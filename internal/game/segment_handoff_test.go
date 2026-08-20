package game

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

var handoffRowPattern = regexp.MustCompile(
	"^\\| `(ECL[1-6]/0x[0-9A-F]{2})` \\| [^|]* \\| [^|]* \\| ([^|]*) \\|")

// segmentEdges 讀轉移圖的段落清單，回傳每一條 `NEWECL` 邊。
func segmentEdges(t *testing.T) [][2]segment.Segment {
	t.Helper()
	payload, err := os.ReadFile("../../docs/audit/ecl-block-graph.md")
	if err != nil {
		t.Fatal(err)
	}
	var edges [][2]segment.Segment
	for _, line := range strings.Split(string(payload), "\n") {
		match := handoffRowPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		source, ok := segment.Lookup(match[1])
		if !ok {
			t.Fatalf("轉移圖有 %s，註冊表沒有", match[1])
		}
		for _, token := range strings.Split(match[2], "、") {
			token = strings.Trim(strings.TrimSpace(token), "`")
			if !strings.HasPrefix(token, "0x") {
				continue
			}
			value, parseErr := strconv.ParseUint(token[2:], 16, 8)
			if parseErr != nil {
				t.Fatalf("轉移圖的 %q 不是 block 編號", token)
			}
			target, found := segment.Lookup(token)
			if !found || target.Block != uint8(value) {
				t.Fatalf("轉移圖指到 %s，註冊表沒有這一段", token)
			}
			edges = append(edges, [2]segment.Segment{source, target})
		}
	}
	return edges
}

// SEG-12：47 條 `NEWECL` 邊每條一個交接。做的事是「在來源段存一份快照 →
// 用它當入口 → 帶著來源 block 當 `LastECL` 進目的段 → 目的段的 initial
// lifecycle 要跑得完」。
//
// ⚠ 來源用的是**段的入口**狀態，不是段真的走完的結束狀態——後者要等
// `SEG-10` 把每段拆成自己的測試。所以這一條擋的是「這條邊本身接不起來」，
// 不是「照劇情走到這裡會如何」。
func TestEveryNewECLEdgeHandsOff(t *testing.T) {
	blocks, records := segmentEntryBlocks(t)
	edges := segmentEdges(t)
	if len(edges) != 47 {
		t.Fatalf("轉移圖有 %d 條出邊，應該是 47 條", len(edges))
	}
	directory := t.TempDir()
	for _, edge := range edges {
		source, target := edge[0], edge[1]
		name := segmentSnapshotName(source) + "→" + segmentSnapshotName(target)
		t.Run(name, func(t *testing.T) {
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
			if err := origin.EnterSegment(source); err != nil {
				t.Fatalf("來源段進不去：%v", err)
			}
			path := filepath.Join(directory, strings.ReplaceAll(name, "→", "-to-")+".json")
			if err := origin.SavePartyFile(path); err != nil {
				t.Fatalf("來源段存不下去：%v", err)
			}

			next := NewStateFromECLBlocks(testCatalog(), blocks, 0x50)
			next.SetMonsterRecords(records)
			if err := next.LoadPartyFile(path); err != nil {
				t.Fatalf("來源段的快照讀不回來：%v", err)
			}
			if err := next.StartStorySegment(
				target.Block, source.Block, target.GameArea, !target.Overland,
			); err != nil {
				t.Fatalf("過這條邊進不去目的段：%v", err)
			}
			if got := next.session.CurrentBlockID(); got != target.Settles() {
				t.Fatalf("過完邊停在 block %#02x，宣告的是 %#02x", got, target.Settles())
			}
			if len(next.PartyFighters()) != len(origin.PartyFighters()) {
				t.Fatalf("隊伍沒跟過來：%d 人，原本 %d 人",
					len(next.PartyFighters()), len(origin.PartyFighters()))
			}
			if next.Area.InDungeon == target.Overland {
				t.Errorf("過完邊的所在層級是 InDungeon=%v，目的段宣告 Overland=%v",
					next.Area.InDungeon, target.Overland)
			}
			switch next.Mode {
			case ModeDungeon, ModeWilderness, ModeEvent, ModeCombat, ModePlace:
			default:
				t.Errorf("過完邊停在不該出現的模式 %v", next.Mode)
			}
		})
	}
}
