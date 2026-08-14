# dos-overlay-16 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 32 | 10 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/751-overlay-init-chain-dependency-graph.md<br>unit 初始化段(retf，不收參數)：依序呼叫 5 個相依 unit 的 0000h — overlay-33、overlay-34、overlay-25、overlay-23、overlay-24。由原始 bytes 解出(IDA 匯出在此漏 byte) | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-23.md |
| `0020` | sub_20 | — | 87 | 36 | 2 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-24.md<br>audit/function-index/pc98-overlay-02.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md |
| `008D` | sub_8D | — | 966 | 385 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0453` | sub_453 | — | 375 | 169 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0643` | sub_643 | — | 936 | 381 | 4 | 3 | ✓ | 待解讀 | — | — | — |
| `09EB` | sub_9EB | — | 135 | 50 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0A7E` | sub_A7E | — | 498 | 183 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0C7E` | sub_C7E | — | 157 | 63 | 2 | 0 | ✓ | 待解讀 | — | — | — |
| `0DEA` | sub_DEA | — | 1436 | 546 | 1 | 5 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `1388` | sub_1388 | — | 264 | 109 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `1490` | sub_1490 | — | 143 | 49 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-16.md |
| `1522` | sub_1522 | — | 348 | 128 | 4 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1490h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `169F` | sub_169F | — | 5 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cbw`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `16A4` | sub_16A4 | — | 10 | 4 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov dl, [di+3F8Eh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `16AE` | sub_16AE | — | 295 | 107 | 3 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1490h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `17DD` | sub_17DD | — | 292 | 106 | 2 | 3 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1490h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1994` | sub_1994 | — | 7 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov di, [bp+6]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `19EA` | sub_19EA | — | 1335 | 390 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1490h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | audit/function-triage.md |
| `1F21` | sub_1F21 | — | 777 | 234 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `228E` | sub_228E | — | 3994 | 1460 | 1 | 7 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `3254` | sub_3254 | — | 707 | 275 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `351C` | sub_351C | — | 108 | 29 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `35A8` | sub_35A8 | — | 116 | 37 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `3748` | sub_3748 | — | 1520 | 575 | 0 | 10 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `3D38` | sub_3D38 | — | 193 | 64 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `3EDD` | sub_3EDD | — | 185 | 75 | 0 | 4 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-16.md |
| `3F96` | sub_3F96 | — | 1135 | 457 | 2 | 6 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 3EDDh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `4405` | sub_4405 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
