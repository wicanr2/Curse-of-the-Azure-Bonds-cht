package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
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
	treasureBlocks, blockErr := loadTreasureItemBlocks(imagePath)
	if blockErr != nil {
		t.Fatalf("load the production ITEM*.DAX catalog: %v", blockErr)
	}
	state.SetTreasureItemBlocks(treasureBlocks)
	// ⚠ 六章的 `MON*CHA`／`MON*SPC`／`MON*ITM` 也要載，理由同上一段。
	// 少了它，走到尤拉什神殿的 `ECL3/0x11:0D0F ADD NPC 16h`（雅麗亞絲）會回
	// 「ADD NPC 0x16 has no MON3CHA Player record」——記錄其實在原版檔裡好好的
	// （422 bytes 的玩家記錄），**是測試自己沒載**。這種錯讀起來像 remake 少了
	// 一段實作，實際上是量測工具的資料缺口。
	for chapter := uint8(1); chapter <= 6; chapter++ {
		if data, loadErr := zipMember(imagePath, fmt.Sprintf("MON%dCHA.DAX", chapter)); loadErr == nil {
			if records, parseErr := loadMonsterRecords(data); parseErr == nil {
				state.SetMonsterRecordsForECL(chapter, records)
				if chapter == 1 {
					state.SetMonsterRecords(records)
				}
			}
		}
		if data, loadErr := zipMember(imagePath, fmt.Sprintf("MON%dSPC.DAX", chapter)); loadErr == nil {
			if affects, parseErr := loadMonsterAffects(data); parseErr == nil {
				state.SetMonsterAffectsForECL(chapter, affects)
				if chapter == 1 {
					state.SetMonsterAffects(affects)
				}
			}
		}
		if data, loadErr := zipMember(imagePath, fmt.Sprintf("MON%dITM.DAX", chapter)); loadErr == nil {
			if items, parseErr := loadMonsterItems(data); parseErr == nil {
				state.SetMonsterItemsForECL(chapter, items)
			}
		}
	}
	keys := newScriptedKeys()
	// 戰鬥動畫在正式前端以 wall-clock 推進；鍵盤測試的「幀」則是緊密迴圈，
	// 不能讓主機快慢決定同一顆 Q 是否被動畫吞掉。每次前端取時固定前進十秒，
	// 只替代 renderer clock，不直呼任何遊戲動作或狀態轉移。
	combatClock := time.Unix(0, 0)
	application := &app{
		state: &state, keys: keys, geoCatalog: geoCatalog,
		ui: newUIRuntime(defaultUISettings(), filepath.Join(t.TempDir(), "ui-settings.json")),
		combatNow: func() time.Time {
			combatClock = combatClock.Add(10 * time.Second)
			return combatClock
		},
	}
	// ⚠ 三支戰鬥地形投影跟正式執行檔一樣要裝（`main.go` 在 RunGame 之前裝）。
	// 少了它們，這一場的每場戰鬥都沒有地形（佈陣不看地面、AI 不算成本），
	// 而怪物的快速面殺法術（火刀法師的臭雲）一出手就是
	// 「TACTICALMAP projection is unavailable」——看起來像輸入層壞了，
	// 其實是**測試自己少裝投影**（同上面少載資料那兩段的形狀）。
	application.state.SetCombatLineTerrain(application.combatLineTerrain())
	application.state.SetCombatMovementTerrain(application.combatMovementTerrain())
	application.state.SetCombatScanMapProvider(application.combatScanTacticalMap)
	return application, keys
}

// tap 按一顆鍵、跑一幀、再放開跑一幀。
//
// ⚠ 放開那一幀不能省：`JustPressed` 的語意是「這一幀剛按下」，不放開的話
// 下一次 `Update()` 會把同一顆鍵再吃一次。
func tap(t *testing.T, application *app, keys *scriptedKeys, key ebiten.Key) {
	t.Helper()
	keys.press(key)
	if err := application.Update(); err != nil {
		block, _ := application.state.CurrentECLBlockID()
		t.Fatalf("按 %v：mode=%s block=0x%02X area=%d message=%q: %v",
			key, modeName(application.state.Mode), block,
			application.state.Area.GameArea, application.state.Message, err)
	}
	keys.release()
	if err := application.Update(); err != nil {
		t.Fatalf("放開 %v：%v", key, err)
	}
}

func typeText(t *testing.T, application *app, keys *scriptedKeys, value string) {
	t.Helper()
	keys.chars = []rune(value)
	if err := application.Update(); err != nil {
		t.Fatalf("輸入 %q：%v", value, err)
	}
	keys.release()
	if err := application.Update(); err != nil {
		t.Fatalf("結束輸入 %q：%v", value, err)
	}
}

// tapWithModifier 走與真人相同的「按住修飾鍵並敲命令鍵、全部放開」。
// ALT+M 的 gate 讀的是 Pressed，而 M 讀 JustPressed；scriptedKeys 的 press
// 會替換整個當幀集合，所以兩顆必須在同一次 press 送入。
func tapWithModifier(t *testing.T, application *app, keys *scriptedKeys, modifier, key ebiten.Key) {
	t.Helper()
	keys.press(modifier, key)
	if err := application.Update(); err != nil {
		t.Fatalf("按 %v+%v：%v", modifier, key, err)
	}
	keys.release()
	if err := application.Update(); err != nil {
		t.Fatalf("放開 %v+%v：%v", modifier, key, err)
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
	candidates := s.routeChoices[signature]
	destinationMenu := s.app.state.Prompt == "從這裡可以前往"
	if destinationMenu && s.previousWorldOrigin != "" {
		filtered := make([]int, 0, len(candidates))
		for _, candidate := range candidates {
			choice := s.route[candidate].Index
			if choice >= 0 && choice < len(s.app.state.Choices) &&
				s.app.state.Choices[choice] != s.previousWorldOrigin {
				filtered = append(filtered, candidate)
			}
		}
		candidates = filtered
	}
	// The recorded campaign file is a union of per-segment decisions, not one
	// literal end-to-end itinerary. Its travel-method menus therefore include
	// exploratory EXIT choices. Replaying EXIT in a continuous session cancels
	// arrival and leaves ECL paused at a branch that the next JOURNEY ON cannot
	// resume as a destination menu. Keep recorded TRAIL/ROAD/WILDERNESS choices,
	// but discard only this non-progressing travel cancellation.
	if len(s.app.state.Choices) >= 2 && s.app.state.Choices[len(s.app.state.Choices)-1] == "離開" {
		hasWilderness := false
		for _, choice := range s.app.state.Choices {
			if choice == "荒野" {
				hasWilderness = true
				break
			}
		}
		if hasWilderness {
			filtered := make([]int, 0, len(candidates))
			for _, candidate := range candidates {
				if s.route[candidate].Index != len(s.app.state.Choices)-1 {
					filtered = append(filtered, candidate)
				}
			}
			candidates = filtered
		}
	}
	index, ok := s.takeRouteStep("menu:"+signature, candidates)
	if !ok {
		return 0, false
	}
	choice := s.route[index].Index
	if destinationMenu {
		s.previousWorldOrigin = s.app.state.LocationName
	}
	return choice, true
}

// routeMove 回答「主線站在這一格時往哪個方向走」。
func (s *keyDrivenSession) routeMove() (int, bool) {
	x, y, _ := s.app.state.DungeonGeometryView()
	block, _ := s.app.state.CurrentECLBlockID()
	// 一般強度路徑在下水道南界與火刀據點入口採原版已驗證的有向交接：
	// `0x03 (10,15) S` → `0x04 (8,0) S`（spec 1184）。純探索啟發式把入口
	// 視為剛走過的舊格，會立刻往北退回下水道，之後在兩段間反覆；固定的只是
	// 玩家剛選擇「往火刀據點前進」這一步，不注入座標，也不封死其他出口。
	if !keyDrivenBoost() {
		if (block == 0x03 && x == 10 && y == 15) || (block == 0x04 && x == 8 && y == 0) {
			return 4, true
		}
	}
	// ⚠ 地圖編號要取**這一步真的會用的那張 grid**（前端手上的那一張），
	// 不是 `State.GeoMapBlock`——錄的時候取的是 `grid.BlockID`，兩者會不一樣。
	// 用錯那一格會讓每一次查表都落空，而落空是**安靜地沒有路線**。
	currentGeo := 0
	if s.app.geoGrid != nil {
		currentGeo = int(s.app.geoGrid.BlockID)
	}
	cell := routeCell{int(block), s.routeGeoBlock(currentGeo), x, y}
	// 放手條件，與選單那一側同源：這一格從上一次有新東西到現在已經照路線走過
	// 這麼多次，路線在這一場就是幫不上忙，交給啟發式。
	if s.moveSinceProgress[cell] > routeMovePatienceValue() {
		return 0, false
	}
	candidates := make([]int, 0, len(s.routeMoves[cell]))
	for _, index := range s.routeMoves[cell] {
		if s.routeDead[[4]int{cell.segment, cell.x, cell.y, s.route[index].Direction}] >= routeDeadAfter {
			continue
		}
		candidates = append(candidates, index)
	}
	index, ok := s.takeRouteStep(
		fmt.Sprintf("cell:%d/%d/%d/%d", cell.segment, cell.geoBlock, cell.x, cell.y),
		candidates)
	if !ok {
		return 0, false
	}
	s.moveSinceProgress[cell]++
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

// routeDeadAfter 是「同一格同一個方向撞幾次之後不再照路線走」。
// ⚠ 太小會把還沒撞開的門也劃掉（1 次 ⇒ 245 → 112 格）；量出來再定。
const routeDeadAfter = 3

// routeMenuPatience 是「同一個選單連續出現幾次都沒有新東西，就不再問路線」。
const routeMenuPatience = 8

// routeMovePatience 是走位那一側的同一個數字：同一格連續照路線走這麼多次都
// 沒有新東西，就不再問路線。
//
// 實測（2,600 幀，同一份路線；「段」欄取最深的一段）：
//
//	耐心=1   111 格／ 66 句／最深 0x02   ← 太早放手，連提爾佛頓都走不完
//	耐心=2   342 格／225 句／最深 0x51   ← 目前這一版
//	耐心=3   335 格／215 句／最深 0x33
//	耐心=4   280 格／205 句／最深 0x33
//	耐心=8   279 格／205 句／最深 0x33
//	耐心=16  274 格／205 句／最深 0x33
//
// ⚠ 選單那一側的 8 不適用在這裡：一個選單按錯只浪費一幀，而走位按錯會讓隊伍
// **離開**原本站對的那一格，代價高得多，所以放手要早。改動走法之後重新量。
const routeMovePatience = 2

func routeMovePatienceValue() int {
	if raw := os.Getenv("COAB_KEY_MOVE_PATIENCE"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			return value
		}
	}
	return routeMovePatience
}

func routeMenuPatienceValue() int {
	if raw := os.Getenv("COAB_KEY_MENU_PATIENCE"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			return value
		}
	}
	return routeMenuPatience
}

// routeCell 是「這一段、這一張圖的這一格」。
//
// ⚠ **段號不能省**：不同段的地圖上都有 (7,13)。
// ⚠ **地圖編號也不能省**：段與地圖不是一對一——實測 16,048 個移動裡有 2,905 個
// 的段號與地圖編號不同。只記段號的話，同一個座標會混到另一張圖的走法，
// 重放時就往牆裡走。
type routeCell struct{ segment, geoBlock, x, y int }

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
// routeGeoBlock 決定查表要不要帶地圖編號。
//
// ⚠ **看路線檔有沒有這一欄，不要無條件帶**：`geo_block` 是後來才加的，早期錄的
// 路線沒有這一欄，反序列化之後全是 0。無條件帶著查，每一格都會查不到，
// 而查不到的後果是**安靜地沒有路線**——報表上看起來只是「走得比較短」。
func (s *keyDrivenSession) routeGeoBlock(current int) int {
	if !s.routeHasGeoBlock {
		return 0
	}
	return current
}

