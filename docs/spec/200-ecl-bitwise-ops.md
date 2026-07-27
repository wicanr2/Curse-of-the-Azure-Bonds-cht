# 第二百輪：ECL bitwise AND／OR bounded semantics

狀態：`READY`

## 證據

`internal/ecl/operand.go` 的 command table 將 `0x2F` 與 `0x30` 分別標為
`AND`、`OR`，兩者都是三個 operand；同一張 table 對 `ADD`／`SUBTRACT`／
`DIVIDE`／`MULTIPLY` 使用相同的三 operand 形式。既有 runtime 的 arithmetic
實作及原始 operand cursor parser 已證實第三個 word operand 是 destination，
前兩個 operand 可由 literal 或 memory 讀值。

本輪只採用這個可由 command table 與既有 VM operand contract 支持的部分：

```text
AND left, right, destination  => memory[destination] = value(left) & value(right)
OR  left, right, destination  => memory[destination] = value(left) | value(right)
```

這是 16-bit、無號、wrap-free 的 bitwise operation；operand code、destination
address 或 runtime memory 的錯誤沿用 arithmetic 的 bounded error boundary。

## Boundary

`0x2D CALL` 仍不實作。雖然 command name 與 arity 已知，但目前沒有足夠證據
確認它是 code-segment call、external routine dispatch 或帶有隱含 context 的
engine service；把它直接當成 `GOSUB` 會污染 return stack／block context。

本規格不宣稱完整 ECL VM，也不改變 `0x2D CALL` 的 unsupported signal。
