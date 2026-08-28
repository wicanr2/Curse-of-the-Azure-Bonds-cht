# 1243：執行期 PNG 圖像第一刀與冒險人物舞台判讀

狀態：`READY`
日期：2026-08-27

## PNG 獨立化

`cmd/runtime-image-export` 以原版 ZIP 為一次性 export oracle，產生
`assets/runtime-images/manifest.json` 與 1,741 張原生 PNG：48 張 `TILES`、65 張
戰場地形、1,625 張 `8X8D1..6` 符號及 3 張 `SKY`。每筆 manifest 保留來源檔、
block、item 與相對路徑。

`cmd/azure-bonds-game` 的 `loadTileImages`、`loadCombatTerrainImages`、
`loadAreaMapSymbols`、`loadSymbolBlock`、`loadSkyImages` 已改讀這份 manifest；它們
不再碰原版圖像 member。啟動參數 `-image-assets` 可指定資產根目錄，預設為
`assets/runtime-images`。

第二刀把 `WALLDEF2..6` 的 20 個 block 寫入同一 manifest；每筆保留來源、block 與
原始 bytes。`loadMapPieceSets` 從 JSON 重建 5×156 cell、續段 record、selector、
symbol block 與全域 ID band，再由獨立 PNG 取得符號像素。`TILES`、三類 combat、
AREA／shared symbols、SKY、WALLDEF 與 8X8D 的 runtime loader 至此均不讀原版圖像
DAX。原版 ZIP 仍供 ECL、GEO、ITEMS、MON* 等非圖像資料使用，不能把本結論擴張成
「remake 已完全不需要原版資料」。

## README 冒險畫面左上人物框

使用者指出 `docs/screenshots/gold-box-layout-adventure.png` 的左上人物似乎沒有塞滿
內框。重新對照 spec 391 的 DOS oracle 與現行 raster 後，判定這不是待修缺陷：

- 原版 HEAD／BODY 合成畫布固定為 88×88，placement 是 `(28,24)`；
- 人物可見裁切區為 90×90、起點 `(27,22)`，因此四周本來就有極窄安全邊；
- 人物姿勢本身帶透明／黑底留白，並非 loader 額外縮小；
- 現行 renderer 採嚴格 2× nearest-neighbour，沒有 letterbox 或非整數縮放；
- spec 1130 的「sub-image 目的地座標扣兩次」已修正，現行 README 圖不是該舊錯誤。

若改用 cover 填滿 90×90，必須非整數放大 88×88 或裁掉手臂／武器，會改變原作像素
與構圖。因此本輪不修改人物縮放，也不為了視覺滿框重拍成錯誤版本。這個判定屬
`material-exact/layout-reconstructed`：88×88 raster 與整數倍率是 exact，畫框位置由
正規化 DOS 擷取重建。

## 驗證

- exporter 實際輸出：`tiles=48 combat=65 symbols=1625 sky=3 walls=20`；
- `go test ./cmd/runtime-image-export ./cmd/azure-bonds-game` 在 CoAB Docker＋Xvfb 通過；
- 所有新輸出檔抽查為目前使用者 UID/GID；
- 以新 PNG loader 重拍 `-inn`；全圖只在右側 roster 的 `x=533..619, y=57..65`
  有 149 個文字像素差異，左上人物舞台逐像素相同。README 人物圖另與 DOS oracle
  人工並列檢視，不做破壞性 cover／stretch。
- 以 WallDef JSON＋symbols PNG 重拍 `-tilverton-dungeon`；第一人稱 stage
  `(40,40)..(239,239)` 為 0 像素差。第一次抽樣曾抓到 3,712 個洋紅差異像素，原因是
  PNG 保留的 EGA index 13 尚未在牆面取圖端套透明鍵；改以專案 `gfx.EGA16[13]`
  （實際 RGB 為 255,82,255，不是概略的 255,85,255）比對後歸零。全圖剩餘 2,718
  個差異只在右側角色文字與下方動態 HUD／命令文字，不屬於牆面 raster。
- `cmd/nonimage-zip-fixture` 從原版 ZIP 移除 51 個圖像 members、保留 43 個非圖像
  members。以同一 fixture 抽樣序幕 PIC、AREA、第一人稱、戰鬥與 BIGPIC，五者均
  正常產生 640×480 非空白畫面；序幕 PIC、第一人稱 stage、戰鬥 viewport 與現行
  基準各為 0 像素差。
