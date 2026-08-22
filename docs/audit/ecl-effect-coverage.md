# ECL 副作用覆蓋（原作可達指令 → remake runtime）

由 `cmd/ecl-effect-coverage` 產生，不要手改。

- 分母是**靜態可達的指令**：`TraceGraph` 跟循序、`GOTO`／`GOSUB`／`RETURN`、`IF` 的兩條路與 `25h ON GOTO`／`26h ON GOSUB` 的每一個目的地；變長指令的長度用 `ecl.RecordEnd`（spec 1110）。同一條指令只算一次，不管玩家會跑過幾遍。
- 狀態來自 `internal/ecl.OpcodeEffects` 的人工判定，不是自動推出來的。`done` 表示副作用已產生**且**有回歸測試或實機路徑驗過；`partial` 表示只做了一部分（多半是把請求記進 result 讓上層處理）；`consumed` 表示只把運算元讀掉——**那是明確的缺口，不是 no-op**。
- 這份報告不看**參數**對不對。`24h COMBAT` 標 `partial` 說的是戰鬥服務分派點已接、回合生命週期未接（`RE-06`），不是說每一場遭遇的編成都對。
- 出現次數是**指令數**不是事件數。一個 `27h TREASURE` 出現在 12 個地方，代表 12 處寶物流程缺口，不代表 12 種不同的寶物規則。

## 摘要

| 狀態 | 指令數 | opcode 數 | 意思 |
|---|---:|---:|---|
| `done` | 13491 | 57 | 副作用已產生，且有回歸測試或實機路徑驗過 |
| `partial` | 686 | 4 | 只做了一部分，多半是把請求記進 result 讓上層處理 |
| **`consumed`** | **0** | **0** | **只讀掉運算元——原作有效果，remake 沒有** |
| 合計（靜態可達指令）| 14177 | 61 | |

## 逐 opcode

