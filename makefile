# Makefile to install YAML manifests and install kind optionally
help:
	@echo "Makefile Help"
	@echo ""
	@echo "Targets:"
	@echo "  all          Installs all YAML manifests."
	@echo "  clean        Deletes all installed resources."
	@echo "  check-syntax Checks the syntax of all manifests without actually installing them."
	@echo "  init         Initializes the Terraform repository."
	@echo "  apply        Applies Terraform configurations."
	@echo "  destroy      Destroys Terraform resources."
	@echo ""
	@echo "Usage:"
	@echo "  make all"
	@echo "  make clean"
	@echo "  make check-syntax"
	@echo "  make init"
	@echo "  make apply"
	@echo "  make destroy"

# Directory where the YAML manifests are located
MANIFESTS_DIR := manifests

# Command to install the manifests
INSTALL_COMMAND := kubectl apply -f

# List all YAML files in the manifests folder
YAML_FILES := $(wildcard $(MANIFESTS_DIR)/*.yaml)

# Get only the file names (without the directory path)
BASE_NAMES := $(notdir $(YAML_FILES))

# Add a prefix for the install target
TARGETS := $(addprefix install-,$(BASE_NAMES))

# Default rule to install all manifests
all: $(TARGETS)

# Rules to install each individual manifest
$(TARGETS): install-%: $(MANIFESTS_DIR)/%
	$(INSTALL_COMMAND) $<

# Rule to clean up all installed resources
clean:
	kubectl delete -f $(MANIFESTS_DIR)

# Rule to check syntax of all manifests without actually installing them
check-syntax:
	for file in $(YAML_FILES); do \
		kubectl apply --dry-run=client -f $$file; \
	done

# Directory where the Terraform code is
TERRAFORM_DIR := config

# Commands
TERRAFORM := terraform
TERRAFORM_CMD := $(TERRAFORM) -chdir=$(TERRAFORM_DIR)
TERRAFORM_INIT := $(TERRAFORM_CMD) init  
TERRAFORM_APPLY := $(TERRAFORM_CMD) apply -auto-approve
TERRAFORM_DESTROY := $(TERRAFORM_CMD) destroy -auto-approve

# Targets
.PHONY: init apply destroy init-upgrade

# Rules
init: ## Initialize Terraform repository
	@echo "Initializing Terraform..."
	$(TERRAFORM_INIT)

apply: ## Apply Terraform configurations
	@echo "Applying Terraform configurations..."
	$(TERRAFORM_APPLY)

destroy: ## Destroy Terraform resources
	@echo "Destroying Terraform resources..."
	$(TERRAFORM_DESTROY)
