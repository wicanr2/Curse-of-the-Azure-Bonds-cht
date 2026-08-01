# 第四百零六輪：DOS GUI 繪製契約與 IDA 交叉驗證（READY）

## 範圍與結論

本規格只關閉冒險畫面的三項契約：HEAD／BODY 人物組合、第一人稱／一般
PIC 的灰色內框，以及 640×480 中文訊息區延伸。它不代表所有選單、角色
資訊頁、戰鬥畫面或動畫已逐像素完成。

- HEAD／BODY 是人物專用舞台；先畫 HEAD，再把 BODY 的列座標增加 5 個
  原版 8px 列後繪製。兩層使用黃色裂紋人物內框。
- 第一人稱與一般 PIC 使用另一個灰色點狀內框；可見內容是原生 88×88，位於
  DOS `(24,24)..(111,111)`，重製版以 2× nearest-neighbour 放到
  `(48,48)..(223,223)`。
- 640×480 不拉伸原版石框。原生前 184 列保持不動，在訊息區插入 40 列，
  再把原版 16 列命令帶移到 `y=224..239`。因此上半部及命令帶素材是
  `material-exact`，新增側牆帶是 `layout-reconstructed`。
- 正文、roster 與命令列使用倚天粗體 16×15；24px 不再是目前一般畫面的
  驗收基準。

## IDA Pro 證據

分析在 Docker 內使用 `/home/anr2/ida_94_official/dist` 的 IDA Pro 9.4，
先由 Borland module／symbol table 把 PC-98 `GAME.OVR` 定位到：

- module 29 `DRAWWIN` → overlay 28，`DRAWWINDOW 0172:0016`；
- module 30 `PORTRAIT` → overlay 29，`SHOWHEAD 0176:0091`、
  `SHOWBODY 0176:00B3`；
- module 31 `THREED` → overlay 30，`CLEAR3DVIEW 017C:018C`、
  `BUILDVIEW 017C:078B`。

`SHOWCHARACTERPORTRAIT` 的 caller 順序是 `SHOWHEAD` 後 `SHOWBODY`。
overlay 29 中 `SHOWHEAD` 直接轉送列座標；`SHOWBODY` 在呼叫共同的
`SHOWPORTRAIT` 前出現 exact bytes `8A 46 0A 30 E4 05 05 00`，即取出列
參數、清 AH、`ADD AX,0005h`。DOS overlay 29 的 `+0091h`／`+00B3h`
亦具有相同核心 bytes，故這不是只屬於 PC-98 的版面推測。

`CLEAR3DVIEW` 則以獨立的顏色欄位與矩形繪製呼叫建立 3D 區；它沒有經過
上述 HEAD／BODY wrapper。IDA 證明函式與資料流，內框像素及 88×88 可見
範圍仍由下列 DOS runtime oracle 量測，兩者不得互相取代。

可重現稽核入口是 `scripts/ida/pc98_gui_layout_audit.idc`。IDA 啟動時雖會
警告主機設定中的 IDAPython 3.14 路徑不存在，本次使用 IDC，反組譯輸出與
分析不依賴 IDAPython。

## 實機素材與雜湊

- `docs/reference/original-dos/tilverton-first-person-demo.png`：
  `79e28faf5d0b483fb765dd52e87efd98a10196a139bee354ddc8dd090681daa4`
- `docs/reference/original-dos/character-head-body-oracle.png`：
  `f53417f4e873a9a05076487b39c70c45270bf0c7c8c153f5269b94c0cdd29c0a`
- 使用者原圖 `docs/reference/user-provided/dos-character-info-layout-20260730.png`：
  `2afa9ba4b37a205571dff582ee69b2a86b305f3f057343bc40dd9ba0ea9320c5`
- 抽出的透明灰色舞台框 `internal/gfx/assets/dos-first-person-stage-frame.png`：
  `f0677007884cf456d672684a10f89818306970fbb01ff24cd151f60bfe0b2872`

## 實作與驗證

- `gfx.ExtendedAdventureFrame` 保留／移動原版 raster，沒有縮放或重畫石紋。
- `gfx.FirstPersonStageFrame` 是原版灰色內框的透明 overlay；3D 與 PIC 先畫
  88×88 內容，再覆蓋該框。HEAD／BODY 仍走 `scene_character` 契約。
- PIC 的 2× blit 直接使用 logical destination 的全域原點；不得先建立帶
  非零 bounds 的 Ebiten sub-image、再以 `(0,0)` 當局部原點，否則 88×88
  事件圖會在內格右／下側被錯誤裁掉。前端 transform 回歸鎖定 88×88 →
  176×176 的 2× scale 與 `(48,48)` 全域目的原點，正式截圖再驗收完整圖像。
- `-tilverton-dungeon` 由正式建立角色、序幕選項與地圖入口抵達第一人稱
  模式，不直接注入地城 state。
- `internal/gfx` 與 `cmd/azure-bonds-game` 測試已在 Docker／Xvfb 通過。
- README 採用本輪重新擷取的旅店與第一人稱畫面；舊截圖若仍出現在歷史
  spec，只能代表當時 checkpoint，不能作目前 GUI fidelity 證據。

## 尚未完成

- 使用者提供的角色資訊頁仍需由正常 `VIEW` 玩家路徑逐區驗收右側 roster、
  下方長文與命令列；本輪只鎖定其中人物舞台契約。
- 一般 PIC 還需依 selector 建立逐張裁切稽核；palette cycle 尚未驗收。
- 戰鬥 frame、法術／弓箭動態與所有服務選單各有獨立規格，不因本輪轉為完成。
