package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 存在事件畫面上的檔，讀回來要按得下「繼續」。
//
// ★ 存檔沒有保存「事件結束要回到哪裡」（`eventReturnMode`）。少了它，讀回來之後
// `Continue()` 會落到 default 分支回 `event has no continuation`——**玩家按下一步
// 就卡住，而且存檔裡每一個欄位都對得上**，欄位比對抓不到這種缺口。
//
// 回到哪裡由已經存下來的隊伍位置決定（地城裡回地城，否則回世界地圖），不是猜的。
func TestEventSaveCanContinueAfterReload(t *testing.T) {
	for _, mode := range []struct {
		name      string
		inDungeon bool
		want      Mode
	}{
		{"世界地圖上的事件", false, ModeWilderness},
		{"地城裡的事件", true, ModeDungeon},
	} {
		t.Run(mode.name, func(t *testing.T) {
			state := NewState(trainingTestCatalog(t))
			state.partyRoster = party.Roster{{
				ID: "p1", Name: "亞勇", Race: party.RaceHuman,
				Class: party.ClassFighter, Level: 1,
				Abilities: party.Abilities{Strength: 14, Intelligence: 10, Wisdom: 10,
					Dexterity: 12, Constitution: 12, Charisma: 10},
			}}
			state.Mode = ModeEvent
			state.Area.InDungeon = mode.inDungeon
			state.Message = "事件內容"

			path := filepath.Join(t.TempDir(), "event.json")
			if err := state.SavePartyFile(path); err != nil {
				t.Fatal(err)
			}
			loaded := NewState(locale.Catalog{Language: "zh-TW",
				Strings: map[string]string{"title": "t"}})
			if err := loaded.LoadPartyFile(path); err != nil {
				t.Fatal(err)
			}
			if loaded.Mode != ModeEvent {
				t.Fatalf("讀回來的模式是 %v，存的是事件", loaded.Mode)
			}
			if loaded.eventReturnMode != mode.want {
				t.Errorf("事件結束要回到 %v，讀回來是 %v", mode.want, loaded.eventReturnMode)
			}
			if err := loaded.Continue(); err != nil {
				t.Fatalf("讀回來按不下繼續：%v", err)
			}
		})
	}
}
