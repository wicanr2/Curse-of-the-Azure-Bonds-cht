package combat

// 佈陣（spec 1200）：PC-98 COMPREP（overlay-10）`17EEh`／`1364h`／`1220h` 的
// 忠實轉錄。原作把兩隊各自的佈署區塊展開成 occupancy（6 列 × 11 欄 × 4 組），
// 以「理想格」為中心做狀態機掃描，通過 occupancy 與地面檢查才寫座標。
//
// ⚠ 這一支是**轉錄不是重寫**：`place` 的控制流照 `1364h` 的跳躍結構走，
// 包含看起來不直覺的分支（部分出界走失敗計數、完全出界走步數上限）。
// 想「簡化」之前先回 spec 1200 的 asm 對照。

// deploymentShapes 是 DSEG `702h`：5 種形狀 × 6 列 × [lo, hi]（含端點；
// lo > hi ＝ 該列沒有格）。形狀 0..3 依隊的方向組選，組 1 一律用形狀 4。
var deploymentShapes = [5][6][2]int{
	{{1, 0}, {1, 0}, {1, 0}, {2, 9}, {3, 10}, {4, 10}},  // 北
	{{0, 2}, {0, 3}, {1, 4}, {2, 5}, {3, 6}, {4, 7}},    // 東
	{{0, 6}, {0, 7}, {1, 8}, {1, 0}, {1, 0}, {1, 0}},    // 南
	{{3, 6}, {4, 7}, {5, 8}, {6, 9}, {7, 10}, {8, 10}},  // 西
	{{0, 6}, {0, 7}, {1, 8}, {2, 9}, {3, 10}, {4, 10}},  // 組 1 專用
}

// 掃描用的四張小表（DSEG `6CEh`／`6DEh`／`6EEh`／`6F2h`／`6FAh`，逐 byte 轉錄）。
var (
	deployWallDirs  = [4][4]int{{8, 4, 6, 2}, {8, 6, 4, 0}, {8, 0, 6, 2}, {8, 2, 0, 4}}
	deployScanDirs  = [4][4]int{{0, 0, 2, 6}, {2, 2, 0, 4}, {4, 4, 2, 6}, {6, 6, 4, 0}}
	deployAxisRemap = [4]int{7, 2, 3, 6}
	deploySeedCol   = [2][4]int{{5, 4, 5, 6}, {3, 8, 7, 2}}
	deploySeedRow   = [2][4]int{{3, 2, 2, 3}, {0, 2, 5, 3}}
)

// deployDX／deployDY 與 DirectionDelta 同表（DSEG `489Eh`／`48A7h`）。
var (
	deployDX = [8]int{0, 1, 1, 1, 0, -1, -1, -1}
	deployDY = [8]int{-1, -1, 0, 1, 1, 1, 0, -1}
)

// GroundCheck 回答「這個戰場格站不站得上去」。原作是 TACMAP entry#19 的
// 障礙判定（spec 1119）＋ `48ACh` 物件表；remake 由呼叫端接戰場地形。
// nil 一律當可站。
type GroundCheck func(x, y int) bool

// WallCheck 回答「隊伍地城格朝這個方向有沒有被牆擋住」——原作 `378h`
// 「牆面從兩側各查一次」，回 true ＝ 走得過。nil ＝ 原作 `7F27h = 3`
// （非地城）那條路：一律當被擋（原作跳過查詢直接立旗標）。
type WallCheck func(direction int) bool

// Deployment 對應原作佈署期間的全域（`7ABAh..7AC8h` 與 `78AAh`）。
type Deployment struct {
	teamX    [2]int // 7ABAh：隊 0 ＝ (0,0)、隊 1 ＝ 距離 × 方向 delta
	teamY    [2]int // 7ABCh
	attempts [2]int // 7ABEh：(隊伍人數 + 1) div 2
	dirGroup [2]int // 7AC0h：地圖朝向 div 2；敵隊取對向
	// occupied 對應 `78AAh`：true ＝ 已占用或不在模板內。
	occupied [2][4][6][11]bool
	// DungeonX/Y 是隊伍的地城座標（`A2A9h`／`A2AAh`），牆面調整用。
	DungeonX, DungeonY int
}

// NewDeployment 對應 `17EEh` 的起點計算與 occupancy 展開。
// mapDirection 是地圖朝向（0..7，`A2ABh`）；distance 是遭遇距離
// （SETUP MONSTER 寫進 `player^[582h]` 的值，spec 1146）。
func NewDeployment(mapDirection uint8, distance, partySize, enemySize int) *Deployment {
	direction := int(mapDirection % 8)
	d := &Deployment{}
	d.teamX[0], d.teamY[0] = 0, 0
	d.dirGroup[0] = direction / 2
	d.teamX[1] = distance * deployDX[direction]
	d.teamY[1] = distance * deployDY[direction]
	d.dirGroup[1] = ((direction + 4) % 8) / 2
	d.attempts[0] = (partySize + 1) / 2
	d.attempts[1] = (enemySize + 1) / 2
	for team := 0; team < 2; team++ {
		for group := 0; group < 4; group++ {
			shape := d.dirGroup[team]
			if group == 1 {
				shape = 4
			}
			for row := 0; row < 6; row++ {
				lo, hi := deploymentShapes[shape][row][0], deploymentShapes[shape][row][1]
				for col := 0; col < 11; col++ {
					d.occupied[team][group][row][col] = col < lo || col > hi
				}
			}
		}
	}
	return d
}

