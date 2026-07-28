# 第三百四十四輪：AREA 俯視地圖與倚天字形

狀態：`READY`

## 完成範圍

- 獨立 engine 的 `areamap.Project` 將 16×16 GEO grid 投影成 cell 與去重牆段，
  並保留 wall type 與 door detail。
- CoAB game-pack JSON 宣告 `tilverton.area-map`、Area 2、`GEO2.DAX` block
  `0x01`、`8X8D2.DAX` 與 2× scale；作品資料沒有寫死進共用 engine。
- 正式 dungeon 按 `A` 開啟 AREA，`A`／`Esc` 返回；舊 `ModeMap` renderer
  也不再顯示 WILDCOM 戰鬥地板。
- 新增 ETen `STDFONT.15` Big5 分區讀取器、optional `SPCFONT.15` 與逐列水平
  膨脹 1px 的 16×15 粗體。字型檔不納入公開 repo。
- Docker／Xvfb 實機證據：
  `docs/screenshots/coab-area-map-remake.png`。

## 驗證

- `golden-box-remake-engine`: `go test ./...`
- CoAB：Docker／Xvfb `go test ./...`
- screenshot：640×480 PNG，GEO2/01、位置 `(7,13)`、方向 N。

## 尚未宣稱完成

`8X8D2` block 1 已確認含 70 個 8×8 symbols，也看得到 AREA 類牆角／門圖形；
但 symbol ID 與 GEO wall/detail 的原版組合表尚未由程式碼或 DOS oracle
證實。本輪採資料正確的 GEO 向量投影，不將它描述為原版逐像素還原。
