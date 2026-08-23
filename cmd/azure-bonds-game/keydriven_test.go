package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"path/filepath"
	"sort"
	"strconv"
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

// routeChoice 回答「主線在這個選單上選了哪一項」。
//
// ⚠ 比對的是**選項文字**不是索引：索引會隨選單內容錯位。
func (s *keyDrivenSession) routeChoice() (int, bool) {
	// ⚠ **只有真的岔路才查路線**。「請按任意鍵或 Enter 繼續」這種單選項在整條
	// 主線裡出現幾千次，讓它們也去查表只會把紀錄用光，隊伍卻一步都沒往前。
	if len(s.app.state.Choices) < 2 {
		return 0, false
	}
	signature := strings.Join(s.app.state.Choices, "｜")
	// ⚠ **路線要有放手條件。** 紮營選單在路線裡出現 2,011 次、其中 2,008 次選
	// 「儲存」——那是錄路線的測試每到檢查點就存一份快照留下的痕跡。照路線按下去
	// 畫面永遠不會有新東西，而候選還有兩千筆，隊伍就在那裡存檔存到跑完
	// （不設這個條件時 1500 幀裡有 700 幀是這樣耗掉的）。同一個選單從上一次有新
	// 東西到現在出現這麼多次，就當作「路線在這裡幫不上忙」，交給啟發式。
	//
	// 實測（4000 幀）三種處置：
	//
	//	不放手           137 格／126 句／6 段（多摸到 `0x51`，之後死在紮營選單）
	//	放手給啟發式     137 格／**131 句**／5 段  ← 目前這一版
	//	跳到「選法不同」的候選  137 格／103 句／5 段
	//
	// ⇒ 沒有一種全面贏。取「放手給啟發式」是因為它句數最多而且**不會讓隊伍把
	// 大半場耗在同一個選單裡**；`0x51` 那一段在另一版是 `0x50` ↔ `0x51` 來回摸到的，
	// 不是留得住的進度。改動走法之後要重新量這三列。
	if s.menuSinceProgress[stuckSignature(s.app.state.Choices)] > routeMenuPatienceValue() {
		return 0, false
	}
	index, ok := s.takeRouteStep("menu:"+signature, s.routeChoices[signature])
	if !ok {
		return 0, false
	}
	return s.route[index].Index, true
}

// routeMove 回答「主線站在這一格時往哪個方向走」。
func (s *keyDrivenSession) routeMove() (int, bool) {
	x, y, _ := s.app.state.DungeonGeometryView()
	block, _ := s.app.state.CurrentECLBlockID()
	cell := routeCell{int(block), x, y}
	index, ok := s.takeRouteStep(fmt.Sprintf("cell:%d/%d/%d", cell.segment, cell.x, cell.y), s.routeMoves[cell])
	if !ok {
		return 0, false
	}
	return s.route[index].Direction, true
}

// takeRouteStep 從一串候選裡輪流拿一個。
//
// ⚠ **不能每次都拿第一個**：那會讓隊伍站在同一格一直按同一個方向，原地震盪
// （舊的指標式重放實測「照路線按 759 次卻只走 57 格」就是這個成因）。
//
// ⚠ 也**不要用過就永久劃掉**：主線在同一格會來回經過很多次，一格的候選用完之後
// 隊伍就再也拿不到路線——實測會在 16 格之間繞掉整整 762 幀。輪流拿兩邊都避開：
// 候選有幾個就輪幾個，回到同一格會換下一個方向。
func (s *keyDrivenSession) takeRouteStep(key string, candidates []int) (int, bool) {
	if len(candidates) == 0 {
		return 0, false
	}
	cursor := s.routeCursor[key] % len(candidates)
	s.routeCursor[key]++
	return candidates[cursor], true
}

