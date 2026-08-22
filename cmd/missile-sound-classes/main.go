// Command missile-sound-classes 把原作 `SHOWARROW` 的武器類別分歧翻成人看得懂的表。
//
// ★ 存在的理由：spec 1186 解出投射武器的第二聲依 `CHARITEMREC.ITEMPTR`（＝物品
// 類別）分歧，但那是 14 個裸數字。要接進 remake、也要讓人能複核，得知道每個
// 數字是哪一種武器，以及**遊戲裡到底有沒有那一類的物品**。
//
// 資料來源兩份，互相獨立：
//
//	ITEMS        類別表（每類 16 bytes，`ITEMREC`）：射程與 `MISSLETYPE` 旗標
//	ITEM1..6.DAX 各章的物品實例：名字，以及「這一類真的出現過嗎」
//
// ⚠ **類別表有一列不代表遊戲裡有那件東西**。分歧鏈上的類別在 corpus 裡一件都
// 沒有，就表示那條分支玩家走不到——那是「不必接」，不是「還沒接」。兩者
// 混在一起會讓覆蓋率看起來比實際差。
//
// 用法：
//
//	go run ./cmd/missile-sound-classes -output docs/audit/missile-sound-classes.md
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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// showArrowBranch 是 `SHOWARROW` 的分歧鏈，照原作的比較順序（spec 1186）。
// 六個 `cmp/jz` 全部落在同一個目標，所以第一組是一條分支不是六條。
var showArrowBranch = []struct {
	sound string
	at    string
	types []uint8
}{
	{"ARROWFX（箭）", "2B4Ah", []uint8{0x09, 0x15, 0x64, 0x1C, 0x1F, 0x49}},
	{"SWISHFX（揮擊）", "2B81h", []uint8{0x02, 0x07, 0x14}},
	{"WHISTLEFX（哨音）", "2BB4h", []uint8{0x55, 0x56}},
	{"WHISTLEFX（哨音）", "2C01h", []uint8{0x65, 0x2F, 0x62}},
}