| opcode | 名稱 | 狀態 | 可達出現次數 | 出現在幾個 block | 還差什麼 |
|---|---|---|---:|---:|---|
| `0x0E` | PICTURE | `partial` | 199 | 23 | 原作 0841h 整支 77 條讀完（spec 1148）：三層分岔——0FFh 關閉、bank1^[5C2h]（＝7EE1h 頭像）非 0FFh 走頭像合成、n >= 78h 是大圖。remake 三個判準都對得上，0FFh 先前什麼都不做、現在發 PictureCloseRequested。partial 剩表現層：實際載入、頭像與身體的分塊合成、以及 4FBAh/4FBBh 那個不重繪的旁路 |
| `0x24` | COMBAT | `partial` | 199 | 24 | COMBAT 是三選一的服務分派點（spec 1095）。分派順序已照抄原作（spec 1149：179Ah 先看 8B69h／8B56h，有怪就直接打，商店旗標排在後面）。199 處分成 153 處真的要打與 46 處走服務分派（docs/audit/ecl-combat-sites.md）；★★ **第四支已經解出來並接上**（spec 1182）：`sub_184B` 讀完是「`bank1^[5C4h] = 1` ⇒ 神殿（overlay-04 ＝ TEMPLE），否則 ⇒ 戰後處理（overlay-05 ＝ DOPOSTCOMBAT）」——那正是那 46 處「沒擺過怪的 24h」的去向，腳本先用 `27h TREASURE` 堆好戰利品再用 `24h` 開分配畫面。remake 因此把零隻怪的 `24h` 改成發 `PostCombatRequested` 而不是 `CombatRequested`（原作的 `24h` 沒有「零隻怪的戰鬥」，有怪才走 `sub_1956`）。⚠ 仍是 `partial`：兩個「有怪要打」旗標 `8B69h`／`8B56h` 在 remake 沒有 producer（用怪物鏈非空代替），而神殿那一支的 `7EE2h` 也沒有 producer |
| `0x2D` | CALL | `partial` | 168 | 22 | 原作 2F02h 整支 124 條讀完（spec 1150）：operand − 7FFFh 之後七路分派。corpus 用到四個——2E10h 125 處（重畫）、B200h 19 處（音效）、C01Eh 13 處（MOVEFORWARD）、6803h 11 處（圖片序列推一格）。B200h 的兩支已判定：選號由 ECL 格 03DE 決定，全 corpus 15 次寫入一律是 5 ⇒ 只走得到 10 那一支；6803h 已接成序列游標。2E10h 的髒旗標模型已經建出來（spec 1172）：STOREVALUE 當場把 C04B/C04C/C04D 鏡射到 720Fh/7210h/7211h 並立 8B68h，CALL C01Eh 當場走一步，CALL 2E10h 取快照後把五個旗標清掉——先前那個「回頭掃同一 block、執行序在前的 SaveWrites」的啟發式整支拿掉了。★ 那個「還要有新寫的 C04D」的附加條件也早就拿掉（spec 1158）：只寫 C04B/C04C 的重畫就是真實移動、朝向刻意不變，corpus 23 處全部生效——4BF0/4BF1 的 producer 是地城主迴圈的移動前快照（spec 1155/1045）。partial 只剩**兩個 remake 這一側的補丁**（原作沒有，理由寫在 spec 1172）：鏡射多記「這三格是哪個 block 的腳本寫的」並要求與目前 block 相同，以及 Written 遮罩每次頂層執行重新起算。兩者都是因為 remake 改變隊伍位置的路徑不只一條、不是每一條都回寫 ECL 暫存器；原作靠引擎每走一步都寫 720Fh 保持三格同步，不需要這兩個補丁 |
| `0x33` | PRINT RETURN | `partial` | 120 | 13 | 原作 2CEAh 整支 14 條讀完（spec 1147）：欄 65A0h := 1、列 65A1h ＋1，兩個分支對游標做的事一樣（8B61h 只決定要不要順手清）。所以它是硬換行——連續兩條會空一行。remake 只記指令邊界（PrintReturnCount），沒有游標模型；缺口在 UI 的行模型，不是 ECL VM |
| `0x09` | SAVE | `done` | 1916 | 25 | 數值與字串記憶體都寫得進去 |
| `0x01` | GOTO | `done` | 1777 | 24 | 控制流 |
| `0x11` | PRINT | `done` | 1310 | 25 | 文字累積進 result.Text |
| `0x03` | COMPARE | `done` | 1243 | 24 | 六個比較旗標，供 16h..1Bh 使用 |
| `0x02` | GOSUB | `done` | 1087 | 24 | 控制流，返回位址進堆疊 |
| `0x12` | PRINTCLEAR | `done` | 1053 | 25 | 清框並開新頁（spec 1104） |
| `0x16` | IF = | `done` | 792 | 24 | 條件不成立時跳過下一條整條指令（spec 1106） |
| `0x00` | EXIT | `done` | 660 | 25 | 結束本次執行，PC 與堆疊由 lifecycle 驅動器接手 |
| `0x17` | IF <> | `done` | 442 | 24 | 同上 |
| `0x04` | ADD | `done` | 364 | 24 | 算術 |
| `0x0B` | LOAD MONSTER | `done` | 276 | 24 | 怪物載入請求交給戰鬥層 |
| `0x13` | RETURN | `done` | 269 | 24 | 控制流 |
| `0x19` | IF > | `done` | 216 | 23 | 同上 |
| `0x1C` | CLEARMONSTERS | `done` | 206 | 24 | 原作 120Eh 逐條讀完（37 條）：清怪物鏈與已放置數（47E6h）、清「有怪要打」旗標（8B69h，spec 1095）、把 6F70h 起 28 bytes 的戰利品池歸零（七種貨幣／寶石／珠寶，spec 1059）、沿 6F8Ch 鏈逐節點 FreeMem(63) 釋放 27h 串進去的物品節點（spec 1087），並把 7603h 設成 8。remake 這一側全部對上——怪物鏈、跨執行累積的戰利品堆都清（docs/audit/ecl-treasure-clear.md）。7603h 設成 8 的語意也讀出來了（spec 1173）：它是**下一批怪物的圖示槽**（`+143h` ＝ `ICONINDEX`，`overlay-33 entry#5` 拿它與 `CPIC` 組檔名），`inc` 在放置迴圈**結束後**才做一次 ⇒ 同一批共用一張圖，而 0..7 是隊伍成員的圖示，所以重設值是 8 不是 0。它是**快取索引不是遊戲狀態**——remake 的每個戰鬥員自己帶 SpriteBlock，可觀察行為相同，因此沒有對應動作、也不該有 |
| `0x0C` | SETUP MONSTER | `done` | 197 | 24 | 怪物編成請求交給戰鬥層 |
| `0x25` | ON GOTO | `done` | 179 | 23 | 動態分支，目的地是字面位址（spec 1110） |
| `0x08` | RANDOM | `done` | 174 | 22 | RANDOM 走 spec 1103 的 ROLLDICE 參數約定 |
| `0x2F` | AND | `done` | 161 | 20 | 位元運算 |
| `0x2B` | HORIZONTAL MENU | `done` | 156 | 24 | 水平選單 |
| `0x2A` | GETTABLE | `done` | 151 | 17 | GETTABLE 讀表（手札編號就是靠它，spec 1108） |
| `0x18` | IF < | `done` | 87 | 21 | 同上 |
| `0x3A` | DELAY | `done` | 87 | 17 | DELAY |
| `0x14` | COMPARE AND | `done` | 66 | 17 | 四運算元比較 |
| `0x1B` | IF >= | `done` | 64 | 20 | 同上 |
| `0x05` | SUBTRACT | `done` | 63 | 15 | 算術 |
| `0x27` | TREASURE | `done` | 63 | 22 | 原作 1B53h 整支 398 條讀完（spec 1151）：前七個運算元以 32-bit **覆寫** DS:6F70h 的戰利品池（1Ch 清的就是它）；第八個 ItemBlock 三選一——< 80h 載 ITEM<片>.DAX 那個區塊並把裡面每一筆都掛上鏈、= 0FFh 不給物品、80h..FEh 隨機產生 n − 80h 件。物品鏈 DS:6F8Ch 的 next 在 +2Ah 且是**前插**，顯示端從鏈頭走（overlay-05:0CF5h），所以清單是反序——remake 已跟上。隨機表的區間 bug 也修了（第二擲 48／49 原作回 59，remake 先前回劍）。隨機那一路已接上 CREATERNDTREASURE（spec 1036）——加值（1d20：1..14 給 ＋1、15..20 給 ＋2）、名稱三段索引、重量、價值、卷軸法術與特殊物品範本都照原作補齊，擲骰順序也對上。★ 多筆 TREASURE 的模型也對上了：**錢幣池覆寫、物品鏈累積**——原作那兩側是兩套資料結構，remake 先前把錢幣也當成累加（spec 1151 增補） |
| `0x20` | NEWECL | `done` | 61 | 23 | NEWECL 終止本次執行並換 block（spec 1104） |
| `0x0A` | LOAD CHARACTER | `done` | 42 | 16 | 角色欄位投影，含 7CE4h 的 and 1 遮罩（ENG-13） |
| `0x1A` | IF <= | `done` | 25 | 17 | 同上 |
| `0x07` | MULTIPLY | `done` | 24 | 7 | 算術 |
| `0x2E` | DAMAGE | `done` | 24 | 12 | 原作 2942h 整支 305 條讀完（spec 1152）：★ 旗標 bit 7 清空時**整個 byte 是次數**——連打 N 下，每下隨機挑一名隊員、用第五個運算元當攻擊值擲 TRYTOHIT，而且每下之間重擲傷害；bit 7 設定才是旗標（bit 6 全隊、bit 5 不擲豁免、bit 4 豁免成功仍吃**全額**、bit 0..4 豁免調整值——bit 4 與調整值欄位重疊，但 corpus 24 處低 5 位全 0）。目標三選一：全隊／目前角色（第五個運算元 bit 7）／隨機一名。★ 目前角色那一路傳給 MAKESAVE 的種類要**減一**且 0 代表不擲豁免，另外兩路不減一、由 bit 5 決定；兩種讀法在 corpus 上結果相同。正式路徑先前只結算全隊那 14 處，現在三種形式都結算（目標由封包自己帶著發出當下選的角色，否則整隊迴圈裡的傷害會全算到同一位身上）。★ 第三條路「單體但隨機挑一名」也接上了：擲一顆 1..隊伍人數挑人，豁免種類**不減一**、要不要擲由 bit 5 決定（corpus 走不到，照 handler 寫，形狀與全隊那一路共用同一個結算器） |
| `0x37` | LOAD PIECES | `done` | 23 | 19 | 與 21h 共用 overlay-02:0C15h..0DA3h 一支 131 條（spec 1087；⚠ IDA 只認到 0D4Ah／104 條，切掉的正是收尾重繪判斷）。三個運算元是三個牆面組槽位的片號，載得進去的槽走 LOADWALLSET(槽, 片)，而那一支收尾寫 [7210h+槽×4] := 片、[7212h+槽×4] := 槽；運算元是 0FFh 的槽由 handler 自己寫成 0FFFFh ⇒ 這三格就是存檔第 9..14 欄（spec 1076／1153）。remake 的資產載入本來就接上了，這一輪補上存檔那一半。★ 三支分派全部照抄了（spec 1153 增補）：`7Fh` ⇒ 只碰槽 1 且片號是 **0**（槽 2／3 連哨兵都不寫）、兩個閘門 `bank0^[1CEh]`／`[1D0h]`（ECL 格 `4BE7h`／`4BE8h`）都非零 ⇒ 只載槽 1／3 且 `0FFh` 整個跳過、否則逐槽迴圈。**分支由 VM 決定**（`WallSetAssignments`）而不是上層猜——沒列出來的槽這一次完全不動。★★ **兩槽分支有四個現場**（`ECL5/0x33`／`0x35`、`ECL6/0x40`／`0x45`）：這四個 block 在自己的 `LOAD PIECES` 之前就把兩個閘門都設成非零，所以**第二個運算元是死資料**——remake 先前把它載進槽 2 並寫進存檔；`7Fh` 那一支 corpus 一次都沒出現，照 handler 寫 |
| `0x15` | VERTICAL MENU | `done` | 22 | 7 | 垂直選單，選項字串與選擇結果都回得去 |
| `0x21` | LOAD FILES | `done` | 22 | 21 | 原作 0C15h 與 37h 共用一支（spec 1087）：載 3D 地圖那一路 handler **自己寫兩格 ECL 記憶體**（spec 1181）——`bank0^[18Ah] := o[1]`（格 `4BC5h`）與 `bank1^[592h] := 0`（格 `7EC9h`），條件是 `o[1] <> 0FFh`、`<> 7Fh` 且 `bank0^[1CCh] <> 0`（格 `4BE6h`，22 處寫入全在 ECL1 的三個世界地圖 block，形狀是「現在是不是第一人稱畫面」）。★ `7EC9h` **腳本讀得到**：全 corpus 34 處存取，腳本一路寫 `FFh`，唯一的讀取端 `ECL2/0x03:00F6h` 是 `COMPARE 7EC9 FF`，而清掉它的只有這個 handler ⇒ 不寫的話腳本會永遠停在那個分支。remake 現在照條件寫那兩格（都走 `recordStore`，會進 `SaveWrites`），並帶出 `LoadFilesLoaded3DMap` 讓上層知道原作走的是哪一路；實際換檔仍由上層做，那是資產層不是副作用 |
| `0x29` | ENCOUNTER MENU | `done` | 21 | 9 | 遭遇選單 |
| `0x0D` | APPROACH | `done` | 20 | 8 | 原作 0801h 逐條讀完（22 條）：距離（bank1^[582h] ＝ ECL 格 7EC1h）大於 0 就減一並用新距離重畫遭遇圖，是 0 就什麼都不做。remake 兩件都做：ApproachEncounter 減格子、ApproachCount 給上層重畫。⚠ 這個減一不影響遭遇選單——29h 進門會重新設一次距離把它蓋回去（spec 1146） |
| `0x3D` | CLEAR BOX | `done` | 17 | 5 | 清空文字框且不印新文字；只有在該次執行沒有新文字時才看得出來，遊戲層據此清掉 Message |
| `0x2C` | PARLAY | `done` | 15 | 10 | PARLAY |
| `0x30` | OR | `done` | 13 | 6 | 位元運算 |
| `0x38` | PROGRAM | `done` | 13 | 10 | 原作 30DDh 整支 104 條讀完（spec 1154）：四個分派值 0（重新初始化＋主選單）／3（DS:4FC7h := 1 ⇒ 宣告隊伍全滅、停掉迴圈）／8（結局）／9（CAMP），corpus 13 處四個值全用得到、沒有一處落在什麼都不做的預設值。★ PROGRAM 8 先跑 overlay-18:10FFh 的結局過場才回主選單問存檔——把字串位移與等鍵呼叫依序取出來，分段是**五頁四次等鍵**，而且第 4 頁是 8 行、其餘四頁各 4 行。remake 先前直接跳到存檔詢問，打通關一句結局都看不到；現在照原作的等鍵位置分頁播完才進選單。⚠ spec 1087 說 4FC7h 是「訓練免費」旗標是把它和 4FC8h 搞混了 |
| `0x40` | DESTROY ITEMS | `done` | 13 | 4 | DESTROY ITEMS |
| `0x28` | ROB | `done` | 10 | 9 | ROB 依比例取走金錢與物品 |
| `0x36` | ADD NPC | `done` | 9 | 5 | ADD NPC 建出隊員 |
| `0x32` | FIND ITEM | `done` | 8 | 3 | FIND ITEM |
| `0x35` | SAVE TABLE | `done` | 8 | 7 | SAVE TABLE |
| `0x39` | WHO | `done` | 7 | 6 | WHO 選人 |
| `0x06` | DIVIDE | `done` | 6 | 6 | 算術 |
| `0x26` | ON GOSUB | `done` | 6 | 6 | 同上，返回位址進堆疊 |
| `0x10` | INPUT STRING | `done` | 5 | 3 | 字串輸入，退格以 rune 為單位（CHT-02） |
| `0x3E` | DUMP | `done` | 5 | 5 | DUMP |
| `0x31` | SPRITE OFF | `done` | 3 | 3 | 關掉畫面上的怪物圖示；同一次執行又要求新畫面時以新的為準（原作也是先關再畫） |
| `0x1D` | PARTYSTRENGTH | `done` | 2 | 2 | 隊伍強度 |
| `0x34` | ECL CLOCK | `done` | 2 | 2 | ECL CLOCK |
| `0x3B` | SPELL | `done` | 2 | 2 | 依行軍順序找持有者，slot 與隊員索引寫回兩個位址，找不到寫 0FFh；依據是呼叫端（COMPARE FFh ＋ LOAD CHARACTER），不是命令表 |
| `0x3F` | FIND SPECIAL | `done` | 2 | 1 | FIND SPECIAL |
| `0x1E` | CHECKPARTY | `done` | 1 | 1 | CHECKPARTY 六個條件 |
| `0x3C` | PROTECTION | `done` | 1 | 1 | 原作是翻手冊的防拷問答（橋上謎題，ECL1 0x50 +1B6Fh）。呼叫端拿到控制權後直接印「YOU MAY PASS.」再 RETURN，沒有分支 ⇒ remake 不附那份對照表，一律通過就是正確行為，不是缺口 |