func buildRouteIndex(route []game.Decision) (map[routeCell][]int, map[string][]int) {
	moves := map[routeCell][]int{}
	choices := map[string][]int{}
	for index, step := range route {
		switch step.Kind {
		case "move":
			key := routeCell{step.Segment, step.GeoBlock, step.FromX, step.FromY}
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

// fatalPickWindow 是「按下這一項之後幾幀內開打，才算是這一項引起的」。
//
// ★ 量出來的：酒館的「揍酒保」是**下一幀**就進戰鬥模式，而「離開商店」到隨機
// 遭遇之間隔著好幾步移動。3 幀足以蓋住前者、擋掉後者。
const fatalPickWindow = 3

// menuPick 是「哪一個選單的第幾項」。簽章與 `stuckSignature` 同一套。
type menuPick struct {
	signature string
	index     int
}

// keyDrivenSession 是一次「只用按鍵」的遊玩紀錄。
type keyDrivenSession struct {
	app    *app
	keys   *scriptedKeys
	frames int
	// wonAt 是第一次由正常按鍵路徑走到正式結局的幀；-1 代表尚未通關。
	wonAt int
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
	// visits 是每一格踏上去過幾次，用來挑「去過最少次」的退路。
	visits map[[3]int]int
	// routeDead 數著「這一格往這個方向，路線說走得通但按下去沒動」幾次。
	//
	// ★ 早期錄的路線把**被擋下來的嘗試**也記了進去（`recordMove` 錄在驗證之前），
	// 那些步驟每一次輪到都會再撞一次牆——巫師塔那一段實測撞了 156 次。
	//
	// ⚠ **不能撞一次就永久劃掉**：上鎖的門在撞開之前也是「按下去沒動」，
	// 劃掉之後隊伍就再也不會往那個方向走了（實測 245 → 112 格、只剩兩段）。
	// 要**撞夠幾次**才算真的走不通。
	routeDead map[[4]int]int
	// routeHasGeoBlock 是「這一份路線檔有沒有記地圖編號」。
	routeHasGeoBlock bool
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
	// moveSinceProgress 是走位那一側的同一件事：從上一次有新東西到現在，
	// 路線在這一格接手過幾次。
	//
	// ★ **`routeDead` 擋不住這一種卡法**：它數的是「路線說走得通、按下去沒動」，
	// 而巫師塔那四格的路線每一步都**真的走得動**——(4,2)→東→(5,2)→南→(5,3)→西→
	// (4,3)→北→(4,2)，四格是一個單向環，路線照著繞永遠不會被判死。實測 15,000 幀
	// 與 2,600 幀停在同一個地方（245 格、第 1961 幀之後再無新東西），因為那不是
	// 幀數不夠，是路線自己把隊伍鎖在環裡。⇒ 走位也要有放手條件（spec 1198）。
	moveSinceProgress map[routeCell]int
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
	// combatTurns 是按了幾次「快速戰鬥」；wipes 是全滅重開幾次。
	// ★ 這兩個一起看才知道「走不遠」是因為打不贏還是因為找不到路。
	combatTurns    int
	wipes          int
	partyWasKilled bool
	wasCombat      bool
	combatNumber   int
	// lastMenuPick 是最後一次真的按下去的選單（簽章與項次）。
	// fatalPicks 是「按下去之後整隊全滅」的那些，之後不再選。
	//
	// ★ 這不是遊戲知識，是「別重複殺死自己的那一步」。少了它，重放每次進酒館
	// 都選第 0 項「揍酒保」——十個酒館客人對六個一級戰士，必輸——然後全滅重開、
	// 計數歸零、再選一次；實測 12,000 幀裡重開 32 次，整場走不出提爾佛頓。
	lastMenuPick  menuPick
	lastMenuFrame int
	hasLastMenu   bool
	// combatStartPick 是**這一場戰鬥開打之前**按的那一項。全滅時標的是它，
	// 不是「全滅之前最後按的那一項」——後者的因果太鬆：離開商店之後走幾步遇上
	// 隨機遭遇再全滅，會把「離開商店」標成致命，於是隊伍再也離不開商店
	// （實測 9,613 幀停在場所畫面）。
	combatStartPick menuPick
	inCombatFrom    bool
	fatalPicks      map[menuPick]bool
	// boosted 記著這一場的隊伍有沒有被撐起來，報表要照實印出來。
	boosted                   bool
	normalGearStage           int
	normalGearShopFound       bool
	normalReadyStage          int
	normalReadyNeedsCharacter bool
	normalReadyDone           bool
	normalSpellStage          int
	normalSpellDone           bool
	normalSkipGearShop        bool
	normalAvoidedFee          bool
	normalMoneyPooled         bool
	normalMoneyTaken          bool
	normalTempleActive        bool
	normalTempleCharacter     int
	normalTempleTreated       map[int]bool
	normalRecoveryActive      bool
	normalRecoveryDaysAdded   int
	// normalInnAttempts records cities where the visible INN action already
	// returned the still-injured party to the same service menu. Some world
	// points expose an INN label but do not route it through PROGRAM 9; retrying
	// forever is a replay-policy loop, not healing. After one attempt, leave the
	// city and use the universally visible CAMP action at the edge.
	normalInnAttempts map[uint8]int
	// previousWorldOrigin prevents a union of per-segment route decisions from
	// sending one continuous session straight back to the world point it just
	// left. It affects only the test driver, never the game's travel graph.
	previousWorldOrigin string
	// lastNarrative 保留角色選擇器前一頁的敘述。ECL 進入「請選擇角色」時會清空
	// Message；若不保留前文，通用卡住策略只看得到六個名字，無法知道攀爬事件
	// 明確要求盜賊。
	lastNarrative string
}

var normalPreparationPurchaseItems = []string{
	"板甲（400 GP）", "盾牌（15 GP）", "長劍（15 GP）",
	"板甲（400 GP）", "盾牌（15 GP）", "釘頭錘（8 GP）",
	"四尺杖（1 GP）",
	"板甲（400 GP）", "盾牌（15 GP）", "長劍（15 GP）",
	"板甲（400 GP）", "盾牌（15 GP）", "長劍（15 GP）",
	"皮甲（5 GP）", "短劍（8 GP）",
}

var normalPreparationReadyItems = []string{
	"未裝備：板甲", "未裝備：盾牌", "未裝備：長劍",
	"未裝備：板甲", "未裝備：盾牌", "未裝備：釘頭錘",
	"未裝備：四尺杖",
	"未裝備：板甲", "未裝備：盾牌", "未裝備：長劍",
	"未裝備：板甲", "未裝備：盾牌", "未裝備：長劍",
	"未裝備：皮甲", "未裝備：短劍",
}

// 法師不能穿甲持盾，盜賊也不能持盾；不能再用「每人固定三件」推買家。
// 這張索引與 ITEMS 的 class-mask 一致，讓採買與整備都走合法的玩家交易。
var normalPreparationCharacterIndices = []int{
	0, 0, 0,
	1, 1, 1,
	2,
	3, 3, 3,
	4, 4, 4,
	5, 5,
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
	// route-current 是完整 internal/game 玩家路徑聯集的唯一現行入口。歷史上
	// route-clean-716 曾被窄測試安靜覆寫成 1,684 步，重放只會走短、不會報錯。
	// 這個寬鬆下界不宣稱覆蓋完整度，只阻止同一種損毀再次冒充正式 oracle。
	if filepath.Base(path) == "route-current.json" && len(route) < 10000 {
		t.Fatalf("現行路線只有 %d 步（最低 10000）；請執行 tools/rebuild-key-route.sh 完整重生", len(route))
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

// ★★ 這張表在「全滅會結束遊戲」之後**整個要重讀**（2026-08-24）。
//
// 在那之前，隊伍全倒了遊戲照樣繼續——`finishCombat` 只印一句「戰鬥失敗。」就回
// 地圖。於是這一場的大半是**一支 HP 全 0 的隊伍**在走（報表裡那些
// 「戰士（HP 0/0）」就是它），走過的格數因此**被灌水**。現在全滅會落到全滅畫面、
// 回標題重開，同一份路線從 476 格掉到 105 格。
//
// ⇒ **105 才是誠實的數字**，476 不是。掉下來的不是能力，是先前多算的部分。
// ⚠ 下一個瓶頸因此換成「重放打不贏架」：隊伍是六個一級戰士（HP 10），
// 而重放的戰鬥處置只會按 Enter（攻擊），實測一場 session 會全滅並重開二十幾次，
// 路線在第一次全滅之後就對不上了。要往下推得先讓它活著，不是加幀數。
//
// 舊表（全滅還不會結束遊戲時量的，留著當對照）：
//
//	 2,500 幀   344 格／215 句／ 11 秒
//	 6,000 幀   457 格／255 句／ 33 秒
//	12,000 幀   572 格／264 句／ 54 秒
//	20,000 幀   602 格／273 句／ 95 秒
//	40,000 幀   612 格／274 句／174 秒
//
// keyDrivenDefaultFrames 是量出來的（第 715 輪：佈署牆檢極性修正＋怪物側
// 自動換裝之後，路線 `route-clean-716`）：
//
//	12,000 幀（帶路線）  636 格／269 句／11 段（穿過猶拉什入口 0x10，第 10,665 幀）
//	12,000 幀（無路線）  522 格／214 句（測試套件預設就是這一條）
//
// ⇒ 幀數用滿、沒有死環：全滅 0、落回原文 0、快速戰鬥 31 場。維持 12,000。
// 第 710 輪（環繞剛修完）同幀數是 314 格／8 段；差異來自戰鬥變真——佈署極性
// 修正後領袖戰是完整的 21 隻，戰鬥吃掉更多幀，但走得更遠。
// ⚠ 環繞修掉之前（夾在最後一項）帶路線只有 137 格／5 段：紮營的「修改」與
// 「改名」兩個選單互踢，「離開」永遠輪不到（spec 1201）。無路線那條同時從
// 94 格跳到 309——啟發式終於能把每一個選單的出口輪到。
// ⚠ 兩條差這麼多是因為**沒有路線就沒有路線知識**：主線的決策點要走到那裡才會
// 出現，而啟發式找不到出城的那一步。路線檔在 `workplace/`（gitignore），
// 所以測試套件預設跑的是無路線那一條——引用數字時要說是哪一條。
//
// ⚠ 這個數字會隨走法與譯文缺口改變：補完酒館傳聞之前它是 1,500，那時候 4,000 幀
// 也只走到 137 格——**不是因為幀數不夠，是因為隊伍撞到一句沒翻的話就停在那裡**。
// 修好巫師塔的單向環之前它是 2,500，那時候 15,000 幀也停在 245 格（spec 1198）。
// **換掉走法之後要重新量這張表**，不要沿用。
const keyDrivenDefaultFrames = 12000

func newKeyDrivenSession(t *testing.T) *keyDrivenSession {
	application, keys := keyDrivenApp(t)
	route := loadRoute(t)
	moves, choices := buildRouteIndex(route)
	hasGeoBlock := false
	for _, step := range route {
		if step.Kind == "move" && step.GeoBlock != 0 {
			hasGeoBlock = true
			break
		}
	}
	return &keyDrivenSession{
		wonAt:            -1,
		routeHasGeoBlock: hasGeoBlock, routeDead: map[[4]int]int{}, visits: map[[3]int]int{},
		route: route, routeCursor: map[string]int{},
		routeMoves: moves, routeChoices: choices,
		app: application, keys: keys,
		cells: map[[3]int]bool{}, tried: map[[4]int]bool{}, modes: map[game.Mode]bool{},
		messages: map[string]bool{}, fallbacks: map[string]bool{},
		blocks: map[uint8]bool{}, menus: map[string]bool{}, menuSeen: map[string]int{},
		modeFrames: map[game.Mode]int{}, stallCells: map[[3]int]int{},
		menuSinceProgress: map[string]int{}, moveSinceProgress: map[routeCell]int{},
		fatalPicks:          map[menuPick]bool{},
		normalTempleTreated: map[int]bool{},
		normalInnAttempts:   map[uint8]int{},
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
	return latinWordRun(text) >= 1
}

// latinWordRun 回傳**最長一串連續的英文單字**有幾個（中間只隔空白或標點才算連續）。
//
// ★ 為什麼需要這個而不是「有沒有漢字」。 原作有一整類句子是**英文骨架 ＋ 中間插一個
// 中文名字**：`THE WALLS PROVE TOO SLIMY FOR 戰士 TO SUCCEED.`。這種句子有漢字，
// 於是「有漢字就算翻好了」的判準直接放行——它一路演到玩家面前而測試全綠。
// 逐格走訪找到它，是因為走到了那一場，不是因為判準抓得到。
func latinWordRun(text string) int {
	// ⚠ **一個字母不算一個單字。** 遊戲內的操作提示是
	// 「↑：前進　K／M：轉向　S：搜尋　L：查看　E：紮營」——那是**翻好的中文**，
	// 但按鍵名稱本來就是英文字母。把單一字母算成單字，這一行會被判成「連續五個
	// 英文單字」而誤報。要兩個字母以上才算。
	best, run, length := 0, 0, 0
	closeWord := func() {
		if length >= 2 {
			run++
			if run > best {
				best = run
			}
		} else if length == 1 {
			// 單一字母**不打斷**整串（`a`、`I` 之類），但也不算一個單字。
			_ = length
		}
		length = 0
	}
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			length++
			continue
		}
		closeWord()
		// 漢字、數字、全形標點都不打斷這一串——插在句子中間的名字與數字正是
		// 這個判準要抓的形狀。
	}
	closeWord()
	return best
}

// latinSentenceRun 是「幾個連續英文單字才算一句沒翻的話」。
//
// ⚠ 不能訂成 1 或 2：畫面上合法的英文碎片是有的（`GP`、`HP`、原作代號、
// `1 GP` 這種單位）。實測 3 是分界——`THE WALLS PROVE TOO SLIMY FOR` 有 6 個。
const latinSentenceRun = 3

// looksUntranslated 是「這一句落回原文了嗎」。
//
// 兩種都算：**整句沒有漢字而且有英文字**（原本的判準），以及**句子裡有連續
// 三個以上的英文單字**——後者專門抓「英文骨架 ＋ 插一個中文名字」那一類，
// 它有漢字，舊判準會放行。
func looksUntranslated(text string) bool {
	if !hasHan(text) {
		return hasLatinWord(text)
	}
	return latinWordRun(text) >= latinSentenceRun
}

// observe 記下這一幀玩家看得到的東西。
func (s *keyDrivenSession) observe() {
	state := s.app.state
	s.modes[state.Mode] = true
	s.modeFrames[state.Mode]++
	if state.Mode == game.ModeCombat && !s.wasCombat {
		s.combatNumber++
		if s.tracing() {
			block, _ := state.CurrentECLBlockID()
			s.moveTrace = append(s.moveTrace, fmt.Sprintf(
				"幀%04d 戰鬥#%d 開始 ECL0x%02X", s.frames, s.combatNumber, block))
			for _, fighter := range state.CombatFighters() {
				s.moveTrace = append(s.moveTrace, fmt.Sprintf(
					"幀%04d 戰鬥#%d 開場 %s side=%d HP=%d/%d AC=%d AB=%d damage=%dd%d%+d attacks=%d quick=%t pos=(%d,%d)",
					s.frames, s.combatNumber, fighter.Name, fighter.Side,
					fighter.HitPoints, fighter.MaxHitPoints,
					fighter.ArmorClass, fighter.AttackBonus, fighter.DamageDiceCount,
					fighter.DamageDiceSides, fighter.DamageBonus, fighter.AttacksPerTurn,
					fighter.QuickFight, fighter.CombatX, fighter.CombatY))
			}
		}
	}
	s.wasCombat = state.Mode == game.ModeCombat
	// 全滅由假轉真才算一次；全滅畫面會停好幾幀，逐幀累加會把它算成幾十次。
	if killed := state.PartyKilled(); killed && !s.partyWasKilled {
		s.wipes++
		if s.tracing() {
			block, _ := state.CurrentECLBlockID()
			s.moveTrace = append(s.moveTrace, fmt.Sprintf(
				"幀%04d 全滅#%d 戰鬥#%d ECL0x%02X message=%.100q",
				s.frames, s.wipes, s.combatNumber, block, state.Message))
			for _, character := range state.PartyRoster() {
				s.moveTrace = append(s.moveTrace, fmt.Sprintf(
					"幀%04d 全滅名冊 %s HP=%d/%d health=%d bleeding=%d coins=%d/%d/%d/%d/%d slots=%v",
					s.frames, character.Name, character.HitPoints, character.MaxHitPoints,
					character.HealthStatus, character.Bleeding, character.Copper,
					character.Silver, character.Electrum, character.Gold,
					character.Platinum, character.SpellSlots))
			}
		}
		if s.inCombatFrom {
			s.fatalPicks[s.combatStartPick] = true
			s.inCombatFrom = false
		}
	} else if !killed {
		s.partyWasKilled = false
	}
	if state.PartyKilled() {
		s.partyWasKilled = true
	}
	// 進戰鬥的那一幀把「開打前按的那一項」釘住；離開戰鬥就放掉。
	if state.Mode == game.ModeCombat {
		// ⚠ 只認**當場**開打的那一項。離開商店之後走了幾步才遇上隨機遭遇，
		// 中間沒有任何選單，於是「開打前按的那一項」還是「離開商店」——
		// 標下去隊伍就再也離不開商店（實測 9,613 幀停在場所畫面）。
		// 「揍酒保」是下一幀就開打；差別在這裡。
		if !s.inCombatFrom && s.hasLastMenu && s.frames-s.lastMenuFrame <= fatalPickWindow {
			s.combatStartPick, s.inCombatFrom = s.lastMenuPick, true
		}
	} else if !state.PartyKilled() {
		s.inCombatFrom = false
	}
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
		s.visits[key]++
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
			if s.tracing() {
				s.moveTrace = append(s.moveTrace, fmt.Sprintf(
					"幀%04d 進段 0x%02X %s｜訊息=%.100q 提示=%.80q 選項=%s",
					s.frames, block, modeName(state.Mode), state.Message, state.Prompt,
					strings.Join(state.Choices, "｜")))
			}
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
		if trimmed == "" || keyDrivenFormattedValue.MatchString(trimmed) || !looksUntranslated(trimmed) {
			continue
		}
		s.fallbacks[trimmed] = true
	}
	for _, text := range []string{state.Message, state.Prompt} {
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			continue
		}
		if text == state.Message {
			s.lastNarrative = trimmed
			if strings.Contains(trimmed, "戰鬥失敗。") {
				s.normalRecoveryActive = true
				s.normalRecoveryDaysAdded = 0
			}
		}
		if !s.messages[trimmed] {
			s.noteProgress()
			if s.tracing() {
				block, _ := state.CurrentECLBlockID()
				s.moveTrace = append(s.moveTrace, fmt.Sprintf(
					"幀%04d 新文字 0x%02X %s｜%.160q", s.frames, block, modeName(state.Mode), trimmed))
				if trimmed == "戰鬥勝利！" || trimmed == "戰鬥失敗。" {
					for _, character := range state.PartyRoster() {
						s.moveTrace = append(s.moveTrace, fmt.Sprintf(
							"幀%04d 戰後名冊 %s HP=%d/%d health=%d bleeding=%d",
							s.frames, character.Name, character.HitPoints, character.MaxHitPoints,
							character.HealthStatus, character.Bleeding))
					}
				}
			}
		}
		s.messages[trimmed] = true
		if looksUntranslated(trimmed) && !keyDrivenKnownFragments[trimmed] {
			s.fallbacks[trimmed] = true
		}
	}
}

// keyDrivenKnownFragments 是 ECL 共用問句子程式被走訪器單獨吐出時的量測例外。
// 原版正常執行會把它們併進呼叫端的同一份文字 run；替短句另寫 all_contains
// 規則反而會因 first-match-wins 攔截完整頁面。這份清單只作用於 Message／Prompt，
// 選項仍逐項接受落回原文檢查。來源契約見 cmd/ecl-text-coverage 與 spec 395／397。
var keyDrivenKnownFragments = map[string]bool{
	"WHAT DO YOU DO ?": true,
	"WHAT DO YOU DO":   true,
	"DO YOU CONTINUE?": true,
}

// 純金額是 locale 自己產生的格式化數值，不是英文原文。限定整串只能是數字與
// `GP`，避免把含 GP 的英文句子一併放過。
var keyDrivenFormattedValue = regexp.MustCompile(`^[0-9]+ GP$`)

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
	s.moveSinceProgress = map[routeCell]int{}
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
		"幀%04d 選單 0x%02X %s 地點=%d/%q 原文=%q 城市=%d 第%d項=%q ← %s｜%s｜訊息=%.60q 提示=%.40q",
		s.frames, block, modeName(s.app.state.Mode),
		s.app.state.Location, s.app.state.LocationName, s.app.state.OriginalLocation,
		s.app.state.Area.CurrentCity, want, chosen, reason,
		strings.Join(s.app.state.Choices, "｜"), s.app.state.Message, s.app.state.Prompt))
}

