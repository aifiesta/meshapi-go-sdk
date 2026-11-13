package meshapi

import (
	"context"
	"errors"
	"testing"
)

func TestLokiPath(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		params   *LokiQueryParams
		want     string
	}{
		{"nil params", "/labels", nil, "/v1/logs/loki/api/v1/labels"},
		{"empty params", "/status/buildinfo", &LokiQueryParams{}, "/v1/logs/loki/api/v1/status/buildinfo"},
		{
			"instant query",
			"/query",
			&LokiQueryParams{Query: `{service="api"}`, Time: "1700000000000000000", Limit: 10},
			"/v1/logs/loki/api/v1/query?limit=10&query=%7Bservice%3D%22api%22%7D&time=1700000000000000000",
		},
		{
			"range query",
			"/query_range",
			&LokiQueryParams{Query: "{a=\"b\"}", Start: "1", End: "2", Step: "60s", Direction: "backward"},
			"/v1/logs/loki/api/v1/query_range?direction=backward&end=2&query=%7Ba%3D%22b%22%7D&start=1&step=60s",
		},
		{
			"series match",
			"/series",
			&LokiQueryParams{Match: []string{"{a=\"1\"}", "", "{b=\"2\"}"}},
			"/v1/logs/loki/api/v1/series?match%5B%5D=%7Ba%3D%221%22%7D&match%5B%5D=%7Bb%3D%222%22%7D",
		},
		{"negative limit omitted", "/labels", &LokiQueryParams{Limit: -1, Since: "1h"}, "/v1/logs/loki/api/v1/labels?since=1h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lokiPath(tt.endpoint, tt.params); got != tt.want {
				t.Errorf("lokiPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogBodiesPath(t *testing.T) {
	tests := []struct {
		name   string
		params *LogBodiesParams
		want   string
	}{
		{"nil params", nil, "/v1/logs/bodies"},
		{"empty params", &LogBodiesParams{}, "/v1/logs/bodies"},
		{"cursor", &LogBodiesParams{Cursor: "abc"}, "/v1/logs/bodies?cursor=abc"},
		{
			"window and limit",
			&LogBodiesParams{Start: "2026-11-12T11:00:00Z", End: "2026-11-12T12:00:00Z", Limit: 50},
			"/v1/logs/bodies?end=2026-11-12T12%3A00%3A00Z&limit=50&start=2026-11-12T11%3A00%3A00Z",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildPath(logsBodiesPath, tt.params.values()); got != tt.want {
				t.Errorf("path = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLokiNamedPathEscaping(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{"plain label", "/label/service/values", "/v1/logs/loki/api/v1/label/service/values"},
		{"detected field", "/detected_field/latency_ms/values", "/v1/logs/loki/api/v1/detected_field/latency_ms/values"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lokiPath(tt.endpoint, nil); got != tt.want {
				t.Errorf("lokiPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogsEmptyNameRejected(t *testing.T) {
	r := &LogsResource{}
	ctx := context.Background()
	tests := []struct {
		name string
		call func() error
	}{
		{"LabelValues", func() error { _, err := r.LabelValues(ctx, "", nil); return err }},
		{"DetectedFieldValues", func() error { _, err := r.DetectedFieldValues(ctx, "", nil); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); !errors.Is(err, errEmptyName) {
				t.Errorf("err = %v, want %v", err, errEmptyName)
			}
		})
	}
}

func TestLogsPermissionValues(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"logs read", OrgPermissionLogsRead, "logs:read"},
		{"logs read bodies", OrgPermissionLogsReadBodies, "logs:read_bodies"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
