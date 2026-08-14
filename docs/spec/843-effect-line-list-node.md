# 843 — 狀態效果清單：顯示行的鏈結節點，與那 21 個效果名稱

- 證據等級：`exact`（`overlay-15:00E69h` 兩平台逐條讀完；字串清單由原始 bytes 取出）
- 作法見 spec 783

## `overlay-15:00E69h` ↔ `overlay-15:00E54h`（entry#14，兩平台各 49 條）

巢狀程序（呼叫端多推一個 `bp` 當靜態鏈）。`retf 8` ＝ 4 個 word：
`(屬性, 字串)` ＋ 靜態鏈。

```pascal
procedure 加一行(屬性: byte; 字串: string);
begin
    暫存 := Copy(字串, 28h);                  { DOS 40 字元／PC-98 3Ch ＝ 60 }
    GetMem(尾^.next, 2Eh);                    { DOS 46 bytes／PC-98 56h ＝ 86 }
    尾 := 尾^.next;
    尾^.next := NIL;
    尾^.文字 := 暫存;
    尾^.屬性 := 屬性;
end;
```

`尾` 不是參數，是**外層函式的區域變數**——靠 `mov di, [bp+靜態鏈]; les di, ss:[di−0Ch]`
取到。這是 Turbo Pascal 巢狀程序存取外層變數的標準碼型。

## 顯示行節點的版面

| 位移（DOS） | 位移（PC-98） | 內容 |
|---|---|---|
| `+00h`..`+27h` | `+00h`..`+3Bh` | 文字（Pascal 字串，DOS 上限 40、PC-98 上限 60） |
| `+29h` | `+51h` | 屬性（一個 byte，呼叫點都傳 0） |
| `+2Ah` | `+52h` | next |
| 節點大小 | | DOS `2Eh` ＝ 46／PC-98 `56h` ＝ 86 |

**PC-98 把字串上限從 40 拉到 60**，節點也跟著從 46 變 86。中文化若沿用 DOS 版面，
一行只能放 20 個全形字；照 PC-98 的 60 bytes 才有 30 個全形字的空間。

## 那 21 個效果名稱（`overlay-15` 內嵌，DOS `0EECh` 起）

`overlay-15:01016h`（505 條，**尚未讀完**）以 `offset loc_EEC` 為起點，把這一串
變長 Pascal 字串當成清單走。字串本體逐條取出如下：

| # | DOS 位移 | 名稱 |
|---|---|---|
| 0 | `0EEDh` | `Funky--` |
| 1 | `0EF5h` | `Dispel Evil` |
| 2 | `0F01h` | `Faerie Fire` |
| 3 | `0F0Dh` | `Fumbling` |
| 4 | `0F16h` | `Helpless` |
| 5 | `0F1Fh` | `Confused` |
| 6 | `0F28h` | `Cause Disease` |
| 7 | `0F36h` | `Hot Fire Shield` |
| 8 | `0F46h` | `Cold Fire Shield` |
| 9 | `0F57h` | `Poisoned` |
| 10 | `0F60h` | `Regenerating` |
| 11 | `0F6Dh` | `Fire Resistance` |
| 12 | `0F7Dh` | `Minor Globe of Invulnerability` |
| 13 | `0F9Ch` | `enfeebled` |
| 14 | `0FA6h` | `invisible to animals` |
| 15 | `0FBBh` | `Invisible` |
| 16 | `0FC5h` | `Camouflaged` |
| 17 | `0FD1h` | `protected from dragon breath` |
| 18 | `0FEEh` | `berserk` |
| 19 | `0FF6h` | `Displaced` |
| 20 | `1000h` | ` <No Spell Effects>` |

**這是變長字串接續排放，不是固定 stride 的表**——所以編號 0..20 是「排列次序」，
**不等於引擎內部的效果碼**。spec 831／832／837／842 列出的那些效果碼
（`1Eh`／`20h`／`23h`／`3Fh`／`6Fh`／`7Dh`／`81h`／`85h`／`89h`／`8Eh`）
**目前都還沒有對應到這張清單的任何一列**，要等 `overlay-15:01016h` 讀完才能定案。

`Funky--` 是佔位用的名字（第 0 列），`' <No Spell Effects>'` 開頭有一個半形空白。

中文化注意：`berserk` 這一列與 spec 831 的 `'goes berserk'` **是兩個獨立字串**；
`Poisoned` 與 spec 796 的 `'is not poisoned.'`、spec 829 的 `'is Poisoned'` 也各自獨立。

## 明確不宣稱

- **沒有宣稱編號 0..20 就是效果碼。**上表只是字串在 overlay 裡的排列次序。
- 沒有宣稱 `+29h`（屬性）的用途——已讀到的三個呼叫點都傳 0。
- 沒有宣稱這條鏈由誰釋放。