// step 依畫面挑一顆鍵按下去。**只用玩家按得到的鍵**，不碰 `state` 的任何方法。
//
// ⚠ 地城裡要**兩輪**挑方向：先找「走得通而且沒去過」的，找不到才退回「走得通」
// 就好。少了第二輪，走到只剩一條回頭路的死巷時會**原地轉圈轉到天荒地老**——
// 那條路走得通，但因為去過所以被拒絕，於是永遠不動。
func (s *keyDrivenSession) step(t *testing.T) {
	t.Helper()
	s.reassertBoost(t)
	// ★ 角色建立在**迴圈裡**也會出現：隊伍全滅之後回標題，按「開始」走的是與
	// 第一次開局同一條路。開頭那段固定的建角序列只跑一次，所以這裡要有一份
	// 能重來的版本——少了它，全滅之後每按一次 Enter 就多加一名角色，
	// 加到第七個回「party already has six characters」，整場 session 就死在那裡。
	if s.app.state.Mode == game.ModeCharacterCreation {
		if len(s.app.state.CreationRoster) >= 6 {
			tap(t, s.app, s.keys, ebiten.KeyD)
			// ⚠ 重開的隊伍也要撐。撐隊伍原本只做在開頭那段固定的建角序列裡，
			// 於是**第一次全滅之後就沒有了**——實測撐過的一場仍然重開 26 次。
			if s.app.state.Mode != game.ModeCharacterCreation {
				s.boostParty(t)
			}
			return
		}
		if !keyDrivenBoost() && s.app.state.CreationCursor < len(s.app.state.CreationRoster) {
			tap(t, s.app, s.keys, ebiten.KeyDown)
			return
		}
		tap(t, s.app, s.keys, ebiten.KeyEnter)
		return
	}
	// ★ 戰鬥交給原作自己的「快速戰鬥」（`Q`）。
	//
	// ⚠ 先前這裡**沒有**戰鬥處置：`step()` 落到選單那一支，看到 `len(Choices) > 1`
	// 就照選單按——而那份 `Choices` 是上一個畫面留下來的（實測整場戰鬥都在
	// 「揍酒保｜喝一杯｜離開」上打轉）。`Update()` 在戰鬥模式把 Enter 導到
	// `CombatAct`，所以隊伍**只會站著平砍**：不移動、不施法、不撤退，
	// 一場 session 全滅重開二十幾次。
	//
	// ⚠ 不要自己寫戰術。`Q` 是原作就有的選項（QUICK），玩家按得出來，
	// 而且它把行動交給引擎自己的 AI ⇒ 這一場量到的是「remake 的戰鬥規則帶得動
	// 這支隊伍嗎」，不是「測試的戰術好不好」。
	if s.app.state.Mode == game.ModeCombat {
		s.playCombatTurn(t)
		return
	}
	if !keyDrivenBoost() && s.normalSafeDungeonRecovery(t) {
		return
	}
	if !keyDrivenBoost() && s.normalTempleRecovery(t) {
		return
	}
	// 地城的正式紮營鍵是 E（C 是角色建立）。指定裝備買完後，先沿原始 GEO
	// 從武器店出口走回已驗證為 `0/0` 的安全紮營格 `(7,13)`，再走
	// TryEncamp／營地查看／整備；不能在出口直接按 E，那格是每小時 100% 遭遇，
	// 休息必然被皇家衛兵中斷（spec 281／ecl_integration_test）。
	if !keyDrivenBoost() && s.normalGearStage >= len(normalPreparationPurchaseItems) && (!s.normalReadyDone || !s.normalSpellDone) &&
		s.app.state.Mode == game.ModeDungeon {
		x, y, _ := s.app.state.DungeonGeometryView()
		if x == 7 && y == 13 {
			tap(t, s.app, s.keys, ebiten.KeyE)
			return
		}
		// 兩個可見出口都要接：從 `(3,12)` 是 E,E,E,S,E；由西側入口進店則
		// continuation 回 `(5,11)`，路線是 S,E,S,E。兩條都由 GEO2/0x01
		// 的實際移動遮罩驗證；轉身與前進仍各用正常鍵盤幀（spec 1233）。
		heading := 2
		if (x == 5 && y == 11) || (x == 6 && y == 12) {
			heading = 4
		}
		if !s.faceHeading(t, heading) {
			return
		}
		tap(t, s.app, s.keys, ebiten.KeyUp)
		return
	}
	// 商店交易結果是 ModeEvent，但仍保留上一層商品 Choices；那份選單不可操作。
	// 每一筆指定購買後都先按 Enter 回商店，尤其第二件不能讓商品殘影落入通用
	// 啟發式；舊策略正是從這裡開始誤買武器、最後把盾牌賣掉。
	if !keyDrivenBoost() && s.app.state.Mode == game.ModeEvent &&
		(s.app.state.OriginalEvent == "BUY" || s.app.state.OriginalEvent == "POOL") &&
		(s.normalGearStage > 0 || s.normalMoneyPooled) {
		tap(t, s.app, s.keys, ebiten.KeyEnter)
		return
	}
	// REST 的完成事件是正式交易結果；只有它明確回報至少一名角色完成記憶，
	// 才承認法術準備完成。不能在「開始休息」時先設旗標，也不能只看前端當幀
	// 尚未刷新的 roster 投影，否則前者會把中斷算成功、後者會無限重睡。
	if !keyDrivenBoost() && s.app.state.Mode == game.ModeEvent &&
		strings.Contains(s.app.state.Message, "名角色的法術記憶") {
		s.normalSpellDone = s.normalPreparedSpellLoadout()
		if !s.normalSpellDone {
			s.normalSpellStage = 0
		}
		tap(t, s.app, s.keys, ebiten.KeyEnter)
		return
	}
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
		if want, ok := s.normalPreparationChoice(); ok {
			if s.moveCursorTo(t, want) {
				s.traceMenu("一般強度準備", want)
				s.lastMenuPick = menuPick{stuckSignature(s.app.state.Choices), want}
				s.lastMenuFrame, s.hasLastMenu = s.frames, true
				tap(t, s.app, s.keys, ebiten.KeyEnter)
				return
			}
		}
		if want, ok := s.routeChoice(); ok {
			if s.moveCursorTo(t, want) {
				s.routeHits++
				s.traceMenu("路線", want)
				s.lastMenuPick = menuPick{stuckSignature(s.app.state.Choices), want}
				s.lastMenuFrame, s.hasLastMenu = s.frames, true
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
			// ★ 環繞而不是夾在最後一項。夾住的死法（第 753 輪實測）：紮營的
			// 「修改」選單escalate 到最後一項「改名」，改名子選單又escalate 到
			// 「返回修改選單」，兩個選單互踢，「離開」永遠輪不到——12,000 幀
			// 有 11,000 幀在這個環裡。環繞讓每一項（含出口）都週期性被試到。
			want := ((s.menuSeen[signature] - 1) / keyDrivenMenuPatience) % count
			// 世界旅行方式的 EXIT 是「取消這一次旅行」，不是通往內容的
			// 第三條路。只有路線查表已經沒有候選、落到通用探索器時才套
			// 這個護欄；錄製主線明確選 EXIT 時仍由上面的 routeChoice 處理。
			// 否則探索器會週期性取消抵達、再選同一目的地，VM continuation
			// 與畫面就在兩個合法選單間往返，幀數再多也沒有新內容。
			if want == count-1 && s.app.state.Choices[count-1] == "離開" {
				for index, choice := range s.app.state.Choices {
					if choice == "荒野" {
						want = index
						break
					}
				}
			}
			// 上一次按這一項之後整隊全滅過就換下一項。全部都試死過就照原樣按
			// ——那代表這個選單怎麼選都會死，硬停在這裡也沒有比較好。
			for probe := 0; probe < count && s.fatalPicks[menuPick{signature, want}]; probe++ {
				want = (want + 1) % count
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
			s.lastMenuPick, s.lastMenuFrame, s.hasLastMenu = menuPick{signature, want}, s.frames, true
			if s.app.state.Prompt == "從這裡可以前往" {
				s.previousWorldOrigin = s.app.state.LocationName
			}
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
				// 撞夠 `routeDeadAfter` 次才不再輪到它——門要先撞開，撞開之前
				// 也是「按下去沒動」。
				s.routeDead[[4]int{int(beforeBlock), beforeX, beforeY, direction}]++
				s.searchForAWayThrough(t)
			}
		}
		return
	}
	target, fresh, ok := s.chooseHeading()
	freshTarget := fresh
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
	beforeX, beforeY, _ := s.app.state.DungeonGeometryView()
	beforeBlock, _ := s.app.state.CurrentECLBlockID()
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
	// ⚠ **踏出去這一步本身就要記一次**，不能只靠 `observe()` 在地城／事件那兩種
	// 畫面下累加：目的地有事件、而那個事件把隊伍**推回原格並切成荒野畫面**時，
	// 那一格從頭到尾不會被 `observe()` 看到 ⇒ `visits` 永遠是 0 ⇒ 它永遠是
	// 「去得最少的」那一格 ⇒ 隊伍在兩格之間乒乓到跑完
	// （實測 `ECL2/0x01` 的 `(2,15)`／`(1,15)` 各 5,882 次）。
	deltaX, deltaY := headingDelta(target)
	s.visits[[3]int{
		int(s.app.geoBlock),
		geo.WrapCoordinate(beforeX+deltaX, geo.Width),
		geo.WrapCoordinate(beforeY+deltaY, geo.Height),
	}]++
	tap(t, s.app, s.keys, ebiten.KeyUp)
	if s.tracing() {
		afterX, afterY, _ := s.app.state.DungeonGeometryView()
		fresh := "退路"
		if freshTarget {
			fresh = "新格"
		}
		s.moveTrace = append(s.moveTrace, fmt.Sprintf(
			"幀%04d 啟發式 ECL0x%02X/GEO0x%02X(%d,%d)→%d（%s）⇒ (%d,%d) %s",
			s.frames, beforeBlock, s.app.geoBlock, beforeX, beforeY, target, fresh,
			afterX, afterY, modeName(s.app.state.Mode)))
	}
}

