package ecl

// EffectStatus 是「這個 opcode 的副作用在 remake 還原到什麼程度」。
//
// ★ 為什麼要有這張表。 `runtime.go` 對每個 opcode 都有 `case`，所以「有沒有
// handler」問不出東西——真正的差別在**那個 case 做了什麼**：有的完整執行，
// 有的只把運算元讀掉再往下走（`3Bh SPELL`、`3Ch PROTECTION` 就是），
// 兩者在程式碼裡長得很像，
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
// ★ 判準（三態的分界，寫死免得下次又憑印象填）：
//
//	done      ECL 看得到的效果（記憶體、比較旗標、控制流）由 VM 產生，
//	          **而且**效果若在 VM 之外（隊伍、物品、畫面、資產），
//	          有**正式路徑**的程式碼把它套用。
//	partial   兩半只做了一半。
//	consumed  兩半都沒有——含「解碼後排進 result，但取用的 API 只有測試呼叫」。
//
// ⚠ **`runtime.go` 有 `case` 不算數。** 本表第一版把 `27h TREASURE` 判成
// `consumed`（其實 `combat_state.go` 有正式消費端）、把 `2Eh DAMAGE` 判成
// `done`（其實只結算全隊封包）。兩個方向都錯過一次，判準才寫成上面這樣：
// 要同時看「VM 產生了什麼」與「誰把它套用」。
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
	0x0D: {EffectDone, "原作 0801h 逐條讀完（22 條）：距離（bank1^[582h] ＝ ECL 格 7EC1h）大於 0 就減一並用新距離重畫遭遇圖，是 0 就什麼都不做。remake 兩件都做：ApproachEncounter 減格子、ApproachCount 給上層重畫。⚠ 這個減一不影響遭遇選單——29h 進門會重新設一次距離把它蓋回去（spec 1146）"},
	0x0E: {EffectPartial, "原作 0841h 整支 77 條讀完（spec 1148）：三層分岔——0FFh 關閉、bank1^[5C2h]（＝7EE1h 頭像）非 0FFh 走頭像合成、n >= 78h 是大圖。remake 三個判準都對得上，0FFh 先前什麼都不做、現在發 PictureCloseRequested。partial 剩表現層：實際載入、頭像與身體的分塊合成、以及 4FBAh/4FBBh 那個不重繪的旁路"},
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
	0x1C: {EffectPartial, "原作 120Eh 逐條讀完（37 條）：清怪物鏈與已放置數（47E6h）、清「有怪要打」旗標（8B69h，spec 1095）、把 6F70h 起 28 bytes 的戰利品池歸零（七種貨幣／寶石／珠寶，spec 1059）、沿 6F8Ch 鏈逐節點 FreeMem(63) 釋放 27h 串進去的物品節點（spec 1087），並把 7603h 設成 8。remake 這一側全部對上——怪物鏈、跨執行累積的戰利品堆都清（docs/audit/ecl-treasure-clear.md）。partial 只剩一項：7603h 設成 8 的語意還沒解讀"},
	0x1D: {EffectDone, "隊伍強度"},
	0x1E: {EffectDone, "CHECKPARTY 六個條件"},
	0x1F: {EffectConsumed, "UNKNOWN_1F 在 corpus 靜態不可達"},
	0x20: {EffectDone, "NEWECL 終止本次執行並換 block（spec 1104）"},
	0x21: {EffectPartial, "LOAD FILES 記下請求；實際換檔由上層做"},
	0x22: {EffectDone, "隊伍突襲判定"},
	0x23: {EffectConsumed, "SURPRISE 在 corpus 靜態不可達；原作結果碼 3 也寫不出去（spec 1087）"},
	0x24: {EffectPartial, "COMBAT 是三選一的服務分派點（spec 1095）。分派順序已照抄原作（spec 1149：179Ah 先看 8B69h／8B56h，有怪就直接打，商店旗標排在後面）。199 處分成 153 處真的要打與 46 處走服務分派（docs/audit/ecl-combat-sites.md）；partial 指的是那 46 處在 remake 走的是別的機制，不是 24h 的請求旗標——兩個旗標都沒有 producer"},
	0x25: {EffectDone, "動態分支，目的地是字面位址（spec 1110）"},
	0x26: {EffectDone, "同上，返回位址進堆疊"},
	0x27: {EffectPartial, "原作 1B53h 整支 398 條讀完（spec 1151）：前七個運算元以 32-bit **覆寫** DS:6F70h 的戰利品池（1Ch 清的就是它）；第八個 ItemBlock 三選一——< 80h 載 ITEM<片>.DAX 那個區塊並把裡面每一筆都掛上鏈、= 0FFh 不給物品、80h..FEh 隨機產生 n − 80h 件。物品鏈 DS:6F8Ch 的 next 在 +2Ah 且是**前插**，顯示端從鏈頭走（overlay-05:0CF5h），所以清單是反序——remake 已跟上。隨機表的區間 bug 也修了（第二擲 48／49 原作回 59，remake 先前回劍）。partial 剩兩項：隨機那一路沒跑 CREATERNDTREASURE（spec 1036）所以加值／名稱三段／重量／價值／卷軸法術都是空的；以及 remake 把同一次執行的多筆 TREASURE 相加而原作是覆寫（corpus 未觀察到兩筆之間沒有 1Ch 的情形）"},
	0x28: {EffectDone, "ROB 依比例取走金錢與物品"},
	0x29: {EffectDone, "遭遇選單"},
	0x2A: {EffectDone, "GETTABLE 讀表（手札編號就是靠它，spec 1108）"},
	0x2B: {EffectDone, "水平選單"},
	0x2C: {EffectDone, "PARLAY"},
	0x2D: {EffectPartial, "原作 2F02h 整支 124 條讀完（spec 1150）：operand − 7FFFh 之後七路分派。corpus 用到四個——2E10h 125 處（重畫）、B200h 19 處（音效）、C01Eh 13 處（MOVEFORWARD）、6803h 11 處（圖片序列推一格）。B200h 的兩支已判定：選號由 ECL 格 03DE 決定，全 corpus 15 次寫入一律是 5 ⇒ 只走得到 10 那一支；6803h 已接成序列游標。partial 只剩 2E10h：原作是 STOREVALUE 當場寫 720Fh/7210h/7211h 並立五個髒旗標之一，CALL 只負責「髒了才重畫」，remake 改用 CALL 當下回頭掃 SaveWrites 的啟發式"},
	0x2E: {EffectPartial, "只有明確指定全隊的封包（旗標 0xC0）會在正式路徑結算；其餘要選定角色的形式仍留在 pending"},
	0x2F: {EffectDone, "位元運算"},
	0x30: {EffectDone, "位元運算"},
	0x31: {EffectDone, "關掉畫面上的怪物圖示；同一次執行又要求新畫面時以新的為準（原作也是先關再畫）"},
	0x32: {EffectDone, "FIND ITEM"},
	0x33: {EffectPartial, "原作 2CEAh 整支 14 條讀完（spec 1147）：欄 65A0h := 1、列 65A1h ＋1，兩個分支對游標做的事一樣（8B61h 只決定要不要順手清）。所以它是硬換行——連續兩條會空一行。remake 只記指令邊界（PrintReturnCount），沒有游標模型；缺口在 UI 的行模型，不是 ECL VM"},
	0x34: {EffectDone, "ECL CLOCK"},
	0x35: {EffectDone, "SAVE TABLE"},
	0x36: {EffectDone, "ADD NPC 建出隊員"},
	0x37: {EffectPartial, "LOAD PIECES 記下請求；資產載入由上層做"},
	0x38: {EffectPartial, "PROGRAM 記下 ID；PROGRAM 8 的通關序列（spec 1087）尚未接完"},
	0x39: {EffectDone, "WHO 選人"},
	0x3A: {EffectDone, "DELAY"},
	0x3B: {EffectDone, "依行軍順序找持有者，slot 與隊員索引寫回兩個位址，找不到寫 0FFh；依據是呼叫端（COMPARE FFh ＋ LOAD CHARACTER），不是命令表"},
	0x3C: {EffectDone, "原作是翻手冊的防拷問答（橋上謎題，ECL1 0x50 +1B6Fh）。呼叫端拿到控制權後直接印「YOU MAY PASS.」再 RETURN，沒有分支 ⇒ remake 不附那份對照表，一律通過就是正確行為，不是缺口"},
	0x3D: {EffectDone, "清空文字框且不印新文字；只有在該次執行沒有新文字時才看得出來，遊戲層據此清掉 Message"},
	0x3E: {EffectDone, "DUMP"},
	0x3F: {EffectDone, "FIND SPECIAL"},
	0x40: {EffectDone, "DESTROY ITEMS"},
}
