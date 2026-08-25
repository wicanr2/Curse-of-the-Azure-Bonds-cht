// Command ecl-treasure-census 盤點 corpus 裡每一處 `27h TREASURE`。
//
// ★ 存在的理由：「全城市／全房間走訪」還沒有分母的三塊之一就是**全部寶物**。
// 這一支把它變成數字：發放點幾處、其中幾處是固定的物品區塊、幾處是隨機生成、
// 幾處只給錢不給物品。
//
// 分類依據（spec 1087 的 `27h` handler）：第八個運算元 `n` 分三段——
//
//	< 80h   從 `ITEM<章>.dax` 讀進第 n 塊（**固定內容**，逐筆 63 bytes 串進物品鏈）
//	80h..   隨機產生 `n − 80h` 件
//	0FFh    不給物品
//
// 前六個運算元是七種貨幣／寶石／珠寶的池（`DS:6F70h + i × 4`，spec 1059）。
//
// ⚠ 這是**分母**：它說「原作總共會在幾個地方發寶物」，不是「remake 發對了幾處」。
// 後者要逐處比對內容，這一支不宣稱。
//
// ⚠ 只數靜態追得到的。`TraceGraph` 走不到的程式碼裡若還有發放點，這裡看不到
// ——所以是**下界**。
//
// 用法：
//
//	go run ./cmd/ecl-treasure-census -output docs/audit/ecl-treasure-census.md
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

type site struct {
	member  string
	block   uint8
	offset  int
	kind    string
	payload int
	money   bool
}

func main() {
	imagePath := flag.String("image", "curseoftheazurebonds.zip", tooltext.Text("h.79f855c8b433"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
	flag.Parse()

	archive, err := zip.OpenReader(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	sites := make([]site, 0, 128)
	for member := 1; member <= 6; member++ {
		name := fmt.Sprintf("ECL%d.DAX", member)
		payload := zipMember(&archive.Reader, name)
		if payload == nil {
			continue
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			log.Fatalf("%s：%v", name, parseErr)
		}
		for _, block := range blocks {
			sites = append(sites, scan(name, block.Entry.ID, block.Data)...)
		}
	}
	sort.Slice(sites, func(i, j int) bool {
		if sites[i].member != sites[j].member {
			return sites[i].member < sites[j].member
		}
		if sites[i].block != sites[j].block {
			return sites[i].block < sites[j].block
		}
		return sites[i].offset < sites[j].offset
	})

	fixed, random, none, dynamic, withMoney := 0, 0, 0, 0, 0
	fixedBlocks := map[string]bool{}
	for _, item := range sites {
		switch item.kind {
		case tooltext.Text("h.77507364a169"):
			fixed++
			fixedBlocks[fmt.Sprintf("%s#%d", item.member, item.payload)] = true
		case tooltext.Text("h.3941bb44df5c"):
			random++
		case tooltext.Text("h.64e4db8efc50"):
			none++
		default:
			dynamic++
		}
		if item.money {
			withMoney++
		}
	}

	var report strings.Builder
	fmt.Fprint(&report, tooltext.Format("h.4e6bdfcbfd87"))
	fmt.Fprint(&report, tooltext.Format("h.c1ae1cbd9d02"))
	fmt.Fprint(&report, tooltext.Text("h.57bcbbd61eed")+
		tooltext.Text("h.d1f38f832155"))
	fmt.Fprint(&report, tooltext.Format("h.0eff7af55886"))
	fmt.Fprint(&report, tooltext.Format("h.7872c79e9548", len(sites)))
	fmt.Fprint(&report, tooltext.Format("h.7c6b37f45270", fixed))
	fmt.Fprint(&report, tooltext.Format("h.834f66664793", len(fixedBlocks)))
	fmt.Fprint(&report, tooltext.Format("h.651e662f3007", random))
	fmt.Fprint(&report, tooltext.Format("h.7acfd3a53f4a", none))
	fmt.Fprint(&report, tooltext.Format("h.ff7d418f0d40", dynamic))
	fmt.Fprint(&report, tooltext.Format("h.8224f938c4bd", withMoney))

	fmt.Fprint(&report, tooltext.Format("h.8077481925b6"))
	for _, item := range sites {
		detail := item.kind
		switch item.kind {
		case tooltext.Text("h.77507364a169"):
			detail = tooltext.Format("h.bb934e321868", item.payload)
		case tooltext.Text("h.3941bb44df5c"):
			detail = tooltext.Format("h.aed4460fb32d", item.payload)
		}
		money := "—"
		if item.money {
			money = tooltext.Text("h.b5141d3d19e9")
		}
		fmt.Fprintf(&report, "| `%s` | `0x%02X` | `%04Xh` | %s | %s |\n",
			item.member, item.block, item.offset, detail, money)
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "sites=%d fixed=%d blocks=%d random=%d none=%d dynamic=%d money=%d\n",
		len(sites), fixed, len(fixedBlocks), random, none, dynamic, withMoney)
}

func scan(member string, block uint8, data []byte) []site {
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return nil
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return nil
	}
	found := make([]site, 0, 4)
	seen := map[int]bool{}
	for _, instruction := range graph.Instructions {
		if instruction.Command.Opcode != 0x27 || seen[instruction.Offset] {
			continue
		}
		seen[instruction.Offset] = true
		item := site{member: member, block: block, offset: instruction.Offset, kind: tooltext.Text("h.fcca5a4d1c91")}
		if len(instruction.Operands) >= 8 {
			if value, ok := constantOperand(instruction.Operands[7]); ok {
				switch {
				case value == 0xFF:
					item.kind = tooltext.Text("h.64e4db8efc50")
				case value >= 0x80:
					item.kind, item.payload = tooltext.Text("h.3941bb44df5c"), int(value-0x80)
				default:
					item.kind, item.payload = tooltext.Text("h.77507364a169"), int(value)
				}
			}
			// 貨幣／寶石池是**第 2..7 個**運算元，不是前六個。
			//
			// ⚠ spec 1087 寫的是 `for i := 1 to 6 do dword[6F70h + i×4] := ADDRESSVALUE(i+1)`
			// ——`ADDRESSVALUE` 的引數是 **1 起算**，所以池子是運算元 2..7、
			// 0 起算就是 `[1..6]`。照「前六個」抄會把第 1 個（不是池子）算進來、
			// 漏掉第 7 個，而兩邊都不會報錯，只是「有沒有帶錢」這一欄悄悄錯。
			for index := 1; index <= 6 && index < len(instruction.Operands); index++ {
				if value, ok := constantOperand(instruction.Operands[index]); ok && value != 0 {
					item.money = true
					break
				}
			}
		}
		found = append(found, item)
	}
	return found
}

func constantOperand(operand ecl.Operand) (uint16, bool) {
	switch operand.Code {
	case 0x00:
		return uint16(operand.Low), true
	case 0x02:
		if !operand.WordSet {
			return 0, false
		}
		return operand.Word, true
	default:
		return 0, false
	}
}

func zipMember(archive *zip.Reader, name string) []byte {
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, name) {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer handle.Close()
		payload, readErr := io.ReadAll(handle)
		if readErr != nil {
			log.Fatal(readErr)
		}
		return payload
	}
	return nil
}
