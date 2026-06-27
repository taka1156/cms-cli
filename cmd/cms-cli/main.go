package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/taka1156/cms-cli/internal"
	"github.com/taka1156/cms-cli/internal/entity"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	jsonNames := entity.JsonNames{
		All:      entity.ALL_JSON_FILE_NAME,
		Category: entity.CATEGORY_JSON_FILE_NAME,
		Tag:      entity.TAG_JSON_FILE_NAME,
	}

	cmd := struct {
		*internal.InitializeConfigCommand
		*internal.SetupProjectCommand
		*internal.AddArticleCommand
		*internal.ConvertArticleCommand
		*internal.PublishArticleCommand
	}{
		internal.NewInitializeConfigCommand(),
		internal.NewSetupProjectCommand(),
		internal.NewAddArticleCommand(),
		internal.NewConvertArticleCommand(),
		internal.NewPublishArticleCommand(),
	}

	// コマンド引数のチェック
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			cmd.Initialize()
			return
		case "setup":
			cmd.Setup()
			return
		case "new":
			cmd.Add()
			return
		case "convert":
			cmd.Convert(jsonNames)
			return
		case "publish":
			cmd.Publish()
			return
		default:
			fmt.Println("Unknown command. Available commands: init, setup, new, convert, publish")
			return
		}
	} else {
		fmt.Println("No command provided. Available commands: init, setup, new, convert, publish")
	}

}
