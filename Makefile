diff-check:
	@git update-index --refresh && git diff-index --quiet HEAD --

.PHONY: lint
lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run --timeout 5m --fix -v ./...

.PHONY: test
test:
	@go test -race -short ./...

.PHONY: test
test-cover:
	@go test -race -timeout=10m ./... -coverprofile=coverage.out

build:
	@go build  -o powwly-cli ./cmd/powwy-cli
	@go build  -o powwly ./cmd/powwy-srv