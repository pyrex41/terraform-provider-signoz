package client

import (
	"encoding/json"
	"testing"
)

func TestParseDashboardData_SingleObject(t *testing.T) {
	raw := json.RawMessage(`{
		"createdAt": "2025-01-01T00:00:00Z",
		"createdBy": "user@example.com",
		"id": "abc-123",
		"locked": false,
		"updatedAt": "2025-01-02T00:00:00Z",
		"updatedBy": "user@example.com",
		"data": {
			"title": "My Dashboard",
			"description": "Test dashboard",
			"tags": ["test"],
			"name": "my-dashboard",
			"widgets": [],
			"layout": [],
			"variables": {},
			"version": "v1"
		}
	}`)

	result, err := parseDashboardData(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "abc-123" {
		t.Errorf("expected ID 'abc-123', got '%s'", result.ID)
	}
	if result.CreatedBy != "user@example.com" {
		t.Errorf("expected CreatedBy 'user@example.com', got '%s'", result.CreatedBy)
	}
	if result.Data.Title != "My Dashboard" {
		t.Errorf("expected title 'My Dashboard', got '%s'", result.Data.Title)
	}
}

func TestParseDashboardData_Array(t *testing.T) {
	raw := json.RawMessage(`[{
		"createdAt": "2025-01-01T00:00:00Z",
		"createdBy": "user@example.com",
		"id": "abc-456",
		"locked": true,
		"updatedAt": "2025-01-02T00:00:00Z",
		"updatedBy": "admin@example.com",
		"data": {
			"title": "Array Dashboard",
			"description": "Test",
			"tags": [],
			"name": "array-dashboard",
			"widgets": [],
			"layout": [],
			"variables": {},
			"version": "v2"
		}
	}]`)

	result, err := parseDashboardData(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "abc-456" {
		t.Errorf("expected ID 'abc-456', got '%s'", result.ID)
	}
	if !result.Locked {
		t.Error("expected Locked to be true")
	}
	if result.Data.Title != "Array Dashboard" {
		t.Errorf("expected title 'Array Dashboard', got '%s'", result.Data.Title)
	}
}

func TestParseDashboardData_EmptyArray(t *testing.T) {
	raw := json.RawMessage(`[]`)

	_, err := parseDashboardData(raw)
	if err == nil {
		t.Fatal("expected error for empty array, got nil")
	}
}

func TestParseDashboardData_Null(t *testing.T) {
	raw := json.RawMessage(`null`)

	_, err := parseDashboardData(raw)
	if err == nil {
		t.Fatal("expected error for null, got nil")
	}
}

func TestParseAlertData_SingleObject(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "rule-789",
		"alert": "Test Alert",
		"alertType": "METRIC_BASED_ALERT",
		"annotations": {"description": "test desc", "summary": "test summary"},
		"condition": {"op": "1", "target": 90},
		"disabled": false,
		"evalWindow": "5m0s",
		"frequency": "1m0s",
		"labels": {"severity": "warning"},
		"preferredChannels": [],
		"ruleType": "threshold_rule",
		"source": "http://localhost:3301/alerts",
		"state": "inactive",
		"version": "v4"
	}`)

	result, err := parseAlertData(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "rule-789" {
		t.Errorf("expected ID 'rule-789', got '%s'", result.ID)
	}
	if result.Alert != "Test Alert" {
		t.Errorf("expected Alert 'Test Alert', got '%s'", result.Alert)
	}
	if result.Labels["severity"] != "warning" {
		t.Errorf("expected severity label 'warning', got '%s'", result.Labels["severity"])
	}
}

func TestParseAlertData_Array(t *testing.T) {
	raw := json.RawMessage(`[{
		"id": "rule-array-1",
		"alert": "Array Alert",
		"alertType": "LOGS_BASED_ALERT",
		"annotations": {"description": "desc", "summary": "sum"},
		"condition": {},
		"evalWindow": "5m0s",
		"frequency": "1m0s",
		"labels": {},
		"ruleType": "threshold_rule",
		"version": "v5"
	}]`)

	result, err := parseAlertData(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "rule-array-1" {
		t.Errorf("expected ID 'rule-array-1', got '%s'", result.ID)
	}
	if result.Alert != "Array Alert" {
		t.Errorf("expected Alert 'Array Alert', got '%s'", result.Alert)
	}
}

func TestParseAlertData_WithSchemaV2Fields(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "rule-v2",
		"alert": "V2 Alert",
		"alertType": "METRIC_BASED_ALERT",
		"annotations": {"description": "desc", "summary": "sum"},
		"condition": {},
		"evalWindow": "5m0s",
		"frequency": "1m0s",
		"labels": {"severity": "critical"},
		"ruleType": "threshold_rule",
		"version": "v5",
		"schemaVersion": "v2",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m0s"}},
		"notificationSettings": {"groupBy": [], "renotify": {"enabled": true, "interval": "30m0s", "alertStates": ["firing"]}}
	}`)

	result, err := parseAlertData(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "rule-v2" {
		t.Errorf("expected ID 'rule-v2', got '%s'", result.ID)
	}
	if result.SchemaVersion != "v2" {
		t.Errorf("expected SchemaVersion 'v2', got '%s'", result.SchemaVersion)
	}
	if result.Evaluation == nil {
		t.Error("expected non-nil Evaluation")
	}
}

func TestParseAlertData_NullTimestamps(t *testing.T) {
	// SigNoz CreateRule response may have null timestamps.
	raw := json.RawMessage(`{
		"id": "rule-new",
		"alert": "New Alert",
		"alertType": "METRIC_BASED_ALERT",
		"annotations": {"description": "desc", "summary": "sum"},
		"condition": {},
		"evalWindow": "5m0s",
		"frequency": "1m0s",
		"labels": {},
		"ruleType": "threshold_rule",
		"version": "v4",
		"state": "",
		"createAt": null,
		"createBy": null,
		"updateAt": null,
		"updateBy": null
	}`)

	result, err := parseAlertData(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "rule-new" {
		t.Errorf("expected ID 'rule-new', got '%s'", result.ID)
	}
	if result.CreateAt != "" {
		t.Errorf("expected empty CreateAt for null, got '%s'", result.CreateAt)
	}
}

func TestParseAlertData_Null(t *testing.T) {
	raw := json.RawMessage(`null`)

	_, err := parseAlertData(raw)
	if err == nil {
		t.Fatal("expected error for null, got nil")
	}
}

func TestParseAlertData_EmptyArray(t *testing.T) {
	raw := json.RawMessage(`[]`)

	_, err := parseAlertData(raw)
	if err == nil {
		t.Fatal("expected error for empty array, got nil")
	}
}

func TestFullDashboardResponse_ArrayFormat(t *testing.T) {
	// Simulate the actual SigNoz 0.104.0 response format where data is an array.
	responseBody := `{
		"status": "success",
		"data": [{
			"id": "dash-001",
			"createdAt": "2025-06-01T10:00:00Z",
			"createdBy": "admin@test.com",
			"updatedAt": "2025-06-01T12:00:00Z",
			"updatedBy": "admin@test.com",
			"locked": false,
			"data": {
				"title": "K8s Dashboard",
				"description": "Kubernetes metrics",
				"name": "k8s",
				"tags": ["kubernetes"],
				"widgets": [],
				"layout": [],
				"variables": {},
				"version": "v1"
			}
		}]
	}`

	var resp dashboardResponse
	err := json.Unmarshal([]byte(responseBody), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != "success" {
		t.Fatalf("expected status 'success', got '%s'", resp.Status)
	}

	dashboard, err := parseDashboardData(resp.Data)
	if err != nil {
		t.Fatalf("failed to parse dashboard data: %v", err)
	}

	if dashboard.ID != "dash-001" {
		t.Errorf("expected ID 'dash-001', got '%s'", dashboard.ID)
	}
	if dashboard.Data.Title != "K8s Dashboard" {
		t.Errorf("expected title 'K8s Dashboard', got '%s'", dashboard.Data.Title)
	}
}

func TestFullAlertResponse_ObjectFormat(t *testing.T) {
	// Simulate the actual SigNoz response for GET /api/v1/rules/{id}.
	responseBody := `{
		"status": "success",
		"data": {
			"id": "01957e2e-a0c4-7d63-9c3c-1e6a4e8f1234",
			"state": "inactive",
			"alert": "CPU Alert",
			"alertType": "METRIC_BASED_ALERT",
			"condition": {"compositeQuery": {"queryType": "builder"}, "op": "1", "target": 80},
			"labels": {"severity": "warning", "managedBy": "terraform"},
			"annotations": {"description": "CPU high", "summary": "CPU usage alert"},
			"disabled": false,
			"evalWindow": "5m0s",
			"frequency": "1m0s",
			"ruleType": "threshold_rule",
			"preferredChannels": ["slack"],
			"source": "http://signoz:3301/alerts",
			"version": "v4",
			"createAt": "2025-06-01T10:00:00Z",
			"createBy": "admin@test.com",
			"updateAt": "2025-06-01T10:00:00Z",
			"updateBy": "admin@test.com"
		}
	}`

	var resp alertResponse
	err := json.Unmarshal([]byte(responseBody), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != "success" {
		t.Fatalf("expected status 'success', got '%s'", resp.Status)
	}

	alert, err := parseAlertData(resp.Data)
	if err != nil {
		t.Fatalf("failed to parse alert data: %v", err)
	}

	if alert.ID != "01957e2e-a0c4-7d63-9c3c-1e6a4e8f1234" {
		t.Errorf("expected UUID ID, got '%s'", alert.ID)
	}
	if alert.Alert != "CPU Alert" {
		t.Errorf("expected alert name 'CPU Alert', got '%s'", alert.Alert)
	}
	if alert.Annotations.Description != "CPU high" {
		t.Errorf("expected description 'CPU high', got '%s'", alert.Annotations.Description)
	}
}
