package game

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func programResult(id uint8) ecl.RunResult {
	return ecl.RunResult{ProgramExit: true, ProgramIDs: []uint8{id}}
}

func TestProgramCatalogCoversEveryDisplayedStableID(t *testing.T) {
	catalog := trainingTestCatalog(t)
	keys := []string{
		"program_main_menu", "program_party_killed_prompt", "program_return_title",
		"program_party_killed_message", "program_victory_prompt", "program_victory_save",
		"program_end_without_save", "program_victory_message",
		"program_victory_save_requested", "program_adventure_ended",
	}
	for _, key := range keys {
		if got := catalog.Text(key, key); got == key {
			t.Fatalf("PROGRAM locale ID %q is absent", key)
		}
	}
}

func TestApplyECLProgramStartMenu(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.Mode = ModeWilderness
	state.eclBlock = []byte{1}

	handled, err := state.applyECLProgram(programResult(0))
	if err != nil || !handled {
		t.Fatalf("PROGRAM 0: handled=%v err=%v", handled, err)
	}
	if state.Mode != ModeTitle || state.session != nil || len(state.eclBlock) != 0 {
		t.Fatalf("PROGRAM 0 did not reset to title: mode=%v block=%d", state.Mode, len(state.eclBlock))
	}
	if state.Message != catalog.Text("program_main_menu", "") {
		t.Fatalf("PROGRAM 0 message=%q", state.Message)
	}
}

// TestApplyECLProgramTrainingHall 釘住「`PROGRAM 0` ＋ `7EA8h == 7Fh` ＝ 訓練場」。
//
// ⚠ 地形碼**不是**判準。下水道那一段的訓練場任何一格隱密房間（地形 `8E`）
// 第一次踏進去就會問，而不是只有 `(11,10)` 那格 `8C`；拿地形碼當條件會讓
// 其餘 16 格落到「回主選單」，隊伍被丟出遊戲而畫面看起來只是回到標題。
func TestApplyECLProgramTrainingHall(t *testing.T) {
	for _, terrain := range []uint8{0x8C, 0x8E, 0x00} {
		state := newProgramState(t)
		state.Mode = ModeEvent
		state.eventReturnMode = ModeDungeon
		state.DungeonWallRoof = terrain
		state.session.SetMemoryValue(programSubFunctionCell, programSubFunctionTraining)
		state.partyRoster = party.Roster{{Name: "亞倫", Class: party.ClassFighter, Level: 1}}

		handled, err := state.applyECLProgram(programResult(0))
		if err != nil || !handled {
			t.Fatalf("地形 %#X：training PROGRAM 0 handled=%v err=%v", terrain, handled, err)
		}
		if state.Mode != ModeWilderness || !state.trainingMenu || len(state.Choices) != 2 {
			t.Fatalf("地形 %#X：沒有進訓練場 mode=%v menu=%v choices=%v",
				terrain, state.Mode, state.trainingMenu, state.Choices)
		}
	}
}

// TestApplyECLProgramWithoutTrainingSubFunctionGoesToTitle 是反面：
// `7EA8h` 不是 `7Fh` 時不可以進訓練場。少了這一條，把條件寫成「一律進訓練場」
// 也會過。
func TestApplyECLProgramWithoutTrainingSubFunctionGoesToTitle(t *testing.T) {
	state := newProgramState(t)
	state.Mode = ModeEvent
	state.eventReturnMode = ModeDungeon
	state.DungeonWallRoof = 0x8C
	state.session.SetMemoryValue(programSubFunctionCell, 0x7C)
	state.partyRoster = party.Roster{{Name: "亞倫", Class: party.ClassFighter, Level: 1}}

	handled, err := state.applyECLProgram(programResult(0))
	if err != nil || !handled {
		t.Fatalf("PROGRAM 0: handled=%v err=%v", handled, err)
	}
	if state.trainingMenu {
		t.Fatal("`7EA8h` 不是 7Fh 卻進了訓練場")
	}
	if state.Mode != ModeTitle {
		t.Fatalf("PROGRAM 0 沒有回標題：mode=%v", state.Mode)
	}
}

