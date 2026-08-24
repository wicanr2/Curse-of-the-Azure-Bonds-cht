# original.geo3.block-10

由 `cmd/map-atlas` 產生，不要手改。`GEO3` 區塊 `0x10`；腳本 `ECL3/0x10`。

```text
       0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 
y= 0   .  .  .  .  .  +  +  .  .  z  .  +  .  .  .  .  ← 走得出去：(0,0)^ (1,0)^ (11,0)^
y= 1   .  .  .  o  .  +  +  .  .  .  .  a  .  .  .  .  
y= 2   .  q  .  .  .  p  +  .  t  .  .  9  .  .  .  .  
y= 3   @  q  .  .  .  +  +  i  .  .  .  v  .  .  g  .  
y= 4   .  .  .  .  +  +  .  .  .  .  .  .  .  .  .  .  
y= 5   .  .  .  r  +  +  .  .  .  .  .  .  n  .  .  .  
y= 6   .  1  .  4  .  2  .  .  .  .  .  .  n  .  h  .  ← 走得出去：(0,6)<
y= 7   .  .  .  .  .  .  .  .  .  .  .  b  .  .  .  h  
y= 8   .  e  .  3  .  .  .  .  .  .  .  b  .  .  .  .  
y= 9   .  .  .  .  .  .  .  k  #  .  .  .  .  .  .  .  
y=10   w  .  .  .  .  .  k  .  .  x  .  .  .  .  .  .  
y=11   .  .  .  .  l  .  .  .  .  .  .  +  .  .  .  m  
y=12   .  .  .  .  .  c  +  .  .  .  .  .  .  .  .  .  
y=13   j  .  .  .  .  .  .  .  .  .  f  .  .  6  .  .  
y=14   j  .  .  .  .  .  .  .  y  .  f  #  f  .  .  .  
y=15   .  .  .  .  .  .  d  .  u  .  .  5  .  8  .  7  
```

`.` 沒有事件的地面　`#` 四面都走不出去　`^v<>` 這一格往那個方向走得出地圖　`@` 宣告的進場錨點

## 離開這張圖的格子

⚠ **走得出去 ≠ 已經接上**：`external_exit` 沒宣告的那幾格，remake 會照 `wrap` 繞回對邊，走不到腳本安排的目的地。

| 格子 | 方向 | game pack 宣告 |
|---|---|---|
| `(0,0)` | 北 | **沒宣告** |
| `(1,0)` | 北 | **沒宣告** |
| `(11,0)` | 北 | `yulash.north-to-moander-pit` |
| `(0,6)` | 西 | **沒宣告** |

## 每格事件

⚠ **有索引不等於站上去就會演**：處理常式自己可能還有守衛（一次性旗標、`RANDOM`、前置劇情、`SEARCH`）。實際演出來什麼看[`cell-sweep`](../../audit/cell-sweep.md)。