// normalTempleRecovery follows the original temple controls: the main menu's
// G/O keys change the current party member and HEAL applies to that member.
// Each injured, non-dead member receives at most one Cure Critical Wounds per
// visit. All money and HP changes still pass through the player-visible menus.
func (s *keyDrivenSession) normalTempleRecovery(t *testing.T) bool {
	state := s.app.state
	if state.Mode == game.ModeDungeon {
		s.normalTempleActive = false
		return false
	}
	if character, index, ok := state.TempleCurrentCharacter(); ok {
		if !s.normalTempleActive {
			s.normalTempleActive = true
			s.normalTempleTreated = map[int]bool{}
		}
		s.normalTempleCharacter = index
		if character.HealthStatus == party.HealthStatusOK && character.HitPoints >= character.MaxHitPoints {
			s.normalTempleTreated[index] = true
		}
		if !s.normalTempleTreated[index] && character.HealthStatus != party.HealthStatusDead {
			if s.moveCursorTo(t, 0) { // HEAL
				tap(t, s.app, s.keys, ebiten.KeyEnter)
				return true
			}
		}
		if index+1 < len(state.PartyRoster()) {
			tap(t, s.app, s.keys, ebiten.KeyO)
			return true
		}
		if s.moveCursorTo(t, len(state.Choices)-1) { // EXIT
			tap(t, s.app, s.keys, ebiten.KeyEnter)
			return true
		}
	}
	if !s.normalTempleActive || state.Mode != game.ModePlace {
		return false
	}
	if strings.HasSuffix(state.Prompt, "需要什麼幫助？") && len(state.Choices) == 11 {
		// 神殿治療的目的只是在敗戰後把昏迷／瀕死者恢復成正 HP，剩餘缺血再由
		// 客棧按每天 1 HP 的自然療傷補滿。固定買 600 GP 的治療致命傷會讓只有
		// 約 320 GP 等值的角色失敗，城市選單又因異常狀態重新進神殿而成死環；
		// 100 GP 的治療輕傷是同一正式服務中的可負擔選項（spec 1234）。
		if s.moveCursorTo(t, 2) { // Cure Light Wounds
			tap(t, s.app, s.keys, ebiten.KeyEnter)
			return true
		}
	}
	if strings.Contains(state.Prompt, "確定施術？") && len(state.Choices) == 2 {
		if s.moveCursorTo(t, 0) {
			s.normalTempleTreated[s.normalTempleCharacter] = true
			tap(t, s.app, s.keys, ebiten.KeyEnter)
			return true
		}
	}
	return false
}

// normalSafeDungeonRecovery uses an explicit player-visible safe-rest message.
// It enters the same CAMP/REST flow as an inn; no HP or clock value is injected.
func (s *keyDrivenSession) normalSafeDungeonRecovery(t *testing.T) bool {
	recoveryRequested := strings.Contains(s.lastNarrative, "可以在這裡安全休息") ||
		s.normalRecoveryActive
	if !recoveryRequested || s.app.state.Mode != game.ModeDungeon {
		return false
	}
	for _, character := range s.app.state.PartyRoster() {
		if character.HealthStatus != party.HealthStatusOK || character.HitPoints < character.MaxHitPoints {
			s.normalRecoveryActive = true
			s.normalRecoveryDaysAdded = 0
			s.normalSpellDone = false
			s.normalSpellStage = 0
			tap(t, s.app, s.keys, ebiten.KeyE)
			return true
		}
	}
	return false
}

