# 第三百四十三輪：原版世界地圖

狀態：`READY`（限 BIGPIC overland 顯示、A–N location data 與繁中 HUD）

## 證據

- 原始 `BIGPIC1.DAX` block `0x79` 是 304×120 月海／谷地世界地圖。
- 使用者提供的 Clue Book PDF 第 35 頁列出同一張圖與所有地名。
- 攻略與說明書確認 CoAB 只能在方形／圓形興趣點間旅行，不能在世界地圖自由走格。
- `WILDCOM` 50×25 是野外戰鬥 background，不是 overland map。

## 實作

- 獨立 engine schema 新增 `overland` map、image resource 與 localized points。
- `worldmap` package 提供作品中立的 point lookup／cardinal selection。
- CoAB JSON 保存 14 個 A–N world values、穩定 ID、繁中 key 與 BIGPIC source
  pixel 座標；`BIGPIC1.DAX`、block `0x79` 與 2× scale 也屬資料。
- `ModeWilderness` 的正式畫面繪製 608×240 原圖、目前位置框、旅行選單與繁中 HUD。
- `-world-map` 提供 deterministic visual verification，不改變正式 ECL route logic。

## 驗證

[`coab-overland-map-remake.png`](../screenshots/coab-overland-map-remake.png)
是 Docker／Xvfb 直接擷取的 640×480 Ebiten 實機圖，目前位置為 world value 4
「立石」。game-pack regression 鎖定 block、14 locations 與代表地點。

尚未完成：將 route adjacency 從 ECL 自動匯出為 JSON graph、Shadowdale 的
AREA overhead map，以及旅途中所有 optional encounters。
