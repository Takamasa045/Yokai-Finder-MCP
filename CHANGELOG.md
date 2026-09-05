# Changelog

## 0.10.0

- Catalog grew to 227 index rows and 227 bilingual cards.
- New names include 夜雀, 火前坊, 隠れ里, 乙姫, 鳳凰, 天逆毎, 鬼童丸, 金神, 毛倡妓, 五徳猫, 塵塚怪王.

## 0.9.0

- Encyclopedia and index are both 207. Every roster row now has a bilingual card.
- `list_yokai` / `list_curated_yokai` limits raised from 200 to 400.
- New names include 大天狗, 木葉天狗, 山姫, 川姫, 団三郎狸, 金長狸, 善狐, 白蛇, 安達ヶ原, 飛頭蛮, 人面魚, 面霊気.
- Remaining obscure roster cards filled (わいら, 煙羅煙羅, 海座頭, 瀬戸大将, 琴古主, ガタロ, and others).

## 0.8.0

- Encyclopedia grew to 172 bilingual cards; the index is now 191 names.
- New entries include 貉, お歯黒べったり, 烏天狗, 狗賓, 飯綱, 八百比丘尼, 隠神刑部, 芝右衛門狸, 白狐, 夜行さん, 青い服の女.
- Roster cards added for 白蔵主, 野狐, 大首, 七人みさき, 逆柱, 朧車, ぼろぼろとん, 波山, オサキ, 目競, 牛頭, 馬頭, and others.

## 0.7.0

- Encyclopedia grew to 133 bilingual cards; the index is now 177 names.
- New well-known entries include 二口女, 手の目, 赤いマント, きさらぎ駅, メリーさん, カシマさん, 金霊, 獺, 山爺.
- Rank-3/4 cards added for names already on the roster (山童, 大入道, 磯女, 釣瓶落とし, ひょうすべ, 抜け首, 鉄鼠, ぬっぺふほふ, 百々目鬼, 一本だたら, 神社姫, and others).

## 0.6.1

- Catalog is validated at process start (required fields, unique names, canonical categories, profiles ⊆ index).
- `yokai://yokai/河童` style URIs resolve: Japanese names are registered as concrete resources because URI templates only match ASCII.
- `get_yokai` includes the index card (tags, tone, famousRank) even when a full encyclopedia profile is present.
- Tool descriptions use live catalog counts so they no longer drift to “160+”.
- Release archives include LICENSE. CI runs gofmt and `go test -race`.

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
