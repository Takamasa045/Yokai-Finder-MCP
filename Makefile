.PHONY: test vet fmt race cover check build export-catalog

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./...

race:
	go test -race ./...

cover:
	go test ./... -cover

check:
	test -z "$$(gofmt -l .)"
	go vet ./...
	go test -race ./...

build:
	go build -o yokai-finder-mcp ./cmd/server

export-catalog:
	go run ./cmd/export-catalog
