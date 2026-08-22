package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// keyDrivenApp 建一個**只有 `Update()` 需要的欄位**的 app，狀態走的是和正式
// 執行檔完全同一條路（`loadECLBlocks` ＋ `NewStateFromECLBlocks`）。
//
// ⚠ 畫面那一層不建：`Draw` 要的字型與圖在 headless 下拿不到，而這條測試問的是
// **按鍵到不到得了劇情**，不是畫得對不對（畫面由 spec 1188 的擷取負責）。
func keyDrivenApp(t *testing.T) (*app, *scriptedKeys) {
	t.Helper()
	root := filepath.Join("..", "..")
	imagePath := filepath.Join(root, "curseoftheazurebonds.zip")
	if _, err := os.Stat(imagePath); err != nil {
		t.Skipf("原版 image 不在：%v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "assets", "locale", "zh-TW.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		t.Fatal(err)
	}
	blocks, initial, err := loadECLBlocks(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	state := game.NewStateFromECLBlocks(catalog, blocks, initial)
	// ⚠ GEO 目錄不給的話，第一格就是「找不到 GEO2 block 0x01」而且**不會報錯**
	// ——畫面停在地城模式、按什麼都不動。正式執行檔走的是同一支 `loadGEOCatalog`。
	geoCatalog, err := loadGEOCatalog(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	state.SetGeoCatalog(geoCatalog)
	// ⚠ 寶物與物品表也要載：主線快照裡有一半的段會在走動時開出寶物，
	// 少了它 `TREASURE` 會直接回錯（`item block 0x05 for area 2 is not loaded`），
	// 而那看起來像「輸入層壞了」，其實是**測試自己少載資料**。
	if itemData, itemErr := zipMember(imagePath, "ITEMS"); itemErr == nil {
		if catalog, parseErr := monster.ParseBaseItems(itemData); parseErr == nil {
			state.SetItemCatalog(catalog)
		}
	}
	if blocks, blockErr := loadTreasureItemBlocks(imagePath); blockErr == nil {
		state.SetTreasureItemBlocks(blocks)
	}
	keys := newScriptedKeys()
	return &app{state: &state, keys: keys, geoCatalog: geoCatalog}, keys
}

// tap 按一顆鍵、跑一幀、再放開跑一幀。
//
// ⚠ 放開那一幀不能省：`JustPressed` 的語意是「這一幀剛按下」，不放開的話
// 下一次 `Update()` 會把同一顆鍵再吃一次。
func tap(t *testing.T, application *app, keys *scriptedKeys, key ebiten.Key) {
	t.Helper()
	keys.press(key)
	if err := application.Update(); err != nil {
		t.Fatalf("按 %v：%v", key, err)
	}
	keys.release()
	if err := application.Update(); err != nil {
		t.Fatalf("放開 %v：%v", key, err)
	}
}

// ★ 這條測試回答「開場到結局」那一列缺的那半：**按鍵到不到得了劇情**。
// 戰役測試直接呼叫 `state.X()`，前端的 `Update()` 一次都沒被跑到。
func TestKeysDriveTheOpeningFromTheTitle(t *testing.T) {
	application, keys := keyDrivenApp(t)
	if application.state.Mode != game.ModeTitle {
		t.Fatalf("新遊戲應該停在標題，實際 %v", application.state.Mode)
	}
	tap(t, application, keys, ebiten.KeyEnter)
	if application.state.Mode == game.ModeTitle {
		t.Fatal("按下 Enter 之後還停在標題：輸入那一層沒接上")
	}
	t.Logf("按一次 Enter 之後：mode=%v message=%.60q choices=%d",
		application.state.Mode, application.state.Message, len(application.state.Choices))
}

// canStepForward 問「照現在的朝向，往前踏得出去嗎」。用的是前端自己那份
// `geoGrid`，和 `moveDungeonPreview` 判斷的是同一份資料。
func (a *app) canStepForward() bool {
	if a.geoGrid == nil {
		return false
	}
	_, _, direction := a.state.DungeonGeometryView()
	deltaX, deltaY := 0, 0
	switch direction {
	case 0:
		deltaY = -1
	case 2:
		deltaX = 1
	case 4:
		deltaY = 1
	case 6:
		deltaX = -1
	}
	return a.state.CanMoveDungeon(*a.geoGrid, deltaX, deltaY, int(direction))
}

// modeName 讓紀錄看得懂。
func modeName(mode game.Mode) string {
	switch mode {
	case game.ModeTitle:
		return "標題"
	case game.ModeWilderness:
		return "荒野"
	case game.ModeEvent:
		return "事件"
	case game.ModeMap:
		return "地圖"
	case game.ModePlace:
		return "場所"
	case game.ModeCombat:
		return "戰鬥"
	case game.ModeJournal:
		return "手札"
	case game.ModeCharacterCreation:
		return "角色建立"
	case game.ModeDungeon:
		return "地城"
	}
	return "?"
}

// keyDrivenSession 是一次「只用按鍵」的遊玩紀錄。
type keyDrivenSession struct {
	app    *app
	keys   *scriptedKeys
	frames int
	// cells 是實際站上過的地城格（含 ECL 段），跨段不會互相蓋掉。
	cells map[[3]int]bool
	// modes 是走到過的畫面模式。
	modes map[game.Mode]bool
	// messages 是演出來的每一句話（去重）。
	messages map[string]bool
	// fallbacks 是**落回原文**的句子——這是這條測試真正在防的東西。
	fallbacks map[string]bool
	// blocks 是走到過的 ECL 段。
	blocks map[uint8]bool
	// doorsFound 是撞到門而開出選單的次數。
	doorsFound int
	// menus 是遇到過的選單（選項以「｜」相接），用來看卡在哪一個決定上。
	menus map[string]bool
}

func newKeyDrivenSession(t *testing.T) *keyDrivenSession {
	application, keys := keyDrivenApp(t)
	return &keyDrivenSession{
		app: application, keys: keys,
		cells: map[[3]int]bool{}, modes: map[game.Mode]bool{},
		messages: map[string]bool{}, fallbacks: map[string]bool{},
		blocks: map[uint8]bool{}, menus: map[string]bool{},
	}
}

// hasHan／hasLatinWord 與戰役測試同一套判準：有漢字就算翻好了；沒有漢字卻有
// 連續兩個以上的英文字母才算落回原文（單獨字母多半是數字旁的單位或原作代號）。
func hasHan(text string) bool {
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

func hasLatinWord(text string) bool {
	run := 0
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			run++
			if run >= 2 {
				return true
			}
			continue
		}
		run = 0
	}
	return false
}

// observe 記下這一幀玩家看得到的東西。
func (s *keyDrivenSession) observe() {
	state := s.app.state
	s.modes[state.Mode] = true
	if state.Mode == game.ModeDungeon || state.Mode == game.ModeEvent {
		x, y, _ := state.DungeonGeometryView()
		block := 0
		if s.app.geoGrid != nil {
			block = int(s.app.geoBlock)
		}
		s.cells[[3]int{block, x, y}] = true
	}
	s.blocks[s.app.geoBlock] = true
	if len(state.Choices) > 0 {
		s.menus[strings.Join(state.Choices, "｜")] = true
	}
	for _, text := range []string{state.Message, state.Prompt} {
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			continue
		}
		s.messages[trimmed] = true
		if !hasHan(trimmed) && hasLatinWord(trimmed) {
			s.fallbacks[trimmed] = true
		}
	}
}

// step 依畫面挑一顆鍵按下去。**只用玩家按得到的鍵**，不碰 `state` 的任何方法。
//
// ⚠ 地城裡要**兩輪**挑方向：先找「走得通而且沒去過」的，找不到才退回「走得通」
// 就好。少了第二輪，走到只剩一條回頭路的死巷時會**原地轉圈轉到天荒地老**——
// 那條路走得通，但因為去過所以被拒絕，於是永遠不動。
func (s *keyDrivenSession) step(t *testing.T) {
	t.Helper()
	if s.app.state.Mode != game.ModeDungeon || s.app.geoGrid == nil {
		tap(t, s.app, s.keys, ebiten.KeyEnter)
		return
	}
	// 門的選單開著就先處理掉：敲門 → 撬鎖 → 撞門，都失敗就退出。
	if s.app.dungeonDoorMenu {
		s.handleDoorMenu(t)
		return
	}
	target, fresh, ok := s.chooseHeading()
	// ⚠ 沒有**新**格子可走時就先去撞門，不要直接退回舊路。`chooseHeading` 只要
	// 有任何一個方向走得通就會給出退路，所以把「撞門」掛在「四面都是牆」底下
	// 等於永遠不會執行——隊伍會沿著已走過的路來回，而門一次都沒被碰過。
	if ok && !fresh && s.app.state.Mode == game.ModeDungeon {
		s.tryDoors(t)
		if s.app.dungeonDoorMenu || s.app.state.Mode != game.ModeDungeon {
			return
		}
		// 撞完一圈沒有門，才照退路走。
		target, _, ok = s.chooseHeading()
	}
	if !ok {
		// ⚠ 沒有走得通的方向**不代表**沒路：**門在 `CanMoveDungeon` 眼裡是牆**。
		// 玩家的做法是**撞上去**——`moveDungeonPreview` 發現擋住的是門（flags
		// 2／3）就開選單。只挑「走得通」的方向會讓隊伍永遠不去碰門，
		// 於是整段地城看起來只有開場那幾格。
		s.tryDoors(t)
		return
	}
	// M 是順時針轉 2；轉到面向目標為止（最多三次）。
	for turn := 0; turn < 4; turn++ {
		_, _, facing := s.app.state.DungeonGeometryView()
		if int(facing) == target {
			break
		}
		tap(t, s.app, s.keys, ebiten.KeyM)
		if s.app.state.Mode != game.ModeDungeon {
			return
		}
	}
	tap(t, s.app, s.keys, ebiten.KeyUp)
}

// tryDoors 朝四個方向各撞一次，看有沒有門。撞到門會開選單，交給下一幀處理。
func (s *keyDrivenSession) tryDoors(t *testing.T) {
	t.Helper()
	for turn := 0; turn < 4; turn++ {
		if s.app.state.Mode != game.ModeDungeon {
			return
		}
		tap(t, s.app, s.keys, ebiten.KeyUp)
		if s.app.dungeonDoorMenu {
			s.doorsFound++
			s.handleDoorMenu(t)
			return
		}
		x, y, _ := s.app.state.DungeonGeometryView()
		if !s.cells[[3]int{int(s.app.geoBlock), x, y}] {
			// 撞出去了（門本來就開著或劇情把隊伍搬走）。
			return
		}
		tap(t, s.app, s.keys, ebiten.KeyM)
	}
}

// handleDoorMenu 依原作提供的選項按鍵：能敲就敲、能撬就撬、能撞就撞。
//
// ⚠ 選項是原作算出來的（`DungeonDoorMenuOptions`），不要自己猜有哪些——
// 上鎖的門沒有「開」這一項，硬按只會什麼都不發生然後看起來像卡住。
func (s *keyDrivenSession) handleDoorMenu(t *testing.T) {
	t.Helper()
	flags, ok := s.app.dungeonDoorFlags()
	if !ok {
		tap(t, s.app, s.keys, ebiten.KeyEscape)
		return
	}
	options := s.app.state.DungeonDoorMenuOptions(flags)
	switch {
	case options.Knock:
		tap(t, s.app, s.keys, ebiten.KeyK)
	case options.Pick:
		tap(t, s.app, s.keys, ebiten.KeyP)
	case options.Bash:
		tap(t, s.app, s.keys, ebiten.KeyB)
	default:
		tap(t, s.app, s.keys, ebiten.KeyEscape)
	}
}

// chooseHeading 挑下一步要朝哪。先要沒去過的，沒有就退回任何走得通的。
//
// ⚠ 起點依幀數輪替：全部都去過時，固定從同一個方向找會讓隊伍在兩格之間
// 來回震盪。輪替是**決定性的**，重跑結果一樣。
// 第二個回傳值是「這個方向通往**沒去過**的格子」。
func (s *keyDrivenSession) chooseHeading() (int, bool, bool) {
	headings := []int{0, 2, 4, 6}
	offset := s.frames % len(headings)
	fallback, hasFallback := 0, false
	for index := range headings {
		heading := headings[(index+offset)%len(headings)]
		deltaX, deltaY := headingDelta(heading)
		if !s.app.state.CanMoveDungeon(*s.app.geoGrid, deltaX, deltaY, heading) {
			continue
		}
		if !hasFallback {
			fallback, hasFallback = heading, true
		}
		x, y, _ := s.app.state.DungeonGeometryView()
		if !s.cells[[3]int{int(s.app.geoBlock), x + deltaX, y + deltaY}] {
			return heading, true, true
		}
	}
	return fallback, false, hasFallback
}

// headingDelta 把 0／2／4／6 換成格子位移。
func headingDelta(heading int) (int, int) {
	switch heading {
	case 0:
		return 0, -1
	case 2:
		return 1, 0
	case 4:
		return 0, 1
	case 6:
		return -1, 0
	}
	return 0, 0
}

// forwardIsNew 問「照現在的朝向往前那一格去過了嗎」。
func (s *keyDrivenSession) forwardIsNew() bool {
	state := s.app.state
	x, y, direction := state.DungeonGeometryView()
	switch direction {
	case 0:
		y--
	case 2:
		x++
	case 4:
		y++
	case 6:
		x--
	}
	block := 0
	if s.app.geoGrid != nil {
		block = int(s.app.geoBlock)
	}
	return !s.cells[[3]int{block, x, y}]
}

// ★ 這條測試回答「開場到結局」那一列缺的那半：**輸入那一層**。
//
// 戰役測試（`TestRealNewGameRunsToTheEnding`）直接呼叫 `state.X()` 推進劇情，
// 前端的 `Update()` 一次都沒被跑到——所以「按鍵到不到得了那些劇情」在報表上
// 一直是空的。這條從標題畫面開始，**只按鍵**：建隊、離開建立畫面、走進地城、
// 觸發劇情，全程走 `(*app).Update()`，和玩家完全同一條路。
//
// ⚠ 這**不是**「按鍵玩到結局」。它證明的是開場那一段按得出來，以及沿路演出來的
// 每一句話都不是英文。走到多遠由 `docs/audit/key-driven-session.json` 記著。
func TestKeysDriveARealSessionFromTheTitle(t *testing.T) {
	session := newKeyDrivenSession(t)
	application, keys := session.app, session.keys
	if application.state.Mode != game.ModeTitle {
		t.Fatalf("新遊戲應該停在標題，實際 %v", application.state.Mode)
	}
	// 標題 → 角色建立 → 加六個角色 → D 完成建立。
	tap(t, application, keys, ebiten.KeyEnter)
	if application.state.Mode != game.ModeCharacterCreation {
		t.Fatalf("按 Enter 應該進角色建立，實際 %s", modeName(application.state.Mode))
	}
	for index := 0; index < 6; index++ {
		tap(t, application, keys, ebiten.KeyEnter)
	}
	if got := len(application.state.CreationRoster); got != 6 {
		t.Fatalf("按六次 Enter 應該有六名隊員，實際 %d", got)
	}
	tap(t, application, keys, ebiten.KeyD)
	if application.state.Mode == game.ModeCharacterCreation {
		t.Fatal("按 D 之後還停在角色建立：完成那條路按不出來")
	}

	for session.frames = 0; session.frames < 600; session.frames++ {
		session.observe()
		session.step(t)
	}
	session.observe()

	if !session.modes[game.ModeDungeon] {
		t.Fatal("整場都沒走進地城：地城那一層按鍵到不了")
	}
	if !session.modes[game.ModeEvent] {
		t.Fatal("整場都沒觸發任何劇情事件")
	}
	if len(session.cells) < 2 {
		t.Fatalf("只站上過 %d 格：走不動", len(session.cells))
	}
	if len(session.fallbacks) > 0 {
		for text := range session.fallbacks {
			t.Errorf("落回原文：%q", text)
		}
		t.Fatalf("按鍵玩到的畫面有 %d 句落回原文", len(session.fallbacks))
	}
	t.Logf("按鍵驅動 %d 幀：走過 %d 格、%d 種畫面、記到 %d 句話、撞到門 %d 次，落回原文 0 句",
		session.frames, len(session.cells), len(session.modes), len(session.messages),
		session.doorsFound)
	// ★ 卡住的原因記在這裡：整場**沒有出現任何真的選單**（只有「按任意鍵繼續」）。
	// 所以走不遠不是「選錯了選項」，是幾何上到不了——這兩種成因的處置完全不同，
	// 分不出來就會往錯的方向修。
	for menu := range session.menus {
		t.Logf("  選單 %s", menu)
	}

	if path := os.Getenv("COAB_KEY_SESSION_JSON"); path != "" {
		if err := session.writeReport(path); err != nil {
			t.Fatalf("報表寫不出來：%v", err)
		}
	}
}

// writeReport 把這一場的量測寫成 JSON，給 `cmd/remake-status` 取用。
//
// ⚠ 相對路徑是**相對 repo 根目錄**，不是測試的工作目錄（Go 把工作目錄設成套件
// 所在的資料夾，照字面開會開在 `cmd/azure-bonds-game/` 底下而失敗）。
func (s *keyDrivenSession) writeReport(path string) error {
	modes := make([]string, 0, len(s.modes))
	for mode := range s.modes {
		modes = append(modes, modeName(mode))
	}
	sort.Strings(modes)
	report := struct {
		Schema    string   `json:"schema"`
		Frames    int      `json:"frames"`
		Cells     int      `json:"cells"`
		Modes     []string `json:"modes"`
		Messages  int      `json:"messages"`
		Fallbacks int      `json:"fallbacks"`
		Doors     int      `json:"doors_found"`
		Menus     int      `json:"menus"`
	}{
		Schema: "coab-key-driven-session/1", Frames: s.frames, Cells: len(s.cells),
		Modes: modes, Messages: len(s.messages), Fallbacks: len(s.fallbacks),
		Doors: s.doorsFound, Menus: len(s.menus),
	}
	encoded, err := json.MarshalIndent(report, "", " ")
	if err != nil {
		return err
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join("..", "..", path)
	}
	return os.WriteFile(path, append(encoded, '\n'), 0o644)
}

// TestKeyDrivenDiagnose 印出逐幀的座標與四面通不通，用來查「走不動」是哪一種：
// 牆擋住、事件把隊伍推回來、還是挑方向的邏輯自己卡住。預設不跑。
func TestKeyDrivenDiagnose(t *testing.T) {
	if os.Getenv("COAB_KEY_EXPLORE") == "" {
		t.Skip("設 COAB_KEY_EXPLORE=1 才跑（診斷用，會印很多行）")
	}
	session := newKeyDrivenSession(t)
	application, keys := session.app, session.keys
	tap(t, application, keys, ebiten.KeyEnter)
	for index := 0; index < 6; index++ {
		tap(t, application, keys, ebiten.KeyEnter)
	}
	tap(t, application, keys, ebiten.KeyD)
	for frame := 0; frame < 30; frame++ {
		session.observe()
		session.step(t)
		state := application.state
		x, y, dir := state.DungeonGeometryView()
		if state.Mode != game.ModeDungeon {
			continue
		}
		can := [4]bool{}
		for index, d := range []int{0, 2, 4, 6} {
			dx, dy := 0, 0
			switch d {
			case 0:
				dy = -1
			case 2:
				dx = 1
			case 4:
				dy = 1
			case 6:
				dx = -1
			}
			if application.geoGrid != nil {
				can[index] = state.CanMoveDungeon(*application.geoGrid, dx, dy, d)
			}
		}
		t.Logf("幀%02d (%d,%d) 朝%d 北%v 東%v 南%v 西%v block=0x%02X",
			frame, x, y, dir, can[0], can[1], can[2], can[3], application.geoBlock)
	}
}

// ★ 這條把「按鍵到不到得了劇情」從**開場**擴到**整條主線**。
//
// 開場那條（`TestKeysDriveARealSessionFromTheTitle`）證明的是遊戲開得起來、
// 前幾段按得動。但主線後面那些段——猶拉什地下、散提爾堡、眼魔洞穴、密斯卓諾——
// 前端的 `Update()` 從來沒有在那些狀態下被跑過：戰役測試直接呼叫 `state.X()`。
//
// 主線在各段落存下來的快照就是那些狀態。這一條把每一份**讀進真的 app**，
// 用**按鍵**推它，確認：推得動、走得動、而且演出來的字沒有落回原文。
//
// ⚠ 這**不是**「按鍵從頭玩到結局」：每一份是各自載入的，不是一條連續的 session。
// 它證明的是**那些狀態下輸入層是活的**，不是「一路按過去到得了」。
func TestKeysDriveEveryCampaignSnapshot(t *testing.T) {
	dir := os.Getenv("COAB_CAMPAIGN_SNAPSHOT_DIR")
	if dir == "" {
		t.Skip("設 COAB_CAMPAIGN_SNAPSHOT_DIR 指向主線快照目錄才跑")
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join("..", "..", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("讀不到快照目錄：%v", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		t.Skip("快照目錄是空的")
	}

	driven, fallbacks := 0, map[string]string{}
	// ★ 已知的**變數插入**頁：那一頁印的是玩家自己取的隊員名字（`0x147B` 的
	// `7C00h` 插入點），固定譯文會把名字吃掉——所以規則層**故意不接**它，
	// 與 `gamepack` 的變數插入台帳（`TestVariableInsertPagesAreWiredAtRuntime`
	// 的 `隊員名字` 那一列）是同一個判斷。
	//
	// ⚠ 這份清單**故意很短而且不會自動長大**：任何**新的**英文落回照樣讓測試紅。
	// 它擋的是已知的量法限制，不是拿來讓報表好看的。
	knownVariableInsert := map[string]bool{
		"THE SPHERE MOVES TOWARD THE OPPOSING MAGE.": true,
	}
	for _, name := range names {
		application, keys := keyDrivenApp(t)
		if err := application.state.LoadPartyFile(filepath.Join(dir, name)); err != nil {
			t.Errorf("%s 讀不回來：%v", name, err)
			continue
		}
		// 先用按鍵把畫面推到可操作狀態，再走幾步。
		moved := false
		for frame := 0; frame < 60; frame++ {
			state := application.state
			for _, text := range []string{state.Message, state.Prompt} {
				trimmed := strings.TrimSpace(text)
				if trimmed != "" && !hasHan(trimmed) && hasLatinWord(trimmed) &&
					!knownVariableInsert[trimmed] {
					fallbacks[trimmed] = name
				}
			}
			// ⚠ 簽章要帶**朝向**：地城裡面對牆時按 M 只轉身，座標與訊息都不變，
			// 少了朝向會把「轉身」判成「沒有反應」——而那正是地城裡最常見的一步。
			_, _, facing := state.DungeonGeometryView()
			before := fmt.Sprintf("%v/%d/%d/%d/%q",
				state.Mode, state.DungeonX, state.DungeonY, facing, state.Message)
			key := ebiten.KeyEnter
			if state.Mode == game.ModeDungeon {
				key = ebiten.KeyUp
				if application.geoGrid != nil && !application.canStepForward() {
					key = ebiten.KeyM
				}
			}
			tap(t, application, keys, key)
			after := application.state
			_, _, afterFacing := after.DungeonGeometryView()
			if fmt.Sprintf("%v/%d/%d/%d/%q",
				after.Mode, after.DungeonX, after.DungeonY, afterFacing, after.Message) != before {
				moved = true
			}
		}
		if !moved {
			t.Errorf("%s：按了 60 幀完全沒有反應——那個狀態下輸入層是死的", name)
			continue
		}
		driven++
	}
	if len(fallbacks) > 0 {
		for text, name := range fallbacks {
			t.Errorf("按鍵推到的畫面落回原文（%s）：%q", name, text)
		}
	}
	t.Logf("按鍵推得動的快照 %d／%d，落回原文 %d 句", driven, len(names), len(fallbacks))
	if path := os.Getenv("COAB_KEY_SNAPSHOT_JSON"); path != "" {
		report := struct {
			Schema    string `json:"schema"`
			Snapshots int    `json:"snapshots"`
			Driven    int    `json:"driven"`
			Fallbacks int    `json:"fallbacks"`
			Known     int    `json:"known_variable_insert"`
		}{
			Schema: "coab-key-driven-snapshots/1", Snapshots: len(names), Driven: driven,
			Fallbacks: len(fallbacks), Known: len(knownVariableInsert),
		}
		encoded, err := json.MarshalIndent(report, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join("..", "..", path)
		}
		if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestKeyDrivenSnapshotDiagnose(t *testing.T) {
	if os.Getenv("COAB_KEY_EXPLORE") == "" {
		t.Skip("診斷用")
	}
	dir := filepath.Join("..", "..", os.Getenv("COAB_CAMPAIGN_SNAPSHOT_DIR"))
	for _, name := range []string{"inside-block-43.json", "inside-block-10.json", "inside-block-01.json"} {
		application, keys := keyDrivenApp(t)
		if err := application.state.LoadPartyFile(filepath.Join(dir, name)); err != nil {
			t.Logf("%s 讀不回來：%v", name, err)
			continue
		}
		s := application.state
		t.Logf("%s 載入後：mode=%s grid=%v geoBlock=%02X msg=%.30q choices=%d",
			name, modeName(s.Mode), application.geoGrid != nil, application.geoBlock,
			s.Message, len(s.Choices))
		tap(t, application, keys, ebiten.KeyEnter)
		s = application.state
		t.Logf("   按 Enter 後：mode=%s grid=%v msg=%.30q",
			modeName(s.Mode), application.geoGrid != nil, s.Message)
	}
}
