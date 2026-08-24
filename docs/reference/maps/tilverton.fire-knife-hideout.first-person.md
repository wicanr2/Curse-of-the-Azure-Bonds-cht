# tilverton.fire-knife-hideout.first-person

由 `cmd/map-atlas` 產生，不要手改。`GEO2` 區塊 `0x04`；腳本 `ECL2/0x04`。

```text
       0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 
y= 0   .  .  .  .  .  .  .  .  1  .  .  2  .  .  .  .  ← 走得出去：(8,0)^ (11,0)^ (13,0)^
y= 1   .  .  .  q  .  p  .  .  .  .  .  .  .  .  .  .  
y= 2   .  .  q  .  q  p  .  p  .  .  .  .  .  .  .  9  
y= 3   .  .  .  q  .  .  .  .  .  .  .  .  .  .  .  l  
y= 4   .  .  .  .  .  .  .  .  .  .  .  .  .  .  a  l  
y= 5   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  l  
y= 6   .  w  w  .  .  .  .  j  .  h  .  .  .  .  .  l  
y= 7   .  .  .  .  .  k  l  l  l  l  .  f  .  .  b  l  
y= 8   w  .  .  .  .  .  .  .  i  l  l  l  l  l  l  l  
y= 9   o  o  5  .  .  .  .  .  .  .  .  .  e  l  .  .  
y=10   o  o  .  .  .  n  n  n  .  .  .  .  .  l  .  .  
y=11   o  o  .  .  .  n  n  n  .  .  .  .  .  d  r  .  
y=12   .  .  .  .  4  .  6  .  .  .  .  .  .  .  .  s  
y=13   .  .  .  7  .  .  8  m  m  m  .  .  .  .  u  .  
y=14   .  .  .  .  .  .  .  m  m  m  .  .  v  .  .  .  
y=15   .  .  .  .  .  .  8  m  m  m  .  .  .  .  .  t  
```

`.` 沒有事件的地面　`#` 四面都走不出去　`^v<>` 這一格往那個方向走得出地圖　`@` 宣告的進場錨點

## 離開這張圖的格子

⚠ **走得出去 ≠ 已經接上**：`external_exit` 沒宣告的那幾格，remake 會照 `wrap` 繞回對邊，走不到腳本安排的目的地。

| 格子 | 方向 | game pack 宣告 |
|---|---|---|
| `(8,0)` | 北 | `tilverton.fire-knife-hideout.e1-north-west` |
| `(11,0)` | 北 | `tilverton.fire-knife-hideout.e1-north-centre` |
| `(13,0)` | 北 | `tilverton.fire-knife-hideout.e1-north-east` |

## 每格事件

⚠ **有索引不等於站上去就會演**：處理常式自己可能還有守衛（一次性旗標、`RANDOM`、前置劇情、`SEARCH`）。實際演出來什麼看[`cell-sweep`](../../audit/cell-sweep.md)。

