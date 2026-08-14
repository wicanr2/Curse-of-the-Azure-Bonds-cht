# dos-overlay-16 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 32 | 10 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/751-overlay-init-chain-dependency-graph.md<br>unit 初始化段(retf，不收參數)：依序呼叫 5 個相依 unit 的 0000h — overlay-33、overlay-34、overlay-25、overlay-23、overlay-24。由原始 bytes 解出(IDA 匯出在此漏 byte) | audit/embedded-strings.md<br>audit/function-index/dos-overlay-00.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-23.md |
| `0020` | sub_20 | — | 87 | 36 | 2 | 1 | ✓ | 已解讀 | exact | docs/spec/758-morale-field-0f7h-round-trip.md<br>C 字串轉 Pascal 字串(retf 4)：掃到 0 得長度 n，Move n bytes 進緩衝，緩衝前一個 byte 填 n，再 StoreString 到目的(上限 0FFh)。⚠ 目的指標在 [bp+0Ah]，超出 retf 4 宣告的參數區 — 堆疊殘留當引數，已由原始 bytes(C4 7E 0A + CA 04)確認 | audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-24.md<br>audit/function-index/pc98-overlay-02.md<br>audit/function-index/pc98-overlay-15.md<br>audit/function-index/pc98-overlay-16.md<br>spec/755-per-character-slots-sound-driver-and-text-io.md |
| `008D` | sub_8D | — | 966 | 385 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `0453` | sub_453 | — | 375 | 169 | 0 | 3 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `0643` | sub_643 | — | 936 | 381 | 4 | 3 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `09EB` | sub_9EB | — | 135 | 50 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/764-fsplit-dbcs-and-eight-slot-longint-table.md<br>兩筆記錄的合併(retf 8)：n := 來源^[11h] − 1；for i := 0 to n 做 目的^[17h+i] |= 來源^[17h+i]、目的^[13h]^[i] &= 來源^[13h]^[i]。同一迴圈一個取聯集一個取交集。⚠ 來源^[11h] 為 0 時 n = 0FFFFh，迴圈會跑 65536 次 | spec/764-fsplit-dbcs-and-eight-slot-longint-table.md |
| `0A7E` | sub_A7E | — | 498 | 183 | 1 | 4 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `0C7E` | sub_C7E | — | 157 | 63 | 2 | 0 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `0DEA` | sub_DEA | — | 1436 | 546 | 1 | 5 | ✓ | 待解讀 | — | — | audit/function-strings.md<br>audit/function-triage.md |
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
| `228E` | sub_228E | — | 3994 | 1460 | 1 | 7 | ✓ | 待解讀 | — | — | audit/function-strings.md<br>audit/function-triage.md |
| `3254` | sub_3254 | — | 707 | 275 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-16.md<br>audit/function-strings.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `351C` | sub_351C | — | 108 | 29 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>載入 CPIC(retf 2)：DS:4F9Dh^[67Ch] > 7(無號)就整支不做；否則 GetMem(p, 1A6h)、本模組 3254h(參數, p)、p^[126h] := 參數、本模組 35A8h(p)，組出 CS:3517h 的 'CPIC' 字串後叫 overlay-33 entry#5 @01D4h(參數, p^[143h])。⚠ 匯出漏了 21h bytes，本判讀由原始 bytes 補回。補洞重匯出(spec 761)已確認本判讀完整 | audit/function-index/pc98-overlay-16.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>spec/760-item-effect-flag-clear-and-file-exists.md<br>spec/761-filled-export-and-dungeon-wall-bits.md<br>spec/768-script-arithmetic-opcodes-and-save-size.md |
| `35A8` | sub_35A8 | — | 116 | 37 | 2 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-16.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `3748` | sub_3748 | — | 1520 | 575 | 0 | 10 | ✓ | 待解讀 | — | — | audit/function-strings.md<br>audit/function-triage.md |
| `3D38` | sub_3D38 | — | 193 | 64 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/768-script-arithmetic-opcodes-and-save-size.md<br>算整隊要多少 bytes(retf 2，參數沒被讀)：每個角色 1A6h(422)、每件物品 3Fh(63，鏈在 +14Dh/next +2Ah)、+0F2h 那條鏈每個節點 9(next 在 +5)，全部累加成 word 回傳。三個常數與既有的 GetMem 大小、物品節點大小一致。⚠ 沒有溢位檢查，+0F2h 那條鏈沒有已知上限 | spec/768-script-arithmetic-opcodes-and-save-size.md |
| `3EDD` | sub_3EDD | — | 185 | 75 | 0 | 4 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-16.md<br>audit/function-strings.md |
| `3F96` | sub_3F96 | — | 1135 | 457 | 2 | 6 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 3EDDh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `4405` | sub_4405 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
