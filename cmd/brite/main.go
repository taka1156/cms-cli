package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/taka1156/brite/internal"
	"github.com/taka1156/brite/internal/entity"
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
		*internal.HelpBriteCommand
		*internal.InitializeConfigCommand
		*internal.SetupProjectCommand
		*internal.AddArticleCommand
		*internal.ConvertArticleCommand
		*internal.PublishArticleCommand
	}{
		internal.NewHelpBriteCommand(),
		internal.NewInitializeConfigCommand(),
		internal.NewSetupProjectCommand(),
		internal.NewAddArticleCommand(),
		internal.NewConvertArticleCommand(),
		internal.NewPublishArticleCommand(),
	}

	switch os.Args[1] {
	case "help":
		cmd.Help()
		return
	case "init":
		absPath, err := parseConfigPath("init", os.Args[2:])
		if err != nil {
			fmt.Printf("Error resolving config path: %v\n", err)
			return
		}
		cmd.Initialize(entity.ClientConfig{ConfigPath: absPath})
	case "setup":
		absPath, err := parseConfigPath("setup", os.Args[2:])
		if err != nil {
			fmt.Printf("Error resolving config path: %v\n", err)
			return
		}
		cmd.Setup(entity.ClientConfig{ConfigPath: absPath})
	case "new":
		absPath, err := parseConfigPath("new", os.Args[2:])
		if err != nil {
			fmt.Printf("Error resolving config path: %v\n", err)
			return
		}
		cmd.Add(entity.ClientConfig{ConfigPath: absPath})
	case "convert":
		absPath, err := parseConfigPath("convert", os.Args[2:])
		if err != nil {
			fmt.Printf("Error resolving config path: %v\n", err)
			return
		}
		cmd.Convert(entity.ClientConfig{ConfigPath: absPath}, jsonNames)
	case "publish":
		absPath, err := parseConfigPath("publish", os.Args[2:])
		if err != nil {
			fmt.Printf("Error resolving config path: %v\n", err)
			return
		}
		cmd.Publish(entity.ClientConfig{ConfigPath: absPath})
	default:
		fmt.Println("Unknown command. Available commands: init, setup, new, convert, publish")
		return
	}
}

func parseConfigPath(cmdName string, args []string) (string, error) {
	fs := flag.NewFlagSet(cmdName, flag.ExitOnError)
	configPath := fs.String("config-path", entity.CONFIG_FILE_NAME, "path to brite.json")
	fs.Parse(args)
	return filepath.Abs(*configPath)
}
