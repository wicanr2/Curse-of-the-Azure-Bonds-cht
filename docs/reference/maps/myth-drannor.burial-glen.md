# myth-drannor.burial-glen

由 `cmd/map-atlas` 產生，不要手改。`GEO6` 區塊 `0x40`；腳本 `ECL6/0x40`。

```text
       0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 
y= 0   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ← 走得出去：(0,0)^ (0,0)< (1,0)^ (2,0)^ (3,0)^ (4,0)^ (5,0)^ (6,0)^ (7,0)^ (8,0)^ (9,0)^ (10,0)^ (11,0)^ (12,0)^ (13,0)^ (14,0)^ (15,0)^ (15,0)>
y= 1   .  m  m  9  m  .  .  .  .  .  i  .  .  .  .  .  ← 走得出去：(0,1)< (15,1)>
y= 2   .  a  m  m  m  8  .  .  .  h  .  .  .  .  .  .  ← 走得出去：(0,2)< (15,2)>
y= 3   .  b  .  .  8  8  .  .  .  .  .  .  .  .  7  .  ← 走得出去：(0,3)< (15,3)>
y= 4   .  .  .  .  .  .  .  .  .  .  4  .  .  6  .  .  ← 走得出去：(0,4)< (15,4)>
y= 5   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ← 走得出去：(0,5)< (15,5)>
y= 6   .  .  .  .  .  .  .  .  .  .  .  .  .  5  .  .  ← 走得出去：(0,6)< (15,6)>
y= 7   .  .  .  .  .  .  .  d  .  .  e  .  .  .  .  .  
y= 8   .  .  .  .  c  .  .  .  .  e  .  .  j  .  k  .  
y= 9   .  4  .  .  .  .  .  g  g  f  .  .  .  .  .  .  
y=10   .  .  .  .  .  .  .  .  .  .  .  .  j  .  l  .  
y=11   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  
y=12   .  .  .  .  .  .  4  .  .  .  .  .  .  .  .  .  ← 走得出去：(15,12)>
y=13   .  .  .  1  .  .  .  .  .  .  .  .  .  .  .  .  ← 走得出去：(0,13)< (15,13)>
y=14   .  .  .  1  .  .  2  .  .  .  .  .  .  3  .  .  ← 走得出去：(0,14)< (15,14)>
y=15   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ← 走得出去：(0,15)v (0,15)< (1,15)v (2,15)v (3,15)v (11,15)v (12,15)v (13,15)v (14,15)v (15,15)> (15,15)v
```

`.` 沒有事件的地面　`#` 四面都走不出去　`^v<>` 這一格往那個方向走得出地圖　`@` 宣告的進場錨點

## 離開這張圖的格子

⚠ **走得出去 ≠ 已經接上**：`external_exit` 沒宣告的那幾格，remake 會照 `wrap` 繞回對邊，走不到腳本安排的目的地。

| 格子 | 方向 | game pack 宣告 |
|---|---|---|
| `(0,0)` | 北 | **沒宣告** |
| `(0,0)` | 西 | **沒宣告** |
| `(1,0)` | 北 | **沒宣告** |
| `(2,0)` | 北 | **沒宣告** |
| `(3,0)` | 北 | **沒宣告** |
| `(4,0)` | 北 | **沒宣告** |
| `(5,0)` | 北 | **沒宣告** |
| `(6,0)` | 北 | **沒宣告** |
| `(7,0)` | 北 | **沒宣告** |
| `(8,0)` | 北 | **沒宣告** |
| `(9,0)` | 北 | **沒宣告** |
| `(10,0)` | 北 | **沒宣告** |
| `(11,0)` | 北 | **沒宣告** |
| `(12,0)` | 北 | **沒宣告** |
| `(13,0)` | 北 | **沒宣告** |
| `(14,0)` | 北 | **沒宣告** |
| `(15,0)` | 北 | **沒宣告** |
| `(15,0)` | 東 | **沒宣告** |
| `(0,1)` | 西 | **沒宣告** |
| `(15,1)` | 東 | **沒宣告** |
| `(0,2)` | 西 | **沒宣告** |
| `(15,2)` | 東 | **沒宣告** |
| `(0,3)` | 西 | **沒宣告** |
| `(15,3)` | 東 | **沒宣告** |
| `(0,4)` | 西 | **沒宣告** |
| `(15,4)` | 東 | **沒宣告** |
| `(0,5)` | 西 | **沒宣告** |
| `(15,5)` | 東 | **沒宣告** |
| `(0,6)` | 西 | **沒宣告** |
| `(15,6)` | 東 | **沒宣告** |
| `(15,12)` | 東 | **沒宣告** |
| `(0,13)` | 西 | **沒宣告** |
| `(15,13)` | 東 | **沒宣告** |
| `(0,14)` | 西 | **沒宣告** |
| `(15,14)` | 東 | **沒宣告** |
| `(0,15)` | 南 | **沒宣告** |
| `(0,15)` | 西 | **沒宣告** |
| `(1,15)` | 南 | **沒宣告** |
| `(2,15)` | 南 | **沒宣告** |
| `(3,15)` | 南 | **沒宣告** |
| `(11,15)` | 南 | **沒宣告** |
| `(12,15)` | 南 | **沒宣告** |
| `(13,15)` | 南 | **沒宣告** |
| `(14,15)` | 南 | **沒宣告** |
| `(15,15)` | 東 | **沒宣告** |
| `(15,15)` | 南 | **沒宣告** |

