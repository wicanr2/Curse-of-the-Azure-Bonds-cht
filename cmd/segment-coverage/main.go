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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/wicanr2/golden-box-remake-engine/dax"
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
	out := flag.String("out", "docs/audit/segment-coverage.md", tooltext.Text("h.aff4479ab1b9"))
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
		if row.translated == tooltext.Text("h.72726d8818f6") {
			wired++
		}
	}
	fmt.Print(tooltext.Format("h.da6f2ccbbc48", len(rows), wired, *out))
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
		row.note = tooltext.Text("h.1e5d9d1f7560") + err.Error()
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
		return tooltext.Text("h.b6c8b0528e55")
	case state.BigPictureRequested:
		return tooltext.Text("h.ba6130decfe8")
	case state.PictureRequested:
		return tooltext.Text("h.e099d7c4c13a")
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
		return tooltext.Text("h.72726d8818f6")
	case hasLatin:
		return tooltext.Text("h.354b28c85333")
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
			return tooltext.Text("h.2e2a749825c6") + err.Error()
		}
	case len(state.Choices) > 0:
		if err := state.Select(0); err != nil {
			return tooltext.Text("h.a1460ec5be6f") + err.Error()
		}
	default:
		return tooltext.Text("h.2b6b63b752d8")
	}
	return fmt.Sprintf("%s → %s／%s", modeName(before), modeName(state.Mode),
		languageOf(state.Message))
}

func modeName(mode game.Mode) string {
	switch mode {
	case game.ModeTitle:
		return tooltext.Text("h.6fe38ed1ee10")
	case game.ModeWilderness:
		return tooltext.Text("h.58f78bc6a875")
	case game.ModeEvent:
		return tooltext.Text("h.c560201b331c")
	case game.ModeMap:
		return tooltext.Text("h.90e1b1b8a537")
	case game.ModePlace:
		return tooltext.Text("h.af2cfdb525ff")
	case game.ModeCombat:
		return tooltext.Text("h.625dd417c2c3")
	case game.ModeJournal:
		return tooltext.Text("h.53ce36b5da9e")
	case game.ModeCharacterCreation:
		return tooltext.Text("h.894d0b190ab5")
	case game.ModeDungeon:
		return tooltext.Text("h.c014266366bb")
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
	out.WriteString(tooltext.Text("h.f945c41aa7de") +
		tooltext.Text("h.abe421dba89d") +
		tooltext.Text("h.752b39680d1b") +
		tooltext.Text("h.f829f7c8448d") +
		tooltext.Text("h.0a8538915c86") +
		tooltext.Text("h.05d34a36f74a") +
		tooltext.Text("h.fa19d9d59bf3") +
		tooltext.Text("h.f8c7e7b29462") +
		tooltext.Text("h.1d23b3e74926") +
		tooltext.Text("h.bc7c19a1b4c7") +
		tooltext.Text("h.41091c0396c3"))
	out.WriteString(tooltext.Text("h.fd2a35a9d136"))
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
	out.WriteString(fmt.Sprintf(tooltext.Text("h.3048482cced7")+
		tooltext.Text("h.2534211b920a"),
		counts[tooltext.Text("h.72726d8818f6")], counts[tooltext.Text("h.354b28c85333")], counts["—"]))
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
