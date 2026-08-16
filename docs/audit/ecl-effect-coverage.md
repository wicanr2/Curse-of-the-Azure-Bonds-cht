# ECL 副作用覆蓋（原作可達指令 → remake runtime）

由 `cmd/ecl-effect-coverage` 產生，不要手改。

- 分母是**靜態可達的指令**：`TraceGraph` 跟循序、`GOTO`／`GOSUB`／`RETURN`、`IF` 的兩條路與 `25h ON GOTO`／`26h ON GOSUB` 的每一個目的地；變長指令的長度用 `ecl.RecordEnd`（spec 1110）。同一條指令只算一次，不管玩家會跑過幾遍。
- 狀態來自 `internal/ecl.OpcodeEffects` 的人工判定，不是自動推出來的。`done` 表示副作用已產生**且**有回歸測試或實機路徑驗過；`partial` 表示只做了一部分（多半是把請求記進 result 讓上層處理）；`consumed` 表示只把運算元讀掉——**那是明確的缺口，不是 no-op**。
- 這份報告不看**參數**對不對。`24h COMBAT` 標 `partial` 說的是戰鬥服務分派點已接、回合生命週期未接（`RE-06`），不是說每一場遭遇的編成都對。
- 出現次數是**指令數**不是事件數。一個 `27h TREASURE` 出現在 12 個地方，代表 12 處寶物流程缺口，不代表 12 種不同的寶物規則。

## 摘要

| 狀態 | 指令數 | opcode 數 | 意思 |
|---|---:|---:|---|
| `done` | 13121 | 47 | 副作用已產生，且有回歸測試或實機路徑驗過 |
| `partial` | 970 | 9 | 只做了一部分，多半是把請求記進 result 讓上層處理 |
| **`consumed`** | **86** | **5** | **只讀掉運算元——原作有效果，remake 沒有** |
| 合計（靜態可達指令）| 14177 | 61 | |

## 逐 opcode

| opcode | 名稱 | 狀態 | 可達出現次數 | 出現在幾個 block | 還差什麼 |
|---|---|---|---:|---:|---|
| `0x27` | TREASURE | `consumed` | 63 | 22 | TREASURE：只讀掉八個運算元。寶物產生、隨機表與拾取流程都還沒接 |
| `0x3D` | CLEAR BOX | `consumed` | 17 | 5 | CLEAR BOX：文字框清除的畫面行為沒有接 |
| `0x31` | SPRITE OFF | `consumed` | 3 | 3 | SPRITE OFF：戰鬥圖示的顯示狀態未還原 |
| `0x3B` | SPELL | `consumed` | 2 | 2 | SPELL：ECL 觸發的法術效果沒有接（ENG-09） |
| `0x3C` | PROTECTION | `consumed` | 1 | 1 | PROTECTION：防護狀態沒有接 |
| `0x1C` | CLEARMONSTERS | `partial` | 206 | 24 | CLEARMONSTERS 清怪物鏈與已放置數；原作另清的四塊 bank1 區域未逐格對上 |
| `0x0E` | PICTURE | `partial` | 199 | 23 | PICTURE 記下 block 與 big-picture 旗標；頭像／身體分塊仍靠上層 |
| `0x24` | COMBAT | `partial` | 199 | 24 | COMBAT 是三選一的服務分派點（spec 1095）；戰鬥本身的回合生命週期仍是 RE-06 |
| `0x2D` | CALL | `partial` | 168 | 22 | CALL 是七路 switch；corpus 只用 2E10h 與 6803h，兩者的 consumer 尚未逐條驗（RE-03） |
| `0x33` | PRINT RETURN | `partial` | 120 | 13 | PRINT RETURN 目前等同換行；原作的游標行為未逐格對上 |
| `0x37` | LOAD PIECES | `partial` | 23 | 19 | LOAD PIECES 記下請求；資產載入由上層做 |
| `0x21` | LOAD FILES | `partial` | 22 | 21 | LOAD FILES 記下請求；實際換檔由上層做 |
| `0x0D` | APPROACH | `partial` | 20 | 8 | APPROACH 只記請求；原作的接近動畫與距離狀態未還原 |
| `0x38` | PROGRAM | `partial` | 13 | 10 | PROGRAM 記下 ID；PROGRAM 8 的通關序列（spec 1087）尚未接完 |
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
| `0x20` | NEWECL | `done` | 61 | 23 | NEWECL 終止本次執行並換 block（spec 1104） |
| `0x0A` | LOAD CHARACTER | `done` | 42 | 16 | 角色欄位投影，含 7CE4h 的 and 1 遮罩（ENG-13） |
| `0x1A` | IF <= | `done` | 25 | 17 | 同上 |
| `0x07` | MULTIPLY | `done` | 24 | 7 | 算術 |
| `0x2E` | DAMAGE | `done` | 24 | 12 | DAMAGE |
| `0x15` | VERTICAL MENU | `done` | 22 | 7 | 垂直選單，選項字串與選擇結果都回得去 |
| `0x29` | ENCOUNTER MENU | `done` | 21 | 9 | 遭遇選單 |
| `0x2C` | PARLAY | `done` | 15 | 10 | PARLAY |
| `0x30` | OR | `done` | 13 | 6 | 位元運算 |
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
| `0x1D` | PARTYSTRENGTH | `done` | 2 | 2 | 隊伍強度 |
| `0x34` | ECL CLOCK | `done` | 2 | 2 | ECL CLOCK |
| `0x3F` | FIND SPECIAL | `done` | 2 | 1 | FIND SPECIAL |
| `0x1E` | CHECKPARTY | `done` | 1 | 1 | CHECKPARTY 六個條件 |

## 尚未還原的指令，依 block

| Block | 可達指令 | 其中未完整還原 |
|---|---:|---:|
| `ECL4.DAX/0x25` | 745 | 90 |
| `ECL4.DAX/0x20` | 722 | 69 |
| `ECL4.DAX/0x21` | 563 | 60 |
| `ECL5.DAX/0x33` | 709 | 59 |
| `ECL3.DAX/0x10` | 913 | 55 |
| `ECL3.DAX/0x12` | 840 | 53 |
| `ECL5.DAX/0x32` | 634 | 53 |
| `ECL6.DAX/0x40` | 766 | 51 |
| `ECL4.DAX/0x22` | 526 | 48 |
| `ECL2.DAX/0x01` | 659 | 47 |
| `ECL3.DAX/0x11` | 788 | 46 |
| `ECL2.DAX/0x02` | 374 | 44 |
| `ECL5.DAX/0x35` | 598 | 44 |
| `ECL2.DAX/0x03` | 703 | 43 |
| `ECL1.DAX/0x50` | 798 | 42 |
| `ECL1.DAX/0x51` | 726 | 40 |
| `ECL6.DAX/0x42` | 603 | 36 |
| `ECL6.DAX/0x43` | 525 | 35 |
| `ECL2.DAX/0x04` | 482 | 30 |
| `ECL5.DAX/0x31` | 388 | 26 |
| `ECL1.DAX/0x52` | 86 | 23 |
| `ECL4.DAX/0x23` | 303 | 22 |
| `ECL3.DAX/0x15` | 334 | 20 |
| `ECL6.DAX/0x45` | 320 | 20 |
| `ECL5.DAX/0x30` | 72 | 0 |
