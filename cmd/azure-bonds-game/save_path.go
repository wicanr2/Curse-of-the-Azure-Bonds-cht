package main

import (
	"os"
	"path/filepath"
)

// defaultPartySavePath keeps mutable player data outside a possibly read-only
// AppImage, application bundle, Program Files directory, or extracted patch.
// The explicit -party-save flag remains authoritative for portable installs,
// tests, and users who deliberately keep saves beside the executable.
func defaultPartySavePath() string {
	root, err := os.UserConfigDir()
	if err != nil || root == "" {
		return "party.json"
	}
	return filepath.Join(root, "curse-of-the-azure-bonds-remake", "party.json")
}
