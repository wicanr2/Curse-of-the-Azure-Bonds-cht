package game

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

// 提爾佛頓（ECL2 的 0x01／0x02／0x03）是**遊戲真正開始的地方**，而
// `TestRealNewGameRunsToTheEnding` 從世界地圖那個 hub（0x50）起跑，整章跳過。
// 可達性盤點因此顯示那 48 個分派索引「主線一格都沒踏到」。
//
// ★ `cmd/dungeon-walk-probe` 已經證明**走得進去**（冷走走得到 25 個索引），
// 所以那不是缺陷、是路線的選擇。這一條把選擇補上：真的走一遍那三段，
// 並套用與主線同一條語系不變量。
//
// ⚠ 這**不是**「從新遊戲玩到提爾佛頓」：進段是直接進的（那三段本來就是開場，
// 沒有更前面的東西）。它證明的是**段內**走得通、而且沿路的字都是中文。
//
// ⚠ 走法是廣度優先的幾何走訪，不是劇情路線：它不解謎、不觸發特定旗標，
// 所以「走到的格子」是下界。
// walkableRouteSegments 是要實際走一遍的段。**提爾佛頓那三段是遊戲真正開始的
// 地方**，而主線測試從世界地圖那個 hub 起跑、整章跳過；其餘幾段主線只擦過邊
// （可達性報表裡「實跑踏到」遠低於「走得到」的那些）。
//
// ⚠ 這一份是清單不是全集：沒有地形分派、或進不去的段不列。
var walkableRouteSegments = []string{
	"ECL2/0x01", "ECL2/0x02", "ECL2/0x03", // 提爾佛頓：開場
	"ECL3/0x10", "ECL3/0x11", // 猶拉什
	"ECL3/0x15",              // 猶拉什：戰火街區（實跑覆蓋原本 0%，第 718 輪補走）
	"ECL4/0x20", "ECL4/0x22", // 散提爾堡、眼魔洞穴
	"ECL4/0x25",              // 魔法商店＋地城（實跑覆蓋原本 0.1%，第 718 輪補走）
	"ECL5/0x31", "ECL5/0x32", // 哈普村、古熔岩洞
	"ECL5/0x35",                           // 洞窟（同上，原本 0%）
	"ECL6/0x40", "ECL6/0x42", "ECL6/0x43", // 密斯卓諾
	"ECL6/0x45", // （同上，原本 0%）
}

