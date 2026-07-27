# 第二百七十一輪：ECL FIND ITEM party query

狀態：`READY`

## 證據

公開 reference `ovr003.CMD_FindItem` 取得一個 `ItemType` 後，清空六個 compare flags，
先令 `compare_flags[1]`（`<>`）為真，再巡訪所有 `TeamList` 角色的全部 item records。
任一 record type 相符便設定 `compare_flags[0]`（`=`）並清除 `<>`，立即返回。查詢不依
selected player、readied flag 或 stack count。

## Contract

- `PartyMemberContext.ItemTypes` 保存作品 roster 對 VM 的 raw item-type projection。
- 有 party context 時，`FIND ITEM` 產生 resolved `FindItemRequest`，並設定 `=`／`<>`
  compare flags；無 context 的 bounded trace 仍保留 unresolved request，不猜結果。
- 查詢範圍是全隊所有 item records；只判斷 type 是否存在，不修改 inventory。
- 同一次 VM run 的 `DESTROY ITEMS` 會從 working inventory view 移除該 type，使後續
  `FIND ITEM` 看見已毀結果；真正 roster mutation仍由 State adapter執行。

regression 覆蓋 found／not-found 的 `IF = → GOTO` 分支，以及
`FIND → DESTROY → FIND` 從 found 轉為 not-found 的順序。
