// Command pc98-ovr-audit validates a Turbo Pascal TPOV file against the
// resident executable and reports literal software interrupts in code only.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/borlanddebug"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98ovr"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
)

func main() {
	interruptText := flag.String("interrupt", "d2", "hexadecimal interrupt number to scan")
	wordText := flag.String("word", "", tooltext.Text("pc98_ovr_audit.word"))
	bytesText := flag.String("bytes", "", tooltext.Text("pc98_ovr_audit.bytes"))
	contextBytes := flag.Int("context", 0, tooltext.Text("pc98_ovr_audit.context"))
	soundFX := flag.Bool("soundfx", false, tooltext.Text("pc98_ovr_audit.soundfx"))
	extractCodeDir := flag.String("extract-code-dir", "", tooltext.Text("pc98_ovr_audit.extract_code"))
	resolveStubText := flag.String("resolve-stub", "", tooltext.Text("pc98_ovr_audit.resolve_stub"))
	resolveCodeText := flag.String("resolve-code", "", tooltext.Text("pc98_ovr_audit.resolve_code"))
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, tooltext.Text("pc98_ovr_audit.usage"))
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}
	if *contextBytes < 0 {
		fatalf(tooltext.Text("pc98_ovr_audit.context_nonnegative"))
	}
	interruptValue, err := strconv.ParseUint(strings.TrimPrefix(*interruptText, "0x"), 16, 8)
	if err != nil {
		fatalf(tooltext.Text("pc98_ovr_audit.interrupt_invalid"), err)
	}
	var wordValue uint64
	if *wordText != "" {
		wordValue, err = strconv.ParseUint(strings.TrimPrefix(*wordText, "0x"), 16, 16)
		if err != nil {
			fatalf(tooltext.Text("pc98_ovr_audit.word_invalid"), err)
		}
	}
	var bytePattern []byte
	if *bytesText != "" {
		bytePattern, err = hex.DecodeString(strings.TrimPrefix(*bytesText, "0x"))
		if err != nil || len(bytePattern) == 0 {
			fatalf(tooltext.Text("pc98_ovr_audit.bytes_invalid"), *bytesText)
		}
	}
	executable := read(flag.Arg(0))
	overlayFile := read(flag.Arg(1))
	overlays, err := pc98ovr.Decode(executable, overlayFile)
	if err != nil {
		fatalf(tooltext.Text("pc98_ovr_audit.decode_failed"), err)
	}
	if *resolveStubText != "" {
		overlayIndex, stubOffset, err := parseOverlayStub(*resolveStubText)
		if err != nil {
			fatalf(tooltext.Text("pc98_ovr_audit.resolve_stub_invalid"), err)
		}
		if overlayIndex < 0 || overlayIndex >= len(overlays) {
			fatalf(tooltext.Text("pc98_ovr_audit.overlay_range"), overlayIndex, len(overlays)-1)
		}
		entry, ok := overlays[overlayIndex].ResolveStub(stubOffset)
		if !ok {
			fatalf(tooltext.Text("pc98_ovr_audit.stub_missing"), overlayIndex, stubOffset)
		}
		fmt.Printf(
			"stub_resolution overlay=%d stub=0x%04X entry=%d code=0x%04X flags=0x%02X exe=0x%X\n",
			overlayIndex, stubOffset, entry.Index, entry.CodeOffset,
			entry.Flags, entry.ExecutableOffset,
		)
	}
	if *resolveCodeText != "" {
		overlayIndex, codeOffset, err := parseOverlayStub(*resolveCodeText)
		if err != nil {
			fatalf(tooltext.Text("pc98_ovr_audit.resolve_code_invalid"), err)
		}
		if overlayIndex < 0 || overlayIndex >= len(overlays) {
			fatalf(tooltext.Text("pc98_ovr_audit.overlay_range"), overlayIndex, len(overlays)-1)
		}
		entries := overlays[overlayIndex].ResolveCode(codeOffset)
		for _, entry := range entries {
			fmt.Printf(
				"code_resolution overlay=%d code=0x%04X entry=%d stub=0x%04X flags=0x%02X exe=0x%X\n",
				overlayIndex, codeOffset, entry.Index, entry.StubOffset,
				entry.Flags, entry.ExecutableOffset,
			)
		}
		if len(entries) == 0 {
			fatalf(tooltext.Text("pc98_ovr_audit.code_missing"), overlayIndex, codeOffset)
		}
	}
	var debugTable borlanddebug.Table
	if *soundFX {
		debugTable, err = borlanddebug.ParseLegacy(executable)
		if err != nil {
			fatalf(tooltext.Text("pc98_ovr_audit.symbols_failed"), err)
		}
	}

	fmt.Printf("exe_sha256=%x\n", sha256.Sum256(executable))
	fmt.Printf("ovr_sha256=%x\n", sha256.Sum256(overlayFile))
	fmt.Printf("overlay_count=%d interrupt=%02X\n", len(overlays), interruptValue)
	if *soundFX {
		table := soundFXSelectorTable()
		if bytes.Index(executable, table) < 0 {
			fatalf(tooltext.Text("pc98_ovr_audit.soundfx_table_missing"))
		}
		fmt.Printf(
			"soundfx_selector_table_files=%s ds_base=0x4838 values=255,0..15\n",
			formatOffsets(pc98ovr.PatternOffsets(executable, table)),
		)
	}
	if *extractCodeDir != "" {
		if err := os.MkdirAll(*extractCodeDir, 0o755); err != nil {
			fatalf(tooltext.Text("pc98_ovr_audit.extract_dir_failed"), err)
		}
	}
	total := 0
	soundFXTotal := 0
	soundFXCounts := make(map[int]int)
	for index, overlay := range overlays {
		offsets := pc98ovr.InterruptOffsets(overlay.Code, byte(interruptValue))
		total += len(offsets)
		wordOffsets := []int(nil)
		if *wordText != "" {
			wordOffsets = pc98ovr.WordOffsets(overlay.Code, uint16(wordValue))
		}
		byteOffsets := []int(nil)
		if len(bytePattern) != 0 {
			byteOffsets = pc98ovr.PatternOffsets(overlay.Code, bytePattern)
		}
		fmt.Printf(
			"index=%d entries=%d fixups=%d exe_offset=0x%X file_offset=0x%X code=0x%X reloc=0x%X int_offsets=%s",
			index, overlay.EntryCount, len(overlay.RelocationOffsets),
			overlay.ExecutableOffset, overlay.FileOffset,
			overlay.CodeSize, overlay.RelocationSize, formatOffsets(offsets),
		)
		if *wordText != "" {
			fmt.Printf(" word_%04X_offsets=%s", wordValue, formatOffsets(wordOffsets))
		}
		if len(bytePattern) != 0 {
			fmt.Printf(" bytes_%X_offsets=%s", bytePattern, formatOffsets(byteOffsets))
		}
		fmt.Println()
		if *contextBytes > 0 {
			for _, offset := range byteOffsets {
				start := max(0, offset-*contextBytes)
				end := min(len(overlay.Code), offset+len(bytePattern)+*contextBytes)
				fmt.Printf(
					"  bytes_context local=0x%X range=0x%X..0x%X hex=%X\n",
					offset, start, end, overlay.Code[start:end],
				)
			}
		}
		if *extractCodeDir != "" {
			path := filepath.Join(*extractCodeDir, fmt.Sprintf("overlay-%02d.bin", index))
			if err := os.WriteFile(path, overlay.Code, 0o644); err != nil {
				fatalf(tooltext.Text("pc98_ovr_audit.extract_code_failed"), index, err)
			}
		}
		if *soundFX {
			for _, call := range pc98ovr.FarCallWordArguments(overlay.Code, 0x0000, 0x0893) {
				selector, ok := soundFXSelector(call.ArgumentAddress)
				if !ok {
					fatalf(
						"overlay %d SOUNDFX call 0x%X uses unknown DS:%04X",
						index, call.CallOffset, call.ArgumentAddress,
					)
				}
				fmt.Printf(
					"soundfx_call overlay=%d module=%s function=%s local=0x%X file=0x%X ds=0x%04X selector=%d symbol=%s\n",
					index, soundFXModuleName(debugTable, index),
					soundFXFunctionName(debugTable, index, call.CallOffset),
					call.CallOffset,
					int(overlay.FileOffset)+call.CallOffset,
					call.ArgumentAddress, selector,
					soundFXSelectorName(debugTable, call.ArgumentAddress),
				)
				soundFXTotal++
				soundFXCounts[selector]++
			}
		}
	}
	fmt.Printf("interrupt_matches=%d\n", total)
	if *soundFX {
		fmt.Printf("soundfx_calls=%d selector_counts=", soundFXTotal)
		for selector := 0; selector <= 15; selector++ {
			if count := soundFXCounts[selector]; count != 0 {
				fmt.Printf("%d:%d,", selector, count)
			}
		}
		if count := soundFXCounts[255]; count != 0 {
			fmt.Printf("255:%d,", count)
		}
		fmt.Println()
	}
}

