# 第二十一輪：NEWECL signal

狀態：`READY`（限 command signal，不含跨 block loader）

`RunSubset` 現在支援 `0x20 NEWECL` 的 operand framing，遇到時回傳 `RunResult.NewECLBlockID` 並停止目前 block；這使上層 session 能載入對應 DAX block，而不是把 block switch 當成普通 fallthrough。

本輪有 synthetic `0x51` regression；下一輪已建立 decoded block session loader。實際 Shadowdale sequence 目前仍會在 bounded path 回到野外／後續 menu，尚未取得可宣稱的 real block switch execution。

## 驗收

```sh
go test ./...
```
