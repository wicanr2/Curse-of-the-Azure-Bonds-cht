// segment-coverage 盤點主線 25 段各自「接到什麼程度」：直接進入每一段，記下
// 入口畫面出了什麼（圖／文字／選單）、文字有沒有中文化，以及那一段的五個
// 生命週期入口在不在。
//
// ★ 存在的理由：`docs/plan/mainline-segmented-verification.md` 階段 2 要按
// 「已接／有幾何沒劇情／完全沒接」排序，而那份分類是憑印象寫的。這支把它變成
// 每次都跑得出來的數字。
//
// ⚠ 「有沒有中文化」是**看有沒有漢字**判定的。原作文字整段大寫英文，所以
// 沒有漢字而有英文字母就是落回原文；兩者都沒有代表那一段入口不出文字。
//
// 用法：
//
//	go run ./cmd/segment-coverage
//	go run ./cmd/segment-coverage -out docs/audit/segment-coverage.md
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

type coverage struct {
	id          string
	label       string
	art         string
	message     string
	translated  string
	choices     int
	choiceState string
	nextStep    string
	note        string
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "locale JSON path")
	out := flag.String("out", "docs/audit/segment-coverage.md", "輸出的 markdown")
	flag.Parse()

	archive, err := zip.OpenReader(*image)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	blocks := map[uint8][]byte{}
	for chapter := 1; chapter <= 6; chapter++ {
		member := fmt.Sprintf("ECL%d.DAX", chapter)
		parsed, err := dax.Parse(memberPayload(archive, member))
		if err != nil {
			log.Fatalf("parse %s: %v", member, err)
		}
		for _, block := range parsed {
			blocks[block.Entry.ID] = block.Data
		}
	}
	// 六章的怪物表都要載：段的入口有可能直接開戰，只載第一章會在別章落空。
	records := map[uint8]map[uint8]monster.Record{}
	for chapter := 1; chapter <= 6; chapter++ {
		member := fmt.Sprintf("MON%dCHA.DAX", chapter)
		payload := memberPayload(archive, member)
		if payload == nil {
			continue
		}
		monsterBlocks, err := dax.Parse(payload)
		if err != nil {
			log.Fatalf("parse %s: %v", member, err)
		}
		chapterRecords := map[uint8]monster.Record{}
		for _, block := range monsterBlocks {
			record, err := monster.Parse(block.Data)
			if err != nil {
				log.Fatalf("%s block %#02x: %v", member, block.Entry.ID, err)
			}
			chapterRecords[block.Entry.ID] = record
		}
		records[uint8(chapter)] = chapterRecords
	}
	localeData, err := os.ReadFile(*localePath)
	if err != nil {
		log.Fatal(err)
	}
	catalog, err := locale.Load(localeData)
	if err != nil {
		log.Fatal(err)
	}

	rows := make([]coverage, 0, 25)
	for _, seg := range segment.All() {
		rows = append(rows, measure(catalog, blocks, records, seg))
	}
	if err := os.WriteFile(*out, []byte(render(rows)), 0o644); err != nil {
		log.Fatal(err)
	}
	wired := 0
	for _, row := range rows {
		if row.translated == "中文" {
			wired++
		}
	}
	fmt.Printf("段=%d 入口文字已中文化=%d → %s\n", len(rows), wired, *out)
}

func measure(catalog locale.Catalog, blocks map[uint8][]byte,
	records map[uint8]map[uint8]monster.Record, seg segment.Segment) coverage {
	row := coverage{id: seg.ID}
	state := game.NewStateFromECLBlocks(catalog, blocks, 0x50)
	for chapter, chapterRecords := range records {
		state.SetMonsterRecordsForECL(chapter, chapterRecords)
	}
	if first, ok := records[1]; ok {
		state.SetMonsterRecords(first)
	}
	if err := state.OpenCharacterCreation(); err != nil {
		row.note = err.Error()
		return row
	}
	if err := state.AddCreationCharacter(0); err != nil {
		row.note = err.Error()
		return row
	}
	if err := state.FinishCharacterCreation(); err != nil {
		row.note = err.Error()
		return row
	}
	if err := state.EnterSegment(seg); err != nil {
		row.note = "進不去：" + err.Error()
		return row
	}
	row.art = artKind(&state)
	row.message = firstLine(state.Message)
	row.translated = languageOf(state.Message)
	row.choices = len(state.Choices)
	row.choiceState = languageOf(strings.Join(state.Choices, " "))
	row.nextStep = advanceOnce(&state)
	return row
}

