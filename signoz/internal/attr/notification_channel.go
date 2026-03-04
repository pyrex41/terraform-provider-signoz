package attr

const (
	// NotificationChannel resource attributes.
	// Name is declared in dashboard.go.
	// Title is declared in dashboard.go.
	// CreatedAt is declared in dashboard.go.
	// UpdatedAt is declared in dashboard.go.
	ChannelType  = "type"
	SlackConfigs = "slack_configs"

	// Slack config sub-attributes.
	ApiUrl       = "api_url"
	Channel      = "channel"
	SendResolved = "send_resolved"
	Text         = "text"
	IconEmoji    = "icon_emoji"
	Username     = "username"

	// Webhook config sub-attributes.
	WebhookConfigs = "webhook_configs"
	Url            = "url"
)
