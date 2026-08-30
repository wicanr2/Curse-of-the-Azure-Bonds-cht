// ecl-address-refs 回答「這個 ECL 位址被誰碰過」與「某一段的 NEWECL 前面守著什麼」。
//
// ★ 存在的理由：段與段之間怎麼交接，是按鍵能不能一路玩下去的關鍵，而那條路
// 通常不是一條指令，是**幾個旗標同時成立**。用 grep 找不到——ECL 是位元組碼，
// 位址藏在運算元裡。
//
// ⚠ 掃的是 `ecl.TraceGraph` 跟得到的可達指令，所以結果是**下界**：跟不到的碼
// 不在裡面。要下「沒有人寫這一格」的結論時要記得這一點。
//
// 用法：
//
//	go run ./cmd/ecl-address-refs -block 2                 # 列出 NEWECL 與它前面的守衛
//	go run ./cmd/ecl-address-refs -block 2 -from 0x93 -to 0xE0
//	go run ./cmd/ecl-address-refs -block 2 -writes 4C17    # 這一段裡誰碰 4C17h
//	go run ./cmd/ecl-address-refs -all 7ED5                # 全 corpus 誰碰 7ED5h
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

// entryStarts 把一個 block 的五個生命週期入口換成走訪起點。
// 解不出來就回 nil（TraceGraph 會退回從位移 0 開始）。
func entryStarts(data []byte) []int {
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return nil
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	return starts
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "")
	member := flag.String("member", "ECL2.DAX", "")
	block := flag.Int("block", 2, "")
	before := flag.Int("before", 24, tooltext.Text("h.62cbb4675a9b"))
	writes := flag.String("writes", "", tooltext.Text("h.bc879d17d6ad"))
	all := flag.String("all", "", tooltext.Text("h.a06677743986"))
	findText := flag.String("text", "", tooltext.Text("h.57795b7720e3"))
	opcode := flag.Int("opcode", -1, tooltext.Text("h.645313783795"))
	from := flag.Int("from", -1, tooltext.Text("h.1da7839df1ed"))
	to := flag.Int("to", 0x10000, tooltext.Text("h.fd89db3c4c90"))
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	var payload []byte
	for _, file := range archive.File {
		if strings.EqualFold(file.Name, *member) {
			handle, _ := file.Open()
			payload, _ = io.ReadAll(handle)
			handle.Close()
		}
	}
	if *all != "" {
		var want uint16
		fmt.Sscanf(*all, "%x", &want)
		fmt.Print(tooltext.Format("h.9edd5e9f2bff", want))
		for _, name := range []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"} {
			var raw []byte
			for _, file := range archive.File {
				if strings.EqualFold(file.Name, name) {
					handle, _ := file.Open()
					raw, _ = io.ReadAll(handle)
					handle.Close()
				}
			}
			if raw == nil {
				continue
			}
			entries, parseErr := dax.Parse(raw)
			if parseErr != nil {
				continue
			}
			for _, entry := range entries {
				// ⚠ 起點要用**五個生命週期入口**，不能傳 nil。傳 nil 時走訪只從
				// 位移 0 開始，跟不到只有入口才進得去的那幾支——`-writes`
				// （單一 block）走的是入口，`-all` 走 nil，於是同一個位址在
				// `-writes` 找得到、在 `-all` 是 0。**那個 0 是儀器的洞，
				// 不是「沒有人碰」**，而它正好長得像可以下結論的答案。
				// 實例：`4BE6` 在 `ECL1/0x52:0109h` 有一條 `SAVE 00 4BE6`，
				// 舊版 `-all` 回 0 處。
				starts := entryStarts(entry.Data)
				graph, graphErr := ecl.TraceGraph(entry.Data, starts, 200000)
				if graphErr != nil {
					continue
				}
				for _, instruction := range graph.Instructions {
					for position, operand := range instruction.Operands {
						if operand.Code != 0x01 || operand.Word != want {
							continue
						}
						fmt.Print(tooltext.Format("h.22c881af5207", name, entry.Entry.ID, instruction.Offset, instruction.Command.Name, position, format(instruction)))
					}
				}
			}
		}
		return
	}
	blocks, err := dax.Parse(payload)
	if err != nil {
		log.Fatal(err)
	}
	var data []byte
	for _, entry := range blocks {
		if int(entry.Entry.ID) == *block {
			data = entry.Data
		}
	}
	if data == nil {
		log.Fatal(tooltext.Format("h.65e78146e93a", *block))
	}
	graph, err := ecl.TraceGraph(data, nil, 200000)
	if err != nil {
		log.Fatal(err)
	}
	list := append([]ecl.Instruction(nil), graph.Instructions...)
	sort.Slice(list, func(a, b int) bool { return list[a].Offset < list[b].Offset })
	if *writes != "" {
		var want uint16
		fmt.Sscanf(*writes, "%x", &want)
		fmt.Print(tooltext.Format("h.aaee9831d7dd", want))
		for _, instruction := range list {
			for position, operand := range instruction.Operands {
				if operand.Code != 0x01 || operand.Word != want {
					continue
				}
				fmt.Print(tooltext.Format("h.733b556cb2a2", instruction.Offset, instruction.Command.Name, position, format(instruction)))
			}
		}
		return
	}
	if *from >= 0 {
		for _, instruction := range list {
			if instruction.Offset < *from || instruction.Offset > *to {
				continue
			}
			fmt.Printf("  %04X %-12s %s\n", instruction.Offset, instruction.Command.Name, format(instruction))
		}
		return
	}
	if *findText != "" {
		for index, instruction := range list {
			hit := false
			for _, operand := range instruction.Operands {
				if operand.Code == 0x80 && strings.Contains(ecl.DecodePackedText(operand.Packed), *findText) {
					hit = true
				}
			}
			if !hit {
				continue
			}
			fmt.Print(tooltext.Format("h.664f9bd7572d", instruction.Offset))
			end := index + *before
			if end > len(list) {
				end = len(list)
			}
			for _, near := range list[index:end] {
				fmt.Printf("  %04X %-16s %s\n", near.Offset, near.Command.Name, format(near))
			}
		}
		return
	}
	if *opcode >= 0 {
		for index, instruction := range list {
			if int(instruction.Command.Opcode) != *opcode {
				continue
			}
			fmt.Print(tooltext.Format("h.fb851e9cce6f", instruction.Command.Name, instruction.Offset))
			start := index - *before
			if start < 0 {
				start = 0
			}
			for _, previous := range list[start : index+*before/2+1] {
				fmt.Printf("  %04X %-12s %s\n", previous.Offset, previous.Command.Name, format(previous))
			}
		}
		return
	}
	for index, instruction := range list {
		if instruction.Command.Opcode != 0x20 {
			continue
		}
		fmt.Print(tooltext.Format("h.472dde689eaf", instruction.Offset))
		start := index - *before
		if start < 0 {
			start = 0
		}
		for _, previous := range list[start : index+1] {
			fmt.Printf("  %04X %-12s %s\n", previous.Offset,
				previous.Command.Name, format(previous))
		}
	}
}

func format(instruction ecl.Instruction) string {
	parts := make([]string, 0, len(instruction.Operands))
	for _, operand := range instruction.Operands {
		// ★ `code=80` 是**內嵌的壓縮文字**，解出來才看得懂這一條在演什麼。
		// 只印 `low` 會得到一個沒有意義的序號。
		if operand.Code == 0x80 && len(operand.Packed) > 0 {
			text := ecl.DecodePackedText(operand.Packed)
			if len(text) > 60 {
				text = text[:60] + "…"
			}
			parts = append(parts, fmt.Sprintf("「%s」", strings.ReplaceAll(text, "\n", "⏎")))
			continue
		}
		parts = append(parts, fmt.Sprintf("[code=%02X low=%02X high=%02X word=%04X set=%v]", operand.Code, operand.Low, operand.High, operand.Word, operand.WordSet))
	}
	return strings.Join(parts, " ")
}
