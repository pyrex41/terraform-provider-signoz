package model

// NotificationChannel model.
type NotificationChannel struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Type           string         `json:"type,omitempty"`
	CreatedAt      string         `json:"created_at,omitempty"`
	UpdatedAt      string         `json:"updated_at,omitempty"`
	SlackConfigs   []SlackConfig  `json:"slack_configs,omitempty"`
	WebhookConfigs []WebhookConfig `json:"webhook_configs,omitempty"`
}

// SlackConfig for Slack notification channel.
type SlackConfig struct {
	SendResolved bool   `json:"send_resolved"`
	ApiUrl       string `json:"api_url"`
	Channel      string `json:"channel,omitempty"`
	Title        string `json:"title,omitempty"`
	Text         string `json:"text,omitempty"`
	IconEmoji    string `json:"icon_emoji,omitempty"`
	Username     string `json:"username,omitempty"`
}

// WebhookConfig for Webhook notification channel.
type WebhookConfig struct {
	SendResolved bool   `json:"send_resolved"`
	Url          string `json:"url"`
}
