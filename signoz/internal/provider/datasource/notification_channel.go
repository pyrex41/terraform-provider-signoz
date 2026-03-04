package datasource

import (
	"context"
	"fmt"

	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &notificationChannelDataSource{}
)

// NewNotificationChannelDataSource is a helper function to simplify the provider implementation.
func NewNotificationChannelDataSource() datasource.DataSource {
	return &notificationChannelDataSource{}
}

// notificationChannelDataSource is the data source implementation.
type notificationChannelDataSource struct {
	client *client.Client
}

// notificationChannelModel maps data source schema data.
type notificationChannelModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	SlackConfigs   types.List   `tfsdk:"slack_configs"`
	WebhookConfigs types.List   `tfsdk:"webhook_configs"`
}

// Configure adds the provider configured client to the data source.
func (d *notificationChannelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *client.Client, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			SigNozNotificationChannel,
		)
		return
	}

	d.client = client
}

// Metadata returns the data source type name.
func (d *notificationChannelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = SigNozNotificationChannel
}

// Schema defines the schema for the data source.
func (d *notificationChannelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a notification channel from SigNoz using its ID.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "ID of the notification channel.",
			},
			attr.Name: schema.StringAttribute{
				Computed:    true,
				Description: "Name of the notification channel.",
			},
			attr.ChannelType: schema.StringAttribute{
				Computed:    true,
				Description: "Type of the notification channel.",
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
				Description: "Slack notification configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						attr.ApiUrl: schema.StringAttribute{
							Computed:    true,
							Sensitive:   true,
							Description: "Slack incoming webhook URL.",
						},
						attr.Channel: schema.StringAttribute{
							Computed:    true,
							Description: "Slack channel.",
						},
						attr.SendResolved: schema.BoolAttribute{
							Computed:    true,
							Description: "Whether to send resolved notifications.",
						},
						attr.Title: schema.StringAttribute{
							Computed:    true,
							Description: "Go template for the Slack message title.",
						},
						attr.Text: schema.StringAttribute{
							Computed:    true,
							Description: "Go template for the Slack message body.",
						},
						attr.IconEmoji: schema.StringAttribute{
							Computed:    true,
							Description: "Slack icon emoji.",
						},
						attr.Username: schema.StringAttribute{
							Computed:    true,
							Description: "Slack username.",
						},
					},
				},
			},
			attr.WebhookConfigs: schema.ListNestedBlock{
				Description: "Webhook notification configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						attr.Url: schema.StringAttribute{
							Computed:    true,
							Description: "Webhook endpoint URL.",
						},
						attr.SendResolved: schema.BoolAttribute{
							Computed:    true,
							Description: "Whether to send resolved notifications.",
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *notificationChannelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data notificationChannelModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := d.client.GetChannel(ctx, data.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, fmt.Errorf("unable to read SigNoz notification channel: %s", err.Error()), SigNozNotificationChannel)
		return
	}

	data.ID = types.StringValue(channel.ID)
	data.Name = types.StringValue(channel.Name)
	data.Type = types.StringValue(channel.Type)
	data.CreatedAt = types.StringValue(channel.CreatedAt)
	data.UpdatedAt = types.StringValue(channel.UpdatedAt)

	// Map slack configs
	if len(channel.SlackConfigs) > 0 {
		slackVals := make([]slackConfigDSModel, 0, len(channel.SlackConfigs))
		for _, sc := range channel.SlackConfigs {
			slackVals = append(slackVals, slackConfigDSModel{
				ApiUrl:       types.StringValue(sc.ApiUrl),
				Channel:      types.StringValue(sc.Channel),
				SendResolved: types.BoolValue(sc.SendResolved),
				Title:        types.StringValue(sc.Title),
				Text:         types.StringValue(sc.Text),
				IconEmoji:    types.StringValue(sc.IconEmoji),
				Username:     types.StringValue(sc.Username),
			})
		}
		slackList, listDiags := types.ListValueFrom(ctx, slackConfigDSObjectType(), slackVals)
		resp.Diagnostics.Append(listDiags...)
		data.SlackConfigs = slackList
	} else {
		data.SlackConfigs = types.ListNull(slackConfigDSObjectType())
	}

	// Map webhook configs
	if len(channel.WebhookConfigs) > 0 {
		webhookVals := make([]webhookConfigDSModel, 0, len(channel.WebhookConfigs))
		for _, wc := range channel.WebhookConfigs {
			webhookVals = append(webhookVals, webhookConfigDSModel{
				Url:          types.StringValue(wc.Url),
				SendResolved: types.BoolValue(wc.SendResolved),
			})
		}
		webhookList, listDiags := types.ListValueFrom(ctx, webhookConfigDSObjectType(), webhookVals)
		resp.Diagnostics.Append(listDiags...)
		data.WebhookConfigs = webhookList
	} else {
		data.WebhookConfigs = types.ListNull(webhookConfigDSObjectType())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// slackConfigDSModel is the datasource equivalent (same fields).
type slackConfigDSModel struct {
	ApiUrl       types.String `tfsdk:"api_url"`
	Channel      types.String `tfsdk:"channel"`
	SendResolved types.Bool   `tfsdk:"send_resolved"`
	Title        types.String `tfsdk:"title"`
	Text         types.String `tfsdk:"text"`
	IconEmoji    types.String `tfsdk:"icon_emoji"`
	Username     types.String `tfsdk:"username"`
}

type webhookConfigDSModel struct {
	Url          types.String `tfsdk:"url"`
	SendResolved types.Bool   `tfsdk:"send_resolved"`
}

func slackConfigDSObjectType() types.ObjectType {
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

func webhookConfigDSObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]tfattr.Type{
			attr.Url:          types.StringType,
			attr.SendResolved: types.BoolType,
		},
	}
}
