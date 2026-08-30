package main

import (
	"archive/zip"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

// 段的入口文字**一段都不准落回原文**。這是 SEG-32（語系不變量）在段入口這一層
// 的閘：落回原文代表那一段的第一個畫面玩家會看到英文。
//
// 同時釘住「入口不出文字」的那三段。它們不是缺漏——是從別段被帶進來的段落，
// 劇情在前一段的轉移與每回合／搜尋生命週期裡，不在 `initial`。多一段或少一段
// 都表示接線的形狀變了，要重新確認。
func TestSegmentEntryTextIsLocalized(t *testing.T) {
	const imagePath = "../../curseoftheazurebonds.zip"
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	blocks := map[uint8][]byte{}
	for chapter := 1; chapter <= 6; chapter++ {
		member := fmt.Sprintf("ECL%d.DAX", chapter)
		parsed, parseErr := dax.Parse(memberPayload(archive, member))
		if parseErr != nil {
			t.Fatalf("parse %s: %v", member, parseErr)
		}
		for _, block := range parsed {
			blocks[block.Entry.ID] = block.Data
		}
	}
	records := map[uint8]map[uint8]monster.Record{}
	for chapter := 1; chapter <= 6; chapter++ {
		member := fmt.Sprintf("MON%dCHA.DAX", chapter)
		payload := memberPayload(archive, member)
		if payload == nil {
			continue
		}
		parsed, parseErr := dax.Parse(payload)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", member, parseErr)
		}
		chapterRecords := map[uint8]monster.Record{}
		for _, block := range parsed {
			record, recordErr := monster.Parse(block.Data)
			if recordErr != nil {
				t.Fatalf("%s block %#02x: %v", member, block.Entry.ID, recordErr)
			}
			chapterRecords[block.Entry.ID] = record
		}
		records[uint8(chapter)] = chapterRecords
	}
	payload, err := os.ReadFile("../../assets/locale/zh-TW.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(payload)
	if err != nil {
		t.Fatal(err)
	}

	var silent []string
	for _, seg := range segment.All() {
		row := measure(catalog, blocks, records, seg)
		if row.note != "" {
			t.Errorf("%s 量不到：%s", seg.ID, row.note)
			continue
		}
		switch row.translated {
		case "原文":
			t.Errorf("%s 的入口文字落回原文：%s", seg.ID, row.message)
		case "—":
			silent = append(silent, seg.ID)
		}
		if row.choiceState == "原文" {
			t.Errorf("%s 的入口選項落回原文", seg.ID)
		}
	}
	sort.Strings(silent)
	const want = "ECL2/0x02 ECL3/0x12 ECL4/0x21"
	if got := strings.Join(silent, " "); got != want {
		t.Errorf("入口不出文字的段是 [%s]，宣告的是 [%s]", got, want)
	}
}

// languageOf 是這份盤點的判準，所以它自己要有正對照與反對照。
func TestLanguageOfSeparatesHanFromOriginalText(t *testing.T) {
	cases := map[string]string{
		"你們進入了火刀據點。":            "中文",
		"YOU HAVE ENTERED AN ANCIENT": "原文",
		"":                            "—",
		"???":                         "—",
		"火刀 HIDEOUT":                 "中文",
	}
	for text, want := range cases {
		if got := languageOf(text); got != want {
			t.Errorf("languageOf(%q) = %q，應為 %q", text, got, want)
		}
	}
}