// fallbackSound 是分歧鏈全部落空時走的那一條（`2C48h`）。
const fallbackSound = "SWISHFX（揮擊）"

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	catalogBytes := member(archive, "ITEMS")
	if catalogBytes == nil {
		log.Fatal("遊戲 image 裡沒有 ITEMS")
	}
	catalog, err := monster.ParseBaseItems(catalogBytes)
	if err != nil {
		log.Fatal(err)
	}

	// 物品實例：類別 → 出現過的名字與件數。
	names := map[uint8]map[string]int{}
	instances := map[uint8]int{}
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
			// ⚠ 原版 DAX 一律走原版解碼（Big5），否則物品名讀成亂碼（spec 1121）。
			items, itemErr := monster.ParseOriginalItems(block.Data)
			if itemErr != nil {
				continue
			}
			for _, item := range items {
				if names[item.Type] == nil {
					names[item.Type] = map[string]int{}
				}
				names[item.Type][item.Name]++
				instances[item.Type]++
			}
		}
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# 投射武器的第二聲：`SHOWARROW` 依物品類別分歧\n\n")
	fmt.Fprintf(&report, "由 `cmd/missile-sound-classes` 產生，不要手改。分歧鏈的位元組證據見 spec 1186。\n\n")
	fmt.Fprintf(&report, "`SHOWARROW` 進場**無條件**放 `ARROWFX`，之後依 `CHARITEMREC.ITEMPTR`"+
		"（＝物品類別，remake 的 `ItemRecord.Type`）在飛行動畫尾端再放一聲。\n\n")
	fmt.Fprintf(&report, "⚠ **類別表有一列不代表遊戲裡有那件東西**。`實例` 是六章 `ITEM*.DAX` 裡"+
		"真的存在的件數；0 就表示那條分支玩家走不到，是「不必接」而不是「還沒接」。\n\n")
	fmt.Fprintf(&report, "⚠ `射程` 與 `MISSLETYPE` 取自類別表（`ITEMREC` 的 `+0Ch`／`+0Eh`）。"+
		"原作的 `USINGMISSLEWEAPON` 用的是 `射程 > 1`，`USINGHURLEDWEAPON` 用的是"+
		"`MISSLETYPE and 14h ＝ 14h`——**兩個判斷式不同**，不要拿其中一個去解釋另一個。\n\n")

	fmt.Fprintf(&report, "| 第二聲 | 位移 | 類別 | 名稱 | 實例 | 射程 | `MISSLETYPE` |\n")
	fmt.Fprintf(&report, "|---|---:|---:|---|---:|---:|---|\n")
	listed := map[uint8]bool{}
	reachable, unreachable := 0, 0
	for _, branch := range showArrowBranch {
		types := append([]uint8(nil), branch.types...)
		sort.Slice(types, func(left, right int) bool { return types[left] < types[right] })
		for _, itemType := range types {
			listed[itemType] = true
			if instances[itemType] > 0 {
				reachable++
			} else {
				unreachable++
			}
			rangeText, flagText := "—", "—"
			if base, ok := catalog.Lookup(itemType); ok {
				rangeText = fmt.Sprintf("%d", base.Range)
				flagText = fmt.Sprintf("`%02Xh`%s", base.AmmunitionType, flagNotes(base))
			}
			fmt.Fprintf(&report, "| %s | `%s` | `%02Xh`（%d）| %s | %d | %s | %s |\n",
				branch.sound, branch.at, itemType, itemType,
				nameList(names[itemType]), instances[itemType], rangeText, flagText)
		}
	}
	fmt.Fprintf(&report, "| %s | `2C48h` | 其餘全部 | — | — | — | — |\n\n", fallbackSound)

	fmt.Fprintf(&report, "| 指標 | 數字 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 分歧鏈點名的類別 | %d |\n", len(listed))
	fmt.Fprintf(&report, "| 其中遊戲裡真的有物品的 | %d |\n", reachable)
	fmt.Fprintf(&report, "| 其中一件都沒有的（走不到）| %d |\n\n", unreachable)

	// 反向：有 `MISSLETYPE` 發射位元、卻沒被分歧鏈點名的類別 → 落到 `2C48h`。
	var fallthroughTypes []uint8
	for _, base := range catalog.Items {
		if listed[base.Type] || !base.IsMissileWeapon() || instances[base.Type] == 0 {
			continue
		}
		fallthroughTypes = append(fallthroughTypes, base.Type)
	}
	sort.Slice(fallthroughTypes, func(left, right int) bool {
		return fallthroughTypes[left] < fallthroughTypes[right]
	})
	fmt.Fprintf(&report, "## 落到預設分支的發射武器\n\n")
	fmt.Fprintf(&report, "有發射位元（`MISSLETYPE` bit 3）、遊戲裡也真的有，但**分歧鏈沒點名**"+
		"⇒ 走 `2C48h` 的 `SWISHFX`。這一份是「預設分支不是空的」的證據："+
		"少了它會以為沒被點名的類別不會走到 `SHOWARROW`。\n\n")
	if len(fallthroughTypes) == 0 {
		fmt.Fprintf(&report, "（沒有。）\n")
	} else {
		fmt.Fprintf(&report, "| 類別 | 名稱 | 實例 | 射程 |\n|---:|---|---:|---:|\n")
		for _, itemType := range fallthroughTypes {
			base, _ := catalog.Lookup(itemType)
			fmt.Fprintf(&report, "| `%02Xh`（%d）| %s | %d | %d |\n",
				itemType, itemType, nameList(names[itemType]), instances[itemType], base.Range)
		}
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "listed=%d reachable=%d unreachable=%d fallthrough=%d\n",
		len(listed), reachable, unreachable, len(fallthroughTypes))
}

func flagNotes(base monster.BaseItem) string {
	notes := make([]string, 0, 2)
	if base.IsMissileWeapon() {
		notes = append(notes, "發射")
	}
	if base.IsThrownWeapon() {
		notes = append(notes, "投擲")
	}
	if len(notes) == 0 {
		return ""
	}
	return "（" + strings.Join(notes, "／") + "）"
}

// nameList 把一個類別觀察到的名字收成一格。名字多於一個時全部列出來——
// 只印第一個會讓「這一類到底是什麼」變成看運氣。
func nameList(observed map[string]int) string {
	if len(observed) == 0 {
		return "**（遊戲裡沒有這一類的物品）**"
	}
	list := make([]string, 0, len(observed))
	for name := range observed {
		list = append(list, name)
	}
	sort.Strings(list)
	return strings.Join(list, "、")
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
