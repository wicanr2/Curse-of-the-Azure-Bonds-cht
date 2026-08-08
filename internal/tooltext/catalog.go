// Package tooltext provides data-driven Traditional Chinese text for
// repository-local diagnostic commands.
package tooltext

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
)

//go:embed messages/zh-TW.json
var localeData []byte

var messages = loadMessages()

func loadMessages() map[string]string {
	var catalog map[string]string
	if err := json.Unmarshal(localeData, &catalog); err != nil {
		panic(fmt.Sprintf("tool text JSON: %v", err))
	}
	if len(catalog) == 0 {
		panic("tool text JSON is empty")
	}
	return catalog
}

// Text returns the localized message identified by id. A missing ID is a
// programming error: silently displaying an ID would hide an incomplete
// migration of a diagnostic command.
func Text(id string) string {
	value, ok := messages[id]
	if !ok {
		panic("missing tool text ID: " + id)
	}
	return value
}

// Format expands a localized message using the format contract stored in the
// locale catalog.
func Format(id string, args ...any) string {
	return fmt.Sprintf(Text(id), args...)
}

// Error creates a localized error without treating its text as a format
// string.
func Error(id string) error {
	return errors.New(Text(id))
}

// Errorf creates a localized formatted error and preserves fmt.Errorf's %w
// wrapping behavior for callers that pass an underlying error.
func Errorf(id string, args ...any) error {
	return fmt.Errorf(Text(id), args...)
}
