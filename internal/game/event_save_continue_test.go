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

// 讀回來之後**畫面上的字**也要對，不只是按得下繼續。
//
// ★ 上面那條補的是「按下一步會不會卡住」。畫面對不對是另一件事，而且
// `*State` 層的測試看不到：讀完檔就直接做下一個動作，`Prompt`／`Choices`
// 在被看到之前就被覆蓋掉了。用**真的前端**把戰役檢查點畫出來才看得到——
// 密斯卓諾墓園的存檔讀回來，畫面寫的是「隊伍已建立．準備開始冒險．」
// 加上「進入城市／繼續旅程／紮營」，也就是剛建好隊伍的世界地圖選單（spec 1188）。
func TestEventSaveRestoresTheScreenNotTheStartMenu(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.partyRoster = party.Roster{{
		ID: "p1", Name: "亞勇", Race: party.RaceHuman,
		Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 14, Intelligence: 10, Wisdom: 10,
			Dexterity: 12, Constitution: 12, Charisma: 10},
	}}
	state.Mode = ModeEvent
	state.Area.InDungeon = true
	state.Location = LocationMythDrannor

	path := filepath.Join(t.TempDir(), "event-screen.json")
	if err := state.SavePartyFile(path); err != nil {
		t.Fatal(err)
	}
	loaded := NewState(trainingTestCatalog(t))
	if err := loaded.LoadPartyFile(path); err != nil {
		t.Fatal(err)
	}

	// ⚠ **正對照**：同一支函式在世界地圖存檔上**仍然要**設出 hub 選單。
	// 少了這一條，「事件存檔沒有選單」在「讀檔根本不再設任何選單」時也會通過，
	// 而那是另一個 bug。
	state.Mode = ModeWilderness
	state.Area.InDungeon = false
	worldPath := filepath.Join(filepath.Dir(path), "world.json")
	if err := state.SavePartyFile(worldPath); err != nil {
		t.Fatal(err)
	}
	world := NewState(trainingTestCatalog(t))
	if err := world.LoadPartyFile(worldPath); err != nil {
		t.Fatal(err)
	}
	if len(world.Choices) != 3 {
		t.Fatalf("世界地圖存檔讀回來應該有 hub 選單，實際 %v", world.Choices)
	}
	if world.Prompt != world.catalog.Text("party_ready", "party_ready") {
		t.Fatalf("世界地圖存檔的提示 ＝ %q", world.Prompt)
	}
	if len(loaded.Choices) != 0 || len(loaded.currentOriginalChoices) != 0 {
		t.Errorf("事件畫面不該帶著世界地圖選單：%v", loaded.Choices)
	}
	if want := loaded.catalog.Text("press_enter", "Press Enter to continue"); loaded.Prompt != want {
		t.Errorf("提示 ＝ %q，want %q", loaded.Prompt, want)
	}
	if loaded.Prompt == loaded.catalog.Text("party_ready", "party_ready") {
		t.Error("讀回來還顯示「隊伍已建立」的開場提示")
	}
}

// 地名也要跟著還原。`Location` 有存，`LocationName`（畫面上那一行）沒有——
// 於是不論人在哪裡，標題永遠是建檔當下的那一個。
//
// ⚠ 還原的來源是**已經還原好的 `Location`**，不是 `Area.CurrentCity`：
// 地城與戰鬥存檔的 `CurrentCity` 可能是舊值或 0，拿它去推會把地名蓋成別的地方
// （`LoadPartyFile` 裡那段註解就是在講這個 hazard）。
func TestLoadRestoresTheLocationNameOutsideTheWorldMap(t *testing.T) {
	for _, item := range []struct {
		name     string
		mode     Mode
		location Location
		key      string
	}{
		{"地城", ModeDungeon, LocationYulash, "yulash"},
		{"事件", ModeEvent, LocationMythDrannor, "myth_drannor"},
	} {
		t.Run(item.name, func(t *testing.T) {
			state := NewState(trainingTestCatalog(t))
			state.partyRoster = party.Roster{{
				ID: "p1", Name: "亞勇", Race: party.RaceHuman,
				Class: party.ClassFighter, Level: 1,
				Abilities: party.Abilities{Strength: 14, Intelligence: 10, Wisdom: 10,
					Dexterity: 12, Constitution: 12, Charisma: 10},
			}}
			state.Mode = item.mode
			state.Area.InDungeon = true
			state.Location = item.location
			// ⚠ `CurrentCity` 故意留 0：這一格在地城存檔裡本來就可能是舊值，
			// 而還原**不可以**依賴它。留 0 才驗得到「沒有走那條路」。
			state.Area.CurrentCity = 0

			path := filepath.Join(t.TempDir(), "location.json")
			if err := state.SavePartyFile(path); err != nil {
				t.Fatal(err)
			}
			loaded := NewState(trainingTestCatalog(t))
			if err := loaded.LoadPartyFile(path); err != nil {
				t.Fatal(err)
			}
			if loaded.Location != item.location {
				t.Fatalf("地點 ＝ %v，want %v", loaded.Location, item.location)
			}
			want := loaded.catalog.Text(item.key, item.key)
			if loaded.LocationName != want {
				t.Errorf("畫面地名 ＝ %q，want %q", loaded.LocationName, want)
			}
			if loaded.LocationName == loaded.catalog.Text("wilderness", "Wilderness") {
				t.Error("讀回來還顯示建檔當下的「荒野」")
			}
		})
	}
}
