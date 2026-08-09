# SSI Golden Box 反組譯工作清單與證據邊界

更新日期：2026-08-09（第 519 輪盤點）

本頁是工作清單，不是「已完成」清單。它回答目前還需要解讀哪些反組譯資料、
每項工作要閉合什麼證據，以及哪些舊斷言已被降級。後續 Gold Box 作品可以沿用
分類方式，但位址、雜湊、版本與劇情資料必須各自保存，不能跨遊戲套用。

## 先看結論

目前不需要把整個 `START.EXE`、`GAME.OVR` 或 `PC98-GAME.EXE` 逐行翻成中文。
真正缺的是「會改變玩家可玩結果」的資料流：輸入／ECL opcode → 外部 routine 或
地圖服務 → state／renderer／戰鬥／存檔 consumer → 可重播的 runtime 結果。

以資料流為單位，目前還有 11 個直接影響正常玩家結果的逆向主題：P0 外部地圖／
正常路徑 3 個、P1 ECL／外部 routine 4 個、戰鬥規則／AI 2 個、存檔／AD&D 角色
規則 2 個；另有 4 個以 fidelity／音訊／發行為主的主題。這些是工作流數量，
不是函式數量或完成百分比；詳見 `docs/spec/517-reverse-engineering-gap-inventory.md`。

第 519 輪只關閉 P0-1 的一個靜態子邊界：ECL selector 已可由 `017F:003Eh`
對到 START control vector 6 與 overlay-30 local `07C6h`，並確認它是兩參數的
16×16 indexed layer read candidate。地圖 plane、目的地 writer、`C04B..C04F`
projection、runtime redraw／位置 consumer 仍未閉合，所以 11＋4 的工作流數量
不變，也沒有資格新增 secret-door 或 movement 規則。

目前可可靠宣稱的範圍是：ECL1–ECL6 的 25 個 block、125 個 initialization entry
已通過無 unsupported-opcode 的邊界 corpus；大量窄規格也已經 `READY`。這不等於
所有外部 routine、所有分支或整條開場到結局路徑都已反組完成。`PLAN.md` 目前只有
4 個明確未勾選項目，但它是歷史累積計畫，不足以代表全部逆向缺口；本頁的優先級
才是接下來的實際順序。

## P0：先關閉目前正常玩家路徑的三個阻塞點

| 優先 | 工作 | 目前證據 | 還缺什麼才可標 `exact` |
|---|---|---|---|
| P0-1 | 火刀戰後 `(1,8)` → `(13,10)` 的 DOS 外部地圖 handoff | ECL2 block 3 `+1B5Bh` 會 `CALL 2E10h`；overlay-02 local `2F23／2F2C` 將 selector 正規化為 `AE11h`，local `2F39` 呼叫 `017F:003Eh`；START control block raw `0x1FA0` 的 vector 6 bytes `CD 3F C6 07 00` 精確對到 overlay-30 local `07C6h`。`07C6h` 只精確到兩個 word、`0..0Fh` bounded 16×16 index、`DS:7206h` far pointer 與 `ES:[DI+0200h]` byte read；其 map plane／writer／consumer 未知。既有 remake trace 沒有 `C04Bh/C04Ch` 寫入；GEO2 block 3 的**關閉狀態 movement graph**沒有合法路徑。overlay 22 `[di+4BF0h]` 仍是獨立 indexed far-pointer table candidate，不能當成 ECL handoff。現行 JSON `set_map_position` 是可重播的 `strong inference` | 找到 `DS:7206h` 初始化／owner、`DS:720F／7210／7213` writer／consumer、`C04B..C04F`／map service projection 與 `CALL 2E10h` runtime consumer 的完整橋接，並以 DOSBox／runtime trace 對上目的 map、座標、方向與暫存器；若引用 `4BF0h`，必須先證明它所屬位址空間與實際 consumer |
| P0-2 | 騎士事件後從 `(13,10)` 到 block 4 E2 `(8,15)` 的正常輸入 | PC-98 `MOVEMENT` 的 `BLOCKCODE` 證明抽樣的 `wall=09/detail=0` 不能普通通過；`S` 只精確到目前角色 record `+594h` bit 0 與 `SHOWLOCATION`。診斷器把該 `wall=09` 邊暫視為開啟後可得到候選路徑，但沒有找到第三平面 writer；攻略的 `~` 仍只有 `layout-only` | DOS／PC-98 任一同版本的 secret-door/search writer、ECL flag predicate 與移動後 map state 必須在同一條 trace 閉合；不能把 static BFS 失敗寫成永久不相連，也不能直接寫入 `(8,15)` |
| P0-3 | block 4 入口後至火刀據點出口／返回世界的外部 handoff | block 4 初始 `LOAD FILES 4,2,FFh`、`LOAD PIECES 1,2,4` 與入口文字已解讀；正常玩家抵達 block 4 仍未取代座標輔助 | 按 terrain／boundary 分段找出 `NEWECL`、地圖服務、返回 world map 的 writer／consumer，並以同一 ECL session 驗證重訪與旗標副作用 |

