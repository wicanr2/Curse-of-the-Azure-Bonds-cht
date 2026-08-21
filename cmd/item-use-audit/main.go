// Command item-use-audit 把原版每一件物品照「戰鬥裡按 USE 會走哪一條路」分類。
//
// ★ 存在的理由：戰鬥選單有 `使用`，但 remake 沒有 USE 的動作本身。
// 「還缺多少」不能憑印象——這一支把它變成可以引用的數字：全部物品幾件、
// 其中卷軸幾件、充能物品幾件，以及那些充能效果有沒有在 game pack 裡宣告成
// 可施放的法術（沒有宣告就沒有目標模式，接不上既有的效果施法路）。
//
// 分類依據（spec 921／1168）：
//
//	卷軸    物品類別 3Ch..3Eh。三個效果槽是三個法術，選一個來唸。
//	充能    `+3Dh > 0` 且 `+3Eh < 80h` 且不是卷軸。效果 ＝ `+3Dh and 7Fh`。
//	其他    按下去什麼都不會發生。
//
// 用法：
//
//	go run ./cmd/item-use-audit -output docs/audit/combat-item-use.md
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

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// scrollTypes 是三種卷軸的物品類別。
//
// ⚠ 這是**從資料推的**：原作判斷「是不是卷軸」的那一支（`loc_11DBh+3`）沒有
// 讀出來。判準是「名字含 Scroll」⟺「類別在這三個裡」——本工具**逐件檢查這個
// 等價關係**，對不上就整支失敗（spec 1168）。
var scrollTypes = map[uint8]bool{0x3C: true, 0x3D: true, 0x3E: true}

type chargedItem struct {
	Chapter  int
	Block    uint8
	Name     string
	Type     uint8
	Charges  uint8
	Effect   uint8
	Declared bool
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	pack, err := gamepack.Default()
	if err != nil {
		log.Fatal(err)
	}
	declared := map[uint8]bool{}
	for _, spell := range pack.CombatPlayerSpells {
		declared[spell.SpellID] = true
	}

	total, scrolls, inert := 0, 0, 0
	charged := make([]chargedItem, 0, 32)
	for chapter := 1; chapter <= 6; chapter++ {
		payload := member(archive, fmt.Sprintf("ITEM%d.DAX", chapter))
		if payload == nil {
			continue
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			log.Fatalf("ITEM%d.DAX: %v", chapter, parseErr)
		}
		for _, block := range blocks {
			// ⚠ 原版 DAX 一律走原版解碼（Big5）：走 remake 那條會把中文物品名讀成亂碼
			// （spec 1121）。
			items, itemErr := monster.ParseOriginalItems(block.Data)
			if itemErr != nil {
				continue
			}
			for _, item := range items {
				total++
				switch {
				case scrollTypes[item.Type]:
					scrolls++
				case item.Affects[1] > 0 && item.Affects[2] < 0x80:
					effect := item.Affects[1] & 0x7F
					charged = append(charged, chargedItem{
						Chapter: chapter, Block: block.Entry.ID, Name: item.Name,
						Type: item.Type, Charges: item.Affects[0], Effect: effect,
						Declared: declared[effect],
					})
				default:
					inert++
				}
			}
		}
	}
	sort.SliceStable(charged, func(left, right int) bool {
		if charged[left].Effect != charged[right].Effect {
			return charged[left].Effect < charged[right].Effect
		}
		return charged[left].Name < charged[right].Name
	})

	var report strings.Builder
	fmt.Fprintf(&report, "# 戰鬥裡按 USE 會走到哪：原版物品逐件分類\n\n")
	fmt.Fprintf(&report, "由 `cmd/item-use-audit` 產生，不要手改。分類依據見 spec 1168。\n\n")
	fmt.Fprintf(&report, "| 類別 | 件數 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 卷軸（類別 `3Ch`..`3Eh`）| %d |\n", scrolls)
	fmt.Fprintf(&report, "| 充能物品 | %d |\n", len(charged))
	fmt.Fprintf(&report, "| 按了不會有事 | %d |\n", inert)
	fmt.Fprintf(&report, "| 合計 | %d |\n\n", total)

	missing := map[uint8]int{}
	for _, item := range charged {
		if !item.Declared {
			missing[item.Effect]++
		}
	}
	fmt.Fprintf(&report, "充能物品的效果編號就是法術主表的列。**沒有在 game pack 的 "+
		"`combat_player_spells` 裡宣告的效果沒有目標模式**，接不上既有的效果施法路——"+
		"那是 USE 還缺的東西：%d 件裡有 %d 個效果還沒宣告。\n\n", len(charged), len(missing))

	fmt.Fprintf(&report, "| 章 | 區塊 | 名稱 | 類別 | 充能 | 效果 | 已宣告 |\n")
	fmt.Fprintf(&report, "|---:|---:|---|---|---:|---|---|\n")
	for _, item := range charged {
		mark := "✅"
		if !item.Declared {
			mark = "—"
		}
		fmt.Fprintf(&report, "| %d | %d | %s | `%02Xh` | %d | `%02Xh` | %s |\n",
			item.Chapter, item.Block, item.Name, item.Type, item.Charges, item.Effect, mark)
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "total=%d scroll=%d charged=%d inert=%d undeclared-effects=%d\n",
		total, scrolls, len(charged), inert, len(missing))
}

func member(archive *zip.ReadCloser, name string) []byte {
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
