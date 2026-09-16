# ChoreSync Incident Triage Runbook (`on-call-engineer/runbook.md`)

This runbook provides deterministic diagnosis, reproduction, and remediation procedures for alerts fired by Prometheus and OpenTelemetry across the ChoreSync infrastructure.

---

## 1. HighHttpErrorRate

### Symptoms
- **Alert**: `HighHttpErrorRate`
- **Severity**: `critical`
- **Threshold**: HTTP 5xx error rate > 5% for sustained 30s–1m window.
- **Metric**: `choresync_http_requests_total{http_status_code=~"5.."}` / `choresync_http_requests_total`
- **User Impact**: Users experience failed requests, red error toasts, and broken state updates.

### Triage & Diagnostics
1. **Inspect Backend Application Logs**:
   ```bash
   docker compose logs backend --tail 100
   ```
   Look for panics, unhandled nil pointers, or `500 Internal Server Error` logged by Chi router.
2. **Inspect OpenTelemetry Error Spans in Tempo**:
   - Query Tempo UI at `http://localhost:3200` or Grafana Explore for spans with `status.code = ERROR` and attribute `http.status_code >= 500`.
   - Identify offending route (e.g. `/api/v1/chores/*`, `/api/v1/auth/*`).
3. **Verify Error Spike in Prometheus**:
   ```bash
   curl -s "http://localhost:9090/api/v1/query?query=sum(rate(choresync_http_requests_total{http_status_code=~\"5..\"}[1m]))"
   ```

### Reproduction Procedure
1. Locate the endpoint indicated by the error logs/spans.
2. Create or extend a Go test case under `backend/tests/` reproducing the exact payload and headers causing the 5xx response.
3. Run the test in the container to prove consistent failure:
   ```bash
   docker run --rm -v $(pwd)/backend:/app -w /app golang:1.22-alpine go test -v -run "<ReproTestName>" ./tests/...
   ```

### Remediation & Verification
1. Implement minimal nil checks, defensive parsing, or correct error-status mapping (e.g. returning 400 Bad Request instead of unhandled 500 for bad user input).
2. Re-run container test suite:
   ```bash
   docker run --rm -v $(pwd)/backend:/app -w /app golang:1.22-alpine go test -v ./...
   ```
3. Verify metric drops to 0 and Prometheus alert clears:
   ```bash
   curl -s "http://localhost:9090/api/v1/alerts" | grep -q "HighHttpErrorRate" && echo "Alert still firing" || echo "Alert cleared"
   ```

---

## 2. HighRequestLatency

### Symptoms
- **Alert**: `HighRequestLatency`
- **Severity**: `warning`
- **Threshold**: P95 request latency > 1.0s for sustained 1m window.
- **Metric**: `choresync_http_server_request_duration_seconds_bucket`
- **User Impact**: Slow UI load times, spinning loaders, sluggish task completion interaction.

### Triage & Diagnostics
1. **Query P95 Latency by Route in Prometheus**:
   ```bash
   curl -s "http://localhost:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(choresync_http_server_request_duration_seconds_bucket[1m]))by(le,http_route))"
   ```
2. **Inspect Distributed Trace in Grafana / Tempo**:
   - Filter Tempo traces with `duration > 1s`.
   - Trace flame graph: Determine if time is spent in backend CPU loop, database locks, or external network call (e.g. SMTP/Resend timeout).

### Reproduction & Remediation
1. Write a targeted benchmark or unit test simulating large payloads or sluggish query execution.
2. Address root cause: Add missing database index in `backend/internal/db/schema.sql`, paginate oversized collections, or wrap third-party email dispatches in non-blocking goroutines.
3. Verify latency drops below threshold.

---

## 3. DatabaseQueryErrors

### Symptoms
- **Alert**: `DatabaseQueryErrors`
- **Severity**: `critical`
- **Threshold**: Database error rate > 0.05 errors/sec for 1m window.
- **Metric**: `choresync_db_query_errors_total`
- **User Impact**: Persistence failures, chore completions not saved, transactions aborted.

### Triage & Diagnostics
1. **Check Postgres Container Health**:
   ```bash
   docker compose ps postgres
   docker compose logs postgres --tail 50
   ```
2. **Inspect DB Spans with Errors**:
   - Look for `postgresql.query` spans in Tempo with `status.code = ERROR`.
   - Check error messages (e.g. foreign key constraint violation, unique key collision, connection pool exhaustion).

### Remediation & Verification
1. Resolve migration inconsistency, connection timeout, or invalid SQL query generation.
2. Run database integration test suite:
   ```bash
   docker compose run --rm backend go test -v ./tests -run "TestPostgres"
   ```
3. Confirm alert resolves in Prometheus.