P0-1 的 remake adapter 可以暫時保留，因為它已標註 `strong inference`；P0-2、P0-3
不得再新增直接座標注入。任何 probe 都必須在測試名稱與規格中明寫
`coordinate-assisted`，不能被算入正常玩家路徑。

### 第 518 輪 P0-1 位址空間排除結果

Docker 內以 IDA Pro 9.4 開啟 `START.EXE.i64` 的 disposable copy，逐一掃描
`START.EXE` 的 IDA segments。`LE16=2E10h` 唯一命中在 `seg043:0x6634`
（IDA EA `0x16634`），IDA 將它解成非 code 的長度前綴字串
`db 16,'. Check install.'`；對候選 EA `0x2E10` 與以 MZ base `0x10000`
換算的 `0x12E10` 都沒有 direct code xref。完整輸入雜湊、工具版本與 raw bytes
見 [`spec 518`](../spec/518-dos-start-ecl-call-address-space-audit.md)。

這只排除「在 resident `START.EXE` 直接找 `sub_2E10`」的搜尋路徑；不表示
`CALL` 沒有經 overlay dispatch、ECL interpreter、far pointer 或 map service
執行。P0-1 仍須找到目的地 producer → `C04B..C04F`／map service projection
→ `CALL 2E10h` consumer 的完整橋接，11 個行為逆向主題的數量不變。

同輪在抽出的 `GAME.OVR overlay-02` 看到連續 code：`0x2F23
sub ax,7FFFh`、`0x2F2C cmp ax,AE11h`、`0x2F39 call far 017F:003Eh`。
`AE11h` 是 `2E10h−7FFFh` 的 16 位元 dispatch selector；這確認下一步不是
尋找 `2E10h` literal，而是解析 `017F:003Eh` 的 module／重定位與後續 consumer。
抽取檔名 `overlay-02` 與公開 reference 的 `ovr003` 標籤尚未證明是同一編號，
兩者暫不合併。

### 第 519 輪 P0-1 靜態 dispatch 子邊界

第 519 輪以 `START.EXE` MZ header `0x7B0`、control image paragraph `017Fh`
與 raw offset `0x1FA0` 重新對齊 overlay manager control block。`+04` 的
`0x3DF87` 與 `+08` 的 `0x147F` 分別吻合 `GAME.OVR` 的 overlay-30 起點與
長度；vector table 在 raw `0x1FC0`，`017F:003Eh = +20h + 6×5`，因此 vector
6 的 bytes 是 `CD 3F C6 07 00`，目標是 local `07C6h`。相鄰 vector 7 的
`CD 3F 41 08 00` 才是 `0841h`；`0841h` 不得再當成 ECL `2E10h` target。

overlay-30 local `07C6h` 的 raw／IDA 連續指令顯示 `retf 4`，不是 `retf 6`；
其真正 read 是 `ES:[DI+0200h]`，不是 `+0100h`。前置 `local 0556h`／
`DS:8B5Eh` guard、`DS:7206h` far pointer 與 bounded index 都保留在
[`spec 519`](../spec/519-dos-overlay-vector-to-cell-layer-accessor.md)。
這項修訂只清掉臨時位移／摘要誤讀，沒有進入 engine、JSON 或 regression。

因此 P0-1 的下一步不再是「找 `017F:003Eh` 的 module」，而是：

1. 找 `DS:7206h` far pointer 的初始化、owner 與 relocation。
2. 分開追 `DS:720F／7210／7213` 及 vector 4 的 `DS:7211／7212` writer／consumer。
3. 將 `C04B..C04F`、目的地 producer、map projection 與 DOSBox runtime redraw／
   位置 trace 接成同一條資料流。

### P0-1 的新線索（尚未是結論，也不縮小成 `4BF0h`）

DOS extracted `overlay-22.bin` 的 IDA Pro 9.4 disposable report 在 overlay-local
`0x0969／0x096D` 看見 `[di+4BF0h]／[di+4BF2h]` writer，在 `0x03CF` 看見同形
far-pointer reader，在 `0x099D` 看見 indexed clear。這只能把「overlay 22 的
indexed table」列入位址空間稽核，不得把相同數字直接命名成 ECL
`4BF0h／4BF1h`、地圖座標或 `set_map_position` 來源。這筆命中目前不會把
P0-1 的 map writer 假設前移；完整 bytes、hash、工具版本與位址基準見
`docs/spec/517-reverse-engineering-gap-inventory.md`。

## P1：可重用引擎與完整遊戲行為

### ECL 與外部 routine

- 盤點所有 `CALL`、`NEWECL`、`LOAD FILES`、`LOAD PIECES`、`PROGRAM`、`COMBAT`、
  `TREASURE`、`INPUT STRING` 的 producer／consumer；優先處理會改變 block、地圖、
  party、旗標或 continuation 的 routine。
- 對每個 work address 保留 ECL work、file offset、overlay-local offset、
  Borland `segment:offset` 與 runtime linear address 的分開欄位。相同十六進位數字
  不能跨位址空間合併。
