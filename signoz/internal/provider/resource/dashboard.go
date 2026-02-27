package resource

import (
	"context"
	"fmt"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &dashboardResource{}
	_ resource.ResourceWithConfigure   = &dashboardResource{}
	_ resource.ResourceWithImportState = &dashboardResource{}
)

// NewDashboardResource is a helper function to simplify the provider implementation.
func NewDashboardResource() resource.Resource {
	return &dashboardResource{}
}

// dashboardResource is the resource implementation.
type dashboardResource struct {
	client *client.Client
}

// dashboardResourceModel maps the resource schema data.
type dashboardResourceModel struct {
	CollapsableRowsMigrated types.Bool   `tfsdk:"collapsable_rows_migrated"`
	CreatedAt               types.String `tfsdk:"created_at"`
	CreatedBy               types.String `tfsdk:"created_by"`
	Description             types.String `tfsdk:"description"`
	ID                      types.String `tfsdk:"id"`
	Layout                  types.String `tfsdk:"layout"`
	Name                    types.String `tfsdk:"name"`
	PanelMap                types.String `tfsdk:"panel_map"`
	Source                  types.String `tfsdk:"source"`
	Tags                    types.List   `tfsdk:"tags"`
	Title                   types.String `tfsdk:"title"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
	UpdatedBy               types.String `tfsdk:"updated_by"`
	UploadedGrafana         types.Bool   `tfsdk:"uploaded_grafana"`
	Variables               types.String `tfsdk:"variables"`
	Version                 types.String `tfsdk:"version"`
	Widgets                 types.String `tfsdk:"widgets"`
}

// Configure adds the provider configured client to the resource.
func (r *dashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected resource configure type. Expected *client.Client, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			operationConfigure, SigNozDashboard,
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *dashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = SigNozDashboard
}

// Schema defines the schema for the resource.
func (r *dashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages dashboard resources in SigNoz.",
		Attributes: map[string]schema.Attribute{
			attr.CollapsableRowsMigrated: schema.BoolAttribute{
				Required: true,
			},
			attr.Description: schema.StringAttribute{
				Required:    true,
				Description: "Description of the dashboard.",
			},
			attr.Layout: schema.StringAttribute{
				Required:    true,
				Description: "Layout of the dashboard.",
				PlanModifiers: []planmodifier.String{
					jsonNormalize(),
				},
			},
			attr.Name: schema.StringAttribute{
				Required:    true,
				Description: "Name of the dashboard.",
			},
			attr.PanelMap: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					jsonNormalize(),
				},
			},
			attr.Source: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Source of the dashboard. By default, it is <SIGNOZ_ENDPOINT>/dashboard.",
			},
			attr.Tags: schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Tags of the dashboard.",
			},
			attr.Title: schema.StringAttribute{
				Required:    true,
				Description: "Title of the dashboard.",
			},
			attr.UploadedGrafana: schema.BoolAttribute{
				Required: true,
			},
			attr.Variables: schema.StringAttribute{
				Required:    true,
				Description: "Variables for the dashboard.",
				PlanModifiers: []planmodifier.String{
					jsonNormalize(),
				},
			},
			attr.Widgets: schema.StringAttribute{
				Required:    true,
				Description: "Widgets for the dashboard.",
				PlanModifiers: []planmodifier.String{
					jsonNormalize(),
				},
			},
			attr.Version: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Version of the dashboard. If not set, uses the server default. The server may normalize this value (e.g., v4 → v5).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// ID is computed by default but can be optionally provided to adopt
			// an existing SigNoz dashboard (e.g., via Crossplane external-name).
			attr.ID: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Unique ID for the dashboard. If provided during creation, the provider will adopt the existing dashboard instead of creating a new one.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the dashboard.",
			},
			attr.CreatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Creator of the dashboard.",
			},
			attr.UpdatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the dashboard.",
			},
			attr.UpdatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Last updater of the dashboard.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body.
	dashboardPayload := &model.Dashboard{
		CollapsableRowsMigrated: plan.CollapsableRowsMigrated.ValueBool(),
		Description:             plan.Description.ValueString(),
		Name:                    plan.Name.ValueString(),
		Title:                   plan.Title.ValueString(),
		UploadedGrafana:         plan.UploadedGrafana.ValueBool(),
		Version:                 plan.Version.ValueString(),
	}

	err := dashboardPayload.SetLayout(plan.Layout)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, SigNozDashboard)
		return
	}
	err = dashboardPayload.SetPanelMap(plan.PanelMap)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, SigNozDashboard)
		return
	}
	dashboardPayload.SetTags(plan.Tags)
	err = dashboardPayload.SetVariables(plan.Variables)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, SigNozDashboard)
		return
	}
	err = dashboardPayload.SetWidgets(plan.Widgets)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, SigNozDashboard)
		return
	}

	// readBackState reads the dashboard from the API after create/adopt.
	// It updates computed fields (ID, timestamps, source) from the server
	// response, but PRESERVES the plan's widgets and variables values.
	//
	// Why: SigNoz's POST handler runs a v5 migration that mutates widget
	// JSON in breaking ways (converts $var → {{.var}}, injects
	// #SIGNOZ_VALUE orderBy, changes op:"in" → op:"="). If we store the
	// mutated response, the Terraform state becomes the broken version,
	// and Crossplane/upjet will adopt it as the desired state. By keeping
	// the plan's values, the next reconciliation cycle detects a diff and
	// PUTs the correct (unmutated) JSON back — SigNoz stores PUT payloads
	// verbatim without migration.
	readBackState := func(dashboardID string) bool {
		dashboard, getErr := r.client.GetDashboard(ctx, dashboardID)
		if getErr != nil {
			addErr(&resp.Diagnostics, getErr, operationCreate, SigNozDashboard)
			return false
		}
		// Update computed-only fields from the server response.
		plan.CreatedAt = types.StringValue(dashboard.CreatedAt)
		plan.CreatedBy = types.StringValue(dashboard.CreatedBy)
		plan.ID = types.StringValue(dashboard.ID)
		plan.Source = types.StringValue(dashboard.Data.Source)
		plan.UpdatedAt = types.StringValue(dashboard.UpdatedAt)
		plan.UpdatedBy = types.StringValue(dashboard.UpdatedBy)

		// For version: use plan value if set, otherwise use server value.
		if plan.Version.IsNull() || plan.Version.IsUnknown() || plan.Version.ValueString() == "" {
			plan.Version = types.StringValue(dashboard.Data.Version)
		}

		// Preserve plan values for widgets, variables, layout, and other
		// user-provided fields. The server may have mutated these during
		// POST (v5 migration), but the plan values are what the user intended.
		// Do NOT overwrite them with the server's mutated response.

		return true
	}

	// If an ID is provided (e.g., from Crossplane external-name), adopt the
	// existing dashboard by updating it instead of creating a new one.
	if !plan.ID.IsNull() && !plan.ID.IsUnknown() && plan.ID.ValueString() != "" {
		existingID := plan.ID.ValueString()
		tflog.Debug(ctx, "Adopting existing dashboard", map[string]any{"id": existingID})

		err = r.client.UpdateDashboard(ctx, existingID, dashboardPayload)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adopting dashboard",
				fmt.Sprintf("Could not adopt dashboard %q, unexpected error: %s", existingID, err.Error()),
			)
			return
		}

		if !readBackState(existingID) {
			return
		}
	} else {
		tflog.Debug(ctx, "Creating dashboard", map[string]any{"dashboard": dashboardPayload})

		dashboard, createErr := r.client.CreateDashboard(ctx, dashboardPayload)
		if createErr != nil {
			resp.Diagnostics.AddError(
				"Error creating dashboard",
				"Could not create dashboard, unexpected error: "+createErr.Error(),
			)
			return
		}

		tflog.Debug(ctx, "Created dashboard", map[string]any{"dashboard": dashboard})

		if !readBackState(dashboard.ID) {
			return
		}
	}

	// Set state to populated data.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dashboardResourceModel
	var diag diag.Diagnostics
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading dashboard", map[string]any{"dashboard": state.ID.ValueString()})

	// Get refreshed dashboard from SigNoz.
	dashboard, err := r.client.GetDashboard(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, SigNozDashboard)
		return
	}

	// Update computed-only fields from the server response.
	state.CreatedAt = types.StringValue(dashboard.CreatedAt)
	state.CreatedBy = types.StringValue(dashboard.CreatedBy)
	state.ID = types.StringValue(dashboard.ID)
	state.Source = types.StringValue(dashboard.Data.Source)
	state.UpdatedAt = types.StringValue(dashboard.UpdatedAt)
	state.UpdatedBy = types.StringValue(dashboard.UpdatedBy)

	// For user-provided fields, we read the server values but use semantic
	// JSON comparison to decide whether to update state. If the server's
	// JSON is semantically identical to our state, keep our state value
	// (preserving formatting). If it genuinely differs (e.g., someone
	// edited the dashboard in the UI), adopt the server's value.
	serverWidgets, err := dashboard.Data.WidgetsToTerraform()
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, SigNozDashboard)
		return
	}
	if !state.Widgets.IsNull() && !state.Widgets.IsUnknown() &&
		semanticallyEqualJSON(state.Widgets.ValueString(), serverWidgets.ValueString()) {
		// Server returned equivalent JSON — keep our state value to prevent
		// drift from server-side normalization (v5 migration, key reordering).
	} else {
		state.Widgets = serverWidgets
	}

	serverVars, err := dashboard.Data.VariablesToTerraform()
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, SigNozDashboard)
		return
	}
	if !state.Variables.IsNull() && !state.Variables.IsUnknown() &&
		semanticallyEqualJSON(state.Variables.ValueString(), serverVars.ValueString()) {
		// Keep state value.
	} else {
		state.Variables = serverVars
	}

	serverLayout, err := dashboard.Data.LayoutToTerraform()
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, SigNozDashboard)
		return
	}
	if !state.Layout.IsNull() && !state.Layout.IsUnknown() &&
		semanticallyEqualJSON(state.Layout.ValueString(), serverLayout.ValueString()) {
		// Keep state value.
	} else {
		state.Layout = serverLayout
	}

	serverPanelMap, err := dashboard.Data.PanelMapToTerraform()
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, SigNozDashboard)
		return
	}
	if !state.PanelMap.IsNull() && !state.PanelMap.IsUnknown() &&
		semanticallyEqualJSON(state.PanelMap.ValueString(), serverPanelMap.ValueString()) {
		// Keep state value.
	} else {
		state.PanelMap = serverPanelMap
	}

	// Always refresh simple scalar fields from the server.
	state.CollapsableRowsMigrated = types.BoolValue(dashboard.Data.CollapsableRowsMigrated)
	state.Description = types.StringValue(dashboard.Data.Description)
	state.Name = types.StringValue(dashboard.Data.Name)
	state.Title = types.StringValue(dashboard.Data.Title)
	state.UploadedGrafana = types.BoolValue(dashboard.Data.UploadedGrafana)
	state.Version = types.StringValue(dashboard.Data.Version)

	state.Tags, diag = dashboard.Data.TagsToTerraform()
	resp.Diagnostics.Append(diag...)

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan, state dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan.
	var err error
	dashboardUpdate := &model.Dashboard{
		CollapsableRowsMigrated: plan.CollapsableRowsMigrated.ValueBool(),
		Description:             plan.Description.ValueString(),
		Name:                    plan.Name.ValueString(),
		Title:                   plan.Title.ValueString(),
		UploadedGrafana:         plan.UploadedGrafana.ValueBool(),
		Version:                 plan.Version.ValueString(),
	}

	err = dashboardUpdate.SetLayout(plan.Layout)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozDashboard)
		return
	}
	err = dashboardUpdate.SetPanelMap(plan.PanelMap)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozDashboard)
		return
	}
	dashboardUpdate.SetTags(plan.Tags)
	err = dashboardUpdate.SetVariables(plan.Variables)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozDashboard)
		return
	}
	err = dashboardUpdate.SetWidgets(plan.Widgets)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozDashboard)
		return
	}

	// Update existing dashboard.
	err = r.client.UpdateDashboard(ctx, state.ID.ValueString(), dashboardUpdate)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozDashboard)
		return
	}

	// Fetch updated dashboard to refresh computed fields only.
	// We preserve the plan's widgets, variables, layout, and other
	// user-provided fields — SigNoz stores PUT payloads verbatim so the
	// response should match, but we keep plan values as the source of truth
	// to prevent any server-side normalization from corrupting state.
	dashboard, err := r.client.GetDashboard(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozDashboard)
		return
	}

	// Update computed-only fields from the server response.
	plan.CreatedAt = types.StringValue(dashboard.CreatedAt)
	plan.CreatedBy = types.StringValue(dashboard.CreatedBy)
	plan.ID = types.StringValue(dashboard.ID)
	plan.Source = types.StringValue(dashboard.Data.Source)
	plan.UpdatedAt = types.StringValue(dashboard.UpdatedAt)
	plan.UpdatedBy = types.StringValue(dashboard.UpdatedBy)

	// Preserve plan values for user-provided fields (widgets, variables,
	// layout, title, description, etc.) — they are the source of truth.

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *dashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing dashboard.
	err := r.client.DeleteDashboard(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationDelete, SigNozDashboard)
		return
	}
}

// ImportState imports Terraform state into the resource.
func (r *dashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
