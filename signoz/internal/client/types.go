package client

import (
	"encoding/json"
	"fmt"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
)

// signozResponse - Maps the response data.
type signozResponse struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data"`
	ErrorType string      `json:"errorType"`
	Error     string      `json:"error"`
}

// alertResponse - Maps the response data of GetAlert and CreateAlert.
// Uses json.RawMessage for Data to handle potential format variations
// across SigNoz versions.
type alertResponse struct {
	Status    string          `json:"status"`
	Error     string          `json:"error"`
	ErrorType string          `json:"errorType"`
	Data      json.RawMessage `json:"data"`
}

// parseAlertData parses the alert data from the response, handling both
// a single object and an array (taking the first element).
func parseAlertData(raw json.RawMessage) (*model.Alert, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, fmt.Errorf("alert data is empty or null")
	}

	// Try single object first.
	var alert model.Alert
	if err := json.Unmarshal(raw, &alert); err == nil {
		return &alert, nil
	}

	// Try array of alerts (take the first element).
	var alerts []model.Alert
	if err := json.Unmarshal(raw, &alerts); err == nil {
		if len(alerts) == 0 {
			return nil, fmt.Errorf("alert data array is empty")
		}
		return &alerts[0], nil
	}

	return nil, fmt.Errorf("failed to parse alert data: unexpected format: %s", string(raw))
}

// dashboardResponse - Maps the response data of CreateDashboard and GetDashboard.
// Uses json.RawMessage for Data to handle both single-object and array responses.
// SigNoz >= 0.104.0 may return data as an array even for single-dashboard endpoints.
type dashboardResponse struct {
	Status    string          `json:"status"`
	Error     string          `json:"error,omitempty"`
	ErrorType string          `json:"errorType,omitempty"`
	Data      json.RawMessage `json:"data"`
}

// parseDashboardData parses the dashboard data from the response, handling both
// a single object and an array (taking the first element).
func parseDashboardData(raw json.RawMessage) (*dashboardData, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, fmt.Errorf("dashboard data is empty or null")
	}

	// Try single object first.
	var single dashboardData
	if err := json.Unmarshal(raw, &single); err == nil && single.ID != "" {
		return &single, nil
	}

	// Try array of dashboards (take the first element).
	var arr []dashboardData
	if err := json.Unmarshal(raw, &arr); err == nil {
		if len(arr) == 0 {
			return nil, fmt.Errorf("dashboard data array is empty")
		}
		return &arr[0], nil
	}

	return nil, fmt.Errorf("failed to parse dashboard data: unexpected format: %s", truncateStr(string(raw), 200))
}

type dashboardData struct {
	CreatedAt string          `json:"createdAt"`
	CreatedBy string          `json:"createdBy"`
	ID        string          `json:"id"`
	Locked    bool            `json:"locked"`
	UpdatedAt string          `json:"updatedAt"`
	UpdatedBy string          `json:"updatedBy"`
	Data      model.Dashboard `json:"data"`
}

// truncateStr truncates a string to maxLen characters for safe logging.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