// ⚠ **不要把重複的選法去掉。** 直覺上該去：紮營選單在路線裡出現 2,011 次，
// 其中 2,008 次選的是「儲存」——那是錄路線的測試每到一個檢查點就存一份快照留下的
// 痕跡，不是玩家的意圖，而輪流取候選會被它稀釋掉。實測（1500 幀，同一份路線，
// cap ＝「同一個選單上同一個選法最多留幾筆」）：
//
//	cap=1   168 格／112 句／最深 0x50
//	cap=2   134 格／ 99 句／最深 0x03
//	cap=3   114 格／ 87 句／最深 0x03
//	cap=6   113 格／ 91 句／最深 0x03
//	cap=12  137 格／120 句／最深 0x51
//	不設限  137 格／124 句／最深 0x51   ← 目前這一版
//
// **兩邊都不是線性的**，而且「走過的格」與「走到多深」會往相反方向動：cap=1 的
// 格數最多（隊伍把時間花在提爾佛頓）卻走不到世界地圖的第二段。要看的是段序列。
// ⇒ 不設限最深、句數最多，而且程式最短。改動走法之後要重新量這張表。

// routeMenuPatience 是「同一個選單連續出現幾次都沒有新東西，就不再問路線」。
const routeMenuPatience = 8

func routeMenuPatienceValue() int {
	if raw := os.Getenv("COAB_KEY_MENU_PATIENCE"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			return value
		}
	}
	return routeMenuPatience
}

// routeCell 是「這一段的這一格」。**段號不能省**：不同段的地圖上都有 (7,13)。
type routeCell struct{ segment, x, y int }

// buildRouteIndex 把路線翻成兩張查得到的表。
//
// ★ 為什麼要換掉往前找的視窗。 原本的重放是「從目前的進度往前找 N 個決策」，
// 而 N 兩邊都不是線性的：太小在第一個岔開的地方就跟丟，太大會把後面的決策提早
// 消耗掉。更根本的問題是**它假設重放和主線走的順序一樣**——按鍵這一場會多走一些
// 冤枉路、少走一些捷徑，順序一旦錯開，視窗再大也只是把錯的步驟吃掉。
//
// 查表式沒有這個假設：**站在哪就查哪一格**，路線裡曾經從這一格走出去的步驟一定
// 找得到，跟中間繞了多遠無關。代價是同一格被走過很多次時，只能照錄下來的順序
// 一個一個用——那是現有資料能給的最好答案。
func buildRouteIndex(route []game.Decision) (map[routeCell][]int, map[string][]int) {
	moves := map[routeCell][]int{}
	choices := map[string][]int{}
	for index, step := range route {
		switch step.Kind {
		case "move":
			key := routeCell{step.Segment, step.FromX, step.FromY}
			moves[key] = append(moves[key], index)
		case "select":
			if len(step.Choices) < 2 {
				continue
			}
			signature := strings.Join(step.Choices, "｜")
			choices[signature] = append(choices[signature], index)
		}
	}
	return moves, choices
}

// moveCursorTo 用方向鍵把選單游標挪到指定項；挪不到就回 false。
func (s *keyDrivenSession) moveCursorTo(t *testing.T, want int) bool {
	t.Helper()
	if want < 0 || want >= len(s.app.state.Choices) {
		return false
	}
	for step := 0; step < len(s.app.state.Choices)+8 && s.app.choiceCursor != want; step++ {
		key := ebiten.KeyDown
		if s.app.choiceCursor > want {
			key = ebiten.KeyUp
		}
		tap(t, s.app, s.keys, key)
		if s.app.state.Mode == game.ModeDungeon {
			return false
		}
	}
	return s.app.choiceCursor == want && want < len(s.app.state.Choices)
}


