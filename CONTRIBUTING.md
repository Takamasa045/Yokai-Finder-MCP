# Contributing

## Develop

```bash
make check
go build -o yokai-finder-mcp ./cmd/server
```

`make check` runs gofmt, `go vet`, and `go test -race`. The embedded catalog is validated when the process starts.

## Add a yokai

1. Edit `internal/yokai/data/index.json` (camelCase: `name`, `nativeName`, `category`, `region`, `blurbJa`, `tags` 3–8, `tone`, `famousRank`).
2. Optionally add a bilingual encyclopedia card to `internal/yokai/data/profiles.json` (`category` is the Japanese key such as `水系`; `categoryEn` is optional; `sources` welcome).
3. Add extra spellings to `internal/yokai/data/aliases.json` if カッパ / ひらがな / English nicknames should resolve.
4. Run `go test ./internal/yokai ./internal/handler`.

Canonical categories are Japanese keys such as `水系`, `付喪神`, `現代伝承`. English hints (`water`, `tsukumogami`) are mapped in code.

`list_yokai` defaults to 20 and caps at 600 (`offset` pages through the rest). Raise `maxRosterLimit` in `internal/handler/handler.go` if the roster grows past that.

Do not add encyclopedia text that looks like generation residue (`Do not merge`, `Keep X as`, `encyclopedia plate`). Rank-1 cards need `legendsJa` (3+) and a `summaryJa` of at least 40 Japanese characters. `python3 scripts/scrub-catalog.py` is a one-shot cleaner plus rank-1 Japanese overlay; after that, `go test ./internal/yokai` should stay green.

## HTTP

Loopback is allowed without a token:

```bash
./yokai-finder-mcp -http 127.0.0.1:8080
```

Binding a public address requires `YOKAI_FINDER_TOKEN`. Clients send `Authorization: Bearer <token>` to `/mcp`. `/healthz` stays open.

See `CHANGELOG.md` for release notes.

## Registry metadata

- `server.json` — MCP Registry schema
- `smithery.yaml` — Smithery start command
- `mcp.json` — local stdio example using the release binary name `yokai-finder-mcp`

Publishing a listing is a manual step on each host; do not commit tokens.
