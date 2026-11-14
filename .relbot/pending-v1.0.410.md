# Release v1.0.410

## Endpoints added

- GET /v1/logs/bodies
- GET /v1/logs/loki/api/v1/detected_field/{name}/values
- GET /v1/logs/loki/api/v1/detected_fields
- GET /v1/logs/loki/api/v1/detected_labels
- GET /v1/logs/loki/api/v1/format_query
- GET /v1/logs/loki/api/v1/index/stats
- GET /v1/logs/loki/api/v1/index/volume
- GET /v1/logs/loki/api/v1/index/volume_range
- GET /v1/logs/loki/api/v1/label/{name}/values
- GET /v1/logs/loki/api/v1/labels
- GET /v1/logs/loki/api/v1/patterns
- GET /v1/logs/loki/api/v1/query
- GET /v1/logs/loki/api/v1/query_range
- GET /v1/logs/loki/api/v1/series
- GET /v1/logs/loki/api/v1/status/buildinfo
- GET /v1/usage/runs/{run_id}
- POST /v1/video/edits
- POST /v1/video/i2v
- POST /v1/video/t2v
- POST /v1/video/v2v

## Schemas changed

- ModelOut.supports_i2v
- ModelOut.supports_t2v
- ModelOut.supports_v2v
- ModelOut.video_options
- ModelProviderOut.supports_i2v
- ModelProviderOut.supports_t2v
- ModelProviderOut.supports_v2v
- ModelProviderOut.video_options
- OrgPermission: 'logs:read'
- OrgPermission: 'logs:read_bodies'
- RejectionFilter.run_id
- UsageEventOut.run_id
- UsageFilter.run_id

_STUB: this is a generated placeholder; full reference, changelog and SDK content is produced by a later automated step._
