package resource

import (
	"context"
	"fmt"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &notificationChannelResource{}
	_ resource.ResourceWithConfigure   = &notificationChannelResource{}
	_ resource.ResourceWithImportState = &notificationChannelResource{}
)

// NewNotificationChannelResource is a helper function to simplify the provider implementation.
func NewNotificationChannelResource() resource.Resource {
	return &notificationChannelResource{}
}

// notificationChannelResource is the resource implementation.
type notificationChannelResource struct {
	client *client.Client
}

// notificationChannelResourceModel maps the resource schema data.
type notificationChannelResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	SlackConfigs   types.List   `tfsdk:"slack_configs"`
	WebhookConfigs types.List   `tfsdk:"webhook_configs"`
}

// slackConfigModel maps a single slack_configs block.
type slackConfigModel struct {
	ApiUrl       types.String `tfsdk:"api_url"`
	Channel      types.String `tfsdk:"channel"`
	SendResolved types.Bool   `tfsdk:"send_resolved"`
	Title        types.String `tfsdk:"title"`
	Text         types.String `tfsdk:"text"`
	IconEmoji    types.String `tfsdk:"icon_emoji"`
	Username     types.String `tfsdk:"username"`
}

// webhookConfigModel maps a single webhook_configs block.
type webhookConfigModel struct {
	Url          types.String `tfsdk:"url"`
	SendResolved types.Bool   `tfsdk:"send_resolved"`
}

func slackConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]tfattr.Type{
			attr.ApiUrl:       types.StringType,
			attr.Channel:      types.StringType,
			attr.SendResolved: types.BoolType,
			attr.Title:        types.StringType,
			attr.Text:         types.StringType,
			attr.IconEmoji:    types.StringType,
			attr.Username:     types.StringType,
		},
	}
}

func webhookConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]tfattr.Type{
			attr.Url:          types.StringType,
			attr.SendResolved: types.BoolType,
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *notificationChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *client.Client, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			operationConfigure, SigNozNotificationChannel,
		)
		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *notificationChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = SigNozNotificationChannel
}

