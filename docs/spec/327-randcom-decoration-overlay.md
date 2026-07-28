# 第三百二十七輪：RANDCOM 地城戰鬥裝飾

狀態：`READY`

## 已有 placement 證據

`mapdata.GenerateDungeonSeeded` 已按 reference `sub_370D3` 還原桌椅 pass：

- 只有 GEO terrain `& 0x40`、牆面組合允許且目標為開放 floor `0x16` 時才嘗試。
- 原版 dice stream 決定 table 與周圍 chairs；remake 以顯式 seed 保持 replay。
- table 寫入 BackgroundTiles entry `0x1A`，chair 寫入 entry `0x1B`。

先前 renderer 只把 `BackgroundTile.TileIndex` 查入 DUNGCOM。entry `0x1A/0x1B`
解出的 graphic ID 是 `0x22/0x23`，超過 DUNGCOM 的 25 張，因而被略過。

## Atlas namespace

| 全域 graphic ID | atlas lookup |
|---:|---|
| `0x00..0x18` | `DUNGCOM[0..24]` |
| `0x22..0x27` | `RANDCOM[0..5]` |

RANDCOM 是透明 decoration，不是 floor。地城 entry 落在 `0x22..0x27` 時，
renderer 先畫開放 DUNGCOM floor `0x16`，再畫 `RANDCOM[id-0x22]`。palette
zero 已由 combat tile codec 轉成透明 sentinel 16，所以桌椅外部不會蓋成方塊。

WILDCOM 保持自己的 `0..33` namespace；不能把 `0x22` 的地城語意套用到野外。
`0x19..0x21` 也不能猜成 DUNGCOM 或 RANDCOM。

## 驗收

- unit tests 鎖定 DUNGCOM 最後一張 24、WILDCOM 最後一張 33、
  RANDCOM `0x22→0`／`0x27→5`，並拒絕 namespace gap `0x21`。
- 原始 GEO catalog 掃描在 seed 1 找到
  `GEO2 block 0x01, center (13,0)` 的可見 table／chair。
- `docs/screenshots/gold-box-layout-combat-randcom.png` 由正式 Ebiten command
  在 Docker/Xvfb 擷取；下方可見桌與椅透明疊在 DUNGCOM 開放地板。

## 邊界

本輪接通圖像 namespace 與已還原的原版 placement，不宣稱 Battle 移動已消費
`BackgroundTile.MoveCost`。terrain cost／impassable collision 必須再將 camera
座標、DungeonFloor 與 Battle movement transaction 串接後驗證。
