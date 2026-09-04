# Yokai-Finder MCP

日本の妖怪を「知らない状態から出会う → 詳しく読む → 関連書籍を探す」までつなぐ [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) サーバーです。

- **提案** … 「怖い」「水のやつ」など曖昧な希望から候補を出す（`suggest_yokai`）
- **索引 161体** … タグ・トーン付きの軽量名簿（`list_yokai`）
- **1体取得** … 名前で図鑑 or 索引カード（`get_yokai`）
- **詳細図鑑 98体** … 日英バイリンガルの伝承・創作フック（出典付きの増補分あり）
- **国立国会図書館（NDL）** … 蔵書のリアルタイム検索と書影URL候補

現在のバージョン: **v0.6.0**

## 特徴

| レイヤー | 内容 |
|----------|------|
| 提案 (`suggest_yokai`) | 雰囲気・題材・場所・用途から候補＋短い推薦理由 |
| 索引 (`list_yokai`) | 161体の名簿（カテゴリ・タグ・トーン・一言紹介） |
| 1体取得 (`get_yokai`) | 日英名で1件。図鑑ありなら詳細、なければ索引カード |
| 図鑑 (`list_yokai` + `hasProfile` / `yokai_of_the_day`) | 98体の詳細プロフィール。`list_curated_yokai` は非推奨 |
| 近傍 (`related_yokai`) | タグ・カテゴリ・トーンが近い妖怪 |
| 比較 (`compare_yokai`) | 2体を並べて共通点・相違点 |
| 書籍 (`search_yokai_books`) | NDL のキーワード検索（`any`）。件名に妖怪・民話がある資料を優先。`verifyCovers` で書影の生存確認 |
| 書影 (`get_cover_thumbnail`) | ISBN / 全国書誌番号から表紙画像URL候補を生成 |
| 基盤 | 公式 MCP Go SDK（stdio または `-http`）、検索キャッシュ（TTL 5分・最大256件） |

### おすすめの使い分け（エージェント向け）

```
ユーザー: 「怖い妖怪とかいる？」「水っぽいの教えて」
    ↓
suggest_yokai       → 6体前後の提案（whySuggested 付き）
    ↓ 気になる名前
get_yokai           → 1体の説明（source=profile なら詳しい伝承）
    ↓ さらに
search_yokai_books  → NDL で本
yokai_of_the_day    → 図鑑枠なら日替わり/名前指定＋おすすめ本も可
list_yokai          → カテゴリやタグで網羅一覧したいとき
```

NDL の蔵書はローカルにコピーしていません。書籍ツールは都度 API を叩きます。索引・図鑑はリポジトリ内の手書きデータ（索引には Tags / Tone / FamousRank 付き）です。

**コンテキストについて:** 提案はデフォルト6件・最大20件、`get_yokai` は常に1体、`list_curated_yokai` はデフォルト10件・詳細フィールドは opt-in、という設計で会話文脈の圧迫を抑えています。

## インストール

### ダウンロード版（Go不要）

