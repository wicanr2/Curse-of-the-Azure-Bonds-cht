# 第二百零八輪：ECL DESTROY ITEMS signal

狀態：`READY`

## 證據

ECL5 block 48 raw scan 在 `FIND ITEM 0x5E/0x60/0x61` 後緊接三個
`DESTROY ITEMS`，operand 同樣是 item ID literal。這是 party inventory mutation，
不是 ECL memory write；原始事件在這段後仍有 control flow。

## Contract

- bounded VM 以 `operandValue` 解析 item ID，保存到 `RunResult.DestroyItemIDs`；
- 不在 VM 直接刪除 party equipment／inventory；
- 從下一個 instruction 繼續，讓後續 ECL event 可觀察；
- 後續 State／party adapter 依 item ownership／count 實作實際消耗。

這是 inventory mutation signal，不代表完整原版 item transfer、cursed item 規則
或存檔 writeback 已完成。
