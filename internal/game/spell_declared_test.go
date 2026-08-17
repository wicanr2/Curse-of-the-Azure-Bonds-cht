package game

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	enginescan "github.com/wicanr2/golden-box-remake-engine/combat/scan"
)

// 每一支宣告過的戰鬥法術都要真的施得出來。
//
// ★ 這條擋的是「宣告數字好看但施法會噴錯」。先前保護法術那一輪就踩過：
// pack 加了兩筆宣告、覆蓋報告立刻變綠，實際施放時三個地方各噴一次錯——
// 因為那三處把法術編號與職業寫死了。宣告與可執行是兩件事，要分別檢查。
func TestEveryDeclaredCombatSpellIsCastable(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	for _, definition := range pack.CombatPlayerSpells {
		entry, found := gamepack.SpellByID(int(definition.SpellID))
		if !found {
			t.Fatalf("法術 0x%02X 不在原作主表裡", definition.SpellID)
		}
		state := declaredSpellState(t, definition.SpellID, definition.CasterClass)
		if err := state.BeginCombatCast(definition.SpellID); err != nil {
			t.Fatalf("%s（0x%02X，%s）開始施法就失敗：%v",
				entry.Name, definition.SpellID, definition.Behavior, err)
		}
		// 線形法術（閃電）要帶地形，別的法術帶了也不影響。
		terrain := func(x, y int) combat.LineCell {
			return combat.LineCell{Valid: x >= 0 && x < 8 && y >= 0 && y < 4}
		}
		if err := state.CombatCastWithTerrain(definition.SpellID, terrain); err != nil {
			t.Fatalf("%s（0x%02X，%s）施放失敗：%v",
				entry.Name, definition.SpellID, definition.Behavior, err)
		}
		if len(state.partyRoster[0].SpellSlots) != 0 {
			t.Fatalf("%s（0x%02X）施放後法術位沒有消耗", entry.Name, definition.SpellID)
		}
	}
}

// 每一支宣告過的法術施放時都要真的發得出聲音。
//
// ★ 這條**不靠** `cmd/combat-spell-coverage-audit`。那支稽核工具與被稽核的
// 施法路在同一輪一起改過（掃描範圍從單一檔案變成整包、behavior 對應從猜名字
// 變成跟分派表），**工具自己把數字從 26／73 改成 73／73 不構成證據**。
// 這裡直接跑一次施法、把 `ConsumeSoundEvents` 收到的東西攤開來看。
func TestEveryDeclaredCombatSpellEmitsSound(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	for _, definition := range pack.CombatPlayerSpells {
		entry, _ := gamepack.SpellByID(int(definition.SpellID))
		state := declaredSpellState(t, definition.SpellID, definition.CasterClass)
		if err := state.BeginCombatCast(definition.SpellID); err != nil {
			t.Fatalf("%s（0x%02X）開始施法失敗：%v", entry.Name, definition.SpellID, err)
		}
		state.ConsumeSoundEvents() // 先清掉開場與選目標留下的
		terrain := func(x, y int) combat.LineCell {
			return combat.LineCell{Valid: x >= 0 && x < 8 && y >= 0 && y < 4}
		}
		if err := state.CombatCastWithTerrain(definition.SpellID, terrain); err != nil {
			t.Fatalf("%s（0x%02X）施放失敗：%v", entry.Name, definition.SpellID, err)
		}
		events := state.ConsumeSoundEvents()
		// 有視覺演出的法術（雲霧、睡眠、火球…）把聲音留給時間軸發，
		// 所以要把演出推到結束再收一次。少了這一步會把「聲音在後面」
		// 誤判成「沒有聲音」。
		if _, playing := state.CombatVisualEvent(); playing {
			if err := state.AdvanceCombatVisual(visualTimelineEnd); err != nil {
				t.Fatalf("%s（0x%02X）推進視覺演出失敗：%v",
					entry.Name, definition.SpellID, err)
			}
			events = append(events, state.ConsumeSoundEvents()...)
		}
		if len(events) == 0 {
			t.Fatalf("%s（0x%02X，%s）施放時一個音效事件都沒有",
				entry.Name, definition.SpellID, definition.Behavior)
		}
		// ⚠ 不要求一定是 `SoundCast`：火球有自己的爆炸聲、閃電有雷聲、
		// 飛箭有弓弦聲。要求特定音效會把「有自己的聲音」判成錯的。
		for _, event := range events {
			if strings.TrimSpace(string(event)) == "" {
				t.Fatalf("%s（0x%02X）排了一個空的音效事件", entry.Name, definition.SpellID)
			}
		}
	}
}

// visualTimelineEnd 大於任何一段演出的長度，用來一次推到底。
const visualTimelineEnd = 10 * time.Second

