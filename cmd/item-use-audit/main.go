// Command item-use-audit 把原版每一件物品照「戰鬥裡按 USE 會走哪一條路」分類。
//
// ★ 存在的理由：「還缺多少」不能憑印象——這一支把它變成可以引用的數字：
// 全部物品幾件、其中卷軸幾件、充能物品幾件，以及那些充能效果 remake 有沒有
// **接上**（`internal/combat` 的 handler 行為表，spec 1169／1170）。
//
// 分類依據（spec 921／1171）：
//
//	卷軸    **類別表的槽**在 `0Bh`..`0Dh`。三個效果槽是三個法術，選一個來唸。
//	充能    `+3Dh > 0` 且 `+3Eh < 80h` 且不是卷軸。效果 ＝ `+3Dh and 7Fh`。
//	其他    按下去什麼都不會發生。
//
// ⚠ 卷軸**不能用物品類別判**：`3Ch` 的名字叫 `Scroll` 但槽是 `0Ah`，原作把它
// 當充能物品（效果 `5Fh`／`60h`）。
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

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

type chargedItem struct {
	Chapter int
	Block   uint8
	Name    string
	Type    uint8
	Charges uint8
	Effect  uint8
	Wired   bool
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

	// 卷軸的判別要查類別表（`ITEMS`），不是物品類別本身（spec 1171）。
	catalogBytes := member(archive, "ITEMS")
	if catalogBytes == nil {
		log.Fatal("遊戲 image 裡沒有 ITEMS")
	}
	catalog, err := monster.ParseBaseItems(catalogBytes)
	if err != nil {
		log.Fatal(err)
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
				case catalog.IsScroll(item.Type):
					scrolls++
				case item.Affects[1] > 0 && item.Affects[2] < 0x80:
					effect := item.Affects[1] & 0x7F
					charged = append(charged, chargedItem{
						Chapter: chapter, Block: block.Entry.ID, Name: item.Name,
						Type: item.Type, Charges: item.Affects[0], Effect: effect,
						Wired: wired(effect),
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
	fmt.Fprintf(&report, "| 卷軸（類別表的槽 `0Bh`..`0Dh`）| %d |\n", scrolls)
	fmt.Fprintf(&report, "| 充能物品 | %d |\n", len(charged))
	fmt.Fprintf(&report, "| 按了不會有事 | %d |\n", inert)
	fmt.Fprintf(&report, "| 合計 | %d |\n\n", total)

	missing := map[uint8]int{}
	for _, item := range charged {
		if !item.Wired {
			missing[item.Effect]++
		}
	}
	fmt.Fprintf(&report, "充能物品的效果編號就是**法術主表的列**：目標模式（`+6`）、"+
		"豁免（`+8`）、效果碼（`+0Ah`）都在表裡，骰子在各自的 handler 裡"+
		"（spec 1169）。`已接` ＝ remake 的 `internal/combat` 有那一支 handler 的"+
		"行為讀法；%d 件裡還有 %d 個效果沒接。\n\n", len(charged), len(missing))

	fmt.Fprintf(&report, "| 章 | 區塊 | 名稱 | 類別 | 充能 | 效果 | 已接 |\n")
	fmt.Fprintf(&report, "|---:|---:|---|---|---:|---|---|\n")
	for _, item := range charged {
		mark := "✅"
		if !item.Wired {
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
	fmt.Fprintf(os.Stderr, "total=%d scroll=%d charged=%d inert=%d unwired-effects=%d\n",
		total, scrolls, len(charged), inert, len(missing))
}

// wired 回答「remake 接得起這個效果嗎」。判斷落在 `internal/combat` 那張表上
// 一處，報表不自己維護第二份清單——兩份會漂移。
func wired(effect uint8) bool {
	_, ok := combat.LookupChargedItemBehaviour(effect)
	return ok
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
