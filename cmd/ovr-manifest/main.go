// Command ovr-manifest 匯出一組 Turbo Pascal overlay（TPOV）容器的完整結構清冊。
//
// 它是「全模組 IDA 建庫」的前置：每個 overlay 的 entry stub handler offset 就是
// IDA 自動分析所需的 entry point 種子。沒有這份種子，raw overlay 進 IDA 只會得到
// 零個函式（實測 DOS GAME.OVR 直接載入僅 1 個函式）。
//
// 本工具只輸出結構事實（offset、size、hash、entry、fixup），不做語意判斷。
//
// 用法：
//
//	ovr-manifest -exe START.EXE -ovr GAME.OVR -platform dos \
//	  -code-dir out/dos-overlays -out out/dos-ovr-manifest.json
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98ovr"
)

type entryJSON struct {
	Index            int    `json:"index"`
	ExecutableOffset int    `json:"executable_offset"`
	StubOffset       uint16 `json:"stub_offset"`
	CodeOffset       uint16 `json:"code_offset"`
	Flags            uint8  `json:"flags"`
}

type overlayJSON struct {
	Index             int         `json:"index"`
	Module            string      `json:"module"`
	ExecutableOffset  int         `json:"executable_offset"`
	FileOffset        uint32      `json:"file_offset"`
	CodeSize          uint16      `json:"code_size"`
	RelocationSize    uint16      `json:"relocation_size"`
	EntryCount        uint16      `json:"entry_count"`
	CodeSHA256        string      `json:"code_sha256"`
	CodeFile          string      `json:"code_file,omitempty"`
	Entries           []entryJSON `json:"entries"`
	RelocationOffsets []uint16    `json:"relocation_offsets"`
}

type manifest struct {
	Schema     string        `json:"schema"`
	Platform   string        `json:"platform"`
	Executable fileJSON      `json:"executable"`
	OverlayBox fileJSON      `json:"overlay_container"`
	Overlays   []overlayJSON `json:"overlays"`
	Totals     totalsJSON    `json:"totals"`
}

type fileJSON struct {
	Path   string `json:"path"`
	Size   int    `json:"size"`
	SHA256 string `json:"sha256"`
}

type totalsJSON struct {
	Overlays  int `json:"overlays"`
	Entries   int `json:"entries"`
	CodeBytes int `json:"code_bytes"`
	Fixups    int `json:"fixups"`
}

func read(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ovr-manifest: %v\n", err)
		os.Exit(1)
	}
	return data
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func main() {
	exePath := flag.String("exe", "", "resident MZ executable（含 overlay control records）")
	ovrPath := flag.String("ovr", "", "TPOV overlay container")
	platform := flag.String("platform", "", "dos 或 pc98（只作標籤，不影響解析）")
	codeDir := flag.String("code-dir", "", "把每段 overlay code 寫成 overlay-NN.bin 的目錄")
	outPath := flag.String("out", "", "輸出 JSON 路徑（預設 stdout）")
	flag.Parse()

	if *exePath == "" || *ovrPath == "" {
		flag.Usage()
		os.Exit(2)
	}

	executable := read(*exePath)
	container := read(*ovrPath)

	overlays, err := pc98ovr.Decode(executable, container)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ovr-manifest: decode: %v\n", err)
		os.Exit(1)
	}

	if *codeDir != "" {
		if err := os.MkdirAll(*codeDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "ovr-manifest: %v\n", err)
			os.Exit(1)
		}
	}

	out := manifest{
		Schema:     "coab-ovr-manifest/1",
		Platform:   *platform,
		Executable: fileJSON{Path: *exePath, Size: len(executable), SHA256: digest(executable)},
		OverlayBox: fileJSON{Path: *ovrPath, Size: len(container), SHA256: digest(container)},
	}

	for index, overlay := range overlays {
		record := overlayJSON{
			Index:             index,
			Module:            fmt.Sprintf("overlay-%02d", index),
			ExecutableOffset:  overlay.ExecutableOffset,
			FileOffset:        overlay.FileOffset,
			CodeSize:          overlay.CodeSize,
			RelocationSize:    overlay.RelocationSize,
			EntryCount:        overlay.EntryCount,
			CodeSHA256:        digest(overlay.Code),
			RelocationOffsets: overlay.RelocationOffsets,
		}
		for _, entry := range overlay.Entries {
			record.Entries = append(record.Entries, entryJSON{
				Index:            entry.Index,
				ExecutableOffset: entry.ExecutableOffset,
				StubOffset:       entry.StubOffset,
				CodeOffset:       entry.CodeOffset,
				Flags:            entry.Flags,
			})
		}
		if *codeDir != "" {
			name := record.Module + ".bin"
			path := filepath.Join(*codeDir, name)
			if err := os.WriteFile(path, overlay.Code, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "ovr-manifest: %v\n", err)
				os.Exit(1)
			}
			record.CodeFile = path
		}
		out.Totals.Entries += len(record.Entries)
		out.Totals.CodeBytes += int(overlay.CodeSize)
		out.Totals.Fixups += len(overlay.RelocationOffsets)
		out.Overlays = append(out.Overlays, record)
	}
	out.Totals.Overlays = len(out.Overlays)

	encoded, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ovr-manifest: %v\n", err)
		os.Exit(1)
	}
	if *outPath == "" {
		os.Stdout.Write(append(encoded, '\n'))
		return
	}
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "ovr-manifest: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*outPath, append(encoded, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "ovr-manifest: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "overlays=%d entries=%d code_bytes=%d fixups=%d → %s\n",
		out.Totals.Overlays, out.Totals.Entries, out.Totals.CodeBytes, out.Totals.Fixups, *outPath)
}
