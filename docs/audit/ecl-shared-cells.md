# ECL↔引擎共用格子清冊（dos）

由 `scripts/ecl_shared_cells.py` 產生，不要手改。
換算依 spec 1096 §二：bank 內位移 ＝ `(位址 − 區基底) × 2`。

- ECL 側變數位址：**81 個**（674 次存取）
- **共用格子（引擎側也存取同一位移）：24 個**（184 次 ECL 存取）
- ECL 私有（引擎側查無存取）：57 個（490 次）

⚠ 引擎側掃描是保守的：`di` 一旦被無法靜態求值的指令改動就停止歸屬，
所以「查無存取」代表**這個方法沒找到**，不代表一定不存在。

## 共用格子（每一格都必須逐格對上）

| ECL 位址 | 區 | bank 位移 | ECL 存取 | ECL 指令 | 引擎側存取點 |
|---|---|---|---|---|---|
| `4BF2h` | 0 | `bank0^[1E4h]` | 33 | COMPARE／SAVE | `overlay-02:0BBB`、`overlay-02:3691`、`overlay-02:3772`、`overlay-11:0056`、`overlay-11:06D8` |
| `7EE1h` | 1 | `bank1^[5C2h]` | 30 | SAVE | `overlay-02:0841`、`overlay-07:01FC`、`overlay-07:0591` |
| `7EC6h` | 1 | `bank1^[58Ch]` | 25 | SAVE | `overlay-09:1388`、`overlay-10:10DD` |
| `7ED2h` | 1 | `bank1^[5A4h]` | 23 | SAVE | `overlay-07:01FC`、`overlay-20:0C9C` |
| `7ED3h` | 1 | `bank1^[5A6h]` | 21 | SAVE | `overlay-07:01FC`、`overlay-20:0C9C` |
| `7ED5h` | 1 | `bank1^[5AAh]` | 8 | COMPARE／SAVE | `overlay-02:3691`、`overlay-14:078E` |
| `4BE6h` | 0 | `bank0^[1CCh]` | 6 | SAVE | `overlay-02:0C15`、`overlay-02:179A`、`overlay-02:2086`、`overlay-02:2F02`、`overlay-02:30DD`、`overlay-02:3772`、`overlay-04:0DD6`、`overlay-06:06F6`、`overlay-07:01FC`、`overlay-07:04BD`、`overlay-07:0591`、`overlay-07:0D70`、`overlay-10:102F`、`overlay-11:0056`、`overlay-11:06D8`、`overlay-16:3748`、`overlay-28:00CC` |
| `4BFBh` | 0 | `bank0^[1F6h]` | 6 | SAVE | `overlay-14:090A`、`overlay-24:2BAA`、`overlay-28:00CC` |
| `4BE7h` | 0 | `bank0^[1CEh]` | 4 | SAVE | `overlay-02:0C15` |
| `4BE8h` | 0 | `bank0^[1D0h]` | 4 | SAVE | `overlay-02:0C15` |
| `7F3Fh` | 1 | `bank1^[67Eh]` | 3 | GETTABLE | `overlay-02:019B` |
| `4CA1h` | 0 | `bank0^[342h]` | 2 | GETTABLE | `overlay-10:0FC7` |
| `7F71h` | 1 | `bank1^[6E2h]` | 2 | SAVE | `overlay-05:1736`、`overlay-23:123F` |
| `7EA8h` | 1 | `bank1^[550h]` | 2 | COMPARE／SAVE | `overlay-02:30DD`、`overlay-16:1F21`、`overlay-16:228E` |
| `4BC9h` | 0 | `bank0^[192h]` | 2 | COMPARE | `overlay-24:2BAA`、`overlay-30:018C` |
| `7ECAh` | 1 | `bank1^[594h]` | 2 | COMPARE | `overlay-02:179A`、`overlay-02:3772`、`overlay-14:083C`、`overlay-14:090A`、`overlay-24:2BAA` |
| `7F12h` | 1 | `bank1^[624h]` | 2 | SAVE | `overlay-16:3748`、`overlay-16:3EDD` |
| `4BF0h` | 0 | `bank0^[1E0h]` | 2 | SAVE | `overlay-02:3772` |
| `4BF1h` | 0 | `bank0^[1E2h]` | 2 | SAVE | `overlay-02:3772` |
| `7F70h` | 1 | `bank1^[6E0h]` | 1 | SAVE | `overlay-05:1736`、`overlay-23:123F` |
| `7ECBh` | 1 | `bank1^[596h]` | 1 | SAVE | `overlay-08:00F3`、`overlay-13:0000` |
| `7EC9h` | 1 | `bank1^[592h]` | 1 | COMPARE | `overlay-02:0C15`、`overlay-14:090A`、`overlay-14:0BAF` |
| `7EE6h` | 1 | `bank1^[5CCh]` | 1 | SAVE | `overlay-05:053C`、`overlay-05:0A7E`、`overlay-05:1736`、`overlay-07:1BE0` |
| `7EC7h` | 1 | `bank1^[58Eh]` | 1 | COMPARE | `overlay-05:053C`、`overlay-05:0A7E`、`overlay-05:1736` |

## ECL 私有（自洽即可，優先度低）

