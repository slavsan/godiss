default: help

.PHONY: help
help:
	@echo "help..."

test:
	@go test -race -v -coverpkg=./... -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
