package resource

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// compactJSONPlanModifier normalizes JSON string values to compact form during
// planning. This prevents "Provider produced inconsistent result" errors when
// the configuration contains pretty-printed JSON but the provider returns
// compact JSON after apply.
type compactJSONPlanModifier struct{}

func (m compactJSONPlanModifier) Description(_ context.Context) string {
	return "Normalizes JSON to compact form."
}

func (m compactJSONPlanModifier) MarkdownDescription(_ context.Context) string {
	return "Normalizes JSON to compact form."
}

func (m compactJSONPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	raw := req.PlanValue.ValueString()
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(raw)); err != nil {
		return // not valid JSON, leave as-is
	}

	resp.PlanValue = types.StringValue(buf.String())
}

// CompactJSON returns a plan modifier that normalizes JSON to compact form.
func CompactJSON() planmodifier.String {
	return compactJSONPlanModifier{}
}
