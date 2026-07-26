package locale

import (
	"encoding/json"
	"fmt"
)

type Catalog struct {
	Language string            `json:"language"`
	Strings  map[string]string `json:"strings"`
}

func Load(data []byte) (Catalog, error) {
	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return Catalog{}, fmt.Errorf("locale JSON: %w", err)
	}
	if catalog.Language == "" || len(catalog.Strings) == 0 {
		return Catalog{}, fmt.Errorf("locale must contain language and strings")
	}
	return catalog, nil
}

func (c Catalog) Text(key, fallback string) string {
	if value, ok := c.Strings[key]; ok && value != "" {
		return value
	}
	return fallback
}
