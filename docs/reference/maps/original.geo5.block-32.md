# original.geo5.block-32

由 `cmd/map-atlas` 產生，不要手改。`GEO5` 區塊 `0x32`；腳本 `ECL5/0x32`。

```text
       0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 
y= 0   9  9  1  1  4  9  .  .  .  .  .  .  .  .  .  .  
y= 1   9  4  .  .  .  9  .  .  .  .  .  .  .  .  .  .  
y= 2   3  .  .  .  a  a  .  9  .  .  .  .  .  .  7  .  
y= 3   6  9  .  a  a  a  .  .  .  .  .  .  6  .  .  .  
y= 4   9  9  .  a  a  a  .  .  8  .  .  .  .  .  .  .  
y= 5   9  6  .  .  4  9  .  .  .  .  .  .  .  2  1  .  ← 走得出去：(15,5)>
y= 6   3  .  .  .  9  9  .  .  .  .  .  .  .  3  .  .  
y= 7   b  b  b  .  .  2  .  .  .  .  .  .  .  3  4  .  
y= 8   b  b  b  .  .  4  .  .  .  .  .  .  .  3  .  5  
y= 9   b  b  b  .  9  9  .  .  .  .  .  .  .  .  .  .  
y=10   3  .  .  .  4  9  .  .  .  a  .  .  .  .  .  .  
y=11   9  4  .  .  9  9  .  .  .  .  .  .  b  .  .  g  
y=12   9  9  .  .  .  2  .  c  .  .  .  .  .  .  .  .  
y=13   .  .  .  8  .  .  .  .  .  .  .  .  .  .  .  e  ← 走得出去：(0,13)<
y=14   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ← 走得出去：(0,14)<
y=15   .  .  .  .  .  .  f  .  .  .  .  .  .  d  .  .  ← 走得出去：(0,15)v (0,15)<
```

`.` 沒有事件的地面　`#` 四面都走不出去　`^v<>` 這一格往那個方向走得出地圖　`@` 宣告的進場錨點

## 離開這張圖的格子

⚠ **走得出去 ≠ 已經接上**：`external_exit` 沒宣告的那幾格，remake 會照 `wrap` 繞回對邊，走不到腳本安排的目的地。

| 格子 | 方向 | game pack 宣告 |
|---|---|---|
| `(15,5)` | 東 | **沒宣告** |
| `(0,13)` | 西 | **沒宣告** |
| `(0,14)` | 西 | **沒宣告** |
| `(0,15)` | 南 | **沒宣告** |
| `(0,15)` | 西 | **沒宣告** |

## 每格事件

⚠ **有索引不等於站上去就會演**：處理常式自己可能還有守衛（一次性旗標、`RANDOM`、前置劇情、`SEARCH`）。實際演出來什麼看[`cell-sweep`](../../audit/cell-sweep.md)。

| 圖上的字 | 索引 | 格子 | 守衛 | 那一場的第一句 |
|---|---:|---|---|---|
| `1` | 1 | `(2,0)`、`(3,0)`、`(14,5)` | `COMPARE AND 7ECA 01 4C04 00 / IF <> / EXIT` | YOU FIND AN ARROW POINTING WEST FAINTLY SCRATCHED |
| `2` | 2 | `(13,5)`、`(5,7)`、`(5,12)` | `COMPARE AND 7ECA 01 4C05 00 / IF <> / EXIT` | AN ARROW MADE OF SMALL STONES POINTS SOUTH, HERE. |
| `3` | 3 | `(0,2)`、`(0,6)`、`(13,6)`、`(13,7)`、`(13,8)`、`(0,10)` | `GOSUB 881B / IF <> / EXIT / COMPARE C04D 00` | YOU HEAR A PATROL APPROACH FROM BEHIND YOU. |
| `4` | 4 | `(4,0)`、`(1,1)`、`(4,5)`、`(14,7)`、`(5,8)`、`(4,10)`、`(1,11)` | `COMPARE 4C06 01 / IF = / EXIT / COMPARE AND 4C60 00 4C61 00` | FOUR FEMALE DARK ELVES STEP FROM THE SHADOWS. |
| `5` | 5 | `(15,8)` | `COMPARE AND 4C60 00 4C61 00 / IF <> / EXIT / COMPARE 4C08 01` | A DARK ELFIN WOMAN STEPS FORWARD. HER HAIR IS DARK |
| `6` | 6 | `(0,3)`、`(12,3)`、`(1,5)` | `COMPARE AND 4C0E 00 4C60 00 / IF <> / EXIT / SAVE 01 4C0E` | SOME DARK ELVES ARE HERE, ATOP A MOUND OF FRESHLY |
| `7` | 7 | `(14,2)` | `COMPARE 4C60 00 / IF <> / EXIT / AND 4C48 02 7F79` | YOU HAVE DISTURBED A BARRACKS FULL OF DARK ELVES |
| `8` | 8 | `(8,4)`、`(3,13)` | `COMPARE 4C60 00 / IF <> / EXIT / AND 4C48 04 7F79` | THIS ROOM IS FILLED WITH CLOYING INCENSE SMOKE. |
| `9` | 9 | `(0,0)`、`(1,0)`、`(5,0)`、`(0,1)`、`(5,1)`、`(7,2)`、`(1,3)`、`(0,4)`、`(1,4)`、`(0,5)`、`(5,5)`、`(4,6)`、`(5,6)`、`(4,9)`、`(5,9)`、`(5,10)`、`(0,11)`、`(4,11)`、`(5,11)`、`(0,12)`、`(1,12)` | `CALL 2E10` | — |
| `a` | 10 | `(4,2)`、`(5,2)`、`(3,3)`、`(4,3)`、`(5,3)`、`(3,4)`、`(4,4)`、`(5,4)`、`(9,10)` | `COMPARE 4C60 00 / IF <> / EXIT / AND 4C48 08 7F79` | THE DOOR IS GUARDED BY A SALAMANDER LED PATROL. |
| `b` | 11 | `(0,7)`、`(1,7)`、`(2,7)`、`(0,8)`、`(1,8)`、`(2,8)`、`(0,9)`、`(1,9)`、`(2,9)`、`(12,11)` | `COMPARE 4C60 00 / IF <> / EXIT / AND 4C48 10 7F79` | MYSTIC SYMBOLS ADORNE THE WALLS. MAGES ARE HERE |
| `c` | 12 | `(7,12)` | `COMPARE AND 4C62 00 4C60 00 / IF <> / EXIT / SETUP MONSTER 3C 00 3C` | CURLED IN THE CENTER OF THIS ROOM IS THE HUGE SKELETAL |
| `d` | 13 | `(13,15)` | `COMPARE 4C60 00 / IF <> / EXIT / AND 4C48 20 7F79` | THIS WAY IS GUARDED BY EFREETI AND DARK ELVES. |
| `e` | 14 | `(15,13)` | `COMPARE 4C66 01 / IF = / EXIT` | AT YOUR APPROACH, THE ELVES COLLAPSE THE TUNNEL. |
| `f` | 15 | `(6,15)` | `COMPARE AND 4C60 01 4C61 01 / IF <> / EXIT / COMPARE AND 4C64 00 4C65 00` | SILK STEPS OUT FROM THE SHADOWS.' CONGRATULATIONS |
| `g` | 16 | `(15,11)` | `CALL 2E10` | — |