| 圖上的字 | 索引 | 格子 | 守衛 | 那一場的第一句 |
|---|---:|---|---|---|
| `1` | 1 | `(8,0)` | `AND 4CFE 01 7F79 / IF <> / EXIT / ADD 01 4CFE 4CFE` | YOU SEE THE REMAINS OF A FIRE KNIFE CHECKPOINT. |
| `2` | 2 | `(11,0)` | `AND 4CFE 02 7F79 / IF <> / EXIT / ADD 02 4CFE 4CFE` | THERE ARE SIGNS THAT THIS IS NORMALLY A |
| `3` | 3 | 圖上沒有這個地形碼 | `GOSUB 961F / SAVE 01 7EE1 / SETUP MONSTER 01 00 01 / SAVE FF 7EE1` | YOU ARE AT A CHECKPOINT. |
| `4` | 4 | `(4,12)` | `GOSUB 961F / SAVE 01 7EE1 / SETUP MONSTER 01 00 01 / SAVE FF 7EE1` | YOU ARE AT A CHECKPOINT. |
| `5` | 5 | `(2,9)` | `EXIT / SAVE 08 7F7A / AND 4C00 7F7A 7F79 / IF <>` | YOU SPOT A CHECKPOINT TO THE |
| `6` | 6 | `(6,12)` | `SAVE 08 7F7A / AND 4C00 7F7A 7F79 / IF <> / EXIT` | YOU SPOT A CHECKPOINT TO THE |
| `7` | 7 | `(3,13)` | `PICTURE 0C` | YOU MEET THE LEADER OF THE FIRE KNIVES. |
| `8` | 8 | `(6,13)`、`(6,15)` | `AND 4CFE 04 7F79 / IF <> / EXIT / ADD 04 4CFE 4CFE` | YOU FOUND THE ARMORY. |
| `9` | 9 | `(15,2)` | `SAVE 02 7F7B / GOTO 86DC / COMPARE 7ECA 01 / IF =` | — |
| `a` | 10 | `(14,4)` | `SAVE 01 7F7B / GOTO 86DC / SAVE 02 7F7B / GOTO 86DC` | — |
| `b` | 11 | `(14,7)` | `SAVE 01 7F7B / GOTO 86DC / SAVE 02 7F7B / GOTO 86DC` | — |
| `c` | 12 | 圖上沒有這個地形碼 | `SAVE 00 7F7B / GOTO 86DC / SAVE 01 7F7B / GOTO 86DC` | — |
| `d` | 13 | `(13,11)` | `SAVE 00 7F7B / GOTO 86DC / SAVE 01 7F7B / GOTO 86DC` | — |
| `e` | 14 | `(12,9)` | `SAVE 01 7F7B / GOTO 86DC / SAVE 02 7F7B / GOTO 86DC` | — |
| `f` | 15 | `(11,7)` | `SAVE 02 7F7B / GOTO 86DC / COMPARE 7ECA 01 / IF =` | — |
| `g` | 16 | 圖上沒有這個地形碼 | `SAVE 00 7F7B / GOTO 86DC / SAVE 01 7F7B / GOTO 86DC` | — |
| `h` | 17 | `(9,6)` | `SAVE 02 7F7B / GOTO 86DC / COMPARE 7ECA 01 / IF =` | — |
| `i` | 18 | `(8,8)` | `SAVE 00 7F7B / GOTO 86DC / SAVE 01 7F7B / GOTO 86DC` | — |
| `j` | 19 | `(7,6)` | `SAVE 02 7F7B / GOTO 86DC / COMPARE 7ECA 01 / IF =` | — |
| `k` | 20 | `(5,7)` | `SAVE 01 7F7B / GOTO 86DC / SAVE 02 7F7B / GOTO 86DC` | — |
| `l` | 21 | `(15,3)`、`(15,4)`、`(15,5)`、`(15,6)`、`(6,7)`、`(7,7)`、`(8,7)`、`(9,7)`、`(15,7)`、`(9,8)`、`(10,8)`、`(11,8)`、`(12,8)`、`(13,8)`、`(14,8)`、`(15,8)`、`(13,9)`、`(13,10)` | `SPRITE OFF / CALL 2E10 / EXIT / SAVE 4BF0 C04B` | — |
| `m` | 22 | `(7,13)`、`(8,13)`、`(9,13)`、`(7,14)`、`(8,14)`、`(9,14)`、`(7,15)`、`(8,15)`、`(9,15)` | `AND 4CFE 08 7F79 / IF <> / EXIT / ADD 08 4CFE 4CFE` | THE ROOM HAS BEEN CONVERTED TO A HOSPITAL. |
| `n` | 23 | `(5,10)`、`(6,10)`、`(7,10)`、`(5,11)`、`(6,11)`、`(7,11)` | `COMPARE 4C0C 00 / IF <> / GOTO 8905 / SAVE 01 4C0C` | THE ROOM SEEMS TO BE USED AS A STORAGE AREA. |
| `o` | 24 | `(0,9)`、`(1,9)`、`(0,10)`、`(1,10)`、`(0,11)`、`(1,11)` | `AND 4CFE 10 7F79 / IF <> / EXIT / SAVE 01 7EE1` | THIS DARK AND SMOKY ROOM IS ADORNED WITH ALL |
| `p` | 25 | `(5,1)`、`(5,2)`、`(7,2)` | `AND 4CFE 20 7F79 / IF <> / EXIT / ADD 20 4CFE 4CFE` | YOU STOP AT THE ENTRANCE TO THIS ROOM. IN FRONT |
| `q` | 26 | `(3,1)`、`(2,2)`、`(4,2)`、`(3,3)` | `AND 4CFE 40 7F79 / IF <> / EXIT / ADD 40 4CFE 4CFE` | ABOUT THE ROOM ARE A NUMBER OF PEOPLE FROZEN IN |
| `r` | 27 | `(14,11)` | `COMPARE 4C10 00 / IF <> / GOTO 8F72` | THIS IS AN ORNATE ROOM, APPARENTLY THE OFFICE OF |
| `s` | 28 | `(15,12)` | `COMPARE 4C11 01 / IF = / EXIT / SAVE 01 4C11` | AS YOU ENTER THIS HALLWAY, YOU DETECT A STRANGE |
| `t` | 29 | `(15,15)` | `COMPARE 4C12 01 / IF = / EXIT / SAVE 01 4C12` | THIS IS AN EXTREMELY WELL ORDERED BEDROOM, |
| `u` | 30 | `(14,13)` | `COMPARE 4C13 01 / IF = / EXIT / SAVE 01 4C13` | THIS ROOM WAS ONCE A LIBRARY, BUT THE SHELVES |
| `v` | 31 | `(12,14)` | `COMPARE 4C14 01 / IF = / EXIT / SAVE 01 4C14` | THIS WAS ONCE A LAB, BUT THE SAME INTENSE FLAME |
| `w` | 32 | `(1,6)`、`(2,6)`、`(0,8)` | `COMPARE 4C15 01 / IF = / EXIT / SAVE 01 4C15` | WITHIN THE ROOM ARE TWO ROWS OF SHROUDED BODIES. |
