package internal

import "fmt"

const HELP_MSG = `brite - A minimal CLI for personal content management.

Usage:
  brite <command> [options]

Commands:
  init      Generates a default brite.json in the current directory.
            Fails if one already exists.

  setup     Creates the directories defined in brite.json
            (subdirectories under articleDir and imageDir).

  new       Interactively creates a new Markdown article with front matter
            under articleDir. You will be prompted for a title, category,
            and tags. The slug is auto-generated as a UUID (cannot be
            entered manually), and the date is automatically set to the
            current datetime.

  convert   Reads brite.json, scans .md files in articleDir, parses their
            front matter, and writes three JSON files (all / category / tag)
            to outputDir. Articles with unregistered categories or tags are
            skipped with a notification.

  publish   Uploads changed images to Cloudflare R2 (diffs are tracked via
            .caches.json), then uploads the output JSON files to R2 as well.
            Requires R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_ENDPOINT,
            and ENABLE_PATH_STYLE_ENDPOINTS environment variables.

Options:
  --config  Path to the config file (default: brite.json)

Examples:
  brite init

  single project workflow:
	brite setup
	brite new
	brite publish

  multiple project workflow:
	brite init --config brite.blog.json
	brite setup --config brite.portfolio.json
	brite new --config brite.blog.json
	brite convert --config brite.portfolio.json
	brite publish --config brite.blog.json`

type HelpBriteCommand struct{}

func NewHelpBriteCommand() *HelpBriteCommand {
	return &HelpBriteCommand{}
}

func (c *HelpBriteCommand) Help() {
	fmt.Println(HELP_MSG)
}