| ECL 位址 | 區 | bank 位移 | ECL 存取 | ECL 指令 |
|---|---|---|---|---|
| `7F79h` | 1 | `bank1^[6F2h]` | 167 | ADD／AND／COMPARE／COMPARE AND／DIVIDE／ENCOUNTER MENU／GETTABLE／LOAD CHARACTER／LOAD MONSTER／MULTIPLY／RANDOM／SAVE／SAVE TABLE／SUBTRACT |
| `7F7Ah` | 1 | `bank1^[6F4h]` | 42 | ADD／AND／COMPARE／DIVIDE／GETTABLE／LOAD MONSTER／MULTIPLY／PARTYSTRENGTH／RANDOM／SUBTRACT |
| `7F7Bh` | 1 | `bank1^[6F6h]` | 20 | ADD／AND／DIVIDE／GETTABLE／LOAD MONSTER／RANDOM |
| `4C01h` | 0 | `bank0^[202h]` | 17 | ADD／COMPARE／SAVE |
| `7F80h` | 1 | `bank1^[700h]` | 15 | ADD／COMPARE／DIVIDE／SAVE／TREASURE |
| `7B2Ch` | 2 | `bank2^[258h]` | 15 | ADD／RANDOM |
| `4C07h` | 0 | `bank0^[20Eh]` | 14 | ADD／COMPARE／MULTIPLY／RANDOM／SAVE |
| `7F7Eh` | 1 | `bank1^[6FCh]` | 14 | COMPARE／GETTABLE／LOAD MONSTER／SAVE |
| `4C06h` | 0 | `bank0^[20Ch]` | 13 | ADD／COMPARE／SAVE |
| `4C02h` | 0 | `bank0^[204h]` | 12 | COMPARE／SAVE |
| `4C05h` | 0 | `bank0^[20Ah]` | 12 | AND／COMPARE／DIVIDE／SUBTRACT |
| `7F7Dh` | 1 | `bank1^[6FAh]` | 12 | COMPARE／GETTABLE／LOAD MONSTER／SAVE |
| `4BFEh` | 0 | `bank0^[1FCh]` | 11 | SAVE |
| `7F7Ch` | 1 | `bank1^[6F8h]` | 11 | COMPARE／GETTABLE／SAVE／SETUP MONSTER |
| `4BFDh` | 0 | `bank0^[1FAh]` | 10 | SAVE |
| `4C00h` | 0 | `bank0^[200h]` | 7 | COMPARE AND／SAVE |
| `4CE1h` | 0 | `bank0^[3C2h]` | 6 | ADD／AND／SUBTRACT |
| `4C9Ch` | 0 | `bank0^[338h]` | 5 | SAVE |
| `4C42h` | 0 | `bank0^[284h]` | 5 | COMPARE |
| `7F7Fh` | 1 | `bank1^[6FEh]` | 5 | GETTABLE／SETUP MONSTER |
| `7F81h` | 1 | `bank1^[702h]` | 5 | ADD／TREASURE |
| `4C9Dh` | 0 | `bank0^[33Ah]` | 4 | GETTABLE |
| `4C17h` | 0 | `bank0^[22Eh]` | 4 | COMPARE／SAVE |
| `4C03h` | 0 | `bank0^[206h]` | 4 | COMPARE／SAVE |
| `4C09h` | 0 | `bank0^[212h]` | 4 | COMPARE／SAVE |
| `4CB8h` | 0 | `bank0^[370h]` | 4 | ADD／COMPARE |
| `4CA3h` | 0 | `bank0^[346h]` | 3 | ADD／COMPARE |
| `4C2Eh` | 0 | `bank0^[25Ch]` | 3 | COMPARE |
| `4C0Eh` | 0 | `bank0^[21Ch]` | 3 | ADD／COMPARE |
| `4C60h` | 0 | `bank0^[2C0h]` | 3 | COMPARE |
| `4C1Ah` | 0 | `bank0^[234h]` | 3 | ADD／COMPARE |
| `4CB9h` | 0 | `bank0^[372h]` | 3 | ADD／COMPARE |
| `4C04h` | 0 | `bank0^[208h]` | 2 | COMPARE／SAVE |
| `4C5Bh` | 0 | `bank0^[2B6h]` | 2 | COMPARE |
| `4C2Dh` | 0 | `bank0^[25Ah]` | 2 | COMPARE |
| `4BC6h` | 0 | `bank0^[18Ch]` | 2 | COMPARE |
| `4C0Ah` | 0 | `bank0^[214h]` | 2 | COMPARE／SAVE |
| `4CE4h` | 0 | `bank0^[3C8h]` | 2 | COMPARE |
| `4C18h` | 0 | `bank0^[230h]` | 2 | COMPARE／SAVE |
| `4CBAh` | 0 | `bank0^[374h]` | 2 | COMPARE／SAVE |
| `7F82h` | 1 | `bank1^[704h]` | 2 | SAVE |
| `4C81h` | 0 | `bank0^[302h]` | 1 | SAVE |
| `7CE4h` | 1 | `bank1^[1C8h]` | 1 | COMPARE |
| `4C13h` | 0 | `bank0^[226h]` | 1 | COMPARE AND |
| `4C2Ch` | 0 | `bank0^[258h]` | 1 | COMPARE |
| `7B09h` | 2 | `bank2^[212h]` | 1 | SAVE |
| `7D00h` | 1 | `bank1^[200h]` | 1 | COMPARE |
| `4C2Ah` | 0 | `bank0^[254h]` | 1 | COMPARE |
| `7CB8h` | 1 | `bank1^[170h]` | 1 | COMPARE |
| `4C47h` | 0 | `bank0^[28Eh]` | 1 | COMPARE |
| `4C5Eh` | 0 | `bank0^[2BCh]` | 1 | COMPARE |
| `4C61h` | 0 | `bank0^[2C2h]` | 1 | COMPARE |
| `4C62h` | 0 | `bank0^[2C4h]` | 1 | COMPARE |
| `4CBBh` | 0 | `bank0^[376h]` | 1 | SAVE |
| `4CBCh` | 0 | `bank0^[378h]` | 1 | COMPARE |
| `4C15h` | 0 | `bank0^[22Ah]` | 1 | SAVE |
| `4C16h` | 0 | `bank0^[22Ch]` | 1 | SAVE |
