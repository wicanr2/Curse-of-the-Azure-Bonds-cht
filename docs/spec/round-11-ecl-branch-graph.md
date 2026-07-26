# 第十一輪：ECL branch target graph

狀態：`DRAFT`

`internal/ecl.TraceGraph` 將 code-segment word operand（`0x8000` 起算）轉成 decoded payload offset，並追蹤靜態可見的 `GOTO/GOSUB` target。

- `GOTO` 只加入 target，不假設 fallthrough。
- `GOSUB` 同時保留 sequential path 與 target。
- `RETURN/EXIT` 終止目前 path。
- data pointer、unknown opcode、越界與 malformed operand 不會被當作 code 執行。

這仍然不是 VM：`IF` 條件、`ON GOTO`、command side effects 與 call stack 尚未執行。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -graph
```
