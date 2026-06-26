package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/taka1156/cms-cli/internal/entity"
)

// json を読み込むだけの共通処理
func loadJson[T entity.CMSConfig | []entity.ImageCache](path string) (T, error) {
	var config T

	configFile, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("%s not found. Run './cmsc init' to create a default configuration", path)
	}

	if err := json.Unmarshal(configFile, &config); err != nil {
		return config, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return config, nil
}
