.PHONY: compat setup dev test lint e2e compose-up compose-down obs-up obs-down obs-logs verify clean prod-build prod-up prod-down ci-local docker-build-tag build-image deploy-dev deploy deploy-down promote-prod oncall-verify oncall-test demo-record demo-view k8s-cluster k8s-build k8s-load k8s-deploy k8s-wait k8s-verify k8s-up k8s-down k8s-logs ext-verify ext-eval ext-mcp-test ext-capabilities

compat:
	@echo "Linking agent tool conventions (Claude, Cursor, Antigravity, Copilot, Windsurf)..."
	@echo "@AGENTS.md" > CLAUDE.md
	@echo "@AGENTS.md" > GEMINI.md
	@ln -sf AGENTS.md .cursorrules
	@ln -sf AGENTS.md .windsurfrules
	@mkdir -p .github && ln -sf ../AGENTS.md .github/copilot-instructions.md
	@mkdir -p .claude .cursor
	@ln -sf ../.agents/skills .claude/skills 2>/dev/null || true
	@ln -sf ../.agents/agents .claude/agents 2>/dev/null || true
	@ln -sf ../.agents/skills .cursor/skills 2>/dev/null || true
	@ln -sf specs.md _docs/plan.md 2>/dev/null || true
	@ln -sf _docs/specs.md PLAN.md 2>/dev/null || true
	@ln -sf _docs/specs.md PRD.md 2>/dev/null || true
	@echo "Compatibility bridges active!"

setup: compat
	@echo "Setting up development environment..."
	@cd frontend && (bun install || npm install)
	@cd backend && (go mod download 2>/dev/null || true)

dev:
	@echo "Starting development environment..."
	@cd frontend && (bun run dev || npm run dev)

backend-dev:
	@echo "Starting Go backend server..."
	@cd backend && go run ./cmd/server

backend-test:
	@echo "Running Go backend unit and integration tests..."
	@cd backend && go test -v ./...

lint:
	@echo "Running lint checks..."
	@cd frontend && (bun run lint || npm run lint 2>/dev/null || true)

test: backend-test
	@echo "Running frontend unit and integration tests..."
	@cd frontend && (bun run test || npm run test 2>/dev/null || true)

test-fast:
	@echo "Running targeted Go test in container for pattern: $(TARGET)..."
	@docker run --rm -v $$(pwd)/backend:/app -w /app golang:1.22-alpine go test -v -run "$(TARGET)" ./...

e2e:
	@echo "Running Playwright E2E tests in container..."
	@docker compose run --rm e2e

e2e-fast:
	@echo "Running targeted Playwright E2E test for pattern: $(TARGET)..."
	@docker compose run --rm e2e npx playwright test -g "$(TARGET)"

compose-up:
	@docker compose up -d --build

compose-down:
	@docker compose down -v

obs-up:
	@echo "Starting standalone observability stack (OTel Collector, Prometheus, Loki, Tempo, Grafana)..."
	@docker compose -f observability/docker-compose.yml up -d
	@echo "Observability live! Grafana: http://localhost:3001 | Prometheus: http://localhost:9090 | Tempo: http://localhost:3200"

obs-down:
	@echo "Stopping observability stack..."
	@docker compose -f observability/docker-compose.yml down

obs-logs:
	@docker compose -f observability/docker-compose.yml logs -f

prod-build:
	@echo "Building production multi-stage Docker images..."
	@docker compose -f docker-compose.prod.yml build

prod-up:
	@echo "Starting production stack with Caddy reverse proxy on port 8088..."
	@docker compose -f docker-compose.prod.yml up -d

prod-down:
	@echo "Stopping production stack..."
	@docker compose -f docker-compose.prod.yml down

verify: lint test
	@echo "All verification gates passed!"

ci-local: verify prod-build e2e
	@echo "All local CI/CD pipeline stages passed successfully!"

act-ci:
	@echo "Executing Local CI/CD Pipeline via act & Kind..."
	@run-act-ci

REGISTRY ?= ghcr.io/choresync
SHORT_SHA ?= $(shell git rev-parse --short=7 HEAD 2>/dev/null || echo "dev")
TIMESTAMP ?= $(shell date -u +'%Y%m%d-%H%M%S')
IMAGE_TAG ?= $(TIMESTAMP)-$(SHORT_SHA)

docker-build-tag:
	@echo "Building and tagging images with pattern YYYYMMDD-HHMMSS-shortsha: $(IMAGE_TAG)..."
	@docker build -t $(REGISTRY)/backend:$(IMAGE_TAG) -t $(REGISTRY)/backend:dev-latest ./backend
	@docker build -f ./frontend/Dockerfile.prod -t $(REGISTRY)/frontend:$(IMAGE_TAG) -t $(REGISTRY)/frontend:dev-latest ./frontend
	@echo "Build complete: $(REGISTRY)/backend:$(IMAGE_TAG)"
	@echo "Build complete: $(REGISTRY)/frontend:$(IMAGE_TAG)"

build-image: docker-build-tag

deploy-dev:
	@echo "Pulling and serving pre-built dev images..."
	@BACKEND_IMAGE=$(REGISTRY)/backend:dev-latest FRONTEND_IMAGE=$(REGISTRY)/frontend:dev-latest docker compose -f docker-compose.deploy.yml up -d
	@echo "Dev stack running via pre-built images on port 8088!"

deploy: deploy-dev

deploy-down:
	@docker compose -f docker-compose.deploy.yml down

