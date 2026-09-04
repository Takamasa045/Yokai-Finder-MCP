# Contributing

## Develop

```bash
go test ./...
go vet ./...
go build -o yokai-finder-mcp ./cmd/server
```

## Add a yokai

1. Edit `internal/yokai/data/index.json` (roster row: Name, NativeName, Category, Region, BlurbJA, Tags 3–8, Tone, FamousRank).
2. Optionally add a bilingual encyclopedia card to `internal/yokai/data/profiles.json` (`Sources` welcome).
3. Add extra spellings to `internal/yokai/data/aliases.json` if カッパ / ひらがな / English nicknames should resolve.
4. Run `go test ./internal/yokai ./internal/handler`.

Canonical categories are Japanese keys such as `水系`, `付喪神`, `現代伝承`. English hints (`water`, `tsukumogami`) are mapped in code.

## HTTP

Loopback is allowed without a token:

```bash
./yokai-finder-mcp -http 127.0.0.1:8080
```

Binding a public address requires `YOKAI_FINDER_TOKEN`. Clients send `Authorization: Bearer <token>` to `/mcp`. `/healthz` stays open.
