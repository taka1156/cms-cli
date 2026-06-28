
<picture>
    <source media="(prefers-color-scheme: dark)" srcset="./logo-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="./logo-light.svg">
    <img alt="Brite logo" src="./logo-light.svg" width="100%" height="100%">
</picture>

[日本語版はこちら](./README.ja.md)

![GitHub Release](https://img.shields.io/github/v/release/taka1156/brite?sort=semver&display_name=release&color=60a5fa&link=https%3A%2F%2Fgithub.com%2Ftaka1156%2Fbrite%2Freleases%2F)
![GitHub Release Date](https://img.shields.io/github/release-date/taka1156/brite?color=60a5fa)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/taka1156/brite/release.yml?logo=github&color=60a5fa)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/taka1156/brite?color=60a5fa&logo=go&logoColor=white)


# brite

A minimal CLI for personal content management based on Markdown. Create, build, and publish content on your own stack.

## Commands

| Command | Description |
|---|---|
| `brite init` | Generates a default `brite.json` in the current directory. Fails if one already exists. |
| `brite setup` | Creates the directories defined in `brite.json` (subdirectories under `articleDir` and `imageDir`). |
| `brite new` | Interactively creates a new Markdown article with front matter under `articleDir`. You will be prompted for a title, category, and tags. The slug is auto-generated as a UUID (cannot be entered manually), and the date is automatically set to the current datetime. |
| `brite convert` | Reads `brite.json`, scans `.md` files in `articleDir`, parses their front matter, and writes three JSON files (`all` / `category` / `tag`) to `outputDir`. Articles with unregistered categories or tags are skipped with a notification. |
| `brite publish` | Uploads changed images to Cloudflare R2 (diffs are tracked via `.caches.json`), then uploads the output JSON files to R2 as well. Requires the environment variables `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_ENDPOINT`, and `ENABLE_PATH_STYLE_ENDPOINTS`. (Not required if you place the binary directly on the host PC or inside a Dev Container without using S3-compatible storage.) |

## Quick Start

```bash
# 1. Initialize the config file
brite init

# 2. Create directories
brite setup

# 3. Create a new article (interactive)
brite new

# 4. Output JSON
brite convert

# 5. Publish to R2
brite publish
```

## Config File: `brite.json`

| Field | Description | Required |
|---|---|---|
| `articleDir` | Directory where `.md` files are stored | true |
| `imageDir` | Root directory for storing images | true |
| `outputDir` | Output path for generated JSON files | true |
| `categories` | Whitelist of allowed categories | true |
| `tags` | Whitelist of allowed tags | true |
| `r2.bucketName` | R2 bucket name | false |
| `r2.baseUrl` | Base URL for image delivery (e.g. `https://assets.your-domain.com`) | false |

## Environment Variables

| Variable | Description |
|---|---|
| `R2_ACCESS_KEY_ID` | R2 access key ID |
| `R2_SECRET_ACCESS_KEY` | R2 secret access key |
| `R2_ENDPOINT` | S3-compatible endpoint URL (e.g. `https://<account_id>.r2.cloudflarestorage.com`) |
| `ENABLE_PATH_STYLE_ENDPOINTS` | Set to `true` if your storage requires path-style URLs (default: `false`) |

## Directory Structure Created by `brite setup`

```
<articleDir>/
<imageDir>/
├── article/
├── category/
└── tag/
```

## R2 Bucket Structure After `brite publish`

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

## Notes on Publishing

- Image uploads are incremental — only added, modified, or deleted files are synced to R2. Diff state is managed in `.caches.json`; committing this file ensures correct diff detection even in CI environments.
- Output JSON files are fully re-uploaded on every publish.
- Uploaded files are served with `Cache-Control: public, max-age=31536000, immutable` to maximize CDN caching and minimize R2 read costs.
- Cloudflare R2 is the recommended storage backend (free tier: 10 GB storage, 1M writes/month, no egress fees). Any S3-compatible storage can be used by changing `R2_ENDPOINT`.
