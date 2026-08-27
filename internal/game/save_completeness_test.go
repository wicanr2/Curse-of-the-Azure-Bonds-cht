package game

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

// 存檔完整性的第二道閘：**存檔 → State → 存檔**，逐欄比對。
//
// ★ 為什麼是這個方向。 `SavePartyFile` 用 20 個位置參數呼叫
// `EncodeGameWithAdventureState`——新增欄位忘了傳，或同型別的兩個參數對調，
// 編譯器都不會吭聲，而症狀是「讀檔之後某個狀態悄悄不見了」。
// 把一份填滿的存檔讀進 State 再存出來，掉的欄位當場現形。
//
// ⚠ 這條測試不驗語意，只驗「進得去、出得來」。
var gameFileRoundTripExemptions = map[string]string{
	// 戰鬥快照要有活著的 Battle 才存得出來，而 Battle 要靠遊戲流程建立；
	// 它自己的完整性由 `internal/combat` 的反射閘與既有的戰鬥存讀測試守。
	"Combat": "需要活的 Battle；由 internal/combat 的快照閘與戰鬥存讀測試覆蓋",
	// 音樂串流快照要有音訊裝置，測試環境沒有。
	"Music": "需要音訊裝置",
	// 一次性音效同上。
	"OneShotAudio": "需要音訊裝置",
	// ECL session 快照要有載入中的 ECL block。
	"ECLSession": "需要載入中的 ECL block",
}

func TestGameFileSurvivesAStateRoundTrip(t *testing.T) {
	source := partySave.GameFile{
		Version: partySave.CurrentGameVersion,
		Characters: party.Roster{{
			ID: "hero", Name: "英雄", Class: party.ClassFighter, Level: 3,
			ClassLevels: [8]uint8{2: 3}, HitPoints: 18, MaxHitPoints: 18,
			HealthStatus: party.HealthStatusOK,
			Abilities:    party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10},
		}},
		// ⚠ 時鐘存兩份：頂層的 `GameTime` 與 `Area.GameTime`。存檔時後者由
		// `s.gameClock` 覆寫，所以兩邊必須一致——只填一邊會在往返之後對不上，
		// 那不是掉欄位，是這份重複本身。
		Area: area.State{GameArea: 3, Current3DMapBlockID: 0x21, InDungeon: true,
			GameTime: [7]uint16{1, 2, 3, 4, 5, 6, 7}},
		Mode:               uint8(ModeDungeon),
		Location:           2,
		MapX:               11,
		MapY:               7,
		DungeonX:           5,
		DungeonY:           9,
		DungeonDir:         2,
		DungeonWallType:    4,
		DungeonWallRoof:    6,
		GameTime:           [7]uint16{1, 2, 3, 4, 5, 6, 7},
		GameAgeCycles:      1234,
		JournalMessageIDs:  []string{"journal.12.1"},
		DungeonSearch:      true,
		DungeonSearchEdges: []string{"tilverton.sewers.wall-09-west"},
	}
	encoded, err := partySave.EncodeGameFile(source)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	inPath := filepath.Join(directory, "in.json")
	if err := os.WriteFile(inPath, encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	state := NewState(testCatalog())
	if err := state.LoadPartyFile(inPath); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(directory, "out.json")
	if err := state.SavePartyFile(outPath); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	got, err := partySave.DecodeGame(written)
	if err != nil {
		t.Fatal(err)
	}

	wantValue, gotValue := reflect.ValueOf(source), reflect.ValueOf(got)
	fileType := wantValue.Type()
	checked := 0
	for index := 0; index < wantValue.NumField(); index++ {
		name := fileType.Field(index).Name
		if reason, exempt := gameFileRoundTripExemptions[name]; exempt {
			// 豁免只給「需要外部資源才組得出來」的快照指標。純量欄位沒有這種
			// 理由——它一定是被漏掉了，不是存不出來。
			if fileType.Field(index).Type.Kind() != reflect.Ptr {
				t.Errorf("GameFile.%s 不是快照指標卻被豁免（%s）：純量欄位沒有「存不出來」這回事",
					name, reason)
			}
			if !gotValue.Field(index).IsZero() {
				t.Errorf("GameFile.%s 被豁免（%s）卻有值，豁免理由要重看", name, reason)
			}
			continue
		}
		checked++
		if !reflect.DeepEqual(wantValue.Field(index).Interface(), gotValue.Field(index).Interface()) {
			t.Errorf("GameFile.%s 沒有存回來：存前 %#v，讀回 %#v",
				name, wantValue.Field(index).Interface(), gotValue.Field(index).Interface())
		}
	}
	if checked == 0 {
		t.Fatal("一個欄位都沒比對到，反射走空了")
	}
	t.Logf("逐欄比對了 %d 個欄位，豁免 %d 個", checked, len(gameFileRoundTripExemptions))
}
