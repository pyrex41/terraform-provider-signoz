package resource

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// jsonNormalize returns a plan modifier that normalizes JSON strings so that
// semantically identical JSON (differing only in whitespace or HTML escaping)
// is treated as equal. This prevents perpetual drift when the SigNoz API
// returns compact JSON but the Crossplane CR spec has pretty-printed JSON.
func jsonNormalize() planmodifier.String {
	return &jsonNormalizeModifier{}
}

type jsonNormalizeModifier struct{}

func (m *jsonNormalizeModifier) Description(_ context.Context) string {
	return "Normalizes JSON strings to prevent false diffs from whitespace or HTML escaping differences."
}

func (m *jsonNormalizeModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m *jsonNormalizeModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If there's no state (new resource) or no plan value, skip.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() ||
		req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	// If they're semantically the same JSON, use the state value to suppress the diff.
	if semanticallyEqualJSON(req.PlanValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

// normalizeJSON compacts JSON and un-escapes HTML entities so that
// `>=` and `\u003e=` are treated as identical.
func normalizeJSON(s string) string {
	n, ok := normalizeJSONValue(s)
	if !ok {
		return s
	}
	return n
}

func semanticallyEqualJSON(a, b string) bool {
	// Keep non-JSON behavior strict to avoid suppressing real diffs.
	an, aok := normalizeJSONValue(a)
	bn, bok := normalizeJSONValue(b)
	if !aok || !bok {
		return a == b
	}
	return an == bn
}

func normalizeJSONValue(s string) (string, bool) {
	if s == "" {
		return s, true
	}

	// First, unmarshal to normalize structure
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "", false
	}

	// Re-marshal with no HTML escaping, compact form
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", false
	}

	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	return string(b), true
}