// keyDrivenMenuPatience 是「同一個選單按同一項幾次還在原地，就換下一項」。
//
// ⚠ 太小會變成輪流選（實測會自己切斷路線）；太大就等於永遠第一項。
//
// ⚠ 實測 2／3／6 三個值結果**完全一樣**（28 格／37 句／只到 `0x01`）——
// 所以**限制不在選單策略上**。再調這個數字不會有任何改變；要離開開場那一段，
// 缺的是**路線知識**（戰役測試裡逐段寫死的那一份）。
const keyDrivenMenuPatience = 6

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
	// tried 是「從這個方向踏進這一格」試過沒有——見 `chooseHeading` 的說明。
	tried map[[4]int]bool
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
	// menuSeen 記著同一個選單看過幾次，用來判斷「卡住了，換下一項」。
	menuSeen map[string]int
	// route 是主線錄下來的選擇（`COAB_DECISION_LOG`）。
	// ★ 這就是「路線知識」：戰役測試裡逐段寫死的劇情決策，用 `State.Select`
	// 錄成一份可重放的清單，重放端用**按鍵**把游標挪到那一項（spec 1191）。
	route []game.Decision
	// routeMoves／routeChoices 是路線翻出來的查表（見 `buildRouteIndex`）。
	routeMoves   map[routeCell][]int
	routeChoices map[string][]int
	// routeHits 是照著路線按下去的次數，用來看路線真的被用到多少。
	routeHits int
	// routeCursor 記著每一格／每一個選單輪到第幾個候選。
	routeCursor map[string]int
	// segmentTrace 是 ECL 段的**變化**序列（含幀號），用來看劇情走到哪一步、
	// 又是在哪一幀被拉回去的。只記變化，不記每一幀。
	segmentTrace []string
	// moveTrace 是逐格的走位紀錄，`COAB_KEY_TRACE` 指定輸出檔才會收集。
	// 用來回答「隊伍到底有沒有站上那一格、站上去之後路線叫它往哪走」。
	moveTrace []string
	// menuSinceProgress 是「從上一次有新東西到現在，這個選單出現過幾次」。
	// ★ 這是路線的**放手條件**：路線在某個選單上一直給同一個答案而畫面沒有任何
	// 新東西時，那個答案在這一場就是錯的，該讓啟發式接手。
	menuSinceProgress map[string]int
	// modeFrames 是每一種畫面待了幾幀；lastProgress 是最後一次踏上新格子的幀號。
	// ★ 這兩個回答的是「**停在哪裡**」——沒有它們，「跑到頂了」與「卡在商店裡」
	// 看起來一模一樣。
	modeFrames   map[game.Mode]int
	lastProgress int
	// stallCells 是「最後一次踏上新格子之後」還在原地繞的那些格子。
	stallCells map[[3]int]int
	// routeBlocked 是「路線說往這邊走，按下去卻沒動」的次數——這個數字把
	// 「路線用完了」與「路線帶不動」分開。
	routeBlocked int
	// searchToggles／looks 是被牆擋住之後按 `S`／`L` 的次數。
	searchToggles int
	looks         int
}

// loadRoute 讀主線錄下來的路線；沒有就回 nil（測試照樣跑，只是沒有路線可循）。
func loadRoute(t *testing.T) []game.Decision {
	t.Helper()
	path := os.Getenv("COAB_ROUTE_JSON")
	if path == "" {
		return nil
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join("..", "..", path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Logf("讀不到路線 %s：%v（這一場就沒有路線可循）", path, err)
		return nil
	}
	var route []game.Decision
	if err := json.Unmarshal(raw, &route); err != nil {
		t.Fatalf("路線解不開：%v", err)
	}
	return route
}

// passability 把四個方向通不通印成一行；`doorNote` 補上門的選單有沒有開。
func (s *keyDrivenSession) passability() string {
	if s.app.geoGrid == nil {
		return "沒有地圖"
	}
	names := []string{"北", "東", "南", "西"}
	parts := make([]string, 0, 4)
	for index, direction := range []int{0, 2, 4, 6} {
		deltaX, deltaY := headingDelta(direction)
		mark := "牆"
		if s.app.state.CanMoveDungeon(*s.app.geoGrid, deltaX, deltaY, direction) {
			mark = "通"
		}
		parts = append(parts, names[index]+mark)
	}
	return strings.Join(parts, "")
}

func (s *keyDrivenSession) doorNote() string {
	if !s.app.dungeonDoorMenu {
		return ""
	}
	flags, ok := s.app.dungeonDoorFlags()
	if !ok {
		return "／門選單開著（讀不到 flags）"
	}
	options := s.app.state.DungeonDoorMenuOptions(flags)
	return fmt.Sprintf("／門選單開著 flags=%d 敲=%v 撬=%v 撞=%v",
		flags, options.Knock, options.Pick, options.Bash)
}

// tracing 回答「這一場要不要收逐格紀錄」。
func (s *keyDrivenSession) tracing() bool { return os.Getenv("COAB_KEY_TRACE") != "" }

// keyDrivenFrames 是這一場要跑幾幀；`COAB_KEY_FRAMES` 可以覆蓋掉，用來量上限。
func keyDrivenFrames() int {
	if raw := os.Getenv("COAB_KEY_FRAMES"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			return value
		}
	}
	return keyDrivenDefaultFrames
}

