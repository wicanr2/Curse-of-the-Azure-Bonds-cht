// ecl-block-text 印出指定 ECL block 裡的英文敘述片段，用來判斷那個 block 是哪個區域。
//
// ⚠ 原作的敘述**全部是大寫**。用 `[A-Z][a-z]+` 這種「首字大寫」的樣式去撈地名
// 會一條都撈不到——那是假零，不是那個 block 沒有文字。
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	member := flag.String("member", "", "ECL 成員，例如 ECL5.DAX；留白就全部")
	block := flag.String("block", "", "block 編號（十六進位，例如 32）；留白就全部")
	limit := flag.Int("limit", 6, "每個 block 印幾條")
	minLength := flag.Int("min-length", 24, "片段最短長度")
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	for chapter := 1; chapter <= 6; chapter++ {
		name := fmt.Sprintf("ECL%d.DAX", chapter)
		if *member != "" && *member != name {
			continue
		}
		var payload []byte
		for _, file := range archive.File {
			if file.Name != name {
				continue
			}
			reader, err := file.Open()
			if err != nil {
				log.Fatal(err)
			}
			payload, err = io.ReadAll(reader)
			reader.Close()
			if err != nil {
				log.Fatal(err)
			}
		}
		if payload == nil {
			continue
		}
		blocks, err := dax.Parse(payload)
		if err != nil {
			log.Fatalf("%s: %v", name, err)
		}
		for _, raw := range blocks {
			if *block != "" {
				wanted, err := strconv.ParseUint(*block, 16, 8)
				if err != nil {
					log.Fatalf("block 不是十六進位：%v", err)
				}
				if uint64(raw.Entry.ID) != wanted {
					continue
				}
			}
			fmt.Printf("== %s/0x%02X ==\n", name, raw.Entry.ID)
			for _, line := range narrativeLines(raw.Data, *minLength, *limit) {
				fmt.Println("  " + line)
			}
		}
	}
}

// narrativeLines 撈出看起來像敘述的片段：夠長、而且大寫字母與空白佔多數。
// 打包過的位元組解出來會是亂碼，用這個比例擋掉。
func narrativeLines(data []byte, minLength, limit int) []string {
	seen := map[string]bool{}
	lines := make([]string, 0, limit)
	candidates := ecl.FindPackedTextCandidates(data)
	sort.SliceStable(candidates, func(i, j int) bool {
		return narrativeScore(candidates[i]) > narrativeScore(candidates[j])
	})
	for _, candidate := range candidates {
		text := strings.TrimSpace(candidate)
		if len(text) < minLength || narrativeScore(text) < 0.9 || seen[text] {
			continue
		}
		seen[text] = true
		lines = append(lines, text)
		if len(lines) >= limit {
			break
		}
	}
	return lines
}

func narrativeScore(text string) float64 {
	if len(text) == 0 {
		return 0
	}
	good := 0
	for _, r := range text {
		switch {
		case r >= 'A' && r <= 'Z', r == ' ', r == ',', r == '.', r == '\'':
			good++
		}
	}
	return float64(good) / float64(len(text))
}
