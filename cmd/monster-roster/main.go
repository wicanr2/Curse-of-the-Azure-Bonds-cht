// monster-roster 匯出 `MON*CHA` 的怪物名單與 `MON*SPC` 的天生效果碼。
//
// ★ 用途是回答「這個遊戲裡有沒有某一類怪物」這種問題，而且**用資料回答**，
// 不是靠記憶或系列作的印象。第一次用它是為了確認 CoAB 有沒有會吸等級的
// 不死生物（沒有，spec 1127）。
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	flag.Parse()

	reader, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	names := map[int]map[byte]string{}
	appearances := map[string][]string{}
	for chapter := 1; chapter <= 6; chapter++ {
		names[chapter] = map[byte]string{}
		for _, block := range blocks(reader, fmt.Sprintf("MON%dCHA.DAX", chapter)) {
			record, err := monster.Parse(block.Data)
			if err != nil {
				continue
			}
			names[chapter][block.Entry.ID] = record.Name
			appearances[record.Name] = append(appearances[record.Name],
				fmt.Sprintf("%d:%d", chapter, block.Entry.ID))
		}
	}

	kinds := map[uint8]map[string]bool{}
	for chapter := 1; chapter <= 6; chapter++ {
		for _, block := range blocks(reader, fmt.Sprintf("MON%dSPC.DAX", chapter)) {
			affects, err := monster.ParseAffects(block.Data)
			if err != nil {
				continue
			}
			name := names[chapter][block.Entry.ID]
			if name == "" {
				name = fmt.Sprintf("chapter%d#%d", chapter, block.Entry.ID)
			}
			for _, affect := range affects {
				if kinds[affect.Kind] == nil {
					kinds[affect.Kind] = map[string]bool{}
				}
				kinds[affect.Kind][name] = true
			}
		}
	}

	roster := make([]string, 0, len(appearances))
	for name := range appearances {
		roster = append(roster, name)
	}
	sort.Strings(roster)
	fmt.Printf("# roster (%d distinct)\n", len(roster))
	for _, name := range roster {
		fmt.Printf("%-28s %s\n", name, strings.Join(appearances[name], " "))
	}

	codes := make([]int, 0, len(kinds))
	for code := range kinds {
		codes = append(codes, int(code))
	}
	sort.Ints(codes)
	fmt.Printf("\n# innate effect codes (%d distinct)\n", len(codes))
	for _, code := range codes {
		owners := make([]string, 0, len(kinds[uint8(code)]))
		for name := range kinds[uint8(code)] {
			owners = append(owners, name)
		}
		sort.Strings(owners)
		fmt.Printf("%02Xh  %s\n", code, strings.Join(owners, ", "))
	}
}

func blocks(reader *zip.ReadCloser, want string) []dax.Block {
	for _, file := range reader.File {
		if !strings.EqualFold(file.Name, want) {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		data, err := io.ReadAll(handle)
		handle.Close()
		if err != nil {
			log.Fatal(err)
		}
		parsed, err := dax.Parse(data)
		if err != nil {
			log.Fatal(err)
		}
		return parsed
	}
	return nil
}
