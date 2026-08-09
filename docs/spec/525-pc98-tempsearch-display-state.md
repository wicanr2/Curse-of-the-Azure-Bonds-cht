# 第五百二十五輪：PC-98 `TEMPSEARCH/BDF1` 暫存狀態與 `SHOWLOCATION` 邊界

狀態：`READY`（靜態資料流邊界；不含秘密門 writer、正常路徑或新遊戲規則）
日期：2026-08-09

## 目的

第 516 輪把 PC-98 `BDF0/BDF1` 列為可能的秘密門／搜尋 bridge 入口，但當時只
完成局部 state 抽樣。本輪利用 PC-98 Borland symbol table、overlay-02／11 的
連續 bytes，以及 IDA Pro 9.4 的 disposable database，確認 `BDF1` 的實際窄
資料流：它是 `TEMPSEARCH` 的暫存欄位，包住 `SHOWLOCATION` 呼叫；目前沒有證據
把它接到第三平面 map writer。

這項結論會移除「繼續擴大 BDF1 raw word 掃描以尋找秘密門 writer」的錯誤工作路由，
但不把秘密門判定為不存在。`LOAD3DMAP`／`MOVEPARTY`／`BLOCKCODE` 與
`SEARCHREC` 的 owner／consumer 仍需獨立追查。

## 輸入與工具

