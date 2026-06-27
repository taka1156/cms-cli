package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/taka1156/brite/internal/entity"
)

func loadJson[T entity.BriteConfig | []entity.ImageCache](path string) (T, error) {
	var config T

	configFile, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("%s not found. Run 'brite init' to create a default configuration", path)
	}

	if err := json.Unmarshal(configFile, &config); err != nil {
		return config, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return config, nil
}
