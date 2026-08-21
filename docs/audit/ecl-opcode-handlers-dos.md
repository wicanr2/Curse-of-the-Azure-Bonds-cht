# DOS `overlay-02` 的 ECL opcode → handler 對照表

由 `sub_3377`（opcode 分派器）逐條讀出。重產方式：

```sh
tools/ida.sh py workplace/re-sweep/dos/overlays/overlay-02.bin.i64 \
  dump_function.py /work/dispatch.json 3377 3861
```

⚠ **表裡的位址範圍與條數以 IDA 的函式邊界為準，而 IDA 會切錯。** 已知三處：
`21h`／`37h` 共用的 `0C15h` 真正結束在 `0DA3h`（131 條，IDA 只認到 `0D4Ah`／104 條，
spec 1153）、`2Fh`／`30h` 共用的 `0DA4h` 真正結束在 `0E12h`（43 條，IDA 拆成 11＋32，
spec 1157）。**收尾一支之前先各自驗一次邊界**，方法是往後 dump 到看見 `retf`。

⚠ **同一個位址出現在兩列代表兩個 opcode 共用一支 handler**，那種 handler 會
**再讀一次 `DS:75FFh`（目前指令碼）分辨自己被誰呼叫**（spec 587）。只實作其中
一個 opcode 的行為會讓另一個靜靜地做錯事。

★ 存在的理由：剩下的 `partial` opcode 要一支一支收尾，而「那一支在哪、多大」
以前每次都要重找。有了這張表就能先排順序——`33h PRINT RETURN` 只有 14 條，
`2Dh CALL` 有 124 條，成本差一個數量級。

⚠ 分派器沒有列出的 opcode（如 `04h`..`06h`、`11h`、`16h`..`1Bh` 的比較指令）
是在分派之前就被主迴圈處理掉的，不代表它們不存在。

