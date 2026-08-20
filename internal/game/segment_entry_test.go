package game

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

func segmentEntryBlocks(t *testing.T) (map[uint8][]byte, map[uint8]monster.Record) {
	t.Helper()
	image, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	all := make(map[uint8][]byte)
	for member := 1; member <= 6; member++ {
		blocks, parseErr := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(member)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			all[block.Entry.ID] = block.Data
		}
	}
	// 開場那一段的 initial lifecycle 會 ADD NPC，沒有 MON1CHA 就進不去。
	records := make(map[uint8]monster.Record)
	monsterBlocks, parseErr := dax.Parse(zipData(t, image, "MON1CHA.DAX"))
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	for _, block := range monsterBlocks {
		record, recordErr := monster.Parse(block.Data)
		if recordErr != nil {
			t.Fatalf("MON1CHA block %#02x: %v", block.Entry.ID, recordErr)
		}
		records[block.Entry.ID] = record
	}
	return all, records
}

// SEG-04 的驗收條件：註冊表裡每一段都進得去。這一條把 25 段全部走一次，
// 段紅的時候看得出是哪一段——而不是整條主線一起紅。
func TestEverySegmentInTheRegistryCanBeEntered(t *testing.T) {
	blocks, records := segmentEntryBlocks(t)
	for _, seg := range segment.All() {
		t.Run(seg.ID, func(t *testing.T) {
			state := NewStateFromECLBlocks(testCatalog(), blocks, 0x50)
			state.SetMonsterRecords(records)
			if err := state.OpenCharacterCreation(); err != nil {
				t.Fatal(err)
			}
			if err := state.AddCreationCharacter(0); err != nil {
				t.Fatal(err)
			}
			if err := state.FinishCharacterCreation(); err != nil {
				t.Fatal(err)
			}
			if err := state.EnterSegment(seg); err != nil {
				t.Fatalf("進不去：%v", err)
			}
			if got := state.session.CurrentBlockID(); got != seg.Settles() {
				t.Fatalf("進去之後停在 block %#02x，宣告的是 %#02x", got, seg.Settles())
			}
			if len(state.PartyFighters()) == 0 {
				t.Fatal("段入口沒有隊伍")
			}
			if seg.Overland {
				if state.Area.InDungeon {
					t.Error("世界地圖段卻標成地城")
				}
			} else if state.Area.GameArea != seg.Member {
				t.Errorf("GEO 檔集 %d 不等於成員編號 %d", state.Area.GameArea, seg.Member)
			}
		})
	}
}

// 段的入口如果一進去就要出圖，那張圖的素材必須已經匯出——否則玩家在那一段
// 的第一個畫面就是「素材尚未載入」。已知缺口列在下表，新的缺口會讓這條紅。
var knownSegmentArtGaps = map[string]string{
	"ECL4/0x22": "入口要 character-area-4-head-25-body-23；頭與身體是不同區塊的組合，" +
		"匯出腳本目前只出 head 與 body 同號的那些",
}

func TestSegmentEntryArtIsExported(t *testing.T) {
	blocks, records := segmentEntryBlocks(t)
	for _, seg := range segment.All() {
		t.Run(seg.ID, func(t *testing.T) {
			state := NewStateFromECLBlocks(testCatalog(), blocks, 0x50)
			state.SetMonsterRecords(records)
			if err := state.OpenCharacterCreation(); err != nil {
				t.Fatal(err)
			}
			if err := state.AddCreationCharacter(0); err != nil {
				t.Fatal(err)
			}
			if err := state.FinishCharacterCreation(); err != nil {
				t.Fatal(err)
			}
			if err := state.EnterSegment(seg); err != nil {
				t.Fatal(err)
			}
			pattern := segmentEntryArtPattern(&state)
			t.Logf("入口素材：%q", pattern)
			reason, known := knownSegmentArtGaps[seg.ID]
			if pattern == "" {
				if known {
					t.Errorf("宣告了素材缺口卻沒要出圖：%s", reason)
				}
				return
			}
			matches, err := filepath.Glob(filepath.Join("../../assets/sprites", pattern))
			if err != nil {
				t.Fatal(err)
			}
			switch {
			case len(matches) > 0 && known:
				t.Errorf("%s 已經匯出了，請把它從已知缺口清單移掉", pattern)
			case len(matches) == 0 && !known:
				t.Errorf("段入口要 %s，但 assets/sprites 沒有這個素材", pattern)
			}
		})
	}
}

// segmentEntryArtPattern 回傳這一段入口畫面需要的素材檔名（glob），
// 對應 cmd/azure-bonds-game 的 drawPictureAnimation 三條分支；不出圖就回空字串。
func segmentEntryArtPattern(state *State) string {
	switch {
	case state.SceneCharacterRequested:
		return fmt.Sprintf("character-area-%d-head-%02X-body-%02X.png",
			state.Area.GameArea, state.SceneHeadBlock, state.SceneBodyBlock)
	case state.BigPictureRequested:
		// 大圖用區塊編號定檔，不跟章節走（見 cmd/azure-bonds-game 的
		// bigPictureSprite）。
		return fmt.Sprintf("bigpic*-block-%02X-item-00.png", state.PictureBlock)
	case state.PictureRequested:
		// 事件圖畫的是動畫的第 0 幀，素材是 animation.json 的那一組。
		return fmt.Sprintf("pic%d-block-%02X-frame-*.png",
			state.Area.GameArea, state.PictureBlock)
	}
	return ""
}
