# `CHECKFX` 的時機分派表

`CHECKFX(timing, subject)` 是一張表：每個 timing 對應一組 effect id，
逐一交給 effect 鏈遍歷（`sub_269`，[spec 577](../spec/577-attempttohit-and-effect-chain-walk.md)）
去看目標身上有沒有該效果，有就 `CALLEFFECT` 分派。

所以規則各處的 `CHECKFX(0Ah)`／`CHECKFX(0Ch)`／`CHECKFX(06h)` 問的都是
「這個時機有哪些效果要介入」。

由 `scripts/checkfx_timing_table.py` 從 IDAPython 匯出的逐指令序列解析，
判準是純結構的（`cmp al, N` ／ `jnz` 的目標即下一個 case 起點）。

| timing | effect id |
|---|---|
| `00h` | （無） |
| `01h` | `25h`、`19h`、`47h`、`45h` |
| `02h` | `4Fh`、`50h`、`91h`、`39h`、`60h`、`7Ah`、`7Bh` |
| `03h` | `40h`、`41h`、`42h`、`43h`、`46h`、`4Fh`、`57h` |
| `04h` | `1Dh`、`06h`、`67h`、`4Bh`、`4Ch`、`86h` |
| `05h` | `1Ch`、`29h`、`68h`、`78h`、`65h`、`73h`、`74h`、`77h`、`5Eh`、`75h`、`3Ch`、`51h`、`52h`、`55h`、`82h`、`8Fh` |
| `06h` | `71h`、`3Dh`、`0Ah`、`14h`、`69h`、`6Ah`、`70h`、`72h`、`76h`、`11h`、`5Dh`、`65h`、`1Ch`、`6Eh`、`49h`、`52h`、`54h`、`81h`、`85h`、`87h`、`3Fh` |
| `07h` | `33h`、`34h`、`35h`、`1Fh`、`03h`、`1Bh`、`88h` |
| `08h` | `63h`、`52h`、`59h`、`48h`、`38h` |
| `09h` | `69h`、`6Ah`、`6Bh`、`6Ch`、`6Dh`、`6Eh`、`6Fh`、`70h`、`7Ch`、`7Dh`、`3Fh`、`81h` |
| `0Ah` | `01h`、`02h`、`21h`、`24h`、`31h`、`06h`、`12h`、`1Ah`、`4Bh`、`4Ch` |
| `0Bh` | `21h`、`11h`、`08h`、`09h`、`2Dh`、`2Eh`、`1Eh`、`07h` |
| `0Ch` | `08h`、`09h`、`0Ah`、`11h`、`14h`、`21h`、`24h`、`2Dh`、`2Eh`、`31h`、`3Dh`、`6Fh`、`7Dh`、`61h`、`32h`、`36h` |
| `0Dh` | `63h`、`64h`、`4Bh` |
| `0Eh` | `53h`、`58h`、`79h`、`56h`、`57h`、`5Ah`、`7Eh`、`80h`、`83h`、`84h`、`8Bh` |
| `0Fh` | `15h`、`1Eh`、`0Bh`、`0Dh`、`4Dh` |
| `10h` | `19h`、`47h`、`25h`、`2Fh`、`30h`、`59h`、`04h` |
| `11h` | `01h`、`02h`、`0Bh` |
| `12h` | `27h`、`2Ah`、`3Ah` |
| `13h` | `62h`、`17h`、`48h`、`38h`、`0Bh` |
| `14h` | `32h`、`36h` |
| `15h` | `23h` |
| `16h` | `8Ah` |
| `17h` | `4Ah` |

## 已知的呼叫點

| 呼叫處 | timing | 出處 |
|---|---|---|
| `PUTDAMAGE` 進入時 | `06h` | [581](../spec/581-putdamage-pipeline.md) |
| `PUTDAMAGE` 無豁免時 | `14h` | 同上 |
| `PUTEFFECT` | `09h` | 同上 |
| `ATTEMPTTOHIT` 對目標 | `0Ah` | [577](../spec/577-attempttohit-and-effect-chain-walk.md) |
| `ATTEMPTTOHIT` 對攻擊者 | `10h` | 同上 |
| `MAKESAVE` | `0Ch` | [582](../spec/582-makesave-and-losedude.md) |
| `KILLDUDE`／`PUTDAMAGE` 死亡後 | `0Dh` | [579](../spec/579-character-status-fields.md) |
