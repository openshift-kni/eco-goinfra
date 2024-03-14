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

install: deps-update
	@echo "Installing needed dependencies"

test:
	go test ./...

