# Changelog

## 0.6.0

- Encyclopedia grew to 98 bilingual cards. Famous-rank 1–2 stay complete; 15 well-known rank-3 names now have profiles (山彦, 磯撫, 大百足, 火車, 枕返し, 天井嘗, 一つ目入道, 牛鬼, 疫病神, 件, アマビコ, 白澤, 骨女, 累, 管狐).
- Agents are told not to call `list_curated_yokai` or `get_cover_thumbnail`; those tools remain only for compatibility and are registered last.
- Name completion also fills `yokai://yokai/{name}` resource URIs. Tool-argument completion still depends on the MCP Go SDK, which does not accept `ref/tool`.
- CI and Release workflows use checkout v5 / setup-go v6, and Release caches `source/go.sum`.

## 0.5.0

- Catalog JSON is camelCase. Profile `category` is the Japanese index key (`水系`); English labels moved to `categoryEn`.
- NDL search uses OpenSearch `any=` (keyword), ranks folklore-tagged books first, retries 5xx/timeouts, and can HEAD-check covers via `verifyCovers`.
- Prompt/resource name completion, `yokai://catalog`, and three missing rank 1–2 profiles (狐火, 人魂, 死神).

## 0.4.0

- Embedded JSON catalog, aliases, Japanese/English category filters, related/compare tools, loopback streamable HTTP, LICENSE, and PR CI.

## 0.3.0

- `suggest_yokai`, `get_yokai`, 50 encyclopedia cards, 161-entry tagged index.
