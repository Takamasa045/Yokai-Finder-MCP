# Yokai-Finder MCP

国立国会図書館（NDL）の蔵書から日本の妖怪に関する書籍を検索できる Model Context Protocol (MCP) サーバーです。書誌検索に加えて、日英バイリンガルの「妖怪図鑑」と日替わりの「今日の妖怪」で、調べものと創作の両方を楽しくします。

## 特徴

- **妖怪書籍検索**: NDL OpenSearch API で妖怪関連の書籍を検索。ISBN がある書籍には書影（表紙画像）URL候補を自動付与
- **妖怪索引（111体）**: 名前・カテゴリ・一言紹介の軽量リスト。どんな妖怪がいるかざっくり眺める入口
- **今日の妖怪**: 日付に連動して毎日1体をピックアップ。伝承・特徴・創作フック・おすすめ書籍・ストーリープロンプトをまとめて返却
- **妖怪図鑑**: 15体のキュレーション済み詳細プロフィール（河童・天狗・雪女・座敷童子・鵺・鎌鼬・一反木綿・天邪鬼など）。要約と豆知識は日英併記
- **書影ツール**: ISBN（10/13桁・ハイフン可）や全国書誌番号から書影URL候補を生成
- **キャッシュ**: NDL API 呼び出しを削減する組み込みキャッシュ（TTL 5分・最大256件）
- **公式 MCP Go SDK**: `modelcontextprotocol/go-sdk` ベースで、stdio トランスポートに完全準拠

## インストール

### 必要要件

- Go 1.23 以上
- Git

### ソースからのビルド

```bash
git clone https://github.com/Takamasa045/Yokai-Finder-MCP.git
cd Yokai-Finder-MCP
go build -o yokai-finder-mcp ./cmd/server
```

## 使い方

### Claude Desktop / Claude Code での設定

ビルド済みバイナリを指定します：