func artKind(state *game.State) string {
	switch {
	case state.SceneCharacterRequested:
		return "人物圖層"
	case state.BigPictureRequested:
		return "大圖"
	case state.PictureRequested:
		return "事件圖"
	}
	return "—"
}

// languageOf 判定一段玩家看得到的字是中文、原文還是沒有字。原作文字是英文，
// 所以「沒有漢字但有英文字母」就是落回原文。
func languageOf(text string) string {
	hasHan, hasLatin := false, false
	for _, glyph := range text {
		switch {
		case unicode.Is(unicode.Han, glyph):
			hasHan = true
		case glyph >= 'A' && glyph <= 'Z', glyph >= 'a' && glyph <= 'z':
			hasLatin = true
		}
	}
	switch {
	case hasHan:
		return "中文"
	case hasLatin:
		return "原文"
	}
	return "—"
}

// advanceOnce 從段的入口再走一步：入口停在事件上就按繼續，停在選單上就選第一項。
// 這一欄回答的是「這一段的第一個畫面之後還走得動嗎」——入口有字不代表段是通的。
func advanceOnce(state *game.State) string {
	before := state.Mode
	switch {
	case state.Mode == game.ModeEvent:
		// 事件模式的選單要按繼續，`Select` 只在世界地圖與地點模式收得下去。
		if err := state.Continue(); err != nil {
			return "繼續失敗：" + err.Error()
		}
	case len(state.Choices) > 0:
		if err := state.Select(0); err != nil {
			return "選第一項失敗：" + err.Error()
		}
	default:
		return "入口沒有可走的下一步"
	}
	return fmt.Sprintf("%s → %s／%s", modeName(before), modeName(state.Mode),
		languageOf(state.Message))
}

func modeName(mode game.Mode) string {
	switch mode {
	case game.ModeTitle:
		return "標題"
	case game.ModeWilderness:
		return "世界地圖"
	case game.ModeEvent:
		return "事件"
	case game.ModeMap:
		return "地圖"
	case game.ModePlace:
		return "地點"
	case game.ModeCombat:
		return "戰鬥"
	case game.ModeJournal:
		return "手札"
	case game.ModeCharacterCreation:
		return "建角"
	case game.ModeDungeon:
		return "地城"
	}
	return "?"
}

func firstLine(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "—"
	}
	runes := []rune(strings.ReplaceAll(text, "\n", " "))
	if len(runes) > 24 {
		return string(runes[:24]) + "…"
	}
	return string(runes)
}

func render(rows []coverage) string {
	var out strings.Builder
	out.WriteString("# 主線分段的接線盤點\n\n" +
		"由 `cmd/segment-coverage` 產生，不要手改。\n\n" +
		"每一段用 `-segment <id>` 直接進入之後量的：入口畫面出了什麼、" +
		"入口文字與選項是中文還是落回原文、從入口再走一步會到哪裡。\n\n" +
		"⚠ 「中文／原文」是**看有沒有漢字**判定的。原作文字整段是英文，" +
		"所以沒有漢字而有英文字母就是落回原文；兩者都沒有代表那裡不出文字。\n" +
		"⚠ 量的是**段的入口**。入口中文不代表整段的文字都接好了——" +
		"段內的逐頁覆蓋看 `cmd/ecl-text-coverage`。\n" +
		"⚠ 「入口不出文字」不等於那一段沒接：有幾段是從別段被帶進來的" +
		"（神殿牢房、城區共用地圖、地下第二層），劇情在前一段的轉移與" +
		"每回合／搜尋生命週期裡，不在 `initial`。\n\n")
	out.WriteString("| 段 | 入口畫面 | 入口文字 | 語言 | 選項 | 選項語言 | 再走一步 |\n")
	out.WriteString("|---|---|---|---|---:|---|---|\n")
	for _, row := range rows {
		if row.note != "" {
			out.WriteString(fmt.Sprintf("| `%s` | %s | | | | | |\n", row.id, row.note))
			continue
		}
		out.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %d | %s | %s |\n",
			row.id, row.art, row.message, row.translated, row.choices,
			row.choiceState, row.nextStep))
	}
	counts := map[string]int{}
	for _, row := range rows {
		counts[row.translated]++
	}
	out.WriteString(fmt.Sprintf("\n## 摘要\n\n| 入口文字 | 段數 |\n|---|---:|\n"+
		"| 中文 | %d |\n| 落回原文 | %d |\n| 入口不出文字 | %d |\n",
		counts["中文"], counts["原文"], counts["—"]))
	return out.String()
}

func memberPayload(archive *zip.ReadCloser, member string) []byte {
	for _, file := range archive.File {
		if file.Name != member {
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
