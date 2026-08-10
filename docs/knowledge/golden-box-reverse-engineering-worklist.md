# SSI Golden Box 反組譯工作清單與證據邊界

更新日期：2026-08-10（第 534 輪盤點）

第 534 輪以中文說明書掃描頁完成關鍵勘誤：印刷頁 3–4 的 `Move characters where?`
列出《光芒之池》／《青色枷的詛咒》／《幽城寶藏》的四個跨遊戲角色轉移方向，
印刷頁 12 的 `ADD CHARACTER` 也要求以 `CURSE／POOL／HILLSFAR` 選擇來源。
因此 `MOVEPARTY` 的產品語意先改列為「角色／隊伍資料轉移」`strong inference`，
不再作為秘密門或 `(13,10)`→`(8,15)` 地圖通路的先驗入口；原始 helper bytes
仍保留，P0-2 地圖 handoff 改追真正的 external map／wall interaction consumer。
原始 RAR SHA-256 與頁面雜湊見 `docs/spec/534-chinese-manual-moveparty-character-transfer.md`。

第 533 輪沒有把靜態 helper 名稱升格成秘密門語意；只用 IDA Pro 9.4 disposable
overlay-14 副本補出 `MOVEPARTY` 的 mode／`+592h` gate、B／P／K helper 的
連續操作與四個 `0x014C` set-writer caller。remake 的 post-knight SEARCH probe
沒有事件且已還原；P0-2 仍缺同版本 runtime consumer／持久性／ECL continuation。

第 532 輪沒有擴大反組譯範圍；先把已存在的 `ENTER CITY`／`JOURNEY ON`／`CAMP`
選項 rule 在開場、荒野、讀檔與紮營返回路徑收斂到 engine＋JSON。這是資料分層
修正，不是新的 ECL／地圖證據；P0-1／P0-2／P0-3 數量與秘密門的未知狀態不變。

第 530 輪新增一條有界 DOSBox 取證：原版 `START STING Wooden` 可在 Docker
暫存副本穩定抵達角色建立，截圖可見年齡與原版資訊列；這只支持角色建立畫面
與資料可見性，不縮小 P0-1／P0-2／P0-3 的正常玩家路徑缺口。測試模式可能
改變啟動／亂數／作弊狀態，故其 trace 不可直接升格為 normal-path evidence。
權威規格為 [`docs/spec/530-dosbox-cheat-assisted-character-creation-evidence.md`](../spec/530-dosbox-cheat-assisted-character-creation-evidence.md)。

本頁是工作清單，不是「已完成」清單。它回答目前還需要解讀哪些反組譯資料、
每項工作要閉合什麼證據，以及哪些舊斷言已被降級。後續 Gold Box 作品可以沿用
分類方式，但位址、雜湊、版本與劇情資料必須各自保存，不能跨遊戲套用。

第 529 輪沒有擴大反組譯範圍；先把目前已可玩的 12 個玩家法術接成共用
engine＋JSON contract，並保留原始 `spell_id`、locale message ID 與驗證邊界。
P0-1／P0-2／P0-3 的逆向缺口數量不變；後續仍只追會阻塞正常玩家路徑的資料流，
不因新 schema 而把未解出的法術、秘密門或地圖 consumer 猜成已完成。

## 先看結論

目前不需要把整個 `START.EXE`、`GAME.OVR` 或 `PC98-GAME.EXE` 逐行翻成中文。
真正缺的是「會改變玩家可玩結果」的資料流：輸入／ECL opcode → 外部 routine 或
地圖服務 → state／renderer／戰鬥／存檔 consumer → 可重播的 runtime 結果。

以資料流為單位，目前還有 11 個直接影響正常玩家結果的逆向主題：P0 外部地圖／
正常路徑 3 個、P1 ECL／外部 routine 4 個、戰鬥規則／AI 2 個、存檔／AD&D 角色
規則 2 個；另有 4 個以 fidelity／音訊／發行為主的主題。這些是工作流數量，
不是函式數量或完成百分比；詳見 `docs/spec/517-reverse-engineering-gap-inventory.md`。

