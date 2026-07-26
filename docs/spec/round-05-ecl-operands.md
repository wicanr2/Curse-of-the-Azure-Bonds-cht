# 第五輪：ECL operand framing

狀態：`DRAFT`

## 已確認的局部契約

ECL 的 operand 讀取遵循下列 cursor 規則：

```text
operand starts at cursor
skip byte at cursor
code = byte[cursor + 1]
low  = byte[cursor + 2]
cursor += 2

if code in {0x01, 0x02, 0x03}:
    cursor += 1
    high = byte[cursor]
    word = (high << 8) | low

after N operands: cursor += 1
```

`internal/ecl.ParseOperands` 已將這個局部契約實作並測試。這是 operand framing，不是完整 opcode interpreter。

## 證據與限制

- `ECL1.DAX` decoded block 的開頭可依此讀出第一個 word operand `0x806A`。
- code `0x00` 是 scalar low byte；code `0x01`–`0x03` 會有 word operand，但其記憶體語意仍依 opcode 而定。
- 後續 command 的 operand 數量與 control-flow semantics 尚未完成，因此規格維持 `DRAFT`。

## 下一步

建立只讀 ECL trace：以已知 command table 做 cursor 前進、記錄未知 opcode 與 block 邊界；遇到未知語意時停止，不執行猜測行為。