- 已讀取的 ECL bytes 只能證明 opcode 與寫入值；要命名為年齡、命中、AC、地圖
  開關或劇情旗標，仍要找到 reader／consumer 或 runtime 效果。
- 對尚未覆蓋的 random encounter、FLEE／交涉、NPC 離隊、物品與劇情 flag，採
  「一個玩家 boundary 一份 spec」；不要為了追求全 binary 函式名稱而製作沒有
  consumer 的名義索引。

### 戰鬥規則與 AI

- 敵我選敵 producer、視線／射程／方向 tie、FLEE／GUARD／HELD action 與敵方
  Quick AI 的完整 caller chain。
- 弓箭、Magic Missile、Fireball、Lightning Bolt、Stinking Cloud、Cloudkill
  等逐項追 caster windup、travel、impact、save、damage、death、persistent
  effect、sound cue 與 ECL handoff；現有共用 timeline 不能取代逐法術 oracle。
- 反組譯與公開影片只能互補：影片證明看見的演出與時間順序，數值、旗標與 RNG
  必須回到原始 bytes／DOSBox。

### 存檔、角色與規則

- `SAVGAM?.DAT` 尚未知欄位、完整 sidecar、副職業、delete／rename semantics、
  player serialization 與跨作品角色轉移。
- AD&D 全規則仍需從 player／monster／spell records 的 consumer 追出，尤其
  age effect、class limit、alignment、特殊能力與戰後恢復不能只靠攻略數字。

## P2：原版 fidelity 與發行

- DOS 地城斜牆／階梯、door／roof overlay、城市／AREA／WILDERNESS 與每張地圖的
  幾何及畫面逐像素校準。
- DOS combat frame 的動態 placement、八方向小人、完整 sprite frame timing，
  以及原版 UI／字型密度的逐狀態對照。
- PC Speaker／Tandy、PC-98 YM2203／MSCDRV 的真實 runtime save/resume、CPU／OPN
  共時 phase、analog mixer gain 與完整音效 producer；已有的合成器測試不代表
  原硬體 cycle-exact。
- 三平台打包、長時間遊玩、完整正常路徑與 release 驗收。

## 本輪移除或降級的錯誤斷言

1. `docs/spec/297-fire-knife-hideout-transition.md` 原本把測試中的直接
   `(8,15)` 寫入描述成「正式 regression crosses `(8,15)`」。這不成立；文件保留
   E2／`NEWECL 4` 的歷史證據，但現行狀態改為 `SUPERSEDED`，正常路徑改由本頁
   P0-2／P0-3 管理。
2. `PLAN.md` 原本把 block 4 `(6,1,S)` 勾成完成；改為未完成，並明示那只是
   direct-entry／coordinate-assisted boundary。`docs/spec/471` 的文字資料化測試
   同樣不再暗示 block 2→3→4 已是正常步行路徑。
3. 移除未能正確表達位址基準的 `scripts/ida/pc98_generic_local_audit.idc`，
   以及重複且容易被誤讀成 instruction xref 的
   `scripts/ida/dos_map_workcell_audit.idc`。保留的 raw candidate scanner 與
   overlay dump 只輸出候選／連續 bytes，不自動升格語意。
4. DOS overlay raw little-endian 掃描目前沒有在已抽出的 overlay 找到 literal
   `4C28h`；這只能表示「沒有直接 literal 命中」，不能表示沒有經指標或通用
   interpreter 的 consumer。此結果不得寫成「4C28 未使用」。

## 每項反組譯工作的完成門檻

只有同時具備下列資料，才能從 `DRAFT` 升為 `READY`：

1. 輸入檔名、SHA-256、平台／版本、工具版本與位址空間。
2. 連續 raw bytes、operand decoding、caller／reader／writer 或等價 runtime trace。
3. 明確的推論等級：`exact`、`strong inference`、`hypothesis`、`unknown`。
4. 正常玩家輸入抵達 boundary；direct-entry 只能作縮小問題的 probe。
5. 對應的 engine／game-pack JSON contract、stable ID 測試與失敗即關閉行為。

第 519 輪後的下一個最有價值工作仍是 P0-1，但範圍已由「解析 vector target」
縮小為「解析 vector target 的資料流與 runtime handoff」：先追
`DS:7206h`、`DS:720F／7210／7213`、vector 4 的 sibling fields，再追
`C04B..C04F` projection 與 DOSBox map／座標／方向 trace。overlay 22
`[di+4BF0h]` 只作獨立的位址空間候選，不再把它當成先驗入口。PC-98 `S`／`BDF1`
目前只有顯示／record state 證據，沒有理由繼續擴大 raw word 掃描。在 writer
閉合前不改 movement 規則、不把 detail 0 泛化成可走門。

最新完整盤點與 GEO 勘誤見 `docs/spec/517-reverse-engineering-gap-inventory.md`；
第 518 輪位址空間稽核見 `docs/spec/518-dos-start-ecl-call-address-space-audit.md`；
第 519 輪 overlay vector／cell-layer 邊界見
`docs/spec/519-dos-overlay-vector-to-cell-layer-accessor.md`。