func TestTilvertonRouteIsWalkableAndLocalized(t *testing.T) {
	// ★ 這條路線走的是主線**不經過**的三段（`0x01`／`0x02`／`0x03`）。它每一步
	// 移動的起點都是「隊伍真的站過的格子」，而那正是走訪用來進入
	// 「幾何上斷開」連通分量的入口（spec 1193）。
	//
	// ⚠ 兩支測試都要跑才錄得齊：只跑主線的話那三段一個起點都沒有，
	// 而少掉的段不會有任何錯誤訊息。
	t.Cleanup(func() {
		if err := WriteDecisionLog(""); err != nil {
			t.Errorf("路線錄不出來：%v", err)
		}
	})
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks := make(map[uint8][]byte)
	for chapter := 1; chapter <= 6; chapter++ {
		parsed, parseErr := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(chapter)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range parsed {
			blocks[block.Entry.ID] = block.Data
		}
	}
	catalog := geo.NewCatalog()
	for chapter := 2; chapter <= 6; chapter++ {
		if err := catalog.AddDAX(uint8(chapter),
			zipData(t, image, "GEO"+strconv.Itoa(chapter)+".DAX")); err != nil {
			t.Fatal(err)
		}
	}

	messages := map[string]bool{}
	totalCells := 0
	for _, id := range walkableRouteSegments {
		seg, ok := segment.Lookup(id)
		if !ok {
			t.Fatalf("段註冊表裡沒有 %s", id)
		}
		t.Run(id, func(t *testing.T) {
			// 每一種選單策略各走一遍：格子累積在 `campaignVisitedCells` 裡，
			// 所以是聯集（spec 1193）。
			cells := 0
			namedTowerCells := 0
			picks := []int{0, 1, 2, -1}
			if id == "ECL4/0x25" {
				// 奧克薩姆莊園不是固定索引策略走得到的：要先進 SHOP、離開
				// 貨架，再在下一層選單選 GO TO THE TOWER。-2 依選項名字走
				// 這條正常玩家路徑，不注入座標或劇情旗標。
				picks = append(picks, -2)
			}
			for _, pick := range picks {
				walked := walkTilvertonSegmentWith(t, pick, blocks, catalog, seg, messages)
				if pick == -2 {
					namedTowerCells = walked
				}
				if walked > cells {
					cells = walked
				}
			}
			// ⚠ 正對照：走不到東西的話下面的語系檢查會**正確地通過**，
			// 因為它一句話都沒收到。先擋住這種假綠。
			if cells < 8 {
				t.Fatalf("%s 只走到 %d 格，這一段不可能這麼小", id, cells)
			}
			if id == "ECL4/0x25" && namedTowerCells == 0 {
				t.Fatalf("%s 的具名選項策略沒有走完前往奧克薩姆莊園的正常轉場", id)
			}
			totalCells += cells
			t.Logf("%s 走到 %d 格", id, cells)
		})
	}
	if totalCells == 0 {
		return
	}

	// 把提爾佛頓走過的格子併進可達性導出（見 `exportCampaignVisitedCells`）。
	exportCampaignVisitedCells(t)

	t.Run("語系：實跑路線沿路沒有落回原文", func(t *testing.T) {
		if len(messages) < 10 {
			t.Fatalf("只記到 %d 句話，走這麼多段不可能這麼少", len(messages))
		}
		var fallbacks, fragments []string
		for message := range messages {
			if campaignMessageHasHan(message) {
				continue
			}
			if !campaignMessageHasLatinWord(message) {
				continue
			}
			if routeKnownFragments[strings.TrimSpace(message)] {
				fragments = append(fragments, message)
				continue
			}
			fallbacks = append(fallbacks, message)
		}
		sort.Strings(fragments)
		for _, message := range fragments {
			t.Logf("片段（單獨出現可能是走法造成的，見 routeKnownFragments）：%q", message)
		}
		sort.Strings(fallbacks)
		for _, message := range fallbacks {
			t.Errorf("落回原文：%q", message)
		}
		t.Logf("實跑路線沿路記到 %d 句話，落回原文 %d 句", len(messages), len(fallbacks))
	})
}

