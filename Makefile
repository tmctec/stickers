# ============================================================================
# Stickers Makefile
# ============================================================================
# 2-Layer System: This Makefile defines WHAT to do.
# Complex logic (>10 lines, OS branching) goes in makehelp.sh (HOW to do it).
#
# The repository holds two Go modules: the library at the root (packages
# flexbox and table) and the demos in example/. Targets cover both.
# ============================================================================

.DEFAULT_GOAL := help

# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------

# Tools
BUILD_CMD := go build
TEST_CMD := go test
LINT_CMD := golangci-lint run
FMT_CMD := gofmt -w .

# Paths
BINARY_DIR := ./bin
COVERAGE_FILE := coverage.out
EXAMPLE_DIR := example

# Benchmarks to run with `make bench`, as a go test -bench pattern
BENCH := .

# ---------------------------------------------------------------------------
# Targets
# ---------------------------------------------------------------------------

##@ General

.PHONY: help
help:  ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

.PHONY: prereqs
prereqs:  ## Check for Go and install golangci-lint if it is missing
	@./makehelp.sh prereqs

##@ Build

.PHONY: all
all: build  ## Alias for build — builds everything

.PHONY: build
build:  ## Compile the library and the demos, and build bin/demo-all
	@$(BUILD_CMD) ./...
	@cd $(EXAMPLE_DIR) && $(BUILD_CMD) ./... && $(BUILD_CMD) -o ../$(BINARY_DIR)/demo-all ./demo-all

##@ Test

.PHONY: test
test: test-unit test-integration  ## Run all tests: the library, then the demos

.PHONY: test-unit
test-unit:  ## Test the flexbox and table packages
	@$(TEST_CMD) ./...

.PHONY: test-integration
test-integration:  ## Test the demos, which render real layouts at several terminal sizes
	@cd $(EXAMPLE_DIR) && $(TEST_CMD) ./...

.PHONY: cover
cover:  ## Test the library and print the coverage of each function
	@$(TEST_CMD) -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -func=$(COVERAGE_FILE)

.PHONY: bench
bench:  ## Run the render, sort and filter benchmarks (BENCH=<pattern> to pick some)
	@$(TEST_CMD) -run '^$$' -bench '$(BENCH)' -benchmem ./...

##@ Quality

.PHONY: lint
lint:  ## Lint the library and the demos with golangci-lint
	@$(LINT_CMD) ./...
	@cd $(EXAMPLE_DIR) && $(LINT_CMD) ./...

.PHONY: fmt
fmt:  ## Format every Go file with gofmt
	@$(FMT_CMD)

.PHONY: check
check: lint test  ## Run lint + test together

.PHONY: security
security:  ## Scan for secrets in the tree and every commit, known vulnerabilities, gosec defects, and banned names (BANNED_NAMES)
	@./makehelp.sh security

##@ Install

.PHONY: install
install: install-dev  ## Alias for install-dev

.PHONY: install-dev
install-dev: build  ## Nothing to install: add the library with go get github.com/tmctec/stickers/v2
	@echo "Nothing to install: stickers is a library. Add it with: go get github.com/tmctec/stickers/v2"

.PHONY: install-production
install-production: build  ## Nothing to install: add the library with go get github.com/tmctec/stickers/v2
	@echo "Nothing to install: stickers is a library. Add it with: go get github.com/tmctec/stickers/v2"

.PHONY: uninstall
uninstall:  ## Nothing to uninstall: the library installs no files
	@echo "Nothing to uninstall: stickers installs no files."

##@ Development

.PHONY: dev
dev: fmt test build  ## Format, test, and build

.PHONY: cycle
cycle: uninstall clean build install  ## Full clean rebuild

.PHONY: run
run: build  ## Run all thirteen demos in one program: tab / shift+tab cycle, a about, q quit
	@$(BINARY_DIR)/demo-all $(ARGS)

.PHONY: demo-all
demo-all: run  ## Alias for run

.PHONY: demo-1
demo-1:  ## FlexBox Simple: grid of colored cells sized by ratio
	@cd $(EXAMPLE_DIR) && go run ./demo-1-flex-box-simple/cmd

.PHONY: demo-2
demo-2:  ## FlexBox Horizontal: columns of cells stacked vertically
	@cd $(EXAMPLE_DIR) && go run ./demo-2-flex-box-horizontal/cmd

.PHONY: demo-3
demo-3:  ## FlexBox with Table: a table inside a cell (arrows, mouse, ctrl+s sort, enter copy, letters filter)
	@cd $(EXAMPLE_DIR) && go run ./demo-3-flex-box-with-table/cmd

.PHONY: demo-4
demo-4:  ## Table Simple String: string columns (arrows, mouse, ctrl+s sort, enter copy, letters filter)
	@cd $(EXAMPLE_DIR) && go run ./demo-4-table-simple-string/cmd

.PHONY: demo-5
demo-5:  ## Table Multi-Type: typed columns that sort numerically (arrows, mouse, ctrl+s, enter, letters)
	@cd $(EXAMPLE_DIR) && go run ./demo-5-table-multi-type/cmd

.PHONY: demo-6
demo-6:  ## FlexBox Nested Borders: boxes nested through a content generator (b border)
	@cd $(EXAMPLE_DIR) && go run ./demo-6-flex-box-nested-borders/cmd

.PHONY: demo-7
demo-7:  ## FlexBox Simple Borders: labeled cells in every border style (b border)
	@cd $(EXAMPLE_DIR) && go run ./demo-7-flex-box-simple-borders/cmd

.PHONY: demo-8
demo-8:  ## FlexBox Fixed Rows: fixed and ratio-based row heights (m fixed/dynamic)
	@cd $(EXAMPLE_DIR) && go run ./demo-8-flex-box-fixed-rows/cmd

.PHONY: demo-9
demo-9:  ## FlexBox Fixed Width: fixed and ratio-based cell widths (m fixed/dynamic)
	@cd $(EXAMPLE_DIR) && go run ./demo-9-flex-box-fixed-width/cmd

.PHONY: demo-10
demo-10:  ## FlexBox Mixed Fixed: fixed sizes next to ratios (b border, m fixed/dynamic, h text)
	@cd $(EXAMPLE_DIR) && go run ./demo-10-flex-box-mixed-fixed/cmd

.PHONY: demo-11
demo-11:  ## FlexBox Cell Config: change a cell's fixed sizes at run time (C-W select, w, h, digits, enter, esc)
	@cd $(EXAMPLE_DIR) && go run ./demo-11-flex-box-cell-config/cmd

.PHONY: demo-12
demo-12:  ## FlexBox Row Alignment: rows narrower than the box placed left, center or right (r)
	@cd $(EXAMPLE_DIR) && go run ./demo-12-flex-box-row-align/cmd

.PHONY: demo-13
demo-13:  ## FlexBox Mouse and Cursor: click a cell to select it, the cursor sits in it (click, arrows)
	@cd $(EXAMPLE_DIR) && go run ./demo-13-flex-box-mouse/cmd

##@ Cleanup

.PHONY: clean
clean:  ## Remove build artifacts, including demo binaries built by hand with go build
	@rm -rf $(BINARY_DIR) $(COVERAGE_FILE)
	@rm -f main cmd demo-all demo-*-flex-box-* $(EXAMPLE_DIR)/demo-*/demo-* $(EXAMPLE_DIR)/demo-*/cmd/cmd

.PHONY: clean-all
clean-all: clean  ## Deep clean: also drop cached test results
	@go clean -testcache
