# dos-overlay-11 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 47 | 13 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化(LOADINIT):依序 far call 八個依賴單元的初始化 entry,最後呼叫 19Ch:2Ah 的 overlay 載入器 | audit/embedded-strings.md<br>audit/overlay-init-graph.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md<br>spec/521-dos-getmem-buffer-owner.md |
| `0056` | sub_56 | — | 1666 | 403 | 0 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ds:7214h, ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1666 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-strings.md |
| `06D8` | sub_6D8 | — | 516 | 164 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `08DC` | sub_8DC | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
