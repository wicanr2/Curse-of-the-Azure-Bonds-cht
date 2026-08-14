# 1083 — ECL 助憶碼表：原作自己留下的 65 個指令名稱，`29h` 是 ENCOUNTER MENU 不是 PARLAY

- 證據等級：`exact`（DOS 側 774 條逐條讀完，無匯出破洞）
- 作法見 spec 783

## `dos overlay-34:00240h`（`retf 2`）

原本待解讀。範圍 `240h`..`904h`。**整支就是一個 65 路的 `case`**：
把目前的 ECL 指令碼 `DS:75FFh` 換成一個字串常數寫進 `arg_2`（上限 `0FFh`）。

指令種類只有 `push`／`mov`／`les`／`cmp`／`jnz`／`jmp`／`call`／`xor`／`retf`／`pop`，
64 次 `cmp` ＋ 64 次 `call`，每一支都是同一個形狀：

```asm
cmp  ax, N
jnz  下一個
mov  di, offset <字串N> ; push cs ; push di
les  di, [bp+arg_2]     ; push es ; push di
mov  ax, 0FFh           ; push ax
call far 0A54h:064Eh                       ; 上限指派
jmp  出口
```

> ★★★ **這是原作開發者自己寫的 ECL 指令名稱**，不是後人推測的。
> 它把既有幾十份 ECL 規格用的名字（`DAMAGE`、`TREASURE`、`WHO`、`DUMP`…）
> 一次全部釘到確切的指令碼上。
> ★ 形狀上這是**留在遊戲裡的除錯／反組譯輸出**（指令碼 `3Eh` 自己就叫 `DUMP`）。

## ★★★ 指令碼 → 助憶碼（完整 65 筆）

| 碼 | 助憶碼 | | 碼 | 助憶碼 |
|---|---|---|---|---|
| `00` | `EXIT` | | `21h` | `LOAD FILES` |
| `01` | `GOTO` | | `22h` | `PARTY SURPRISE` |
| `02` | `GOSUB` | | `23h` | `SURPRISE` |
| `03` | `COMPARE` | | `24h` | `COMBAT` |
| `04` | `ADD` | | `25h` | `ON GOTO` |
| `05` | `SUBTRAT`（原文就少一個 C） | | `26h` | `ON GOSUB` |
| `06` | `DIVIDE` | | `27h` | `TREASURE` |
| `07` | `MULTIPLY` | | `28h` | `ROB` |
| `08` | `RANDOM` | | **`29h`** | **`ENCOUNTER MENU`** |
| `09` | `SAVE` | | `2Ah` | `GETTABLE` |
| `0Ah` | `LOAD CHARACTER` | | `2Bh` | `HORIZONTAL MENU` |
| `0Bh` | `LOAD MONSTER` | | **`2Ch`** | **`PARLAY`** |
| `0Ch` | `SETUP MONSTER` | | `2Dh` | `CALL` |
| `0Dh` | `APPROACH` | | `2Eh` | `DAMAGE` |
| `0Eh` | `PICTURE` | | `2Fh` | `AND` |
| `0Fh` | `INPUT NUMBER` | | `30h` | `OR` |
| `10h` | `INPUT STRING` | | `31h` | `SPRITE OFF` |
| `11h` | `PRINT` | | `32h` | `FIND ITEM` |
| `12h` | `PRINTCLEAR` | | `33h` | `PRINT RETURN` |
| `13h` | `RETURN` | | `34h` | `ECL CLOCK` |
| `14h` | `COMPARE AND` | | `35h` | `SAVE TABLE` |
| `15h` | `VERTICAL MENU` | | `36h` | `ADD NPC` |
| `16h` | `IF = `（尾端有空白） | | **`37h`** | **`LOAD PIECES`** |
| `17h` | `IF <>` | | `38h` | `PROGRAM` |
| `18h` | `IF <` | | `39h` | `WHO` |
| `19h` | `IF >` | | `3Ah` | `DELAY` |
| `1Ah` | `IF <=` | | `3Bh` | `SPELL` |
| `1Bh` | `IF >=` | | `3Ch` | `PROTECTION` |
| `1Ch` | `CLEARMONSTERS` | | `3Dh` | `CLEAR BOX` |
| `1Dh` | `PARTYSTRENGTH` | | `3Eh` | `DUMP` |
| `1Eh` | `CHECKPARTY` | | `3Fh` | `FIND SPECIAL` |
| **`1Fh`** | **沒有分支** | | `40h` | `DESTROY ITEMS` |
| `20h` | `NEWECL` | | | |

> ★★★ **`1Fh` 沒有任何分支**：落到這個指令碼時 `arg_2` **完全不被寫入**，
> 呼叫端拿到的是上一次的內容。⇒ **`1Fh` 是未使用的指令碼**（或它的名稱
> 在某次改版被拿掉了）。remake 的反組譯器要把 `1Fh` 標成未定義，不要留空。
>
> ★★★ **常數在 `CS:` 裡的排列順序與指令碼順序不一致**：
> `'LOAD PIECES'`（`CS:129h`）夾在 `'LOAD FILES'`（`21h`）與
> `'PARTY SURPRISE'`（`22h`）之間，但它對應的是 **`37h`**。
> ⇒ **不能靠「字串在資料段裡的順序」推指令碼**，一定要看 `cmp` 的那個常數。
> 這正是「二手推論（排列順序）輸給一手資料（比較指令）」的例子。

## ★★★ 更正既有規格：`29h` 的名字

spec 611 與 spec 1041 把 `29h` 叫做 **PARLAY**。**本表推翻這個名字**：

- `29h` ＝ **`ENCOUNTER MENU`**
- `2Ch` ＝ **`PARLAY`**（另一個指令）

⚠ 但兩份規格的**行為描述是對的**：spec 1041 從 DOS 側讀出來的選項行是
`'~COMBAT ~WAIT ~FLEE ~PARLAY'` ／ `'~COMBAT ~WAIT ~FLEE ~ADVANCE'`
——那本來就是**遭遇選單**，`PARLAY` 只是其中一個選項。
⇒ **原作的命名與 spec 1041 的一手證據互相印證**；錯的只有規格上的標題。

★ 對照 spec 1039：`2Bh` ＝ `HORIZONTAL MENU` **完全吻合**。

## 中文化

65 個助憶碼**全部是除錯輸出，不面向玩家，一律不要翻譯**。
⚠ `'SUBTRAT'`（少一個 `C`）與 `'IF = '`（尾端多一個空白）是原作的瑕疵，
照抄即可——改了反而對不上原版的 dump 輸出。

## 明確不宣稱

- 沒有宣稱 `DS:75FFh` 是由誰寫入的（本支只讀）。
- 沒有宣稱 `arg_2`（輸出字串）拿去印在哪裡。
- 沒有宣稱這 65 個名稱各自的參數格式（既有的 ECL 規格各自涵蓋）。
- 沒有宣稱 `1Fh` 在直譯器裡實際會做什麼（本支只是沒有名字）。
- 沒有宣稱指令碼 `40h` 之後還有沒有指令。
- 沒有宣稱 PC-98 側有沒有同一張表。
