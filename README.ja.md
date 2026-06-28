
<picture>
    <source media="(prefers-color-scheme: dark)" srcset="./logo-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="./logo-light.svg">
    <img alt="Brite logo" src="./logo-light.svg" width="100%" height="100%">
</picture>

[English version](./README.md)

![GitHub Release](https://img.shields.io/github/v/release/taka1156/brite?sort=semver&display_name=release&color=60a5fa&link=https%3A%2F%2Fgithub.com%2Ftaka1156%2Fbrite%2Freleases%2F)
![GitHub Release Date](https://img.shields.io/github/release-date/taka1156/brite?color=60a5fa)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/taka1156/brite/release.yml?logo=github&color=60a5fa)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/taka1156/brite?color=60a5fa&logo=go&logoColor=white)

# brite

Markdown ベースの個人向けコンテンツ管理のための、最小限の CLI。独自のスタック上で、コンテンツの作成、ビルド、公開を行うことができます。

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `brite init` | カレントディレクトリにデフォルトの `brite.json` を生成します。既に存在する場合は失敗します。 |
| `brite setup` | `brite.json` に定義されたディレクトリ（`articleDir` および `imageDir` 配下のサブディレクトリ）を作成します。 |
| `brite new` | 対話形式でフロントマター付きの新規Markdown記事を `articleDir` 配下に作成します。タイトル・カテゴリ・タグを順に質問されます。slugはUUIDで自動採番され（手動入力不可）、dateは現在日時が自動で埋め込まれます。 |
| `brite convert` | `brite.json` を読み込み、`articleDir` 内の `.md` ファイルを走査してフロントマターを解析し、`outputDir` に `all` / `category` / `tag` の三つのJSONを書き出します。未登録のカテゴリ・タグを持つ記事はスキップされ、通知が表示されます。 |
| `brite publish` | 画像の差分をCloudflare R2にアップロードし（差分は `.caches.json` で管理）、出力JSONもR2にアップロードします。`R2_ACCESS_KEY_ID`・`R2_SECRET_ACCESS_KEY`・`R2_ENDPOINT`・`ENABLE_PATH_STYLE_ENDPOINTS`の環境変数が必要です。(このバイナリをホストPCやDev container内においてプロジェクト内で直接配置して運用するなどS3互換ストレージを使わない場合は、不要です。) |

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
| `articleDir` | `.md` ファイルを格納するディレクトリ | true |
| `imageDir` | 画像を保存するルートディレクトリ | true |
| `outputDir` | 生成されるJSONの出力先パス | true |
| `categories` | 許可されるカテゴリのホワイトリスト | true |
| `tags` | 許可されるタグのホワイトリスト | true |
| `r2.bucketName` | R2バケット名 | false |
| `r2.baseUrl` | 画像配信用のベースURL（例: `https://assets.your-domain.com`） | false |

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