// Schema defines the schema for the resource.
func (r *notificationChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages notification channel resources in SigNoz.",
		Attributes: map[string]schema.Attribute{
			attr.Name: schema.StringAttribute{
				Required:    true,
				Description: "Name of the notification channel.",
			},
			attr.ChannelType: schema.StringAttribute{
				Computed:    true,
				Description: "Type of the notification channel (derived from config block).",
			},
			attr.ID: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Unique ID for the notification channel. If provided during creation, the provider will adopt the existing channel.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the notification channel.",
			},
			attr.UpdatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the notification channel.",
			},
		},
		Blocks: map[string]schema.Block{
			attr.SlackConfigs: schema.ListNestedBlock{
				Description: "Slack notification configuration. At most one block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						attr.ApiUrl: schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "Slack incoming webhook URL.",
						},
						attr.Channel: schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Slack channel to post to (e.g. #ops-alerts).",
						},
						attr.SendResolved: schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
							Description: "Whether to send a notification when the alert resolves.",
						},
						attr.Title: schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Go template for the Slack message title.",
						},
						attr.Text: schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Go template for the Slack message body.",
						},
						attr.IconEmoji: schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Slack icon emoji for the bot.",
						},
						attr.Username: schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Slack username for the bot.",
						},
					},
				},
			},
			attr.WebhookConfigs: schema.ListNestedBlock{
				Description: "Webhook notification configuration. At most one block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						attr.Url: schema.StringAttribute{
							Required:    true,
							Description: "Webhook endpoint URL.",
						},
						attr.SendResolved: schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
							Description: "Whether to send a notification when the alert resolves.",
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *notificationChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationChannelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := buildChannelPayload(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var channel *model.NotificationChannel
	var err error

	// If an ID is provided (e.g., from Crossplane external-name), adopt the
	// existing channel by updating it instead of creating a new one.
	if !plan.ID.IsNull() && !plan.ID.IsUnknown() && plan.ID.ValueString() != "" {
		existingID := plan.ID.ValueString()
		tflog.Debug(ctx, "Adopting existing channel", map[string]any{"id": existingID})

		err = r.client.UpdateChannel(ctx, existingID, payload)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adopting notification channel",
				fmt.Sprintf("Could not adopt channel %q: %s", existingID, err.Error()),
			)
			return
		}

		channel, err = r.client.GetChannel(ctx, existingID)
		if err != nil {
			addErr(&resp.Diagnostics, err, operationCreate, SigNozNotificationChannel)
			return
		}
	} else {
		tflog.Debug(ctx, "Creating notification channel", map[string]any{"name": payload.Name})

		channel, err = r.client.CreateChannel(ctx, payload)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating notification channel",
				"Could not create notification channel: "+err.Error(),
			)
			return
		}
	}

	// Save plan config blocks before the API read overwrites them.
	planSlackConfigs := plan.SlackConfigs
	planWebhookConfigs := plan.WebhookConfigs

	mapChannelToState(ctx, channel, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve plan config blocks when the API response doesn't include them.
	// The SigNoz GET endpoint may omit nested config objects, which causes
	// Terraform's post-apply consistency check to fail with
	// "block count changed from 1 to 0".
	if plan.SlackConfigs.IsNull() && !planSlackConfigs.IsNull() {
		plan.SlackConfigs = planSlackConfigs
	}
	if plan.WebhookConfigs.IsNull() && !planWebhookConfigs.IsNull() {
		plan.WebhookConfigs = planWebhookConfigs
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *notificationChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save existing config blocks before the API read overwrites them.
	prevSlackConfigs := state.SlackConfigs
	prevWebhookConfigs := state.WebhookConfigs

	channel, err := r.client.GetChannel(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, SigNozNotificationChannel)
		return
	}
	if channel == nil {
		// Channel not found — remove from state so Terraform knows to recreate.
		resp.State.RemoveResource(ctx)
		return
	}

	mapChannelToState(ctx, channel, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve config blocks when the API doesn't return them.
	if state.SlackConfigs.IsNull() && !prevSlackConfigs.IsNull() {
		state.SlackConfigs = prevSlackConfigs
	}
	if state.WebhookConfigs.IsNull() && !prevWebhookConfigs.IsNull() {
		state.WebhookConfigs = prevWebhookConfigs
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *notificationChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state notificationChannelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := buildChannelPayload(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	payload.ID = state.ID.ValueString()

	err := r.client.UpdateChannel(ctx, state.ID.ValueString(), payload)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozNotificationChannel)
		return
	}

	// Read back the channel to capture server-computed fields (updated_at, etc.).
	// If the read succeeds and includes config blocks, use them; otherwise
	// preserve the plan values (the API may not return nested configs on GET).
	channel, err := r.client.GetChannel(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozNotificationChannel)
		return
	}

	mapChannelToState(ctx, channel, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve plan config blocks when the API response doesn't include them.
	// The SigNoz GET endpoint may omit nested config objects.
	var originalPlan notificationChannelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &originalPlan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.SlackConfigs.IsNull() && !originalPlan.SlackConfigs.IsNull() {
		plan.SlackConfigs = originalPlan.SlackConfigs
	}
	if plan.WebhookConfigs.IsNull() && !originalPlan.WebhookConfigs.IsNull() {
		plan.WebhookConfigs = originalPlan.WebhookConfigs
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *notificationChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteChannel(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationDelete, SigNozNotificationChannel)
		return
	}
}

// ImportState imports Terraform state into the resource.
func (r *notificationChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildChannelPayload constructs the API model from the Terraform plan.
func buildChannelPayload(ctx context.Context, plan *notificationChannelResourceModel, diags *diag.Diagnostics) *model.NotificationChannel {
	payload := &model.NotificationChannel{
		Name: plan.Name.ValueString(),
	}

	// Extract slack_configs
	if !plan.SlackConfigs.IsNull() && !plan.SlackConfigs.IsUnknown() {
		var slackModels []slackConfigModel
		diags.Append(plan.SlackConfigs.ElementsAs(ctx, &slackModels, false)...)
		for _, sc := range slackModels {
			payload.SlackConfigs = append(payload.SlackConfigs, model.SlackConfig{
				SendResolved: sc.SendResolved.ValueBool(),
				ApiUrl:       sc.ApiUrl.ValueString(),
				Channel:      sc.Channel.ValueString(),
				Title:        sc.Title.ValueString(),
				Text:         sc.Text.ValueString(),
				IconEmoji:    sc.IconEmoji.ValueString(),
				Username:     sc.Username.ValueString(),
			})
		}
	}

	// Extract webhook_configs
	if !plan.WebhookConfigs.IsNull() && !plan.WebhookConfigs.IsUnknown() {
		var webhookModels []webhookConfigModel
		diags.Append(plan.WebhookConfigs.ElementsAs(ctx, &webhookModels, false)...)
		for _, wc := range webhookModels {
			payload.WebhookConfigs = append(payload.WebhookConfigs, model.WebhookConfig{
				SendResolved: wc.SendResolved.ValueBool(),
				Url:          wc.Url.ValueString(),
			})
		}
	}

	return payload
}

// mapChannelToState maps the API response to the Terraform state model.
func mapChannelToState(ctx context.Context, channel *model.NotificationChannel, state *notificationChannelResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(channel.ID)
	state.Name = types.StringValue(channel.Name)
	state.Type = types.StringValue(channel.Type)
	state.CreatedAt = types.StringValue(channel.CreatedAt)
	state.UpdatedAt = types.StringValue(channel.UpdatedAt)

	// Map slack configs
	if len(channel.SlackConfigs) > 0 {
		var slackElems []slackConfigModel
		for _, sc := range channel.SlackConfigs {
			slackElems = append(slackElems, slackConfigModel{
				ApiUrl:       types.StringValue(sc.ApiUrl),
				Channel:      types.StringValue(sc.Channel),
				SendResolved: types.BoolValue(sc.SendResolved),
				Title:        types.StringValue(sc.Title),
				Text:         types.StringValue(sc.Text),
				IconEmoji:    types.StringValue(sc.IconEmoji),
				Username:     types.StringValue(sc.Username),
			})
		}
		slackList, listDiags := types.ListValueFrom(ctx, slackConfigObjectType(), slackElems)
		diags.Append(listDiags...)
		state.SlackConfigs = slackList
	} else {
		state.SlackConfigs = types.ListNull(slackConfigObjectType())
	}

	// Map webhook configs
	if len(channel.WebhookConfigs) > 0 {
		var webhookElems []webhookConfigModel
		for _, wc := range channel.WebhookConfigs {
			webhookElems = append(webhookElems, webhookConfigModel{
				Url:          types.StringValue(wc.Url),
				SendResolved: types.BoolValue(wc.SendResolved),
			})
		}
		webhookList, listDiags := types.ListValueFrom(ctx, webhookConfigObjectType(), webhookElems)
		diags.Append(listDiags...)
		state.WebhookConfigs = webhookList
	} else {
		state.WebhookConfigs = types.ListNull(webhookConfigObjectType())
	}
}
