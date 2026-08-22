# `CHECKFX` 的呼叫點：每個時機是誰在什麼時候問的

由 `cmd/checkfx-callsites` 產生，不要手改。理由與兩種呼叫的差別見該檔註解。

`CHECKFX` ＝ `overlay-23 entry#4`（模組內位移 `03FEh`）。時機清單見 [`checkfx-timing-table.md`](checkfx-timing-table.md)。

| 時機 | 呼叫點數 | 在哪 |
|---|---:|---|
| `00h` | 1 | `overlay-13`/sub_1144 @ `11BAh`（far） |
| `01h` | 1 | `overlay-13`/sub_1144 @ `117Ch`（far） |
| `04h` | 1 | `overlay-13`/sub_192 @ `023Dh`（far） |
| `05h` | 1 | `overlay-13`/sub_192 @ `024Bh`（far） |
| `06h` | 1 | `overlay-23` @ `1FEDh`（near） |
| `07h` | 1 | `overlay-08`/sub_26B @ `02A1h`（far） |
| `08h` | 1 | `overlay-10`/sub_1C3E @ `1D7Eh`（far） |
| `09h` | 3 | `overlay-22`/sub_4289 @ `42D5h`（far）、`overlay-22`/sub_45B5 @ `45DAh`（far）、`overlay-23` @ `2330h`（near） |
| `0Ah` | 1 | `overlay-23` @ `127Dh`（near） |
| `0Bh` | 1 | `overlay-22`/sub_F06 @ `0FFCh`（far） |
| `0Ch` | 1 | `overlay-23` @ `134Dh`（near） |
| `0Dh` | 4 | `overlay-12`/sub_299C @ `2A0Bh`（far）、`overlay-13`/sub_31D @ `0639h`（far）、`overlay-23` @ `008Ch`（near）、`overlay-23` @ `22CEh`（near） |
| `0Eh` | 1 | `overlay-09`/sub_DB1 @ `0DDBh`（far） |
| `0Fh` | 1 | `overlay-08`/sub_26B @ `033Bh`（far） |
| `10h` | 2 | `overlay-23` @ `120Bh`（near）、`overlay-23` @ `128Ah`（near） |
| `11h` | 1 | `overlay-09`/sub_1388 @ `141Dh`（far） |
| `12h` | 4 | `overlay-13`/sub_0 @ `005Eh`（far）、`overlay-13`/sub_124 @ `0179h`（far）、`overlay-13`/sub_DD9 @ `0E6Bh`（far）、`overlay-22`/sub_3804 @ `3900h`（far） |
| `13h` | 1 | `overlay-08`/sub_997 @ `09D2h`（far） |
| `14h` | 1 | `overlay-23` @ `2022h`（near） |
| `15h` | 1 | `overlay-08`/sub_26B @ `0357h`（far） |
| `16h` | 1 | `overlay-10`/sub_1C3E @ `1D8Ch`（far） |

共 30 處呼叫點。

## 沒有呼叫點的時機

`02h`、`03h` ——分派表裡有效果碼，但這一支找不到任何呼叫端。

⚠ **不要直接讀成「原作不會用到」。** 這一支只看得到兩種形狀：far-call 表裡的跨 overlay 呼叫，以及 overlay-23 內部的 `E8` 近呼叫，而且時機必須是 `mov al, imm` 推進去的。常駐執行檔那一側因為重定位，沒有辦法用同一個位元組樣式掃。**要下「這個時機是死的」這種結論，得先把常駐側與非立即數的呼叫都排除掉。**