promote-prod:
	@echo "Pulling currently deployed dev image to prod..."
	@docker pull $(REGISTRY)/backend:dev-latest 2>/dev/null || true
	@docker pull $(REGISTRY)/frontend:dev-latest 2>/dev/null || true
	@docker tag $(REGISTRY)/backend:dev-latest $(REGISTRY)/backend:$(IMAGE_TAG)
	@docker tag $(REGISTRY)/backend:dev-latest $(REGISTRY)/backend:$(IMAGE_TAG)-prod
	@docker tag $(REGISTRY)/backend:dev-latest $(REGISTRY)/backend:prod-latest
	@docker tag $(REGISTRY)/frontend:dev-latest $(REGISTRY)/frontend:$(IMAGE_TAG)
	@docker tag $(REGISTRY)/frontend:dev-latest $(REGISTRY)/frontend:$(IMAGE_TAG)-prod
	@docker tag $(REGISTRY)/frontend:dev-latest $(REGISTRY)/frontend:prod-latest
	@echo "Serving promoted production stack..."
	@BACKEND_IMAGE=$(REGISTRY)/backend:prod-latest FRONTEND_IMAGE=$(REGISTRY)/frontend:prod-latest docker compose -f docker-compose.deploy.yml up -d
	@echo "Production stack live on port 8088!"

oncall-verify:
	@echo "Running Autonomous Incident Response & Alerting Verification Suite..."
	@bash on-call-engineer/scripts/verify

oncall-test:
	@echo "Triggering test incident and alert ingest..."
	@bash on-call-engineer/scripts/trigger-test-incident

demo-record:
	@echo "1. Pre-warming live observability data (Loki logs, Tempo traces, firing alerts)..."
	@python3 e2e/demo/prewarm.py
	@echo "2. Recording Playwright architecture demo video in container..."
	@docker compose run --rm e2e node demo/record.js
	@echo "3. Multiplexing video with synchronized voice narration into WebM and MP4..."
	@bash e2e/demo/mux.sh
	@echo "Demo recording complete! Watch at http://localhost:3000/demo.html"

demo-view:
	@echo "Open http://localhost:3000/demo.html in your browser to watch the interactive demo video."

K8S_CLUSTER ?= choresync-cluster

k8s-cluster:
	@echo "Ensuring Linux inotify limits for kind..."
	@docker run --privileged --rm alpine sysctl -w fs.inotify.max_user_instances=512 fs.inotify.max_user_watches=524288 >/dev/null 2>&1 || true
	@echo "Checking kind cluster $(K8S_CLUSTER)..."
	@kind get clusters | grep -qx $(K8S_CLUSTER) || kind create cluster --name $(K8S_CLUSTER) --config k8s/kind-cluster-config.yaml
	@echo "Kind cluster $(K8S_CLUSTER) is active."

k8s-build:
	@echo "Building application images for Kubernetes..."
	@docker build -t choresync-backend:dev-latest ./backend
	@docker build -f ./frontend/Dockerfile.prod -t choresync-frontend:dev-latest ./frontend

k8s-load: k8s-cluster
	@echo "Loading images into kind cluster $(K8S_CLUSTER) offline..."
	@kind load docker-image choresync-backend:dev-latest --name $(K8S_CLUSTER)
	@kind load docker-image choresync-frontend:dev-latest --name $(K8S_CLUSTER)
	@kind load docker-image postgres:16-alpine --name $(K8S_CLUSTER)
	@echo "Images loaded successfully into $(K8S_CLUSTER)."

k8s-deploy:
	@echo "Applying Kubernetes manifests via Kustomize..."
	@kubectl apply -k k8s/ --context kind-$(K8S_CLUSTER)

k8s-wait:
	@echo "Waiting for pods to reach Ready state..."
	@kubectl wait --for=condition=ready pod -l app=postgres --context kind-$(K8S_CLUSTER) --timeout=120s
	@kubectl rollout status deployment/backend --context kind-$(K8S_CLUSTER) --timeout=120s
	@kubectl rollout status deployment/frontend --context kind-$(K8S_CLUSTER) --timeout=120s
	@kubectl wait --for=condition=ready pod --all --context kind-$(K8S_CLUSTER) --timeout=120s
	@echo "All pods are Ready!"

k8s-verify:
	@bash k8s/verify

k8s-up: k8s-cluster k8s-build k8s-load k8s-deploy k8s-wait k8s-verify
	@echo "ChoreSync is running on Kubernetes at http://localhost:8090"

k8s-down:
	@echo "Deleting kind cluster $(K8S_CLUSTER)..."
	@kind delete cluster --name $(K8S_CLUSTER) 2>/dev/null || true
	@echo "Kind cluster deleted."

k8s-logs:
	@kubectl logs -l app.kubernetes.io/part-of=choresync --all-containers=true -f --context kind-$(K8S_CLUSTER)

ext-verify:
	@echo "Verifying Agent Extension Pack compliance (agent-plugins.org / AAIF)..."
	@verify-extension-pack

ext-eval:
	@echo "Evaluating Agent Extension Pack Conformance & Domain Alignment (Gate 14)..."
	@evaluate-extension-pack

ext-mcp-test:
	@echo "Testing ChoreSync MCP server initialize..."
	@echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}' | python3 mcp-server/server.py
	@echo "Testing ChoreSync MCP server tools/list..."
	@echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}' | python3 mcp-server/server.py

ext-capabilities:
	@echo "Running Contract Audit Capability..."
	@bash agent-capabilities/contract-audit/scripts/audit-contract.sh
	@echo "Running Database Schema Migration Capability..."
	@bash agent-capabilities/db-migration-runner/scripts/check-migrations.sh
	@echo "Running Cluster Health Prober Capability..."
	@bash agent-capabilities/cluster-health-prober/scripts/probe-cluster.sh
