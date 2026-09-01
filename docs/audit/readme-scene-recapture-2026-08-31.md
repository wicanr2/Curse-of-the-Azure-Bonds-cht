# README 左上場景重拍收據（2026-08-31）

## 範圍與方法

依使用者指定，重新檢查 README 中帶左上 PIC／第一人稱 scene 的代表畫面。所有執行、
建置與 PNG 產生均在 `coab-go-ebiten:1.24` 的一次性 Docker／Xvfb 容器內完成；原始
theme 使用隔離的 `coab-ui-settings/1` 設定，modern theme 使用乾淨預設。輸入 repository
為唯讀，只有 `/tmp` 與明確截圖輸出目錄可寫。

不是對舊 PNG 後製位移。旅店與第一人稱均由 `cmd/azure-bonds-game` 的正式 checkpoint
重新進入；截圖功能等待三個 frame 後直接讀取 640×480 framebuffer。

原版 PIC／第一人稱顯示契約來自現行 renderer 與 stage-frame alpha mask：88×88 原始
scene 以整數 2× 畫入 `(48,48)..(223,223)`，正好是 176×176 透明顯示孔；modern 圖以
cover 裁切到相同 rectangle。內框在 scene 之後疊上，沒有減掉 clip origin，也沒有讓
scene 像素越界蓋住 roster／訊息區。

## 結果

| README 圖 | checkpoint／theme | 新 SHA-256 | 判定 |
|---|---|---|---|
| `gold-box-layout-adventure.png` | `-inn`／original | `466826fbfdd1d60b827d7e5468d2b0d0b937914c814a1e802a8c6c25054ea034` | 已替換；scene 與原版內框對齊，文字不被框線遮住 |
| `tilverton-first-person-remake.png` | `-tilverton-dungeon`／original | `eba3aea12c2c9d7eaef31692a8c51812c17327988c447dd01e0b3b301dc02b7c` | **已由 2026-09-01 的畫面證據推翻**：88×88 scene 只填了 `(24,24)..(111,111)`，沒有填滿 stage frame 延伸到 `(22,22)..(119,119)` 的透明孔；修正契約見 spec 1250 |
| `a6-modern-first-person.png` | `-tilverton-dungeon`／modern-a6 | `b94fc759519733cd185ae57e13c7d7b508963a656a6c8317fa86650771e05718` | 已替換；金色內框包住相同 scene rectangle，命令列未被遮擋 |
| `a6-modern-redraw-slice.png` | `-inn`／modern-a6 | `317b168e87f18ef97cfe65b5a4d524bdcb8b23103d76ff0154439fbcf8e56cb6` | 重拍與版控檔 byte-identical，不需替換 |
| `a6-modern-picture-animation.png` | `-inner-ritual`／modern-a6 | `83d23238dc4298c9eb776d08965d47dadd7a52c0e70889ea1f7e94f77f47d500` | 重拍與版控檔 byte-identical，不需替換 |

## 沒有拿來覆蓋的錯誤 checkpoint

manifest 舊稱開場圖可由 `-segment ECL1/0x52` 重生；現行程式以該命令（含
`-geo-set 1 -geo-block 1` 對照）得到的是 `找不到 GEO1 block 0x01`，不是原圖中的正常
敘事頁。這證明舊 generation mode 已過期，不證明開場玩家路徑壞掉。為避免把正常舊圖
換成錯誤畫面，本輪保留 `opening-prologue-remake.png`，並把「建立可重生的正常開場
checkpoint」留作獨立驗證工具缺口；在修好前不得宣稱那張已由現行命令重生。

## 驗收

- `cmd/screenshot-audit` 必須驗證 manifest 尺寸與雜湊。
- README 實際引用上述檔案，不另存一份未引用的新圖。
- 原版與 modern-a6 各至少一張人物 scene 及一張第一人稱 scene 經人工原尺寸檢視。
- Docker 工作後不得留下本輪容器或映像。

## 2026-09-01 勘誤與重拍

使用者以 README 現圖指出第一人稱 stage 左上仍未填滿。重新以 frame alpha 幾何
檢查後，確認本文件原先的「176×176 scene 填滿透明孔」是錯誤斷言：透明孔實際
延伸到 native `(22,22)..(119,119)`，而 DOS 背景只畫到
`(24,24)..(111,111)`；斜邊牆片又會在最後把左上角蓋成黑色。

依 [spec 1250](../spec/1250-first-person-inset-fill.md) 修正後，以相同
`-tilverton-dungeon` 正常 checkpoint、original theme、640×480 framebuffer 重拍：

- 新 SHA-256：`289fca9187b64bb734d067e9d646edb295fd846ecb05e6e548e34f1ef2e84faf`
- 背景填滿 stage inset；頂部使用同一場景的 sky palette 封口至 native row 39。
- 中央牆、側牆與地板仍使用原本牆片座標，沒有縮放或後製裁切。
