# 第二百零九輪：ECL inventory state adapter

狀態：`READY`

## 證據

ECL5 block 48 text 明確描述暗精靈在日光下武器／護甲腐朽，raw command sequence
以 `FIND ITEM 0x5E/0x60/0x61` 查詢後執行同 ID 的 `DESTROY ITEMS`。這些 ID
必須作用到 persistent party roster，而不是只存在 VM signal。

## Contract

- State 對 `RunResult.DestroyItemIDs` 逐一套用到所有 roster character；
- `party.Character.DestroyItemType` 移除該 type 的全部 units，包含 readied item，
  以區分 ECL effect 與玩家手動 `RemoveItem` 的安全限制；
- `FIND ITEM` query result 仍未虛構，沒有 matching equipment 時不產生 mutation；
- raw ECL signal 與 save／party roster 是分層的，VM 不直接依賴 party package。

這是已驗證事件的 inventory side effect，不代表全部 ECL item rules、cursed item
effects 或所有作品的 item ID namespace 都相同。
