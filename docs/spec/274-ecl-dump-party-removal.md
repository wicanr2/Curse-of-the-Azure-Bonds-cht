# 第二百七十四輪：ECL DUMP party removal

狀態：`READY`

## 證據

公開 reference `CMD_Dump` 不是 diagnostic print：它以目前 `SelectedPlayer` 呼叫
`FreeCurrentPlayer(player, true, false)`，從 `TeamList` 移除角色、釋放 combat icon、
減少 party size，並將回傳角色寫入 `SelectedPlayer`／`LastSelectedPlayer`。被移除 index
大於零時回傳前一位；移除第一位時回傳新的第一位；隊伍為空時回傳 null。

原始 ECL5 block `0x30:+0x020E` 在 Akabar 離隊 routine 中明確解碼為 `DUMP (0x3E)`。

## Contract

- bounded VM 產生 ordered `DumpRequest`，並在 working `PartyContext` 移除 selected member。
- public runner／BlockSession 先 deep-copy caller context；同一 mutable copy 跨 NEWECL 使用，避免離隊角色在 target block 重新出現，也不修改呼叫端 roster projection。
- working inventory、selected-player index 與 `0x7C00` 姓名同步更新，後續
  FIND ITEM／FIND SPECIAL／party rules 看到移除後隊伍。
- State 依 request index 從 persistent `partyRoster` 與 fighter projection 移除同 ID；
  selected player 改成 reference predecessor，隊伍空時清除 selection。
- ECL DUMP 不套用玩家 `ALTER DROP` 的「至少保留一人」限制；它是 script-controlled
  NPC／party departure side effect。DOS sidecar 實體刪除仍由既有 save transaction處理。

regression 覆蓋中間角色移除後的 FIND SPECIAL、新 selected player、roster／fighter同步、
最後一人移除、DUMP 跨 NEWECL working-party persistence／caller isolation，以及 real ECL5
Akabar DUMP opcode位置。
