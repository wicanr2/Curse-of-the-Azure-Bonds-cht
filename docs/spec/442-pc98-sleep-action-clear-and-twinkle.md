# 第四百四十二輪：PC-98 睡眠術 Action 清除與 TWINKLE

狀態：`READY`

## 範圍與結論

本輪閉合動態 Sleep `35h` 在加入、受傷移除與自然到期時的 Action／演出順序。
結論只涵蓋下列已證明邊界，不代表完整法術動畫或 active battle 存檔完成。

1. `INITEFFPROX` 將 effect `33h／34h／35h` 的 add／remove callback 都設為
   overlay 12 local `0075h`；typed overlay stub 解析至 overlay 24 local
   `2A5Bh` `CLEARACTION`。它清除 Action 內四個欄位，不消耗 memorized slot。
2. `PUTDAMAGE` 先移除 Sleep，再進一般施法中斷。因此睡眠中的施法 action
   被 `CLEARACTION` 清除，不能再由通用 damage interrupt 重複消耗法術格。
3. `CLOCK_` 到期同樣經 `SPELLOFF`／callback；加入、傷害喚醒與自然醒來
   都會清除 pending spell、target、move 與 guard 狀態。
4. Sleep 成功字串是 overlay 22 local `2536h` 的 CP932 Pascal string
   「は眠りに落ちた。」。`PUTEFFECT` 成功寫入後呼叫 `TWINKLE(success=1)`；
   魔抗分支在此之前返回，所以抵抗者沒有 TWINKLE。
5. `TWINKLE` 在戰鬥狀態以動態圖示 `16h` 表示成功、`17h` 表示失敗；這兩個
   值不是 DAX block。overlay 24 local `1C76h` 以四次 writer 呼叫建立
   24×6 動態圖示。
6. 成功分支逐目標播放 `SPELLHITFX`。effect callback 本身沒有文字、圖形或
   音效呼叫，因此傷害喚醒與自然到期不可自行加入「醒來」閃爍或音效。
7. `GAMESPEED=4` 時，成功 TWINKLE 的 delay-only 時間是
   `(4+1) × 4 × (4×18ms) = 1440ms`；繪圖本身的耗時尚未計入。

## 實作契約

- `Battle` 在 effect `35h` 成功加入、傷害移除與自然到期時共用
  `clearActionState`，並保持「先移除 effect，再判斷一般 damage interrupt」。
- `VisualTwinkle` 是作品中立的逐目標呈現 transaction；每名成功目標使用
  1440ms，抵抗者不進清單。State 先發施法聲，再於各目標發 spell-hit 聲。
- renderer 依 IDA 證據使用四格、24×6 native geometry；目前顏色／圖元因
  尚無 PC-98 runtime capture，標記為 `layout-reconstructed`，不得宣稱
  pixel-exact。

## 證據等級

- callback table、typed stub、`CLEARACTION` 欄位寫入、`PUTEFFECT→TWINKLE`、
  `SPELLHITFX`、四格 writer 與 delay loop：`exact`（IDA 指令、bytes、
  Borland symbol/type 與既有 runtime／raw bytes 交叉驗證）。
- remake 四格顏色與圖元：`layout-reconstructed`。
- draw-call overhead、PC-98 實機 palette-cycle 像素：`unknown`。

## 非破壞性工具紀錄

`scripts/ida/pc98_held_action_clear_audit.idc`、
`pc98_sleep_lifecycle_audit.idc` 與 `pc98_monster_affect_loader_audit.idc` 只在
唯讀來源的 `/tmp` 副本建立資料庫與 ledger；沒有改名、patch 或覆寫原始
overlay。報告必須保留 local offset、原始 bytes、反組譯文字與上述證據等級。

| 證據 | SHA-256／規模 |
|---|---|
| overlay 12 | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` |
| overlay 22 | `c54729525d576c11d731d64a1b06ee2547b2562b73e3708a1beaafc535cabbe8` |
| overlay 23 | `a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9` |
| overlay 24 | `0b5d1bebb367414fac93287449a230a36355c075efdc30640a4994bc11d5cd5f` |
| held-action IDA 腳本 | `c0d3f4e3a9bebad87dde6fbf4a30b25064869511d8ffed44090523062bb75522` |
| Sleep lifecycle IDA 腳本 | `b0c5a80cf1675676524d0cc32951baf9d0a373cb76272caff2cec1c1b6b21891` |
| monster affect IDA 腳本 | `f6c5bc1b9b4d88d984d5d93c37c912253b95cc45f46bfa080b566b6410426482` |
| overlay 24 ledger | `05c349a4e5fd0926763c3d5f03bc21e59726f68c180de16c4860138d2203d9bb`／351 行 |

## 尚未完成

- PC-98 runtime 擷取，用來確認動態圖示的確切 palette／pixels 與 draw overhead。
- active battle save round-trip，包括 effect linked list、Action 與 scheduler。
- Quick Sleep、固定原版戰場 wall／corner 動態及 DOS 平台同效果逐幀對照。
