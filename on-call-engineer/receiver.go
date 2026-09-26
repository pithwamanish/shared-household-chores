package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Canonical Normalized Incident Data Models (matching payload-schema.json)

type AlertInfo struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Severity string `json:"severity"`
}

type ServiceInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type DeploymentInfo struct {
	Environment string `json:"environment"`
}

type MetricInfo struct {
	Name          string  `json:"name"`
	ObservedValue float64 `json:"observed_value"`
	Threshold     float64 `json:"threshold"`
	Duration      string  `json:"duration,omitempty"`
}

type NormalizedIncident struct {
	Alert        AlertInfo      `json:"alert"`
	Service      ServiceInfo    `json:"service"`
	Deployment   DeploymentInfo `json:"deployment"`
	Metric       MetricInfo     `json:"metric"`
	DashboardURL string         `json:"dashboard_url"`
	RunbookURL   string         `json:"runbook_url"`
	Timestamp    string         `json:"timestamp"`
}

type WebhookResponse struct {
	Status      string   `json:"status"`
	Message     string   `json:"message"`
	IncidentIDs []string `json:"incident_ids"`
}

type CLIOutput struct {
	IncidentID string             `json:"incident_id"`
	Path       string             `json:"path"`
	Payload    NormalizedIncident `json:"payload"`
}

// Config resolves environment settings and directories
type Config struct {
	Port         int
	IncidentsDir string
}

func resolveConfig() Config {
	port := 5050
	if val := os.Getenv("RECEIVER_PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			port = p
		}
	}

	incidentsDir := os.Getenv("INCIDENTS_DIR")
	if incidentsDir == "" {
		// Default to ./incidents
		incidentsDir = filepath.Join(".", "incidents")
	}

	_ = os.MkdirAll(incidentsDir, 0755)

	return Config{
		Port:         port,
		IncidentsDir: incidentsDir,
	}
}

// Helper: safe string extraction from map
func getString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if val, ok := m[k]; ok && val != nil {
			if s, ok := val.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// Helper: safe float extraction from map
func getFloat(m map[string]interface{}, defaultVal float64, keys ...string) float64 {
	for _, k := range keys {
		if val, ok := m[k]; ok && val != nil {
			switch v := val.(type) {
			case float64:
				return v
			case float32:
				return float64(v)
			case int:
				return float64(v)
			case int64:
				return float64(v)
			case string:
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					return f
				}
			}
		}
	}
	return defaultVal
}

