package meshapi

import (
	"context"
	"encoding/json"
)

// ModelsResource provides access to the GET /v1/models model catalog.
type ModelsResource struct {
	http *httpClient
}

// List returns the models currently available in the catalog.
// Use only model ids returned here in requests.
func (r *ModelsResource) List(ctx context.Context) (*ModelList, error) {
	var out ModelList
	if err := r.http.get(ctx, "/v1/models", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Cost breakdown (spec v1.0.371)
// ---------------------------------------------------------------------------

// CostLineOut is one line item of a request's cost breakdown.
//
// Breaking in v1.0.371: RatePer1MUSD and Tokens are now optional and nullable,
// because non-token charges (for example images or video seconds) have no
// per-1M-token rate or token count. Read Quantity, Unit and RateUSD for those
// lines instead.
type CostLineOut struct {
	// RatePer1MUSD is the per-1M-token rate as a decimal string; nil for
	// lines that are not token-priced.
	RatePer1MUSD *string `json:"rate_per_1m_usd,omitempty"`
	// Tokens is the token count for the line; nil for lines that are not
	// token-priced.
	Tokens *int `json:"tokens,omitempty"`
	// Quantity is the number of units billed on this line.
	Quantity *json.Number `json:"quantity,omitempty"`
	// Unit names what Quantity counts.
	Unit *string `json:"unit,omitempty"`
	// RateUSD is the rate applied per Unit, as a decimal string.
	RateUSD *string `json:"rate_usd,omitempty"`
	// Charged reports whether this line was billed to the caller.
	Charged *bool `json:"charged,omitempty"`
	// Modifiers carries the rate-card adjustments that applied to this line.
	Modifiers json.RawMessage `json:"modifiers,omitempty"`
}

// CostBreakdownOut is the itemised cost of a request.
type CostBreakdownOut struct {
	// Lines are the individual cost line items.
	Lines []CostLineOut `json:"lines,omitempty"`
	// RoundingUSD is the rounding adjustment applied to the total, as a
	// decimal string.
	RoundingUSD *string `json:"rounding_usd,omitempty"`
}

// ---------------------------------------------------------------------------
// Usage filter (spec v1.0.371)
// ---------------------------------------------------------------------------

// UsageFilter narrows usage event listings, counts, summaries and exports.
type UsageFilter struct {
	// Provider filters by provider slug.
	Provider []string `json:"provider,omitempty"`
	// FaultOwner filters by who a failure is attributed to.
	FaultOwner []string `json:"fault_owner,omitempty"`
	// IsBYOK filters to requests served on your own provider key (true) or not (false).
	IsBYOK *bool `json:"is_byok,omitempty"`
	// Stream filters to streaming (true) or non-streaming (false) requests.
	Stream *bool `json:"stream,omitempty"`
	// TrafficType filters by traffic class.
	TrafficType []string `json:"traffic_type,omitempty"`
	// IsRetryAttempt selects which attempts to include. Omit it to get only the
	// served attempt; pass []bool{true, false} to include retry attempts too.
	IsRetryAttempt []bool `json:"is_retry_attempt,omitempty"`
}
