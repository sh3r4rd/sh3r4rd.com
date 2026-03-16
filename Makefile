copy:
	@aws s3 cp ./dist/index.html s3://$(bucket)/ \
		--content-type "text/html" \
		--cache-control "no-cache, no-store, must-revalidate"

sync:
	@aws s3 sync ./dist s3://$(bucket)/ --exclude index.html --exclude "images/*" --delete

build:
	@npm run build

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
	@cd $(INFRA_DIR)/lambda-src/email-parser && $(GOFLAGS) go build -o ../../../.build/email-parser/bootstrap ./cmd/handler/

build-api-handler:
	@mkdir -p $(BUILD_DIR)/api-handler
	@cd $(INFRA_DIR)/lambda-src/api-handler && $(GOFLAGS) go build -o ../../../.build/api-handler/bootstrap .