// NormalizeAlert normalizes raw Alertmanager payloads or direct alert objects into canonical schema
func NormalizeAlert(alertRaw map[string]interface{}, commonLabels, commonAnnotations map[string]interface{}) NormalizedIncident {
	labels := make(map[string]interface{})
	for k, v := range commonLabels {
		labels[k] = v
	}
	if rawLabels, ok := alertRaw["labels"].(map[string]interface{}); ok {
		for k, v := range rawLabels {
			labels[k] = v
		}
	}

	annotations := make(map[string]interface{})
	for k, v := range commonAnnotations {
		annotations[k] = v
	}
	if rawAnnotations, ok := alertRaw["annotations"].(map[string]interface{}); ok {
		for k, v := range rawAnnotations {
			annotations[k] = v
		}
	}

	// Alert block
	var alertMap map[string]interface{}
	if a, ok := alertRaw["alert"].(map[string]interface{}); ok {
		alertMap = a
	}

	alertName := getString(labels, "alertname", "alert")
	if alertName == "" && alertMap != nil {
		alertName = getString(alertMap, "name")
	}
	if alertName == "" {
		alertName = "UnknownAlert"
	}

	state := getString(alertRaw, "status", "state")
	if state == "" && alertMap != nil {
		state = getString(alertMap, "state")
	}
	if strings.ToLower(state) == "resolved" {
		state = "resolved"
	} else {
		state = "firing"
	}

	severity := getString(labels, "severity")
	if severity == "" && alertMap != nil {
		severity = getString(alertMap, "severity")
	}
	severity = strings.ToLower(severity)
	if severity != "critical" && severity != "warning" && severity != "info" {
		severity = "critical"
	}

	// Service block
	var serviceMap map[string]interface{}
	if s, ok := alertRaw["service"].(map[string]interface{}); ok {
		serviceMap = s
	}

	serviceName := getString(labels, "service", "service_name")
	if serviceName == "" && serviceMap != nil {
		serviceName = getString(serviceMap, "name")
	}
	if serviceName == "" {
		serviceName = "api-backend"
	}

	version := getString(labels, "version", "service_version")
	if version == "" && serviceMap != nil {
		version = getString(serviceMap, "version")
	}
	if version == "" {
		version = "dev-latest"
	}

	// Deployment block
	var deployMap map[string]interface{}
	if d, ok := alertRaw["deployment"].(map[string]interface{}); ok {
		deployMap = d
	}

	env := getString(labels, "environment", "deployment_environment")
	if env == "" && deployMap != nil {
		env = getString(deployMap, "environment")
	}
	env = strings.ToLower(env)
	if env != "development" && env != "staging" && env != "production" {
		env = "development"
	}

	// Dashboard & Runbook URLs
	dashURL := getString(annotations, "dashboard_url")
	if dashURL == "" {
		dashURL = getString(alertRaw, "dashboard_url")
	}
	if dashURL == "" {
		dashURL = fmt.Sprintf("http://localhost:3001/d/overview?var-environment=%s", env)
	}

	runbookURL := getString(annotations, "runbook_url")
	if runbookURL == "" {
		runbookURL = getString(alertRaw, "runbook_url")
	}
	if runbookURL == "" {
		runbookURL = "on-call-engineer/runbook.md"
	}

	// Metric block
	var metricMap map[string]interface{}
	if m, ok := alertRaw["metric"].(map[string]interface{}); ok {
		metricMap = m
	}

	metricName := getString(labels, "__name__")
	if metricName == "" {
		metricName = getString(annotations, "metric_name")
	}
	if metricName == "" && metricMap != nil {
		metricName = getString(metricMap, "name")
	}
	if metricName == "" {
		metricName = "http_requests_total"
	}

	observedVal := 5.2
	if metricMap != nil {
		observedVal = getFloat(metricMap, 5.2, "observed_value")
	}

	thresholdVal := 5.0
	if metricMap != nil {
		thresholdVal = getFloat(metricMap, 5.0, "threshold")
	}

	duration := getString(labels, "for")
	if duration == "" && metricMap != nil {
		duration = getString(metricMap, "duration")
	}
	if duration == "" {
		duration = "1m"
	}

	// Timestamp
	ts := getString(alertRaw, "startsAt", "timestamp")
	if ts == "" {
		ts = time.Now().UTC().Format(time.RFC3339)
	}

	return NormalizedIncident{
		Alert: AlertInfo{
			Name:     alertName,
			State:    state,
			Severity: severity,
		},
		Service: ServiceInfo{
			Name:    serviceName,
			Version: version,
		},
		Deployment: DeploymentInfo{
			Environment: env,
		},
		Metric: MetricInfo{
			Name:          metricName,
			ObservedValue: observedVal,
			Threshold:     thresholdVal,
			Duration:      duration,
		},
		DashboardURL: dashURL,
		RunbookURL:   runbookURL,
		Timestamp:    ts,
	}
}

// SaveIncident persists the normalized incident payload to incidents/
func SaveIncident(normalized NormalizedIncident, incidentsDir string) (string, string, error) {
	if err := os.MkdirAll(incidentsDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create incidents dir: %w", err)
	}

	timestampKey := time.Now().Unix()
	alertSlug := strings.ToLower(normalized.Alert.Name)
	alertSlug = strings.ReplaceAll(alertSlug, " ", "-")
	incidentID := fmt.Sprintf("inc-%d-%s", timestampKey, alertSlug)
	incidentPath := filepath.Join(incidentsDir, fmt.Sprintf("%s.json", incidentID))

	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal incident: %w", err)
	}

	if err := os.WriteFile(incidentPath, data, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write incident file: %w", err)
	}

	log.Printf("[ON-CALL RECEIVER] Incident recorded: ID=%s Alert=%s Status=%s Path=%s\n",
		incidentID, normalized.Alert.Name, normalized.Alert.State, incidentPath)

	return incidentID, incidentPath, nil
}

