// Command pc98-sfx-audit reports the PC-9801 speaker-effect program imported
// from the user's exact local GAME.EXE.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98sfx"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "用法：pc98-sfx-audit GAME.EXE")
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	game, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fatalf("讀取失敗：%v", err)
	}
	effects, err := pc98sfx.Import(game)
	if err != nil {
		fatalf("解析失敗：%v", err)
	}
	report := struct {
		GameSHA256 string           `json:"game_sha256"`
		Effects    []pc98sfx.Effect `json:"effects"`
	}{
		GameSHA256: pc98sfx.GameSHA256,
		Effects:    effects,
	}
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatalf("建立報告失敗：%v", err)
	}
	fmt.Println(string(output))
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
