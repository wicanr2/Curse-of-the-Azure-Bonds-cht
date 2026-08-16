package ecl

// EffectStatus 是「這個 opcode 的副作用在 remake 還原到什麼程度」。
//
// ★ 為什麼要有這張表。 `runtime.go` 對每個 opcode 都有 `case`，所以「有沒有
// handler」問不出東西——真正的差別在**那個 case 做了什麼**：有的完整執行，
// 有的只把運算元讀掉再往下走（`0x27 TREASURE` 就是），兩者在程式碼裡長得很像，
// 在測試裡也一樣綠。表把差別寫死，`cmd/ecl-effect-coverage` 再乘上 corpus 的
// 出現次數，才知道「副作用還差多少」是多大的數字。
type EffectStatus string

const (
	// EffectDone：副作用在 remake 已經產生，且有回歸測試或實機路徑驗過。
	EffectDone EffectStatus = "done"
	// EffectPartial：做了一部分——通常是把請求記進 result 讓上層處理，
	// 或只還原了原作行為的一支。
	EffectPartial EffectStatus = "partial"
	// EffectConsumed：只把運算元讀掉、繼續往下走。**不是 no-op 的委婉說法**：
	// 它明確表示「這條指令在原作有效果，remake 目前沒有」。
	EffectConsumed EffectStatus = "consumed"
)

// OpcodeEffect 描述一個 opcode 的副作用還原狀態。
type OpcodeEffect struct {
	Status EffectStatus
	// Note 寫「還差什麼」或「憑什麼說做完了」，不寫這個 opcode 是什麼意思
	// ——那在 KnownCommands 裡。
	Note string
}

// OpcodeEffects 逐個 opcode 的還原狀態。
//
// ⚠ 新增 opcode 沒有登記會讓 `cmd/ecl-effect-coverage` 的 fail-closed 檢查變紅，
// 不會被靜靜略過。
var OpcodeEffects = map[byte]OpcodeEffect{
	0x00: {EffectDone, "結束本次執行，PC 與堆疊由 lifecycle 驅動器接手"},
	0x01: {EffectDone, "控制流"},
	0x02: {EffectDone, "控制流，返回位址進堆疊"},
	0x03: {EffectDone, "六個比較旗標，供 16h..1Bh 使用"},
	0x04: {EffectDone, "算術"},
	0x05: {EffectDone, "算術"},
	0x06: {EffectDone, "算術"},
	0x07: {EffectDone, "算術"},
	0x08: {EffectDone, "RANDOM 走 spec 1103 的 ROLLDICE 參數約定"},
	0x09: {EffectDone, "數值與字串記憶體都寫得進去"},
	0x0A: {EffectDone, "角色欄位投影，含 7CE4h 的 and 1 遮罩（ENG-13）"},
	0x0B: {EffectDone, "怪物載入請求交給戰鬥層"},
	0x0C: {EffectDone, "怪物編成請求交給戰鬥層"},
	0x0D: {EffectPartial, "APPROACH 只記請求；原作的接近動畫與距離狀態未還原"},
	0x0E: {EffectPartial, "PICTURE 記下 block 與 big-picture 旗標；頭像／身體分塊仍靠上層"},
	0x0F: {EffectConsumed, "INPUT NUMBER 在 corpus 靜態不可達，沒有輸入通道"},
	0x10: {EffectDone, "字串輸入，退格以 rune 為單位（CHT-02）"},
	0x11: {EffectDone, "文字累積進 result.Text"},
	0x12: {EffectDone, "清框並開新頁（spec 1104）"},
	0x13: {EffectDone, "控制流"},
	0x14: {EffectDone, "四運算元比較"},
	0x15: {EffectDone, "垂直選單，選項字串與選擇結果都回得去"},
	0x16: {EffectDone, "條件不成立時跳過下一條整條指令（spec 1106）"},
	0x17: {EffectDone, "同上"},
	0x18: {EffectDone, "同上"},
	0x19: {EffectDone, "同上"},
	0x1A: {EffectDone, "同上"},
	0x1B: {EffectDone, "同上"},
	0x1C: {EffectPartial, "CLEARMONSTERS 清怪物鏈與已放置數；原作另清的四塊 bank1 區域未逐格對上"},
	0x1D: {EffectDone, "隊伍強度"},
	0x1E: {EffectDone, "CHECKPARTY 六個條件"},
	0x1F: {EffectConsumed, "UNKNOWN_1F 在 corpus 靜態不可達"},
	0x20: {EffectDone, "NEWECL 終止本次執行並換 block（spec 1104）"},
	0x21: {EffectPartial, "LOAD FILES 記下請求；實際換檔由上層做"},
	0x22: {EffectDone, "隊伍突襲判定"},
	0x23: {EffectConsumed, "SURPRISE 在 corpus 靜態不可達；原作結果碼 3 也寫不出去（spec 1087）"},
	0x24: {EffectPartial, "COMBAT 是三選一的服務分派點（spec 1095）；戰鬥本身的回合生命週期仍是 RE-06"},
	0x25: {EffectDone, "動態分支，目的地是字面位址（spec 1110）"},
	0x26: {EffectDone, "同上，返回位址進堆疊"},
	0x27: {EffectConsumed, "TREASURE：只讀掉八個運算元。寶物產生、隨機表與拾取流程都還沒接"},
	0x28: {EffectDone, "ROB 依比例取走金錢與物品"},
	0x29: {EffectDone, "遭遇選單"},
	0x2A: {EffectDone, "GETTABLE 讀表（手札編號就是靠它，spec 1108）"},
	0x2B: {EffectDone, "水平選單"},
	0x2C: {EffectDone, "PARLAY"},
	0x2D: {EffectPartial, "CALL 是七路 switch；corpus 只用 2E10h 與 6803h，兩者的 consumer 尚未逐條驗（RE-03）"},
	0x2E: {EffectDone, "DAMAGE"},
	0x2F: {EffectDone, "位元運算"},
	0x30: {EffectDone, "位元運算"},
	0x31: {EffectConsumed, "SPRITE OFF：戰鬥圖示的顯示狀態未還原"},
	0x32: {EffectDone, "FIND ITEM"},
	0x33: {EffectPartial, "PRINT RETURN 目前等同換行；原作的游標行為未逐格對上"},
	0x34: {EffectDone, "ECL CLOCK"},
	0x35: {EffectDone, "SAVE TABLE"},
	0x36: {EffectDone, "ADD NPC 建出隊員"},
	0x37: {EffectPartial, "LOAD PIECES 記下請求；資產載入由上層做"},
	0x38: {EffectPartial, "PROGRAM 記下 ID；PROGRAM 8 的通關序列（spec 1087）尚未接完"},
	0x39: {EffectDone, "WHO 選人"},
	0x3A: {EffectDone, "DELAY"},
	0x3B: {EffectConsumed, "SPELL：ECL 觸發的法術效果沒有接（ENG-09）"},
	0x3C: {EffectConsumed, "PROTECTION：防護狀態沒有接"},
	0x3D: {EffectConsumed, "CLEAR BOX：文字框清除的畫面行為沒有接"},
	0x3E: {EffectDone, "DUMP"},
	0x3F: {EffectDone, "FIND SPECIAL"},
	0x40: {EffectDone, "DESTROY ITEMS"},
}
