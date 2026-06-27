package internal

import (
	"fmt"
	"os"

	"github.com/taka1156/brite/internal/entity"
)

type SetupProjectCommand struct{}

func NewSetupProjectCommand() *SetupProjectCommand {
	return &SetupProjectCommand{}
}

func (c *SetupProjectCommand) Setup() {
	config, err := loadJson[entity.BriteConfig](entity.CONFIG_FILE_NAME)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	dirs := []string{
		config.ArticleDir,
		config.ImageDir + "/article",
		config.ImageDir + "/category",
		config.ImageDir + "/tag",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	fmt.Println("Success! Project setup completed.")
}
