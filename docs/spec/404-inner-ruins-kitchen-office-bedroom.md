# 404：內部遺跡廚房、辦公室與臥房

狀態：`READY`

## 問題

第 403 輪已從外圍遺跡下水道正式進入 ECL／GEO `43h` 的
`(15,15,N)`，但尚未證明玩家能在這張 26 種 terrain 的內部地圖繼續探索。
本輪先處理出生點附近不會提前觸發最終主線的三個房間：

- terrain `8Ch`：廚房；
- terrain `8Bh`：辦公室；
- terrain `8Ah`：豪華臥房。

## 原始資料

- `ECL6.DAX`
  - SHA-256：
    `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb`
  - decoded block `43h`：5,363 bytes。
- `GEO6.DAX`
  - SHA-256：
    `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c`
  - decoded block `43h`：1,026 bytes。

所有分析與測試均在 `--network none` 的 Docker 容器內執行；原始 ZIP
唯讀使用。

## GEO 座標與正常路線

GEO6 block `43h` 的相關格：

| Terrain | 座標 |
|---|---|
| `8Dh` | `(15,15)`，下水道出生格，本身不顯示 terrain 事件 |
| `8Ch` | `(13,14)` |
| `8Bh` | `(11,12)`、`(12,12)` |
| `8Ah` | `(9,12)`、`(10,12)` |

由 `(15,15)` 起的合法路線：

```text
(15,15) → (15,14) → (14,14) → (13,14)  terrain 8Ch
         → (13,13) → (13,12) → (12,12)  terrain 8Bh
         → (11,12) → (10,12)             terrain 8Ah
```

每一步均由 `geo.Grid.CanMoveDungeonWrapped` 驗證雙側牆面，不是直接指定
事件 entry。

## Raw ECL 行為

### terrain 8Ch：廚房

- `+0DEAh` 寫 `4C06h=1`。
- 初次且 `4C06h!=1` 時顯示：廚房裡的奴隸躲到桌下，對隊伍沒有威脅。
- 按鍵後離開；重訪不再顯示文字。

### terrain 8Bh：辦公室

- `+0D8Ch` 寫 `4C05h=1`。
- 初次且 `4C05h!=1` 時顯示：房間被改為辦公室，牆上滿是班恩聖徽。
- 按鍵後離開；重訪不再顯示文字。

### terrain 8Ah：豪華臥房

- `+0D01h` 寫 `4C04h=1`；旗標在詢問前寫入，因此拒絕搜刮也會消耗事件。
- 選 `NO` 直接離開，不給財寶。
- 選 `YES` 發出沒有怪物的 `COMBAT` 財寶服務邊界，exact TREASURE：

  ```text
  [0,0,0,5000,5000,12,15], ItemBlock FFh
  ```

  即 5,000 GP、5,000 PP、12 gems、15 jewelry；以本作幣值換算共
  30,000 gold，不產生 ITEM6 隨機物品。

## 跨 block 旗標碰撞

單獨從乾淨 ECL session 進入 block `43h` 時，廚房與辦公室文字都會正常
顯示。但 Standing Stone 起始的「走遍已完成支線」玩家路徑在抵達前已經：

- 於 block `40h:+0F52h` 寫 `4C05h=1`；
- 於 block `42h:+0F70h` 寫 `4C06h=1`。

block `43h` 再於 `+0D8Ch／+0DEAh` 使用相同絕對地址。`SAVE` 寫的是
SAVGAM ECL 全域記憶體，目前沒有 `LOAD FILES` 建立 block-local bank 的
bytes 或 runtime 證據。因此完整前置路徑會忠實地讓辦公室與廚房 one-shot
靜默；remake 不得擅自清零、換頁或重播。

這不代表文字無用：較短、未觸發相同全域旗標的合法玩家路徑仍可看見；
raw regression 也必須保存兩段翻譯與各自的一次性行為。

## 實作與驗證

- game-pack 新增三個 stable message ID 與繁中：
  - `myth-drannor.inner.kitchen`
  - `myth-drannor.inner.office`
  - `myth-drannor.inner.bedroom`
- `TestRealInnerRuinsKitchenOfficeAndBedroom` 直接執行原始 ECL bytes，驗證
  三個 writer、重訪、臥房 YES／NO 與 exact 財寶。
- `TestRealPlayerPathStandingStoneToBurialGlen` 從正常遊戲入口沿合法 GEO
  延伸至 `(10,12)`，驗證：
  - `4C05／4C06` 跨 block 碰撞造成的原版靜默；
  - 臥房繁中詢問；
  - 30,000 gold、12 gems、15 jewelry；
  - 財寶服務後回到地城探索。

## 完成邊界

本輪只完成 block `43h` 出生點附近三個房間。犬舍、活動雕像、私人禮拜堂、
祭司宿舍、魔法圓、樓梯、二樓、提朗瑟克斯最終事件、結局，以及所有相關
怪物能力、動畫、法術與音效仍未完成。

