# dos-overlay-29 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call far ptr 19Ch:2Ah`（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-17.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/754-small-predicates-and-wrappers.md |
| `000C` | sub_C | — | 133 | 51 | 3 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-29.md<br>spec/753-small-utility-routines.md<br>spec/754-small-predicates-and-wrappers.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `0091` | sub_91 | — | 34 | 17 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>單一呼叫包裝：整個 body 只準備參數並執行 `call near ptr sub_C`（body 共 34 bytes，已逐條讀完） | audit/function-index/dos-overlay-29.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `00B3` | sub_B3 | — | 37 | 18 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>包裝(retf 8)：f(a: byte, b: byte, p: 遠指標) → 本模組 sub_Ch(a, b+5, 0, p)。兩個 byte 參數零延伸成 word | audit/function-index/dos-overlay-29.md<br>audit/function-index/pc98-overlay-29.md<br>spec/754-small-predicates-and-wrappers.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `010B` | sub_10B | — | 1104 | 453 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `055B` | sub_55B | — | 105 | 41 | 2 | 1 | ✓ | 待解讀 | — | — | spec/749-combat-teardown-and-battlefield-grid.md<br>spec/750-combat-setup.md |
| `05DD` | sub_5DD | — | 260 | 110 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `06E1` | sub_6E1 | — | 84 | 36 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/758-morale-field-0f7h-round-trip.md<br>兩種畫法的分岔(retf 6)：第一個參數非 0 時叫本模組 0091h(y, x, 遠指標@DS:726Fh) 與 00B3h(y, x, 遠指標@DS:7274h)；為 0 時改叫 000Ch(y, x, 0, 遠指標@DS:726Fh)。確認 spec 754 那兩組「哨兵 byte + 4 bytes」的 4 bytes 是遠指標 | audit/function-index/pc98-overlay-29.md<br>spec/758-morale-field-0f7h-round-trip.md |
| `0754` | sub_754 | — | 184 | 74 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0813` | sub_813 | — | 110 | 51 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0881` | sub_881 | — | 33 | 13 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>retf，無參數：01A0h:0252h() 後 0297h:1110h(1, 1, 遠指標@DS:728Ch)。⚠ PC-98 對應那支多一個 02A8h:10D5h(@DS:0A327h)，已對原始 bytes 確認不是匯出誤差 | spec/753-small-utility-routines.md |
| `08A2` | sub_8A2 | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
