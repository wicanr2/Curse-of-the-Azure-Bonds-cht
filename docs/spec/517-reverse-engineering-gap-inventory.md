# 第 517 輪：逆向缺口盤點與 GEO 斷言勘誤

狀態：`READY`（範圍盤點；不是新的遊戲規則）
日期：2026-08-09

## 目的

第 516 輪已完成秘密門／搜尋資料流的窄抽樣，但沒有找到可以支持新規則的
第三平面 writer。本輪把「還差多少反組譯」改用玩家可觀察的資料流計算，並
刪除一個容易誤導後續工作的 GEO 斷言。這份盤點不把 `READY` 規格數量當成
反組譯完成百分比，也不要求把每一行 `START.EXE`、`GAME.OVR` 或 PC-98
overlay 都翻譯成中文。

## 第 517 輪新增 DOS overlay 候選

在 Docker 內以 IDA Pro 9.4 disposable database 分析 DOS extracted
`overlay-22.bin`（SHA-256
`980714e67b1f9a0077404fbb4144ac3f2d48d688aec8e74f4318c23ad9829c9b`；完整
`GAME.OVR` SHA-256 `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215`），
使用 `scripts/ida/dos_overlay22_4bf0_dataflow_audit.idc`，得到下列**原始
overlay-local 指令證據**：

| offset | raw／IDA 指令 | 可證明的最小語意 | 等級 |
|---|---|---|---|
| `0x0969..0x0970` | `mov [di+4BF0h],cx`；`mov [di+4BF2h],bx` | 以 `di` 為索引基底的 4-byte table candidate writer | `exact bytes／unknown meaning` |
| `0x099D..0x09A3` | `xor ax,ax`；`mov [di+4BF0h],ax`；`mov [di+4BF2h],ax` | 以 `di= index<<2` 清除多筆同形 table candidate | `exact bytes／unknown meaning` |
| `0x03CF..0x03D5` | `les ax,[di+4BF0h]`；`mov ds:729C／729E,ax／dx` | 以同一 indexed table candidate 讀取 far pointer | `exact bytes／unknown meaning` |
| `0x0955..0x0979` | 讀 object record、呼叫 local `0x03F5`，再寫入 table candidate | writer 的局部 caller／producer 形狀 | `strong inference` |

這些 bytes 證明的是「overlay 22 在其 DOS 資料段位址空間有一個以
`[di+4BF0h]` 為候選基址的 indexed far-pointer table」，**不證明它就是 ECL
scalar work address `4BF0h／4BF1h`**。同名數字可能分屬 resident data、ECL
work memory 或不同載入基址；在 DS／segment、table owner、`CALL 2E10h` 的
projection 與 runtime consumer 尚未閉合前，不能把這個新 writer 直接接到
`set_map_position`，也不能用它命名地圖座標。

## 目前還差多少

以「尚未閉合的行為資料流」為單位，目前有 **11 個直接影響正常玩家結果的
逆向主題**，另有 **4 個以畫面／音訊／發行為主的 fidelity 主題**。這是工作
流數量，不是函式數量；每個主題可能要由 ECL bytes、IDA database、原版執行
追蹤與 game-pack 測試共同閉合。

### 11 個逆向主題

| 群組 | 數量 | 尚缺的閉合內容 |
|---|---:|---|
| P0 外部地圖／正常路徑 | 3 | 火刀戰後目的地 producer→`CALL 2E10h` redraw／位置 consumer、騎士後秘密門／搜尋 bridge、block 4 後續地圖與返回世界 handoff |
| ECL 與外部 routine | 4 | `CALL／NEWECL／LOAD FILES／LOAD PIECES／PROGRAM／COMBAT／TREASURE／INPUT STRING` 的高影響 producer／consumer；work address 跨位址空間索引；random／FLEE／NPC／物品／旗標邊界 |
| 戰鬥規則與 AI | 2 | 敵我選敵與 Quick AI caller chain；近戰、弓箭、法術、持續效果、音效與戰後 handoff 的逐項演出／規則橋接 |
| 存檔與 AD&D 角色規則 | 2 | `SAVGAM?.DAT` 完整欄位及 sidecar；角色／怪物／法術 consumer、年齡、職業限制、alignment、特殊能力與戰後恢復 |

### 4 個 fidelity／發行主題

1. DOS 地城、城市、AREA、WILDERNESS 與 combat 的幾何／sprite／frame timing
   逐狀態校準。
2. PC Speaker、Tandy、PC-98 YM2203／MSCDRV 的真實 runtime 音訊與完整 producer。
3. 完整正常玩家路徑、跨存檔／長時間回歸與三平台打包。
4. 原版忠實 theme 的逐畫面稽核，以及日後可切換的獨立美化 theme。

