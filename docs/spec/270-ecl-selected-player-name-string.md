# 第二百七十輪：ECL selected-player name string

狀態：`READY`

## 證據

公開 reference `ovr008.vm_CopyStringFromMemory` 先依 address range 選擇 memory type；
在 player-memory range `0x7C00..0x7FFF` 中，`location == 0x7C00` 是明確特例，直接
複製 `gbl.SelectedPlayer.name`。`CMD_LoadCharacter` 會先依 operand value 更新
`SelectedPlayer`，因此後續 code `0x81` string-memory operand 可以用 `0x7C00` 比較姓名。

## Contract

- `PartyMemberContext.Name` 是作品 roster 對共用 VM 的 renderer-neutral 姓名投影。
- 成功的 `LOAD CHARACTER` 將該姓名寫入 resumable `RuntimeState.Strings[0x7C00]`。
- 後續 `COMPARE`、`IF`、`GOTO` 使用既有 string／control-flow semantics，不新增作品專屬 opcode。
- 無效 selector 不改寫 `0x7C00`；`0x4B00`、`0x7A00`、其他 `0x7C00` player fields 與
  `0x8000` ECL-memory strings仍需各自證據，不以姓名特例外推。

synthetic regression 以 roster name `HI` 執行 `LOAD CHARACTER 1 → LOAD CHARACTER 0
(not found) → COMPARE [0x7C00], "HI" → IF = → GOTO success`，證明失敗 lookup 保留
上一個 selected player，且只輸出 `YES`。
