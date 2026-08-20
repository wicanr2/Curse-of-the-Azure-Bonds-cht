# 第五百四十三輪：正常主線 coverage 盤點與 IDA 地圖 cell 稽核

> 更新（第 547 輪）：本檔保留的 `C04B..C04F` map projection gap 已由 DOS
> overlay-07 原始 setter／getter bytes 補齊；目前的 Cave E1 anchor 是
> `(5,7,W)`，不是本檔歷史記錄的 `(4,5,N)`。完整勘誤見
> [`spec 547`](./547-normal-beholder-cave-presentation-state.md)。本檔仍不宣稱
> DOS redraw／完整地城 coverage 已完成。

狀態：`READY`（本輪只封存正常主線的 Hap／熔岩洞／法師塔 coverage，不宣稱全城市、全地城或完整結局）

## 本輪結論

同一個新遊戲 ECL session 現在可由第 542 輪的艾森布拉城外正常走到 Hap，沿
GEO5 block `0x32` 逐格完成民宅／阿卡巴／旅店／伊弗利特解放，經外部出口進入
熔岩洞，打贏守門巡邏，走到熔岩池，經巫師塔入口、德拉坎德羅斯與黑龍事件返回
洞穴，再完成「友善交涉→同格重訪→15 隻火蜥蜴→防火桶耐熱失敗」分支。測試使用
`internal/game/campaign_normal_test.go` 的 `TestRealNewGameRunsToTheEnding`；
選項由 game-pack stable ID 解析，移動由 `MoveDungeon` 執行，沒有注入劇情旗標或
直接切入戰鬥。這是本輪第一條把 Hap／熔岩洞／法師塔接在同一 session 的正常
玩家路徑證據。

本輪也修正一個會破壞巫師塔回洞穴路徑的引擎邊界：ECL 同區塊重繪前寫入
`C04B/C04C/C04D` 的暫存座標，不應覆蓋已有地城游標。`original.geo5.block-33`
在 JSON 宣告 `spawn=(7,15,W)` 作為目的地錨點；可重用 engine 以「有 spawn 的
地圖」保留 live cursor，沒有把 Hap 座標寫死在 State。外部出口的 `roof_type=2`
也由 engine schema／CoAB JSON 宣告，讓 `(15,5,E)` 的正常邊界交易保留原版
presentation selector。DOS `C04E/C04F` 的完整 producer→buffer→SearchLocation
資料流仍未達 `exact`，所以不把這項修正擴大成全 GEO plane fidelity 完成。

既有固定夾具另外覆蓋希爾斯法、尤拉什、摩安德之坑、散提爾堡與 Myth Drannor 的
許多事件，並通過終戰／`PROGRAM 8` 局部驗證；固定夾具會直接提交 ECL terrain／
work memory，不能替代「新遊戲一路到結局」。

## 本輪正常玩家路徑邊界

| 階段 | 正常 session 證據 | 結果 |
|---|---|---|
| Hap 城外→村落 | 第 542 輪世界 handoff 後，以 GEO5 block `0x32` 逐格移動 | `hap.leave`、`hap.map-route` 與村落主要故事分支接通 |
| 村落→熔岩洞 | JSON external exit `(15,5,E)`，`wall_type=7`、`roof_type=2`，正常 `MoveDungeon` 邊界交易 | ECL5 block `0x32`、`LOAD PIECES [8,FF,FF]`、入口／伏擊接通 |
| 熔岩洞→巫師塔 | `(6,15,W)` 的 ECL5 per-turn handoff，JSON map anchor 保留 `(7,15,W)` | block `0x33` courtyard、arrival、WAIT narrative、攻擊巫師、守住屋頂與 CAVES 回程接通 |
| 熔岩池重訪 | 第一次 `WAIT→PARLAY_NICE` 回地城；同一 cell 再跑 per-turn，選 `COMBAT` | 15 隻火蜥蜴、火防桶、WHO、耐熱失敗與拒絕重試回地城接通 |

「完整主線」仍未宣稱：本測試在防火桶後尚未沿同一 session 接回尤拉什／摩安德
之坑／散提爾堡／Myth Drannor 結局；城市全房間、所有分支、完整存檔與重訪也仍
在 `WORKLIST.md`。

