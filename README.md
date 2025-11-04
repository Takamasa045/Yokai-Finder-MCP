# Yokai-Finder MCP

国立国会図書館の蔵書から日本の妖怪に関する書籍を検索するための Model Context Protocol (MCP) サーバーです。

## 特徴

- **妖怪書籍検索**: 国立国会図書館APIを使用して妖怪に関する書籍を検索
- **柔軟な検索**: 妖怪名、地域、カテゴリーでの検索が可能
- **キャッシュ機能**: API呼び出しを削減し、パフォーマンスを向上させる組み込みキャッシュ
- **妖怪スポットライト**: キュレーション済みの妖怪プロフィールと創作アイディア、推薦書籍をまとめて取得
- **MCPプロトコル準拠**: Model Context Protocolに完全準拠

## インストール

### 必要要件

- Go 1.23 以上
- Git

### ソースからのビルド

```bash
git clone https://github.com/Takamasa045/Yokai-Finder-MCP.git
cd Yokai-Finder-MCP
go build -o yokai-finder-mcp cmd/server/server.go
```

## 使い方

### MCPサーバーとして

サーバーを直接実行してMCPサーバーとして使用できます：

```bash
./yokai-finder-mcp
```

サーバーはJSON-RPCプロトコルを使用してstdin/stdoutで通信します。

### Claude Desktop での設定

Claude Desktop の MCP 設定に以下を追加してください：

```json
{
  "mcpServers": {
    "yokai-finder": {
      "command": "go",
      "args": ["run", "cmd/server/server.go"],
      "cwd": "/path/to/yokai-finder-mcp",
      "env": {}
    }
  }
}
```

または、ビルド済みのバイナリを使用：

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

## 利用可能なツール

### search_yokai_books

国立国会図書館で妖怪に関する書籍を検索します。

**パラメータ:**
- `name` (任意): 検索する妖怪の名前（例：'河童'、'天狗'、'九尾の狐'）
- `region` (任意): 妖怪に関連する地域や都道府県（例：'岩手'、'京都'）
- `category` (任意): 妖怪のカテゴリー（例：'水妖'、'山妖'、'動物妖怪'）
- `limit` (任意): 返す結果の最大数（デフォルト：10、最大：100）

**使用例:**
```json
{
  "name": "search_yokai_books",
  "arguments": {
    "name": "河童",
    "limit": 5
  }
}
```

### yokai_of_the_day

キュレーション済みの妖怪プロフィールと創作フック、さらには関連書籍のおすすめをまとめて取得します。物語づくりやブレインストーミングの出発点に最適です。

**パラメータ:**
- `name` (任意): 特定の妖怪名（英語または日本語）を指定するとその妖怪を返します
- `category` (任意): 「water spirit」「snow phantom」など、カテゴリのキーワードでフィルタリング
- `region` (任意): 「kumamoto」「honshu」など、地域に関連するヒントでフィルタリング
- `seed` (任意): 数値を指定するとランダム選択が再現可能になります
- `limit` (任意): 推薦書籍の最大数（デフォルト: 5、最大: 10）

**使用例:**
```json
{
  "name": "yokai_of_the_day",
  "arguments": {
    "category": "water spirit",
    "limit": 3,
    "seed": 108
  }
}
```

### list_curated_yokai

キュレーション済みの妖怪プロフィールの一覧を取得し、カテゴリや地域、キーワードで絞り込めます。返却される各プロフィールは必要に応じて伝承・特徴・モチーフ・創作フックを含められます。

**パラメータ:**
- `term` (任意): 名前、概要、伝承などに含まれるキーワード
- `category` (任意): カテゴリのキーワードでフィルタリング
- `region` (任意): 地域のキーワードでフィルタリング
- `seed` (任意): 同じ値を指定すると結果のシャッフル順序も再現可能
- `limit` (任意): 返すプロフィールの最大数（デフォルト: 10、最大: 50）
- `includeLegends` (任意): 伝承を含めるかどうか
- `includeTraits` (任意): 特徴を含めるかどうか
- `includeMotifs` (任意): モチーフを含めるかどうか
- `includeCreativeHooks` (任意): 創作フックを含めるかどうか

**使用例:**
```json
{
  "name": "list_curated_yokai",
  "arguments": {
    "category": "water",
    "includeLegends": true,
    "limit": 5
  }
}
```

## 開発

### プロジェクト構造

```
yokai-finder-mcp/
├── cmd/
│   └── server/
│       └── server.go          # メインサーバーエントリーポイント
├── internal/
│   ├── cache/
│   │   └── cache.go          # キャッシュ機能
│   ├── handler/
│   │   └── handler.go        # MCPリクエストハンドラー
│   └── ndl/
│       └── ndl.go            # 国立国会図書館APIクライアント
├── pkg/
│   └── types/
│       └── types.go          # 型定義
├── go.mod
├── mcp.json                  # MCP設定
└── README.md
```

### テストの実行

```bash
go test ./...
```

### ビルド

```bash
go build -o yokai-finder-mcp cmd/server/server.go
```

## API

サーバーは以下のメソッドを持つMCPプロトコルを実装しています：

- `initialize`: MCP接続の初期化
- `tools/list`: 利用可能なツールの一覧表示
- `tools/call`: ツールの実行

## キャッシュ

サーバーには以下の機能を持つ組み込みキャッシュが含まれています：
- TTL: 30分
- 最大サイズ: 100エントリ
- 期限切れエントリの自動クリーンアップ

## ライセンス

MIT

## コントリビューション

プルリクエストを歓迎します！お気軽にご貢献ください。

## 謝辞

- APIを提供してくださった国立国会図書館
- プロトコル仕様を策定したModel Context Protocol