// newProgramState 給一個帶 session 的最小 State：`PROGRAM 0` 的判準讀的是
// ECL 記憶體的 `7EA8h`，沒有 session 就讀不到。
func newProgramState(t *testing.T) *State {
	t.Helper()
	session, err := ecl.NewBlockSession(map[uint8][]byte{0x03: {0, 0}}, 0x03)
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(trainingTestCatalog(t))
	state.session = session
	return &state
}

func TestHallChoiceRecognizesTrainingProgramContext(t *testing.T) {
	if !isTrainingProgramChoice("HALL", programResult(0)) {
		t.Fatal("HALL PROGRAM 0 was not recognized as training")
	}
	if isTrainingProgramChoice("BAR", programResult(0)) ||
		isTrainingProgramChoice("HALL", programResult(3)) {
		t.Fatal("unrelated program context was misclassified as training")
	}
}

func TestApplyECLProgramPartyKilled(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)

	handled, err := state.applyECLProgram(programResult(3))
	if err != nil || !handled || !state.PartyKilled() {
		t.Fatalf("PROGRAM 3: handled=%v killed=%v err=%v", handled, state.PartyKilled(), err)
	}
	if state.Mode != ModeWilderness || len(state.Choices) != 1 {
		t.Fatalf("PROGRAM 3 did not expose terminal menu: mode=%v choices=%v", state.Mode, state.Choices)
	}
	if state.Prompt != catalog.Text("program_party_killed_prompt", "") ||
		state.Choices[0] != catalog.Text("program_return_title", "") ||
		state.Message != catalog.Text("program_party_killed_message", "") {
		t.Fatalf("PROGRAM 3 localized terminal=%q choices=%v message=%q",
			state.Prompt, state.Choices, state.Message)
	}
	if got, want := state.ConsumeSoundEvents(), []SoundEvent{SoundCrash}; !reflect.DeepEqual(got, want) {
		t.Fatalf("PROGRAM 3 sound events=%v, want %v", got, want)
	}
	if err := state.Select(0); err != nil || state.Mode != ModeTitle {
		t.Fatalf("PROGRAM 3 title choice: mode=%v err=%v", state.Mode, err)
	}
	if state.Message != catalog.Text("program_adventure_ended", "") {
		t.Fatalf("PROGRAM 3 title message=%q", state.Message)
	}
}

func TestApplyECLProgramVictoryHealsAndRequestsSave(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{
		ID: "hero", HitPoints: 0, MaxHitPoints: 12,
		HealthStatus: party.HealthStatusDying, Bleeding: 4,
	}}
	state.party = []combat.Fighter{{
		ID: "hero", Side: combat.SideParty, HitPoints: 0, MaxHitPoints: 12,
		DeathOverlay: true, DownedCorpse: true,
	}}

	handled, err := state.applyECLProgram(programResult(8))
	if err != nil || !handled || !state.GameWon() {
		t.Fatalf("PROGRAM 8: handled=%v won=%v err=%v", handled, state.GameWon(), err)
	}
	character, fighter := state.partyRoster[0], state.party[0]
	if character.HitPoints != 12 || character.HealthStatus != party.HealthStatusOK || character.Bleeding != 0 {
		t.Fatalf("PROGRAM 8 roster recovery=%+v", character)
	}
	if fighter.HitPoints != 12 || fighter.DeathOverlay || fighter.DownedCorpse {
		t.Fatalf("PROGRAM 8 fighter recovery=%+v", fighter)
	}
	// ★ 存檔詢問排在結局過場**之後**（spec 1154）。
	if err := advanceEndingScene(t, &state); err != nil {
		t.Fatal(err)
	}
	if state.Prompt != catalog.Text("program_victory_prompt", "") ||
		len(state.Choices) != 2 || state.Choices[0] != catalog.Text("program_victory_save", "") ||
		state.Choices[1] != catalog.Text("program_end_without_save", "") ||
		state.Message != catalog.Text("program_victory_message", "") {
		t.Fatalf("PROGRAM 8 localized terminal=%q choices=%v message=%q",
			state.Prompt, state.Choices, state.Message)
	}
	if err := state.Select(0); err != nil {
		t.Fatalf("victory save choice: %v", err)
	}
	if state.Mode != ModeTitle || !state.ConsumeSaveRequest() {
		t.Fatalf("victory save did not return title/request save: mode=%v", state.Mode)
	}
	if state.Message != catalog.Text("program_victory_save_requested", "") {
		t.Fatalf("victory save title message=%q", state.Message)
	}
}

