package internal

import (
	"fmt"
	"os"
)

type SetupProjectCommand struct{}

func NewSetupProjectCommand() *SetupProjectCommand {
	return &SetupProjectCommand{}
}

func (c *SetupProjectCommand) Setup() {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	dirs := []string{
		config.ArticleDir,
		config.ImagesDir + "/article",
		config.ImagesDir + "/category",
		config.ImagesDir + "/tag",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	fmt.Println("Success! Project setup completed.")
}
