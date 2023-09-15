# Makefile to install YAML manifests

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