// LoadAllIncidents returns all recorded incidents sorted newest first
func LoadAllIncidents(incidentsDir string) ([]struct {
	Filename string
	Incident NormalizedIncident
}, error) {
	entries, err := os.ReadDir(incidentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var results []struct {
		Filename string
		Incident NormalizedIncident
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(incidentsDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		var inc NormalizedIncident
		if err := json.Unmarshal(data, &inc); err == nil {
			results = append(results, struct {
				Filename string
				Incident NormalizedIncident
			}{
				Filename: entry.Name(),
				Incident: inc,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Filename > results[j].Filename
	})

	return results, nil
}

// HTML Dashboard Template
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>ChoreSync On-Call Alert Receiver</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; margin: 0; padding: 24px; }
    .container { max-width: 960px; margin: 0 auto; }
    h1 { color: #38bdf8; margin-bottom: 4px; }
    .badge { background: #0284c7; color: white; padding: 4px 10px; border-radius: 9999px; font-size: 13px; }
    .card { background: #1e293b; border-radius: 8px; padding: 20px; margin-top: 20px; border: 1px solid #334155; }
    table { width: 100%; border-collapse: collapse; text-align: left; }
    th { padding: 10px; background: #0f172a; color: #94a3b8; font-size: 13px; border-bottom: 2px solid #334155; }
    td { padding: 10px; border-bottom: 1px solid #334155; }
    a { color: #38bdf8; text-decoration: none; }
    a:hover { text-decoration: underline; }
    .links { display: flex; gap: 16px; margin-top: 12px; }
    .firing { background: #dc2626; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
    .resolved { background: #16a34a; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
  </style>
</head>
<body>
  <div class="container">
    <div style="display:flex; justify-content:space-between; align-items:center;">
      <div>
        <h1>🚨 ChoreSync On-Call Alert Receiver (Go Engine)</h1>
        <p style="color:#94a3b8;margin-top:0;">Autonomous incident webhook endpoint listening on port 5050</p>
      </div>
      <div><span class="badge">● Online & Ready</span></div>
    </div>

    <div class="card">
      <h3 style="margin-top:0;color:#e2e8f0;">Quick Observability Navigation</h3>
      <div class="links">
        <a href="http://localhost:3001" target="_blank">📊 Grafana UI (3001)</a>
        <a href="http://localhost:9090" target="_blank">🔥 Prometheus Alerts & Metrics (9090)</a>
        <a href="http://localhost:9093" target="_blank">🔔 Alertmanager (9093)</a>
        <a href="/incidents" target="_blank">📄 Incidents JSON API</a>
        <a href="/healthz" target="_blank">❤️ Healthcheck</a>
      </div>
    </div>

    <div class="card">
      <h3 style="margin-top:0;color:#e2e8f0;">Recorded Incidents ({{len .Incidents}})</h3>
      <table>
        <thead>
          <tr>
            <th>Incident File</th>
            <th>Alert Name</th>
            <th>State</th>
            <th>Service / Version</th>
            <th>Timestamp</th>
          </tr>
        </thead>
        <tbody>
          {{range .Incidents}}
          <tr>
            <td style="font-family:monospace;font-size:13px;">{{.Filename}}</td>
            <td><strong>{{.Incident.Alert.Name}}</strong></td>
            <td>
              {{if eq .Incident.Alert.State "firing"}}
              <span class="firing">FIRING</span>
              {{else}}
              <span class="resolved">RESOLVED</span>
              {{end}}
            </td>
            <td>{{.Incident.Service.Name}} ({{.Incident.Service.Version}})</td>
            <td style="font-size:12px;color:#94a3b8;">{{.Incident.Timestamp}}</td>
          </tr>
          {{else}}
          <tr>
            <td colspan="5" style="padding:20px;text-align:center;color:#94a3b8;">No incidents recorded yet.</td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>
</body>
</html>`

func createRouter(cfg Config) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"on-call-receiver"}` + "\n"))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"on-call-receiver"}` + "\n"))
	})

	// Incidents API
	mux.HandleFunc("/incidents", func(w http.ResponseWriter, r *http.Request) {
		incidents, err := LoadAllIncidents(cfg.IncidentsDir)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"LOAD_FAILED","message":"%s"}`, err), http.StatusInternalServerError)
			return
		}
		var list []NormalizedIncident
		for _, item := range incidents {
			list = append(list, item.Incident)
		}
		if list == nil {
			list = []NormalizedIncident{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(list)
	})

	// Webhook receiver
	webhookHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"METHOD_NOT_ALLOWED"}`, http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error":"READ_BODY_FAILED"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var rawMap map[string]interface{}
		if err := json.Unmarshal(body, &rawMap); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "INVALID_JSON",
				"message": err.Error(),
			})
			return
		}

		var savedIDs []string

		// Check if Alertmanager envelope with alerts array
		if alertsList, ok := rawMap["alerts"].([]interface{}); ok && len(alertsList) > 0 {
			commonLabels, _ := rawMap["commonLabels"].(map[string]interface{})
			commonAnnotations, _ := rawMap["commonAnnotations"].(map[string]interface{})

			for _, aItem := range alertsList {
				if aMap, ok := aItem.(map[string]interface{}); ok {
					norm := NormalizeAlert(aMap, commonLabels, commonAnnotations)
					id, _, err := SaveIncident(norm, cfg.IncidentsDir)
					if err == nil {
						savedIDs = append(savedIDs, id)
					}
				}
			}
		} else {
			// Direct single alert or normalized payload
			norm := NormalizeAlert(rawMap, nil, nil)
			id, _, err := SaveIncident(norm, cfg.IncidentsDir)
			if err == nil {
				savedIDs = append(savedIDs, id)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(WebhookResponse{
			Status:      "success",
			Message:     "Alerts received and normalized",
			IncidentIDs: savedIDs,
		})
	}

	mux.HandleFunc("/webhook", webhookHandler)
	mux.HandleFunc("/alert", webhookHandler)
	mux.HandleFunc("/api/v1/alerts", webhookHandler)

	// HTML Dashboard
	tmpl := template.Must(template.New("dashboard").Parse(dashboardHTML))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			http.NotFound(w, r)
			return
		}
		incidents, _ := LoadAllIncidents(cfg.IncidentsDir)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, struct {
			Incidents []struct {
				Filename string
				Incident NormalizedIncident
			}
		}{
			Incidents: incidents,
		})
	})

	return mux
}

func main() {
	cfg := resolveConfig()

	// CLI mode: check for --stdin
	if len(os.Args) > 1 && os.Args[1] == "--stdin" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
		if len(strings.TrimSpace(string(data))) > 0 {
			var rawMap map[string]interface{}
			if err := json.Unmarshal(data, &rawMap); err != nil {
				log.Fatalf("Invalid JSON on stdin: %v", err)
			}
			norm := NormalizeAlert(rawMap, nil, nil)
			incID, incPath, err := SaveIncident(norm, cfg.IncidentsDir)
			if err != nil {
				log.Fatalf("Failed to save incident: %v", err)
			}
			out, _ := json.MarshalIndent(CLIOutput{
				IncidentID: incID,
				Path:       incPath,
				Payload:    norm,
			}, "", "  ")
			fmt.Println(string(out))
			return
		}
	}

	router := createRouter(cfg)
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	log.Printf("[ON-CALL RECEIVER] Listening for alerts on http://%s/webhook (Go Engine)\n", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("[ON-CALL RECEIVER] Server failed: %v", err)
	}
}
