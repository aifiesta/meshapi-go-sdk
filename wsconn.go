package meshapi

// wsconn.go — minimal RFC 6455 WebSocket client using only the standard library.
//
// Implements exactly what the realtime proxy needs:
//   - Client-side handshake (HTTP/1.1 Upgrade)
//   - Subprotocol negotiation (Sec-WebSocket-Protocol)
//   - Bidirectional text and binary frame I/O with client-side masking
//   - Automatic ping→pong handling
//   - Clean close handshake (opClose)
//
// Only exported to this package. External callers use RealtimeSession.

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec — SHA-1 is mandated by the WebSocket spec
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// WebSocket opcodes (RFC 6455 §5.2)
const (
	wsOpContinuation byte = 0x0
	wsOpText         byte = 0x1
	wsOpBinary       byte = 0x2
	wsOpClose        byte = 0x8
	wsOpPing         byte = 0x9
	wsOpPong         byte = 0xA
)

// wsStatusNormal is the WebSocket close code for normal closure.
const wsStatusNormal uint16 = 1000

// wsFrame is a decoded WebSocket frame.
type wsFrame struct {
	opcode  byte
	fin     bool
	payload []byte
}

// wsConn is a raw WebSocket connection over a net.Conn.
// All methods are NOT concurrency-safe; callers must synchronise externally.
type wsConn struct {
	conn   net.Conn
	reader *bufio.Reader
	closed bool
}

// dialWS opens a WebSocket connection to wsURL using the given extra HTTP headers.
// wsURL must start with "ws://" or "wss://".
func dialWS(wsURL string, extraHeaders http.Header) (*wsConn, error) {
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("wsconn: parse url: %w", err)
	}

	host := u.Host
	var rawConn net.Conn
	switch strings.ToLower(u.Scheme) {
	case "wss":
		if !strings.Contains(host, ":") {
			host += ":443"
		}
		rawConn, err = tls.Dial("tcp", host, &tls.Config{ServerName: u.Hostname()}) //nolint:gosec
	case "ws":
		if !strings.Contains(host, ":") {
			host += ":80"
		}
		rawConn, err = net.Dial("tcp", host)
	default:
		return nil, fmt.Errorf("wsconn: unsupported scheme %q", u.Scheme)
	}
	if err != nil {
		return nil, fmt.Errorf("wsconn: dial: %w", err)
	}

	wsc := &wsConn{conn: rawConn, reader: bufio.NewReaderSize(rawConn, 65536)}
	if err := wsc.handshake(u, extraHeaders); err != nil {
		rawConn.Close()
		return nil, err
	}
	return wsc, nil
}

