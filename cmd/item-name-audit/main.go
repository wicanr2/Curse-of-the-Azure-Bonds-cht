// Command item-name-audit 把原版每一件物品的名稱組出來，英文與繁中並排。
//
// ★ 存在的理由：物品名稱不是一句字串，是**三個名稱編號**去查
// `DS:1040h + N×15h` 那張 255 筆的詞表，再照 `overlay-24:0467h` 的順序串起來
// （spec 1178）。缺一個詞，玩家看到的就是少一截的名字，而且**不會有任何錯誤**
// ——所以「還缺多少」要有一支能引用的數字，不能憑印象。
//
// 用法：
//
//	go run ./cmd/item-name-audit -output docs/audit/item-names.md
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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// 詞表在 DOS 版的資料段：第 N 筆（N 從 1 起算）在 `DS:1040h + N×15h`，
// 每筆是一個 Pascal `string[20]`。
const (
	wordTableBase   = 0x1040
	wordTableStride = 0x15
	wordTableCount  = 0xFF
)

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "遊戲 image zip")
	dseg := flag.String("dseg", "workplace/re-sweep/dos/dseg/dos-dseg-dseg.bin", "DOS 資料段 dump")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "語系檔")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	english, err := loadWordTable(*dseg)
	if err != nil {
		log.Fatal(err)
	}
	localeBytes, err := os.ReadFile(*localePath)
	if err != nil {
		log.Fatal(err)
	}
	catalog, err := locale.Load(localeBytes)
	if err != nil {
		log.Fatal(err)
	}

	items, err := collectOriginalItems(*image)
	if err != nil {
		log.Fatal(err)
	}

	type row struct {
		english string
		chinese string
		count   int
	}
	rows := map[string]*row{}
	usedNumbers := map[uint8]int{}
	for _, item := range items {
		parts := make([]string, 0, 3)
		for slot := 3; slot >= 1; slot-- {
			number := item.NameNumbers[slot-1]
			if number == 0 {
				continue
			}
			usedNumbers[number]++
			parts = append(parts, english[number])
		}
		// 台帳看的是「全部鑑定後」的名字：`HiddenNameFlags` 只影響玩家當下看到
		// 什麼，不影響有沒有翻譯；數量前綴同理，不是名稱的一部分。
		visible := item
		visible.HiddenNameFlags = 0
		visible.Count = 0
		key := strings.Join(parts, " ")
		if existing, ok := rows[key]; ok {
			existing.count++
			continue
		}
		rows[key] = &row{
			english: key,
			chinese: monster.LocalizedItemName(visible, catalog),
			count:   1,
		}
	}

	missing := make([]uint8, 0, 8)
	for number := range usedNumbers {
		if catalog.Text(fmt.Sprintf("item_name_%02X", number), "") == "" {
			missing = append(missing, number)
		}
	}
	sort.Slice(missing, func(left, right int) bool { return missing[left] < missing[right] })

	keys := make([]string, 0, len(rows))
	for key := range rows {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var report strings.Builder
	fmt.Fprintf(&report, "# 原版物品名稱逐件對照\n\n")
	fmt.Fprintf(&report, "由 `cmd/item-name-audit` 產生，不要手改。組名規則見 spec 1178。\n\n")
	fmt.Fprintf(&report, "| 指標 | 數字 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 物品件數 | %d |\n", len(items))
	fmt.Fprintf(&report, "| 相異名稱 | %d |\n", len(keys))
	fmt.Fprintf(&report, "| 用到的名稱編號 | %d |\n", len(usedNumbers))
	fmt.Fprintf(&report, "| 其中缺譯 | %d |\n\n", len(missing))
	if len(missing) > 0 {
		fmt.Fprintf(&report, "缺譯的名稱編號：")
		for _, number := range missing {
			fmt.Fprintf(&report, " `%02Xh`", number)
		}
		report.WriteString("\n\n")
	}
	fmt.Fprintf(&report, "| 件數 | 原文 | 繁中 |\n|---:|---|---|\n")
	for _, key := range keys {
		fmt.Fprintf(&report, "| %d | %s | %s |\n", rows[key].count, rows[key].english, rows[key].chinese)
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "items=%d names=%d numbers=%d missing=%d\n",
		len(items), len(keys), len(usedNumbers), len(missing))
}

// loadWordTable 讀出那 255 個名稱成分。回傳的索引就是名稱編號本身（1 起算）。
func loadWordTable(path string) (map[uint8]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	words := make(map[uint8]string, wordTableCount)
	for number := 1; number <= wordTableCount; number++ {
		offset := wordTableBase + number*wordTableStride
		if offset+wordTableStride > len(data) {
			return nil, fmt.Errorf("資料段只有 %d 位元組，放不下第 %d 筆", len(data), number)
		}
		length := int(data[offset])
		if length >= wordTableStride {
			return nil, fmt.Errorf("第 %d 筆的長度位元組是 %d，超過 %d", number, length, wordTableStride-1)
		}
		words[uint8(number)] = string(data[offset+1 : offset+1+length])
	}
	return words, nil
}

// collectOriginalItems 讀出六章 `ITEM*.DAX` 裡的每一件物品。
func collectOriginalItems(image string) ([]monster.ItemRecord, error) {
	archive, err := zip.OpenReader(image)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	items := make([]monster.ItemRecord, 0, 256)
	for chapter := 1; chapter <= 6; chapter++ {
		payload := member(archive, fmt.Sprintf("ITEM%d.DAX", chapter))
		if payload == nil {
			continue
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			return nil, fmt.Errorf("ITEM%d.DAX: %w", chapter, parseErr)
		}
		for _, block := range blocks {
			// ⚠ 原版 DAX 一律走原版解碼（spec 1121）。
			parsed, itemErr := monster.ParseOriginalItems(block.Data)
			if itemErr != nil {
				continue
			}
			items = append(items, parsed...)
		}
	}
	return items, nil
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
