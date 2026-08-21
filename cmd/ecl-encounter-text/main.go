// ecl-encounter-text 盤點每一處 `29h ENCOUNTER MENU` 的旁白有沒有接上譯文。
//
// ★ 存在的理由：`cmd/ecl-text-coverage` 的分母裡**沒有這個 opcode 的文字**。那一支
// 走的是「印字指令連成的一頁」，`ENCOUNTER MENU` 不在它收的指令裡，所以那份稽核
// 報「未接上 0 群」的時候，這一批文字連被數到都沒有——**查無與不存在長得一模一樣**。
//
// `29h` 帶三句旁白（運算元 9、10、11），原作依距離挑一句。remake 目前只取
// **第一句非空的**（`internal/ecl/runtime.go` 的 `0x29`），所以：
//
//   - 第一句沒有規則 ⇒ 玩家真的會看到原文，是中文化缺口。
//   - 第二、三句沒有規則 ⇒ remake 現在演不到，是**還原度**的缺口，不是中文化的。
//
// 兩者分開列，不要混成一個數字。
//
// 用法：
//
//	go run ./cmd/ecl-encounter-text
//	go run ./cmd/ecl-encounter-text -out docs/audit/ecl-encounter-menu-text.md
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

	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

const (
	opEncounterMenu = 0x29
	// 旁白在運算元 9..11，跟 `internal/ecl/runtime.go` 的 `0x29` 同一組編號。
	firstPromptOperand = 9
	lastPromptOperand  = 11
)

type prompt struct {
	member int
	block  uint8
	offset int
	// distance 是運算元 2（`bank1^[580h]`，距離上限）的字面值；不是立即數就是 -1。
	distance int
	// slot 是這一句排在第幾個旁白運算元（9／10／11）。
	slot int
	// shown 為真代表 remake 真的會把這一句演出來。
	shown   bool
	text    string
	ruleID  string
	message string
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	out := flag.String("out", "docs/audit/ecl-encounter-menu-text.md", "輸出的 markdown")
	localeName := flag.String("locale", "zh-TW", "要查的語系")
	flag.Parse()

	pack, err := gamepack.Default()
	if err != nil {
		log.Fatal(err)
	}
	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	prompts := make([]prompt, 0, 256)
	for member := 1; member <= 6; member++ {
		payload := memberPayload(archive, fmt.Sprintf("ECL%d.DAX", member))
		if payload == nil {
			log.Fatalf("image 裡沒有 ECL%d.DAX", member)
		}
		blocks, err := dax.Parse(payload)
		if err != nil {
			log.Fatal(err)
		}
		for _, block := range blocks {
			found, err := scanBlock(pack, *localeName, member, block.Entry.ID, block.Data)
			if err != nil {
				log.Fatalf("ECL%d/0x%02X: %v", member, block.Entry.ID, err)
			}
			prompts = append(prompts, found...)
		}
	}
	sort.Slice(prompts, func(a, b int) bool {
		if prompts[a].member != prompts[b].member {
			return prompts[a].member < prompts[b].member
		}
		if prompts[a].block != prompts[b].block {
			return prompts[a].block < prompts[b].block
		}
		if prompts[a].offset != prompts[b].offset {
			return prompts[a].offset < prompts[b].offset
		}
		return prompts[a].slot < prompts[b].slot
	})
	if err := os.WriteFile(*out, []byte(render(prompts, *localeName)), 0o644); err != nil {
		log.Fatal(err)
	}
	shown, shownGap, hiddenGap := 0, 0, 0
	for _, item := range prompts {
		switch {
		case item.shown && item.ruleID == "":
			shown, shownGap = shown+1, shownGap+1
		case item.shown:
			shown++
		case item.ruleID == "":
			hiddenGap++
		}
	}
	fmt.Printf("旁白=%d 演得到=%d 演得到但沒接=%d 演不到且沒接=%d → %s\n",
		len(prompts), shown, shownGap, hiddenGap, *out)
}

