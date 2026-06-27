package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/taka1156/cms-cli/internal/entity"
)

type InitializeConfigCommand struct{}

func NewInitializeConfigCommand() *InitializeConfigCommand {
	return &InitializeConfigCommand{}
}

// 初期化コマンドの処理
func (c *InitializeConfigCommand) Initialize() {
	configName := entity.CONFIG_FILE_NAME

	// すでにファイルがある場合は上書きを防ぐ
	if _, err := os.Stat(configName); err == nil {
		fmt.Println("Error: cmsc.json already exists in this directory.")
		return
	}

	// デフォルト設定の作成
	defaultConfig := entity.CMSConfig{
		Schema:     "./cms.schema.json",
		ArticleDir: "./articles",
		ImageDir:   "./images",
		OutputDir:  "./dist",
		Categories: []string{"tech", "life", "hobby"},
		Tags:       []string{"Go", "CLI"},
	}

	// JSONに変換
	jsonBytes, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		fmt.Printf("Error generating default config: %v\n", err)
		return
	}

	// ファイル書き出し
	if err := os.WriteFile(configName, jsonBytes, 0644); err != nil {
		fmt.Printf("Error writing cmsc.json: %v\n", err)
		return
	}

	fmt.Println("Success! Created default cmsc.json with schema link.")
}
