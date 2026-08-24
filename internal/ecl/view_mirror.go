package ecl

// ViewMirror 是原作 `DS:720Fh`／`7210h`／`7211h` 那三格，加上決定「要不要重畫」
// 的五個髒旗標（spec 1150）。
//
// ★ 為什麼要有它。 `2Dh CALL 2E10h` 在原作**不是「重畫」而是「髒了才重畫」**，
// 而座標與朝向不是在那一刻才生效——`STOREVALUE` 收到 `C04B`／`C04C`／`C04D`
// 就**當場**寫這三格並順手把 `8B68h` 設 1（DOS `overlay-07:0E89h`..`0F09h`）。
// 第二個證人是 `SHOWLOCATION`（`overlay-24:2BAAh`）：狀態列組的就是
// `Str(720Fh) + ',' + Str(7210h)` 再接 `DS:2540h + 7211h × 3` 的方向名表。
//
// ⚠ 沒有這個鏡射時，remake 只能用「`CALL` 當下回頭掃同一 block、執行序在前的
// `SaveWrites`」這個啟發式來猜。兩者在**同一次執行內**幾乎等價，差別是視窗：
// 髒旗標只有重畫才會清，所以跨執行、跨 block 的座標寫入也算數，而視窗會漏掉。
type ViewMirror struct {
	// X／Y／Facing 對應 `720Fh`／`7210h`／`7211h`。Facing 存的是**折過的**
	// 0／2／4／6——原作在 `C04D` 那一路就是這樣折的。
	X      uint16
	Y      uint16
	Facing uint16
	// Known 為 false 代表這三格還沒有人寫過，呼叫端要沿用自己的位置。
	Known bool
	// Dirty 是五個旗標的位元集合。
	Dirty uint8
	// Block 是最後一次寫這三格的腳本所在的 block。
	//
	// ⚠ **原作沒有這個**（`720Fh` 是全域，誰寫都算），而且它擋的不是引擎行為：
	// spec 1183 普查過這三格的全部寫入者，`INTERPET`（block 載入）一次都沒寫，
	// 所以原作根本沒有「換 block 的進場放置」——落點是腳本自己寫的。
	// 這一格擋的是 remake 的**資料缺口**：腳本沒寫進場座標的地圖靠 game pack
	// 宣告的 spawn 補值，不比對 block 的話會被上一個 block 的座標蓋掉。
	Block uint8
	// ScreenMode／PrevScreenMode 對應 `4FBAh`／`4FBBh`（spec 1148／1150）：
	// `STOREVALUE` 發現 ECL 格 `4BE6h`（第一人稱模式旗標）的新值與舊值不同時，
	// 先 `4FBBh := 4FBAh`，再依新值是 0 或非 0 把 `4FBAh` 設成 3（非第一人稱）
	// 或 4（第一人稱）（DOS `overlay-07:0DA0h`..`0DC9h`）。
	// 0 是 BSS 開機值——兩格都是 0 時 `PICTURE 0FFh` 的旁路條件不成立，走重繪。
	ScreenMode     uint16
	PrevScreenMode uint16
	// threeDGate 是上一次寫進 `4BE6h` 的值，判「換值」用。原作比的是記憶體格
	// 本身的舊值；`Store` 在寫入之後才被呼叫、拿不到舊值，所以自己留一份。
	// 開機值 0 與原作 BSS 相同。
	threeDGate uint16
}

// 五個髒旗標，各自的生產者見 spec 1150。
const (
	// ViewDirtyPicture ＝ `8B62h`：`0Eh PICTURE` 開圖。
	ViewDirtyPicture uint8 = 1 << iota
	// ViewDirtyWindow ＝ `8B65h`：`ECL2` entry#8（PC-98 `GODRAWWINDOW`）設，
	// `31h SPRITE OFF` 會清。
	ViewDirtyWindow
	// ViewDirtyCell ＝ `8B67h`：寫 ECL 格 `C059`／`C05F`。
	ViewDirtyCell
	// ViewDirtyCoords ＝ `8B68h`：寫 ECL 格 `C04B`／`C04C`／`C04D`。
	ViewDirtyCoords
	// ViewDirtyPartyCell ＝ `8B6Ah`：寫 ECL 格 `4BFD`／`4BFE`。
	ViewDirtyPartyCell
)

// 三格座標與兩組髒旗標來源的 ECL 位址。
const (
	viewCellX       uint16 = 0xC04B
	viewCellY       uint16 = 0xC04C
	viewCellFacing  uint16 = 0xC04D
	viewCellTerrain uint16 = 0xC059
	viewCellWall    uint16 = 0xC05F
	viewCellPartyA  uint16 = 0x4BFD
	viewCellPartyB  uint16 = 0x4BFE
	// viewCellThreeD 是第一人稱模式旗標（`4BE6h`，引擎側同一格是
	// `bank0^[1CCh]`，spec 1096／1181）；`ScreenMode` 的來源。
	viewCellThreeD uint16 = 0x4BE6
)