| opcode | 助憶碼 | 運算元數 | handler | 大小 | 條數 | 副作用狀態 |
|---|---|---:|---|---|---:|---|
| `00h` | EXIT | 0 | `0052h` | `0052h`..`00E8h` | 49 | `done` |
| `01h` | GOTO | 1 | `00E8h` | `00E8h`..`0107h` | 14 | `done` |
| `02h` | GOSUB | 1 | `0107h` | `0107h`..`011Eh` | 11 | `done` |
| `03h` | COMPARE | 2 | `011Eh` | `011Eh`..`019Bh` | 50 | `done` |
| `07h` | MULTIPLY | 3 | `019Bh` | `019Bh`..`0244h` | 64 | `done` |
| `08h` | RANDOM | 2 | `0244h` | `0244h`..`02A2h` | 35 | `done` |
| `09h` | SAVE | 2 | `02A2h` | `02A2h`..`02F0h` | 30 | `done` |
| `0Ah` | LOAD CHARACTER | 1 | `02F0h` | `02F0h`..`03CAh` | 73 | `done` |
| `0Bh` | LOAD MONSTER | 3 | `0466h` | `0466h`..`06D5h` | 284 | `done` |
| `0Ch` | SETUP MONSTER | 3 | `03CAh` | `03CAh`..`0461h` | 52 | `done` |
| `0Dh` | APPROACH | 0 | `0801h` | `0801h`..`083Dh` | 22 | `done` |
| `0Eh` | PICTURE | 1 | `0841h` | `0841h`..`092Ch` | 77 | `partial` |
| `0Fh` | INPUT NUMBER | 2 | `092Ch` | `092Ch`..`0970h` | 28 | `consumed` |
| `10h` | INPUT STRING | 2 | `0972h` | `0972h`..`09EAh` | 51 | `done` |
| `12h` | PRINTCLEAR | 1 | `09EAh` | `09EAh`..`0A86h` | 69 | `done` |
| `13h` | RETURN | 0 | `0A86h` | `0A86h`..`0AD6h` | 30 | `done` |
| `14h` | COMPARE AND | 4 | `0AD6h` | `0AD6h`..`0B3Bh` | 40 | `done` |
| `15h` | VERTICAL MENU | 0 | `0EBDh` | `0EBDh`..`0F0Fh` | 29 | `done` |
| `1Bh` | IF >= | 0 | `0B3Bh` | `0B3Bh`..`0BBBh` | 43 | `done` |
| `1Ch` | CLEARMONSTERS | 0 | `120Eh` | `120Eh`..`1271h` | 37 | `partial` |
| `1Dh` | PARTYSTRENGTH | 1 | `1271h` | `1271h`..`13B6h` | 120 | `done` |
| `1Eh` | CHECKPARTY | 6 | `1416h` | `1416h`..`14FBh` | 141 | `done` |
| `20h` | NEWECL | 1 | `0BBBh` | `0BBBh`..`0C15h` | 33 | `done` |
| `22h` | PARTY SURPRISE | 2 | `1636h` | `1636h`..`1654h` | 11 | `done` |
| `23h` | SURPRISE | 4 | `16D2h` | `16D2h`..`1777h` | 66 | `consumed` |
| `24h` | COMBAT | 0 | `179Ah` | `179Ah`..`17B5h` | 18 | `partial` |
| `26h` | ON GOSUB | 0 | `1A9Bh` | `1A9Bh`..`1B30h` | 57 | `done` |
| `27h` | TREASURE | 8 | `1B53h` | `1B53h`..`1F46h` | 398 | `partial` |
| `28h` | ROB | 3 | `1F46h` | `1F46h`..`2029h` | 78 | `done` |
| `29h` | ENCOUNTER MENU | 14 | `2086h` | `2086h`..`2785h` | 649 | `done` |
| `2Ah` | GETTABLE | 3 | `0E13h` | `0E13h`..`0E71h` | 35 | `done` |
| `2Bh` | HORIZONTAL MENU | 0 | `1082h` | `1082h`..`120Eh` | 161 | `done` |
| `2Ch` | PARLAY | 6 | `27A8h` | `27A8h`..`2847h` | 66 | `done` |
| `2Dh` | CALL | 1 | `2F02h` | `2F02h`..`3073h` | 124 | `partial` |
| `2Eh` | DAMAGE | 5 | `2942h` | `2942h`..`2C8Fh` | 305 | `partial` |
| `2Fh` | AND | 3 | `0DA4h` | `0DA4h`..`0E12h` | 43 | `done` |
| `30h` | OR | 3 | `0DA4h` | `0DA4h`..`0E12h` | 43 | `done` |
| `31h` | SPRITE OFF | 0 | `2C8Fh` | `2C8Fh`..`2CB5h` | 12 | `done` |
| `32h` | FIND ITEM | 1 | `2847h` | `2847h`..`28F3h` | 61 | `done` |
| `33h` | PRINT RETURN | 0 | `2CEAh` | `2CEAh`..`2D15h` | 14 | `partial` |
| `34h` | ECL CLOCK | 2 | `2CB5h` | `2CB5h`..`2CEAh` | 22 | `done` |
| `35h` | SAVE TABLE | 3 | `0E71h` | `0E71h`..`0EBDh` | 29 | `done` |
| `36h` | ADD NPC | 2 | `2DA9h` | `2DA9h`..`2E16h` | 38 | `done` |
| `37h` | LOAD PIECES | 3 | `0C15h` | `0C15h`..`0DA3h` | 131 | `partial` |
| `38h` | PROGRAM | 1 | `30DDh` | `30DDh`..`321Eh` | 104 | `done` |
| `39h` | WHO | 1 | `2D5Eh` | `2D5Eh`..`2DA9h` | 36 | `done` |
| `3Ah` | DELAY | 0 | `28F3h` | `28F3h`..`2903h` | 7 | `done` |
| `3Bh` | SPELL | 3 | `2E16h` | `2E16h`..`2F02h` | 83 | `done` |
| `3Ch` | PROTECTION | 1 | `321Fh` | `321Fh`..`3251h` | 20 | `done` |
| `3Dh` | CLEAR BOX | 0 | `2D15h` | `2D15h`..`2D5Eh` | 23 | `done` |
| `3Eh` | DUMP | 0 | `3251h` | `3251h`..`3284h` | 19 | `done` |
| `3Fh` | FIND SPECIAL | 1 | `3284h` | `3284h`..`32D8h` | 34 | `done` |
| `40h` | DESTROY ITEMS | 1 | `32D8h` | `32D8h`..`3377h` | 56 | `done` |

## 摘要

| 項目 | 數 |
|---|---:|
| 分派器列出的 opcode | 52 |
| handler 指令數合計 | 4023 |
| 其中副作用仍是 `partial` | 9（合計 1181 條）|