## 尚未還原的指令，依 block

| Block | 可達指令 | 其中未完整還原 |
|---|---:|---:|
| `ECL4.DAX/0x25` | 745 | 55 |
| `ECL4.DAX/0x21` | 563 | 50 |
| `ECL4.DAX/0x20` | 722 | 45 |
| `ECL3.DAX/0x10` | 913 | 39 |
| `ECL3.DAX/0x12` | 840 | 39 |
| `ECL5.DAX/0x33` | 709 | 37 |
| `ECL2.DAX/0x01` | 659 | 36 |
| `ECL3.DAX/0x11` | 788 | 35 |
| `ECL4.DAX/0x22` | 526 | 35 |
| `ECL5.DAX/0x32` | 634 | 34 |
| `ECL6.DAX/0x40` | 766 | 33 |
| `ECL2.DAX/0x02` | 374 | 28 |
| `ECL6.DAX/0x42` | 603 | 25 |
| `ECL5.DAX/0x35` | 598 | 24 |
| `ECL6.DAX/0x43` | 525 | 24 |
| `ECL2.DAX/0x03` | 703 | 23 |
| `ECL4.DAX/0x23` | 303 | 20 |
| `ECL1.DAX/0x50` | 798 | 17 |
| `ECL1.DAX/0x52` | 86 | 17 |
| `ECL2.DAX/0x04` | 482 | 17 |
| `ECL5.DAX/0x31` | 388 | 17 |
| `ECL1.DAX/0x51` | 726 | 16 |
| `ECL3.DAX/0x15` | 334 | 11 |
| `ECL6.DAX/0x45` | 320 | 9 |
| `ECL5.DAX/0x30` | 72 | 0 |
