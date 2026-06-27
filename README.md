
 <picture>
 	  <source media="(prefers-color-scheme: dark)" srcset="./logo-dark.svg">
	  <source media="(prefers-color-scheme: light)" srcset="./logo-light.svg">
	  <img alt="Brite logo" src="./logo-light.svg" width="100%" height="100%">
  </picture>


  ![GitHub Release](https://img.shields.io/github/v/release/taka1156/brite?sort=semver&display_name=release&color=60a5fa&link=https%3A%2F%2Fgithub.com%2Ftaka1156%2Fbrite%2Freleases%2F)
  ![GitHub Release Date](https://img.shields.io/github/release-date/taka1156/brite?color=60a5fa)
  ![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/taka1156/brite/release.yml?logo=github&color=60a5fa)
  ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/taka1156/brite?color=60a5fa&logo=go&logoColor=white)


A minimal CLI for Markdown-based personal content management. Write, build, and publish on your own stack.

<details>
<summary><strong>English</strong></summary>

## Commands

| Command | Description |
|---|---|
| `brite init` | Create a default `brite.json` config file in the current directory. Fails if one already exists. |
| `brite setup` | Create the directories defined in `brite.json` (`articleDir` and `imageDir` sub-directories). |
| `brite new` | Interactively create a new Markdown post with front matter under `articleDir`. Prompts for title, category, and tags. The slug is auto-generated as a UUID (never entered manually) and the date is filled automatically with the current timestamp. |
| `brite convert` | Read `brite.json`, walk `articleDir` for `.md` files, parse each front matter, and build/write the output JSON (`all` / `byCategory` / `byTag`) to `outputDir`. Posts with unregistered categories or tags are skipped with a notice. |
| `brite publish` | Upload changed images to Cloudflare R2 (diff-based, tracked via `.caches.json`) and upload the output JSON files to R2. Requires `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, and `R2_ENDPOINT`, `ENABLE_PATH_STYLE_ENDPOINTS` environment variables. |

## Quick Start

```bash
# 1. Initialize config
brite init

# 2. Create directories
brite setup

# 3. Create a new post (interactive)
brite new

# 4. Build the output JSON
brite convert

# 5. Publish to R2
brite publish
```

## Config file: `brite.json`

| Field | Description |
|---|---|
| `articleDir` | Directory where `.md` files are stored |
| `imageDir` | Root directory for images |
| `outputDir` | Path for the generated JSON output |
| `categories` | Whitelist of allowed categories |
| `tags` | Whitelist of allowed tags |
| `r2.bucketName` | R2 bucket name |
| `r2.baseUrl` | Public base URL for images (e.g. `https://assets.your-domain.com`) |

## Environment variables

| Variable | Description |
|---|---|
| `R2_ACCESS_KEY_ID` | R2 Access Key ID |
| `R2_SECRET_ACCESS_KEY` | R2 Secret Access Key |
| `R2_ENDPOINT` | S3-compatible endpoint URL (e.g. `https://<account_id>.r2.cloudflarestorage.com`) |
| `ENABLE_PATH_STYLE_ENDPOINTS` | Set `true` if your storage requires path-style URLs (default: `false`) |

## Directory structure created by `brite setup`

```
<articleDir>/
<imageDir>/
├── article/
├── category/
└── tag/
```

## R2 bucket structure after `brite publish`

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
- R2 is recommended for its generous free tier (10 GB storage, 1M writes/month) and zero egress fees. Any S3-compatible storage can be used by changing `R2_ENDPOINT`.

</details>

<details>
<summary><strong>日本語</strong></summary>

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `brite init` | カレントディレクトリにデフォルトの `brite.json` を生成します。既に存在する場合は失敗します。 |
| `brite setup` | `brite.json` に定義されたディレクトリ（`articleDir` および `imageDir` 配下のサブディレクトリ）を作成します。 |
| `brite new` | 対話形式でフロントマター付きの新規Markdown記事を `articleDir` 配下に作成します。タイトル・カテゴリ・タグを順に質問されます。slugはUUIDで自動採番され（手動入力不可）、dateは現在日時が自動で埋め込まれます。 |
| `brite convert` | `brite.json` を読み込み、`articleDir` 内の `.md` ファイルを走査してフロントマターを解析し、`outputDir` に `all` / `byCategory` / `byTag` を含むJSONを書き出します。未登録のカテゴリ・タグを持つ記事はスキップされ、通知が表示されます。 |
| `brite publish` | 画像の差分をCloudflare R2にアップロードし（差分は `.caches.json` で管理）、出力JSONもR2にアップロードします。`R2_ACCESS_KEY_ID`・`R2_SECRET_ACCESS_KEY`・`R2_ENDPOINT`・`ENABLE_PATH_STYLE_ENDPOINTS`の環境変数が必要です。 |

## クイックスタート

```bash
# 1. 設定ファイルを初期化
brite init

# 2. ディレクトリを作成
brite setup

# 3. 新規記事を作成（対話形式）
brite new

# 4. JSONを出力
brite convert

# 5. R2に公開
brite publish
```

## 設定ファイル: `brite.json`

| フィールド | 説明 |
|---|---|
| `articleDir` | `.md` ファイルを格納するディレクトリ |
| `imageDir` | 画像を保存するルートディレクトリ |
| `outputDir` | 生成されるJSONの出力先パス |
| `categories` | 許可されるカテゴリのホワイトリスト |
| `tags` | 許可されるタグのホワイトリスト |
| `r2.bucketName` | R2バケット名 |
| `r2.baseUrl` | 画像配信用のベースURL（例: `https://assets.your-domain.com`） |

## 環境変数

| 変数名 | 説明 |
|---|---|
| `R2_ACCESS_KEY_ID` | R2のアクセスキーID |
| `R2_SECRET_ACCESS_KEY` | R2のシークレットアクセスキー |
| `R2_ENDPOINT` | S3互換エンドポイントURL（例: `https://<account_id>.r2.cloudflarestorage.com`） |
| `ENABLE_PATH_STYLE_ENDPOINTS` | パス形式のURLが必要なストレージを使う場合は `true` を指定（デフォルト: `false`） |

## `brite setup` で作成されるディレクトリ構造

```
<articleDir>/
<imageDir>/
├── article/
├── category/
└── tag/
```

## `brite publish` 後のR2バケット構造

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

- 画像アップロードは差分のみ実行されます。追加・変更・削除されたファイルのみR2に同期されます。差分の状態は `.caches.json` で管理されるため、このファイルをコミットしておくことでCI環境でも正しく差分検知が機能します。
- 出力JSONはpublishのたびに毎回全件アップロードされます。
- アップロードするファイルには `Cache-Control: public, max-age=31536000, immutable` を付与し、CDNキャッシュを最大限活用してR2の読み込みコストを抑えます。
- ストレージはCloudflare R2を推奨します（ストレージ10GB・書き込み100万回/月の無料枠、egressコスト無料）。`R2_ENDPOINT` を変更すればS3互換の任意のストレージも利用可能です。

</details>