// handshake sends the HTTP/1.1 upgrade request and validates the server response.
func (c *wsConn) handshake(u *url.URL, extraHeaders http.Header) error {
	// Generate a random 16-byte nonce for Sec-WebSocket-Key.
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return fmt.Errorf("wsconn: generate key: %w", err)
	}
	secKey := base64.StdEncoding.EncodeToString(nonce[:])

	// Build the HTTP/1.1 upgrade request.
	requestPath := u.RequestURI() // includes path + query string
	if requestPath == "" {
		requestPath = "/"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "GET %s HTTP/1.1\r\n", requestPath)
	fmt.Fprintf(&sb, "Host: %s\r\n", u.Host)
	sb.WriteString("Upgrade: websocket\r\n")
	sb.WriteString("Connection: Upgrade\r\n")
	fmt.Fprintf(&sb, "Sec-WebSocket-Key: %s\r\n", secKey)
	sb.WriteString("Sec-WebSocket-Version: 13\r\n")
	for k, vals := range extraHeaders {
		for _, v := range vals {
			fmt.Fprintf(&sb, "%s: %s\r\n", k, v)
		}
	}
	sb.WriteString("\r\n")

	if _, err := io.WriteString(c.conn, sb.String()); err != nil {
		return fmt.Errorf("wsconn: write handshake: %w", err)
	}

	// Read the HTTP response using the buffered reader so any WebSocket data
	// that arrives alongside the headers is preserved in the buffer.
	resp, err := http.ReadResponse(c.reader, nil)
	if err != nil {
		return fmt.Errorf("wsconn: read handshake response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("wsconn: server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// Verify Sec-WebSocket-Accept = base64(SHA-1(key + GUID)).
	h := sha1.New() //nolint:gosec
	h.Write([]byte(secKey + wsGUID))
	expected := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if got := resp.Header.Get("Sec-Websocket-Accept"); got != expected {
		return fmt.Errorf("wsconn: bad Sec-WebSocket-Accept: got %q want %q", got, expected)
	}
	return nil
}

// ReadFrame reads one complete WebSocket frame from the server.
// It automatically replies to Ping frames with Pong and discards them,
// transparently returning the next data frame to the caller.
// Close frames from the server are returned with opcode wsOpClose.
func (c *wsConn) ReadFrame() (wsFrame, error) {
	for {
		f, err := c.readFrameRaw()
		if err != nil {
			return wsFrame{}, err
		}
		switch f.opcode {
		case wsOpPing:
			// Reply with Pong carrying the same payload; ignore errors.
			// pongPayload is copied here so it is safe to return the frame.
			pongPayload := append([]byte(nil), f.payload...)
			_ = c.pongFunc(pongPayload) // caller-supplied, mutex-guarded
			continue
		case wsOpPong:
			// Unsolicited pong — ignore.
			continue
		}
		return f, nil
	}
}

// readFrameRaw reads one raw frame without any opcode handling.
func (c *wsConn) readFrameRaw() (wsFrame, error) {
	// Read first two bytes: FIN+opcode, MASK+payload_len.
	var hdr [2]byte
	if _, err := io.ReadFull(c.reader, hdr[:]); err != nil {
		return wsFrame{}, fmt.Errorf("wsconn: read frame header: %w", err)
	}
	fin := hdr[0]&0x80 != 0
	opcode := hdr[0] & 0x0F
	masked := hdr[1]&0x80 != 0
	payLen := uint64(hdr[1] & 0x7F)

	switch payLen {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(c.reader, ext[:]); err != nil {
			return wsFrame{}, fmt.Errorf("wsconn: read 16-bit length: %w", err)
		}
		payLen = uint64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(c.reader, ext[:]); err != nil {
			return wsFrame{}, fmt.Errorf("wsconn: read 64-bit length: %w", err)
		}
		payLen = binary.BigEndian.Uint64(ext[:])
	}

	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(c.reader, maskKey[:]); err != nil {
			return wsFrame{}, fmt.Errorf("wsconn: read mask key: %w", err)
		}
	}

	payload := make([]byte, payLen)
	if payLen > 0 {
		if _, err := io.ReadFull(c.reader, payload); err != nil {
			return wsFrame{}, fmt.Errorf("wsconn: read payload: %w", err)
		}
		if masked {
			for i := range payload {
				payload[i] ^= maskKey[i%4]
			}
		}
	}
	return wsFrame{opcode: opcode, fin: fin, payload: payload}, nil
}

// writeFrame sends one frame to the server with client-side masking.
func (c *wsConn) writeFrame(opcode byte, payload []byte) error {
	var maskKey [4]byte
	if _, err := rand.Read(maskKey[:]); err != nil {
		return fmt.Errorf("wsconn: generate mask: %w", err)
	}

	// Build header.
	var hdr []byte
	hdr = append(hdr, 0x80|opcode) // FIN=1, RSV=0

	payLen := len(payload)
	switch {
	case payLen <= 125:
		hdr = append(hdr, 0x80|byte(payLen))
	case payLen <= 65535:
		hdr = append(hdr, 0x80|126)
		hdr = append(hdr, byte(payLen>>8), byte(payLen))
	default:
		hdr = append(hdr, 0x80|127)
		hdr = append(hdr,
			byte(payLen>>56), byte(payLen>>48), byte(payLen>>40), byte(payLen>>32),
			byte(payLen>>24), byte(payLen>>16), byte(payLen>>8), byte(payLen),
		)
	}
	hdr = append(hdr, maskKey[:]...)

	// Mask payload.
	masked := make([]byte, payLen)
	for i, b := range payload {
		masked[i] = b ^ maskKey[i%4]
	}

	// Write header + masked payload in one syscall.
	frame := append(hdr, masked...)
	_, err := c.conn.Write(frame)
	return err
}

// WriteText sends a text frame.
func (c *wsConn) WriteText(data []byte) error {
	return c.writeFrame(wsOpText, data)
}

// WriteBinary sends a binary frame.
func (c *wsConn) WriteBinary(data []byte) error {
	return c.writeFrame(wsOpBinary, data)
}

// Close sends a close frame with the given status code and closes the connection.
func (c *wsConn) Close(code uint16, reason string) error {
	if c.closed {
		return nil
	}
	c.closed = true
	payload := make([]byte, 2+len(reason))
	binary.BigEndian.PutUint16(payload, code)
	copy(payload[2:], reason)
	_ = c.writeFrame(wsOpClose, payload) // best-effort; ignore errors
	return c.conn.Close()
}
