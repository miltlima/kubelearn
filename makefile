SHELL := /bin/bash
.DEFAULT_GOAL := help

# -----------------------------------------------------------------------------
# Configuration
# -----------------------------------------------------------------------------
KUBE_CONTEXT ?= kind-kubelearn
MANIFESTS_DIR ?= manifests
TERRAFORM_DIR ?= config
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
TERRAFORM := terraform -chdir=$(TERRAFORM_DIR)
NPM := npm --prefix $(FRONTEND_DIR)

.PHONY: \
	help \
	manifests-apply \
	manifests-install \
	apply-% \
	manifests-delete \
	manifests-lint \
	terraform-init \
	terraform-plan \
	terraform-apply \
	terraform-destroy \
	kube-context \
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
	@echo "  TERRAFORM_DIR   Terraform configuration directory (default: $(TERRAFORM_DIR))"
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

manifests-lint: ## Dry-run manifests for basic validation
	@echo "Validating manifests in $(MANIFESTS_DIR)"
	@for file in $(YAML_FILES); do \
		echo "Checking $$file"; \
		kubectl apply --dry-run=client -f $$file >/dev/null; \
	done

# -----------------------------------------------------------------------------
# Terraform
# -----------------------------------------------------------------------------
terraform-init: ## Initialize Terraform state
	@echo "Initializing Terraform in $(TERRAFORM_DIR)"
	@$(TERRAFORM) init

terraform-plan: terraform-init ## Show Terraform execution plan
	@echo "Generating Terraform plan"
	@$(TERRAFORM) plan

terraform-apply: terraform-init ## Apply Terraform infrastructure
	@echo "Applying Terraform configuration"
	@$(TERRAFORM) apply -auto-approve

terraform-destroy: terraform-init ## Destroy Terraform infrastructure
	@echo "Destroying Terraform infrastructure"
	@$(TERRAFORM) destroy -auto-approve

# -----------------------------------------------------------------------------
# Runtime helpers
# -----------------------------------------------------------------------------
logs-dir:
	@$(MKDIR_P) $(LOG_DIR)

backend-build: ## Build backend binary
	@echo "Building backend binary at $(BACKEND_BIN)"
	@$(MKDIR_P) $(dir $(BACKEND_BIN))
	@go build -o $(BACKEND_BIN) ./$(BACKEND_DIR)

backend-start: backend-build logs-dir ## Start backend service in the background
	@echo "Starting backend (logging to $(LOG_DIR)/backend.log)"
	@nohup ./$(BACKEND_BIN) > $(LOG_DIR)/backend.log 2>&1 &
	@echo $$! > $(LOG_DIR)/backend.pid

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
	@nohup $(NPM) run start > $(LOG_DIR)/frontend.log 2>&1 &
	@echo $$! > $(LOG_DIR)/frontend.pid

frontend-stop: ## Stop frontend dev server if running
	@echo "Stopping frontend"
	@if [ -f $(LOG_DIR)/frontend.pid ]; then \
		kill "$$(cat $(LOG_DIR)/frontend.pid)" >/dev/null 2>&1 || true; \
		rm -f $(LOG_DIR)/frontend.pid; \
	else \
		pkill -f "$(FRONTEND_DIR)" >/dev/null 2>&1 || true; \
		pkill -f "react-scripts start" >/dev/null 2>&1 || true; \
	fi

kube-context: ## Switch kubectl to the KubeLearn cluster context
	@echo "Switching kubectl context to '$(KUBE_CONTEXT)'"
	@attempt=0; \
	until kubectl config get-contexts $(KUBE_CONTEXT) >/dev/null 2>&1; do \
		if [ $$attempt -ge 20 ]; then \
			echo "Context '$(KUBE_CONTEXT)' not found. Aborting." >&2; \
			exit 1; \
		fi; \
		attempt=$$((attempt+1)); \
		echo "Waiting for context '$(KUBE_CONTEXT)' to become available..."; \
		sleep 3; \
	done
	@kubectl config use-context $(KUBE_CONTEXT)

infra-up: terraform-apply ## Provision infrastructure and switch kubectl context
	@$(MAKE) kube-context
	@echo "Infrastructure provisioned and context set to '$(KUBE_CONTEXT)'"

kubelearn: infra-up backend-start frontend-start ## Provision infra, switch context, start backend & frontend
	@echo "KubeLearn environment ready."
	@echo "  Backend log : $(LOG_DIR)/backend.log"
	@echo "  Frontend log: $(LOG_DIR)/frontend.log"

stop-kubelearn: backend-stop frontend-stop ## Stop backend and frontend processes
	@echo "KubeLearn services stopped"

clean: stop-kubelearn ## Stop services and remove build artifacts
	@echo "Removing backend binary and logs"
	@rm -rf $(BACKEND_BIN) $(LOG_DIR)
