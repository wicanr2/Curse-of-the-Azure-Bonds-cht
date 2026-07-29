// Command pc98-symbol-audit reports selected symbols from legacy Borland
// debug information appended to a DOS MZ executable.
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/borlanddebug"
)

func main() {
	match := flag.String("match", "", "不分大小寫的名稱子字串；空白表示全部")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "用法：pc98-symbol-audit [選項] GAME.EXE")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	executable, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fatalf("讀取失敗：%v", err)
	}
	table, err := borlanddebug.ParseLegacy(executable)
	if err != nil {
		fatalf("解析失敗：%v", err)
	}
	fmt.Printf("exe_sha256=%x\n", sha256.Sum256(executable))
	fmt.Printf(
		"debug_offset=0x%X version=0x%04X symbols=%d names=%d modules=%d flags=0x%02X\n",
		table.Header.FileOffset, table.Header.Version, len(table.Symbols),
		len(table.Names), table.Header.ModuleCount, table.Header.ProgramFlags,
	)
	needle := strings.ToUpper(*match)
	matches := 0
	for _, symbol := range table.Symbols {
		if symbol.Name == "" || (needle != "" && !strings.Contains(strings.ToUpper(symbol.Name), needle)) {
			continue
		}
		fmt.Printf(
			"index=%d name_index=%d name=%q address=%04X:%04X type=%d flags=0x%02X\n",
			symbol.Index, symbol.NameIndex, symbol.Name, symbol.Segment,
			symbol.Offset, symbol.TypeIndex, symbol.Flags,
		)
		matches++
	}
	fmt.Printf("matches=%d\n", matches)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
