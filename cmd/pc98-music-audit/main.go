// Command pc98-music-audit verifies the local PC-98 music bridge binaries.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, tooltext.Text("pc98_music_audit.usage"))
		os.Exit(2)
	}
	game, err := os.ReadFile(os.Args[1])
	if err != nil {
		fatal(err)
	}
	driver, err := os.ReadFile(os.Args[2])
	if err != nil {
		fatal(err)
	}
	report, err := pc98music.AuditBridge(game, driver)
	if err != nil {
		fatal(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
