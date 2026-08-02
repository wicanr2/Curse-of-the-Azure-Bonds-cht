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
	showModules := flag.Bool("modules", false, "列出 compiler modules，不列 symbols")
	showTypes := flag.Bool("types", false, "列出 legacy type table，不列 symbols")
	showMembers := flag.Bool("members", false, "列出 legacy member table，不列 symbols")
	typeIndex := flag.Int("type-index", -1, "只列出指定的 1-based type index；同時啟用 -types")
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
		"debug_offset=0x%X version=0x%04X symbols=%d names=%d modules=%d types=%d members=%d flags=0x%02X\n",
		table.Header.FileOffset, table.Header.Version, len(table.Symbols),
		len(table.Names), table.Header.ModuleCount, len(table.Types), len(table.Members),
		table.Header.ProgramFlags,
	)
	fmt.Printf(
		"type_table=0x%X member_table=0x%X member_bytes=%d data_pool=0x%X\n",
		table.TypeTableOffset, table.MemberTableOffset, table.MemberTableSize, table.DataPoolOffset,
	)
	needle := strings.ToUpper(*match)
	matches := 0
	if *showMembers {
		for _, member := range table.Members {
			if needle != "" && !strings.Contains(strings.ToUpper(member.Name), needle) {
				continue
			}
			fmt.Printf(
				"member=%d flags=0x%02X name_index=%d name=%q type=%d\n",
				member.Index, member.Flags, member.NameIndex, member.Name, member.TypeIndex,
			)
			matches++
		}
		fmt.Printf("matches=%d\n", matches)
		return
	}
	if *showModules {
		for _, module := range table.Modules {
			if needle != "" && !strings.Contains(strings.ToUpper(module.Name), needle) {
				continue
			}
			fmt.Printf(
				"module=%d name_index=%d name=%q language=%d model_flags=0x%02X symbols=%d+%d sources=%d+%d correlations=%d+%d\n",
				module.Index, module.NameIndex, module.Name, module.Language, module.ModelFlags,
				module.SymbolIndex, module.SymbolCount, module.SourceIndex, module.SourceCount,
				module.CorrelationIndex, module.CorrelationCount,
			)
			matches++
		}
		fmt.Printf("matches=%d\n", matches)
		return
	}
	if *showTypes || *typeIndex >= 0 {
		for _, entry := range table.Types {
			if *typeIndex >= 0 && entry.Index != *typeIndex {
				continue
			}
			if needle != "" && !strings.Contains(strings.ToUpper(entry.Name), needle) {
				continue
			}
			fmt.Printf(
				"type=%d id=0x%02X name_index=%d name=%q size=0x%04X detail=%02X%02X%02X\n",
				entry.Index, entry.ID, entry.NameIndex, entry.Name, entry.Size,
				entry.Detail[0], entry.Detail[1], entry.Detail[2],
			)
			matches++
		}
		fmt.Printf("matches=%d\n", matches)
		return
	}
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