1. [最新Release](https://github.com/Takamasa045/Yokai-Finder-MCP/releases/latest) から自分の環境に合うファイルをダウンロードします。

   | 環境 | ファイル名の末尾 |
   |------|------------------|
   | Mac（Apple Silicon: M1/M2/M3/M4） | `darwin_arm64.tar.gz` |
   | Mac（Intel） | `darwin_amd64.tar.gz` |
   | Linux（Intel/AMD） | `linux_amd64.tar.gz` |
   | Linux（ARM64） | `linux_arm64.tar.gz` |
   | Windows（Intel/AMD） | `windows_amd64.zip` |
   | Windows（ARM64） | `windows_arm64.zip` |

2. 展開します。macOS / Linux の例:

   ```bash
   tar -xzf yokai-finder-mcp_vX.Y.Z_darwin_arm64.tar.gz
   cd yokai-finder-mcp_vX.Y.Z_darwin_arm64
   ```

   WindowsではZIPを展開します。各アーカイブには実行ファイルとこのREADMEが入っています。macOSで実行をブロックされた場合は、展開後のフォルダで `xattr -d com.apple.quarantine yokai-finder-mcp` を一度だけ実行してください。

3. 次の「Codexでの最短登録」または「Claude Desktop / Claude Codeでの設定」で、展開した実行ファイルを指定します。Goのインストールやビルドは不要です。

Release内の`checksums.txt`も一緒にダウンロードすれば、ダウンロードファイルのSHA-256を照合できます。macOS / Linuxでは、展開前のフォルダで `shasum -a 256 -c checksums.txt` を実行してください。

### 必要要件（ソースからビルドする場合）

- Go 1.23 以上
- Git

### ソースからのビルド（開発者向け）

```bash
git clone https://github.com/Takamasa045/Yokai-Finder-MCP.git
cd Yokai-Finder-MCP
go build -o yokai-finder-mcp ./cmd/server
```

## 使い方

### まずは自然言語で

MCPを一度登録したら、ツール名やJSONを書く必要はありません。新しいチャットを開き、普通の日本語で話しかけてください。

最初の一言は、たとえば次のようになります。

> 妖怪ファインダーを使って、怖いけれど有名すぎない妖怪を5体教えて

ほかにも、次のように依頼できます。

- 「水辺に出る妖怪を探して」
- 「子ども向けの、ちょっとかわいい妖怪を教えて」
- 「河童について、伝承も含めて詳しく教えて」
- 「創作の敵キャラに使えそうな妖怪を提案して」
- 「京都に関係する妖怪を一覧にして」
- 「今日の妖怪を1体紹介して」
- 「のっぺらぼうについて読める本を探して」
- 「その中から一番不気味な妖怪を詳しく調べて」
- 「その妖怪を題材に短編小説の設定を考えて」

AIが依頼内容に応じて、提案・図鑑・書籍検索などのツールを自動的に使い分けます。確実にこのMCPを使わせたい場合は、依頼の最初に「妖怪ファインダーを使って」と付けてください。

妖怪の提案・索引・図鑑はローカルデータで利用できます。国立国会図書館から本を探す機能にはインターネット接続が必要です。

### Codexでの最短登録

ダウンロード版またはビルドした実行ファイルを、絶対パスで一度登録します。

```bash
codex mcp add yokai-finder -- /absolute/path/to/yokai-finder-mcp
```

登録後に新しいチャットを開き、上記のように自然言語で依頼してください。

### Claude Desktop / Claude Code での設定

ビルド済みバイナリを指定する場合:

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

`go run` で直接起動する場合:

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

任意の環境変数:

| 変数 | 説明 |
|------|------|
| `YOKAI_FINDER_VERSION` | サーバーが報告するバージョン文字列（未設定時はビルド時の `0.6.0`） |
| `YOKAI_FINDER_TOKEN` | `-http` でループバック以外に bind するときの Bearer トークン |

stdio のほか、Streamable HTTP でも起動できます。既定ではループバックのみ、トークンなしで許可します。

```bash
./yokai-finder-mcp -http 127.0.0.1:8080
# MCP endpoint: http://127.0.0.1:8080/mcp
# health:       http://127.0.0.1:8080/healthz
```

`0.0.0.0:8080` のような公開アドレスには 16 文字以上の `YOKAI_FINDER_TOKEN` が必須です。クライアントは `Authorization: Bearer <token>` を付けます。このバイナリ自体は TLS を話しません。インターネットに出すときは reverse proxy で HTTPS を終端してください。

## 利用可能なツール

### 1. `suggest_yokai` — 曖昧な希望から提案（入口）

名前を知らなくても使えます。「怖い妖怪」「水のやつ」「創作向けのかわいい系」など、雰囲気から候補を返します。各候補に短い推薦理由（`whySuggested`）が付きます。

**パラメータ**

| 名前 | 必須 | 説明 |
|------|------|------|
| `vibe` | 任意 | 雰囲気（例: `怖い` `かわいい` `不気味` `ほのぼの`） |
| `theme` | 任意 | 題材（例: `水` `狐` `付喪神` `呪い`） |
| `setting` | 任意 | 場所・場面（例: `山` `川` `夜の町` `学校`） |
| `audience` | 任意 | 用途（例: `子ども` `入門` `創作` `ホラー`） |
| `term` | 任意 | 上記に当てはまらない自由語 |
| `limit` | 任意 | 件数（デフォルト **6**、最大 **20**） |
| `seed` | 任意 | 同じ条件でも別候補にしたいときのシード |

```json
{ "name": "suggest_yokai", "arguments": { "vibe": "怖い", "limit": 6 } }
```

```json
{ "name": "suggest_yokai", "arguments": { "theme": "水", "audience": "入門" } }
```

返却の `hasProfile: true` は詳細図鑑あり。次は `get_yokai` がおすすめです。

### 2. `get_yokai` — 1体を名前で取得

日本語名または英語名で1体だけ返します。

| `source` | 意味 |
|----------|------|
| `profile` | 図鑑98体のいずれか。伝承・豆知識など詳細あり |
| `index` | 索引のみ。一言紹介＋タグ。本は `search_yokai_books` へ |
| （`found: false`） | 未収録。`suggest_yokai` / `list_yokai` / NDL 検索を案内 |

```json
{ "name": "get_yokai", "arguments": { "name": "河童" } }
```

```json
{ "name": "get_yokai", "arguments": { "name": "八岐大蛇" } }
```

### 3. `list_yokai` — 妖怪索引（161体）

網羅一覧・カテゴリ閲覧用の軽量索引です。名前・カテゴリ・地域・一言紹介（`blurbJa`）・`tags` / `tone` / `famousRank`・`hasProfile` を返します。  
**雰囲気だけの質問には `suggest_yokai` を優先**してください。

**パラメータ**

| 名前 | 必須 | 説明 |
|------|------|------|
| `term` | 任意 | 名前・カテゴリ・地域・紹介文・タグへのキーワード |
| `category` | 任意 | カテゴリ（例: `水系` `water` `付喪神` `狐狸` `現代伝承`）。日英どちらでも同じ集合になります |
| `region` | 任意 | 地域のヒント（例: `東北` `tohoku` `九州` `海`） |
| `tag` | 任意 | タグ（例: `怖い` `かわいい` `入門`） |
| `tone` | 任意 | `gentle` / `comic` / `horror` / `solemn` / `tragic` / `mysterious` / `playful` |
| `famousRankMin` / `famousRankMax` | 任意 | 有名度 1（有名）〜5（マイナー） |
| `hasProfile` | 任意 | `true` なら図鑑カードがあるものだけ |
| `limit` | 任意 | 最大件数（デフォルト・最大とも **200**。全161体を一度に返せる） |

```json
{ "name": "list_yokai", "arguments": {} }
```

```json
{ "name": "list_yokai", "arguments": { "category": "付喪神" } }
```

**索引のカテゴリ例**

予言 / 付喪神 / 古典 / 変化 / 屋敷 / 山系 / 憑きもの / 死霊 / 気象 / 水系 / 海 / 狐狸 / 現代伝承 / 田畑 / 町 / 異形 / 道中 / 霊火 / 鬼

`get_yokai` は別名（`カッパ` `かわっぱ` `Kappa`）でも解決します。見つからないときは `suggestions` に「もしかして」候補が付きます。

### 4. `list_curated_yokai` — 非推奨（互換用）

詳細図鑑の一覧は **`list_yokai` に `hasProfile: true`** を付けてください。このツールは残していますが、新規のエージェント向け説明では使いません。

キュレーション済みの**詳細**プロフィールを一覧・検索します。`summaryJa` / `funFactJa` など日本語フィールド付き。広く眺めたいときは先に `list_yokai` を使ってください。

**パラメータ**

| 名前 | 必須 | 説明 |
|------|------|------|
| `term` | 任意 | 名前・伝承・要約（日本語含む）へのキーワード |
| `category` | 任意 | カテゴリヒント（英語表記寄り。例: `water`） |
| `region` | 任意 | 地域ヒント（例: `tohoku` `kyoto`） |
| `seed` | 任意 | 指定すると並びを決定的にシャッフル |
| `limit` | 任意 | 最大件数（デフォルト 10、最大 200） |
| `includeLegends` | 任意 | 伝承スニペットを含める |
| `includeTraits` | 任意 | 特徴を含める |
| `includeMotifs` | 任意 | モチーフを含める |
| `includeCreativeHooks` | 任意 | 創作フックを含める |

```json
{ "name": "list_curated_yokai", "arguments": { "limit": 50 } }
```

```json
{ "name": "list_curated_yokai", "arguments": { "term": "狐", "includeLegends": true } }
```

### 5. `yokai_of_the_day` — 今日の妖怪

引数なしだと、**JST の日付**から決まる「今日の1体」（詳細図鑑98体から）を紹介します。同じ日は同じ妖怪。プロフィール・伝承・創作フック・NDLのおすすめ書籍・ストーリープロンプトを一度に返します。

**パラメータ**

| 名前 | 必須 | 説明 |
|------|------|------|
| `name` | 任意 | 特定の妖怪名（英語・日本語。図鑑にある名前のみ） |
| `category` | 任意 | 図鑑カテゴリのヒント |
| `region` | 任意 | 図鑑地域のヒント |
| `seed` | 任意 | 日替わりではなくシード基準で選択 |
| `limit` | 任意 | 推薦書籍数（デフォルト 5、最大 10） |

```json
{ "name": "yokai_of_the_day", "arguments": {} }
```

NDL が使えない場合も図鑑データだけで応答し、`notes` に理由を書きます（沈黙の失敗はしません）。

### 6. `search_yokai_books` — 書籍検索（NDL）

国立国会図書館 OpenSearch のキーワード検索（`any`）です。題名一致だけではなく、件名・記述にも当たります。件名や題に「妖怪」「民話」「伝承」がある資料を先に並べます。空の条件だけは「妖怪」で検索します。

**パラメータ**

| 名前 | 必須 | 説明 |
|------|------|------|
| `name` | 任意 | 妖怪名やキーワード（例: `河童` `天狗`） |
| `region` | 任意 | 関連地域（例: `岩手` `京都`） |
| `category` | 任意 | カテゴリ・主題（例: `民話`） |
| `limit` | 任意 | 最大件数（デフォルト 10、最大 50） |
| `verifyCovers` | 任意 | `true` なら書影 URL を HEAD して死んでいる候補を落とす（遅い） |

```json
{ "name": "search_yokai_books", "arguments": { "name": "のっぺらぼう", "limit": 5 } }
```

### 7. `get_cover_thumbnail` — 書影URL

ISBN（10/13桁、ハイフン・スペース可）または全国書誌番号（JP番号）から書影URL候補を返します。ISBN-10 は ISBN-13 に正規化されます。候補は上から順に試してください（資料により提供有無が異なります）。

**パラメータ**

| 名前 | 必須 | 説明 |
|------|------|------|
| `isbn` | ※ | ISBN-10 または ISBN-13 |
| `jpno` | ※ | 全国書誌番号 |

※ `isbn` と `jpno` のどちらか一方は必須。

```json
{ "name": "get_cover_thumbnail", "arguments": { "isbn": "4-04-883992-6" } }
```

### 8. `related_yokai` — 近い妖怪

名前を1つ渡し、タグ・カテゴリ・トーンが近い索引エントリを返します。

```json
{ "name": "related_yokai", "arguments": { "name": "河童", "limit": 6 } }
```

### 9. `compare_yokai` — 2体比較

```json
{ "name": "compare_yokai", "arguments": { "left": "河童", "right": "天狗" } }
```

Resources: `yokai://catalog`（件数とカテゴリ）、`yokai://yokai/{name}`（例: `yokai://yokai/河童`）。  
Prompts: `yokai_story` / `yokai_character_sheet`（名前は completion で補完できます）。

## 収録データ

### 詳細図鑑（98体）

`list_curated_yokai` / `yokai_of_the_day` の対象。データ: `internal/yokai/data/profiles.json`  
いずれも索引にも載り、`list_yokai` では `hasProfile: true` になります。

河童 / 天狗 / 雪女 / 九尾の狐 / 化け猫 / ぬりかべ / ろくろ首 / 海坊主 / 餓者髑髏 / アマビエ / 座敷童子 / 鵺 / 鎌鼬 / 一反木綿 / 天邪鬼 / のっぺらぼう / 垢嘗 / 提灯お化け / 唐傘お化け / 化け狸 / 猫又 / 木霊 / 山姥 / 鬼 / 酒呑童子 / ぬらりひょん / 豆腐小僧 / 子泣き爺 / 産女 / 濡女 / 絡新婦 / 土蜘蛛 / 輪入道 / 覚 / 泥田坊 / 目目連 / 青行燈 / 玉藻前 / 清姫 / 橋姫 / 船幽霊 / 見越入道 / 小豆洗い / 口裂け女 / 犬神 / 雷獣 / 一つ目小僧 / 茨木童子 / 人魚 / 付喪神 / 八岐大蛇 / 幽霊 / 怨霊 / お岩 / お菊 / トイレの花子さん / なまはげ / べとべとさん / 貧乏神 / 生霊 / テケテケ / 人面犬 / 獏 / 麒麟 / 葛の葉 / 狐の嫁入り / 分福茶釜 / 鬼火 / 餓鬼 / あやかし / 八尺様 / くねくね / お露 / 鬼婆 / 夜叉 / 蛟 / だいだらぼっち / 送り犬 / うわん / 影女 / 狐火 / 人魂 / 死神 / 山彦 / 磯撫 / 大百足 / 火車 / 枕返し / 天井嘗 / 一つ目入道 / 牛鬼 / 疫病神 / 件 / アマビコ / 白澤 / 骨女 / 累 / 管狐

### 索引（161体）

`list_yokai` の対象。データ: `internal/yokai/data/index.json`  
上記98体に加え、古典・地方伝承・都市伝説・付喪神・瑞獣などを広く収録。

河童 / 天狗 / 雪女 / 九尾の狐 / 化け猫 / ぬりかべ / ろくろ首 / 海坊主 / 餓者髑髏 / アマビエ / 座敷童子 / 鵺 / 鎌鼬 / 一反木綿 / 天邪鬼 / のっぺらぼう / 垢嘗 / 提灯お化け / 唐傘お化け / 化け狸 / 猫又 / 犬神 / 木霊 / 山彦 / 山姥 / 山童 / 鬼 / 酒呑童子 / 茨木童子 / ぬらりひょん / 一つ目小僧 / 一つ目入道 / 見越入道 / 大入道 / 豆腐小僧 / 小豆洗い / 子泣き爺 / 砂かけ婆 / 産女 / 濡女 / 磯女 / 磯撫 / 船幽霊 / 人魚 / 絡新婦 / 土蜘蛛 / 大百足 / 雷獣 / 鬼火 / 狐火 / 人魂 / 釣瓶落とし / つるべ火 / 輪入道 / 朧車 / 火車 / 覚 / ひょうすべ / 泥田坊 / 枕返し / 天井嘗 / 目目連 / 毛羽毛現 / 化け草履 / 付喪神 / 青行燈 / 青鷺火 / 油すまし / 油赤子 / 白粉婆 / 橋姫 / 清姫 / 玉藻前 / 白蔵主 / 野狐 / 天狐 / 蛟 / 野槌 / 百々目鬼 / 陰摩羅鬼 / 抜け首 / 大首 / 影女 / 逆柱 / 袖引き小僧 / 小豆はかり / 餓鬼 / 幽霊 / あやかし / わいら / 煙羅煙羅 / 一目連 / 大口真神 / 白澤 / 彭侯 / 口裂け女 / テケテケ / トイレの花子さん / 人面犬 / 川太郎 / 瀬戸大将 / ぼろぼろとん / 川赤子 / 波山 / おとろし / うわん / 塗仏 / 苧うに / 髪切り / 馬頭 / 牛頭 / 八岐大蛇 / 牛鬼 / だいだらぼっち / ぬっぺふほふ / べとべとさん / 送り犬 / 貧乏神 / 疫病神 / 生霊 / 怨霊 / お岩 / お菊 / 獏 / 件 / アマビコ / 神社姫 / 海座頭 / 赤舌 / 鳴釜 / 琵琶牧々 / 琴古主 / 行灯お化け / 片輪車 / 骨女 / 後神 / 七人みさき / 一本だたら / 手長足長 / 雪入道 / 不知火 / 狐の嫁入り / 狸囃子 / 管狐 / オサキ / 葛の葉 / 鉄鼠 / 目競 / なまはげ / 麒麟 / 夜叉 / 羅刹 / 鬼婆 / ガタロ / 水虎 / 八尺様 / くねくね / 死神 / お露 / 累 / 分福茶釜

## 開発

### プロジェクト構造

```
yokai-finder-mcp/
├── cmd/
│   ├── server/              # MCP サーバー（stdio / 公式 Go SDK）
│   └── debug/               # 手元確認用ヘルパー
├── internal/
│   ├── cache/               # NDL 検索結果キャッシュ
│   ├── handler/             # ツールのビジネスロジック
│   ├── ndl/                 # NDL OpenSearch クライアント + 書影URL
│   └── yokai/
│       ├── data/index.json     # 索引 161体
│       ├── data/profiles.json  # 詳細図鑑 98体
│       ├── data/aliases.json   # カッパ / かわっぱ などの別名
│       ├── index.go            # 索引 API
│       └── profiles.go         # 図鑑 API
├── pkg/types/               # 共有型定義
├── go.mod
├── mcp.json
└── README.md
```

### テスト

```bash
go test ./...
go test ./... -cover
bash scripts/test-release-build.sh
```

### 妖怪を追加する

**索引を増やす（一覧に顔を出す）**

`internal/yokai/data/index.json` にエントリを追記します。

- 必須イメージ: `Name` / `NativeName` / `Category` / `Region` / `BlurbJA`
- 件数が 200 を超える場合は `list_yokai` の limit 上限も合わせて見直してください

**詳細図鑑を増やす（伝承付きにする）**

`internal/yokai/data/profiles.json` に `Profile` を追記します。

- `Summary` / `FunFact` … 英語
- `SummaryJA` / `FunFactJA` … 日本語
- `SearchQuery` … NDL 向け日本語キーワード
- 索引に同名（`Name` または `NativeName`）があると `hasProfile: true` になります
- 件数が 200 を超える場合は `list_curated_yokai` の limit 上限も合わせて見直してください

### MCP レジストリ

`server.json` は [MCP Registry](https://github.com/modelcontextprotocol/registry) 向けのメタデータ、`smithery.yaml` は Smithery 向けです。掲載は各サービスの公開手順に従います。このリポジトリから自動投稿はしていません。

## ライセンス

MIT

## 謝辞

- 蔵書 API を提供してくださった [国立国会図書館](https://www.ndl.go.jp/)
- 書影フォールバックの [openBD](https://openbd.jp/)
- プロトコル仕様の [Model Context Protocol](https://modelcontextprotocol.io/)