// normalPreparationChoice 用玩家實際按得到的服務選單，讓六名角色各完成一次
// 合法訓練，再於確認過的裝備店購買各職業可用的防具、盾牌與武器，最後逐人紮營
// 整備，再讓牧師與法師透過 MAGIC → MEMORIZE → REST 各準備一支一級法術。
// 這不是 boost：錢、劇情扣款、價格、XP、升級擲骰、裝備限制、AC 投影、法術槽
// 與休息時間全走正式規則；這裡只取代「錄製主線時沒有理財、換裝與準備法術」
// 的選單決策。
func (s *keyDrivenSession) normalPreparationChoice() (int, bool) {
	if keyDrivenBoost() || len(s.app.state.Choices) == 0 {
		return 0, false
	}
	choices := s.app.state.Choices
	prompt := s.app.state.Prompt
	// POSTCOM deliberately lets an unconscious／dying party survive (spec
	// 1204). When the visible result is defeat, start the same player-visible
	// CAMP → REST recovery used after ordinary attrition before the generic
	// ready-party policy chooses JOURNEY ON. Without this ordering, a roster at
	// 0 HP can travel through several cities until the next ECL DAMAGE command
	// turns a recoverable defeat into a real party wipe.
	if strings.Contains(s.app.state.Message, "戰鬥失敗。") {
		for _, character := range s.app.state.PartyRoster() {
			if character.HitPoints < character.MaxHitPoints {
				s.normalRecoveryActive = true
				s.normalRecoveryDaysAdded = 0
				break
			}
		}
	}
	// 一般整備完成後，城市服務已不是目前路線的目標。通用「卡住才換」會把
	// 客棧／商店／訓練場逐間輪巡，再從城外選紮營回城，形成完全可操作但永遠
	// 不繼續旅行的循環。依玩家看得到的出口逐層離開；不指定目的地、不注入座標。
	if s.normalReadyDone && s.normalSpellDone && !s.normalRecoveryActive &&
		s.app.state.Mode == game.ModeWilderness {
		// 城市服務數量依城市不同：阿沙本福德有六項，匕首瀑布只有四項。
		// 用玩家可見的首／末項辨識，不把任一城市的形狀當成全域契約。
		if len(choices) >= 4 && choices[0] == "客棧" && choices[len(choices)-1] == "離開" {
			return len(choices) - 1, true
		}
	}
	// 城市客棧是手冊明定的安全休息服務，而且走正式 INN 選單會同步 roster
	// 與戰鬥投影。長途連戰之間若已受傷，先住店再離城；不能明知殘血仍照
	// 錄製路線去酒館閒逛，最後把「測試不會理財」誤報成平衡缺陷。
	if len(choices) >= 1 && choices[0] == "客棧" {
		for _, character := range s.app.state.PartyRoster() {
			if character.HealthStatus != party.HealthStatusOK && len(choices) > 3 && choices[3] == "神殿" {
				return 3, true
			}
		}
		for _, character := range s.app.state.PartyRoster() {
			if character.HitPoints < character.MaxHitPoints {
				s.normalRecoveryActive = true
				s.normalRecoveryDaysAdded = 0
				if !s.normalPreparedSpellLoadout() {
					s.normalSpellDone = false
					s.normalSpellStage = 0
				}
				city := s.app.state.Area.CurrentCity
				if s.normalInnAttempts[city] > 0 {
					// The previous visible INN transaction returned without
					// healing or opening CAMP. Leave and recover at the edge.
					return len(choices) - 1, true
				}
				s.normalInnAttempts[city]++
				return 0, true
			}
		}
	}
	// ECL 城市客棧以 PROGRAM 9 開啟正式紮營服務；它不會像簡化的場所
	// INN adapter 一樣瞬間補滿。依自然療傷契約，每 24 小時恢復 1 HP，
	// 所以用選單加入「全隊最大缺血量」天數，再開始休息。整段仍會經過
	// AdvanceGameTimeHours、隨機中斷與 roster 投影，沒有直接改角色數值。
	if s.normalRecoveryActive {
		if len(choices) == 3 && choices[0] == "進入城市" &&
			choices[1] == "繼續旅程" && choices[2] == "紮營" {
			return 2, true
		}
		if prompt == "紮營選單" {
			if !s.normalSpellDone && s.normalSpellStage < 2 {
				for index, choice := range choices {
					if choice == "法術" {
						return index, true
					}
				}
			}
			maxMissing := 0
			for _, character := range s.app.state.PartyRoster() {
				if missing := character.MaxHitPoints - character.HitPoints; missing > maxMissing {
					maxMissing = missing
				}
			}
			if maxMissing <= 0 {
				s.normalRecoveryActive = false
				s.normalRecoveryDaysAdded = 0
				return len(choices) - 1, true
			}
			for index, choice := range choices {
				if choice == "休息" {
					return index, true
				}
			}
		}
		if len(choices) == 4 && strings.HasPrefix(choices[0], "開始休息（") {
			maxMissing := 0
			for _, character := range s.app.state.PartyRoster() {
				if missing := character.MaxHitPoints - character.HitPoints; missing > maxMissing {
					maxMissing = missing
				}
			}
			// REST 設定會跨客棧保留，不能假設每次進來都從 24 小時開始。
			// 先前只數「這一輪按了幾次增加」，使舊的 936 小時再加 42 天，
			// 最後一次睡到 1,944 小時。直接讀玩家看得到的目前時數，往
			// `最大缺血 × 24 小時` 雙向校正，才不會累積漂移（spec 1234）。
			currentHours := 0
			if _, err := fmt.Sscanf(choices[0], "開始休息（%d 小時）", &currentHours); err == nil {
				targetHours := maxMissing * 24
				if targetHours < 24 {
					targetHours = 24
				}
				switch {
				case currentHours < targetHours:
					return 1, true // 增加 24 小時
				case currentHours > targetHours:
					return 2, true // 減少 24 小時
				default:
					return 0, true // 開始休息
				}
			}
		}
	}
	// 這是下水道天花板的可選支線；實跑選「是」並派盜賊後會進入
	// 「冒險告一段落」標題頁，而不是世界地圖。正常通關重放在玩家看得到
	// 完整警告時選「否」，避免把已打贏的戰役誤算成全滅重開。
	if len(choices) == 2 && choices[0] == "是" && choices[1] == "否" &&
		strings.Contains(s.app.state.Message, "只有盜賊爬得上") {
		return 1, true
	}
	// 主線錄製檔在兩個火刀關卡都選了投降：第一次被送回盜賊公會，第二次則
	// 讓首領勝利事件永遠不會發生，隊伍只會在 0x03／0x04 間反覆走。正常通關
	// 路徑拒絕投降，讓正式戰鬥與戰後 `NEWECL 0x50` 有機會執行。
	if len(choices) == 2 && choices[0] == "是" && choices[1] == "否" &&
		strings.Contains(s.app.state.Message, "火刀要求你們立刻投降") {
		return 1, true
	}
	// 下水道天花板事件的上一頁已明說「只有盜賊爬得上」，下一頁卻只留下
	// 共用的「請選擇角色」提示。照玩家可見前文選盜賊；不能讓通用的卡住輪替
	// 從第一名戰士開始亂試，因為錯選會直接切斷這條正常主線。
	if prompt == "請選擇角色" && strings.Contains(s.lastNarrative, "只有盜賊爬得上") {
		for index, choice := range choices {
			if choice == "盜賊" {
				return index, true
			}
		}
	}
	// 裝備買齊後若先回到荒野入口，應在離城前選 CAMP；只等地城的 E 鍵會讓
	// 第一場皇家守衛戰先發生，牧師與法師根本還沒有準備法術。
	if s.normalGearStage >= len(normalPreparationPurchaseItems) && !s.normalReadyDone && len(choices) == 3 &&
		choices[0] == "進入城市" && choices[1] == "繼續旅程" && choices[2] == "紮營" {
		return 2, true
	}
	// 酒館鬥毆是可選支線；一般強度驗證不應在完成任何整備前，因為
	// 「永遠選第一項」而主動攻擊十名酒客，製造一場與主線無關的全滅。
	if len(choices) == 3 && choices[0] == "揍酒保" && choices[1] == "喝一杯" && choices[2] == "離開" {
		return 2, true
	}
	if len(choices) == 3 && choices[0] == "攻擊" && choices[1] == "保持冷靜" && choices[2] == "撤退" {
		return 1, true
	}
	// 火刀據點明示前方是旋轉刀刃；選「闖入刀刃」會讓滿血且已整備的六人
	// 當場全滅，並非戰鬥平衡或 QUICK 失敗。正常玩家路徑採可見的「等待」，
	// 保留原事件傷害但不故意選致命選項（spec 1233）。
	if len(choices) == 3 && choices[0] == "闖入刀刃" && choices[1] == "等待" && choices[2] == "撤退" {
		return 1, true
	}
	// 法師塔地上明示〈避開塔中陷阱〉；正常玩家應讀取而不是讓錄製路線
	// 拒讀提示後繼續踩樓梯傷害。只選玩家看得到的「是」，不直接設旗標。
	if len(choices) == 2 && choices[0] == "是" && choices[1] == "否" &&
		strings.Contains(s.app.state.Message, "避開塔中陷阱") {
		// 紙條本身就是爆裂符文陷阱；閱讀後會提示「不要閱讀爆裂符文」並
		// 立刻爆炸。正常強度路線依玩家可見標題避開它，不把全滅當成路線進展。
		return 1, true
	}
	// 艾森布拉往哈普有「小徑／荒野／離開」三條玩家可見選擇。錄製路線選
	// 小徑會固定撞上三條黑龍；一般隊伍在已知這條路的致命遭遇後合法改走
	// 荒野，不需要降低黑龍數值或注入角色資源。
	if len(choices) == 3 && choices[0] == "小徑" && choices[1] == "荒野" && choices[2] == "離開" &&
		strings.Contains(s.app.state.Message, "前往哈普") {
		return 1, true
	}
	if !s.normalAvoidedFee && len(choices) == 3 && choices[0] == "如實相告" &&
		choices[1] == "說謊" && choices[2] == "離開" {
		s.normalAvoidedFee = true
		return 1, true
	}
	if prompt == "訓練哪一位角色？" {
		if s.app.state.Message == "訓練費用是 1000 GP。" {
			// 這句是餘額不足後回到名單的拒絕結果；繼續挑同一人只會
			// 在事件／訓練場兩個畫面間無限往返。
			s.normalMoneyTaken = true
			return len(choices) - 1, true
		}
		roster := s.app.state.PartyRoster()
		for index := 0; index < len(roster); index++ {
			// 已昏迷或垂死的角色依法不能受訓；若仍反覆選他，前端只會
			// 回覆拒絕訊息，整場重放便永遠停在訓練場。
			// 正式扣款會把五種硬幣都折成 copper 再換算；買下昂貴防具後的
			// 找零不一定仍留在 GP／PP。只數這兩欄會把仍有 1,070 GP 價值的
			// 戰士誤判成付不起訓練費。
			fundsCopper := uint64(roster[index].Copper) + uint64(roster[index].Silver)*10 +
				uint64(roster[index].Electrum)*100 + uint64(roster[index].Gold)*200 +
				uint64(roster[index].Platinum)*1000
			if s.tracing() {
				s.moveTrace = append(s.moveTrace, fmt.Sprintf(
					"幀%04d 訓練候選 %s level=%d health=%d coins=%d/%d/%d/%d/%d worth=%d",
					s.frames, roster[index].Name, roster[index].Level, roster[index].HealthStatus,
					roster[index].Copper, roster[index].Silver, roster[index].Electrum,
					roster[index].Gold, roster[index].Platinum, fundsCopper/200))
			}
			if roster[index].HealthStatus == party.HealthStatusOK && roster[index].Level < 5 && fundsCopper/200 >= 1000 {
				return index, true
			}
		}
		// 每人原始 300 PP 只能支付一次 1,000 GP（200 PP）訓練；六人
		// 都升到 2 級後離開，剩餘整備不再嘗試從商店取款。
		s.normalMoneyTaken = true
		return len(choices) - 1, true
	}
	if strings.Contains(prompt, "要支付 1000 GP 訓練嗎？") {
		return 0, true
	}
	if s.normalGearStage >= len(normalPreparationPurchaseItems) {
		switch {
		case prompt == "紮營選單":
			if s.normalReadyStage >= len(normalPreparationReadyItems) {
				s.normalReadyDone = true
			}
			s.normalSpellDone = s.normalPreparedSpellLoadout()
			want := "查看"
			if s.normalReadyStage > 0 && !s.normalSpellDone {
				want = "法術"
			} else if s.normalSpellDone {
				want = "離開"
			}
			for index, choice := range choices {
				if choice == want {
					return index, true
				}
			}
		case prompt == "選擇要查看的角色":
			if s.normalReadyStage >= len(normalPreparationReadyItems) {
				s.normalReadyDone = true
				return len(choices) - 1, true
			}
			s.normalReadyNeedsCharacter = false
			return normalPreparationCharacterIndices[s.normalReadyStage], true
		case strings.HasPrefix(prompt, "選擇 ") && strings.HasSuffix(prompt, " 要整備或卸下的物品"):
			if s.normalReadyNeedsCharacter {
				return len(choices) - 1, true
			}
			want := normalPreparationReadyItems[s.normalReadyStage]
			for index, choice := range choices {
				if choice == want {
					s.normalReadyStage++
					if s.normalReadyStage >= len(normalPreparationReadyItems) ||
						normalPreparationCharacterIndices[s.normalReadyStage] != normalPreparationCharacterIndices[s.normalReadyStage-1] {
						s.normalReadyNeedsCharacter = true
					}
					return index, true
				}
			}
			// 指定物品不存在時只略過該件，避免誤整備別的背包內容。
			s.normalReadyStage++
			if s.normalReadyStage >= len(normalPreparationReadyItems) ||
				normalPreparationCharacterIndices[s.normalReadyStage] != normalPreparationCharacterIndices[s.normalReadyStage-1] {
				s.normalReadyNeedsCharacter = true
			}
			return len(choices) - 1, true
		case prompt == "法術選單":
			if s.normalSpellDone {
				return len(choices) - 1, true // 返回紮營選單
			}
			if s.normalSpellStage < 2 {
				return 1, true // MEMORIZE
			}
			return 4, true // REST
		case prompt == "選擇要準備法術的角色":
			switch s.normalSpellStage {
			case 0:
				return 1, true // 牧師（建角名單是戰士、牧師、魔法師…）
			case 1:
				return 2, true // 法師
			default:
				return len(choices) - 1, true
			}
		case strings.HasSuffix(prompt, " 的可用法術"):
			// 五級牧師的完整容量是 5/5/1（本模板智慧含額外格；spec 1211／1241）。
			// 一環保留既有五支控制組，二環準備五格人類定身術，三環準備一格
			// 祈禱術。三者都走正式 MEMORIZE → REST 與原版逐環容量，不修改敵人、
			// 角色數值或法術規則；這一場量的是合法高環資源能否改變正常戰局。
			// 法師固定法術書仍屬 spec 1217 的待決產品分支，本測試不替它選。
			targets := []struct{ index, count int }{
				{0, 1}, {1, 1}, {2, 1}, {3, 1}, {5, 1}, {9, 5}, {20, 1},
			}
			if s.normalSpellStage == 1 {
				// 法師四格都準備目前唯一完成戰鬥規則的燃燒之手；原版
				// 同法術可重複佔槽，清單顯示為「燃燒之手 (4)」。
				targets = []struct{ index, count int }{{0, 4}}
			}
			for _, target := range targets {
				if target.index < len(choices) && memorizeChoiceCount(choices[target.index]) < target.count {
					return target.index, true
				}
			}
			s.normalSpellStage++
			return len(choices) - 2, true
		case len(choices) == 4 && strings.HasPrefix(choices[0], "開始休息（"):
			// 只有回到紮營選單、並從 roster 看到實際 SpellSlots 後，才把
			// normalSpellDone 設成 true；按下「開始休息」本身不是完成證據。
			return 0, true // 預設 24 小時後開始休息
		}
	}
	if s.app.state.Mode != game.ModePlace {
		return 0, false
	}
	if len(choices) == 9 && choices[0] == "購買" {
		switch {
		case s.normalSkipGearShop:
			s.normalSkipGearShop = false
			return len(choices) - 1, true
		case s.normalGearShopFound && !s.normalMoneyPooled:
			// 每名角色原有的 300 PP 足以支付自己的皮甲／盾牌與一次訓練。
			// 先前在這裡集中全部金幣，買裝雖成功，之後六人個別餘額卻全是 0，
			// 訓練場依法全部拒絕。此路徑不需要 pool，保留個人訓練費。
			s.normalMoneyPooled = true
			return 0, true
		case s.normalGearStage < len(normalPreparationPurchaseItems):
			return 0, true
		case s.normalGearStage >= len(normalPreparationPurchaseItems):
			return len(choices) - 1, true // 指定裝備買完就離店，禁止啟發式繞到販售
		}
	}
	if s.normalGearStage < len(normalPreparationPurchaseItems) && prompt == "選擇要購買的物品" {
		want := normalPreparationPurchaseItems[s.normalGearStage]
		for index, choice := range choices {
			if choice == want {
				if !s.normalMoneyPooled {
					s.normalGearShopFound = true
					return len(choices) - 1, true
				}
				s.normalGearStage++
				return index, true
			}
		}
		// 這不是盔甲店。退出商品清單，讓同一間商店仍可執行集中／
		// 取出金幣；不可落回啟發式反覆購入清單上的第一件雜物。
		s.normalSkipGearShop = true
		return len(choices) - 1, true
	}
	if s.normalGearStage < len(normalPreparationPurchaseItems) && prompt == "選擇要購買物品的角色" {
		return normalPreparationCharacterIndices[s.normalGearStage], true
	}
	return 0, false
}

