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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	// Detail 是型別記錄尾巴那三個位元組。record（`id=30`）用它指到成員表：
	// 前兩個位元組是第一個成員的索引，第三個是成員數（spec 1164）。
	Detail [3]byte `json:"detail"`
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
	ovrPath := flag.String("ovr", "", tooltext.Text("h.657ad409401b"))
	outPath := flag.String("out", "", tooltext.Text("h.5b913c9420b2"))
	recordName := flag.String("record", "", tooltext.Text("h.45419c201b6c"))
	allRecords := flag.Bool("records", false, tooltext.Text("h.142f2364d416"))
	overlayNames := flag.Bool("overlay-names", false, tooltext.Text("h.a9a72282beb1"))
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
	if *allRecords {
		names := map[string]bool{}
		for _, entry := range table.Types {
			if entry.ID == borlanddebug.RecordTypeID && entry.Name != "" {
				names[entry.Name] = true
			}
		}
		ordered := make([]string, 0, len(names))
		for name := range names {
			ordered = append(ordered, name)
		}
		sort.Strings(ordered)
		fmt.Print(tooltext.Format("h.9574ed5b94e5"))
		fmt.Print(tooltext.Format("h.29df8fab5955"))
		for _, name := range ordered {
			fields, layoutErr := table.RecordLayout(name)
			if layoutErr != nil {
				fmt.Print(tooltext.Format("h.eb6661f0d65f", name, layoutErr))
				continue
			}
			total := fields[len(fields)-1].Offset + fields[len(fields)-1].Size
			fmt.Print(tooltext.Format("h.4ba3342e9c3d", name, len(fields), total))
			fmt.Print(tooltext.Format("h.130a40e36193"))
			for _, field := range fields {
				typeName := field.TypeName
				if typeName == "" {
					typeName = fmt.Sprintf("（id %d）", field.TypeID)
				}
				fmt.Printf("| `+%03Xh` | %d | `%s` | %s |\n", field.Offset, field.Size, field.Name, typeName)
			}
			fmt.Println()
		}
		return
	}

	if *recordName != "" {
		fields, layoutErr := table.RecordLayout(*recordName)
		if layoutErr != nil {
			fmt.Fprintf(os.Stderr, "borland-symbols: %v\n", layoutErr)
			os.Exit(1)
		}
		fmt.Print(tooltext.Format("h.a67d9420f54b", *recordName, len(fields), fields[len(fields)-1].Offset+fields[len(fields)-1].Size))
		for _, field := range fields {
			typeName := field.TypeName
			if typeName == "" {
				typeName = fmt.Sprintf("id=%d", field.TypeID)
			}
			fmt.Printf("+%03Xh %3d %-22s %s\n", field.Offset, field.Size, field.Name, typeName)
		}
		return
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
			Detail: t.Detail,
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

	if *overlayNames {
		// ★ overlay 沒有自己的名字欄位，但**裡面的符號有**：Turbo Pascal 每個
		// overlay 單元的第一個 entry 一律是 `LOADxxx`（overlay manager 的載入
		// stub），所以那個名字就是單元名。剩下的符號一起列出來當佐證。
		perModule := map[string][]string{}
		for _, symbol := range table.Symbols {
			info, ok := segmentToOverlay[symbol.Segment]
			if !ok || symbol.Name == "" {
				continue
			}
			perModule[info.module] = append(perModule[info.module], symbol.Name)
		}
		modules := make([]string, 0, len(perModule))
		for module := range perModule {
			modules = append(modules, module)
		}
		sort.Strings(modules)
		fmt.Print(tooltext.Format("h.2d7c7a33d88b"))
		fmt.Print(tooltext.Format("h.a8b04de04339"))
		fmt.Print(tooltext.Text("h.a1285930d6d0") +
			tooltext.Text("h.390bf90cd436") +
			tooltext.Text("h.4bebf5d1a587") +
			tooltext.Text("h.f32d14bd843e"))
		fmt.Print(tooltext.Format("h.b7664f5dbc3c"))
		for _, module := range modules {
			names := perModule[module]
			unit := ""
			for _, name := range names {
				if strings.HasPrefix(name, "LOAD") {
					unit = strings.TrimPrefix(name, "LOAD")
					break
				}
			}
			if unit == "" {
				unit = "—"
			}
			fmt.Printf("| `%s` | **%s** | %s |\n", module, unit, strings.Join(names, "、"))
		}
		return
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
