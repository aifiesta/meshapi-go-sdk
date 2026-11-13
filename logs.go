package meshapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

// ---------------------------------------------------------------------------
// Logs permissions
// ---------------------------------------------------------------------------

// Organisation permissions that gate the /v1/logs endpoints.
const (
	// OrgPermissionLogsRead allows reading the org's request records through
	// the Loki-compatible endpoints under /v1/logs/loki/api/v1.
	OrgPermissionLogsRead = "logs:read"

	// OrgPermissionLogsReadBodies allows reading the org's exported request
	// and response bodies through GET /v1/logs/bodies. It is never implied by
	// OrgPermissionLogsRead.
	OrgPermissionLogsReadBodies = "logs:read_bodies"
)

const (
	logsBodiesPath = "/v1/logs/bodies"
	lokiBasePath   = "/v1/logs/loki/api/v1"
)

// ---------------------------------------------------------------------------
// Logs types
// ---------------------------------------------------------------------------

// LokiQueryParams holds the query-string parameters accepted by the
// Loki-compatible read endpoints. Zero values are omitted from the request.
// Time values are passed through as given (RFC 3339 or Unix nanoseconds).
type LokiQueryParams struct {
	// Query is a LogQL expression or stream selector.
	Query string
	// Start is the inclusive start of the time window.
	Start string
	// End is the end of the time window.
	End string
	// Since is a duration relative to End (for example "1h").
	Since string
	// Time is the evaluation time for instant queries.
	Time string
	// Limit caps the number of entries returned.
	Limit int
	// Step is the resolution step for range queries.
	Step string
	// Interval is the interval for range queries and volume ranges.
	Interval string
	// Direction is "forward" or "backward".
	Direction string
	// Match holds series selectors, sent as repeated match[] parameters.
	Match []string
}

func (p *LokiQueryParams) values() url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	set := func(k, val string) {
		if val != "" {
			v.Set(k, val)
		}
	}
	set("query", p.Query)
	set("start", p.Start)
	set("end", p.End)
	set("since", p.Since)
	set("time", p.Time)
	set("step", p.Step)
	set("interval", p.Interval)
	set("direction", p.Direction)
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	for _, m := range p.Match {
		if m != "" {
			v.Add("match[]", m)
		}
	}
	return v
}

// LogBodiesParams holds the query-string parameters for GET /v1/logs/bodies.
type LogBodiesParams struct {
	// Start is the start of the time window.
	Start string
	// End is the end of the time window.
	End string
	// Limit caps the number of records in one page.
	Limit int
	// Cursor is the opaque NextCursor from a previous page.
	Cursor string
}

func (p *LogBodiesParams) values() url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.Start != "" {
		v.Set("start", p.Start)
	}
	if p.End != "" {
		v.Set("end", p.End)
	}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Cursor != "" {
		v.Set("cursor", p.Cursor)
	}
	return v
}

// LogBodiesPage is one page of exported request/response bodies.
type LogBodiesPage struct {
	// Data holds the body records in this page.
	Data []json.RawMessage `json:"data"`
	// NextCursor is an opaque cursor for the next page; nil on the last page.
	NextCursor *string `json:"next_cursor,omitempty"`
}

// ---------------------------------------------------------------------------
// Logs resource
// ---------------------------------------------------------------------------

// LogsResource reads the organisation's request logs.
//
// The Loki-compatible methods return the upstream JSON unchanged, so callers
// can decode it into the Loki response shape they need. A response larger
// than the gateway limit returns a 422 MeshAPIError with code
// "log_result_too_large".
type LogsResource struct {
	http *httpClient
}

func buildPath(path string, v url.Values) string {
	if len(v) == 0 {
		return path
	}
	return path + "?" + v.Encode()
}

func lokiPath(endpoint string, p *LokiQueryParams) string {
	return buildPath(lokiBasePath+endpoint, p.values())
}

func (r *LogsResource) lokiGet(ctx context.Context, endpoint string, p *LokiQueryParams) (json.RawMessage, error) {
	var out json.RawMessage
	if err := r.http.get(ctx, lokiPath(endpoint, p), &out); err != nil {
		return nil, err
	}
	return out, nil
}

var errEmptyName = errors.New("meshapi: name must not be empty")

// Query runs an instant LogQL query (GET /v1/logs/loki/api/v1/query).
func (r *LogsResource) Query(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/query", p)
}

// QueryRange runs a LogQL query over a time range (GET /v1/logs/loki/api/v1/query_range).
func (r *LogsResource) QueryRange(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/query_range", p)
}

// Labels lists label names (GET /v1/logs/loki/api/v1/labels).
func (r *LogsResource) Labels(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/labels", p)
}

// LabelValues lists the values of one label (GET /v1/logs/loki/api/v1/label/{name}/values).
func (r *LogsResource) LabelValues(ctx context.Context, name string, p *LokiQueryParams) (json.RawMessage, error) {
	if name == "" {
		return nil, errEmptyName
	}
	return r.lokiGet(ctx, "/label/"+url.PathEscape(name)+"/values", p)
}

// Series lists streams matching the selectors in p.Match (GET /v1/logs/loki/api/v1/series).
func (r *LogsResource) Series(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/series", p)
}

// DetectedFields lists fields detected in matching log lines (GET /v1/logs/loki/api/v1/detected_fields).
func (r *LogsResource) DetectedFields(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/detected_fields", p)
}

// DetectedFieldValues lists the values of one detected field (GET /v1/logs/loki/api/v1/detected_field/{name}/values).
func (r *LogsResource) DetectedFieldValues(ctx context.Context, name string, p *LokiQueryParams) (json.RawMessage, error) {
	if name == "" {
		return nil, errEmptyName
	}
	return r.lokiGet(ctx, "/detected_field/"+url.PathEscape(name)+"/values", p)
}

// DetectedLabels lists labels detected in matching streams (GET /v1/logs/loki/api/v1/detected_labels).
func (r *LogsResource) DetectedLabels(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/detected_labels", p)
}

// FormatQuery returns p.Query in normalised LogQL form (GET /v1/logs/loki/api/v1/format_query).
func (r *LogsResource) FormatQuery(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/format_query", p)
}

// IndexStats returns index statistics for a query (GET /v1/logs/loki/api/v1/index/stats).
func (r *LogsResource) IndexStats(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/index/stats", p)
}

// IndexVolume returns log volume for a query (GET /v1/logs/loki/api/v1/index/volume).
func (r *LogsResource) IndexVolume(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/index/volume", p)
}

// IndexVolumeRange returns log volume over a time range (GET /v1/logs/loki/api/v1/index/volume_range).
func (r *LogsResource) IndexVolumeRange(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/index/volume_range", p)
}

// Patterns returns log line patterns for a query (GET /v1/logs/loki/api/v1/patterns).
func (r *LogsResource) Patterns(ctx context.Context, p *LokiQueryParams) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/patterns", p)
}

// BuildInfo returns the Loki build information (GET /v1/logs/loki/api/v1/status/buildinfo).
func (r *LogsResource) BuildInfo(ctx context.Context) (json.RawMessage, error) {
	return r.lokiGet(ctx, "/status/buildinfo", nil)
}

// Bodies returns one page of the org's exported request/response bodies
// (GET /v1/logs/bodies). It requires OrgPermissionLogsReadBodies. Pass the
// returned NextCursor in p.Cursor to read the next page.
func (r *LogsResource) Bodies(ctx context.Context, p *LogBodiesParams) (*LogBodiesPage, error) {
	var out LogBodiesPage
	if err := r.http.get(ctx, buildPath(logsBodiesPath, p.values()), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
