package game

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

// SEG-11 的往返閘：段的邊界狀態要存得下去、讀得回來，否則「下一段用上一段的
// 快照當入口」這件事就不成立。這裡量的是 25 段的入口狀態——段的結束狀態要等
// 每段有自己的測試（SEG-10）之後才存得到。
func TestSegmentEntrySnapshotRoundTrips(t *testing.T) {
	blocks, records := segmentEntryBlocks(t)
	directory := t.TempDir()
	for _, seg := range segment.All() {
		t.Run(seg.ID, func(t *testing.T) {
			origin := NewStateFromECLBlocks(testCatalog(), blocks, 0x50)
			origin.SetMonsterRecords(records)
			if err := origin.OpenCharacterCreation(); err != nil {
				t.Fatal(err)
			}
			if err := origin.AddCreationCharacter(0); err != nil {
				t.Fatal(err)
			}
			if err := origin.FinishCharacterCreation(); err != nil {
				t.Fatal(err)
			}
			if err := origin.EnterSegment(seg); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(directory, segmentSnapshotName(seg)+".json")
			if err := origin.SavePartyFile(path); err != nil {
				t.Fatalf("存不下去：%v", err)
			}
			restored := NewStateFromECLBlocks(testCatalog(), blocks, 0x50)
			restored.SetMonsterRecords(records)
			if err := restored.LoadPartyFile(path); err != nil {
				t.Fatalf("讀不回來：%v", err)
			}
			assertSegmentStateMatches(t, &origin, &restored)
		})
	}
}

func assertSegmentStateMatches(t *testing.T, origin, restored *State) {
	t.Helper()
	if got, want := restored.session.CurrentBlockID(), origin.session.CurrentBlockID(); got != want {
		t.Errorf("block %#02x，原本是 %#02x", got, want)
	}
	if restored.Mode != origin.Mode {
		t.Errorf("模式 %v，原本是 %v", restored.Mode, origin.Mode)
	}
	if restored.Area.GameArea != origin.Area.GameArea || restored.Area.InDungeon != origin.Area.InDungeon {
		t.Errorf("章節／地城 %d/%v，原本是 %d/%v",
			restored.Area.GameArea, restored.Area.InDungeon,
			origin.Area.GameArea, origin.Area.InDungeon)
	}
	if restored.GeoMapSet != origin.GeoMapSet || restored.GeoMapBlock != origin.GeoMapBlock {
		t.Errorf("GEO %d/%#02x，原本是 %d/%#02x",
			restored.GeoMapSet, restored.GeoMapBlock, origin.GeoMapSet, origin.GeoMapBlock)
	}
	if restored.DungeonX != origin.DungeonX || restored.DungeonY != origin.DungeonY ||
		restored.DungeonDirection != origin.DungeonDirection {
		t.Errorf("地城座標 (%d,%d,%d)，原本是 (%d,%d,%d)",
			restored.DungeonX, restored.DungeonY, restored.DungeonDirection,
			origin.DungeonX, origin.DungeonY, origin.DungeonDirection)
	}
	if len(restored.PartyFighters()) != len(origin.PartyFighters()) {
		t.Errorf("隊伍 %d 人，原本是 %d 人",
			len(restored.PartyFighters()), len(origin.PartyFighters()))
	}
}

// segmentSnapshotName 把段的 id 變成檔名。
func segmentSnapshotName(seg segment.Segment) string {
	return strings.ReplaceAll(seg.ID, "/", "-")
}