## ECL5 block 0x31 的靜態條件

以下是原始 DOS ECL 的 payload offset；`ECL5.DAX` block ID `0x31` 在這裡以
decimal `49` 表示。這些是 ECL 位址空間，不可和 overlay local offset、GEO
座標或檔案 offset 混用。

| payload offset | 原始操作 | 目前可證明的結果 | 等級 |
|---|---|---|---|
| `+04A5..+04B2` | 讀／比較 `C04F`，再寫 `7EC6` | SearchLocation 先做 terrain／wall selector 的前置判斷 | `exact`（ECL bytes） |
| `+04B7..+04C0` | `COMPARE [4BC9],0Eh`、`IF <=` | `4BC9 <= 14` 會跳過後續事件；不得替它命名成時間或房間編號 | `exact`（writer 語意仍未知） |
| `+0507..+0513` | `RANDOM 99 → 7F79`、比較 `10` | 哈普巡邏不是每次進 cell 都必然出現；seed／持續亂數位置會改變結果 | `exact` |
| `+0518..+051F` | `COMPARE 7ECA,1` | 一次性 LOOK／SEARCH dispatch 與一般 per-turn 呼叫不同 | `exact` |
| `+052C..+0534` | 比較並寫入 `4C02` | 民宅／後續哈普故事共用 visited gate；必須保存同一 ECL session | `exact` |
| `+0B1A..+0B50` | `4C02`、`4C5E`、`4C5F` gate；寫 `7F7A`／`4C5F`；PICTURE `3B` | 阿卡巴分支受「民宅旗標、伊弗利特旗標、阿卡巴相關旗標」共同決定，不能只看 `C04F=8A` | `exact` |
| `+0B59..+0BC1` | 阿卡巴文字、加入選項與 `ADD NPC 3B` | NPC／PICTURE／選單／ADD NPC 必須以同一 continuation 串起 | `exact` |

ECL5 block `0x32` 的 SearchLocation 在 payload `+05B5` 將 `C04F & 0x7F` 送入
`ON GOTO`；terrain `0x89` 的熔岩池敘事由 per-turn entry 觸發，Search entry
只留下 redraw `CALL`。因此正常測試必須先讓圖片／PRESS 邊界續跑，再選
`COMBAT／WAIT／FLEE／PARLAY`，不能只驗證某個 terrain byte 命中。

這也解釋為何既有固定測試可以在同一位置先把 `DungeonWallRoof` 設成 `0x80`、
再設成 `0x8A` 而通過：它是在測 ECL handler，本身不是 GEO 移動的證據。

## IDA Pro 9.4 非破壞性證據

本輪使用授權的 `ida-pro-9.4-ver2:uidfix-v1` Docker image，將原始 overlay 複製
到一次性分析目錄後執行 `scripts/ida/dos_overlay30_vector_audit.idc`。原始
`GAME.OVR` 與工作區檔案沒有被寫入；IDA database 只存在於容器／暫存副本。

| 項目 | 識別資訊 |
|---|---|
| 輸入 | `workplace/ida406/overlays/overlay-30.bin`；SHA-256 `444893f6d239cc57f555287786e3704bc801dd63f3d1f4d4ac14e8742652d468` |
| script | `scripts/ida/dos_overlay30_vector_audit.idc`；SHA-256 `10f49e0799e9925d5407ac3569b1c2ec1b4fa75812be5bd2c08560092cab1f10` |
| 報告 | `/tmp` 稽核輸出；SHA-256 `907c6ab3d992ed7cf45d764ce6a3e14a0e9ca2dd2e94a529145159a48afb25cd` |
| 位址基準 | DOS `GAME.OVR` 的 overlay-30 local offset；IDA segment `0x0000..0x147F` |
| control vector | START `017F:003E`，zero-based vector index 6，目標 local `0x07C6` |

IDA 報告保留原始 bytes／原始 IDA 名稱與完整連續指令，不做 rename。可直接回查
的 exact 片段如下：

