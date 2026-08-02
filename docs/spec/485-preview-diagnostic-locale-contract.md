# 第 485 輪：AREA／GEO／素材預覽診斷 locale 契約

狀態：`READY`

## 問題

素材載入、AREA、GEO geometry 與地城研究預覽器曾直接保存繁中診斷、座標、
selector 與門選項。世界地圖日期還以全形空白切割完整時間字串。這些工具雖非
原版遊戲畫面，仍是玩家／研究者可見 UI，也會造成 renderer 與 locale 耦合。

## READY 契約

1. `PreviewLabel` 表示十三種 static 診斷 chrome；動態 selector、GEO set／block、
   座標、方向、wall／roof、door flags 與錯誤由 State format contract 插入。
2. `LOAD PIECES` selectors 與 State `LoadPieces` 的正式型別是 `uint16`。不得因目前
   樣本值小於 256 就縮窄為 `uint8`；focused compile gate 已阻擋此錯誤。
3. 地城預覽門選項使用 Pick／Knock booleans 選四種 stable ID，不在 renderer
   拼接局部中英文片段。
4. 世界地圖日期直接由 typed `GameTimeDisplay` 的 day／month／year 格式化，
   不再拆解已翻譯的 `GameTimeText`。
5. preview 診斷屬 CoAB adapter，不抽入 Golden Box engine；GEO／AREA codec 與
   door options 的既有 typed 機制不變。
6. 本輪不變更 preview geometry、AREA 11×11 projection、GEO collision、
   WALLDEF rendering、地城規則或正式玩家畫面。

## 驗證

- 正式 catalog test 覆蓋十三種 static label、十一種 dynamic format、四種門
  option 組合與 typed 日期。
- `uint16` selector 由真實 app call site 編譯驗證；不接受 cast 掩蓋。
- AREA／GEO／dungeon preview 與 Ebiten tests 由 Docker／Xvfb 正式 gate 重跑。
- Go 漢字稽核 `262→238`；frontend `50→26`、runtime 212、localization 0
  不變。

本輪沒有正式遊戲版面變更，不新增 README 截圖，也不把研究預覽器誤稱原版
像素級畫面。
