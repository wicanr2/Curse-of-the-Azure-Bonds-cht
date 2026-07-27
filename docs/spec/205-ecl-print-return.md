# 第二百零五輪：ECL PRINT RETURN boundary

狀態：`READY`

## 證據

ECL3 block 16 entry 4 在 `PRINTCLEAR` 與三段 Yulash text 後遇到 opcode `0x33`
`PRINT RETURN`；原始 cursor trace 顯示該 command 沒有 operand，後面仍有事件
文字／menu continuation。它不是 `EXIT`，也不是 `RETURN` 的 ECL call-stack
操作。

## Contract

- bounded VM 消費 `0x33` 並增加 `RunResult.PrintReturnCount`；
- 從下一個 instruction 繼續，保留 raw text 與 menu control flow；
- renderer-facing text window／cursor layout 留給 game UI adapter；
- 不修改 ECL call stack、runtime memory 或 party state。

這讓真實 ECL3 text event 可以越過 `PRINT RETURN` 到下一個可觀測邊界，並不宣稱
已完成 DOS text window 的像素級排版。
