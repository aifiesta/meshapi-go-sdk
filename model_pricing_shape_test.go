package meshapi

// ModelPricing must describe the pricing object the gateway actually returns.
//
// Found during the MESH-472 versioning audit: this SDK declares PromptUSDPer1K /
// CompletionUSDPer1K — which prod stopped returning in v1.0.135, and which now have
// ZERO references anywhere in the gateway — and does not declare
// input_usd_per_unit / output_usd_per_unit, which are what replaced them.
//
// encoding/json silently ignores unknown object keys, so the undeclared fields were
// not merely undocumented: they were DISCARDED. That matters most for models which
// are not token-priced — per-second video, per-image, per-1k-chars — because their
// per-1M fields are null BY DESIGN, so the per-unit rate is the only price on the
// wire. A caller listing such a model saw no price at all.
//
// The pre-existing testdata/model_list.json cannot catch this: it encodes the old
// shape, so TestContract_ModelList asserts the SDK parses a response the gateway no
// longer sends. model_list_current.json is what prod sends today.

import (
	"encoding/json"
	"testing"
)

func currentModels(t *testing.T) []ModelInfo {
	t.Helper()
	var models []ModelInfo
	if err := json.Unmarshal(loadFixture(t, "model_list_current.json"), &models); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return models
}

func findModel(t *testing.T, models []ModelInfo, id string) ModelInfo {
	t.Helper()
	for _, m := range models {
		if m.ID == id {
			return m
		}
	}
	t.Fatalf("fixture has no model %q", id)
	return ModelInfo{}
}

func TestPricing_TokenRowExposesThePerUnitRate(t *testing.T) {
	// For token rows the per-unit rate equals the per-1M rate, so either is usable.
	m := findModel(t, currentModels(t), "openai/gpt-4o-mini")
	if m.Pricing == nil {
		t.Fatal("expected pricing")
	}

	if got := deref(m.Pricing.PricingUnit); got != "per_1m_tokens" {
		t.Errorf("PricingUnit = %q, want per_1m_tokens", got)
	}
	if got := deref(m.Pricing.PromptUSDPer1M); got != "0.15000000" {
		t.Errorf("PromptUSDPer1M = %q", got)
	}
	if got := deref(m.Pricing.InputUSDPerUnit); got != "0.15000000" {
		t.Errorf("InputUSDPerUnit = %q, want 0.15000000", got)
	}
	if got := deref(m.Pricing.OutputUSDPerUnit); got != "0.60000000" {
		t.Errorf("OutputUSDPerUnit = %q, want 0.60000000", got)
	}
}

func TestPricing_NonTokenRowHasItsRateOnlyInPerUnit(t *testing.T) {
	// The case the missing fields actually broke. A per-second video model has null
	// for both per-1M fields — a per-1M-token figure is meaningless for it — so
	// InputUSDPerUnit is the ONLY place the price exists.
	m := findModel(t, currentModels(t), "bytedance/seedance-2-5")
	if m.Pricing == nil {
		t.Fatal("expected pricing")
	}

	if got := deref(m.Pricing.PricingUnit); got != "per_second" {
		t.Errorf("PricingUnit = %q, want per_second", got)
	}
	if m.Pricing.PromptUSDPer1M != nil || m.Pricing.CompletionUSDPer1M != nil {
		t.Error("per-1M fields should be null for a per-second row")
	}
	if got := deref(m.Pricing.InputUSDPerUnit); got != "10.70000000" {
		t.Errorf("InputUSDPerUnit = %q, want 10.70000000", got)
	}
	if got := deref(m.Pricing.OutputUSDPerUnit); got != "6.40000000" {
		t.Errorf("OutputUSDPerUnit = %q, want 6.40000000", got)
	}
}

func TestPricing_RateIsOnlyInterpretableAlongsidePricingUnit(t *testing.T) {
	// InputUSDPerUnit is a bare number; PricingUnit is what makes it a price. A
	// response carrying one without the other is not usable.
	for _, m := range currentModels(t) {
		if m.Pricing != nil && m.Pricing.InputUSDPerUnit != nil && m.Pricing.PricingUnit == nil {
			t.Errorf("%s has a rate but no pricing_unit", m.ID)
		}
	}
}

func TestPricing_RetiredPer1KFieldsAreAbsentFromACurrentResponse(t *testing.T) {
	// The honest outcome of the drift: against a real gateway these are nil forever.
	// A caller branching on them silently sees "unpriced".
	m := findModel(t, currentModels(t), "openai/gpt-4o-mini")
	if m.Pricing.PromptUSDPer1K != nil || m.Pricing.CompletionUSDPer1K != nil {
		t.Error("the retired per-1k fields should be absent from a current response")
	}
}

func TestPricing_RetiredShapeStillParses(t *testing.T) {
	// Kept declared, not deleted: a caller that still reads them keeps compiling, and
	// an old recorded response still round-trips.
	var models []ModelInfo
	if err := json.Unmarshal(loadFixture(t, "model_list.json"), &models); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := deref(models[0].Pricing.PromptUSDPer1K); got != "0.000150" {
		t.Errorf("PromptUSDPer1K = %q, want 0.000150", got)
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
