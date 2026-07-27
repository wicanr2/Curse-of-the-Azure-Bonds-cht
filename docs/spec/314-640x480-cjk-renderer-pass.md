# 第三百一十四輪：640×480 圖像放大與中文 HUD 排版

狀態：`READY`

## 問題證據

- Ebiten logical canvas 已是 `640×480`，事件 PICTURE、人物與戰鬥小人也已使用
  nearest-neighbor 整數放大；但正式 dungeon floor viewport 仍直接繪製原始
  24×24 tile，與右側 2× wall stamps 倍率不一致。
- Combat 把所有操作與最多七個法術提示接成單一 24px 中文長行，必然可能超過
  640px。Dungeon event message 與 controls 也缺少寬度管理。
- `ModePlace` choices 在共用 wilderness/place branch 與後續 place branch 各畫一次，
  造成不同 line-height 的重影。
- 文件雖已規定正文 24px、緊湊欄位 16×15 級，但 renderer 原本只有一個 face。

## 實作契約

- 維持 640×480 logical canvas，不把整張 320×240 framebuffer 放大。
- Dungeon floor 改為 6×3 viewport；每個原始 24×24 tile 以 nearest-neighbor 2×
  顯示，佔 48×48 logical pixels，與 wall art 同為整數倍率。
- 外部 CJK font 同時建立 24px 正文 face 與 16px compact face。未指定 `-font`
  時可依 Linux／macOS／Windows 的常見系統 CJK font 路徑自動尋找，但 repo
  不提交來源或授權未確認的字型。
- Combat controls 與 spell shortcuts 分成兩行 compact text；Dungeon message
  以 Unicode rune 每 22 字、最多兩行排版，door／controls 使用 compact face。
- `ModePlace` choices 只經單一 layout 繪製一次。

## 驗收

- Unicode wrapping regression 證明繁中不會在 UTF-8 byte 中間切斷，並限制行數。
- `cmd/azure-bonds-game` 在完整 X11／ALSA build dependencies 下可編譯。
- `go test ./...` 通過；正式 640×480 畫面維持 nearest-neighbor 原始素材與
  高解析 CJK 文字兩條獨立 raster pipeline。