| 輸入／工具 | SHA-256／版本 | 位址基準 |
|---|---|---|
| PC-98 `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland `segment:offset`／resident IDA linear |
| PC-98 baseline `PC98-GAME.EXE.i64` | `0aa83775a39b5bd8cf10b5ed04c0508845127aed5f2f299656ec7f3641249696` | IDA database；只在容器 disposable copy 上開啟 |
| PC-98 overlay-02 | `49a35a10c3a08def6fcf7d17cfa48fb58e6ce70d88d754163c54c1d074d338a4` | overlay-local code offset |
| PC-98 overlay-11 | `6025b37b88c923ead20c61592392684bc2b92880feae1cc2412c9f2bb88d8ba0` | overlay-local code offset |
| IDA | `ida-pro-9.4-ver2:uidfix-v1`／IDA Pro 9.4 | 非破壞性 xref／連續 bytes |
| Borland symbol audit | `pc98-symbol-audit`，SHA 對應 repo 現行 source | 附加 debug table，不回寫 binary |
| 本輪 xref script | `scripts/ida/pc98_search_state_xref_audit.idc`；SHA `4e6d45d9ec4c3ce177dc0a046f6482882eaffc6f524e39d6598e5842f4842e34` | IDA linear／Borland data segment 對照 |
| 本輪 map loader script | `scripts/ida/pc98_load3dmap_dataflow_audit.idc`；SHA `cab51a5f32add2fc76380045c734489d1fd22b3cc0284d23cba20b5a8969ada0` | overlay-local code offset；只分析 disposable raw copy |

PC-98 Borland data segment `0C29h` 在 resident database 的 IDA linear base 是
`01C290h`，所以：

```text
Borland 0C29:BDF0h  -> IDA EA 028080h
Borland 0C29:BDF1h  -> IDA EA 028081h
```

兩個位址空間不得與 overlay-local `0BDF0h／0BDF1h` 的 raw operand 混成同一個
位址。overlay 的 `ds:` 是執行時 data segment；報告中的 `local` 只表示該 overlay
檔案的 code offset。

## Borland symbol／type table 結果

原始 `PC98-GAME.EXE` 的 legacy debug table 為 version `0208h`，共有 1725 個
symbols、1531 個 types、641 個 members。與本輪相關的正式名稱是：

| 原始名稱 | Borland 定位／type | 可證明內容 | 等級 |
|---|---|---|---|
| `TEMPSEARCH` | `0C29:BDF1h`，type 8 | `BDF1h` 是一個具正式名稱的 data symbol | `exact` |
| `SEARCHREC` | type 1458，size `002Bh`，detail `00 56 02` | 存在一個 43-byte named type | `exact`；欄位尚未解讀 |
| `SEARCHING` | absolute symbol `0000:0001h` | 存在名為 `SEARCHING` 的符號 | `exact`；用途／值語意未知 |
| `SEARCH` | absolute symbol `0000:02CAh` | 存在名為 `SEARCH` 的符號 | `exact`；用途／值語意未知 |
| `NOSECRET`／`SECRET` | absolute symbols | 存在這兩個名稱 | `exact`；不能由名稱直接命名 map state |
| `HIDDEN` | absolute symbol `0000:0002h` | 存在名為 `HIDDEN` 的符號 | `exact`；不能由名稱直接命名 wall detail |
| `LOAD3DMAP` | `017C:1253h` | 3D map loader 的正式入口；`0402h` gate 與 `THE3DMAP` 四平面 copy 已閉合 | `exact`；selector／欄位語意與 writer 未知 |
| `BLOCKCODE` | `017C:04DEh` | movement 的 wall/detail 判定入口 | `exact` |
| `WALLCODE` | `017C:060Dh` | wall code handler 入口 | `exact` |
| `MOVEPARTY`／`PREMOVEPARTY` | `00C9:0BCCh`／`00C9:08E4h` | movement producer／前置入口 | `exact`；secret-door bridge 未閉合 |

`SEARCHREC` 的存在不等於已知道 record 欄位；`SECRET`、`HIDDEN` 等名稱也不能
單獨證明第三平面 writer 或 detail `1`。在取得 member ordinal、record consumer
與 map state trace 前，這些只可作下一個查詢入口。

## overlay-02：`TEMPSEARCH` 的暫存／還原資料流

overlay-02 local `3BB8h..3BFDh` 的連續 bytes 如下。這段保留原始 operand 與
overlay-local 位址，不以自訂函式名覆蓋它們：

```text
3BB8  C43E097F       les  di, ds:7F09h
3BBC  268B859405     mov  ax, es:[di+594h]
3BC1  25FDFF         and  ax, 0FFFDh
3BC4  A2F1BD         mov  ds:0BDF1h, al
3BC7  C43E097F       les  di, ds:7F09h
3BCB  26C78594050100 mov  es:[di+594h], 1
3BD2  C6066CA601     mov  ds:0A66Ch, 1
3BD7  9A25007201     call far 0172:0025h
3BDC  FF36197F       push word ds:7F19h
3BE0  0E             push cs
3BE1  E8C1FD         call local 39A5h
3BE4  803E977800     cmp  ds:7897h, 0
3BE9  7404           jz   3BEFh
3BEB  0E             push cs
3BEC  E88FF6         call local 327Eh
3BEF  A0F1BD         mov  al, ds:0BDF1h
3BF2  30E4           xor  ah, ah
3BF4  C43E097F       les  di, ds:7F09h
3BF8  2689859405     mov  es:[di+594h], ax
3BFD  9ADE004A01     call far 014A:00DEh
```

可閉合的最小語意是：

1. 從目前 record `ES:[DI+594h]` 取值，清除 bit 1（`AND FFFDh`），存入
   `TEMPSEARCH/BDF1` 的低 byte。
2. 暫時把目前 record 的 `+594h` 寫成 `1`。
3. 執行中間 UI／狀態流程，再把 `TEMPSEARCH` 載回同一個 record `+594h`。
4. 最後呼叫 far stub `014A:00DEh`。既有 TPOV resolver 將這個 stub 解析到
   overlay-24 `SHOWLOCATION` local `2E8Ch`；這只證明顯示／位置狀態 handoff，
   不證明 map third-plane mutation。

「暫存 record state 包住 `SHOWLOCATION`」是由連續 writer→reader→call 順序得到
的 `strong inference`；每個 raw operand 與寫回動作本身是 `exact`。`+594h` 的
完整角色欄位語意仍不可只靠這一段命名。

## overlay-11：初始化，不是秘密門 writer

INIT overlay-11 local `06BEh..06CDh` 是：

```text
06BE  C606F1BD00  mov byte ds:0BDF1h, 0
06C3  C606F2BD00  mov byte ds:0BDF2h, 0
06C8  C606F3BD00  mov byte ds:0BDF3h, 0
06CD  C606F4BD00  mov byte ds:0BDF4h, 0
```

這是初始清零的 `exact` 證據。它不能證明 `BDF1` 是秘密門旗標，也不能證明
`BDF2..BDF4` 的欄位名稱；同一段沒有 third-plane address-taken writer。

## `LOAD3DMAP` 與 `THE3DMAP`／`BLOCKCODE` 的新邊界

同一份 Borland symbol table 還提供 `THE3DMAP=0C29:A2A0h`、
`LOAD3DMAP=017C:1253h`、`BLOCKCODE=017C:04DEh` 與
`WALLCODE=017C:060Dh`。本輪以 PC-98 overlay-30 disposable raw database
（SHA-256 `a8cd1b853e1c755e2031a1b5154364b11245cf501b68eff89fd8432a4f7f9b07`）
新增 `scripts/ida/pc98_load3dmap_dataflow_audit.idc`（SHA-256
`cab51a5f32add2fc76380045c734489d1fd22b3cc0284d23cba20b5a8969ada0`）抽樣
三個正式 symbol 邊界。

`LOAD3DMAP 017C:1253h` 的連續 bytes 可閉合下列資料流：

```text
1277  mov  al,[bp+6]
1285  call far 0723:0824       ; GETDATABLOCK symbol boundary
128A  cmp  word [bp-2],0
1290  cmp  word [bp-2],0402h
12C5  mov  word [bp-8],2
12D4  les  di,ds:0A2A0h       ; THE3DMAP pointer
12DA  mov  ax,0100h
12E3  add  word [bp-8],0100h
12F2  les  di,ds:0A2A0h
12FC  mov  ax,0100h
1305  add  word [bp-8],0100h
1314  les  di,ds:0A2A0h
1318  add  di,0200h
1322  mov  ax,0100h
1327  add  word [bp-8],0100h
1336  les  di,ds:0A2A0h
133A  add  di,0300h
1340  mov  ax,0100h
1344  call far 0A65:1B0Dh
134D  mov  al,[bp+6]
1356  mov  es:[di+18Ah],ax
```

更完整的第一、二次 copy 也保留在同一 report；四次 call 的 destination 是
`THE3DMAP +000h/+100h/+200h/+300h`，每次 `0100h`，source 從成功解碼輸出的
`+002h` 依序前進。`0402h` gate、四平面 destination 與 selector 寫入
`active record +18Ah` 是 `exact bytes／call shape`；`+18Ah` 的正式欄位名仍是
`unknown`，不能因 selector 而命名成區域、座標或秘密門旗標。

`BLOCKCODE 017C:04DEh` 先將兩個座標參數限制在 `0..0Fh`，呼叫
`WALLCODE 017C:060Dh`。在 direction／mode byte 為 `6／4／2／0` 的分支中，
它以 `THE3DMAP` 指標加 `(y<<4)+x` 讀取：

- `+300h` 的 byte 依 `C0h／30h／0Ch／03h` mask 與 `6／4／2` shift 取出 detail。
- `+000h` 與 `+100h` 依 `F0h／0Fh` mask 取出相鄰 wall code。
- 讀取、mask、座標公式與 `retf 6` 是 `exact`；把每個 plane 命名成秘密門、
  普通門或 terrain 仍需既有 round-56 幾何證據與 writer／runtime trace。

`WALLCODE` 局部也直接讀 `DS:0BDF0h`，但該段沒有讀 `BDF1h`；這再次支持本輪
的分界：`BDF1/TEMPSEARCH` 是 overlay-02 的暫存／顯示狀態，而 map loader／
movement reader 使用的是 `THE3DMAP` 四平面。這仍然**沒有**找到 `wall=09` 的
第三平面 writer，也沒有證明 `SEARCH` 會把 detail `0` 改成 `1`。

因此 P0-2 的 loader／buffer／普通 movement reader 靜態子邊界可列為已閉合，
但正常秘密門路徑仍未完成；下一個窄工作是 `SEARCHREC` 的 member owner、
`MOVEPARTY` 的 `P/BASH/KNOCK/ENTER` 交易，以及該交易是否寫回 `THE3DMAP` 的
第三平面。不能直接把 `LOAD3DMAP` 的 copy 當成 writer。

## xref 與負面結果的界線

本輪新增的 resident IDA script 以 baseline `.i64` 的原始關係查詢
`028080h／028081h`。resident `BDF0` 有一個 direct data reference（IDA EA
`018ADBh`，`mov al, byte_28080`）；resident `BDF1` 沒有 direct reference。
這只能表示 resident database 中的 direct xref 數量，不包含 overlay code 的
runtime segment、指標／暫存器間接讀寫，因此不能寫成「BDF1 沒有 consumer」。

跨 extracted PC-98 overlays 的 byte-pattern 候選只看到：overlay-02 的上述兩個
`BDF1` 讀寫，以及 overlay-11 的初始化；這個 raw scan 仍只是候選 inventory，
正式結論以 overlay-02 的連續 code 與 TPOV `SHOWLOCATION` resolver 為準。

## 對 P0-2 的影響

本輪**沒有**關閉 `(13,10)`→`(8,15)` 的秘密門路徑，也沒有新增 `secret_door`
JSON、movement 特判或直接座標注入。可以移除的錯誤路由是：

- 不再把 `BDF1/TEMPSEARCH` 當成第三平面 map writer 的先驗入口。
- 不再把 `S`→`SHOWLOCATION` 或暫時 `+594h=1` 寫入當成「秘密門已開」。
- 不再因 `SEARCHREC`、`SECRET`、`HIDDEN` 名稱存在，就替 `wall=09/detail=0`
  建立可走規則。

下一個窄工作改為：在已閉合的 `LOAD3DMAP／BLOCKCODE` reader 邊界之外，追
`MOVEPARTY (00C9:0BCCh)`、`SEARCHREC` type/member owner 與第三平面 writer
之間的 writer→projection→movement consumer；仍須以同版本 DOS／PC-98 runtime
trace 證明 map state 變更，才可實作正常玩家路徑。

## 驗證命令與結論

在 Docker 內以 IDA Pro 9.4 disposable copy 執行：

```text
cp PC98-GAME.EXE.i64 /tmp/pc98-search.i64
idat -A -Sscripts/ida/pc98_search_state_xref_audit.idc /tmp/pc98-search.i64
idat -A -Sscripts/ida/pc98_load3dmap_dataflow_audit.idc /tmp/pc98-overlay30.bin
idat -A -Sscripts/ida/pc98_overlay2_search_state_audit.idc /tmp/pc98-overlay02.bin
idat -A -Sscripts/ida/pc98_overlay11_bdf1_audit.idc /tmp/pc98-overlay11.bin
```

結果重現了 `BDF1` 的 `A2F1BD` writer、`A0F1BD` reader、`014A:00DEh` far stub
與 INIT 清零，且沒有對原始 binary／baseline `.i64` 寫入。這輪是 `SAMPLED`
的靜態反組譯 boundary，不是完整 runtime gate。