第 519–524 輪已連續關閉 P0-1 的六個靜態子邊界：ECL selector 已可由
`017F:003Eh` 對到 START control vector 6 與 overlay-30 local `07C6h`，確認
兩參數的 16×16 indexed layer read；第 520 輪補上 `DS:720F／7210` 的 movement
欄位 bridge；第 521 輪再把 `0A54:0329h` 對到 resident Borland
`GetMem(Pointer &,Word)`；第 522 輪再以 `Move`／`FreeMem`／`sub_16C3E` 閉合
`DS:7206h` 四個 `0100h` destination plane 的 writer 幾何與暫存生命週期；第 523
輪又以 `START.EXE` control vector 26 與 overlay-02 `401Fh` branch 閉合
overlay-07 `1B3Fh` 的靜態 dispatcher entry；第 524 輪再以 `GEO`／`.dax` source
fragment、`0402h` gate 與既有 GEO corpus 閉合來源 record／四平面 payload 格式。
仍未閉合的是 selector producer、正常鍵盤輸入與 loader runtime、欄位 consumer、
`C04B..C04F` projection 與 runtime redraw／位置 consumer，所以 11＋4 的工作流
數量不變，也沒有資格新增 secret-door 或 movement 規則。

第 525 輪沿 PC-98 Borland symbol table 追查 `TEMPSEARCH/BDF1`，並由 overlay-02
`3BB8h..3BFDh` 的連續 writer→reader→`014A:00DEh` stub 閉合：它暫存目前
角色 record `+594h` 的低 byte、短暫寫入 `+594h=1`，再還原並呼叫
`SHOWLOCATION`。因此 `BDF1` 不再是秘密門第三平面 writer 的先驗入口；這不代表
秘密門不存在，只代表後續應轉向 `LOAD3DMAP／BLOCKCODE／MOVEPARTY` 的玩家結果
資料流。第 526 輪確認 `SEARCHREC` 是 DOS 檔案搜尋 record，不再是 map writer
owner；完整證據見 `docs/spec/525-pc98-tempsearch-display-state.md` 與
`docs/spec/526-pc98-moveparty-map-writer-searchrec-correction.md`。

第 527 輪沿同一 P0-2 資料流追查 overlay-14 的 MOVEPARTY action branch。raw
bytes 精確保留 B／P／K token 與 local 0x02F5／0x05B4／0x0714 helper call；
local 0x014C 會把 THE3DMAP +300h 的 selected 2-bit field 寫成 raw 01，且
B／P helper 各有兩個直接 call；local 0x003E 則是 movement-result 分支使用的
field-clear writer。這仍不能命名成開門／關門或新增 secret-door，因 action
predicate、BLOCKCODE consumer、ECL flag、重訪／存檔持久性與同版本 runtime
trace 尚未閉合。完整靜態邊界見
docs/spec/527-pc98-moveparty-action-writer-boundary.md；可重現 audit 為
scripts/research/pc98_overlay14_action_writer_audit.py。

第 528 輪再沿 MOVEPARTY 的連續 bytes 閉合 `017C:0039` 回傳 `AL=1／2／3` 的
控制流、`0164:0039` action input、B／P／K dispatch、P 路徑的第二次
`017C:0039` call，以及非零結果到 local `0x0807`、共同 `014A:00DE` tail 的
call-site。這仍是 exact raw control flow，不是秘密門／開門語意；特別不能把
`017C:0039` 直接合併成 overlay-30 local `0039` 或 `BLOCKCODE`。第 527 輪
規格的 B helper 第二個 caller `0x05A5` 已勘誤為 raw near-call 位址 `0x05A4`。
NP2kai 只取得 BIOS、NP2 menu 與 FDD selector 畫面，Disk 1 尚未產生可用遊戲
loader／地城 trace，因此 P0-2 仍不新增 movement／secret-door JSON。完整邊界見
[`spec 528`](../spec/528-pc98-moveparty-action-transaction-boundary.md)，唯讀
audit 為 `scripts/research/pc98_overlay14_action_transaction_audit.py`。

目前可可靠宣稱的範圍是：ECL1–ECL6 的 25 個 block、125 個 initialization entry
已通過無 unsupported-opcode 的邊界 corpus；大量窄規格也已經 `READY`。這不等於
所有外部 routine、所有分支或整條開場到結局路徑都已反組完成。`PLAN.md` 目前只有
4 個明確未勾選項目，但它是歷史累積計畫，不足以代表全部逆向缺口；本頁的優先級
才是接下來的實際順序。

## P0：先關閉目前正常玩家路徑的三個阻塞點

### 目前剩餘的反組譯量（以資料流工作單位計）