func TestApplyECLProgramVictoryCanEndWithoutSave(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{ID: "hero", HitPoints: 1, MaxHitPoints: 1}}

	handled, err := state.applyECLProgram(programResult(8))
	if err != nil || !handled {
		t.Fatalf("PROGRAM 8: handled=%v err=%v", handled, err)
	}
	if err := advanceEndingScene(t, &state); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeTitle || state.ConsumeSaveRequest() ||
		state.Message != catalog.Text("program_adventure_ended", "") {
		t.Fatalf("victory no-save mode=%v requested=%v message=%q",
			state.Mode, state.saveRequested, state.Message)
	}
}

func TestApplyECLProgramCamp(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.Mode = ModeEvent
	handled, err := state.applyECLProgram(programResult(9))
	if err != nil || !handled {
		t.Fatalf("PROGRAM 9: handled=%v err=%v", handled, err)
	}
	if state.Mode != ModeWilderness || !state.campMenu {
		t.Fatalf("PROGRAM 9 did not enter camp: mode=%v camp=%v", state.Mode, state.campMenu)
	}
}

// advanceEndingScene 把結局過場整個翻完，讓斷言回到「過場之後」那一刻。
func advanceEndingScene(t *testing.T, state *State) error {
	t.Helper()
	for guard := 0; state.endingScene; guard++ {
		if guard > len(endingSceneKeys) {
			return fmt.Errorf("結局過場翻不完，停在第 %d 頁", state.endingPageIndex)
		}
		if err := state.Select(0); err != nil {
			return err
		}
	}
	return nil
}

// TestApplyECLProgramVictoryPlaysTheEndingScene 釘住結局過場本身：
// 原作 `PROGRAM 8` 先跑 `overlay-18:10FFh` 那段五頁的敘事才回主選單問存檔
// （spec 1082／1154）。先前 remake 直接跳到存檔詢問，打通關一句結局都看不到。
func TestApplyECLProgramVictoryPlaysTheEndingScene(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewState(catalog)
	state.partyRoster = party.Roster{{ID: "hero", HitPoints: 1, MaxHitPoints: 1}}

	if _, err := state.applyECLProgram(programResult(8)); err != nil {
		t.Fatal(err)
	}
	for page := range endingSceneKeys {
		if !state.endingScene || state.endingPageIndex != page {
			t.Fatalf("第 %d 頁：endingScene=%v index=%d", page+1, state.endingScene, state.endingPageIndex)
		}
		want := catalog.Text(endingSceneKeys[page], "")
		if want == "" {
			t.Fatalf("結局第 %d 頁沒有譯文（key %q）", page+1, endingSceneKeys[page])
		}
		if state.Message != want {
			t.Fatalf("結局第 %d 頁 ＝ %q，預期 %q", page+1, state.Message, want)
		}
		if len(state.Choices) != 1 || len(state.currentOriginalChoices) != 1 ||
			state.currentOriginalChoices[0] != "PRESS BUTTON OR RETURN TO CONTINUE." {
			t.Fatalf("結局第 %d 頁的選項 ＝ %v／%v", page+1, state.Choices, state.currentOriginalChoices)
		}
		if state.programEndMenu {
			t.Fatalf("結局第 %d 頁就開了存檔選單", page+1)
		}
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}
	if state.endingScene {
		t.Fatal("五頁演完之後過場還沒結束")
	}
	if !state.programEndMenu || len(state.Choices) != 2 {
		t.Fatalf("過場之後沒有進勝利選單：menu=%v choices=%v", state.programEndMenu, state.Choices)
	}
}

