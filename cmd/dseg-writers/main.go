// Command dseg-writers 回答「誰寫這個 DS 位址」——把 resident 與每個 overlay
// 掃過一遍，逐處列出**寫入**、**讀取**與**取位址**。
//
// ★ 存在的理由：這個問題在本專案反覆出現，而每次手工推都容易得到假零
// （spec 1095 的 `7EE2h`、spec 1153 的 `4BE7h` 都是這樣栽的）。把它變成工具，
// 至少讓「零」這個答案附帶一份可以檢查的掃描範圍。
//
// ⚠ **這是位元組線性掃描，不是反組譯。** 16-bit x86 是變長指令，所以立即數或
// 位移裡剛好長得像 opcode 的位元組會被誤判 ⇒ **會有偽陽性**。這是刻意選的取捨：
// 問「誰寫 X」的時候，偽陽性只要去看一眼就排除，偽陰性卻會直接變成錯誤結論。
// 走控制流的掃描（像 `cmd/ecl-cell-refs`）反過來——它掃得到的只有被認成程式碼
// 的部分，沒被認出來的整段消失。兩種掃描的失敗方向相反，所以下「不存在」的
// 結論之前兩種都要跑。
//
// ⚠ 掃不到的形式：段前綴（`es:`／`ss:`）、以及**位址在執行時算出來**的存取。
// 後者用「取位址」那一欄補：Turbo Pascal 的全域要被指標寫，位址一定得先進
// 暫存器，而那是 `mov r16, imm16` 或 `push imm16`，掃得到。
//
// 用法：
//
//	./tools/go.sh run ./cmd/dseg-writers -cells 720F,7210,7211 \
//	  -root workplace/re-sweep/dos -resident workplace/re-sweep/dos/START.EXE \
//	  -names docs/audit/overlay-module-names.md -output docs/audit/dseg-writers-720F.md
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Access 是一次存取：位元組位移、種類、以及碰到的位址。
type Access struct {
	Offset  int
	Mnemoni string
	Address uint16
}

// Kind 把三種存取分開。讀取也要列出來——`0` 這個答案只有在「有讀沒寫」與
// 「連讀都沒有」分得開的時候才有意義。
type Kind int

const (
	KindWrite Kind = iota
	KindRead
	KindAddress
)

func (k Kind) String() string {
	switch k {
	case KindWrite:
		return "write"
	case KindRead:
		return "read"
	default:
		return "address-taken"
	}
}

// aluName 是 `80h`／`81h`／`83h` 那組的 `/r` 欄。`/7` 是 `CMP`——它不寫記憶體，
// 分錯的話會憑空生出一個寫入者（本工具第一版就把 THREED 的 `cmpb` 算成寫入）。
var aluName = [8]string{"add", "or", "adc", "sbb", "and", "sub", "xor", "cmp"}

// scan 把一塊位元組裡對 wanted 的存取全部列出來。
func scan(buf []byte, wanted map[uint16]bool) map[Kind][]Access {
	out := map[Kind][]Access{}
	add := func(kind Kind, offset int, mnemonic string, address uint16) {
		if !wanted[address] {
			return
		}
		out[kind] = append(out[kind], Access{Offset: offset, Mnemoni: mnemonic, Address: address})
	}
	word := func(i int) uint16 { return uint16(buf[i]) | uint16(buf[i+1])<<8 }
	// 直接定址是 mod=00、rm=110，後面跟 disp16。
	isDisp16 := func(modrm byte) bool { return modrm&0xC7 == 0x06 }
	reg := func(modrm byte) int { return int(modrm>>3) & 7 }
	for i := 0; i+4 < len(buf); i++ {
		op := buf[i]
		next := buf[i+1]
		switch {
		case op == 0xA0 || op == 0xA1:
			add(KindRead, i, "mov acc,[addr]", word(i+1))
		case op == 0xA2 || op == 0xA3:
			add(KindWrite, i, "mov [addr],acc", word(i+1))
		case (op == 0xC6 || op == 0xC7) && next == 0x06:
			add(KindWrite, i, "mov [addr],imm", word(i+2))
		case (op == 0x88 || op == 0x89) && isDisp16(next):
			add(KindWrite, i, "mov [addr],reg", word(i+2))
		case (op == 0x8A || op == 0x8B) && isDisp16(next):
			add(KindRead, i, "mov reg,[addr]", word(i+2))
		case (op == 0x86 || op == 0x87) && isDisp16(next):
			add(KindWrite, i, "xchg", word(i+2))
		case (op == 0x84 || op == 0x85) && isDisp16(next):
			add(KindRead, i, "test", word(i+2))
		case (op == 0xFE || op == 0xFF) && (next == 0x06 || next == 0x0E):
			add(KindWrite, i, "inc/dec", word(i+2))
		case op == 0xFF && next == 0x36:
			add(KindRead, i, "push [addr]", word(i+2))
		case (op == 0x80 || op == 0x81 || op == 0x83) && isDisp16(next):
			name := aluName[reg(next)]
			kind := KindWrite
			if name == "cmp" {
				kind = KindRead
			}
			add(kind, i, name+" [addr],imm", word(i+2))
		case op < 0x40 && op&7 <= 1 && isDisp16(next):
			// 00..3D 那一整片 ALU：`op&7 == 0/1` 是「記憶體 ← 暫存器」（寫），
			// `2/3` 是「暫存器 ← 記憶體」（讀）。`38h`／`39h` 是 CMP，只讀。
			name := aluName[op>>3]
			kind := KindWrite
			if name == "cmp" {
				kind = KindRead
			}
			add(kind, i, name+" [addr],reg", word(i+2))
		case op < 0x40 && (op&7 == 2 || op&7 == 3) && isDisp16(next):
			add(KindRead, i, aluName[op>>3]+" reg,[addr]", word(i+2))
		case op >= 0xB8 && op <= 0xBF:
			add(KindAddress, i, "mov reg,imm16", word(i+1))
		case op == 0x68:
			add(KindAddress, i, "push imm16", word(i+1))
		case op == 0x8D && isDisp16(next):
			add(KindAddress, i, "lea", word(i+2))
		}
	}
	return out
}

