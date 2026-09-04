FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o yokai-finder-mcp ./cmd/server

FROM alpine:3.21
RUN apk --no-cache add ca-certificates \
    && adduser -D -H -u 65532 nonroot
WORKDIR /app
COPY --from=builder /app/yokai-finder-mcp .
USER 65532:65532
ENTRYPOINT ["./yokai-finder-mcp"]
