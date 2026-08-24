# myth-drannor.inner-ruins

由 `cmd/map-atlas` 產生，不要手改。`GEO6` 區塊 `0x43`；腳本 `ECL6/0x43`。

```text
       0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 
y= 0   .  .  .  .  .  .  .  .  k  .  .  .  .  .  .  .  
y= 1   .  .  .  .  .  .  q  .  .  .  j  .  .  .  i  .  
y= 2   .  .  .  .  .  .  .  .  .  l  .  .  .  .  .  i  
y= 3   .  .  p  .  .  .  .  .  .  .  .  .  .  .  .  h  
y= 4   .  .  .  .  .  .  .  .  .  .  .  .  .  .  h  .  
y= 5   .  .  o  .  .  .  .  .  .  .  m  .  .  .  g  .  
y= 6   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  g  
y= 7   .  .  .  .  .  .  .  .  .  .  n  .  .  .  .  f  
y= 8   .  .  .  .  .  .  .  .  .  .  .  .  .  .  f  .  
y= 9   .  7  .  .  8  .  .  .  .  .  .  .  .  .  .  .  
y=10   6  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  
y=11   .  .  .  .  .  5  4  3  .  .  .  .  .  .  e  .  
y=12   .  .  .  .  .  .  .  .  .  a  a  b  b  .  .  .  
y=13   .  .  .  .  .  .  .  2  9  9  .  .  .  .  .  .  
y=14   .  .  .  .  .  .  .  .  .  .  .  .  .  c  .  .  
y=15   .  .  .  .  .  .  1  .  .  .  .  .  .  .  .  d  
```

`.` 沒有事件的地面　`#` 四面都走不出去　`^v<>` 這一格往那個方向走得出地圖　`@` 宣告的進場錨點

## 離開這張圖的格子

這張圖沒有走得出去的邊界格。

## 每格事件

⚠ **有索引不等於站上去就會演**：處理常式自己可能還有守衛（一次性旗標、`RANDOM`、前置劇情、`SEARCH`）。實際演出來什麼看[`cell-sweep`](../../audit/cell-sweep.md)。

| 圖上的字 | 索引 | 格子 | 守衛 | 那一場的第一句 |
|---|---:|---|---|---|
| `1` | 1 | `(6,15)` | `CALL 2E10` | — |
| `2` | 2 | `(7,13)` | `COMPARE 4C00 01 / IF = / EXIT / SAVE 01 4C00` | AS YOU ENTER, YOU HEAR A VOICE. 'FINALLY YOU HAVE |
| `3` | 3 | `(7,11)` | `COMPARE 4C00 01 / IF = / EXIT / SAVE 01 4C00` | AS YOU ENTER, YOU HEAR A VOICE. 'FINALLY YOU HAVE |
| `4` | 4 | `(6,11)` | `COMPARE 4C00 01 / IF = / EXIT / SAVE 01 4C00` | AS YOU ENTER, YOU HEAR A VOICE. 'FINALLY YOU HAVE |
| `5` | 5 | `(5,11)` | `COMPARE 4C00 01 / IF = / EXIT / SAVE 01 4C00` | AS YOU ENTER, YOU HEAR A VOICE. 'FINALLY YOU HAVE |
| `6` | 6 | `(0,10)` | `CALL 2E10` | — |
| `7` | 7 | `(1,9)` | `COMPARE 4C01 01 / IF = / EXIT / SAVE 01 4C01` | A SULFEROUS SMELL ASSAULTS YOUR NOSTRILS. THE ROOM |
| `8` | 8 | `(4,9)` | `COMPARE 4C02 01 / IF = / EXIT / SAVE 01 4C02` | THE ROOM SEEMS FILLED WITH BONES AND HIDEOUS |
| `9` | 9 | `(8,13)`、`(9,13)` | `COMPARE 4C03 01 / IF = / EXIT / SAVE 01 4C03` | YOU HAVE WALKED INTO AN ELEGANT PRIVATE CHAPEL. |
| `a` | 10 | `(9,12)`、`(10,12)` | `COMPARE 4C04 01 / IF = / EXIT / SAVE 01 4C04` | YOU HAVE ENTERED AN ELEGANT BEDROOM WITH A |
| `b` | 11 | `(11,12)`、`(12,12)` | `COMPARE 4C05 01 / IF = / EXIT / SAVE 01 4C05` | THIS ROOM HAS BEEN CONVERTED TO AN OFFICE. THE |
| `c` | 12 | `(13,14)` | `COMPARE AND 4C06 00 4C07 00 / IF <> / EXIT / SAVE 01 4C06` | YOU HAVE COME INTO THE KITCHEN. SLAVES DIVE UNDER |
| `d` | 13 | `(15,15)` | `COMPARE 4C07 00 / IF = / EXIT` | THE SEWER HAS COLLAPSED. THIS IS NO LONGER A WAY OUT. |
| `e` | 14 | `(14,11)` | `COMPARE AND 4C08 00 4C00 00 / IF <> / EXIT / SAVE 01 4C08` | TIERED BEDS LINE THE WALLS, FILLED WITH MEDITATING |
| `f` | 15 | `(15,7)`、`(14,8)` | `COMPARE AND 4C09 00 4C00 00 / IF <> / EXIT / SAVE 01 4C09` | THE ROOM IS FILLED WITH WORSHIPPING PRIESTS. THEY |
| `g` | 16 | `(14,5)`、`(15,6)` | `COMPARE 4C0A 01 / IF = / EXIT / SAVE 01 4C0A` | A HIGH PRIEST IS HERE SUMMONING UP THE POWER OF |
| `h` | 17 | `(15,3)`、`(14,4)` | `COMPARE 4C0B 01 / IF = / EXIT / SAVE 01 4C0B` | RATS SKITTER UNDER BAGS AS YOU OPEN THE DOOR INTO |
| `i` | 18 | `(14,1)`、`(15,2)` | `COMPARE 4C0C 01 / IF = / EXIT` | THIS ROOM IS LINED WITH SHELVES OF MOULDERING BOOKS. |
| `j` | 19 | `(10,1)` | `EXIT / COMPARE 4C0E 01 / IF <> / EXIT` | THE ROOM CONTAINS OLD BIERS AND CASKETS. |
| `k` | 20 | `(8,0)` | `COMPARE 4C0E 01 / IF <> / EXIT / PRINTCLEAR 00` | THE ROOM CONTAINS OLD BIERS AND CASKETS. |
| `l` | 21 | `(9,2)` | `COMPARE AND 4C0F 00 4C0E 00 / IF <> / EXIT / SAVE 01 4C0F` | THE ROOM CONTAINS OLD BIERS AND CASKETS. |
| `m` | 22 | `(10,5)` | `COMPARE 4C10 01 / IF = / EXIT / SAVE 01 4C10` | THE STENCH OF PRESERVING FLUIDS IS STRONG IN HERE. |
| `n` | 23 | `(10,7)` | `—` | STAIRS LEAD UP HERE. DO YOU WANT TO GO UP? |
| `o` | 24 | `(2,5)` | `—` | STAIRS LEAD DOWN HERE. DO YOU WANT TO DESCEND? |
| `p` | 25 | `(2,3)` | `CALL 2E10` | — |
| `q` | 26 | `(6,1)` | `SETUP MONSTER 47 00 47` | 'THE POWER OF YOUR BONDS HAS RETURNED. GROVEL AT |