func main() {
	cells := flag.String("cells", "", "DS addresses to census, comma-separated hex (e.g. 720F,7210,7211)")
	root := flag.String("root", "workplace/re-sweep/dos", "directory holding overlays/overlay-*.bin")
	resident := flag.String("resident", "", "resident executable (START.EXE on DOS, GAME.EXE on PC-98)")
	names := flag.String("names", "docs/audit/overlay-module-names.md", "overlay-to-unit name table; empty disables unit names")
	output := flag.String("output", "", "Markdown output path (empty prints to stdout)")
	flag.Parse()

	wanted := map[uint16]bool{}
	var ordered []uint16
	for _, text := range strings.Split(*cells, ",") {
		text = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), "0x"))
		if text == "" {
			continue
		}
		value, err := strconv.ParseUint(text, 16, 16)
		if err != nil {
			log.Fatalf("cannot parse address %q: %v", text, err)
		}
		if !wanted[uint16(value)] {
			ordered = append(ordered, uint16(value))
		}
		wanted[uint16(value)] = true
	}
	if len(wanted) == 0 {
		log.Fatal("-cells needs at least one address")
	}

	unitNames := map[string]string{}
	if *names != "" {
		unitNames = parseUnitNames(*names)
	}

	type fileResult struct {
		label   string
		unit    string
		results map[Kind][]Access
	}
	var files []fileResult
	paths, err := filepath.Glob(filepath.Join(*root, "overlays", "overlay-*.bin"))
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(paths)
	if *resident != "" {
		paths = append([]string{*resident}, paths...)
	}
	scanned := 0
	for _, path := range paths {
		buf, readErr := os.ReadFile(path)
		if readErr != nil {
			log.Fatalf("%s: %v", path, readErr)
		}
		scanned += len(buf)
		label := strings.TrimSuffix(filepath.Base(path), ".bin")
		files = append(files, fileResult{label: label, unit: unitNames[label], results: scan(buf, wanted)})
	}
	if scanned == 0 {
		log.Fatalf("read zero bytes: no overlays/overlay-*.bin under -root %q", *root)
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# Access census for `%s`\n\n", strings.Join(hexList(ordered), "`, `"))
	fmt.Fprintf(&report, "Generated by `cmd/dseg-writers`; do not hand-edit. Scanned %d files, %d bytes.\n\n",
		len(files), scanned)
	fmt.Fprintf(&report, "This is a linear byte scan, not a disassembly: it yields false positives "+
		"(immediate or displacement bytes that happen to look like an opcode) but no false negatives "+
		"for the encodings it knows. A control-flow scan fails the other way round, so run both before "+
		"concluding that nobody writes an address. Rationale and blind spots: see the `cmd/dseg-writers` "+
		"header comment and spec 1183.\n\n")

	for _, kind := range []Kind{KindWrite, KindAddress, KindRead} {
		fmt.Fprintf(&report, "## %s\n\n", kind)
		fmt.Fprintf(&report, "| file | unit | offset | form | address |\n|---|---|---|---|---|\n")
		rows := 0
		for _, file := range files {
			for _, access := range file.results[kind] {
				unit := file.unit
				if unit == "" {
					unit = "-"
				}
				fmt.Fprintf(&report, "| `%s` | %s | `%04Xh` | `%s` | `%04Xh` |\n",
					file.label, unit, access.Offset, access.Mnemoni, access.Address)
				rows++
			}
		}
		if rows == 0 {
			fmt.Fprintf(&report, "| - | - | - | - | - |\n")
		}
		fmt.Fprintf(&report, "\n%d site(s).\n\n", rows)
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	for _, kind := range []Kind{KindWrite, KindAddress, KindRead} {
		count := 0
		for _, file := range files {
			count += len(file.results[kind])
		}
		fmt.Fprintf(os.Stderr, "%s=%d ", kind, count)
	}
	fmt.Fprintln(os.Stderr)
}

func hexList(values []uint16) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, fmt.Sprintf("%04Xh", value))
	}
	return out
}

// unitNameRow 讀 `docs/audit/overlay-module-names.md` 那張表的前兩欄。
var unitNameRow = regexp.MustCompile("^\\|\\s*`(overlay-\\d+)`\\s*\\|\\s*\\*\\*([^*]+)\\*\\*\\s*\\|")

func parseUnitNames(path string) map[string]string {
	payload, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}
	}
	out := map[string]string{}
	for _, line := range strings.Split(string(payload), "\n") {
		match := unitNameRow.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		if name := strings.TrimSpace(match[2]); name != "\u2014" {
			out[match[1]] = name
		}
	}
	return out
}
