package ecl

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
