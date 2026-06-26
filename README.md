# cms-cli

A minimal CLI for managing Markdown content with front matter, organized into a single JSON output by category and tag.

<details>
<summary><strong>English</strong></summary>

## Commands

| Command | Description |
|---|---|
| `cmsc init` | Create a default `cmsc.json` config file in the current directory. Fails if one already exists. |
| `cmsc setup` | Create the directories defined in `cmsc.json` (`articleDir` and `imageDir` sub-directories). |
| `cmsc new` | Interactively create a new Markdown post with front matter under `articleDir`. Prompts for title, category, and tags. The slug is auto-generated as a UUID (never entered manually) and the date is filled automatically with the current timestamp. |
| `cmsc convert` | Read `cmsc.json`, walk `articleDir` for `.md` files, parse each front matter, and build/write the output JSON (`all` / `byCategory` / `byTag`) to `outputDir`. Posts with unregistered categories or tags are skipped with a notice. |

## Quick Start

```bash
# 1. Initialize config
cmsc init

# 2. Create directories
cmsc setup

# 3. Create a new post (interactive)
cmsc new

# 4. Build the output JSON
cmsc convert
```

## Config file: `cmsc.json`

| Field | Description |
|---|---|
| `articleDir` | Directory where `.md` files are stored |
| `imageDir` | Root directory for images |
| `outputDir` | Path for the generated JSON output |
| `categories` | Whitelist of allowed categories |
| `tags` | Whitelist of allowed tags |

## Directory structure created by `cmsc setup`

```
<articleDir>/
<imageDir>/
├── article/
├── category/
└── tag/
```

</details>

<details>
<summary><strong>日本語</strong></summary>

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `cmsc init` | カレントディレクトリにデフォルトの `cmsc.json` を生成します。既に存在する場合は失敗します。 |
| `cmsc setup` | `cmsc.json` に定義されたディレクトリ（`articleDir` および `imageDir` 配下のサブディレクトリ）を作成します。 |
| `cmsc new` | 対話形式でフロントマター付きの新規Markdown記事を `articleDir` 配下に作成します。タイトル・カテゴリ・タグを順に質問されます。slugはUUIDで自動採番され（手動入力不可）、dateは現在日時が自動で埋め込まれます。 |
| `cmsc convert` | `cmsc.json` を読み込み、`articleDir` 内の `.md` ファイルを走査してフロントマターを解析し、`outputDir` に `all` / `byCategory` / `byTag` を含むJSONを書き出します。未登録のカテゴリ・タグを持つ記事はスキップされ、通知が表示されます。 |

## クイックスタート

```bash
# 1. 設定ファイルを初期化
cmsc init

# 2. ディレクトリを作成
cmsc setup

# 3. 新規記事を作成（対話形式）
cmsc new

# 4. JSONを出力
cmsc convert
```

## 設定ファイル: `cmsc.json`

| フィールド | 説明 |
|---|---|
| `articleDir` | `.md` ファイルを格納するディレクトリ |
| `imageDir` | 画像を保存するルートディレクトリ |
| `outputDir` | 生成されるJSONの出力先パス |
| `categories` | 許可されるカテゴリのホワイトリスト |
| `tags` | 許可されるタグのホワイトリスト |

## `cmsc setup` で作成されるディレクトリ構造

```
<articleDir>/
<imageDir>/
├── article/
├── category/
└── tag/
```

</details>