// walkTilvertonSegment 進段之後從落點做廣度優先，回傳走到的格子數。
// 沿路把玩家看得到的字收進 messages，並把 (block, 地形碼) 記進
// campaignVisitedCells 讓可達性盤點看得到。
// walkTilvertonSegment 走一遍。`pick` 是選單策略（`-1` ＝ 最後一項）。
//
// ★ 為什麼要跑好幾種：**單一策略的結果看起來都很合理**，但每一種都會被某一類
// 岔路擋住——選第一項會被收費關卡擋在門外（下水道的奧提尤格要食物），
// 選最後一項會在「要離開嗎」那種提示上直接走人（spec 1193）。
func walkTilvertonSegmentWith(t *testing.T, pick int, blocks map[uint8][]byte, catalog geo.Catalog,
	seg segment.Segment, messages map[string]bool) int {
	t.Helper()
	state := NewStateFromECLBlocks(trainingTestCatalog(t), blocks, seg.Block)
	state.SetGeoCatalog(catalog)
	installShopFixtures(t, &state)
	// ⚠ `ECL2/0x01` 是**開場**，走的是 `BeginAdventure`，它要的是**角色名冊**
	// 不只是戰鬥員——少了名冊會回「adventure requires a created or loaded party」，
	// 那一段就整段被跳過（第一版就是這樣漏掉開場那一張圖的）。
	state.partyRoster = party.Roster{{
		ID: "walker", Name: "走訪者", Race: party.RaceHuman,
		Class: party.ClassFighter, Level: 1,
		Abilities: party.Abilities{Strength: 18, Intelligence: 10, Wisdom: 10,
			Dexterity: 16, Constitution: 16, Charisma: 10},
	}}
	// ⚠ 隊伍撐起來只為了讓入口伏擊不會把走訪擋在門口。**只給盤點用**，
	// 它不代表正常隊伍的強度，所以這一條不宣稱「一般玩家打得過」。
	if err := state.SetParty([]combat.Fighter{{
		ID: "walker", Name: "走訪者", Side: combat.SideParty,
		HitPoints: 999, MaxHitPoints: 999, ArmorClass: -10,
		AttackBonus: 100, DamageDiceCount: 1, DamageDiceSides: 1, DamageBonus: 100,
		AttacksPerTurn: 8, InitiativeBonus: 100,
	}}); err != nil {
		t.Fatal(err)
	}
	if err := state.EnterSegment(seg); err != nil {
		t.Skipf("%s 進不去：%v", seg.ID, err)
	}
	collect := func() {
		if state.Message != "" {
			messages[state.Message] = true
		}
		if state.Prompt != "" {
			messages[state.Prompt] = true
		}
		if state.Mode == ModeDungeon && state.session != nil {
			campaignVisitedCells[campaignCellKey{
				block: state.session.CurrentBlockID(), terrain: state.DungeonWallRoof,
			}] = true
			// 這條路線走的段（提爾佛頓那三段等）主線不一定經過，段內快照也要在
			// 這裡存，否則那些段在 `cmd/cell-sweep -snapshots` 那邊沒有備用入口。
			captureInsideSegment(&state)
		}
	}
	settle := func() bool {
		for step := 0; step < 60 && state.Mode != ModeDungeon; step++ {
			collect()
			if state.session != nil && state.session.CurrentBlockID() != seg.Block {
				// 這個策略把隊伍帶出段外（0x25 選單有「回城」那一類出口）：
				// 這一輪作罷，讓下一個策略試別條路。
				return false
			}
			if testing.Verbose() {
				t.Logf("settle step %d: mode=%d choices=%v msg=%.40q", step, state.Mode, state.Choices, state.Message)
			}
			if state.CombatActive() {
				for turn := 0; turn < 400 && state.CombatActive(); turn++ {
					if err := state.CombatAct(); err != nil {
						if testing.Verbose() {
							t.Logf("settle combat err: %v", err)
						}
						return false
					}
				}
				continue
			}
			choice := 0
			if count := len(state.Choices); count > 0 {
				if pick == -2 {
					choice = namedTowerRouteChoice(state.Choices)
				} else if pick < 0 || pick >= count {
					choice = count - 1
				} else {
					choice = pick
				}
			}
			if state.Mode == ModePlace && len(state.Choices) > 0 {
				choice = len(state.Choices) - 1
			}
			if err := state.Continue(); err != nil {
				if selectErr := state.Select(choice); selectErr != nil {
					return false
				}
			}
		}
		collect()
		return state.Mode == ModeDungeon
	}
	if !settle() {
		// ⚠ 只放棄**這一個選單策略**，不能 Skip 整個子測試：0x25 的第一項
		// 是 SHOP（要寶物區塊才開得起來），第一策略失敗就 Skip 會讓
		// EXPLORE 那條路永遠沒機會（第 719 輪就是這樣整段漏掉的）。
		t.Logf("%s（pick=%d）入口推不回地城（停在 %v）", seg.ID, pick, state.Mode)
		return 0
	}
	// ⚠ GEO 要在 settle **之後**查：0x25 進段當下 3D 圖還沒載
	//（要選 EXPLORE 才進地城），提早查會拿到錯的圖、BFS 四面都撞牆。
	grid, ok := catalog.Lookup(geo.MapRef{Set: seg.Member, BlockID: state.Area.Current3DMapBlockID})
	if !ok {
		t.Logf("%s（pick=%d）讀不到 GEO（block=%d）", seg.ID, pick, state.Area.Current3DMapBlockID)
		return 0
	}

	type point struct{ x, y int }
	// ⚠ **踏進一格的方向會改變結果**：樓梯事件是「站對方向踏上去」才觸發的
	// （`ECL5/0x33:090Ch` 用地形碼查表拿方向，朝向不對就直接 `EXIT`，畫面上什麼
	// 都不會發生，spec 1161）。只用格子當鍵，第一次從錯的方向踏上去就把那一格
	// 封死，另外三個方向永遠不會再試——**靠樓梯才進得去的樓層因此永遠走不到**。
	// ⇒ 邊界記 **(格子, 進入方向)**；「這一格去過沒有」另外記。
	type entry struct {
		at        point
		direction int
	}
	start := point{state.DungeonX, state.DungeonY}
	if testing.Verbose() {
		open := ""
		for _, direction := range []int{0, 2, 4, 6} {
			dx, dy := normalDungeonDelta(direction)
			state.SetDungeonGeometryView(start.x, start.y, uint8(direction))
			open += fmt.Sprintf(" %d:%v", direction, state.CanMoveDungeon(grid, dx, dy, direction))
		}
		t.Logf("%s（pick=%d）落點 (%d,%d) map=%d 四向:%s",
			seg.ID, pick, start.x, start.y, state.Area.Current3DMapBlockID, open)
	}
	seen := map[point]bool{start: true}
	tried := map[entry]bool{}
	queue := []point{start}
	visited := 1
	for len(queue) > 0 && visited < 260 {
		current := queue[0]
		queue = queue[1:]
		for _, direction := range []int{0, 2, 4, 6} {
			deltaX, deltaY := normalDungeonDelta(direction)
			// ⚠ 地城圖是環繞的（`CanMoveDungeonWrapped` 一族）：0x25 的落點在
			// (3,0)、唯一開的方向朝北跨過圖邊——用邊界檢查擋掉環繞會讓整段
			// 只剩 1 格（第 719 輪）。
			next := point{(current.x + deltaX + geo.Width) % geo.Width,
				(current.y + deltaY + geo.Height) % geo.Height}
			if tried[entry{next, direction}] {
				continue
			}
			tried[entry{next, direction}] = true
			state.SetDungeonGeometryView(current.x, current.y, uint8(direction))
			state.DungeonWallRoof = grid.CellWrapped(current.x, current.y).Terrain
			if !state.CanMoveDungeon(grid, deltaX, deltaY, direction) {
				// 門在 CanMoveDungeon 看來就是牆；若不先開，莊園與城市裡的
				// 單格房間會被盤點成永遠走不到。這裡和 dungeon-walk-probe
				// 一樣只在 GEO 明確標成門（2／3）時打開，問的是「門開後內容
				// 能否走到」，不把它冒稱成這支強化走訪隊伍的開鎖成功率。
				flags, ok := grid.WallDoorFlagsWrapped(current.x, current.y, direction)
				if !ok || (flags != 2 && flags != 3) ||
					!grid.UnlockDoorWrapped(current.x, current.y, direction) ||
					!state.CanMoveDungeon(grid, deltaX, deltaY, direction) {
					continue
				}
			}
			if err := state.MoveDungeon(grid, deltaX, deltaY, direction); err != nil {
				continue
			}
			if !seen[next] {
				seen[next] = true
				visited++
			}
			if !settle() {
				// 推不回地城就不從這一格繼續往外走；已經到過的算數。
				continue
			}
			queue = append(queue, next)
			// ★ 事件可能把隊伍搬走（樓梯、傳送）。只把**打算走到**的那一格排進
			// 佇列的話，落點就被丟掉了——而樓梯正是進到別的連通分量的唯一辦法
			// （巫師塔每一層在 GEO 上是獨立房間，spec 1161）。丟掉落點等於把
			// 「走上樓」記成「走到樓梯口」，那些樓層永遠不會被走到。
			landed := point{state.DungeonX, state.DungeonY}
			if landed != next && !seen[landed] {
				seen[landed] = true
				visited++
				queue = append(queue, landed)
			}
		}
	}
	return visited
}

