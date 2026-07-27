# 第六十七輪：ECL LOAD FILES → GEO map request

狀態：`READY`（限第一 operand 的 GEO block selector 與 renderer catalog bridge）

## 已確認

reference `ovr003.CMD_LoadFiles` 讀取三個 command-set operand；operand 1 會存入 local
`var_3`。當它不是 `0xFF`／`0x7F` 且處於 dungeon 時，會寫入
`Area1.current_3DMap_block_id` 並呼叫 `Load3DMap(var_3)`。

本輪實作：

- `ecl.RunSubset` 的 opcode `0x21` 解碼三個 operand，輸出 `LoadFilesRequested` 與 `[3]uint16`。
- `game.State` 在 `Area.InDungeon` 為 true 時，將有效第一 operand 保存為
  `GeoMapBlock` pending request。第 280 輪依 reference `var_3=operand 1` 修正早期誤寫的
  「第三值」；第三 operand 實際屬 outdoor BIGPIC branch。
- Ebiten app 消費 request，從 `geo.Catalog` 查找原始 set/block，並同時更新 G geometry 與 D dungeon floor。
- state／ECL regression 驗證 request 只能消費一次；catalog regression 驗證所有 16 個原始 GEO IDs。

## 尚未完成

`inDungeon`／`game_area`／完整 Area1 save loader、WALLDEF set reload、非 dungeon big-picture side effect、ECL 全 opcode 與完整玩家流程仍待完成。