因此目前不應報告「還剩 N 個函式」或「反組譯已完成 X%」。可靠的說法是：
格式與大量窄行為已可重播，但仍有 3 個 P0 路徑 handoff、8 個 P1 行為主題
（ECL 4、戰鬥 2、存檔／規則 2）尚未閉合。

## 第 517 輪證據勘誤

### GEO 斷言

原本「`(1,8)`、`(13,10)`、`(8,15)` 不在同一 GEO component」的寫法過度
延伸了靜態圖分析。正確表述是：

- 在 `GEO2.DAX` block 3 的**目前第三平面／門狀態未變更圖**中，從 `(13,10)`
  到 `(8,15)` 沒有符合目前 movement predicate 的路徑；這是 geometry fact。
- 若在診斷器中暫時把 `wall=09` 邊視為已開啟，會得到一條經過
  `(10,12) → (9,12)` 的候選路徑。這只能指出值得追查的秘密門候選，不能證明
  原版 `SEARCH`、`S` 或某個 ECL flag 會開啟它。
- 因此不能把 static BFS 失敗寫成「地圖永久不相連」，也不能把候選邊直接
  改成可走；`MoveDungeon` 目前維持 fail-closed 是正確的暫態行為。

### `S`／搜尋斷言

PC-98 overlay 14 的連續 bytes 目前只精確證明：輸入 `S` 會切換目前角色
record `+594h` bit 0，然後呼叫 TPOV `014A:00DEh`；resolver 將該 stub 解為
overlay 24 `SHOWLOCATION` local `0x2E8Ch`。它沒有精確證明秘密門 writer。
`BLOCKCODE` 讀取第三平面 detail `0..3`、detail 0 不是普通可走格，也不等於
已找到 `S → detail 1` 的橋。

CoAB remake 的 bounded probe 在已知候選位置呼叫 `SearchDungeonLocation` 時，
觀察到的是 `ModeEvent` 與 `PRESS BUTTON OR RETURN TO CONTINUE.` 邊界；這是
remake runtime 的 entry observation，不是原版秘密門語意，也不計入正常路徑
完成證據。probe 留在 `workplace/` 供後續追查，不能移入正式 regression。

## 第 517 輪當時的下一個反組譯順序

1. 先追 DOS ECL2 block 3 `CALL 2E10h` 的 redraw／位置 consumer，以及它之前
   的目的地 producer；並分開解析 ECL work address 與 overlay 22 的
   `[di+4BF0h]` indexed table。要找的是 producer→projection→consumer 的完整
   橋接，不是先假定兩者同址。
2. 若 DOS writer 仍只有間接取址，才回到 PC-98 `BDF1` caller／address-taken
   map service，並使用原始位址空間分欄；不得把 `BDF1`、`+594h`、`+300h`
   對映成 DOS `4C28h`，也不得把 overlay `[di+4BF0h]` 自動對映成 ECL
   `4BF0h／4BF1h`。
3. 只有 writer、movement consumer 與 runtime state 同時閉合，才建立中立
   `secret_door`／`search` JSON contract；在此之前不新增 movement 特判。

本輪沒有新增遊戲規則，亦沒有把 direct-entry 或 coordinate-assisted probe
勾成正常玩家路徑。

## 第 519 輪更新

上述第 517 輪的第一項已完成一個**靜態 dispatch 子邊界**，但不是完成 P0-1：
`ECL 2E10h → selector AE11h → START control vector 017F:003Eh →
overlay-30 local 07C6h` 已由 raw control block 與 extracted `GAME.OVR`
offset 閉合。`07C6h` 只精確到兩個 word 參數、16×16 bounded index、
`DS:7206h` far pointer 與 `ES:[DI+0200h]` byte read；map plane、座標 writer、
`C04B..C04F` projection 與 runtime consumer 仍是未知。

因此目前反組譯路由改為：

1. 追 `DS:7206h` far pointer 的初始化／owner 與 overlay relocation。
2. 追 `DS:720F／7210／7213` 及 vector 4 的 `DS:7211／7212` writer／consumer，
   不跨位址空間合併相同十六進位數字。
3. 回到 `C04B..C04F` 與 DOSBox 原版 runtime，閉合目的地 producer → map
   projection → redraw／位置 consumer。
4. 只有正常輸入、writer、movement consumer 與 runtime state 同時閉合後，才
   重評估 `secret_door`／`search` JSON；本輪不新增 movement 特判。

完整證據見 [`spec 519`](./519-dos-overlay-vector-to-cell-layer-accessor.md)。
