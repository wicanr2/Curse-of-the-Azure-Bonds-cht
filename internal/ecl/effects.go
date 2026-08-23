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
	0x0E: {EffectPartial, "原作 0841h 整支 77 條讀完（spec 1148）：三層分岔——0FFh 關閉、bank1^[5C2h]（＝7EE1h 頭像）非 0FFh 走頭像合成、n >= 78h 是大圖。remake 三個判準都對得上，0FFh 先前什麼都不做、現在發 PictureCloseRequested。partial 剩表現層：實際載入、頭像與身體的分塊合成、以及 4FBAh/4FBBh 那個不重繪的旁路（旁路的語意已解出：4FBAh 是現在的畫面模式、4FBBh 是上一次的，3 ＝ 非第一人稱、4 ＝ 第一人稱，因為兩者都由 ECL 格 4BE6 的新值 0／非 0 決定，而 4BE6 就是第一人稱模式旗標 ⇒「前後都還在第一人稱就不重繪」；⚠ 旁路連 8B62h/8B65h 的清除一起跳過，所以走旁路時「圖還開著」這個狀態留著。remake 還沒建這個模型）"},
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
	0x1C: {EffectDone, "原作 120Eh 逐條讀完（37 條）：清怪物鏈與已放置數（47E6h）、清「有怪要打」旗標（8B69h，spec 1095）、把 6F70h 起 28 bytes 的戰利品池歸零（七種貨幣／寶石／珠寶，spec 1059）、沿 6F8Ch 鏈逐節點 FreeMem(63) 釋放 27h 串進去的物品節點（spec 1087），並把 7603h 設成 8。remake 這一側全部對上——怪物鏈、跨執行累積的戰利品堆都清（docs/audit/ecl-treasure-clear.md）。7603h 設成 8 的語意也讀出來了（spec 1173）：它是**下一批怪物的圖示槽**（`+143h` ＝ `ICONINDEX`，`overlay-33 entry#5` 拿它與 `CPIC` 組檔名），`inc` 在放置迴圈**結束後**才做一次 ⇒ 同一批共用一張圖，而 0..7 是隊伍成員的圖示，所以重設值是 8 不是 0。它是**快取索引不是遊戲狀態**——remake 的每個戰鬥員自己帶 SpriteBlock，可觀察行為相同，因此沒有對應動作、也不該有"},
	0x1D: {EffectDone, "隊伍強度"},
	0x1E: {EffectDone, "CHECKPARTY 六個條件"},
	0x1F: {EffectConsumed, "UNKNOWN_1F 在 corpus 靜態不可達"},
	0x20: {EffectDone, "NEWECL 終止本次執行並換 block（spec 1104）"},
	0x21: {EffectDone, "原作 0C15h 與 37h 共用一支（spec 1087）：載 3D 地圖那一路 handler **自己寫兩格 ECL 記憶體**（spec 1181）——`bank0^[18Ah] := o[1]`（格 `4BC5h`）與 `bank1^[592h] := 0`（格 `7EC9h`），條件是 `o[1] <> 0FFh`、`<> 7Fh` 且 `bank0^[1CCh] <> 0`（格 `4BE6h`，22 處寫入全在 ECL1 的三個世界地圖 block，形狀是「現在是不是第一人稱畫面」）。★ `7EC9h` **腳本讀得到**：全 corpus 34 處存取，腳本一路寫 `FFh`，唯一的讀取端 `ECL2/0x03:00F6h` 是 `COMPARE 7EC9 FF`，而清掉它的只有這個 handler ⇒ 不寫的話腳本會永遠停在那個分支。remake 現在照條件寫那兩格（都走 `recordStore`，會進 `SaveWrites`），並帶出 `LoadFilesLoaded3DMap` 讓上層知道原作走的是哪一路；實際換檔仍由上層做，那是資產層不是副作用"},
	0x22: {EffectDone, "隊伍突襲判定"},
	0x23: {EffectConsumed, "SURPRISE 在 corpus 靜態不可達；原作結果碼 3 也寫不出去（spec 1087）"},
	0x24: {EffectDone, "COMBAT 是三選一的服務分派點（spec 1095）。分派順序已照抄原作（spec 1149：179Ah 先看 8B69h／8B56h，有怪就直接打，商店旗標排在後面）。199 處分成 153 處真的要打與 46 處走服務分派（docs/audit/ecl-combat-sites.md）；★★ **第四支已經解出來並接上**（spec 1182）：`sub_184B` 讀完是「`bank1^[5C4h] = 1` ⇒ 神殿（overlay-04 ＝ TEMPLE），否則 ⇒ 戰後處理（overlay-05 ＝ DOPOSTCOMBAT）」——那正是那 46 處「沒擺過怪的 24h」的去向，腳本先用 `27h TREASURE` 堆好戰利品再用 `24h` 開分配畫面。remake 因此把零隻怪的 `24h` 改成發 `PostCombatRequested` 而不是 `CombatRequested`（原作的 `24h` 沒有「零隻怪的戰鬥」，有怪才走 `sub_1956`）。★★★ 四支全部接上而且都有實機路徑：打（怪物鏈非空 ＝ 原作的 `8B69h`，spec 1145）、商店、神殿（後兩者的 producer 是腳本自己，`7F6Ch` 9 處／`7EE2h` 4 處，已用原版資料跑過）、戰後處理。★★ 第一支的另一個條件 `DS:8B56h` ＝ **決鬥旗標**，整個執行檔只有 `GODUEL`（`overlay-07 entry#25`）設它，而 `GODUEL` 只有 `2Dh CALL` 的 `8000h`／`8001h` 到得了——**corpus 兩個都沒用到**（spec 1150），所以那一支在 CoAB 走不到，remake 用怪物鏈非空覆蓋了整個可達條件"},
	0x25: {EffectDone, "動態分支，目的地是字面位址（spec 1110）"},
	0x26: {EffectDone, "同上，返回位址進堆疊"},
	0x27: {EffectDone, "原作 1B53h 整支 398 條讀完（spec 1151）：前七個運算元以 32-bit **覆寫** DS:6F70h 的戰利品池（1Ch 清的就是它）；第八個 ItemBlock 三選一——< 80h 載 ITEM<片>.DAX 那個區塊並把裡面每一筆都掛上鏈、= 0FFh 不給物品、80h..FEh 隨機產生 n − 80h 件。物品鏈 DS:6F8Ch 的 next 在 +2Ah 且是**前插**，顯示端從鏈頭走（overlay-05:0CF5h），所以清單是反序——remake 已跟上。隨機表的區間 bug 也修了（第二擲 48／49 原作回 59，remake 先前回劍）。隨機那一路已接上 CREATERNDTREASURE（spec 1036）——加值（1d20：1..14 給 ＋1、15..20 給 ＋2）、名稱三段索引、重量、價值、卷軸法術與特殊物品範本都照原作補齊，擲骰順序也對上。★ 多筆 TREASURE 的模型也對上了：**錢幣池覆寫、物品鏈累積**——原作那兩側是兩套資料結構，remake 先前把錢幣也當成累加（spec 1151 增補）"},
	0x28: {EffectDone, "ROB 依比例取走金錢與物品"},
	0x29: {EffectDone, "遭遇選單"},
	0x2A: {EffectDone, "GETTABLE 讀表（手札編號就是靠它，spec 1108）"},
	0x2B: {EffectDone, "水平選單"},
	0x2C: {EffectDone, "PARLAY"},
	0x2D: {EffectPartial, "原作 2F02h 整支 124 條讀完（spec 1150）：operand − 7FFFh 之後七路分派。corpus 用到四個——2E10h 125 處（重畫）、B200h 19 處（音效）、C01Eh 13 處（MOVEFORWARD）、6803h 11 處（圖片序列推一格）。B200h 的兩支已判定：選號由 ECL 格 03DE 決定，全 corpus 15 次寫入一律是 5 ⇒ 只走得到 10 那一支；6803h 已接成序列游標。2E10h 的髒旗標模型已經建出來（spec 1172）：STOREVALUE 當場把 C04B/C04C/C04D 鏡射到 720Fh/7210h/7211h 並立 8B68h，CALL C01Eh 當場走一步，CALL 2E10h 取快照後把五個旗標清掉——先前那個「回頭掃同一 block、執行序在前的 SaveWrites」的啟發式整支拿掉了。★ 那個「還要有新寫的 C04D」的附加條件也早就拿掉（spec 1158）：只寫 C04B/C04C 的重畫就是真實移動、朝向刻意不變，corpus 23 處全部生效——4BF0/4BF1 的 producer 是地城主迴圈的移動前快照（spec 1155/1045）。★ 七個選擇子已全量普查（`TestDuelCallSelectorsAreUnusedWhileTheOthersAreNot`，先做正對照才讓零算數）：corpus 用四個（125／19／13／11），`C018h`／`8000h`／`8001h` 三個都是 0，而且沒有第八個 ⇒ **未接的目標都走不到**。★ 投影的時機也對上了（spec 1172）：原作的 STOREVALUE 一寫 C04B 就當場改 720Fh，隊伍在那一刻就搬了，2E10h 只是重畫；remake 改成「執行結束時座標還髒著就收尾投影一次」，所以「寫了座標卻沒有 2E10h」的執行也會搬隊伍（實例 ECL3/0x10:0C7Eh 指揮官帶你走側門，隔壁那條搶劫結局 0D24h 才有重畫）。Written 遮罩因此整格移除。partial 只剩**一個 remake 這一側的擋板**：鏡射多記「這三格是哪個 block 的腳本寫的」並要求與目前 block 相同。★ 它擋的不是引擎行為——spec 1183 普查過 720Fh/7210h/7211h 的全部寫入者（STOREVALUE 的鏡射、MOVEFORWARD、INIT、MOVEMENT、LOADSAVE 的 5 byte 區塊，共五處），INTERPET（擁有 GOECL 與 block 載入）一次都沒寫卻讀了 23 次 ⇒ 原作沒有「換 block 的引擎進場放置」，進新地圖的落點是腳本自己寫的（ECL2/0x03:00DBh 甚至還帶條件 4C2A==1）。這一格擋的是 remake 的資料缺口：腳本沒寫進場座標的地圖靠 game pack 宣告的 spawn 補值。所以拿不拿得掉是資料問題，而已知有一處對不上——提爾弗頓下水道腳本說 (0,0) 朝南、宣告的 spawn 是 (0,1)，還沒用原版判過誰對"},
	0x2E: {EffectDone, "原作 2942h 整支 305 條讀完（spec 1152）：★ 旗標 bit 7 清空時**整個 byte 是次數**——連打 N 下，每下隨機挑一名隊員、用第五個運算元當攻擊值擲 TRYTOHIT，而且每下之間重擲傷害；bit 7 設定才是旗標（bit 6 全隊、bit 5 不擲豁免、bit 4 豁免成功仍吃**全額**、bit 0..4 豁免調整值——bit 4 與調整值欄位重疊，但 corpus 24 處低 5 位全 0）。目標三選一：全隊／目前角色（第五個運算元 bit 7）／隨機一名。★ 目前角色那一路傳給 MAKESAVE 的種類要**減一**且 0 代表不擲豁免，另外兩路不減一、由 bit 5 決定；兩種讀法在 corpus 上結果相同。正式路徑先前只結算全隊那 14 處，現在三種形式都結算（目標由封包自己帶著發出當下選的角色，否則整隊迴圈裡的傷害會全算到同一位身上）。★ 第三條路「單體但隨機挑一名」也接上了：擲一顆 1..隊伍人數挑人，豁免種類**不減一**、要不要擲由 bit 5 決定（corpus 走不到，照 handler 寫，形狀與全隊那一路共用同一個結算器）"},
	0x2F: {EffectDone, "位元運算"},
	0x30: {EffectDone, "位元運算"},
	0x31: {EffectDone, "關掉畫面上的怪物圖示；同一次執行又要求新畫面時以新的為準（原作也是先關再畫）"},
	0x32: {EffectDone, "FIND ITEM"},
	0x33: {EffectDone, "原作 2CEAh 整支 14 條讀完（spec 1147）：欄 65A0h := 1、列 65A1h ＋1，兩個分支對游標做的事一樣（8B61h 只決定要不要順手清）。所以它是硬換行——連續兩條會空一行。★ 玩家看得到的那一半已經接完（cmd/ecl-print-return-audit，spec 1147）：走得到的 33h 共 120 條、110 個換行段落，其中 10 段連著兩條會空行，而 7 段兩側的文字落在同一個顯示頁裡；兩個語系各七則插 \\n\\n，位置照原作那兩條 33h 前後的段落切，三條回歸測試釘住，清單來自 docs/audit/ecl-print-return.json 不是手寫的 ⇒ 原作多出一處看得見的空行測試就會紅。★ VM 這一側**沒有東西可做**：`65A0h`／`65A1h` 是 DOS 資料段的游標，全 corpus 沒有任何 ECL 指令碰得到它（`cmd/ecl-address-refs -all`，正對照：同一次掃描 `4C2A` 6 處、`7EA8` 6 處、`4BE6` 多處）⇒ 腳本觀察不到游標，沒有可分歧的控制流。⚠ 這個 0 是**修好儀器之後**量的：舊版 `-all` 傳 nil 起點，跟不到只有生命週期入口才進得去的碼，同一個位址在 `-writes` 找得到、在 `-all` 是 0。"},
	0x34: {EffectDone, "ECL CLOCK"},
	0x35: {EffectDone, "SAVE TABLE"},
	0x36: {EffectDone, "ADD NPC 建出隊員"},
	0x37: {EffectDone, "與 21h 共用 overlay-02:0C15h..0DA3h 一支 131 條（spec 1087；⚠ IDA 只認到 0D4Ah／104 條，切掉的正是收尾重繪判斷）。三個運算元是三個牆面組槽位的片號，載得進去的槽走 LOADWALLSET(槽, 片)，而那一支收尾寫 [7210h+槽×4] := 片、[7212h+槽×4] := 槽；運算元是 0FFh 的槽由 handler 自己寫成 0FFFFh ⇒ 這三格就是存檔第 9..14 欄（spec 1076／1153）。remake 的資產載入本來就接上了，這一輪補上存檔那一半。★ 三支分派全部照抄了（spec 1153 增補）：`7Fh` ⇒ 只碰槽 1 且片號是 **0**（槽 2／3 連哨兵都不寫）、兩個閘門 `bank0^[1CEh]`／`[1D0h]`（ECL 格 `4BE7h`／`4BE8h`）都非零 ⇒ 只載槽 1／3 且 `0FFh` 整個跳過、否則逐槽迴圈。**分支由 VM 決定**（`WallSetAssignments`）而不是上層猜——沒列出來的槽這一次完全不動。★★ **兩槽分支有四個現場**（`ECL5/0x33`／`0x35`、`ECL6/0x40`／`0x45`）：這四個 block 在自己的 `LOAD PIECES` 之前就把兩個閘門都設成非零，所以**第二個運算元是死資料**——remake 先前把它載進槽 2 並寫進存檔；`7Fh` 那一支 corpus 一次都沒出現，照 handler 寫"},
	0x38: {EffectDone, "原作 30DDh 整支 104 條讀完（spec 1154）：四個分派值 0（重新初始化＋主選單）／3（DS:4FC7h := 1 ⇒ 宣告隊伍全滅、停掉迴圈）／8（結局）／9（CAMP），corpus 13 處四個值全用得到、沒有一處落在什麼都不做的預設值。★ PROGRAM 8 先跑 overlay-18:10FFh 的結局過場才回主選單問存檔——把字串位移與等鍵呼叫依序取出來，分段是**五頁四次等鍵**，而且第 4 頁是 8 行、其餘四頁各 4 行。remake 先前直接跳到存檔詢問，打通關一句結局都看不到；現在照原作的等鍵位置分頁播完才進選單。⚠ spec 1087 說 4FC7h 是「訓練免費」旗標是把它和 4FC8h 搞混了"},
	0x39: {EffectDone, "WHO 選人"},
	0x3A: {EffectDone, "DELAY"},
	0x3B: {EffectDone, "依行軍順序找持有者，slot 與隊員索引寫回兩個位址，找不到寫 0FFh；依據是呼叫端（COMPARE FFh ＋ LOAD CHARACTER），不是命令表"},
	0x3C: {EffectDone, "原作是翻手冊的防拷問答（橋上謎題，ECL1 0x50 +1B6Fh）。呼叫端拿到控制權後直接印「YOU MAY PASS.」再 RETURN，沒有分支 ⇒ remake 不附那份對照表，一律通過就是正確行為，不是缺口"},
	0x3D: {EffectDone, "清空文字框且不印新文字；只有在該次執行沒有新文字時才看得出來，遊戲層據此清掉 Message"},
	0x3E: {EffectDone, "DUMP"},
	0x3F: {EffectDone, "FIND SPECIAL"},
	0x40: {EffectDone, "DESTROY ITEMS"},
}