// keyDrivenDefaultFrames 是量出來的：真正的進展在第 1,937 幀停下，2,500／3,000／
// 4,000／10,000／25,000 幀的結果**完全相同**（236 格、191 句、同一條段序列）。
// 2,500 幀跑一次約 10 秒。
//
// ⚠ 這個數字會隨走法與譯文缺口改變：補完酒館傳聞之前它是 1,500，那時候 4,000 幀
// 也只走到 137 格——**不是因為幀數不夠，是因為隊伍撞到一句沒翻的話就停在那裡**。
const keyDrivenDefaultFrames = 2500

func newKeyDrivenSession(t *testing.T) *keyDrivenSession {
	application, keys := keyDrivenApp(t)
	route := loadRoute(t)
	moves, choices := buildRouteIndex(route)
	return &keyDrivenSession{
		route: route, routeCursor: map[string]int{},
		routeMoves: moves, routeChoices: choices,
		app: application, keys: keys,
		cells: map[[3]int]bool{}, tried: map[[4]int]bool{}, modes: map[game.Mode]bool{},
		messages: map[string]bool{}, fallbacks: map[string]bool{},
		blocks: map[uint8]bool{}, menus: map[string]bool{}, menuSeen: map[string]int{},
		modeFrames: map[game.Mode]int{}, stallCells: map[[3]int]int{},
		menuSinceProgress: map[string]int{},
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
	s.modeFrames[state.Mode]++
	if state.Mode == game.ModeDungeon || state.Mode == game.ModeEvent {
		x, y, _ := state.DungeonGeometryView()
		block := 0
		if s.app.geoGrid != nil {
			block = int(s.app.geoBlock)
		}
		key := [3]int{block, x, y}
		if !s.cells[key] {
			s.noteProgress()
		}
		s.cells[key] = true
		s.stallCells[key]++
	}
	if block, ok := state.CurrentECLBlockID(); ok {
		if !s.blocks[block] {
			s.noteProgress()
		}
		s.blocks[block] = true
		entry := fmt.Sprintf("0x%02X@%d", block, s.frames)
		if len(s.segmentTrace) == 0 ||
			strings.SplitN(s.segmentTrace[len(s.segmentTrace)-1], "@", 2)[0] != fmt.Sprintf("0x%02X", block) {
			s.segmentTrace = append(s.segmentTrace, entry)
		}
	}
	if len(state.Choices) > 0 {
		signature := strings.Join(state.Choices, "｜")
		// ⚠ **新的選單不算「有進展」。** 商店與神殿的選單把金幣與件數寫進選項
		// （「戰士（HP 0/0，0 GP）」），數字一變簽章就變 ⇒ 每一幀都是「新選單」，
		// 於是停滯偵測永遠不會觸發，而報表會寫「最後一次有新東西在第 24672 幀」
		// 卻和第 1500 幀的格數、句數、段序列一模一樣。
		s.menus[signature] = true
		s.menuSinceProgress[stuckSignature(state.Choices)]++
	}
	// ★ **選項也要查落回原文**。 原本只查 `Message` 與 `Prompt`，於是
	// 「ASK ABOUT INJURIES｜IGNORE THEM」這種整組英文選單一路演到玩家面前而
	// 測試全綠——落回原文的判準漏了玩家最常讀的那一塊。
	for _, option := range state.Choices {
		trimmed := strings.TrimSpace(option)
		if trimmed == "" || hasHan(trimmed) || !hasLatinWord(trimmed) {
			continue
		}
		s.fallbacks[trimmed] = true
	}
	for _, text := range []string{state.Message, state.Prompt} {
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			continue
		}
		if !s.messages[trimmed] {
			s.noteProgress()
		}
		s.messages[trimmed] = true
		if !hasHan(trimmed) && hasLatinWord(trimmed) {
			s.fallbacks[trimmed] = true
		}
	}
}

