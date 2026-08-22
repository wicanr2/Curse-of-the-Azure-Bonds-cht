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
	imagePath := flag.String("image", "curseoftheazurebonds.zip", "原版 image ZIP")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
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
		case "固定區塊":
			fixed++
			fixedBlocks[fmt.Sprintf("%s#%d", item.member, item.payload)] = true
		case "隨機":
			random++
		case "不給物品":
			none++
		default:
			dynamic++
		}
		if item.money {
			withMoney++
		}
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# 原版在哪裡發寶物：`27h TREASURE` 逐處盤點\n\n")
	fmt.Fprintf(&report, "由 `cmd/ecl-treasure-census` 產生，不要手改。分類依據見 spec 1087 的 `27h` handler。\n\n")
	fmt.Fprintf(&report, "⚠ 這是**分母**：它說原作總共會在幾個地方發寶物，不是 remake 發對了幾處。"+
		"⚠ 只數靜態追得到的發放點，所以是**下界**。\n\n")
	fmt.Fprintf(&report, "| 分類 | 處數 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| **發放點合計** | **%d** |\n", len(sites))
	fmt.Fprintf(&report, "| 固定物品區塊（`n < 80h`）| %d |\n", fixed)
	fmt.Fprintf(&report, "| 　其中相異的 `ITEM<章>` 區塊 | %d |\n", len(fixedBlocks))
	fmt.Fprintf(&report, "| 隨機生成（`n >= 80h`）| %d |\n", random)
	fmt.Fprintf(&report, "| 不給物品（`n = 0FFh`）| %d |\n", none)
	fmt.Fprintf(&report, "| `n` 來自記憶體、靜態看不出來 | %d |\n", dynamic)
	fmt.Fprintf(&report, "| 有帶貨幣／寶石池的 | %d |\n\n", withMoney)

	fmt.Fprintf(&report, "| ECL 檔 | 段 | 位移 | 物品 | 帶錢 |\n|---|---:|---:|---|---|\n")
	for _, item := range sites {
		detail := item.kind
		switch item.kind {
		case "固定區塊":
			detail = fmt.Sprintf("固定區塊 %d", item.payload)
		case "隨機":
			detail = fmt.Sprintf("隨機 %d 件", item.payload)
		}
		money := "—"
		if item.money {
			money = "是"
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
		item := site{member: member, block: block, offset: instruction.Offset, kind: "動態"}
		if len(instruction.Operands) >= 8 {
			if value, ok := constantOperand(instruction.Operands[7]); ok {
				switch {
				case value == 0xFF:
					item.kind = "不給物品"
				case value >= 0x80:
					item.kind, item.payload = "隨機", int(value-0x80)
				default:
					item.kind, item.payload = "固定區塊", int(value)
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
