# 第一百六十輪：LOAD PIECES 地圖素材 adapter

狀態：`READY`（限 WALLDEF／8X8D selector 載入，不宣稱完整牆面 renderer）

## 反組譯證據

公開 CoAB 重製程式的 `CMD_LoadFiles` 將 `LOAD PIECES` 的三個 VM operand 依序視為 symbol set 1、2、3 的 WALLDEF block selector；`LoadWalldef` 讀取 `WALLDEF{game_area}.DAX`。當一個 DAX block 含多個 5×156 record 時，後續 record 的 8×8D block ID 使用 `selector * 10 + recordIndex + 1`；單一 record 則使用 selector 本身。

- [參考 CMD_LoadFiles](https://github.com/simeonpilgrim/coab/blob/master/engine/ovr003.cs)
- [參考 LoadWalldef](https://github.com/simeonpilgrim/coab/blob/master/engine/ovr031.cs)

## 實作結果

- `gfx.ParsePieceSet` 將一個 WALLDEF selector 解析為 `WallDefs`，並依上述規則找出對應 8×8D block。
- `cmd/azure-bonds-game` 消費 `State.ConsumeLoadPiecesRequest()`，從 `WALLDEF{area}.DAX`／`8X8D{area}.DAX` 載入三個 selector，保存 `PieceSet` 並在 dungeon preview 顯示載入狀態。
- synthetic regression 覆蓋單 record 與多 record 的 symbol block ID；原始 `curseoftheazurebonds.zip` regression 覆蓋 area 2 selectors `[1,2,3]`。

## 邊界

本輪不把 WALLDEF row／column ID 猜成最終牆面圖層座標，也不修改 GEO floor composition、碰撞、camera 或 8×8D 的畫面拼接；這些仍需下一個 renderer／反組譯規格。`LOAD PIECES` 的 `0x7F` 特殊分支亦保留未接入。