// stuckSignature 是「這是不是同一個選單」用的簽章：把選項裡的**數字**抹掉。
//
// ★ 為什麼不能直接用選項原文。 有兩種選單會把會變的值寫進選項——商店與神殿寫
// 金幣與件數（「戰士（HP 0/0，0 GP）」），肖像編輯器寫目前的編號（「頭部：4D」）。
// 拿原文當簽章，**每按一次就是一個「沒看過的選單」**，於是「同一個選單看太多次
// 就換一項」這條停滯偵測完全不會觸發：實測隊伍在肖像編輯器裡按了六百次
// 「頭部上一個」，把編號從 4D 一路減到 03，而報表看起來只是「停在荒野」。
//
// ⚠ 抹掉數字**不能**用在查路線那一邊：那邊要跟錄下來的選項逐字相同。
func stuckSignature(choices []string) string {
	return stuckDigits.ReplaceAllString(strings.Join(choices, "｜"), "#")
}

var stuckDigits = regexp.MustCompile(`[0-9A-F]+`)

// noteProgress 記下「這一幀有新東西」。
//
// ★ **「有進展」不能只算新格子**：隊伍走到世界地圖之後就不再踏地城的格子，
// 於是格數凍住、看起來像卡死，實際上劇情還在往前。新的段、新的一句話、
// 新的選單都算——這一行是判斷「真的停了沒」的唯一依據。
func (s *keyDrivenSession) noteProgress() {
	s.lastProgress = s.frames
	s.stallCells = map[[3]int]int{}
	s.menuSinceProgress = map[string]int{}
}

// traceMenu 記下這一幀在哪個選單上按了第幾項，以及那是路線給的還是啟發式給的。
func (s *keyDrivenSession) traceMenu(reason string, want int) {
	if !s.tracing() {
		return
	}
	block, _ := s.app.state.CurrentECLBlockID()
	chosen := ""
	if want >= 0 && want < len(s.app.state.Choices) {
		chosen = s.app.state.Choices[want]
	}
	s.moveTrace = append(s.moveTrace, fmt.Sprintf(
		"幀%04d 選單 0x%02X %s 第%d項=%q ← %s｜%s", s.frames, block,
		modeName(s.app.state.Mode), want, chosen, reason,
		strings.Join(s.app.state.Choices, "｜")))
}

