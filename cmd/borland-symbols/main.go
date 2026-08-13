// Command borland-symbols 把 16-bit Borland／Turbo Pascal 除錯資訊匯出成 JSON，
// 並在有 TPOV overlay 容器時，把每個符號的 segment 對應回 overlay module。
//
// 這是 PC-98 側的語意骨幹：PC-98 `GAME.EXE` 保留了完整的 module／symbol／type
// 表，符號位址的 segment 等於 overlay control segment、offset 等於 overlay-local
// code offset（本工具會逐筆檢查 offset < code_size 才標記歸屬）。DOS `START.EXE`
// 沒有這張表，因此 DOS 的命名只能靠與 PC-98 的結構對應另行證明。
//
// 用法：
//
//	borland-symbols -exe PC98-GAME.EXE [-ovr PC98-GAME.OVR] -out out/pc98-symbols.json
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/borlanddebug"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98ovr"
)

type symbolJSON struct {
	Index     int    `json:"index"`
	Name      string `json:"name"`
	Segment   uint16 `json:"segment"`
	Offset    uint16 `json:"offset"`
	TypeIndex uint16 `json:"type_index"`
	Flags     byte   `json:"flags"`
	Module    string `json:"overlay_module,omitempty"`
	InCode    bool   `json:"overlay_code_offset,omitempty"`
}

type typeJSON struct {
	Index     int    `json:"index"`
	ID        byte   `json:"id"`
	Name      string `json:"name"`
	Size      uint16 `json:"size"`
	NameIndex uint16 `json:"name_index"`
}

type memberJSON struct {
	Index     int    `json:"index"`
	Name      string `json:"name"`
	TypeIndex uint16 `json:"type_index"`
	Flags     byte   `json:"flags"`
}

type moduleJSON struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	Language byte   `json:"language"`
}

type output struct {
	Schema     string       `json:"schema"`
	Executable string       `json:"executable"`
	SHA256     string       `json:"sha256"`
	DebugInfo  headerJSON   `json:"debug_header"`
	Modules    []moduleJSON `json:"modules"`
	Symbols    []symbolJSON `json:"symbols"`
	Types      []typeJSON   `json:"types"`
	Members    []memberJSON `json:"members"`
	Segments   []segJSON    `json:"overlay_segments"`
}

type headerJSON struct {
	FileOffset  int    `json:"file_offset"`
	Version     uint16 `json:"version"`
	SymbolCount uint16 `json:"symbol_count"`
	ModuleCount uint16 `json:"module_count"`
	TypeCount   uint16 `json:"type_count"`
	MemberCount uint16 `json:"member_count"`
}

type segJSON struct {
	Segment  uint16 `json:"segment"`
	Module   string `json:"module"`
	CodeSize uint16 `json:"code_size"`
	Symbols  int    `json:"symbols"`
}

func main() {
	exePath := flag.String("exe", "", "resident MZ executable")
	ovrPath := flag.String("ovr", "", "TPOV overlay 容器（可省略；提供才做 segment→overlay 歸屬）")
	outPath := flag.String("out", "", "輸出 JSON（預設 stdout）")
	flag.Parse()
	if *exePath == "" {
		flag.Usage()
		os.Exit(2)
	}

	executable, err := os.ReadFile(*exePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "borland-symbols: %v\n", err)
		os.Exit(1)
	}
	table, err := borlanddebug.ParseLegacy(executable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "borland-symbols: %v\n", err)
		os.Exit(1)
	}
	// MZ header 的 paragraph 數決定 load image 起點；overlay control 的 file
	// offset 減掉它才是 resident segment（spec 412 的 08B0h → 段 008Bh）。
	headerParagraphs := int(uint32(executable[8]) | uint32(executable[9])<<8)
	loadImageStart := headerParagraphs * 16

	sum := sha256.Sum256(executable)
	out := output{
		Schema:     "coab-borland-symbols/1",
		Executable: filepath.Base(*exePath),
		SHA256:     hex.EncodeToString(sum[:]),
		DebugInfo: headerJSON{
			FileOffset:  table.Header.FileOffset,
			Version:     table.Header.Version,
			SymbolCount: table.Header.SymbolCount,
			ModuleCount: table.Header.ModuleCount,
			TypeCount:   table.Header.TypeCount,
			MemberCount: table.Header.MemberCount,
		},
	}
	for _, module := range table.Modules {
		out.Modules = append(out.Modules, moduleJSON{
			Index: module.Index, Name: module.Name, Language: module.Language,
		})
	}
	for _, t := range table.Types {
		out.Types = append(out.Types, typeJSON{
			Index: t.Index, ID: t.ID, Name: t.Name, Size: t.Size, NameIndex: t.NameIndex,
		})
	}
	for _, m := range table.Members {
		out.Members = append(out.Members, memberJSON{
			Index: m.Index, Name: m.Name, TypeIndex: m.TypeIndex, Flags: m.Flags,
		})
	}

	type overlayInfo struct {
		module   string
		codeSize uint16
	}
	segmentToOverlay := map[uint16]overlayInfo{}
	if *ovrPath != "" {
		container, err := os.ReadFile(*ovrPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "borland-symbols: %v\n", err)
			os.Exit(1)
		}
		overlays, err := pc98ovr.Decode(executable, container)
		if err != nil {
			fmt.Fprintf(os.Stderr, "borland-symbols: decode: %v\n", err)
			os.Exit(1)
		}
		for index, overlay := range overlays {
			offset := overlay.ExecutableOffset - loadImageStart
			if offset < 0 || offset%16 != 0 {
				continue
			}
			segmentToOverlay[uint16(offset/16)] = overlayInfo{
				module:   fmt.Sprintf("overlay-%02d", index),
				codeSize: overlay.CodeSize,
			}
		}
	}

	perSegment := map[uint16]int{}
	for _, symbol := range table.Symbols {
		record := symbolJSON{
			Index: symbol.Index, Name: symbol.Name, Segment: symbol.Segment,
			Offset: symbol.Offset, TypeIndex: symbol.TypeIndex, Flags: symbol.Flags,
		}
		if info, ok := segmentToOverlay[symbol.Segment]; ok {
			record.Module = info.module
			// 只有落在 code span 內才可宣稱是 overlay-local code offset。
			record.InCode = symbol.Offset < info.codeSize
			perSegment[symbol.Segment]++
		}
		out.Symbols = append(out.Symbols, record)
	}
	for segment, info := range segmentToOverlay {
		out.Segments = append(out.Segments, segJSON{
			Segment: segment, Module: info.module, CodeSize: info.codeSize,
			Symbols: perSegment[segment],
		})
	}

	encoded, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "borland-symbols: %v\n", err)
		os.Exit(1)
	}
	if *outPath == "" {
		os.Stdout.Write(append(encoded, '\n'))
		return
	}
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "borland-symbols: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*outPath, append(encoded, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "borland-symbols: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "symbols=%d modules=%d types=%d overlay_segments=%d → %s\n",
		len(out.Symbols), len(out.Modules), len(out.Types), len(out.Segments), *outPath)
}
