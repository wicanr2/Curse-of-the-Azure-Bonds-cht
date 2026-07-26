# 第十二輪：ECL 初始化入口

狀態：`READY`（限 bounded command-set／operand parsing；不包含 VM 執行）

公開的 CoAB 重寫程式顯示，ECL block 載入後先跳過兩個 block prefix bytes，接著由 `vm_init_ecl` 連續呼叫五次單一 command-set loader。每筆 command-set 的資料形狀是：

```text
[原 VM command byte][operand code][word low][word high]
```

當 operand code 為 `1`、`2` 或 `3` 時，low/high 組成 code-segment word。五個 word 的語意依序是：run、search、pre-camp、camp-interrupted、initial entry point。

`internal/ecl.EntryPoints` 目前只解析這個已觀察到的 word-valued header，回傳原始 `0x8000` code address 與讀取後 cursor；不執行 VM，也不把結果直接視為已驗證的劇情流程。若 header 不是 word operand 或資料截斷，會安全回傳錯誤。

同一套 cursor parser 也已處理 `0x80 length payload` 的 compressed-string operand，並將 payload 暫存於 `Operand.Packed`；`TraceAt` 可從指定 decoded payload offset 開始，避免把初始化 header 當成 executable entry。未知 opcode、條件判斷與 command side effects 仍會停止或不執行。

## 驗收

```sh
go test ./internal/ecl
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -entrypoints
```

實際 `ECL1.DAX` 的三個 block 已能解析出第五個 initial entry point `0x8014`。CLI 的 `-graph` 現在優先從第五個入口開始；若 header 不完整才保留 offset 0 fallback。CLI 的 `-trace -trace-start 20` 已在 block 81 解出 `AS YOU DEPART...` 等原始事件文字。下一步是建立完整 regression，並確認 code address 與 packed text／選單事件的對齊。
