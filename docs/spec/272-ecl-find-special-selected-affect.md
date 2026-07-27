# 第二百七十二輪：ECL FIND SPECIAL selected affect

狀態：`READY`

## 證據

公開 reference command table 將 `0x3F` 定義為單 operand `FIND SPECIAL`；`0x3D` 是
零 operand `CLEAR BOX`。`CMD_FindSpecial` 清空六個 compare flags，解析 affect ID，
再呼叫 `gbl.SelectedPlayer.HasAffect`：找到時 `=` 為真，否則 `<>` 為真。

## Contract

- `RuntimeState` 保存 selected-player index 與有效旗標，與 memory、strings、compare flags
  一起跨 menu／WHO pause 與 shared BlockSession／NEWECL context。
- 成功的 `LOAD CHARACTER` 以 zero-based selector 更新 selected index；完成的 `WHO`
  selection 以 UI 的 0-based roster index 更新同一狀態。
- `FIND SPECIAL` 只查 selected member 的 active `PartyMemberContext.Effects`，產生
  `FindSpecialRequest` 並設定 `=`／`<>`；沒有 party context 或尚未選角時維持 unresolved。
- inactive FX records 不進 `PartyMemberContext.Effects`，因此不會被誤判為 active special。

regression 覆蓋 `LOAD CHARACTER → FIND SPECIAL → IF/GOTO`，以及 WHO pause 後選第二位
角色 resume，再查該角色 affect 的 transaction。