```json
{
  "mcpServers": {
    "yokai-finder": {
      "command": "/path/to/yokai-finder-mcp/yokai-finder-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

`go run` で直接起動する場合：

```json
{
  "mcpServers": {
    "yokai-finder": {
      "command": "go",
      "args": ["run", "./cmd/server"],
      "cwd": "/path/to/yokai-finder-mcp",
      "env": {}
    }
  }
}
```

## 利用可能なツール

### list_yokai — 妖怪索引（111体）

「どんな妖怪がいるか」を一覧するための軽量索引です。各エントリは名前・カテゴリ・地域・一言紹介（`blurbJa`）のみ。詳細伝承は図鑑、本は NDL 検索へつなげます。

**パラメータ:**
- `term` (任意): 名前・カテゴリ・地域・紹介文へのキーワード
- `category` (任意): カテゴリ（例:「水系」「付喪神」「狐狸」）
- `region` (任意): 地域のヒント（例:「東北」「九州」）
- `limit` (任意): 最大件数（デフォルト: 111、最大: 111）

```json
{ "name": "list_yokai", "arguments": { "category": "付喪神" } }
```

`hasProfile: true` の妖怪は `list_curated_yokai` / `yokai_of_the_day` で深掘りできます。気になった名前は `search_yokai_books` の `name` に渡して書籍を探せます。

### yokai_of_the_day — 今日の妖怪

引数なしで呼ぶと、日付から決まる「今日の1体」を紹介します（同じ日は必ず同じ妖怪）。プロフィール・伝承・創作フック・NDLのおすすめ書籍・ストーリープロンプトが一度に届きます。

**パラメータ:**
- `name` (任意): 特定の妖怪名（英語・日本語どちらでも）を指定するとその妖怪を返します
- `category` (任意): 「water spirit」「house spirit」など、カテゴリのヒントで絞り込み
- `region` (任意): 「kumamoto」「tohoku」など、地域のヒントで絞り込み
- `seed` (任意): 数値を指定すると日替わりではなくシード基準の選択になります
- `limit` (任意): 推薦書籍の最大数（デフォルト: 5、最大: 10）

```json
{ "name": "yokai_of_the_day", "arguments": {} }
```

NDL が利用できない場合も、図鑑データだけで応答し `notes` にその旨を記載します（沈黙の失敗をしません）。

### search_yokai_books — 書籍検索

NDL OpenSearch API で妖怪関連の書籍を検索します。ISBN が取得できた書籍には `coverImageCandidates`（書影URL候補）が付きます。

**パラメータ:**
- `name` (任意): 妖怪の名前やキーワード（例:「河童」「天狗」）
- `region` (任意): 関連する地域（例:「岩手」「京都」）
- `category` (任意): カテゴリ（例:「水妖」「民話」）
- `limit` (任意): 最大件数（デフォルト: 10、最大: 50）

```json
{ "name": "search_yokai_books", "arguments": { "name": "河童", "limit": 5 } }
```

### list_curated_yokai — 妖怪図鑑（詳細15体）

キュレーション済みの**詳細**プロフィールを一覧・検索します。`summaryJa` で日本語の要約も返ります。顔ぶれを広く知りたいときは先に `list_yokai` を使ってください。

**パラメータ:**
- `term` (任意): 名前・伝承・要約（日本語含む）へのキーワードマッチ
- `category` (任意): カテゴリで絞り込み
- `region` (任意): 地域で絞り込み
- `seed` (任意): 指定すると並び順を決定的にシャッフル
- `limit` (任意): 最大件数（デフォルト: 10、最大: 50）
- `includeLegends` / `includeTraits` / `includeMotifs` / `includeCreativeHooks` (任意): 伝承・特徴・モチーフ・創作フックを含めるか

```json
{ "name": "list_curated_yokai", "arguments": { "term": "座敷", "includeLegends": true } }
```

### get_cover_thumbnail — 書影URL

ISBN（10/13桁、ハイフン・スペース可）または全国書誌番号（JP番号）から書影URL候補を返します。ISBN-10 は自動で ISBN-13 に変換されます。候補は上から順に試してください（資料により提供有無が異なります）。

**パラメータ:**
- `isbn` (任意): ISBN-10 または ISBN-13
- `jpno` (任意): 全国書誌番号

```json
{ "name": "get_cover_thumbnail", "arguments": { "isbn": "4-04-883992-6" } }
```

## 収録データ

### 索引（111体）— `list_yokai`

河童・天狗・雪女・九尾の狐・化け猫・のっぺらぼう・垢嘗・唐傘お化け・酒呑童子・ぬらりひょん・絡新婦・輪入道・口裂け女・テケテケ・牛頭馬頭 など、有名どころを111体。データは `internal/yokai/index_data.go`。

### 詳細図鑑（15体）— `list_curated_yokai` / `yokai_of_the_day`

河童 / 天狗 / 雪女 / 九尾の狐 / 化け猫 / ぬりかべ / ろくろ首 / 海坊主 / 餓者髑髏 / アマビエ / 座敷童子 / 鵺 / 鎌鼬 / 一反木綿 / 天邪鬼

## 開発

### プロジェクト構造

```
yokai-finder-mcp/
├── cmd/
│   ├── server/            # MCPサーバー エントリーポイント（公式 Go SDK）
│   └── debug/             # 手元確認用ヘルパー
├── internal/
│   ├── cache/             # 検索結果キャッシュ
│   ├── handler/           # ツールのビジネスロジック
│   ├── ndl/               # NDL OpenSearch クライアント + 書影URL
│   └── yokai/             # 索引 (index_data.go) + 図鑑 (profiles_data.go)
├── pkg/types/             # 共有型定義
├── go.mod
├── mcp.json
└── README.md
```

### テストの実行

```bash
go test ./...
go test ./... -cover   # カバレッジ確認
```

### 妖怪を追加するには

**索引（ざっくり一覧）を増やす**  
`internal/yokai/index_data.go` に `IndexEntry` を1つ追記します（`Name` / `NativeName` / `Category` / `Region` / `BlurbJA`）。

**詳細図鑑を増やす**  
`internal/yokai/profiles_data.go` にエントリを1つ追記します。`Summary`/`FunFact` は英語、`SummaryJA`/`FunFactJA` は日本語で書き、`SearchQuery` に NDL 検索に適した日本語キーワードを設定してください。索引側にも同名エントリがあると `hasProfile: true` になります。

## ライセンス

MIT

## 謝辞

- API を提供してくださった国立国会図書館
- 書影フォールバックの [openBD](https://openbd.jp/)
- プロトコル仕様を策定した Model Context Protocol