// TestEndingSceneTextIsDistinctPerPage 擋掉「五頁都指到同一個 key」這種
// 看起來會過的接法——每一頁的譯文必須互不相同。
func TestEndingSceneTextIsDistinctPerPage(t *testing.T) {
	catalog := trainingTestCatalog(t)
	seen := make(map[string]int, len(endingSceneKeys))
	for index, key := range endingSceneKeys {
		text := catalog.Text(key, "")
		if text == "" {
			t.Fatalf("第 %d 頁的 key %q 沒有譯文", index+1, key)
		}
		if previous, ok := seen[text]; ok {
			t.Fatalf("第 %d 頁與第 %d 頁的譯文相同", index+1, previous+1)
		}
		seen[text] = index
	}
}

func TestPrepareEndingScenePreviewUsesNormalPageSelections(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	if err := state.PrepareEndingScenePreview(4); err != nil {
		t.Fatal(err)
	}
	page, ok := state.EndingScenePage()
	if !ok || page != 3 {
		t.Fatalf("EndingScenePage() = (%d, %v), want zero-based page 3", page, ok)
	}
	if state.Message != state.catalog.Text("ending_page_4", "") || len(state.Choices) != 1 {
		t.Fatalf("page 4 state: message=%q choices=%v", state.Message, state.Choices)
	}
	if err := state.PrepareEndingScenePreview(0); err == nil {
		t.Fatal("page zero unexpectedly accepted")
	}
	if err := state.PrepareEndingScenePreview(6); err == nil {
		t.Fatal("page six unexpectedly accepted")
	}
}

// 全滅回標題之後再開新局，章節與地圖狀態要收回開局值（seg001.Init）。
//
// ★ 這條釘住的缺陷：`resetSessionForNewGame` 先前只重設 session 與隊伍，
// `Area.GameArea` 留著上一局的章。全新開局不經過「ENTER CITY」那條會改章的路
// （`BeginAdventure` 直接 Reset 到 0x01），於是重開的新隊伍帶著章 6 走進提爾佛頓：
// GEO 載成「找不到 GEO6 block 0x01」、商店 TREASURE 拿章 6 查 ITEM 區塊 1 而整個
// `Update()` 回錯（實測擋住五份 `inside-block-42-*` 快照走訪）。
func TestReturnToTitleAfterPartyKilledResetsCampaignArea(t *testing.T) {
	catalog := trainingTestCatalog(t)
	state := NewStateFromECLBlocks(catalog, map[uint8][]byte{0x01: {}}, 0x01)

	// 模擬一局走到密斯卓諾（章 6）時全滅的狀態。
	state.Area.GameArea = 6
	state.Area.InDungeon = true
	state.GeoMapSet = 6
	state.GeoMapBlock = 0x42
	state.DungeonX, state.DungeonY, state.DungeonDirection = 4, 9, 2

	if handled, err := state.applyECLProgram(programResult(3)); err != nil || !handled {
		t.Fatalf("PROGRAM 3: handled=%v err=%v", handled, err)
	}
	if err := state.Select(0); err != nil || state.Mode != ModeTitle {
		t.Fatalf("返回標題：mode=%v err=%v", state.Mode, err)
	}

	if state.Area.GameArea != 2 || state.Area.InDungeon {
		t.Fatalf("章節沒有收回開局值：GameArea=%d InDungeon=%v",
			state.Area.GameArea, state.Area.InDungeon)
	}
	if state.GeoMapSet != 2 || state.GeoMapBlock != 1 {
		t.Fatalf("GEO 檔集沒有收回開局值：set=%d block=%02X",
			state.GeoMapSet, state.GeoMapBlock)
	}
	if set, block, ok := state.ConsumeGeoMapRequest(); !ok || set != 2 || block != 1 {
		t.Fatalf("前端沒有收到重載 GEO2/0x01 的請求：set=%d block=%02X ok=%v",
			set, block, ok)
	}
	if state.DungeonX != 7 || state.DungeonY != 13 || state.DungeonDirection != 0 {
		t.Fatalf("隊伍座標沒有收回開局值：(%d,%d) dir=%d",
			state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
}