// tryPlace 對應 `1220h`：occupancy → 座標公式 → 地面檢查 → 占格。
// 回傳 (x, y, ok)。
func (d *Deployment) tryPlace(team, group, teamY, teamX, row, col int, ground GroundCheck) (int, int, bool) {
	if row < 0 || row > 5 || col < 0 || col > 10 {
		return 0, 0, false
	}
	if d.occupied[team][group][row][col] {
		return 0, 0, false
	}
	x := teamY*5 + teamX*6 + col + 22
	y := teamY*5 + row + 10
	if ground != nil && !ground(x, y) {
		return 0, 0, false
	}
	d.occupied[team][group][row][col] = true
	return x, y, true
}

// Place 對應 `1364h`：為一名戰鬥員找空位。回傳 (x, y, ok)；ok 為假
// 表示四個組都放不下——原作在這種情況把戰鬥員**從戰鬥移除**。
//
// walls 只在「隊 0、方向組為偶數」的第一次失敗與組升級時被查詢；
// nil 對應原作的非地城模式（跳過查詢、直接視為被擋）。
func (d *Deployment) Place(team int, ground GroundCheck, walls WallCheck) (int, int, bool) {
	if team < 0 || team > 1 {
		return 0, 0, false
	}
	// 區域變數對照：firstPass=var_3、giveUp=var_4、state=var_7、fails=var_F、
	// group=var_14、tx/ty=var_11/var_12、seedCol/Row=var_15/var_16、
	// col/row=var_17/var_18、ring=var_10、steps=var_13、outOfBlock=var_5、
	// placed=var_2。
	firstPass, giveUp := true, false
	state := 1
	fails, group := 0, 0
	tx, ty := d.teamX[team], d.teamY[team]
	var seedCol, seedRow, col, row, ring, steps int
	placed := false
	wallOpen := func(direction int) bool {
		if walls == nil {
			return false // 7F27h = 3：跳過查詢，視為被擋
		}
		return walls(direction)
	}
	// ⚠ 迴圈上限是 harness 的保險，不是原作語意：原作靠「步數上限 → 換組 →
	// 四組用盡」收斂。掃描空間是 4 組 × 6×11 格加上重設，遠小於這個數。
	for iteration := 0; iteration < 4096; iteration++ {
		// loop_1398
		scan := deployScanDirs[d.dirGroup[team]][group] / 2
		switch state {
		case 1:
			axis := deployAxisRemap[(scan+2)%4]
			k := 0
			if group > 0 {
				k = 1
			}
			seedCol = deploySeedCol[k][scan] + deployDX[axis]*fails
			seedRow = deploySeedRow[k][scan] + deployDY[axis]*fails
			col, row = seedCol, seedRow
			ring, state, steps = 1, 2, 1
		case 2:
			axis := deployAxisRemap[(scan+1)%4]
			col = seedCol + deployDX[axis]*ring
			row = seedRow + deployDY[axis]*ring
			state = 3
			steps++
		case 3:
			axis := deployAxisRemap[(scan+3)%4]
			col = seedCol + deployDX[axis]*ring
			row = seedRow + deployDY[axis]*ring
			state = 2
			ring++
			steps++
		}
		// 1587：候選是否在佈署區塊內。
		outOfBlock := col < 0 || row < 0 || col > 10 || row > 5
		// 15A9：state 還是 1（剛被牆面調整重設）就直接跳到組升級判斷。
		if state > 1 {
			if outOfBlock {
				fullyOut := (col < 0 || col >= 11) && (row < 0 || row >= 6)
				// 15F6：失敗計數（可能觸發牆面調整）＋ 重設掃描。
				fail := func() {
					fails++
					if team == 0 && d.dirGroup[team]%2 == 0 && group == 0 && fails == 1 {
						// 161F：第一次失敗時查隊伍地城格四周的牆，
						// 有任何一個方向被擋就再多推一步。
						blocked := false
						for k := 1; k <= 3; k++ {
							if !wallOpen(deployWallDirs[d.dirGroup[team]][k]) {
								blocked = true
							}
						}
						if blocked {
							fails++
						}
					}
					state, firstPass = 1, false
				}
				if !fullyOut {
					// 部分出界 → 直接失敗計數。
					fail()
				} else {
					// 15C8：完全出界 → 步數超過上限才算失敗
					// （首組上限是 attempts、其後 11）。
					if firstPass {
						if steps >= d.attempts[team] {
							fail()
						}
					} else if steps > 11 {
						fail()
					}
				}
			}
		}
		// 16AF：完全出界時嘗試升級到下一個組（把起點往開著的方向移）。
		if outOfBlock {
			fullyOut := (col < 0 || col >= 11) && (row < 0 || row >= 6)
			if fullyOut {
				placed = false
				state = 0
				for group < 3 && state != 1 {
					group++
					dir := deployWallDirs[d.dirGroup[team]][group]
					if !wallOpen(dir) {
						tx = d.teamX[team] + deployDX[dir]
						ty = d.teamY[team] + deployDY[dir]
						fails = 0
						state = 1
					}
				}
				if state != 1 {
					giveUp = true
				}
			}
		}
		// 17A1：界內候選才真的試放。
		if !outOfBlock {
			var x, y int
			x, y, placed = d.tryPlace(team, group, ty, tx, row, col, ground)
			if placed {
				return x, y, true
			}
		}
		if giveUp {
			return 0, 0, false
		}
	}
	return 0, 0, false
}