```text
07C6  55 89 E5 83 EC 02       push bp; mov bp,sp; sub sp,2
07CC  8A 46 08 ...             讀 [bp+8]，再呼叫 local 0556h
07F0..0816                      將兩個 byte 參數限制到 0..0Fh
081A  ...                        x/y 轉成 word
0824  B1 04 D3 E0                x 乘 16
0828  C4 3E 06 72                讀 DS:7206h 的 far pointer
082C  03 F8 03 FA                加上 x*16 與 y
0830  26 8A 85 00 02             讀 ES:[DI+0200h]
0838..083E                      回傳 byte；retf 4
```

因此本輪能安全下的結論是：該 DOS 外部 routine 會把受限的兩個座標組合成
`16×16` cell index，從 `DS:7206h` 指向的緩衝區讀取 `+0200h` plane。它**不能**
單獨證明 `C04F` 的正式欄位名稱、所有四個 plane 的順序、GEO terrain 與
ECL room selector 的一對一映射；後三項仍是 `strong inference／unknown`，必須
再接 producer→buffer→C04F→SearchLocation 的 runtime trace。

這項證據與 PC-98 overlay-22 的同輪稽核互相獨立：PC-98 報告確認閃電效果 routine
在 local `0x5F70..0x6007` 逐格走 target pointer、呼叫傷害／效果 helper；它只支持
戰鬥演出與傷害橋接，不可被擴大成完整弓箭／法術／音效完成證明。PC-98 輸入
`workplace/ida406/pc98-overlays/overlay-22.bin` SHA-256 為
`c54729525d576c11d731d64a1b06ee2547b2562b73e3708a1beaafc535cabbe8`，稽核 script
為 `scripts/ida/pc98_monster_lightning_audit.idc`，SHA-256
`a0e5eb77edc686e49f7370b9d6504885fc2ea783f6dc7c056fb239c95c6073ee`。

### START `4BC9` direct-xref audit

本輪另以同一 IDA Pro 9.4 Docker image 稽核 `START.EXE` 的 `4BC9` 直接 resident
data xref。原始執行檔、基準 `.i64` 與資料庫均保持唯讀；報告只依 IDA 建立的
direct data xref，不能涵蓋經指標／暫存器的間接 writer。

| 項目 | 識別資訊 |
|---|---|
| 原始輸入 | `workplace/ida406/START.EXE`；SHA-256 `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` |
| 基準 database | `workplace/ida406/START.EXE.i64`；分析副本 SHA-256 `9df802ee4ef71fb2eda83257e0ed2d87adf0ee2d10241d3bdbdc6bc369fe47eb` |
| script | `scripts/ida/dos_start_4bc9_writer_audit.idc`；SHA-256 `e58c91df8c5109772d6e2158853af99a2171b4977afba95eaa63205e925986c3` |
| 位址基準 | IDA DSEG linear EA；`DSEG=0x1C1C0..0x250C0`，DS-relative `0x4BC9` → linear EA `0x20D89` |
| 結果 | direct data xref count `0`；`4BC9` 欄位 producer／正式名稱維持 `unknown` |

這個「direct xref=0」只排除 IDA 已建立的直接 data reference，不能推論沒有
writer。正常 remake 的 `4BC6..4BCC` clock mirror 與 `4BC9→hour` 仍標為
`strong inference`，因為 ECL 分支、runtime clock transaction 與公開
cross-implementation mapping 一致；若要升為 `exact`，仍需完成間接 producer→
work memory→consumer trace。

## 全世界／城市／地城 coverage 基線

下表區分「正常 session」、「固定 ECL fixture」與「只有資料宣告」。只有第一類
可以支持從玩家移動抵達；第二類可以支持局部 opcode／事件；第三類不支持可玩性。

### 城市與區域

