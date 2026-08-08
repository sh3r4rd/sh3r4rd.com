copy:
	@aws s3 cp ./dist/index.html s3://$(bucket)/ \
		--content-type "text/html" \
		--cache-control "no-cache, no-store, must-revalidate"

sync:
	@aws s3 sync ./dist s3://$(bucket)/ --exclude index.html --exclude "images/*" --delete

build:
	@npm run build

lint:
	@npm run lint

deploy: build copy sync

server:
	@npm run dev

# Lambda builds — compile Go binaries for AWS Lambda (arm64)
INFRA_DIR := infra/recruiter-dashboard
BUILD_DIR := $(INFRA_DIR)/.build
GOFLAGS   := GOOS=linux GOARCH=arm64 CGO_ENABLED=0

build-lambdas: build-email-parser build-api-handler

build-email-parser:
	@mkdir -p $(BUILD_DIR)/email-parser
	@cd $(INFRA_DIR)/lambda-src/email-parser && $(GOFLAGS) go build -o $(CURDIR)/$(BUILD_DIR)/email-parser/bootstrap ./cmd/handler/

build-api-handler:
	@mkdir -p $(BUILD_DIR)/api-handler
	@cd $(INFRA_DIR)/lambda-src/api-handler && $(GOFLAGS) go build -o $(CURDIR)/$(BUILD_DIR)/api-handler/bootstrap .

# Go tests
test-go: test-email-parser test-api-handler

test-email-parser:
	@cd $(INFRA_DIR)/lambda-src/email-parser && go test -v -race ./...

test-api-handler:
	@cd $(INFRA_DIR)/lambda-src/api-handler && \
		RECRUITER_TABLE=test CORS_ALLOW_ORIGIN=http://localhost DATE_INDEX_NAME=date-index \
		AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1 \
		go test -v -race ./...

# Terraform
tf-init:
	@terraform -chdir=$(INFRA_DIR) init

tf-validate: tf-init
	@terraform -chdir=$(INFRA_DIR) validate

tf-fmt:
	@terraform -chdir=$(INFRA_DIR) fmt -recursive

tf-fmt-check:
	@terraform -chdir=$(INFRA_DIR) fmt -recursive -check

tf-plan:
	@terraform -chdir=$(INFRA_DIR) plan

# terraform.tfvars <-> SSM sync (SSM is the durable source of truth; the file is git-ignored)
#
# TFVARS_ARGS passes flags through to the script — chiefly --force, which skips
# the confirmation prompt for non-interactive use:
#   make tf-vars-push TFVARS_ARGS=--force
# tf-vars-diff exits non-zero when drift exists so it can gate a script or CI
# check; the resulting `Error 1` is the drift signal, not a malfunction.
TFVARS_ARGS ?=

tf-vars-pull:
	@$(INFRA_DIR)/scripts/tfvars.sh pull $(TFVARS_ARGS)

tf-vars-push:
	@$(INFRA_DIR)/scripts/tfvars.sh push $(TFVARS_ARGS)

tf-vars-diff:
	@$(INFRA_DIR)/scripts/tfvars.sh diff $(TFVARS_ARGS)

# Clean build artifacts
clean:
	@rm -rf dist $(BUILD_DIR)

# Run all CI checks locally
ci: lint build test-go tf-fmt-check tf-validate
