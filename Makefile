.DEFAULT_GOAL := help
.PHONY: help generate test 
current_dir = $(shell pwd)

help:
	@echo "you may:"
	@echo "   make test     - run all tests inside tests/"
	@echo "   make generate - re-generate client code"

generate:
	@rm ./api.yaml ./client.gen.go || true
	@patch --output=./api.yaml ./patch/api.yaml.orig ./patch/api.yaml.diff
	@go tool oapi-codegen -config cfg.yaml api.yaml
	@patch --output=./client.gen.go ./client.gen.go.orig ./patch/de-ptrize.diff
	@rm ./client.gen.go.orig
	@go mod tidy

test:
	@(cd ${current_dir}/tests && go test -v -vet=all -timeout 3m -run ./...)
