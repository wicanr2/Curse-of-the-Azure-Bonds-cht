// Command pc98-ovr-audit validates a Turbo Pascal TPOV file against the
// resident executable and reports literal software interrupts in code only.
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98ovr"
)

func main() {
	interruptText := flag.String("interrupt", "d2", "hexadecimal interrupt number to scan")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "用法：pc98-ovr-audit [選項] GAME.EXE GAME.OVR")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}
	interruptValue, err := strconv.ParseUint(strings.TrimPrefix(*interruptText, "0x"), 16, 8)
	if err != nil {
		fatalf("無效 interrupt：%v", err)
	}
	executable := read(flag.Arg(0))
	overlayFile := read(flag.Arg(1))
	overlays, err := pc98ovr.Decode(executable, overlayFile)
	if err != nil {
		fatalf("解析失敗：%v", err)
	}

	fmt.Printf("exe_sha256=%x\n", sha256.Sum256(executable))
	fmt.Printf("ovr_sha256=%x\n", sha256.Sum256(overlayFile))
	fmt.Printf("overlay_count=%d interrupt=%02X\n", len(overlays), interruptValue)
	total := 0
	for index, overlay := range overlays {
		offsets := pc98ovr.InterruptOffsets(overlay.Code, byte(interruptValue))
		total += len(offsets)
		fmt.Printf(
			"index=%d entries=%d exe_offset=0x%X file_offset=0x%X code=0x%X reloc=0x%X int_offsets=%s\n",
			index, overlay.EntryCount, overlay.ExecutableOffset, overlay.FileOffset,
			overlay.CodeSize, overlay.RelocationSize, formatOffsets(offsets),
		)
	}
	fmt.Printf("interrupt_matches=%d\n", total)
}

func read(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		fatalf("讀取 %s 失敗：%v", path, err)
	}
	return data
}

func formatOffsets(offsets []int) string {
	if len(offsets) == 0 {
		return "-"
	}
	values := make([]string, len(offsets))
	for index, offset := range offsets {
		values[index] = fmt.Sprintf("0x%X", offset)
	}
	return strings.Join(values, ",")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
