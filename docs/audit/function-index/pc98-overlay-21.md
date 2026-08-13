# pc98-overlay-21 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADMONEY | 23 | 6 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `001B` | sub_1B | MAXIUMWEIGHT | 32 | 12 | 3 | 1 | ✓ | 待解讀 | — | — | — |
| `003B` | sub_3B | LOSEWEIGHT | 28 | 10 | 3 | 0 | ✓ | 待解讀 | — | — | — |
| `0057` | sub_57 | GAINWEIGHT | 28 | 10 | 4 | 0 | ✓ | 待解讀 | — | — | — |
| `0073` | sub_73 | TOOHEAVY | 84 | 31 | 3 | 2 | ✓ | 待解讀 | — | — | — |
| `00C7` | sub_C7 | CASHPOOL | 129 | 51 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0148` | sub_148 | CHANGECHARMONEY | 85 | 31 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `019D` | sub_19D | CHANGEPOOLMONEY | 86 | 34 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `0218` | sub_218 | ADDPLATINUM | 145 | 53 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `02AB` | sub_2AB | GETMONEYINPUT | 461 | 203 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0485` | sub_485 | GIVEMONEY | 188 | 71 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0541` | sub_541 | POOLMONEY | 210 | 79 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0613` | sub_613 | NUMOFPCS | 86 | 30 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0669` | sub_669 | SHAREPOOL | 948 | 382 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0A1D` | sub_A1D | DROPCASH | 124 | 50 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0AA6` | sub_AA6 | — | 206 | 87 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0B9B` | sub_B9B | GETMONEYTYPE | 520 | 219 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `0DE8` | sub_DE8 | TAKEMONEY | 12 | 4 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1070` | sub_1070 | CHECKTREASURE | 84 | 30 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `10C4` | sub_10C4 | — | 56 | 25 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `10FC` | sub_10FC | CREATERNDTREASURE | 1455 | 538 | 0 | 4 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `147D` | sub_147D | — | 143 | 51 | 2 | 2 |  | 待解讀 | — | — | — |
| `151F` | sub_151F | — | 15 | 5 | 5 | 2 |  | 待解讀 | — | — | — |
| `152E` | sub_152E | — | 115 | 38 | 1 | 2 |  | 待解讀 | — | — | — |
| `15A1` | sub_15A1 | — | 382 | 141 | 2 | 1 |  | 待解讀 | — | — | — |
| `1683` | sub_1683 | — | 10 | 3 | 2 | 1 |  | 待解讀 | — | — | — |
| `19CA` | sub_19CA | — | 12 | 3 | 1 | 0 |  | 待解讀 | — | — | audit/function-triage.md |
| `19E0` | sub_19E0 | — | 31 | 14 | 2 | 0 |  | 待解讀 | — | — | — |
| `1A3E` | sub_1A3E | APPRAISE | 2263 | 151 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2315` | sub_2315 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