## 每格事件

⚠ **有索引不等於站上去就會演**：處理常式自己可能還有守衛（一次性旗標、`RANDOM`、前置劇情、`SEARCH`）。實際演出來什麼看[`cell-sweep`](../../audit/cell-sweep.md)。

| 圖上的字 | 索引 | 格子 | 守衛 | 那一場的第一句 |
|---|---:|---|---|---|
| `1` | 1 | `(3,13)`、`(3,14)` | `COMPARE 4CBE 01 / IF = / EXIT / SAVE 01 4CBE` | AN ELFISH SPIRIT APPEARS AND GREETS YOU. |
| `2` | 2 | `(6,14)` | `COMPARE 4CBF 01 / IF = / EXIT` | A RED WEB STRETCHES ACROSS THE PASSAGE, |
| `3` | 3 | `(13,14)` | `COMPARE 4CC0 01 / IF = / EXIT / SAVE 01 4CC0` | A SPIRIT APPEARS BEFORE YOU. 'I AM THE SPIRIT OF |
| `4` | 4 | `(10,4)`、`(1,9)`、`(6,12)` | `COMPARE 4CC1 04 / IF = / EXIT / COMPARE 7F81 01` | A THRI-KREEN IS EXCAVATING A GRAVE HERE. AT YOUR |
| `5` | 5 | `(13,6)` | `COMPARE 4CC2 01 / IF = / EXIT / SAVE 14 7EE1` | HE MAKES A GESTURE OF FRIENDSHIP |
| `6` | 6 | `(13,4)` | `PRINTCLEAR 00 / PICTURE FF / CALL 2E10 / EXIT` | — |
| `7` | 7 | `(14,3)` | `COMPARE 4C05 01 / IF = / EXIT / SAVE 01 7F82` | NEAR THE ENTRANCE TO THIS BUILDING IS A CRUSHED |
| `8` | 8 | `(5,2)`、`(4,3)`、`(5,3)` | `COMPARE 4CC3 01 / IF = / EXIT` | NEAR THE ENTRANCE TO THIS BUILDING IS A CRUSHED |
| `9` | 9 | `(3,1)` | `COMPARE 4CC4 01 / IF = / EXIT` | TWO SUITS OF ARMOR FLANK THIS STAIRWAY, RADIATING |
| `a` | 10 | `(1,2)` | `COMPARE 4CC5 01 / IF = / EXIT / SAVE 01 4CC5` | AS YOU APPROACH THE STAIRS A VOICE CRIES OUT,' |
| `b` | 11 | `(1,3)` | `COMPARE 4CC6 01 / IF = / EXIT / SAVE 01 4CC6` | A SPIRIT APPEARS BEFORE YOU.' |
| `c` | 12 | `(4,8)` | `COMPARE 4CC7 01 / IF = / EXIT / SAVE 01 4CC7` | A FIGURE APPEARS FROM THE SHADOWS. 'HAIL BONDED ONES!' |
| `d` | 13 | `(7,7)` | `COMPARE 4CB8 0A / IF >= / EXIT / RANDOM 09 7F79` | — |
| `e` | 14 | `(10,7)`、`(9,8)` | `COMPARE 4CC8 01 / IF = / EXIT / SETUP MONSTER 40 02 40` | A PARTY OF THRI-KREEN BAR YOUR ENTRANCE. |
| `f` | 15 | `(9,9)` | `COMPARE 4CC9 01 / IF = / EXIT` | GUARDS HERE PREPARE FOR COMBAT. |
| `g` | 16 | `(7,9)`、`(8,9)` | `COMPARE 4CCA 01 / IF = / EXIT` | THE THRI-KREEN HAVE BIVOUACKED HERE. THEY |
| `h` | 17 | `(9,2)` | `COMPARE 4CCB 01 / IF = / EXIT / SETUP MONSTER 41 00 41` | WEBS FESTOON THIS MAUSOLEUM. THE WEBS ARE INHABITED. |
| `i` | 18 | `(10,1)` | `COMPARE 4CCC 01 / IF = / EXIT` | YOU SEE A FUNNEL OF WEBS. |
| `j` | 19 | `(12,8)`、`(12,10)` | `COMPARE 4CCD 01 / IF = / EXIT / SETUP MONSTER 41 01 41` | AS YOU ENTER, SPIDERS COME OUT OF THE SOLID WALLS. |
| `k` | 20 | `(14,8)` | `COMPARE 4CCE 01 / IF = / EXIT / SETUP MONSTER 41 02 41` | GLOWING SPIDERS SKITTER FORWARD AT YOUR APPROACH. |
| `l` | 21 | `(14,10)` | `COMPARE 4CCF 01 / IF = / EXIT / SETUP MONSTER 41 02 41` | SPIDERS HAVE GATHERED A PILE OF BONES HERE |
| `m` | 22 | `(1,1)`、`(2,1)`、`(4,1)`、`(2,2)`、`(3,2)`、`(4,2)` | `PRINTCLEAR 00 / PICTURE FF / CALL 2E10 / EXIT` | — |
