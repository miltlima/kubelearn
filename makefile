SHELL := /bin/bash
.DEFAULT_GOAL := help

# -----------------------------------------------------------------------------
# Configuration
# -----------------------------------------------------------------------------
KUBE_CONTEXT ?= kind-kubelearn
MANIFESTS_DIR ?= manifests
GATEWAY_API_VERSION ?= v1.2.1
BACKEND_DIR ?= cmd
BACKEND_BIN ?= bin/kubelearn
FRONTEND_DIR ?= kubelearn-frontend
LOG_DIR ?= logs

# -----------------------------------------------------------------------------
# Derived variables
# -----------------------------------------------------------------------------
YAML_FILES := $(sort $(wildcard $(MANIFESTS_DIR)/*.yaml))
MANIFEST_TARGETS := $(patsubst $(MANIFESTS_DIR)/%,apply-%,$(YAML_FILES))
MKDIR_P := mkdir -p
NPM := npm --prefix $(FRONTEND_DIR)

.PHONY: \
	help \
	manifests-apply \
	manifests-install \
	apply-% \
	manifests-delete \
	manifests-lint \
	gateway-api-install \
	build \
	cluster-up \
	cluster-down \
	infra-up \
	backend-build \
	backend-start \
	backend-stop \
	frontend-install \
	frontend-start \
	frontend-stop \
	kubelearn \
	stop-kubelearn \
	logs-dir \
	clean

help: ## Show available targets and configurable variables
	@echo "KubeLearn automation targets"
	@echo
	@grep -hE '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"} {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "Variables:"
	@echo "  KUBE_CONTEXT    kubectl context to switch to (default: $(KUBE_CONTEXT))"
	@echo "  MANIFESTS_DIR   directory containing Kubernetes manifests (default: $(MANIFESTS_DIR))"
	@echo "  GATEWAY_API_VERSION  Gateway API release to install (default: $(GATEWAY_API_VERSION))"
	@echo "  FRONTEND_DIR    Frontend app directory (default: $(FRONTEND_DIR))"
	@echo "  BACKEND_BIN     Location for compiled backend binary (default: $(BACKEND_BIN))"
	@echo "  LOG_DIR         Directory for runtime logs and PID files (default: $(LOG_DIR))"

# -----------------------------------------------------------------------------
# Kubernetes manifests
# -----------------------------------------------------------------------------
manifests-apply: $(MANIFEST_TARGETS) ## Apply all manifests in $(MANIFESTS_DIR)

manifests-install: manifests-apply ## Backwards compatibility alias for manifests-apply

apply-%: $(MANIFESTS_DIR)/% ## Apply a single manifest (usage: make apply-filename.yaml)
	@echo "Applying $<"
	@kubectl apply -f $<

manifests-delete: ## Delete all manifests in $(MANIFESTS_DIR)
	@echo "Deleting manifests in $(MANIFESTS_DIR)"
	@kubectl delete -f $(MANIFESTS_DIR) --ignore-not-found

gateway-api-install: ## Install the Gateway API CRDs and the lab gateway class (needs internet)
	@echo "Installing Gateway API $(GATEWAY_API_VERSION)"
	@kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/$(GATEWAY_API_VERSION)/standard-install.yaml
	@kubectl wait --for=condition=Established --timeout=60s crd/gatewayclasses.gateway.networking.k8s.io crd/gateways.gateway.networking.k8s.io crd/httproutes.gateway.networking.k8s.io
	@kubectl apply -f $(MANIFESTS_DIR)/addons

manifests-lint: ## Dry-run manifests for basic validation
	@echo "Validating manifests in $(MANIFESTS_DIR)"
	@for file in $(YAML_FILES); do \
		echo "Checking $$file"; \
		kubectl apply --dry-run=client -f $$file >/dev/null; \
	done


# -----------------------------------------------------------------------------
# Runtime helpers
# -----------------------------------------------------------------------------
logs-dir:
	@$(MKDIR_P) $(LOG_DIR)

build: frontend-install ## Build the complete binary, with the web interface embedded
	@echo "Building the web interface"
	@$(NPM) run build
	@echo "Building $(BACKEND_BIN) with the interface embedded"
	@$(MKDIR_P) $(dir $(BACKEND_BIN))
	@go build -tags embed -o $(BACKEND_BIN) ./$(BACKEND_DIR)

backend-build: ## Build backend binary
	@echo "Building backend binary at $(BACKEND_BIN)"
	@$(MKDIR_P) $(dir $(BACKEND_BIN))
	@go build -o $(BACKEND_BIN) ./$(BACKEND_DIR)

backend-start: backend-build logs-dir ## Start backend service in the background
	@echo "Starting backend (logging to $(LOG_DIR)/backend.log)"
	@nohup ./$(BACKEND_BIN) > $(LOG_DIR)/backend.log 2>&1 & echo $$! > $(LOG_DIR)/backend.pid

backend-stop: ## Stop backend service if running
	@echo "Stopping backend"
	@if [ -f $(LOG_DIR)/backend.pid ]; then \
		kill "$$(cat $(LOG_DIR)/backend.pid)" >/dev/null 2>&1 || true; \
		rm -f $(LOG_DIR)/backend.pid; \
	else \
		pkill -f "$(BACKEND_BIN)" >/dev/null 2>&1 || true; \
	fi

frontend-install: ## Install frontend dependencies
	@echo "Installing frontend dependencies"
	@$(NPM) install

frontend-start: frontend-install logs-dir ## Start frontend dev server in the background
	@echo "Starting frontend (logging to $(LOG_DIR)/frontend.log)"
	@nohup $(NPM) run start > $(LOG_DIR)/frontend.log 2>&1 & echo $$! > $(LOG_DIR)/frontend.pid

frontend-stop: ## Stop frontend dev server if running
	@echo "Stopping frontend"
	@if [ -s $(LOG_DIR)/frontend.pid ]; then \
		kill "$$(cat $(LOG_DIR)/frontend.pid)" >/dev/null 2>&1 || true; \
	fi
	@rm -f $(LOG_DIR)/frontend.pid
	@# npm does not always pass the signal on, so also stop this project's dev server process.
	@pkill -f "$(FRONTEND_DIR)/node_modules/react-scripts" >/dev/null 2>&1 || true

cluster-up: backend-build ## Create the kind cluster, switch kubectl to it and apply the lab resources
	@./$(BACKEND_BIN) up

cluster-down: backend-build ## Delete the kind cluster
	@./$(BACKEND_BIN) down

infra-up: cluster-up ## Alias for cluster-up

kubelearn: cluster-up backend-start frontend-start ## Create the lab cluster, then start backend & frontend (development)
	@echo "KubeLearn environment ready."
	@echo "  Backend log : $(LOG_DIR)/backend.log"
	@echo "  Frontend log: $(LOG_DIR)/frontend.log"

stop-kubelearn: backend-stop frontend-stop ## Stop backend and frontend processes
	@echo "KubeLearn services stopped"

clean: stop-kubelearn ## Stop services and remove build artifacts
	@echo "Removing backend binary and logs"
	@rm -rf $(BACKEND_BIN) $(LOG_DIR)