func scanBlock(pack *goldenbox.Pack, localeName string, member int, id uint8,
	data []byte) ([]prompt, error) {
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return nil, err
	}
	starts := make([]int, 0, len(points))
	for _, point := range points {
		starts = append(starts, int(point)-ecl.CodeAddressBase)
	}
	graph, err := ecl.TraceGraph(data, starts, len(data)*8)
	if err != nil {
		return nil, err
	}
	unique := map[int]ecl.Instruction{}
	for _, instruction := range graph.Instructions {
		unique[instruction.Offset] = instruction
	}
	offsets := make([]int, 0, len(unique))
	for offset := range unique {
		offsets = append(offsets, offset)
	}
	sort.Ints(offsets)

	out := make([]prompt, 0, 8)
	for _, offset := range offsets {
		instruction := unique[offset]
		if instruction.Command.Opcode != opEncounterMenu {
			continue
		}
		if len(instruction.Operands) <= lastPromptOperand {
			continue
		}
		// remake 依距離挑一句（`ecl.EncounterPromptSlots`）。距離上限不是立即數時
		// **每一句都當成演得到**——那種情況要跑起來才知道是哪一句，靜態上不能
		// 排除任何一句，而漏判會讓缺口消失。
		distance := literalDistance(instruction.Operands[1])
		shownSlot := -1
		if distance >= 0 {
			for _, slot := range ecl.EncounterPromptSlots(distance) {
				operand := instruction.Operands[firstPromptOperand+slot]
				if strings.TrimSpace(textOf(operand)) == "" {
					continue
				}
				shownSlot = firstPromptOperand + slot
				break
			}
		}
		for slot := firstPromptOperand; slot <= lastPromptOperand; slot++ {
			text := strings.TrimSpace(textOf(instruction.Operands[slot]))
			if text == "" {
				continue
			}
			item := prompt{member: member, block: id, offset: offset, slot: slot,
				distance: distance,
				shown:    distance < 0 || slot == shownSlot, text: text}
			// ⚠ 一句一句比對，不是整條指令合起來比：三句旁白只有一句會被演出來，
			// 合起來比會讓沒接上的那兩句被第一句的規則蓋掉。
			if result := pack.MatchText([]string{text}, localeName); result.Matched {
				item.ruleID, item.message = result.RuleID, result.Message
			}
			out = append(out, item)
		}
	}
	return out, nil
}

// literalDistance 取距離上限的字面值。運算元不是立即數（是記憶體位址）時回 -1
// ——那種情況要跑起來才知道，不能靜態宣稱。
func literalDistance(operand ecl.Operand) int {
	if operand.WordSet {
		return -1
	}
	return int(operand.Low)
}

func textOf(operand ecl.Operand) string {
	if operand.Code != 0x80 || len(operand.Packed) == 0 {
		return ""
	}
	return ecl.DecodePackedText(operand.Packed)
}

func render(prompts []prompt, localeName string) string {
	var out strings.Builder
	out.WriteString("# `29h ENCOUNTER MENU` 的旁白有沒有接上譯文\n\n" +
		"由 `cmd/ecl-encounter-text` 產生，不要手改。\n\n" +
		"`29h` 帶三句旁白（運算元 9、10、11），原作依距離挑一句。remake 目前只取\n" +
		"**第一句非空的**（`internal/ecl/runtime.go` 的 `0x29`）。所以缺口分兩種：\n\n" +
		"- **演得到但沒接**：玩家真的會看到原文，是中文化缺口。\n" +
		"- **演不到且沒接**：remake 現在演不到那一句，是還原度的缺口。\n\n" +
		"⚠ 一句一句比對 `all_contains`，不是把三句合起來比——合起來比會讓沒接上的\n" +
		"那兩句被第一句的規則蓋掉。\n\n" +
		"⚠ `cmd/ecl-text-coverage` 的分母裡**沒有這個 opcode**，所以那份報告的\n" +
		"「未接上 0 群」與這一份不衝突：那裡從來沒數過這批文字。\n\n")
	out.WriteString("| 段 | 位移 | 距離上限 | 旁白 | remake 演得到 | 規則 | 原文 |\n")
	out.WriteString("|---|---|---:|---:|---|---|---|\n")
	for _, item := range prompts {
		shown := "否"
		if item.shown {
			shown = "是"
		}
		rule := "**沒有規則**"
		if item.ruleID != "" {
			rule = "`" + item.ruleID + "`"
		}
		text := item.text
		if len([]rune(text)) > 64 {
			text = string([]rune(text)[:64]) + "…"
		}
		distance := fmt.Sprintf("%d", item.distance)
		if item.distance < 0 {
			distance = "動態"
		}
		out.WriteString(fmt.Sprintf("| `ECL%d/0x%02X` | `%#04x` | %s | %d | %s | %s | %s |\n",
			item.member, item.block, item.offset, distance, item.slot, shown, rule, text))
	}
	shown, shownGap, hidden, hiddenGap, sites := 0, 0, 0, 0, map[string]bool{}
	for _, item := range prompts {
		sites[fmt.Sprintf("%d/%d/%d", item.member, item.block, item.offset)] = true
		if item.shown {
			shown++
			if item.ruleID == "" {
				shownGap++
			}
			continue
		}
		hidden++
		if item.ruleID == "" {
			hiddenGap++
		}
	}
	out.WriteString(fmt.Sprintf("\n## 摘要（語系 `%s`）\n\n| 項目 | 數 |\n|---|---:|\n"+
		"| `29h` 的處數 | %d |\n| 旁白總句數 | %d |\n"+
		"| remake 演得到的句數 | %d |\n| **演得到但沒接上譯文** | **%d** |\n"+
		"| remake 演不到的句數 | %d |\n| 其中沒接上譯文 | %d |\n",
		localeName, len(sites), len(prompts), shown, shownGap, hidden, hiddenGap))
	return out.String()
}

func memberPayload(archive *zip.ReadCloser, member string) []byte {
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, member) {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil
		}
		defer reader.Close()
		payload, err := io.ReadAll(reader)
		if err != nil {
			return nil
		}
		return payload
	}
	return nil
}
