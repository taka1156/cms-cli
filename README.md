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
| `cmsc publish` | Upload changed images to Cloudflare R2 (diff-based, tracked via `.cache.json`) and upload the output JSON files to R2. Requires `R2_ACCESS_KEY_ID` and `R2_SECRET_ACCESS_KEY` environment variables. |

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

# 5. Publish to R2
cmsc publish
```

## Config file: `cmsc.json`

| Field | Description |
|---|---|
| `articleDir` | Directory where `.md` files are stored |
| `imageDir` | Root directory for images |
| `outputDir` | Path for the generated JSON output |
| `categories` | Whitelist of allowed categories |
| `tags` | Whitelist of allowed tags |
| `r2.endpoint` | S3-compatible endpoint URL (e.g. `https://<account_id>.r2.cloudflarestorage.com`) |
| `r2.bucketName` | R2 bucket name |
| `r2.baseUrl` | Public base URL for images (e.g. `https://assets.your-domain.com`) |

## Environment variables

| Variable | Description |
|---|---|
| `R2_ACCESS_KEY_ID` | R2 Access Key ID |
| `R2_SECRET_ACCESS_KEY` | R2 Secret Access Key |

## Directory structure created by `cmsc setup`

```
<articleDir>/
<imageDir>/
├── article/
├── category/
└── tag/
```

## R2 bucket structure after `cmsc publish`

```
<bucketName>/
├── <imageDir>/
│   ├── article/
│   ├── category/
│   └── tag/
└── <outputDir>/
    ├── all.json
    ├── category.json
    └── tag.json
```

## publish notes

- Image uploads are diff-based. Only added, modified, or deleted images are synced. Diff state is tracked in `.caches.json` (commit this file to keep CI in sync).
- Output JSON files are always fully re-uploaded on every publish.
- `Cache-Control: public, max-age=31536000, immutable` is set on all uploaded files to maximize CDN cache efficiency and minimize R2 read costs.
- R2 is recommended for its generous free tier (10 GB storage, 1M writes/month) and zero egress fees. Any S3-compatible storage can be used by changing `r2.endpoint`.

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
| `cmsc publish` | 画像の差分をCloudflare R2にアップロードし（差分は `.caches.json` で管理）、出力JSONもR2にアップロードします。`R2_ACCESS_KEY_ID` と `R2_SECRET_ACCESS_KEY` の環境変数が必要です。 |

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

# 5. R2に公開
cmsc publish
```

## 設定ファイル: `cmsc.json`

| フィールド | 説明 |
|---|---|
| `articleDir` | `.md` ファイルを格納するディレクトリ |
| `imageDir` | 画像を保存するルートディレクトリ |
| `outputDir` | 生成されるJSONの出力先パス |
| `categories` | 許可されるカテゴリのホワイトリスト |
| `tags` | 許可されるタグのホワイトリスト |
| `r2.endpoint` | S3互換エンドポイントURL（例: `https://<account_id>.r2.cloudflarestorage.com`） |
| `r2.bucketName` | R2バケット名 |
| `r2.baseUrl` | 画像配信用のベースURL（例: `https://assets.your-domain.com`） |

## 環境変数

| 変数名 | 説明 |
|---|---|
| `R2_ACCESS_KEY_ID` | R2のアクセスキーID |
| `R2_SECRET_ACCESS_KEY` | R2のシークレットアクセスキー |

## `cmsc setup` で作成されるディレクトリ構造

```
<articleDir>/
<imageDir>/
├── article/
├── category/
└── tag/
```

## `cmsc publish` 後のR2バケット構造

```
<bucketName>/
├── <imageDir>/
│   ├── article/
│   ├── category/
│   └── tag/
└── <outputDir>/
    ├── all.json
    ├── category.json
    └── tag.json
```

## publishについて

- 画像アップロードは差分のみ実行されます。追加・変更・削除されたファイルのみR2に同期されます。差分の状態は `.cache.json` で管理されるため、このファイルをコミットしておくことでCI環境でも正しく差分検知が機能します。
- 出力JSONはpublishのたびに毎回全件アップロードされます。
- アップロードするファイルには `Cache-Control: public, max-age=31536000, immutable` を付与し、CDNキャッシュを最大限活用してR2の読み込みコストを抑えます。
- ストレージはCloudflare R2を推奨します（ストレージ10GB・書き込み100万回/月の無料枠、egressコスト無料）。`r2.endpoint` を変更すればS3互換の任意のストレージも利用可能です。

</details>
