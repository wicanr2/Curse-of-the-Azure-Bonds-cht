# 第二百六十九輪：ECL LOAD CHARACTER roster transaction

狀態：`READY`

## 證據

`ovr003.CMD_LoadCharacter` 以 `vm_GetCmdValue(1)` 取得 player selector，先保留 bit 7，再以 `player_index & 0x7f` 查 `TeamList`。第三百一十輪以 ECL5 阿卡巴搜尋子程序補正：它明確從 0 走到 7，因此有效索引是 zero-based。查不到時設 `player_not_found`，不會阻止後續 ECL instruction。bit 7 另控制 restore／party-summary 副作用，不能直接等同於 WHO 的 UI 選擇。

## Contract

- bounded VM 保留既有 `LoadCharacterAddresses` raw word signal，並新增 `LoadCharacterRequest{Address, Value, PlayerIndex, HighBitSet}`。
- `PlayerIndex` 是低 7 bits 的 zero-based selector；State 直接對應 persistent `partyRoster`，成功後更新共用 selected-player ID。
- index 超出 roster 會保留 `LoadCharacterNotFound`，不清除上一個有效 selected player；
  bit 7 只保留 restore／redraw 語意，不改變低七位的角色索引。
- `HighBitSet` 只保存 reference restore／redraw flag；完整 `FreeCurrentPlayer`、party summary redraw 與 `0x7C00` 姓名以外的 DOS external string context 仍是下一個 engine boundary。
- CoAB `PartyContext` 會將 script name 提供給 runtime 的 `0x7C00` selected-player string slot，因此後續 `0x81` string operand可被 `COMPARE`／`IF` 使用；其他 DOS memory regions 仍不會被猜測填入。

這一輪完成的是「ECL selector → party roster」資料橋，不宣稱已完成所有原版角色 record 載入或 UI redraw side effects。