// Store 重現 `STOREVALUE` 的鏡射那一段：**當場**寫三格並立旗標。
//
// ⚠ 這一支要接在唯一那條寫入路徑上。原作只有 `STOREVALUE` 會寫變數，所有會
// 寫變數的 opcode 都走它——「哪些 opcode 寫過座標」在原作裡不是分類問題，
// 任何一個都算（spec 1159）。
func (m *ViewMirror) Store(address, value uint16, block uint8) {
	if m == nil {
		return
	}
	switch address {
	case viewCellX, viewCellY, viewCellFacing:
		m.Block = block
	}
	switch address {
	case viewCellX:
		m.X = value
		m.Known = true
		m.Dirty |= ViewDirtyCoords
	case viewCellY:
		m.Y = value
		m.Known = true
		m.Dirty |= ViewDirtyCoords
	case viewCellFacing:
		// 原作把 `C04D` 折成 0／2／4／6 再寫 `7211h`。
		m.Facing = (value & 3) * 2
		m.Known = true
		m.Dirty |= ViewDirtyCoords
	case viewCellTerrain, viewCellWall:
		m.Dirty |= ViewDirtyCell
	case viewCellPartyA, viewCellPartyB:
		m.Dirty |= ViewDirtyPartyCell
	case viewCellThreeD:
		// `4FBAh`／`4FBBh` 只在**換值**時輪替（`overlay-07:0DA0h`）：
		// 同值重寫不動它們——ECL1 的世界地圖 block 會反覆寫 0。
		if value != m.threeDGate {
			m.PrevScreenMode = m.ScreenMode
			if value == 0 {
				m.ScreenMode = 3
			} else {
				m.ScreenMode = 4
			}
			m.threeDGate = value
		}
	}
}

// Adopt 把引擎自己搬隊伍的結果同步進三格，並把座標那一條旗標**清掉**。
//
// ★ 原作那一側寫的是 `720Fh` 這幾個全域本身，不經過 `STOREVALUE`，所以不會
// 立旗標；而地城主迴圈走完一步就重畫，重畫又把旗標清掉。remake 的
// `SetMemoryValue` 就是那條路：正常移動、載入存檔、進地城、傳送落點都走它。
//
// ⚠ 不清旗標會留下一個看不見的地雷：腳本寫過座標但那一次執行沒有 `2E10h`，
// 旗標就一直立著；玩家接著自己走幾步，下一次任何重畫都會把「腳本指定過位置」
// 重新宣告一次，於是進場錨點被誤判成已被腳本蓋掉。
func (m *ViewMirror) Adopt(address, value uint16) {
	if m == nil {
		return
	}
	switch address {
	case viewCellX:
		m.X = value
		m.Known = true
		m.Dirty &^= ViewDirtyCoords
	case viewCellY:
		m.Y = value
		m.Known = true
		m.Dirty &^= ViewDirtyCoords
	case viewCellFacing:
		m.Facing = (value & 3) * 2
		m.Known = true
		m.Dirty &^= ViewDirtyCoords
	}
}

// StepForward 是 `2Dh CALL C01Eh`（`MOVEFORWARD`）對三格做的事。
//
// ★ 原作的 `MoveForward` **當場**改地圖暫存器，所以同一次執行裡排在後面的
// `2E10h` 看到的是走完之後的位置。remake 把 `CALL` 的副作用留到執行結束後才
// 套用，如果鏡射不在 VM 裡跟著走，中間那些 `2E10h` 的快照就會停在走之前，
// 下一次投影會把隊伍拉回起點——盜賊公會的走位動畫會少掉第一步。
//
// ⚠ 走一步**不弄髒**：原作那一支寫的是全域本身，不經過 `STOREVALUE`。
func (m *ViewMirror) StepForward() {
	if m == nil || !m.Known {
		return
	}
	switch m.Facing {
	case 0:
		m.Y = (m.Y + viewGridSize - 1) % viewGridSize
	case 2:
		m.X = (m.X + 1) % viewGridSize
	case 4:
		m.Y = (m.Y + 1) % viewGridSize
	case 6:
		m.X = (m.X + viewGridSize - 1) % viewGridSize
	}
}

// viewGridSize 是地城格子的邊長。原作的前進沒有碰撞判斷，走出邊界就繞回來。
const viewGridSize uint16 = 16

// ClearDirty 是重畫之後那五條歸零（原作在 `WALLCODE` 之前逐個清）。
func (m *ViewMirror) ClearDirty() {
	if m == nil {
		return
	}
	m.Dirty = 0
}