| 區域 | game-pack／原始區塊 | 目前證據 | 狀態 |
|---|---|---|---|
| 提爾佛頓 | ECL2、GEO2 block 1/3/4 | 下水道→E2→火刀據點、城市設施與公會多為正常或固定切片 | 部分完成 |
| 阿沙本福德 | ECL1 world handoff、城市 service | 同一新遊戲 session 進城、河畔酒館、Tavern Tale 28、離城 | 正常主線已驗證 |
| 艾森布拉 | ECL1／城市 facility | 同一新遊戲 session 到城外；城市場所另有固定事件覆蓋 | 城外正常；城內非完整 session |
| 哈普／哈普圖斯 | ECL5 block `0x31`、GEO5 `0x32` | 同一正常 session 已完成村落主要路徑、熔岩洞、巫師塔回程；全部村民事件、可選房間與重訪仍未完 | 正常主線局部完成 |
| 希爾斯法 | ECL1 world handoff、ECL3／GEO3 | 火刀偽裝伏擊、碼頭／城市事件有固定 fixture | 非正常完整 session |
| 尤拉什 | ECL3 `0x10/0x11`、GEO3 | 紅羽衛、等候室、間諜、外部 block handoff 有固定／局部正常測試 | 非完整城市 coverage |
| 散提爾堡 | ECL4／GEO4 `0x20/0x21/0x25` | 內城、衛兵、神殿／beholder 部分固定測試 | 非完整城市 coverage |
| Myth Drannor | ECL6／GEO6 `0x40/0x42/0x43/0x45` | Burial Glen、Outer／Inner Ruins 與終戰固定／座標輔助測試 | 尚未由新遊戲 session 串到結局 |

### 目前 game-pack 的 19 個第一人稱／區域 map declaration

| map ID | script block | geometry block | coverage |
|---|---:|---:|---|
| `tilverton.first-person` | — | `0x01` | 城市事件局部 |
| `tilverton.sewers.first-person` | `0x03` | `0x03` | 正常入口／固定房間 |
| `tilverton.fire-knife-hideout.first-person` | `0x04` | `0x04` | 正常至首領前 |
| `zhentil-keep.inner-city` | — | `0x20` | 固定 |
| `myth-drannor.burial-glen` | `0x40` | `0x40` | 固定／座標輔助 |
| `myth-drannor.outer-ruins` | `0x42` | `0x42` | 固定／座標輔助 |
| `myth-drannor.inner-ruins` | `0x43` | `0x43` | 固定／終戰局部 |
| `zhentil-keep.dark-shrine` | — | `0x21` | 固定 |
| `zhentil-keep.beholder-cave` | `0x22` | `0x25` | 固定 |
| `original.geo3.block-10` | `0x10` | `0x10` | 尤拉什固定／局部正常 |
| `original.geo3.block-11-level-1` | `0x11` | `0x11` | 固定 |
| `original.geo3.block-11-level-2` | `0x12` | `0x11` | 宣告／局部 |
| `original.geo3.block-15` | `0x15` | `0x15` | 宣告 |
| `original.geo5.block-32` | `0x32` | `0x32` | Hap／熔岩洞正常主線局部；全房間未驗證 |
| `original.geo5.block-33` | `0x33` | `0x33` | 巫師塔正常進入／回洞穴局部；全事件未驗證 |
| `original.geo5.block-35` | `0x35` | `0x35` | 宣告 |
| `original.geo6.block-45` | `0x45` | `0x45` | 結局區域固定局部 |
| `moonsea.overland` | — | `0x79` | 14 點 arrival／路網 baseline |
| `tilverton.area-map` | — | `0x01` | 區域圖資料宣告 |

「有 declaration」不等於「每個房間可走、每個事件有副作用、離開後可重訪」。
下一個 coverage artifact 必須逐格保存：原始 GEO cell、方向、C04E/C04F 投影、
ECL entry／resume PC、旗標前後、戰鬥／寶物／手札及離開再進入結果。

## 下一個可驗收工作

1. 依 IDA overlay-30 的 plane buffer evidence，繼續追 `MoveDungeon` 的
   `C04E/C04F` producer；目前只把 map anchor 與 JSON roof selector 用於已驗證
   邊界，不把 `grid.Cell.Terrain` 擴大宣稱成原版正式 room selector。
2. 以同一 ECL session 從防火桶後續接尤拉什、摩安德之坑、散提爾堡與 Myth Drannor，
   每個 handoff 都保留圖片／選單／戰鬥／旗標／重訪的正常輸入證據。
3. 完成全城市／全地城房間與結局 gate 前，維持固定夾具、coordinate-assisted 與
   normal session 分層；不能用既有 `PROGRAM 8` fixture 代替新遊戲到結局。
4. 每完成一個區域就更新 `WORKLIST.md`、`docs/project-status.md`、本表與
   coverage 測試；在全主線通過前，不做「完成 remake」或三平台／推廣片聲明。
