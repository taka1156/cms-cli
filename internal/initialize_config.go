package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/taka1156/brite/internal/entity"
)

type InitializeConfigCommand struct{}

func NewInitializeConfigCommand() *InitializeConfigCommand {
	return &InitializeConfigCommand{}
}

func (c *InitializeConfigCommand) Initialize(clientConfig entity.ClientConfig) {
	configName := clientConfig.ConfigPath

	// Check if the config file already exists
	if _, err := os.Stat(configName); err == nil {
		fmt.Println("Error: brite.json already exists in this directory.")
		return
	}

	// Create default configuration
	defaultConfig := entity.BriteConfig{
		Schema:     "./brite.schema.json",
		ArticleDir: "./articles",
		ImageDir:   "./images",
		OutputDir:  "./dist",
		Categories: []string{"tech", "life", "hobby"},
		Tags:       []string{"Go", "CLI"},
	}

	jsonBytes, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		fmt.Printf("Error generating default config: %v\n", err)
		return
	}

	if err := os.MkdirAll(filepath.Dir(configName), 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	if err := os.WriteFile(configName, jsonBytes, 0644); err != nil {
		fmt.Printf("Error writing brite.json: %v\n", err)
		return
	}

	fmt.Println("Success! Created default brite.json with schema link.")
}
