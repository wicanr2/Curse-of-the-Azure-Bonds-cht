# 第二百六十九輪：ECL LOAD CHARACTER roster transaction

狀態：`READY`

## 證據

`ovr003.CMD_LoadCharacter` 以 `vm_GetCmdValue(1)` 取得 player selector，先保留 bit 7，再以 `player_index & 0x7f` 查 `TeamList`；有效索引從 1 開始。查不到時設 `player_not_found`，不會阻止後續 ECL instruction。bit 7 另控制 restore／party-summary 副作用，不能直接等同於 WHO 的 UI 選擇。

## Contract

- bounded VM 保留既有 `LoadCharacterAddresses` raw word signal，並新增 `LoadCharacterRequest{Address, Value, PlayerIndex, HighBitSet}`。
- `PlayerIndex` 是低 7 bits 的 1-based selector；State 以 `PlayerIndex-1` 對應 persistent `partyRoster`，成功後更新共用 selected-player ID。
- `PlayerIndex == 0` 或超出 roster 會保留 `LoadCharacterNotFound`，不清除上一個有效 selected player。
- `HighBitSet` 只保存 reference restore／redraw flag；完整 `FreeCurrentPlayer`、party summary redraw 與 DOS external string context 仍是下一個 engine boundary。

這一輪完成的是「ECL selector → party roster」資料橋，不宣稱已完成所有原版角色 record 載入或 UI redraw side effects。
