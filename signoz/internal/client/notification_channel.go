package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// channelPath - URL path for notification channel APIs.
	channelPath = "api/v1/channels"
)

// GetChannel - Returns a specific notification channel.
// Returns nil, nil when the channel is not found (empty ID or API returns empty data).
func (c *Client) GetChannel(ctx context.Context, channelID string) (*model.NotificationChannel, error) {
	if channelID == "" {
		return nil, nil
	}

	url, err := url.JoinPath(c.hostURL.String(), channelPath, channelID)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var bodyObj channelResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal channel response: %w (body: %s)", err, truncateStr(string(body), 500))
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "GetChannel: error while fetching channel", map[string]any{
			"error": bodyObj.Error,
			"type":  bodyObj.ErrorType,
		})
		return nil, fmt.Errorf("error while fetching channel: %s", bodyObj.Error)
	}

	channel, err := parseChannelData(bodyObj.Data)
	if err != nil {
		// Treat parse errors from empty data as not-found
		tflog.Debug(ctx, "GetChannel: channel not found or empty data", map[string]any{"channelID": channelID, "error": err.Error()})
		return nil, nil
	}

	tflog.Debug(ctx, "GetChannel: channel fetched", map[string]any{"channelID": channel.ID})

	return channel, nil
}

// CreateChannel - Creates a new notification channel.
func (c *Client) CreateChannel(ctx context.Context, payload *model.NotificationChannel) (*model.NotificationChannel, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url, err := url.JoinPath(c.hostURL.String(), channelPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var bodyObj channelResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal create channel response: %w (body: %s)", err, truncateStr(string(body), 500))
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "CreateChannel: error while creating channel", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
		})
		return nil, fmt.Errorf("error while creating channel: %s", bodyObj.Error)
	}

	channel, err := parseChannelData(bodyObj.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created channel data: %w", err)
	}

	if channel.ID == "" {
		return nil, fmt.Errorf("created channel has empty ID (response: %s)", truncateStr(string(body), 500))
	}

	tflog.Debug(ctx, "CreateChannel: channel created", map[string]any{"channelID": channel.ID})

	return channel, nil
}

// UpdateChannel - Updates an existing notification channel.
func (c *Client) UpdateChannel(ctx context.Context, channelID string, payload *model.NotificationChannel) error {
	rb, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.hostURL.String(), channelPath, channelID)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}

	// SigNoz API may return an empty body on successful update (e.g. 204 No Content).
	if len(body) == 0 {
		tflog.Debug(ctx, "UpdateChannel: channel updated (empty response body)", map[string]any{"channelID": channelID})
		return nil
	}

	var bodyObj signozResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "UpdateChannel: error while updating channel", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while updating channel: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "UpdateChannel: channel updated", map[string]any{"channelID": channelID})

	return nil
}

// DeleteChannel - Deletes an existing notification channel.
func (c *Client) DeleteChannel(ctx context.Context, channelID string) error {
	url, err := url.JoinPath(c.hostURL.String(), channelPath, channelID)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}

	var bodyObj signozResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "DeleteChannel: error while deleting channel", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while deleting channel: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "DeleteChannel: channel deleted", map[string]any{"channelID": channelID})

	return nil
}
