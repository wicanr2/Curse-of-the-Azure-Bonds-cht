# 第二百零七輪：ECL FIND ITEM signal

狀態：`READY`

## 證據

ECL5 block 48 entry 4 在 `LOAD CHARACTER` 後回到事件開頭，重複執行
`FIND ITEM` operand literals `0x5E`、`0x60`、`0x61`，每次後面都有 `IF =`／
`GOTO`。這是 party inventory query，不是單純 ECL memory arithmetic。

## Contract

- bounded VM 以 `operandValue` 解析 item ID，保存到 `RunResult.FindItemIDs`；
- 注入 party inventory context 時設定 `=`／`<>` compare flags；沒有 context 的 raw trace 保留 unresolved request；
- 從下一個 instruction 繼續，保留原始 branch structure；
- State／party adapter 後續可依 query ID 實際設定 found/not-found result。

這讓 ECL5 real entry 越過 item-query engine boundary；第二百七十一輪已完成全隊
item-type lookup 與 compare branch，其他作品的 item namespace仍須各自驗證。
