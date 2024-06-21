GO_PACKAGES=$(shell go list ./... | grep -v vendor)
.PHONY: lint \
        deps-update \
        vet
vet:
	go vet ${GO_PACKAGES}

lint:
	@echo "Running go lint"
	scripts/golangci-lint.sh

deps-update:
	go mod tidy && \
	go mod vendor

lib-sync:
	export FLAGS_v=100;
	go run ./internal/sync

install: deps-update
	@echo "Installing needed dependencies"

test:
	go test ./...

coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