目前仍需解讀的範圍是 **11 個行為資料流主題**與 **4 個 fidelity／發行主題**，
不是 15 個函式，也不是 15 個完成百分點。就目前最急迫的 P0-1 而言，`GetMem`
owner、四平面 writer 幾何、GEO DAX source／decoded layout 已不再是缺口；剩餘仍
可拆成五個可驗證 runtime 邊界，其中第一項只剩 selector producer／正式 map
consumer：

1. `DS:7206h` 四平面的 selector producer、正式 runtime consumer 與 map projection；
   `.dax` source、`0402h` decoded record 與四個 `0100h` payload plane 已由第 524
   輪 `READY` spec 精確閉合。
2. overlay-07 `1B3Fh` 的正常鍵盤 producer／control-loader runtime；其
   control vector 26 與 overlay-02 `401Fh` 靜態 caller 已由第 523 輪閉合。
3. `DS:7212／7213`、`7Fh` sentinel 與 vector 7 `0841h` 的正式 consumer。
4. ECL `CALL 2E10h`、`C04B..C04F` 與 map service 之間的 projection／位置
   handoff。
5. DOSBox 原版實機中的目的地、座標、方向、重繪與戰後續跑 trace。

P0-2 與 P0-3 各自仍是完整的正常路徑資料流缺口；P1 的 8 個主題（ECL／外部
routine 4、戰鬥／AI 2、存檔／AD&D 規則 2）尚未因本輪靜態稽核而縮減。

| 優先 | 工作 | 目前證據 | 還缺什麼才可標 `exact` |
|---|---|---|---|
| P0-1 | 火刀戰後 `(1,8)` → `(13,10)` 的 DOS 外部地圖 handoff | ECL2 block 3 `+1B5Bh` 會 `CALL 2E10h`；overlay-02 local `2F23／2F2C` 將 selector 正規化為 `AE11h`，local `2F39` 呼叫 `017F:003Eh`；START control block raw `0x1FA0` 的 vector 6 bytes `CD 3F C6 07 00` 精確對到 overlay-30 local `07C6h`。overlay-07 local `1B3F` 已精確找到 `DS:720F／7210` 的 `0..0Fh` 循環更新與 vector 6／4 呼叫；overlay-11 已精確找到 `DS:7206h`＋`0400h` 的 Borland `GetMem(Pointer &,Word)` call-site 及初始 `7／0Dh／0或2`。第 522 輪又以 overlay-30 `133Ah..1475h` 的四次 `Move` 精確閉合 `DS:7206h +000/+100/+200/+300` 各 `0100h` 的 destination 幾何，並確認暫存 pointer 由 `FreeMem` 回收；第 523 輪再證明 `START.EXE` 的 `006B:00A2h` 是 vector 26，entry bytes `CD 3F 3F 1B 00` 指向 overlay-07 local `1B3Fh`，overlay-02 local `3002h` 在 `401Fh` branch 呼叫該 entry；第 524 輪以 overlay-30 local `1310h` 的 `GEO`／`.dax` Pascal fragments、`DS:5BEEh` area-value input、`0402h` gate 與 GEO2–GEO6 decoded corpus 閉合 `GEO<area>.dax` source／四平面 payload layout。仍未閉合的是 selector producer、正常鍵盤 producer／control-loader runtime、欄位 consumer、map projection 與 runtime consumer。overlay-30 vector 6 仍只精確到 `ES:[DI+0200h]` byte read，overlay-28 另有 `DS:7213h` consumer 與獨立 vector 7 `0841h` 路徑；既有 remake trace 沒有 `C04Bh/C04Ch` 寫入；GEO2 block 3 的**關閉狀態 movement graph**沒有合法路徑。overlay 22 `[di+4BF0h]` 仍是獨立 indexed far-pointer table candidate，不能當成 ECL handoff。現行 JSON `set_map_position` 是可重播的 `strong inference` | 追 `DS:5BEEh`／`[bp+6]` 的 selector producer、`006B` control-loader 與正常鍵盤 producer、DS:7212／7213 的正式 consumer、`C04B..C04F`／map service projection 與 `CALL 2E10h` runtime consumer 的完整橋接，並以 DOSBox／runtime trace 對上目的 map、座標、方向與暫存器；若引用 `+300h` 或 `4BF0h`，必須先證明其所屬位址空間與實際 consumer |
| P0-2 | 騎士事件後從 `(13,10)` 到 block 4 E2 `(8,15)` 的正常輸入 | PC-98 `LOAD3DMAP (017C:1253h)` 已精確以 `0402h` gate 將四個 `0100h` plane 複製到 named `THE3DMAP (0C29:A2A0h)`；`BLOCKCODE／WALLCODE` 已精確讀取 `THE3DMAP` 的 `+000／+100／+300` 與 mask／座標公式。第 528／533 輪精確保存 `MOVEPARTY` 的 `AL=1／2／3` 分流、`0164:0039` action input、B/P/K helper、共同續跑 call-site 與 writer；第 534 輪中文說明書則把它的產品語意改列為跨遊戲角色轉移候選，不再把它當作地圖秘密門／P0-2 movement consumer。`wall=09/detail=0` 仍不能普通通過，真正的地圖 action、external map consumer 與 runtime state 未知。`S`、`TEMPSEARCH/BDF1`、`SEARCHREC` 的既有邊界仍不等於第三平面 writer。NP2kai FDD selector 尚未產生遊戲 loader／地城 trace | DOS／PC-98 任一同版本的地圖 wall interaction、external map handoff、ECL flag predicate 與移動後 map state 必須在同一條 trace 閉合；`MOVEPARTY` 另需角色資料檔／selector／record round-trip 證據，不能直接寫入 `(8,15)` |
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