// declaredSpellState 起一場只有一個施法者與一個敵人的戰鬥。
//
// 施法者兩個職業都有（牧師 ＋ 法師），這樣同一個骨架可以測兩邊的法術；
// 受傷是為了讓治療類有合法目標。
func declaredSpellState(t *testing.T, spellID uint8, casterClass string) *State {
	t.Helper()
	value := NewState(testCatalog())
	state := &value
	// 睡眠、閃電、雲霧那幾支要有戰術地圖才施得出來（掃描產生器）。
	state.SetCombatScanMapProvider(func() (enginescan.TacticalMap, error) {
		return enginescan.TacticalMap{
			Width: 8, Height: 4,
			Tiles:       bytes.Repeat([]byte{1}, 32),
			Definitions: []enginescan.TerrainDefinition{{LOS: 1, SYM: 0}},
		}, nil
	})
	class := party.ClassCleric
	if casterClass == "magic_user" {
		class = party.ClassMagicUser
	}
	state.partyRoster = party.Roster{{ID: "caster", Name: "施法者",
		Class: class, Level: 9, HitPoints: 20, MaxHitPoints: 40,
		BaseMaxHitPoints: 30,
		ClassLevels:      [8]uint8{0: 9, 5: 9},
		HealthStatus:     party.HealthStatusOK,
		Abilities: party.Abilities{Strength: 12, Intelligence: 16, Wisdom: 16,
			Dexterity: 12, Constitution: 14, Charisma: 10},
		SpellSlots: []uint8{spellID}}}
	partyFighters := []combat.Fighter{{ID: "caster", Name: "施法者", Side: combat.SideParty,
		HitPoints: 20, MaxHitPoints: 40, ArmorClass: 10, InitiativeBonus: 20,
		HasCombatPosition: true, CombatX: 1, CombatY: 1,
		SavingThrows: []uint8{14, 14, 14, 14, 14},
		// 治療失明要有失明的人、移除詛咒要有被詛咒的人，這兩支才有合法目標。
		MonsterAffects: []combat.MonsterAffect{
			{Kind: 0x21, Duration: 10, Active: true},
			{Kind: 0x24, Duration: 10, Active: true},
		}}}
	enemies := []combat.Fighter{{ID: "orc", Name: "獸人", Side: combat.SideEnemy,
		HitPoints: 40, MaxHitPoints: 40, ArmorClass: 10,
		HasCombatPosition: true, CombatX: 2, CombatY: 1,
		SavingThrows: []uint8{14, 14, 14, 14, 14}}}
	// 出貨的那條路（`cmd/azure-bonds-game`）會開視覺時間軸，雲霧那一類的
	// 聲音就是由時間軸發的。關著測會把「聲音在時間軸裡」誤判成沒有聲音。
	state.EnableCombatVisualTimeline(true)
	if err := state.StartCombat(partyFighters, enemies, 11); err != nil {
		t.Fatal(err)
	}
	return state
}

// 復原術把被吸掉的一級還回去：HP 三格各加 `少的 HP ÷ 少的級數`，
// 職業等級加一，經驗值補到那一級的門檻。
func TestRestorationReturnsOneDrainedLevel(t *testing.T) {
	character := party.Character{Class: party.ClassFighter,
		ClassLevels: [8]uint8{2: 4}, Level: 4,
		HitPoints: 20, MaxHitPoints: 30, BaseMaxHitPoints: 25,
		DrainedLevels: 2, DrainedHitPoints: 14, Experience: 100}
	if !restoreDrainedLevel(&character) {
		t.Fatal("有兩級被吸掉，應該還得回來")
	}
	if character.MaxHitPoints != 37 || character.BaseMaxHitPoints != 32 ||
		character.HitPoints != 27 {
		t.Fatalf("HP 三格是 %d／%d／%d，want 37／32／27",
			character.MaxHitPoints, character.BaseMaxHitPoints, character.HitPoints)
	}
	if character.DrainedLevels != 1 || character.DrainedHitPoints != 7 {
		t.Fatalf("剩下 %d 級／%d HP，want 1／7",
			character.DrainedLevels, character.DrainedHitPoints)
	}
	if character.ClassLevels[2] != 5 {
		t.Fatalf("戰士等級是 %d，want 5", character.ClassLevels[2])
	}
	// 戰士第 4 級的門檻是 18001，經驗值不足時補到剛好。
	if character.Experience != 18001 {
		t.Fatalf("經驗值是 %d，want 18001", character.Experience)
	}
}

// 沒有被吸掉的等級時什麼都不做——**不要**憑空加 HP。
func TestRestorationDoesNothingWithoutDrainedLevels(t *testing.T) {
	character := party.Character{Class: party.ClassFighter,
		ClassLevels: [8]uint8{2: 4}, HitPoints: 20, MaxHitPoints: 30}
	if restoreDrainedLevel(&character) {
		t.Fatal("沒有被吸掉的等級，不該回報還了")
	}
	if character.HitPoints != 20 || character.MaxHitPoints != 30 {
		t.Fatal("沒有可還的等級時不該動 HP")
	}
}

// 被吸掉的兩格要能寫進 DOS 角色記錄再讀回來，否則存檔一次就歸零，
// 而歸零之後復原術永遠是「沒有可恢復的等級」——看起來像法術壞了。
func TestDrainedLevelsSurviveTheDOSRecordRoundTrip(t *testing.T) {
	character := party.Character{Name: "測試", Class: party.ClassMagicUser,
		ClassLevels: [8]uint8{5: 4}, DrainedLevels: 3, DrainedHitPoints: 21}
	blank := make([]byte, party.DOSPlayerRecordSize)
	blank[0x74], blank[0x75] = 7, 5 // 人類法師：讀回來時種族／職業要合法
	data, err := party.PatchDOSPlayerRecord(blank, character)
	if err != nil {
		t.Fatal(err)
	}
	record, err := party.ParseDOSPlayerRecord(data, "測試")
	if err != nil {
		t.Fatal(err)
	}
	if record.DrainedLevels != 3 || record.DrainedHitPoints != 21 {
		t.Fatalf("讀回來是 %d 級／%d HP，want 3／21",
			record.DrainedLevels, record.DrainedHitPoints)
	}
}
