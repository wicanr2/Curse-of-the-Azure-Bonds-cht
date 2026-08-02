# 第 483 輪：冒險 chrome 與世界地圖 locale 契約

狀態：`READY`

## 問題

Ebiten 冒險 renderer 曾直接保存選擇／存讀檔提示、事件繼續、暗影谷 AREA、
世界地圖、PICTURE 缺素材診斷、角色欄標題、戰鬥角色檢視及倒地標記等繁中
文字。世界地圖地名查詢還把 locale 固定成 `zh-TW`，使 State 即使使用其他
catalog，地圖仍可能混入繁中。

## READY 契約

1. `PlayerUILabel` 是 renderer-neutral typed identity；State 將十五種玩家介面
   語意映射到 stable locale ID。Ebiten 只決定畫面狀態、字型、座標與顏色。
2. AREA 座標與世界地圖目前位置由 State format contract 插入 runtime 值；
   座標、地名不能複製進 locale 或 frontend。
3. game-pack 世界目的地 `message_id` 必須使用 State catalog language 查詢，
   不得在 renderer 固定 `zh-TW`。
4. 一般事件、HEAD／BODY、BIGPIC 與 PIC 共用同一 continue identity；只有原作
   BIGPIC 版面需要的合併提示使用獨立 ID。
5. 倒地 overlay 三條 renderer branch 共用同一 stable ID，不因 sprite 是否
   可用而產生文字副本。
6. 本輪不變更 640×480 geometry、DOS 石框、HEAD／BODY anchor、PICTURE cover、
   世界地圖座標、戰鬥 timeline、sprite 或輸入行為。

## 驗證

- 正式 catalog table test 覆蓋十五種 static label、AREA 動態座標、世界地圖
  動態地名與 State locale language。
- 既有 adventure、overland、picture、HEAD／BODY、combat view、death overlay
  及 Ebiten tests 由 Docker／Xvfb 正式 gate 重跑。
- Go 漢字稽核 `300→275`；frontend `88→63`、runtime 212、localization 0
  不變。

本輪只改文字來源，不新增 README 截圖，也不擴大宣稱 GUI 已完成像素級終驗。