### 第 520 輪 P0-1 movement／consumer bridge

本輪沿所有 direct DS operand candidate 做有界 IDA 稽核，新增下列可回查資料：

- `overlay-07:1B3Fh` 依 `DS:7211h` 的 `0／2／4／6` 更新 `DS:720Fh／7210h`，
  兩欄各在 `0..0Fh` 循環；接著 `017F:003Eh` 的 AL 寫入 `DS:7213h`，
  `017F:0034h` 的 AL 寫入 `DS:7212h`，最後寫 `DS:8B68h=1`。
- `overlay-11:00E9h` 把 `DS:7206h` 與 `0400h` 傳給 `0A54:0329h`；
  `03F2..03FCh`／`078Eh..0798h` 將兩欄初始化為 `07h／0Dh`，方向欄為
  `00h／02h`。這是 call-site／writer exact；第 521 輪已將 callee 對到
  `GetMem(Pointer &,Word)`，第 522 輪再由 `Move`／`FreeMem` 閉合四平面
  destination 的 offset／count；buffer 內容的 `.dax` record 與正式 map 語意仍未知。
- `overlay-30:06BDh`（vector 4）讀取 `+000h／+100h` 的高低 nibble；
  `07C6h`（vector 6）讀取 `+200h` byte。`overlay-28:00CCh` 呼叫 vector 6、
  比較 `DS:7213h` 與 `7Fh`，稍後另呼叫 vector 7 `017F:0043h → 0841h`。
- `overlay-14:003Eh` 是另一個 control segment `00D2h` 的 `+300h` mask
  read/write candidate；不能因 selector 編號相同就與 overlay-30 vector 4
  合併。

這些 bytes 把 `DS:720F／7210` 的 writer 與 vector consumer 範圍縮小，但沒有
減少 11 個行為主題；完整 report／hash 見
[`spec 520`](../spec/520-dos-movement-to-overlay-cell-layer-bridge.md)。

### 第 521 輪 P0-1 `GetMem` owner 邊界

第 521 輪在 resident `START.EXE.i64` 中將 `0A54:0329h` 解析為
`seg050:1A54:0329h`／IDA EA `0x1A869`，原始符號為 Borland
`GetMem(Pointer &,Word)`；overlay-11 的 `DS:7206h` 與 `0400h` call-site 因此
可標成「配置 1 KiB pointer buffer」的 `strong inference`。這移除了「callee
完全未知」的錯誤現況描述，但仍不能把 buffer 直接命名成 GEO、wall、terrain 或
secret-door plane。詳細位址、hash、入口 bytes 與 direct xrefs 見
[`spec 521`](../spec/521-dos-getmem-buffer-owner.md)。

### 第 522 輪 P0-1 四平面 writer 邊界