| 圖上的字 | 索引 | 格子 | 守衛 | 那一場的第一句 |
|---|---:|---|---|---|
| `1` | 1 | `(1,6)` | `COMPARE C04D 01 / IF <> / EXIT / GOSUB 9D3C` | EAST. |
| `2` | 2 | `(5,6)` | `COMPARE C04D 03 / IF <> / EXIT / GOSUB 9D3C` | WEST |
| `3` | 3 | `(3,8)` | `COMPARE C04D 00 / IF <> / EXIT / GOSUB 9D3C` | NORTH |
| `4` | 4 | `(3,6)` | `COMPARE 4C34 FF / IF = / GOTO 8619 / SAVE 02 4C31` | — |
| `5` | 5 | `(11,15)` | `COMPARE C04D 01 / IF <> / EXIT / GOSUB 9D3C` | EAST |
| `6` | 6 | `(13,13)` | `COMPARE C04D 02 / IF <> / EXIT / GOSUB 9D3C` | SOUTH |
| `7` | 7 | `(15,15)` | `COMPARE C04D 03 / IF <> / EXIT / GOSUB 9D3C` | WEST |
| `8` | 8 | `(13,15)` | `COMPARE 4C33 FF / IF = / GOTO 8619 / SAVE 01 4C31` | — |
| `9` | 9 | `(11,2)` | `COMPARE C04D 00 / IF <> / EXIT / GOSUB 9D3C` | NORTH |
| `a` | 10 | `(11,1)` | `COMPARE 4C32 FF / IF = / GOTO 8619 / SAVE 00 4C31` | — |
| `b` | 11 | `(11,7)`、`(11,8)` | `COMPARE 4C46 01 / IF = / EXIT / SAVE 01 4C46` | SHAMBLING MOUNDS ATTEMPT TO DRAG A CLERICS BODY AWAY. |
| `c` | 12 | `(5,12)` | `—` | YOU'VE COME UPON A DESTROYED CHECKPOINT.  THE MARK OF ZH… |
| `d` | 13 | `(6,15)` | `COMPARE 4C50 01 / IF = / EXIT / SAVE 01 4C50` | — |
| `e` | 14 | `(1,8)` | `COMPARE 4C51 01 / IF = / EXIT / SAVE 01 4C51` | — |
| `f` | 15 | `(10,13)`、`(10,14)`、`(12,14)` | `COMPARE 4C52 01 / IF = / EXIT / SAVE 01 4C52` | — |
| `g` | 16 | `(14,3)` | `COMPARE 4C53 01 / IF = / EXIT / SAVE 01 4C53` | — |
| `h` | 17 | `(14,6)`、`(15,7)` | `AND 4C54 01 7F79 / IF <> / EXIT / ADD 01 4C54 4C54` | — |
| `i` | 18 | `(7,3)` | `AND 4C54 02 7F79 / IF <> / EXIT / ADD 02 4C54 4C54` | — |
| `j` | 19 | `(0,13)`、`(0,14)` | `AND 4C54 04 7F79 / IF <> / EXIT / ADD 04 4C54 4C54` | — |
| `k` | 20 | `(7,9)`、`(6,10)` | `AND 4C54 08 7F79 / IF <> / EXIT / ADD 08 4C54 4C54` | — |
| `l` | 21 | `(4,11)` | `AND 4C54 10 7F79 / IF <> / EXIT / ADD 10 4C54 4C54` | — |
| `m` | 22 | `(15,11)` | `AND 4C54 20 7F79 / IF <> / EXIT / ADD 20 4C54 4C54` | — |
| `n` | 23 | `(12,5)`、`(12,6)` | `AND 4C54 40 7F79 / IF <> / EXIT / ADD 40 4C54 4C54` | A FILTHY GROUP HAS BEEN PICKING THROUGH THE RUBBLE.  THE… |
| `o` | 24 | `(3,1)` | `COMPARE 4C42 FF / IF <> / GOTO 9421 / COMPARE 4C4F 00` | THIS IS THE RED GUARDS MESS HALL. |
| `p` | 25 | `(5,2)` | `COMPARE 4C42 FF / IF = / GOTO 8852` | THIS IS THE BARRACKS.  THE ROOM IS ABOUT HALF |
| `q` | 26 | `(1,2)`、`(1,3)` | `COMPARE 4C41 00 / IF = / GOTO 8363 / COMPARE 4C41 FF` | TROOPS COME BURSTING OUT OF THE COMMANDER'S |
| `r` | 27 | `(3,5)` | `COMPARE 4BF0 04 / IF <> / GOTO 818C / COMPARE 4BF1 05` | YOU ARE PICKED UP AND MARCHED TO |
| `s` | 28 | 圖上沒有這個地形碼 | `COMPARE 4C36 01 / IF = / GOTO 818C / SAVE 01 4C36` | — |
| `t` | 29 | `(8,2)` | `COMPARE 4C37 01 / IF = / GOTO 818C / SAVE 01 4C37` | — |
| `u` | 30 | `(8,15)` | `COMPARE 4C38 01 / IF = / GOTO 818C / SAVE 01 4C38` | — |
| `v` | 31 | `(11,3)` | `COMPARE 4C4E 01 / IF = / GOTO 818C / SAVE 01 4C4E` | NOTICES A POSSIBLE SINKHOLE |
| `w` | 32 | `(0,10)` | `COMPARE 4C3A 01 / IF = / GOTO 818C / SAVE 01 4C3A` | — |
| `x` | 33 | `(9,10)` | `COMPARE 4C3B 01 / IF = / GOTO 818C / SAVE 01 4C3B` | — |
| `y` | 34 | `(8,14)` | `COMPARE 4C3C 01 / IF = / GOTO 818C / SAVE 01 4C3C` | — |
| `z` | 35 | `(9,0)` | `COMPARE 4C3D 01 / IF = / GOTO 818C / SAVE 01 4C3D` | — |
| `+` | 36 | `(11,11)` | `COMPARE 4C3E 01 / IF = / GOTO 818C / SAVE 01 4C3E` | NOTICES THAT THE WALLS ARE VERY |
| `+` | 37 | `(6,12)` | `COMPARE 4C43 01 / IF = / EXIT / SAVE 01 4C43` | A RED PLUME GUARD GROWLS, 'NOBODY'S GOING TO MAKE US GO … |
| `+` | 38 | `(11,0)` | `PICTURE 10` | YOU SEE BEFORE THE PIT CREATED BY MOANDER IN HIS |
| `+` | 39 | `(5,0)`、`(6,0)`、`(5,1)`、`(6,1)`、`(6,2)`、`(5,3)`、`(6,3)`、`(4,4)`、`(5,4)`、`(4,5)`、`(5,5)` | `PICTURE FF` | — |
