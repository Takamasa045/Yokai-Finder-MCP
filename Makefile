.PHONY: test vet build

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o yokai-finder-mcp ./cmd/server

export-catalog:
	go run ./cmd/export-catalog