// step 依畫面挑一顆鍵按下去。**只用玩家按得到的鍵**，不碰 `state` 的任何方法。
//
// ⚠ 地城裡要**兩輪**挑方向：先找「走得通而且沒去過」的，找不到才退回「走得通」
// 就好。少了第二輪，走到只剩一條回頭路的死巷時會**原地轉圈轉到天荒地老**——
// 那條路走得通，但因為去過所以被拒絕，於是永遠不動。
func (s *keyDrivenSession) step(t *testing.T) {
	t.Helper()
	if s.app.state.Mode != game.ModeDungeon || s.app.geoGrid == nil {
		// ⚠ 選單**預設按第一項**。試過依幀數輪流選：走過的格子從 28 掉到 27、
		// 記到的話從 23 掉到 16——因為輪到「離開」那一項就真的離開了，
		// 整條路被自己切斷。冷走那邊多策略有效，是因為它**每一種策略各跑一趟**
		// 再取聯集；這裡是**一條連續的 session**，換選項不是多一條路，是換一條路。
		//
		// ★ 但「永遠第一項」會**卡在同一個選單上**：世界地圖的
		// 「進入城市｜繼續旅程｜紮營」永遠選第一項，隊伍就再也離不開開場那一座城，
		// 整場 session 停在 ECL `0x01`。
		//
		// ⇒ 折衷是**卡住才換**：同一個選單看到第 N 次還在原地，就往下挪一項。
		// 這不是亂試，是「這條路走過了，換下一條」——而且預設仍然是第一項，
		// 所以不會像輪流選那樣自己切斷路線。
		// ★ 先看**路線**：主線錄下來的下一個決策如果面對的是同一組選項，
		// 就照它按。路線走完或對不上才退回啟發式。
		if want, ok := s.routeChoice(); ok {
			if s.moveCursorTo(t, want) {
				s.routeHits++
				s.traceMenu("路線", want)
				tap(t, s.app, s.keys, ebiten.KeyEnter)
				return
			}
			s.traceMenu("路線挪不到游標", want)
		}
		if count := len(s.app.state.Choices); count > 1 {
			// ⚠ 簽章**只看選項，不看訊息**：商店每按一項就換一句回應
			// （「目前沒有可提取的隊伍金幣。」…），把訊息放進簽章的話計數永遠
			// 歸零，**卡住偵測等於沒有**——而商店的「離開商店」是最後一項，
			// 於是整場 session 就困在店裡。
			signature := stuckSignature(s.app.state.Choices)
			s.menuSeen[signature]++
			want := (s.menuSeen[signature] - 1) / keyDrivenMenuPatience
			if want >= count {
				want = count - 1
			}
			// ⚠ `a.choiceCursor` **跨模式留著**：在 58 項的商店選單挪到第 8 項之後
			// 換到只有 3 項的荒野選單，游標還是 8 ⇒ 按下去會拿到
			// `choice 8 is invalid in mode 1`。所以要**兩個方向都會動**，
			// 不能只往下挪。
			for step := 0; step < count+8 && s.app.choiceCursor != want; step++ {
				key := ebiten.KeyDown
				if s.app.choiceCursor > want {
					key = ebiten.KeyUp
				}
				tap(t, s.app, s.keys, key)
				if s.app.state.Mode == game.ModeDungeon {
					return
				}
			}
			// ⚠ 挪游標的過程中**模式可能已經變了**（某些選項一選就換畫面），
			// 所以要拿**當下**的選項數再檢查一次，不能用迴圈前抓的那個。
			if current := len(s.app.state.Choices); s.app.choiceCursor >= current {
				// 還是挪不進範圍就不要按下去——寧可這一幀什麼都不做。
				return
			}
			s.traceMenu("卡住換下一項", want)
		}
		tap(t, s.app, s.keys, ebiten.KeyEnter)
		return
	}
	// 門的選單開著就先處理掉：敲門 → 撬鎖 → 撞門，都失敗就退出。
	if s.app.dungeonDoorMenu {
		s.handleDoorMenu(t)
		return
	}
	// ★ 先看路線：主線在這一格往哪走。轉到那個方向再踏出去，全程只用按鍵。
	if direction, ok := s.routeMove(); ok {
		beforeX, beforeY, _ := s.app.state.DungeonGeometryView()
		beforeBlock, _ := s.app.state.CurrentECLBlockID()
		for turn := 0; turn < 4; turn++ {
			_, _, facing := s.app.state.DungeonGeometryView()
			if int(facing) == direction {
				break
			}
			tap(t, s.app, s.keys, ebiten.KeyM)
			if s.app.state.Mode != game.ModeDungeon {
				return
			}
		}
		s.routeHits++
		tap(t, s.app, s.keys, ebiten.KeyUp)
		if s.tracing() {
			afterX, afterY, _ := s.app.state.DungeonGeometryView()
			afterBlock, _ := s.app.state.CurrentECLBlockID()
			note := ""
			if afterX == beforeX && afterY == beforeY && s.app.state.Mode == game.ModeDungeon {
				note = " 沒動：" + s.passability() + s.doorNote()
			}
			s.moveTrace = append(s.moveTrace, fmt.Sprintf(
				"幀%04d 路線 0x%02X(%d,%d)→%d ⇒ 0x%02X(%d,%d) %s%s",
				s.frames, beforeBlock, beforeX, beforeY, direction,
				afterBlock, afterX, afterY, modeName(s.app.state.Mode), note))
		}
		if s.app.state.Mode == game.ModeDungeon {
			afterX, afterY, _ := s.app.state.DungeonGeometryView()
			if afterX == beforeX && afterY == beforeY {
				// ⚠ 路線說往這邊走、按下去卻沒動：那一步在這一場**用不出來**
				// （門沒開、旗標不同、或這一格根本不是主線走到的那個狀態）。
				// 記下來，免得把「路線用完了」與「路線帶不動」混成同一件事。
				s.routeBlocked++
				s.searchForAWayThrough(t)
			}
		}
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

// searchForAWayThrough 是「路線說有路、按下去卻是牆」時玩家會做的事：先開搜尋
// （`S`），再看一眼（`L`）。
//
// ★ 為什麼需要這一段。 原作有**搜尋才會出現的邊**——`tilverton.sewers` 的
// `(10,12)` 往西就是一面要搜尋才過得去的牆（game pack 的 `search_edges`）。
// 主線的路線紀錄從那裡穿過去，因為錄路線的那條測試是在規則層走的；按鍵這一場
// 不搜尋就永遠是一面牆，而畫面上看起來只是「走不動」。
//
// ⚠ 搜尋只開一次：`ToggleDungeonSearch` 是切換，每次都按會一開一關，
// 於是**看起來一直在搜尋，實際上有一半的時間是關的**。
func (s *keyDrivenSession) searchForAWayThrough(t *testing.T) {
	t.Helper()
	if s.app.state.Mode != game.ModeDungeon {
		return
	}
	if !s.app.state.DungeonSearchEnabled {
		tap(t, s.app, s.keys, ebiten.KeyS)
		s.searchToggles++
		return
	}
	tap(t, s.app, s.keys, ebiten.KeyL)
	s.looks++
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
	action := "退出"
	switch {
	case options.Knock:
		action = "敲門"
		tap(t, s.app, s.keys, ebiten.KeyK)
	case options.Pick:
		action = "撬鎖"
		tap(t, s.app, s.keys, ebiten.KeyP)
	case options.Bash:
		action = "撞門"
		tap(t, s.app, s.keys, ebiten.KeyB)
	default:
		tap(t, s.app, s.keys, ebiten.KeyEscape)
	}
	if s.tracing() {
		x, y, facing := s.app.state.DungeonGeometryView()
		s.moveTrace = append(s.moveTrace, fmt.Sprintf(
			"幀%04d 門 (%d,%d)朝%d flags=%d %s ⇒ %.30q 選單%v",
			s.frames, x, y, facing, flags, action, s.app.state.Message,
			s.app.dungeonDoorMenu))
	}
}

// chooseHeading 挑下一步要朝哪。先要沒去過的，沒有就退回任何走得通的。
//
// ⚠ **「去過沒有」要以（格子, 進入方向）為單位**：樓梯與傳送事件是「站對方向
// 踏上去」才觸發的，只記格子的話，第一次從錯的方向踏上去就把那一格封死，
// 另外三個方向永遠不會再試（spec 1193）。
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
		target := [4]int{int(s.app.geoBlock), x + deltaX, y + deltaY, heading}
		if !s.tried[target] {
			s.tried[target] = true
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

	// ⚠ 幀數上限是**量出來的不是猜的**：換掉走法之後要重新量一次到頂的位置，
	// 用 `COAB_KEY_FRAMES` 掃一遍再把預設值訂在轉折點上（見 spec 1197）。
	// 不要為了「看起來跑得久」而拖慢整個測試套件。
	for session.frames = 0; session.frames < keyDrivenFrames(); session.frames++ {
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
	t.Logf("按鍵驅動 %d 幀：走過 %d 格、%d 種畫面、記到 %d 句話、撞到門 %d 次、"+
		"照路線按 %d 次（路線共 %d 步，查表覆蓋 %d 格／%d 種選單），落回原文 0 句",
		session.frames, len(session.cells), len(session.modes), len(session.messages),
		session.doorsFound, session.routeHits, len(session.route),
		len(session.routeMoves), len(session.routeChoices))
	x, y, facing := application.state.DungeonGeometryView()
	spent := make([]string, 0, len(session.modeFrames))
	for mode, frames := range session.modeFrames {
		spent = append(spent, fmt.Sprintf("%s %d", modeName(mode), frames))
	}
	sort.Strings(spent)
	stall := make([]string, 0, len(session.stallCells))
	for cell, count := range session.stallCells {
		stall = append(stall, fmt.Sprintf("0x%02X(%d,%d)×%d", cell[0], cell[1], cell[2], count))
	}
	sort.Strings(stall)
	if len(stall) > 12 {
		stall = append(stall[:12], fmt.Sprintf("…共 %d 格", len(session.stallCells)))
	}
	t.Logf("  停在：%s (%d,%d) 朝 %d；最後一次有新東西在第 %d 幀；"+
		"各畫面幀數 %s；路線帶不動 %d 次", modeName(application.state.Mode), x, y, facing,
		session.lastProgress, strings.Join(spent, "、"), session.routeBlocked)
	t.Logf("  被牆擋住之後：開搜尋 %d 次、看一眼 %d 次", session.searchToggles, session.looks)
	t.Logf("  卡住之後在這些格子之間繞：%s", strings.Join(stall, " "))
	trace := session.segmentTrace
	if len(trace) > 40 {
		trace = append(append([]string{}, trace[:20]...), append([]string{"…"}, trace[len(trace)-20:]...)...)
	}
	t.Logf("  ECL 段的變化序列（段@幀）：%s", strings.Join(trace, " → "))
	// ★ 卡住的原因記在這裡：整場**沒有出現任何真的選單**（只有「按任意鍵繼續」）。
	// 所以走不遠不是「選錯了選項」，是幾何上到不了——這兩種成因的處置完全不同，
	// 分不出來就會往錯的方向修。
	for menu := range session.menus {
		t.Logf("  選單 %s", menu)
	}
	blocks := make([]string, 0, len(session.blocks))
	for block := range session.blocks {
		blocks = append(blocks, fmt.Sprintf("0x%02X", block))
	}
	sort.Strings(blocks)
	t.Logf("  走到過的 ECL 段：%s", strings.Join(blocks, " "))

	if path := os.Getenv("COAB_KEY_TRACE"); path != "" {
		if !filepath.IsAbs(path) {
			path = filepath.Join("..", "..", path)
		}
		if err := os.WriteFile(path,
			[]byte(strings.Join(session.moveTrace, "\n")+"\n"), 0o644); err != nil {
			t.Fatalf("逐格紀錄寫不出來：%v", err)
		}
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
		Segments  int      `json:"segments"`
		RouteHits int      `json:"route_hits"`
		RouteLen  int      `json:"route_steps"`
	}{
		Schema: "coab-key-driven-session/1", Frames: s.frames, Cells: len(s.cells),
		Modes: modes, Messages: len(s.messages), Fallbacks: len(s.fallbacks),
		Doors: s.doorsFound, Menus: len(s.menus),
		Segments: len(s.blocks), RouteHits: s.routeHits, RouteLen: len(s.route),
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

// TestKeyDrivenRouteProbe 探路用：從開場走到世界地圖選單，逐項試「進入城市／
// 繼續旅程／紮營」會走到哪。預設不跑。
func TestKeyDrivenRouteProbe(t *testing.T) {
	if os.Getenv("COAB_KEY_EXPLORE") == "" {
		t.Skip("探路用")
	}
	for pick := 0; pick < 3; pick++ {
		session := newKeyDrivenSession(t)
		application, keys := session.app, session.keys
		tap(t, application, keys, ebiten.KeyEnter)
		for index := 0; index < 6; index++ {
			tap(t, application, keys, ebiten.KeyEnter)
		}
		tap(t, application, keys, ebiten.KeyD)
		// 走到出現三選一的世界地圖選單為止。
		found := false
		for frame := 0; frame < 400 && !found; frame++ {
			session.frames = frame
			session.observe()
			if len(application.state.Choices) == 3 &&
				strings.Contains(strings.Join(application.state.Choices, "｜"), "繼續旅程") {
				found = true
				break
			}
			session.step(t)
		}
		if !found {
			t.Logf("pick=%d：400 幀內沒看到世界地圖選單", pick)
			continue
		}
		for step := 0; step < 3 && application.choiceCursor != pick; step++ {
			tap(t, application, keys, ebiten.KeyDown)
		}
		before, _ := application.state.CurrentECLBlockID()
		tap(t, application, keys, ebiten.KeyEnter)
		for frame := 0; frame < 60; frame++ {
			session.frames = frame
			session.observe()
			session.step(t)
		}
		after, _ := application.state.CurrentECLBlockID()
		blocks := make([]string, 0, len(session.blocks))
		for block := range session.blocks {
			blocks = append(blocks, fmt.Sprintf("0x%02X", block))
		}
		sort.Strings(blocks)
		t.Logf("pick=%d（%s）：段 0x%02X → 0x%02X；走到過 %s；mode=%s msg=%.40q",
			pick, application.state.Choices, before, after, strings.Join(blocks, " "),
			modeName(application.state.Mode), application.state.Message)
	}
}
