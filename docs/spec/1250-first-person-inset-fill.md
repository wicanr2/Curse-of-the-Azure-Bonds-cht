# 1250 — 第一人稱場景填滿內框

- 狀態：`READY`
- 證據等級：`exact`（現行 PNG／stage-frame alpha 幾何）；`project decision`
  （使用者 2026-09-01 指定內框不得留下黑色空帶）

## 問題

`tilverton-first-person-remake.png` 的牆與背景仍依 DOS 的 88×88 區域
`(24,24)..(111,111)` 繪製，但 stage frame 的內側透明孔延伸到
`(22,22)..(119,119)`。因此場景四周留下 2～8 個 native pixels 的黑色空帶；
先前 `docs/audit/readme-scene-recapture-2026-08-31.md` 只檢查 176×176 區域有內容，
卻誤寫成「填滿透明孔」。

## 實作契約

1. 保留 DOS 88×88 牆片座標與 nearest-neighbour 像素，不縮放牆片。
2. 背景的天空、地平線與地面向外延伸到 stage frame 的安全內框
   `(22,22)..(119,119)`，作為牆片下方的完整底色；內框仍最後疊上。
   牆片合成後再用同一個 sky palette 封滿 `(22,22)..(119,39)` 的頂部帶，
   避免斜邊牆片的黑底重新蓋回左上角；不得覆蓋 row 40 起的側牆與中央牆。
3. 不把牆片中的合法黑色改成透明，也不改變 spec 1131 已證實的洋紅透明鍵。
4. 重新由 `-tilverton-dungeon` 正式 checkpoint 擷取 README 現用 PNG；不得後製裁切。
5. 回歸測試鎖定填滿後三段背景的矩形範圍，並確認原始 `BuildBackground` 結果未被修改。
6. 回歸測試另鎖定頂部封口 rectangle，避免調整牆片順序後恢復黑角。

## 共用 engine 抽取（2026-09-01）

使用者確認 Pool of Radiance 採相同機制後，演算法已移至作品中立的
`viewport.FillBackgroundToStageInset`。CoAB 只提供自己的 native
`StageInset{22,22,98,98,WallTop:40}` 並依序繪製 `Backdrop → walls → PostWall`；
Pool 以自己的 frame 幾何消費同一 API。engine 不含作品名稱、palette RGB 或場景常數。
