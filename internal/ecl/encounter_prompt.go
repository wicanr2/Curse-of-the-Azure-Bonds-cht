package ecl

// 遭遇距離的兩格工作記憶體。
//
// ★ 位址是從 `bank1^[X] ＝ ECL 格 0x7C00 ＋ X ÷ 2` 換算來的，三個獨立的已知點
// 都對得上（`bank1^[5C4h]`＝`7EE2h` 營地、`bank1^[5C2h]`＝`7EE1h`、
// `bank1^[550h]`＝`7EA8h`，見 docs/audit/ecl-shared-cells.md）。第四個佐證是
// `ECL4/0x20` `+00B5h` 的 `SAVE 00 7EC0`——腳本自己會把距離上限歸零。
const (
	// EncounterMaxDistanceCell ＝ `bank1^[580h]`：`0Ch SETUP MONSTER` 與
	// `29h ENCOUNTER MENU` 的運算元 2 寫進來的**距離上限**。
	EncounterMaxDistanceCell = 0x7EC0
	// EncounterDistanceCell ＝ `bank1^[582h]`：**當下的距離**。原作由地圖座標
	// 算出再被上限夾住；`0Dh APPROACH` 與遭遇選單的 `ADVANCE` 會把它減一。
	EncounterDistanceCell = 0x7EC1
)

// InitEncounterDistance 依原作 `0Ch`（`overlay-02:03CAh`）與 `29h`
// （`overlay-02:2177h`）共用的那一段設定距離：先從地圖算，再被上限夾住。
//
// ⚠ remake 沒有「從地圖算距離」的模型（原作走 `overlay-07 entry#6(座標, 朝向)`），
// 所以直接取上限——夾住那一步在這個近似下永遠成立，不是被省略掉。
func InitEncounterDistance(memory map[uint16]uint16, maxDistance uint16) {
	if memory == nil {
		return
	}
	memory[EncounterMaxDistanceCell] = maxDistance
	memory[EncounterDistanceCell] = maxDistance
}

// ApproachEncounter 把當下距離減一，回傳有沒有真的減。距離已經是 0 就什麼都不做
// ——原作 `0Dh`（`overlay-02:0801h`）的第一條就是 `cmp ... 0 / jbe`。
func ApproachEncounter(memory map[uint16]uint16) bool {
	if memory == nil || memory[EncounterDistanceCell] == 0 {
		return false
	}
	memory[EncounterDistanceCell]--
	return true
}

// EncounterPromptOperand 是 `29h ENCOUNTER MENU` 第一句旁白的運算元位置。
// 三句旁白在運算元 9、10、11。
const EncounterPromptOperand = 9

// EncounterPromptSlots 回傳三句旁白的查看順序（0..2，對應運算元 9、10、11），
// 呼叫端取**第一個非空**的那一句。
//
// 原作（DOS `overlay-02:2233h`／`2273h`／`22BEh`）依距離分三支：距離 0 從第 1 句
// 開始、距離 1 從第 2 句、距離 2 從第 3 句，接著在三句之間**循環**往下找，直到
// 找到非空的或繞回起點。三句就是近／中／遠三種描述——猶拉什地下那一處是
// 「蔓生怪撲上來推你回走廊」「蔓生怪想脫身」「你看見遠處有蔓生怪」，語意與
// 距離一一對上。
//
// ⚠ 距離大於 2 時原作沒有對應分支（`2229h` 只比 0／1／2），落到那裡讀的是沒寫過
// 的暫存區。這裡一律先夾到 0..2，不重現那個未初始化行為。
func EncounterPromptSlots(distance int) [3]int {
	start := distance
	if start < 0 {
		start = 0
	}
	if start > 2 {
		start = 2
	}
	var order [3]int
	for step := 0; step < 3; step++ {
		order[step] = (start + step) % 3
	}
	return order
}
