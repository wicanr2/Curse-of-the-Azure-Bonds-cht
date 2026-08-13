# ECL handler 的 operand 取用 vs 宣稱 arity

由 `scripts/ecl_handler_operand_audit.py` 產生，不要手改。

平台 `pc98`：dispatcher `373Eh`，opcode 來源 `ds:0A891h`，helper segment `0062h`（＝overlay-07，ECL2 單元的 control block）。

`ADDRESSVALUE` 等名稱來自 PC-98 的 Borland 除錯表，經 stub offset →
entry index → code offset 解析而得，不是猜的。DOS 沒有符號表，兩平台
的 stub 佈局相同才能沿用同一組名稱——這一點本身是 `strong inference`。

| opcode | 指令名（remake） | 宣稱 arity | `READVAR` 參數 | handler | `READVAR` | `ADDRESSVALUE` | `STOREVALUE` | `ADDFNC` | 一致 |
|---|---|---:|---|---|---:|---:|---:|---:|---|
| `00h` | EXIT | 0 | — | `0052h` | 0 | 0 | 0 | 0 | ✓ |
| `01h` | GOTO | 1 | 1 | `00E8h` | 1 | 0 | 0 | 1 | ✓ |
| `02h` | GOSUB | 1 | 1 | `0107h` | 1 | 0 | 0 | 0 | ✓ |
| `03h` | COMPARE | 2 | 2 | `011Eh` | 1 | 2 | 0 | 0 | ✓ |
| `04h` | ADD | 3 | 3 | `01B7h` | 1 | 1 | 1 | 1 | ✓ |
| `05h` | SUBTRACT | 3 | 3 | `01B7h` | 1 | 1 | 1 | 1 | ✓ |
| `06h` | DIVIDE | 3 | 3 | `01B7h` | 1 | 1 | 1 | 1 | ✓ |
| `07h` | MULTIPLY | 3 | 3 | `01B7h` | 1 | 1 | 1 | 1 | ✓ |
| `08h` | RANDOM | 2 | 2 | `025Ah` | 1 | 1 | 1 | 1 | ✓ |
| `09h` | SAVE | 2 | 2 | `02B8h` | 1 | 1 | 1 | 1 | ✓ |
| `0Ah` | LOAD CHARACTER | 1 | 1 | `0306h` | 1 | 1 | 0 | 0 | ✓ |
| `0Bh` | LOAD MONSTER | 3 | 3 | `049Ch` | 1 | 3 | 0 | 0 | ✓ |
| `0Ch` | SETUP MONSTER | 3 | 3 | `03E0h` | 1 | 3 | 0 | 0 | ✓ |
| `0Dh` | APPROACH | 0 | — | `0858h` | 0 | 0 | 0 | 0 | ✓ |
| `0Eh` | PICTURE | 1 | 1 | `08A7h` | 1 | 1 | 0 | 0 | ✓ |
| `0Fh` | INPUT NUMBER | 2 | 2 | `0992h` | 1 | 0 | 1 | 1 | ✓ |
| `10h` | INPUT STRING | 2 | 2 | `09DDh` | 1 | 0 | 0 | 1 | ✓ |
| `11h` | PRINT | 1 | 1 | `0A5Dh` | 1 | 1 | 0 | 0 | ✓ |
| `12h` | PRINTCLEAR | 1 | 1 | `0A5Dh` | 1 | 1 | 0 | 0 | ✓ |
| `13h` | RETURN | 0 | — | `0AF9h` | 0 | 0 | 0 | 0 | ✓ |
| `14h` | COMPARE AND | 4 | 4 | `0B49h` | 1 | 1 | 0 | 0 | ✓ |
| `15h` | VERTICAL MENU | 0 | 3 | `0F29h` | 1 | 0 | 0 | 1 | ✗ |
| `16h` | IF = | 0 | — | `0BAEh` | 0 | 0 | 0 | 0 | ✓ |
| `17h` | IF <> | 0 | — | `0BAEh` | 0 | 0 | 0 | 0 | ✓ |
| `18h` | IF < | 0 | — | `0BAEh` | 0 | 0 | 0 | 0 | ✓ |
| `19h` | IF > | 0 | — | `0BAEh` | 0 | 0 | 0 | 0 | ✓ |
| `1Ah` | IF <= | 0 | — | `0BAEh` | 0 | 0 | 0 | 0 | ✓ |
| `1Bh` | IF >= | 0 | — | `0BAEh` | 0 | 0 | 0 | 0 | ✓ |
| `1Ch` | CLEARMONSTERS | 0 | — | `1294h` | 0 | 0 | 0 | 0 | ✓ |
| `1Dh` | PARTYSTRENGTH | 1 | 1 | `12F7h` | 1 | 0 | 1 | 1 | ✓ |
| `1Eh` | CHECKPARTY | 6 | 6 | `149Ch` | 1 | 2 | 0 | 1 | ✓ |
| `20h` | NEWECL | 1 | 1 | `0C26h` | 1 | 1 | 0 | 0 | ✓ |
| `21h` | LOAD FILES | 3 | 3 | `0C81h` | 1 | 1 | 0 | 0 | ✓ |
| `22h` | PARTY SURPRISE | 2 | 2 | `16BCh` | 1 | 0 | 1 | 2 | ✓ |
| `23h` | SURPRISE | 4 | 4 | `1758h` | 1 | 1 | 0 | 0 | ✓ |
| `24h` | COMBAT | 0 | — | `1820h` | 0 | 0 | 0 | 0 | ✓ |
| `25h` | ON GOTO | 0 | — | `1B28h` | 0 | 0 | 0 | 0 | ✓ |
| `26h` | ON GOSUB | 0 | — | `1B28h` | 0 | 0 | 0 | 0 | ✓ |
| `27h` | TREASURE | 8 | 8 | `1BEAh` | 1 | 2 | 0 | 0 | ✓ |
| `28h` | ROB | 3 | 3 | `2125h` | 1 | 3 | 0 | 0 | ✓ |
| `29h` | ENCOUNTER MENU | 14 | 14 | `222Ch` | 1 | 6 | 15 | 1 | ✓ |
| `2Ah` | GETTABLE | 3 | 3 | `0E7Fh` | 1 | 1 | 0 | 2 | ✓ |
| `2Bh` | HORIZONTAL MENU | 0 | 2、? | `10C2h` | 2 | 1 | 1 | 1 | ✗ |
| `2Ch` | PARLAY | 6 | 6 | `2940h` | 1 | 1 | 1 | 1 | ✓ |
| `2Dh` | CALL | 1 | 1 | `30C6h` | 1 | 0 | 0 | 1 | ✓ |
| `2Eh` | DAMAGE | 5 | 5 | `2ACEh` | 1 | 5 | 0 | 0 | ✓ |
| `2Fh` | AND | 3 | 3 | `0E10h` | 1 | 2 | 0 | 0 | ✓ |
| `30h` | OR | 3 | 3 | `0E10h` | 1 | 2 | 0 | 0 | ✓ |
| `31h` | SPRITE OFF | 0 | — | `2E06h` | 0 | 0 | 0 | 0 | ✓ |
| `32h` | FIND ITEM | 1 | 1 | `29FBh` | 1 | 1 | 0 | 0 | ✓ |
| `33h` | PRINT RETURN | 0 | — | `2E61h` | 0 | 0 | 0 | 0 | ✓ |
| `34h` | ECL CLOCK | 2 | 2 | `2E2Ch` | 1 | 2 | 0 | 0 | ✓ |
| `35h` | SAVE TABLE | 3 | 3 | `0EDDh` | 1 | 2 | 1 | 1 | ✓ |
| `36h` | ADD NPC | 2 | 2 | `2F5Fh` | 1 | 2 | 0 | 0 | ✓ |
| `37h` | LOAD PIECES | 3 | 3 | `0C81h` | 1 | 1 | 0 | 0 | ✓ |
| `38h` | PROGRAM | 1 | 1 | `3393h` | 1 | 1 | 0 | 0 | ✓ |
| `39h` | WHO | 1 | 1 | `2EDAh` | 1 | 0 | 0 | 0 | ✓ |
| `3Ah` | DELAY | 0 | — | `2AA7h` | 0 | 0 | 0 | 0 | ✓ |
| `3Bh` | SPELL | 3 | 3 | `2FDAh` | 1 | 1 | 2 | 2 | ✓ |
| `3Ch` | PROTECTION | 1 | 1 | `3520h` | 1 | 0 | 0 | 0 | ✓ |
| `3Dh` | CLEAR BOX | 0 | — | `2E8Ch` | 0 | 0 | 0 | 0 | ✓ |
| `3Eh` | DUMP | 0 | — | `35AFh` | 0 | 0 | 0 | 0 | ✓ |
| `3Fh` | FIND SPECIAL | 1 | 1 | `364Bh` | 1 | 1 | 0 | 0 | ✓ |
| `40h` | DESTROY ITEMS | 1 | 1 | `369Fh` | 1 | 1 | 0 | 0 | ✓ |

`READVAR` 參數與宣稱 arity 一致的有 62／64 個 opcode。

不一致的 opcode：`15h`、`2Bh`。

不一致的都是**變長指令**：原版先 `READVAR(固定前綴)`，解出一個 operand
得到後續數量 N，`dec` ECL PC 回退一格，再 `READVAR(N)`。`KnownCommands`
把它們宣告成 0 並由 parser 特別處理，因此不是錯誤，但固定前綴的長度
（本表第四欄的第一個數字）是原版事實，remake 應該對得上。