func memorizeChoiceCount(choice string) int {
	if !strings.HasPrefix(choice, "*") {
		return 0
	}
	open := strings.LastIndex(choice, "(")
	if open < 0 || !strings.HasSuffix(choice, ")") {
		return 1
	}
	count, err := strconv.Atoi(choice[open+1 : len(choice)-1])
	if err != nil || count < 2 {
		return 1
	}
	return count
}

func (s *keyDrivenSession) normalPreparedSpellLoadout() bool {
	clericCounts := map[uint8]int{}
	mageBurningHands := 0
	for _, character := range s.app.state.PartyRoster() {
		if character.HasClass(party.ClassCleric) {
			for _, spellID := range character.SpellSlots {
				clericCounts[spellID]++
			}
		}
		if character.HasClass(party.ClassMagicUser) {
			for _, spellID := range character.SpellSlots {
				if spellID == game.BurningHandsSpellID {
					mageBurningHands++
				}
			}
		}
	}
	for _, spellID := range []uint8{
		game.BlessSpellID, game.CurseSpellID, game.CureLightWoundsSpellID,
		game.CauseLightWoundsSpellID, game.ProtectionFromEvilSpellID,
	} {
		if clericCounts[spellID] < 1 {
			return false
		}
	}
	if clericCounts[23] < 5 || clericCounts[42] < 1 {
		return false
	}
	return mageBurningHands >= 4
}

// keyDrivenBoost 決定要不要把隊伍撐起來（`COAB_KEY_BOOST=0` 關掉）。
//
// ★★ **為什麼預設要撐。** 這一場問的是「按鍵到得了多少內容」，而擋在前面的是
// **戰術**不是內容：六個一級戰士（HP 10、AC 9）在提爾佛頓就打不贏十個酒館客人，
// 而重放的戰鬥處置只有原作的「快速戰鬥」。實測未撐的隊伍在 12,000 幀裡全滅重開
// 三十幾次，路線在第一次全滅之後就對不上，整場走不出開場那一段。
// `cmd/cell-sweep`／`cmd/dungeon-walk-probe` 早就是這樣做的（`gamecorpus.BoostParty`）。
//
// ⚠ **所以這一場的數字不能讀成「正常隊伍走得到這麼多」。** 它證明的是
// 「這些內容按鍵到得了、而且演出來是中文」；「正常隊伍打不打得贏」是另一把尺，
// 由戰鬥那一側的測試負責。報表會把這件事印出來，不要只引用格數。
func keyDrivenBoost() bool { return os.Getenv("COAB_KEY_BOOST") != "0" }

// boostNote 把「隊伍撐過」這件事印在數字旁邊，免得格數被引用成「正常隊伍走得到」。
func boostNote(boosted bool) string {
	if !boosted {
		return "（隊伍**沒有**撐過：這一場連戰術一起量）"
	}
	return "（⚠ 隊伍撐過：量的是內容按鍵到不到得了，**不是**正常隊伍打不打得贏）"
}

// reassertBoost 在每一幀（非戰鬥時）檢查撐過的隊伍還在不在。
//
// ★ 為什麼需要。 撐隊伍只在建角完成那一刻做一次是不夠的：**任何從 roster
// 重投影隊伍的事件都會把它洗掉**——實測開場序幕（「所有裝備都不見了」）在
// 第 33 幀就把 999×6 打回 10×6，之後整場都是六個一級戰士在打，全滅重開
// 二十幾次、路線在第一次全滅之後就對不上（第 715 輪）。
// 戰鬥中不重套：戰鬥裡的 HP 變化是戰鬥規則的結果，蓋掉會偽造戰局。
func (s *keyDrivenSession) reassertBoost(t *testing.T) {
	t.Helper()
	if !s.boosted || s.app.state.Mode == game.ModeCombat {
		return
	}
	party := s.app.state.PartyFighters()
	if len(party) == 0 {
		return
	}
	// 兩個破口都要看：投影把撐過的欄位打回原形（開場序幕、換裝、施法解除），
	// 以及**戰利品把武器塞回 roster**——下一場的自動換裝會拿它重投影，
	// 撐過的攻擊欄位就沒了（HP 留著、輸出掉回 1d8，實測 41 回合被磨死）。
	need := false
	for _, fighter := range party {
		if fighter.MaxHitPoints < 999 || fighter.AttackBonus < 100 || fighter.AttacksPerTurn < 8 {
			need = true
			break
		}
	}
	if !need {
		for _, character := range s.app.state.PartyRoster() {
			if len(character.Equipment) > 0 {
				need = true
				break
			}
		}
	}
	if need {
		s.boostParty(t)
	}
}

// boostParty 把隊伍撐到足以走完內容盤點的程度，欄位與 `gamecorpus.BoostParty`
// 同一組。⚠ 這裡不能直接用那一支：它吃 `*game.State`，而這一場的 State 在 app 裡。
func (s *keyDrivenSession) boostParty(t *testing.T) {
	t.Helper()
	if !keyDrivenBoost() {
		return
	}
	// ★ 裝備要一併清掉。撐上去的攻擊／傷害欄位是**裝備衍生欄位**，戰鬥中
	// 快速戰鬥隊員的自動換裝（spec 1120）會從 roster 的裝備重投影、把它們
	// 蓋回真實武器的值——實測 999 HP 的隊伍拿著 1d8 在奧提尤格金字塔被
	// 磨了 41 回合全滅（第 715 輪）。roster 沒有裝備，自動換裝就是 no-op，
	// 撐過的欄位才活得過戰鬥。這也呼應開場劇情：「所有裝備都不見了」。
	roster := s.app.state.PartyRoster()
	for index := range roster {
		if len(roster[index].Equipment) > 0 {
			roster[index].Equipment = nil
		}
		// roster 與 Fighter 必須一起撐。神殿治療、戰後同步等正式路徑會把
		// Fighter HP 寫回角色；只撐 Fighter 會產生 999/5 的非法角色記錄。
		roster[index].HitPoints, roster[index].MaxHitPoints = 999, 999
		roster[index].BaseMaxHitPoints = 999
		roster[index].HealthStatus = party.HealthStatusOK
		roster[index].Bleeding = 0
	}
	if err := s.app.state.SetPartyRoster(roster); err != nil {
		t.Fatalf("同步強化 roster 失敗：%v", err)
	}
	party := s.app.state.PartyFighters()
	if len(party) == 0 {
		t.Fatal("建角完成之後隊伍是空的")
	}
	for index := range party {
		party[index].HitPoints, party[index].MaxHitPoints = 999, 999
		party[index].ArmorClass = -10
		party[index].AttackBonus = 100
		party[index].DamageDiceCount, party[index].DamageDiceSides = 1, 1
		party[index].DamageBonus = 100
		party[index].AttacksPerTurn = 8
		party[index].InitiativeBonus = 100
	}
	if err := s.app.state.SetParty(party); err != nil {
		t.Fatalf("撐隊伍失敗：%v", err)
	}
	s.boosted = true
}

// playCombatTurn 走一次戰鬥的按鍵。子選單／子模式先收掉，再按 `Q`。
//
// ⚠ 順序不能反：`Q` 在移動模式或施法選目標時是別的意思，先收掉才按得對。
func (s *keyDrivenSession) playCombatTurn(t *testing.T) {
	t.Helper()
	// QUICK 住在角色記錄中，上一戰全員委派會帶進下一場。StartCombat 會在
	// 新遭遇邊界先 yield，這一幀要先按原版的 Space 收回玩家角色；否則下一個
	// 推進鍵仍可能讓整場由 AI 同步算完，沒有戰術或施法介入機會。
	hasManualPlayer := false
	hasPersistentQuick := false
	for _, fighter := range s.app.state.CombatFighters() {
		if fighter.Side != combat.SideParty || fighter.ControlMorale >= 0x80 || fighter.HitPoints <= 0 {
			continue
		}
		if fighter.QuickFight {
			hasPersistentQuick = true
		} else {
			hasManualPlayer = true
		}
	}
	if hasPersistentQuick && !hasManualPlayer {
		tap(t, s.app, s.keys, ebiten.KeySpace)
		return
	}
	// QUICK 的法術 AI 是原版獨立的 ALT+M 開關，而且每場戰鬥重設為關閉。
	// 正常隊伍已透過 CAMP 準備法術；先以正式組合鍵允許 AI 使用它們，否則
	// 單按 Q 只會近戰，記憶槽整場原封不動（spec 424）。
	if !s.app.state.CombatQuickMagicEnabled() {
		tapWithModifier(t, s.app, s.keys, ebiten.KeyAltLeft, ebiten.KeyM)
		return
	}
	switch {
	case s.app.combatSpeedMenu, s.app.combatDoneMenu:
		tap(t, s.app, s.keys, ebiten.KeyEscape)
	case s.app.state.CombatViewActive():
		tap(t, s.app, s.keys, ebiten.KeyEscape)
	case s.app.state.CombatMoveMode():
		tap(t, s.app, s.keys, ebiten.KeyEscape)
	case s.app.state.CombatCastingSpell() != 0:
		// B/S/... 先進入選法術狀態；真人下一步按 Enter 才確認施法。
		tap(t, s.app, s.keys, ebiten.KeyEnter)
	default:
		s.combatTurns++
		if s.tracing() && s.combatTurns == 1 {
			for _, character := range s.app.state.PartyRoster() {
				s.moveTrace = append(s.moveTrace, fmt.Sprintf(
					"幀%04d 角色法術 %s class=%d level=%d health=%d HP=%d/%d slots=%v known=%v",
					s.frames, character.Name, character.Class, character.Level,
					character.HealthStatus, character.HitPoints, character.MaxHitPoints,
					character.SpellSlots, character.KnownSpells))
			}
			for _, fighter := range s.app.state.CombatFighters() {
				s.moveTrace = append(s.moveTrace, fmt.Sprintf(
					"幀%04d 戰鬥開場 %s side=%d HP=%d/%d AC=%d AB=%d damage=%dd%d%+d attacks=%d quick=%t pos=(%d,%d)",
					s.frames, fighter.Name, fighter.Side, fighter.HitPoints, fighter.MaxHitPoints,
					fighter.ArmorClass, fighter.AttackBonus, fighter.DamageDiceCount,
					fighter.DamageDiceSides, fighter.DamageBonus, fighter.AttacksPerTurn,
					fighter.QuickFight, fighter.CombatX, fighter.CombatY))
			}
		}
		// 原版 C 會先開正式 CAST 清單；祝福是整備時放進牧師第一個槽的
		// 第一項，因此下一幀以 Enter 選定。這條路同時證明前端沒有再把
		// C 偷綁成單一「詛咒術」快捷鍵。
		if s.app.combatSpellMenu {
			spellID := uint8(0)
			if s.app.state.CombatCanCastBurningHands() {
				spellID = game.BurningHandsSpellID
			} else if s.app.state.CombatCanCastBless() {
				spellID = game.BlessSpellID
			}
			choices := s.app.state.CombatSpellChoices()
			for index, choice := range choices {
				if choice.SpellID != spellID {
					continue
				}
				if s.app.combatSpellCursor != index {
					tap(t, s.app, s.keys, ebiten.KeyDown)
					return
				}
				tap(t, s.app, s.keys, ebiten.KeyEnter)
				return
			}
			// CAST 清單在開啟後若已不再包含預定法術，安全退出並讓本回合
			// 落到 QUICK；不能任選第 0 項，也不能留在空選單裡。
			tap(t, s.app, s.keys, ebiten.KeyEscape)
			return
		}
		// 已確認、仍在原版施法延遲中的動作，要等排程重新輪到施法者才用 Enter
		// 推進。法術等待期間若目前是另一名手動隊員，仍須先完成該隊員的正常
		// 回合；把「場上有人有待決法術」誤當成「目前就是施法者」會讓重放每幀
		// 都替另一名角色按 Enter，永遠輪不回施法者（spec 1233）。
		for _, candidate := range s.app.state.CombatFighters() {
			if candidate.CombatAction.SpellID != 0 {
				current, hasCurrent := s.app.state.CombatActiveFighter()
				if s.tracing() {
					active := "none"
					if hasCurrent {
						active = fmt.Sprintf("%s side=%d hp=%d quick=%t control=%#x", current.Name,
							current.Side, current.HitPoints, current.QuickFight, current.ControlMorale)
					}
					s.moveTrace = append(s.moveTrace, fmt.Sprintf(
						"幀%04d 待決法術 %s spell=%#02x delay=%d target=%q point=%t(%d,%d)；active=%s",
						s.frames, candidate.Name, candidate.CombatAction.SpellID,
						candidate.CombatAction.Delay, candidate.CombatAction.TargetID,
						candidate.CombatAction.HasTargetPoint, candidate.CombatAction.TargetX,
						candidate.CombatAction.TargetY, active))
				}
				if !hasCurrent || current.ID == candidate.ID {
					tap(t, s.app, s.keys, ebiten.KeyEnter)
					return
				}
				break
			}
		}
		// ⚠ 不是自己活著的隊員回合就按 Enter（`CombatAct` 那顆鍵），不是 `Q`。
		// 排程中的快速施法（「開始吟唱…」）會把回合表走到盡頭再把控制交回
		// 玩家，這時 `Q`（快速戰鬥切換）的第一步就是「要有隊員回合」——
		// 永遠回「it is not a living party turn」，而推進戰局的
		// `advanceCombatToParty` 只有動作成功才會被叫到 ⇒ 整場停格。
		// 真人玩家按 Enter 就過得去；重放照做（第 715 輪）。
		if fighter, ok := s.app.state.CombatActiveFighter(); ok &&
			fighter.Side == combat.SideParty && fighter.HitPoints > 0 && !fighter.QuickFight {
			// 法師確實準備了燃燒之手；只有既有的鄰接敵人／職業／記憶槽
			// 閘門成立時，才透過正式 C → CAST 清單施放。法師只有這一支
			// 記憶法術，所以 Enter 會選中它，不需要測試專用直呼。
			if s.app.state.CombatCanCastBurningHands() && s.combatSpellChoiceAvailable(game.BurningHandsSpellID) {
				tap(t, s.app, s.keys, ebiten.KeyC)
				return
			}
			// 整備路徑已讓牧師用正式 MEMORIZE → REST 準備祝福；在第一個
			// 合法牧師回合按原版公開的 C → CAST 施放，不能讓 QUICK 的「不自動施法」
			// 把玩家實際擁有的法術永遠留在槽裡。
			if s.app.state.CombatCanCastBless() && s.combatSpellChoiceAvailable(game.BlessSpellID) {
				for _, character := range s.app.state.PartyRoster() {
					if character.ID == fighter.ID && character.Class == party.ClassCleric {
						tap(t, s.app, s.keys, ebiten.KeyC)
						return
					}
				}
			}
			// 已委派的前一名角色必須維持 QUICK。若在下一名手動角色接棒時按
			// Space，會把前一名收回；下一幀 Q 又只委派目前角色，兩個動作互相
			// 抵消，戰局便永遠停在同一輪。只有戰鬥開場「所有玩家角色都已
			// QUICK」時，才由函式頂端的閘門收回一次，確保新遭遇仍有玩家輸入。
			if os.Getenv("COAB_KEY_MANUAL_COMBAT") == "1" {
				tap(t, s.app, s.keys, ebiten.KeyEnter)
			} else {
				tap(t, s.app, s.keys, ebiten.KeyQ)
			}
		} else {
			tap(t, s.app, s.keys, ebiten.KeyEnter)
		}
	}
	// 戰鬥最後一個按鍵可能在同一個 Update 內完成 POSTCOM，下一幀 observe()
	// 看見的已是後續 ECL 頁面。不能只靠 state.Message 的跨幀取樣記戰敗；在
	// 正式戰鬥訊息仍可見的提交點同步立起恢復旗標，否則很短的敗戰頁會漏掉，
	// 0 HP 隊伍便繼續走到下一個 DAMAGE 才真正全滅。
	if strings.Contains(s.app.state.CombatMessage(), "戰鬥失敗。") {
		s.normalRecoveryActive = true
		s.normalRecoveryDaysAdded = 0
	}
	if s.tracing() {
		alive, hp, enemies := 0, 0, 0
		for _, fighter := range s.app.state.CombatFighters() {
			if fighter.HitPoints <= 0 {
				continue
			}
			if fighter.Side == combat.SideParty {
				alive++
				hp += fighter.HitPoints
			} else {
				enemies++
			}
		}
		s.moveTrace = append(s.moveTrace, fmt.Sprintf(
			"幀%04d 戰鬥 我方 %d 人／HP %d，敵方 %d 人 ⇒ %.40q",
			s.frames, alive, hp, enemies, s.app.state.CombatMessage()))
	}
}