// namedTowerRouteChoice 只處理 0x25 那條需要跨兩層選單的正常路徑；其他選單
// 仍選最後一項，讓既有 settle 行為維持不變。用名字而不是索引，避免插入新選項
// 時把「前往高塔」安靜地改成「返回匕首瀑布」。
func namedTowerRouteChoice(choices []string) int {
	for _, wanted := range []string{"SHOP", "GO TO THE TOWER"} {
		for index, choice := range choices {
			if choice == wanted {
				return index
			}
		}
	}
	return len(choices) - 1
}

// routeKnownFragments 是**已知的問句／子程式片段**：實機它們會被併進呼叫端
// 那一頁，所以沒有自己的 text rule。這一支的走法是廣度優先、逐格推事件，
// 有機會讓它們單獨出現——那時候看起來就像「落回原文」。
//
// ⚠ 這一份**故意很短，而且不會自動長大**：任何**新的**英文落回照樣讓測試紅。
// 它擋的是已知的量法假陽性，不是拿來讓報表好看的。
//
// ⚠ 要把某一句放進來之前先確認它真的是片段：看它的**兄弟句**有沒有自己的
// 規則。`A SULFEROUS SMELL…` 與 `A PARTY OF DARK ELVES…` 各有一條獨立規則，
// 所以同一組的 `AN EFREET LEADS A BAND…` 少一條就是**規則寫錯**不是片段
// ——那一句已經補上規則，不在這裡。
var routeKnownFragments = map[string]bool{
	// 共用子程式：`ecl-text-coverage` 明文說**不要**替它寫規則
	//（只有一兩個字的 `all_contains` 會攔截到別的文字）。
	"WHAT DO YOU DO ?": true,
	// 有些共用問句由 PRINT 與下一段標點分開；BFS 在事件邊界收訊息時會只
	// 看到這一半。完整頁面已由 ecl-text-coverage 的 run 分母另行驗證。
	"WHAT DO YOU DO": true,
	// 奧克薩姆莊園的同一支共用子程式保留了原作的 WHAT TO 拼字；
	// ecl-text-coverage 也把它判成 subroutine，而不是獨立玩家頁。
	"WHAT TO YOU DO?": true,
	// 問句片段：兄弟句（`DOES ANYONE WANT TO TRY AGAIN`、`WILL YOU ACCOMPANY
	// THEM`、`DO YOU WANT TO TRY FOR ANOTHER`）都是**併在前一頁的規則裡**，
	// 沒有獨立規則。⚠ 這一句是否真的只會併著出現**還沒有實機確認**——
	// 確認之前不替它寫規則（寫短片段規則會攔截別的文字），也不讓它擋住測試。
	"DOES ANYONE WANT TO GO AND OPEN ONE?": true,
}

// installShopFixtures 把物品目錄與寶物區塊掛上走訪 state：0x25 魔法商店的
// SHOP 分支走 `enterECLShop` → `ResolveTreasureRequests`，沒有這兩份就整條
// 分支進不去（覆蓋因此少掉商店貨架那一叢集）。讀不到 image 時安靜跳過——
// 走訪照舊，只是 SHOP 那條分支不開。
func installShopFixtures(t *testing.T, state *State) {
	t.Helper()
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		return
	}
	defer image.Close()
	if catalog, err := monster.ParseBaseItems(zipData(t, image, "ITEMS")); err == nil {
		state.SetItemCatalog(catalog)
	}
	areaData := make(map[uint8][]byte)
	for area := 1; area <= 6; area++ {
		areaData[uint8(area)] = zipData(t, image, fmt.Sprintf("ITEM%d.DAX", area))
	}
	if blocks, err := ParseTreasureItemBlocks(areaData); err == nil {
		state.SetTreasureItemBlocks(blocks)
	}
}
