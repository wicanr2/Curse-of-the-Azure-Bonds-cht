# Release 圖像 PNG 獨立化盤點（2026-08-27）

狀態：盤點完成；轉換可行，runtime 尚未完全切換。

## 目標與邊界

- 原版 `curseoftheazurebonds.zip` 與其中 DAX 只作一次性 export oracle。
- release 使用可版控、可驗雜湊的獨立 PNG；需要索引、動畫座標或牆面組合資料時
  另配 JSON manifest，不把語意藏回二進位封裝。
- 本表只盤點圖像。ECL、GEO、ITEMS、MON* 等非圖像資料目前仍由原版 ZIP 載入，
  是否改成獨立 game-pack 是下一階段，不可把「圖像 PNG 化」寫成「已完全脫離原版」。

## 已獨立為 PNG

`assets/` 目前有 **787 張 PNG**：`assets/sprites/` 780 張、Journal 6 張、原版參考
截圖 1 張。runtime 已直接從 `assets/sprites/` 載入下列玩家可見素材：

| 類別 | PNG 數 | runtime 狀態 |
|---|---:|---|
| CPIC1..6 戰鬥角色／怪物圖示 | 156 | 直接讀 PNG |
| SPRIT1..6 遭遇動畫 | 138 | PNG＋`animation.json` |
| PIC1..6 場景動畫 | 152 | PNG＋`animation.json` |
| CHEAD／CBODY／party 合成 | 202 | 直接讀 PNG |
| HEAD2..6／BODY2..6／場景人物合成 | 102 | 直接讀 PNG |
| COMSPR 投射物／戰鬥符號 | 26 | 直接讀 PNG |
| BIGPIC1／2／6 | 4 | 直接讀 PNG |
| Journal 插圖／地圖 | 6 | 直接讀 PNG |

上述數字依檔名前綴計數；合成圖與來源 layer 都保留，因此不是「原版唯一圖塊數」。
來源與每張 block／item／frame 已記在 `assets/sprites/README.md`。

## 尚在 runtime 從原版 DAX 解碼的圖像

| 類別 | 現行 loader | PNG 化 | 額外資料 |
|---|---|---|---|
| `TILES.DAX` 地圖磚 | `loadTileImages` | 直接逐 item 匯出 | block／item → 線性 tile index manifest |
| `DUNGCOM`／`WILDCOM`／`RANDCOM` | `loadCombatTerrainImages` | 直接逐 24×24 tile 匯出 | source＋tile index manifest |
| AREA map 符號 | `loadAreaMapSymbols` | 直接逐 8×8 item 匯出 | symbol file／block／item |
| 第一人稱共用符號 | `loadSymbolBlock` | 直接逐 8×8 item 匯出 | group、first_id、item count |
| `SKY.DAX` | `loadSkyImages` | 每個使用 block 匯一張 PNG | block ID 對應表 |
| `WALLDEF*.DAX`＋`8X8D*.DAX` | `loadMapPieceSets` | 8×8 symbol 可匯 PNG | **還需 JSON 保存 WallDef 156-byte cell、selector、record、symbol block 與全域 ID band** |

結論：**全部玩家可見圖像都能轉為 PNG 獨立存放。**唯一不能只靠裸 PNG 的是
第一人稱牆片：PNG 保存像素，JSON 保存牆型如何以 8×8 symbol 組合；兩者合起來
才能取代 `PieceSet`，不能把 contact sheet 當 runtime atlas。

## 現有工具與缺口

- `scripts/render_previews` 已可重生 780 張 sprites、PIC、BIGPIC 與動畫 manifest。
- `cmd/combat-tile-export` 已能逐張輸出三類戰場 tile。
- `cmd/symbol-export` 現在輸出放大 contact sheet，適合人工稽核，**不適合 runtime**；
  要增加原尺寸逐 item 模式與 manifest。
- `cmd/picture-probe` 可匯單張 SKY／picture，適合診斷；要補批次 export。
- `TILES`、SKY、AREA symbols、wall symbols／WallDef manifest 尚未有統一的 release
  asset generator。

## Release 現況與遷移順序

`tools/package-release.sh` 會把整個 `assets/` 放進 patch，但執行檔啟動仍要求
`-image curseoftheazurebonds.zip`，而 tile／terrain／symbol／sky／wall loader 都直接
讀該 ZIP。`patch` 因此尚不是圖像自足；`full-local` 甚至會複製原版 ZIP。

建議依風險拆成三刀：

1. 匯出並切換 TILES、三類 combat terrain、AREA/shared symbols、SKY；這些都是
   PNG＋簡單索引，先建立「刪掉原版圖像 member 仍可開畫面」的測試。
2. 匯出 wall symbol PNG＋WallDef JSON，讓 `loadMapPieceSets` 改由 manifest 重建
   `PieceSet`；以既有 1,498 張第一人稱零差異樣本抽查代表地圖。
3. 打包時禁止把原版 ZIP 當圖像依賴；另以缺少圖像 DAX member 的 fixture 啟動
   opening、AREA、第一人稱、戰鬥與 BIGPIC 五類代表畫面。

完成第 3 刀後，才能宣稱「release 圖像只來自獨立 PNG」。非圖像資料是否仍需要
原版 ZIP，必須在發行說明中另行揭露。