func (s *keyDrivenSession) combatSpellChoiceAvailable(spellID uint8) bool {
	for _, choice := range s.app.state.CombatSpellChoices() {
		if choice.SpellID == spellID {
			return true
		}
	}
	return false
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

// tryDoors 朝**走不通**的方向各撞一次，看那裡是不是門。撞到門會開選單，
// 交給 `handleDoorMenu`。
//
// ⚠ **不能對走得通的方向按前進**：那只會把隊伍帶走。舊版是「轉一圈、每一面
// 按一次前進」，於是站在只有一面是門的死角時，第一次前進就沿著開著的那一面
// 離開了，門永遠不會被碰到——尤拉什 `(4,7)(4,8)(5,7)` 那個死角實測繞了 1,700 幀、
// 門選單一次都沒開，而出口就是 `(4,8)` 東面那一道上鎖的門（spec 1198）。
//
// ⚠ 也**不能**只看目前朝向：門可能在背後。四個方向都要輪，只是先過濾掉
// 走得通的那些。
func (s *keyDrivenSession) tryDoors(t *testing.T) {
	t.Helper()
	for _, heading := range []int{0, 2, 4, 6} {
		if s.app.state.Mode != game.ModeDungeon {
			return
		}
		deltaX, deltaY := headingDelta(heading)
		if s.app.state.CanMoveDungeon(*s.app.geoGrid, deltaX, deltaY, heading) {
			// 走得通就不是門，按下去只會離開這一格。
			continue
		}
		if !s.faceHeading(t, heading) {
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
	}
}

// faceHeading 用 `M`（順時針轉 2）把隊伍轉到指定朝向；轉不過去或中途離開地城
// 就回 false。
func (s *keyDrivenSession) faceHeading(t *testing.T, heading int) bool {
	t.Helper()
	for turn := 0; turn < 4; turn++ {
		if _, _, facing := s.app.state.DungeonGeometryView(); int(facing) == heading {
			return true
		}
		tap(t, s.app, s.keys, ebiten.KeyM)
		if s.app.state.Mode != game.ModeDungeon {
			return false
		}
	}
	return false
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
	fallbackVisits := 0
	for index := range headings {
		heading := headings[(index+offset)%len(headings)]
		deltaX, deltaY := headingDelta(heading)
		if !s.app.state.CanMoveDungeon(*s.app.geoGrid, deltaX, deltaY, heading) {
			continue
		}
		x, y, _ := s.app.state.DungeonGeometryView()
		// ⚠ 鄰格座標要**照移動的規則繞回來**（`geo.WrapCoordinate`）。
		// 不繞的話，站在 `(15,0)` 往北算出來的是 `(15,−1)`，而隊伍真的踏上去
		// 之後會在 `(15,15)`；於是 `visits[(15,−1)]` 永遠是 0、
		// `visits[(15,16)]` 也永遠是 0，兩格互相看起來都是「去得最少的」
		// ⇒ 隊伍在地圖上下邊界之間乒乓。實測在 `ECL2/0x03` 的
		// `(15,0)`／`(15,15)` 各繞了 3,500 次、把 7,000 幀花光。
		neighbour := [3]int{
			int(s.app.geoBlock),
			geo.WrapCoordinate(x+deltaX, geo.Width),
			geo.WrapCoordinate(y+deltaY, geo.Height),
		}
		// ⚠ **退路要挑去過最少次的那一格**，不能挑「第一個走得通的」。
		// 只挑第一個會讓隊伍在兩格之間來回：實測新錄的路線把隊伍帶到下水道
		// 天花板那一格之後，就在那一格與鄰格之間震盪到跑完，因為那一格有事件、
		// 每次踏上去都要重答一次。去過次數最少的那一格會把隊伍推離熱點。
		if visits := s.visits[neighbour]; !hasFallback || visits < fallbackVisits {
			fallback, hasFallback, fallbackVisits = heading, true, visits
		}
		target := [4]int{neighbour[0], neighbour[1], neighbour[2], heading}
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
	// 標題 → 原版種族／性別／職業／陣營 → 擲點 → 姓名 → 儲存 → D 完成隊伍。
	tap(t, application, keys, ebiten.KeyEnter)
	if application.state.Mode != game.ModeCharacterCreation {
		t.Fatalf("按 Enter 應該進角色建立，實際 %s", modeName(application.state.Mode))
	}
	if application.state.GuidedActive || application.state.CreationOuterStep != game.CreationOuterMenu {
		t.Fatalf("正常入口應停在原版外層功能選單：active=%v outer=%d", application.state.GuidedActive, application.state.CreationOuterStep)
	}
	creationTheme := application.ui.settings.Theme
	tap(t, application, keys, ebiten.KeyF2)
	firstTheme := application.ui.settings.Theme
	if firstTheme == creationTheme || application.state.CreationOuterStep != game.CreationOuterMenu {
		t.Fatalf("建隊外層 F2 沒有切換 theme 或改壞流程：before=%q after=%q outer=%d",
			creationTheme, application.ui.settings.Theme, application.state.CreationOuterStep)
	}
	tap(t, application, keys, ebiten.KeyF2)
	if application.ui.settings.Theme == firstTheme || (application.ui.settings.Theme != "original" && application.ui.settings.Theme != "modern-a6") {
		t.Fatalf("建隊外層第二次 F2 沒有切到另一個有效 theme：first=%q second=%q", firstTheme, application.ui.settings.Theme)
	}
	tap(t, application, keys, ebiten.KeyC)
	if !application.state.GuidedActive || application.state.GuidedStep != game.CreationStepRace {
		t.Fatalf("按 C 應到原版種族選單：active=%v step=%d", application.state.GuidedActive, application.state.GuidedStep)
	}
	for index := 0; index < 4; index++ {
		tap(t, application, keys, ebiten.KeyEnter)
	}
	if application.state.GuidedStep != game.CreationStepAbilities || application.state.GuidedDraft.Abilities == (party.Abilities{}) {
		t.Fatalf("選完四段應自動擲點：step=%d abilities=%+v", application.state.GuidedStep, application.state.GuidedDraft.Abilities)
	}
	tap(t, application, keys, ebiten.KeyEnter)
	typeText(t, application, keys, "測試者")
	tap(t, application, keys, ebiten.KeyEnter)
	tap(t, application, keys, ebiten.KeyY)
	if got := len(application.state.CreationRoster); got != 0 {
		t.Fatalf("原版流程儲存後尚未加入隊伍，實際 %d", got)
	}
	tap(t, application, keys, ebiten.KeyA)
	tap(t, application, keys, ebiten.KeyC)
	tap(t, application, keys, ebiten.KeyEnter)
	if got := len(application.state.CreationRoster); got != 1 {
		t.Fatalf("從 Curse 角色清單加入後應有一名隊員，實際 %d", got)
	}
	tap(t, application, keys, ebiten.KeyEscape)
	tap(t, application, keys, ebiten.KeyB)
	if application.state.Mode == game.ModeCharacterCreation {
		t.Fatal("按 D 之後還停在角色建立：完成那條路按不出來")
	}
	session.boostParty(t)

	// ⚠ 幀數上限是**量出來的不是猜的**：換掉走法之後要重新量一次到頂的位置，
	// 用 `COAB_KEY_FRAMES` 掃一遍再把預設值訂在轉折點上（見 spec 1197）。
	// 不要為了「看起來跑得久」而拖慢整個測試套件。
	dungeonThemeChecked := false
	for session.frames = 0; session.frames < keyDrivenFrames(); session.frames++ {
		session.observe()
		if !dungeonThemeChecked && application.state.Mode == game.ModeDungeon {
			before := application.ui.settings.Theme
			x, y, facing := application.state.DungeonGeometryView()
			tap(t, application, keys, ebiten.KeyF2)
			if application.ui.settings.Theme == before {
				t.Fatal("第一張地圖內 F2 沒有切換 theme")
			}
			if gotX, gotY, gotFacing := application.state.DungeonGeometryView(); gotX != x || gotY != y || gotFacing != facing {
				t.Fatalf("切換 theme 改變地圖狀態：before=(%d,%d,%d) after=(%d,%d,%d)", x, y, facing, gotX, gotY, gotFacing)
			}
			tap(t, application, keys, ebiten.KeyF2)
			dungeonThemeChecked = true
		}
		// 一般強度報表量的是一條連續冒險。全滅後由標題建立新隊伍已是另一場
		// attempt，不能把多場走過的格子與段數聯集起來冒充單次進度。
		if !keyDrivenBoost() && application.state.PartyKilled() {
			break
		}
		session.step(t)
		if application.state.GameWon() {
			session.wonAt = session.frames
			break
		}
	}
	session.observe()
	if !dungeonThemeChecked {
		t.Fatal("按鍵路徑未在第一張地圖實際測到 F2 theme 切換")
	}
	// 失敗斷言之前先落 trace；否則早停場景會在原本的收尾寫檔前 t.Fatal，
	// COAB_KEY_TRACE 反而只對通過的 run 有效。
	if path := os.Getenv("COAB_KEY_TRACE"); path != "" {
		if !filepath.IsAbs(path) {
			path = filepath.Join("..", "..", path)
		}
		if err := os.WriteFile(path,
			[]byte(strings.Join(session.moveTrace, "\n")+"\n"), 0o644); err != nil {
			t.Fatalf("逐格紀錄寫不出來：%v", err)
		}
	}

	if !session.modes[game.ModeDungeon] {
		t.Fatal("整場都沒走進地城：地城那一層按鍵到不了")
	}
	if !session.modes[game.ModeEvent] {
		t.Fatal("整場都沒觸發任何劇情事件")
	}
	if len(session.cells) < 2 {
		t.Fatalf("只站上過 %d 格：走不動", len(session.cells))
	}
	if !keyDrivenBoost() && session.frames >= keyDrivenFrames() && len(session.blocks) == 1 && session.blocks[0x01] {
		spellState := make([]string, 0, len(application.state.PartyRoster()))
		for _, character := range application.state.PartyRoster() {
			spellState = append(spellState, fmt.Sprintf("%s slots=%v", character.Name, character.SpellSlots))
		}
		t.Fatalf("一般強度路線在 %d 幀上限仍未離開 ECL 0x01；mode=%v event=%q killed=%t message=%q prompt=%q choices=%q spells=%s",
			session.frames, application.state.Mode, application.state.OriginalEvent,
			application.state.PartyKilled(), application.state.Message, application.state.Prompt,
			application.state.Choices, strings.Join(spellState, "；"))
	}
	t.Logf("按鍵驅動 %d 幀：走過 %d 格、%d 種畫面、記到 %d 句話、撞到門 %d 次、"+
		"照路線按 %d 次（路線共 %d 步，查表覆蓋 %d 格／%d 種選單），"+
		"快速戰鬥 %d 次、全滅重開 %d 次（記住 %d 個致命選項），落回原文 %d 句%s",
		session.frames, len(session.cells), len(session.modes), len(session.messages),
		session.doorsFound, session.routeHits, len(session.route),
		len(session.routeMoves), len(session.routeChoices),
		session.combatTurns, session.wipes, len(session.fatalPicks), len(session.fallbacks),
		boostNote(session.boosted))
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
	if application.state.Mode == game.ModeCombat {
		t.Logf("  戰鬥停格 frontend spellMenu=%t cursor=%d doneMenu=%t speedMenu=%t casting=%#02x quickMagic=%t message=%q",
			application.combatSpellMenu, application.combatSpellCursor,
			application.combatDoneMenu, application.combatSpeedMenu,
			application.state.CombatCastingSpell(), application.state.CombatQuickMagicEnabled(),
			application.state.Message)
		t.Logf("  戰鬥停格 gates bless=%t burningHands=%t spellChoices=%+v",
			application.state.CombatCanCastBless(), application.state.CombatCanCastBurningHands(),
			application.state.CombatSpellChoices())
		if active, ok := application.state.CombatActiveFighter(); ok {
			t.Logf("  戰鬥停格 active=%s side=%d hp=%d quick=%t control=%#x action=%+v",
				active.Name, active.Side, active.HitPoints, active.QuickFight,
				active.ControlMorale, active.CombatAction)
		} else {
			t.Logf("  戰鬥停格 active=none")
		}
		if event, ok := application.state.CombatVisualEvent(); ok {
			t.Logf("  戰鬥停格 visual serial=%d elapsed=%s duration=%s event=%+v",
				event.Serial, application.state.CombatVisualElapsed(), event.Duration(), event)
		}
		for _, fighter := range application.state.CombatFighters() {
			t.Logf("  戰鬥停格 fighter=%s side=%d hp=%d quick=%t control=%#x action=%+v",
				fighter.Name, fighter.Side, fighter.HitPoints, fighter.QuickFight,
				fighter.ControlMorale, fighter.CombatAction)
		}
	}
	for pick := range session.fatalPicks {
		t.Logf("  致命選項：第 %d 項 於 %.60s", pick.index, pick.signature)
	}
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

	// ⚠ 這一條**排在報表後面**：落回原文代表隊伍走到了還沒翻的新內容，
	// 那一刻最想知道的是「走到哪裡才碰到的」。排在前面的話 t.Fatalf 會把
	// 走過幾格、段序列、選單清單全部吃掉，只剩兩行英文，看不出是哪一段的內容。
	if len(session.fallbacks) > 0 {
		for text := range session.fallbacks {
			t.Errorf("落回原文：%q", text)
		}
		t.Fatalf("按鍵玩到的畫面有 %d 句落回原文", len(session.fallbacks))
	}
	// 一般強度的真實全滅是合法玩家結果，不是 remake 缺陷。這場已在上方
	// PartyKilled 閘門停止並寫出報表；不得為了讓自動測試綠而改隊伍、敵人或
	// 戰術重跑。強化的輸入可達性樣本仍保留 REQUIRE_WIN 硬閘門。
	if os.Getenv("COAB_KEY_REQUIRE_WIN") == "1" && session.wonAt < 0 &&
		(keyDrivenBoost() || !application.state.PartyKilled()) {
		t.Fatalf("要求正常按鍵通關，但 %d 幀後仍未抵達結局", session.frames)
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
	type partySummary struct {
		Name           string `json:"name"`
		Level          int    `json:"level"`
		Experience     uint32 `json:"experience"`
		HitPoints      int    `json:"hit_points"`
		MaxHP          int    `json:"max_hit_points"`
		HealthStatus   uint8  `json:"health_status"`
		Bleeding       int    `json:"bleeding"`
		Gold           uint16 `json:"gold"`
		Platinum       uint16 `json:"platinum"`
		Equipment      int    `json:"equipment"`
		ProjectedHP    int    `json:"projected_hit_points"`
		ProjectedMaxHP int    `json:"projected_max_hit_points"`
	}
	projected := make(map[string]combat.Fighter)
	for _, fighter := range s.app.state.PartyFighters() {
		projected[fighter.ID] = fighter
	}
	partyRows := make([]partySummary, 0, len(s.app.state.PartyRoster()))
	for _, character := range s.app.state.PartyRoster() {
		fighter := projected[character.ID]
		partyRows = append(partyRows, partySummary{
			Name: character.Name, Level: character.Level, Experience: character.Experience,
			HitPoints: character.HitPoints, MaxHP: character.MaxHitPoints,
			HealthStatus: uint8(character.HealthStatus), Bleeding: character.Bleeding,
			Gold: character.Gold, Platinum: character.Platinum, Equipment: len(character.Equipment),
			ProjectedHP: fighter.HitPoints, ProjectedMaxHP: fighter.MaxHitPoints,
		})
	}
	report := struct {
		Schema    string         `json:"schema"`
		Frames    int            `json:"frames"`
		Cells     int            `json:"cells"`
		Modes     []string       `json:"modes"`
		Messages  int            `json:"messages"`
		Fallbacks int            `json:"fallbacks"`
		Doors     int            `json:"doors_found"`
		Menus     int            `json:"menus"`
		Segments  int            `json:"segments"`
		RouteHits int            `json:"route_hits"`
		RouteLen  int            `json:"route_steps"`
		Won       bool           `json:"won"`
		WonAt     int            `json:"won_at_frame,omitempty"`
		Boosted   bool           `json:"boosted"`
		Wipes     int            `json:"party_wipes"`
		Party     []partySummary `json:"party"`
	}{
		Schema: "coab-key-driven-session/1", Frames: s.frames, Cells: len(s.cells),
		Modes: modes, Messages: len(s.messages), Fallbacks: len(s.fallbacks),
		Doors: s.doorsFound, Menus: len(s.menus),
		Segments: len(s.blocks), RouteHits: s.routeHits, RouteLen: len(s.route),
		Won: s.wonAt >= 0, WonAt: s.wonAt,
		Boosted: s.boosted, Wipes: s.wipes, Party: partyRows,
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
// snapshotTap 與 tap 相同，只是**不會 t.Fatalf**：把錯誤帶回去讓呼叫端記帳。
//
// ★ 為什麼不能沿用 `tap`。 114 份快照裡有一份推到一半回錯，`t.Fatalf` 會把
// 整趟走訪停在那裡，剩下的一份都沒跑 ⇒ 報表上看起來像「量不到」，而實際上是
// 「113 份好的、1 份撞到一個還沒判的東西」。這個專案的其他報表一律拆成
// 「已判定／還沒判的」，這裡也一樣。
func snapshotTap(application *app, keys *scriptedKeys, key ebiten.Key) error {
	keys.press(key)
	if err := application.Update(); err != nil {
		return fmt.Errorf("按 %v：%w", key, err)
	}
	keys.release()
	if err := application.Update(); err != nil {
		return fmt.Errorf("放開 %v：%w", key, err)
	}
	return nil
}

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
	// blocked 是「推到一半回錯」的快照：這一份還沒判是 remake 的缺口還是
	// 量測的限制，所以獨立記一格，不混進 `driven` 也不當成沒反應。
	blocked := map[string]string{}
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
			switch {
			case state.Mode == game.ModeDungeon:
				key = ebiten.KeyUp
				if application.geoGrid != nil && !application.canStepForward() {
					key = ebiten.KeyM
				}
			case state.Mode == game.ModeCharacterCreation &&
				len(state.CreationRoster) >= 6:
				// ⚠ 建角滿六名之後再按 Enter 會回
				// `party already has six characters` 並讓 `Update()` 報錯 ⇒
				// 整個快照走訪死在那一份上。完成建立的鍵是 `D`，
				// 與 `step()` 那一側同一條路（那裡是全滅重開走到的）。
				key = ebiten.KeyD
			}
			if err := snapshotTap(application, keys, key); err != nil {
				blocked[name] = err.Error()
				break
			}
			after := application.state
			_, _, afterFacing := after.DungeonGeometryView()
			if fmt.Sprintf("%v/%d/%d/%d/%q",
				after.Mode, after.DungeonX, after.DungeonY, afterFacing, after.Message) != before {
				moved = true
			}
		}
		if reason, stopped := blocked[name]; stopped {
			t.Logf("%s：推到一半停下來——%s", name, reason)
			continue
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
	t.Logf("按鍵推得動的快照 %d／%d，落回原文 %d 句，推到一半回錯 %d 份",
		driven, len(names), len(fallbacks), len(blocked))
	// ⚠ 閘釘在**相異成因**不是份數。份數會隨快照數量漂（先前同一個缺陷就擋住
	// 五份 `inside-block-42-*`：全滅重開沒把章節收回開局值，新隊伍帶著章 6 走進
	// 提爾佛頓，商店 TREASURE 拿章 6 查 ITEM 區塊 1 而回錯——已修在
	// `resetSessionForNewGame`）。目前宣告 0 種，多一種就紅。
	// 不要為了讓測試綠而把它調大。
	causes := map[string]int{}
	for _, reason := range blocked {
		causes[reason]++
	}
	if len(causes) > 0 {
		t.Errorf("推到一半回錯的相異成因有 %d 種，宣告的上限是 0：%v", len(causes), causes)
	}
	if path := os.Getenv("COAB_KEY_SNAPSHOT_JSON"); path != "" {
		report := struct {
			Schema    string `json:"schema"`
			Snapshots int    `json:"snapshots"`
			Driven    int    `json:"driven"`
			Fallbacks int    `json:"fallbacks"`
			Known     int    `json:"known_variable_insert"`
			Blocked   int    `json:"blocked"`
		}{
			Schema: "coab-key-driven-snapshots/1", Snapshots: len(names), Driven: driven,
			Fallbacks: len(fallbacks), Known: len(knownVariableInsert),
			Blocked: len(blocked),
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
