# 第二百零六輪：ECL LOAD CHARACTER signal

狀態：`READY`

## 證據

ECL5 block 48 entry 4 的 raw trace 在 `NEWECL 0x50` 後執行：

```text
SAVE 0 -> [0x7F79]
LOAD CHARACTER [0x7F79]
COMPARE [0x7CB8] with text "AKABAR BEL AKAS"
```

`0x0A` 是一個 word-address、single-operand command，後面仍有 compare／branch，
不是 ECL `RETURN` 或終止指令。

## Contract

- bounded VM 將 address 保存為 `RunResult.LoadCharacterAddresses`；
- 視為 synchronous external character/string load，從下一個 instruction 繼續；
- 不虛構 DOS player record、姓名或 string-memory side effect；
- 後續 State adapter 可依 address 與 party roster 提供實際角色資料。

這個切片的目標是越過 ECL5 真實事件的下一個 engine boundary，不宣稱完整
character-record import 或所有 compare branch 已完成。
