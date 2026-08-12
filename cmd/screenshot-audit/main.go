package main

import (
	"fmt"
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
	fmt.Printf("截圖 manifest 驗證通過：%d 張現有圖片、%d 項 planned 缺口\n", len(manifest.Screenshots), len(manifest.Planned))
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "截圖 manifest 驗證失敗：%v\n", err)
	os.Exit(1)
}
