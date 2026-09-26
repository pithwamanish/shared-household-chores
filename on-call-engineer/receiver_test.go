package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestPayloadSchemaSyntax(t *testing.T) {
	data, err := os.ReadFile("payload-schema.json")
	if err != nil {
		t.Fatalf("Failed to read payload-schema.json: %v", err)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("Failed to parse payload-schema.json: %v", err)
	}

	if schema["title"] != "NormalizedAlertIncidentPayload" {
		t.Errorf("Expected title 'NormalizedAlertIncidentPayload', got %v", schema["title"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing or invalid properties in schema")
	}

	for _, req := range []string{"alert", "service", "deployment", "metric", "dashboard_url", "runbook_url", "timestamp"} {
		if _, exists := props[req]; !exists {
			t.Errorf("Missing required property in schema: %s", req)
		}
	}
}

func TestAutonomyPolicyStructure(t *testing.T) {
	data, err := os.ReadFile("autonomy-policy.json")
	if err != nil {
		t.Fatalf("Failed to read autonomy-policy.json: %v", err)
	}

	var policy map[string]interface{}
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatalf("Failed to parse autonomy-policy.json: %v", err)
	}

	if policy["title"] != "ChoreSync On-Call Autonomy Policy" {
		t.Errorf("Expected title 'ChoreSync On-Call Autonomy Policy', got %v", policy["title"])
	}

	levels, ok := policy["autonomy_levels"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing autonomy_levels")
	}
	if _, exists := levels["LEVEL_1_AUTO_EXECUTE"]; !exists {
		t.Errorf("Missing LEVEL_1_AUTO_EXECUTE in policy")
	}

	if _, exists := policy["forbidden_patterns"]; !exists {
		t.Errorf("Missing forbidden_patterns in policy")
	}
}

func TestNormalizeAlert(t *testing.T) {
	rawAlert := map[string]interface{}{
		"status": "firing",
		"labels": map[string]interface{}{
			"alertname":   "HighHttpErrorRate",
			"severity":    "critical",
			"service":     "api-backend",
			"environment": "development",
			"version":     "dev-latest",
			"for":         "1m",
		},
		"annotations": map[string]interface{}{
			"dashboard_url": "http://localhost:3001/d/overview?var-environment=development",
			"runbook_url":   "on-call-engineer/runbook.md#high-http-error-rate",
			"metric_name":   "choresync_http_requests_total",
		},
	}

	norm := NormalizeAlert(rawAlert, nil, nil)

	if norm.Alert.Name != "HighHttpErrorRate" {
		t.Errorf("Expected alert name HighHttpErrorRate, got %s", norm.Alert.Name)
	}
	if norm.Alert.State != "firing" {
		t.Errorf("Expected alert state firing, got %s", norm.Alert.State)
	}
	if norm.Alert.Severity != "critical" {
		t.Errorf("Expected alert severity critical, got %s", norm.Alert.Severity)
	}
	if norm.Service.Name != "api-backend" {
		t.Errorf("Expected service name api-backend, got %s", norm.Service.Name)
	}
	if norm.Deployment.Environment != "development" {
		t.Errorf("Expected environment development, got %s", norm.Deployment.Environment)
	}
	if norm.Metric.Name != "choresync_http_requests_total" {
		t.Errorf("Expected metric name choresync_http_requests_total, got %s", norm.Metric.Name)
	}
	if norm.Metric.Duration != "1m" {
		t.Errorf("Expected duration 1m, got %s", norm.Metric.Duration)
	}
}

func TestSaveAndLoadIncident(t *testing.T) {
	tempDir := t.TempDir()

	incident := NormalizedIncident{
		Alert: AlertInfo{
			Name:     "TestAlert",
			State:    "firing",
			Severity: "warning",
		},
		Service: ServiceInfo{
			Name:    "api-backend",
			Version: "1.0.0",
		},
		Deployment: DeploymentInfo{
			Environment: "staging",
		},
		Metric: MetricInfo{
			Name:          "cpu_usage",
			ObservedValue: 88.5,
			Threshold:     80.0,
			Duration:      "5m",
		},
		DashboardURL: "http://example.com/dash",
		RunbookURL:   "http://example.com/runbook",
		Timestamp:    "2026-09-26T14:00:00Z",
	}

	id, path, err := SaveIncident(incident, tempDir)
	if err != nil {
		t.Fatalf("SaveIncident failed: %v", err)
	}
	if id == "" || path == "" {
		t.Errorf("Invalid id or path returned")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Incident file does not exist at %s", path)
	}

	loaded, err := LoadAllIncidents(tempDir)
	if err != nil {
		t.Fatalf("LoadAllIncidents failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("Expected 1 incident loaded, got %d", len(loaded))
	}
	if loaded[0].Incident.Alert.Name != "TestAlert" {
		t.Errorf("Loaded alert name mismatch: %s", loaded[0].Incident.Alert.Name)
	}
}

func TestHTTPHandlers(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Port:         5050,
		IncidentsDir: tempDir,
	}
	router := createRouter(cfg)

	// 1. Test /healthz
	reqHealth := httptest.NewRequest("GET", "/healthz", nil)
	recHealth := httptest.NewRecorder()
	router.ServeHTTP(recHealth, reqHealth)
	if recHealth.Code != http.StatusOK {
		t.Errorf("/healthz returned code %d", recHealth.Code)
	}

	// 2. Test POST /webhook with Alertmanager envelope
	webhookPayload := map[string]interface{}{
		"status": "firing",
		"alerts": []interface{}{
			map[string]interface{}{
				"status": "firing",
				"labels": map[string]interface{}{
					"alertname":   "DatabaseQueryErrors",
					"severity":    "critical",
					"service":     "api-backend",
					"environment": "development",
				},
				"annotations": map[string]interface{}{
					"runbook_url": "on-call-engineer/runbook.md#database-query-errors",
				},
			},
		},
	}
	body, _ := json.Marshal(webhookPayload)
	reqHook := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	recHook := httptest.NewRecorder()
	router.ServeHTTP(recHook, reqHook)
	if recHook.Code != http.StatusOK {
		t.Errorf("POST /webhook returned code %d: %s", recHook.Code, recHook.Body.String())
	}

	var resp WebhookResponse
	if err := json.Unmarshal(recHook.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse webhook response: %v", err)
	}
	if resp.Status != "success" || len(resp.IncidentIDs) != 1 {
		t.Errorf("Unexpected webhook response: %+v", resp)
	}

	// 3. Test GET /incidents
	reqInc := httptest.NewRequest("GET", "/incidents", nil)
	recInc := httptest.NewRecorder()
	router.ServeHTTP(recInc, reqInc)
	if recInc.Code != http.StatusOK {
		t.Errorf("GET /incidents returned code %d", recInc.Code)
	}
	var incidents []NormalizedIncident
	if err := json.Unmarshal(recInc.Body.Bytes(), &incidents); err != nil {
		t.Fatalf("Failed to parse /incidents JSON: %v", err)
	}
	if len(incidents) != 1 || incidents[0].Alert.Name != "DatabaseQueryErrors" {
		t.Errorf("Unexpected incidents returned: %+v", incidents)
	}

	// 4. Test GET / (HTML Dashboard)
	reqDash := httptest.NewRequest("GET", "/", nil)
	recDash := httptest.NewRecorder()
	router.ServeHTTP(recDash, reqDash)
	if recDash.Code != http.StatusOK {
		t.Errorf("GET / returned code %d", recDash.Code)
	}
	if !bytes.Contains(recDash.Body.Bytes(), []byte("DatabaseQueryErrors")) {
		t.Errorf("Dashboard HTML does not contain DatabaseQueryErrors")
	}
}