第 522 輪已把「配置後 writer／四平面 layout 完全未知」從目前斷言中移除。IDA
在 overlay-30 local `133Ah..1475h` 證明四次 `Move` 的 source 是暫存 pointer
`+002h/+102h/+202h/+302h`，destination 是 `DS:7206h` 的
`+000h/+100h/+200h/+300h`，每次 `0100h` bytes；隨後同一暫存 pointer 與返回
word 傳給 resident `FreeMem`。`0636:08DEh` 在 IDA `0x16C3E` 的原始名稱仍是
`sub_16C3E`，但 function dump 已證明它會做 `BlockRead`、暫存配置／資料轉換／
釋放；前置 `Store string`／`Concat` 的 stack cleanup 也與其 `retf 12h` 閉合。
這使 writer 幾何與暫存生命週期達 `exact`，但不替四個 offset 命名成 wall、door、
terrain 或 secret-door。完整 hash／位址空間見
[`spec 522`](../spec/522-dos-buffer-four-plane-fill.md)；工具為
`scripts/ida/dos_overlay30_buffer_copy_audit.idc` 與
`scripts/ida/dos_start_buffer_routines_audit.idc`。

### 第 523 輪 P0-1 control vector／dispatcher entry 邊界

第 523 輪修正「overlay-07 `1B3Fh` 沒有 direct xref，所以 entry 仍完全未知」的
目前斷言。`START.EXE` control block raw `0x0E60` 對應 runtime selector `006Bh`；
其 vector table 從 `006B:0020h` 開始，每項 5 bytes。vector 26 位於
`006B:00A2h`／raw `0x0F02`／IDA EA `0x10752`，連續 bytes 為
`CD 3F 3F 1B 00`，local target 是 overlay-07 `1B3Fh`。overlay-02 local
`2FFD..3007` 另有 `cmp ax,401Fh` 後的 `call far 006B:00A2h`。因此
control vector 與靜態 dispatcher caller 已達 `exact`。

這不等於已找到普通鍵盤 producer：`401Fh` 與既有 ECL `C01Eh` 的對照目前仍是
`strong inference`，`006B` loader 如何在 runtime 執行 `CD 3F`、相鄰按鍵如何
進入其他 selector、`DS:7212／7213` 如何被正式 consumer 使用，都仍是
`unknown`。overlay-07 自身的 `direct_cref_count=0` 與
`raw_LE16_target_1B3F_count=0` 只能表示沒有本地 literal／direct xref，不能再
被寫成「沒有 caller」。完整位址空間、hash、bytes 與推論分層見
[`spec 523`](../spec/523-dos-overlay07-vector26-entry.md)。

### 第 524 輪 P0-1 GEO source／decoded payload 邊界

第 524 輪把「overlay-30 只知道某個 `.dax`、`0402h` 與四個 plane 的正式來源／
格式未知」從現行斷言移除。IDA raw bytes 在 overlay-30 local `1310h`／`1314h`
是 Pascal `GEO`／`.dax` fragments；local `1341h` 讀 `DS:5BEEh` 並以 width `1`
呼叫 resident numeric-to-string helper；local `1361h`／`137Bh` 以
`Store string`／`Concat` 組合；local `1393h` 把 `[bp+6]`、output pointer 與
decoded-size word 傳給 `0636:08DEh`。local `139Eh` 只讓 `0402h` 通過，成功後從
output `+002h` 將四段各 `0100h` 搬到 `DS:7206h +000/+100/+200/+300`。

同一份 archive 的 `GEO2.DAX` 三個 block 全部 decoded `0x402` bytes；`GEO3`–
`GEO6` 也相同。既有 round 56 parser 已以 real corpus 閉合前 2-byte prefix＋四
個 `0x100` plane 的格式與 plane 語意。因此目前應寫成：`GEO` source／`0402h`
decoded payload／四平面 layout `exact`；`GEO<DS:5BEEh 十進位值>.dax` 完整 file-open
trace、`DS:5BEEh` 正式欄位名、`[bp+6]` selector producer、runtime consumer 與
map projection 仍是 `strong inference／unknown`。完整 hash／raw bytes／工具見
[`spec 524`](../spec/524-dos-overlay30-geo-loader-source.md)。

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
5. 本輪取代「overlay-07 `1B3Fh` 沒有 direct xref，所以沒有 caller／entry 仍未知」
   的現行斷言：`START.EXE` control vector 26 與 overlay-02 `401Fh` branch 已
   精確證明一條靜態 dispatcher entry；但普通鍵盤 producer、control-loader
   runtime 與返回資料流仍是 `unknown`。`raw_LE16_target_1B3F_count=0` 只適用
   overlay-07 自身的有界掃描，不得外推成整個程式沒有 caller。