func parseOverlayStub(value string) (int, uint16, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, 0, tooltext.Error("pc98_ovr_audit.stub_format")
	}
	overlayIndex, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil || overlayIndex < 0 {
		return 0, 0, tooltext.Errorf("pc98_ovr_audit.overlay_invalid", parts[0])
	}
	stubOffset, err := strconv.ParseUint(strings.TrimPrefix(parts[1], "0x"), 16, 16)
	if err != nil {
		return 0, 0, tooltext.Errorf("pc98_ovr_audit.offset_invalid", parts[1])
	}
	return int(overlayIndex), uint16(stubOffset), nil
}

func soundFXModuleName(table borlanddebug.Table, overlayIndex int) string {
	moduleIndex := overlayIndex + 1
	if moduleIndex < 0 || moduleIndex >= len(table.Modules) {
		return "?"
	}
	return table.Modules[moduleIndex].Name
}

func soundFXFunctionName(
	table borlanddebug.Table, overlayIndex, callOffset int,
) string {
	moduleName := soundFXModuleName(table, overlayIndex)
	if moduleName == "?" {
		return "?"
	}
	loadName := "LOAD" + strings.ToUpper(moduleName)
	var segment uint16
	foundSegment := false
	for _, symbol := range table.Symbols {
		if strings.ToUpper(symbol.Name) == loadName && symbol.Segment != 0 {
			segment = symbol.Segment
			foundSegment = true
			break
		}
	}
	if !foundSegment {
		return "?"
	}
	bestName := "?"
	bestOffset := -1
	for _, symbol := range table.Symbols {
		if symbol.Segment != segment || symbol.Name == "" ||
			symbol.Offset == 0xFFFF || int(symbol.Offset) > callOffset {
			continue
		}
		if int(symbol.Offset) >= bestOffset {
			bestName = symbol.Name
			bestOffset = int(symbol.Offset)
		}
	}
	return bestName
}

func soundFXSelectorName(table borlanddebug.Table, address uint16) string {
	for _, symbol := range table.Symbols {
		if symbol.Segment == 0x0C29 && symbol.Offset == address {
			return symbol.Name
		}
	}
	return "?"
}

func soundFXSelectorTable() []byte {
	table := make([]byte, 17*2)
	for index, selector := range []uint16{
		255, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	} {
		binary.LittleEndian.PutUint16(table[index*2:], selector)
	}
	return table
}

func soundFXSelector(address uint16) (int, bool) {
	const base = uint16(0x4838)
	if address < base || address > base+16*2 || (address-base)%2 != 0 {
		return 0, false
	}
	index := int((address - base) / 2)
	if index == 0 {
		return 255, true
	}
	return index - 1, true
}

func read(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		fatalf(tooltext.Text("pc98_ovr_audit.read_failed"), path, err)
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
