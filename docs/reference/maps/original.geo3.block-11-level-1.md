# original.geo3.block-11-level-1

由 `cmd/map-atlas` 產生，不要手改。`GEO3` 區塊 `0x11`；腳本 `ECL3/0x11`。

```text
       0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 
y= 0   .  .  .  1  .  .  .  .  .  .  .  .  g  .  .  .  
y= 1   .  .  .  .  .  #  c  c  .  a  .  .  .  .  a  .  
y= 2   .  .  3  .  .  #  .  .  .  .  #  .  .  #  .  .  
y= 3   .  .  .  .  .  .  .  .  .  a  .  c  c  .  a  .  
y= 4   .  5  4  .  .  .  7  i  .  a  .  .  f  .  a  .  
y= 5   .  .  .  2  h  7  7  .  .  .  #  .  .  #  .  .  
y= 6   .  .  8  #  .  .  .  .  .  .  9  9  b  .  .  .  
y= 7   .  .  .  8  .  .  .  .  .  .  .  .  .  .  .  .  
y= 8   .  .  .  .  .  #  b  .  .  .  #  7  .  #  4  .  
y= 9   .  .  .  9  g  .  .  .  .  .  7  7  .  4  4  .  
y=10   .  .  9  .  #  a  .  .  .  .  .  .  e  .  .  .  
y=11   .  .  .  .  a  a  a  j  3  .  #  8  .  #  .  6  ← 走得出去：(0,11)<
y=12   f  .  .  .  .  .  .  .  3  3  .  .  .  .  .  6  
y=13   d  .  d  .  #  .  .  .  .  .  .  5  .  .  .  .  
y=14   .  #  d  .  .  .  e  6  .  .  #  .  .  #  2  1  
y=15   .  .  .  .  .  .  .  #  .  5  .  5  d  .  h  .  
```

`.` 沒有事件的地面　`#` 四面都走不出去　`^v<>` 這一格往那個方向走得出地圖　`@` 宣告的進場錨點

## 離開這張圖的格子

⚠ **走得出去 ≠ 已經接上**：`external_exit` 沒宣告的那幾格，remake 會照 `wrap` 繞回對邊，走不到腳本安排的目的地。

| 格子 | 方向 | game pack 宣告 |
|---|---|---|
| `(0,11)` | 西 | **沒宣告** |

## 每格事件

⚠ **有索引不等於站上去就會演**：處理常式自己可能還有守衛（一次性旗標、`RANDOM`、前置劇情、`SEARCH`）。實際演出來什麼看[`cell-sweep`](../../audit/cell-sweep.md)。

| 圖上的字 | 索引 | 格子 | 守衛 | 那一場的第一句 |
|---|---:|---|---|---|
| `1` | 1 | `(3,0)`、`(15,14)` | `AND 4C4B 01 7F79 / IF <> / EXIT / ADD 4C4B 01 4C4B` | THE BODY OF A DEAD CULTIST LIES IN A |
| `2` | 2 | `(3,5)`、`(14,14)` | `AND 4C4B 02 7F79 / IF <> / EXIT / ADD 4C4B 02 4C4B` | A BLEEDING CLERIC CRAWLS OUT OF THE DOOR |
| `3` | 3 | `(2,2)`、`(8,11)`、`(8,12)`、`(9,12)` | `AND 4C4B 04 7F79 / IF <> / EXIT / ADD 4C4B 04 4C4B` | GREEN ICHOR COVERS THE FLOOR AND WALLS. |
| `4` | 4 | `(2,4)`、`(14,8)`、`(13,9)`、`(14,9)` | `AND 4C4B 08 7F79 / IF <> / EXIT / ADD 4C4B 08 4C4B` | A PILE OF DEAD CLERICS, SHAMBLING MOUNDS AND |
| `5` | 5 | `(1,4)`、`(11,13)`、`(9,15)`、`(11,15)` | `COMPARE 4C2E 00 / IF <> / EXIT / COMPARE 4C5B 00` | YOU SEE A FEMALE FIGHTER AND A STRANGE-LOOKING |
| `6` | 6 | `(15,11)`、`(15,12)`、`(7,14)` | `—` | YOU SEE STAIRS LEADING DOWN TO THE SOUTH. |
| `7` | 7 | `(6,4)`、`(5,5)`、`(6,5)`、`(11,8)`、`(10,9)`、`(11,9)` | `COMPARE 4C2D 80 / IF = / GOTO 85A3 / COMPARE 4C00 01` | — |
| `8` | 8 | `(2,6)`、`(3,7)`、`(11,11)` | `COMPARE 4C2D 80 / IF = / GOTO 85A3 / COMPARE 4C01 01` | — |
| `9` | 9 | `(10,6)`、`(11,6)`、`(3,9)`、`(2,10)` | `COMPARE 4C2D 80 / IF = / GOTO 85A3 / COMPARE 4C02 01` | — |
| `a` | 10 | `(9,1)`、`(14,1)`、`(9,3)`、`(14,3)`、`(9,4)`、`(14,4)`、`(5,10)`、`(4,11)`、`(5,11)`、`(6,11)` | `COMPARE 4C2D 80 / IF = / GOTO 85A3 / COMPARE 4C05 01` | — |
| `b` | 11 | `(12,6)`、`(6,8)` | `COMPARE 4C2D 80 / IF = / GOTO 85A3 / COMPARE 4C06 01` | — |
| `c` | 12 | `(6,1)`、`(7,1)`、`(11,3)`、`(12,3)` | `COMPARE 4C2D 80 / IF = / GOTO 85A3 / COMPARE 4C07 01` | — |
| `d` | 13 | `(0,13)`、`(2,13)`、`(2,14)`、`(12,15)` | `COMPARE 4C2D 80 / IF = / GOTO 85A3 / COMPARE 4C03 01` | — |
| `e` | 14 | `(12,10)`、`(6,14)` | `COMPARE 4C2D 80 / IF = / GOTO 85A3 / COMPARE 4C04 01` | GIANT SLUGS ARE CROSSING YOUR PATH. |
| `f` | 15 | `(12,4)`、`(0,12)` | `COMPARE 4C5B 00 / IF = / EXIT / COMPARE 4C2D FF` | YOU ARE ATTACKED BY A LARGE FORCE OF |
| `g` | 16 | `(12,0)`、`(4,9)` | `GOSUB 97A6 / COMPARE 4C2E FF / IF = / EXIT` | — |
| `h` | 17 | `(4,5)`、`(14,15)` | `COMPARE 4C2D 01 / IF = / EXIT / COMPARE 4C5B 00` | — |
| `i` | 18 | `(7,4)` | `COMPARE 4C2D 01 / IF = / EXIT / COMPARE 4C5B 00` | — |
| `j` | 19 | `(7,11)` | `GOSUB 97A6 / COMPARE 4C2E FF / IF = / EXIT` | — |