6. 本輪取代「overlay-30 的 `.dax` record／`0402h`／四個 plane 來源完全未知」的
   現行斷言：raw `GEO`／`.dax` fragments、`0402h` gate、`+002h` payload 與
   GEO2–GEO6 corpus 已閉合 source／format；只保留完整 file-open trace、區域欄位
   正式命名、selector producer 與 runtime map consumer 的未知，不得再重做同一份
   `.dax` header／四平面掃描。
7. 第 525 輪取代「`BDF1/TEMPSEARCH` 是秘密門／第三平面 map writer 候選，應繼續
   從 BDF1 raw word 掃描找 map service」的路由：overlay-02 `3BB8h..3BFDh`
   精確顯示它是目前角色 `+594h` 的暫存／還原，尾端呼叫 `014A:00DEh` 的
   `SHOWLOCATION` stub；overlay-11 `06BEh..06CDh` 只是初始化清零。第 526 輪又以
   Borland type／DOS `INT 21h` consumer 確認 `SEARCHREC` 是檔案搜尋 record；仍
   不可把 `SEARCHREC`、`SECRET` 或 `HIDDEN` 名稱升格成 map writer／可走秘密門。

## 每項反組譯工作的完成門檻

只有同時具備下列資料，才能從 `DRAFT` 升為 `READY`：

1. 輸入檔名、SHA-256、平台／版本、工具版本與位址空間。
2. 連續 raw bytes、operand decoding、caller／reader／writer 或等價 runtime trace。
3. 明確的推論等級：`exact`、`strong inference`、`hypothesis`、`unknown`。
4. 正常玩家輸入抵達 boundary；direct-entry 只能作縮小問題的 probe。
5. 對應的 engine／game-pack JSON contract、stable ID 測試與失敗即關閉行為。

第 526 輪後的下一個最有價值工作仍是 P0-1 與 P0-2 的正式 consumer，但不再重做
vector target、GEO `.dax` payload 或 BDF1 raw word 掃描：P0-1 先追 overlay-30
`[bp+6]` selector／`DS:5BEEh` producer 與 `DS:7206h` 四平面 runtime consumer，
再以原版 DOSBox 實際鍵盤輸入追 `006B` control-loader、`401Fh` 與相鄰 selector，
最後閉合 `DS:7212／7213` consumer、`C04B..C04F` projection，以及目的 map、座標、
方向、重繪與戰後續跑 trace。P0-2 的 `LOAD3DMAP (017C:1253h)`、
`THE3DMAP (0C29:A2A0h)`、`BLOCKCODE (017C:04DEh)`／`WALLCODE (017C:060Dh)`
loader／buffer／普通 movement reader 靜態邊界已由第 525 輪閉合；下一步改追
真正的 `wall=09/detail=0` external map／wall interaction consumer、ECL predicate
與 runtime state，不再把 `MOVEPARTY (00C9:0BCCh)` 的 `P/BASH/KNOCK/ENTER`
當作地圖交易先驗。`MOVEPARTY` 另列為跨遊戲角色資料轉移研究，需追角色檔案／
selector／record round-trip；`SEARCHREC` 不再列為 map owner，`BDF1` 只保留為
顯示狀態證據。
overlay 22 `[di+4BF0h]` 仍只作獨立的位址空間候選，不把它當成先驗入口。
在正式 consumer 與 runtime 閉合前不改 movement 規則、不把 detail 0 泛化成可走門。

最新完整盤點與 GEO 勘誤見 `docs/spec/517-reverse-engineering-gap-inventory.md`；
第 518 輪位址空間稽核見 `docs/spec/518-dos-start-ecl-call-address-space-audit.md`；
第 519 輪 overlay vector／cell-layer 邊界見
`docs/spec/519-dos-overlay-vector-to-cell-layer-accessor.md`；第 520 輪 movement／
consumer bridge 見 `docs/spec/520-dos-movement-to-overlay-cell-layer-bridge.md`；
第 521 輪 GetMem owner 見 `docs/spec/521-dos-getmem-buffer-owner.md`；第 522 輪
四平面 writer 見 `docs/spec/522-dos-buffer-four-plane-fill.md`；第 523 輪
control vector／dispatcher entry 見
`docs/spec/523-dos-overlay07-vector26-entry.md`；第 524 輪 GEO loader source／
四平面 payload 見 `docs/spec/524-dos-overlay30-geo-loader-source.md`；第 525 輪
`TEMPSEARCH/BDF1` 暫存／`SHOWLOCATION` 邊界見
`docs/spec/525-pc98-tempsearch-display-state.md`。
