package main

import (
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"os"
	"path/filepath"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/screenshotmanifest"
)

func main() {
	manifestPath := "docs/screenshots/manifest.json"
	if len(os.Args) > 1 {
		manifestPath = os.Args[1]
	}
	manifest, err := screenshotmanifest.Load(manifestPath)
	if err != nil {
		fail(err)
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(manifestPath), "..", ".."))
	if err := screenshotmanifest.Validate(manifest, root); err != nil {
		fail(err)
	}
	fmt.Print(tooltext.Format("h.a91cf8eb75e4", len(manifest.Screenshots), len(manifest.Planned)))
}

func fail(err error) {
	fmt.Fprint(os.Stderr, tooltext.Format("h.2ae6e26c8170", err))
	os.Exit(1)
}
