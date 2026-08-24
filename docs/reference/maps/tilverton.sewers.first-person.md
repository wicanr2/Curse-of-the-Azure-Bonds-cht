# tilverton.sewers.first-person

由 `cmd/map-atlas` 產生，不要手改。`GEO2` 區塊 `0x03`；腳本 `ECL2/0x03`。

```text
       0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 
y= 0   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ← 走得出去：(0,0)^ (4,0)^ (9,0)^ (11,0)^ (15,0)^
y= 1   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  
y= 2   .  .  .  .  .  .  .  .  .  .  .  7  7  8  8  8  
y= 3   .  .  .  .  .  d  .  .  .  .  g  7  7  8  8  8  
y= 4   .  .  .  .  .  .  .  .  .  .  .  .  .  .  8  8  
y= 5   .  .  .  .  #  2  .  .  .  .  .  .  .  j  .  .  
y= 6   .  d  .  .  .  .  .  .  .  .  .  .  .  9  e  e  
y= 7   .  .  .  .  .  b  .  .  .  .  .  .  .  e  e  e  
y= 8   .  1  .  .  .  .  .  .  .  .  .  .  .  d  e  e  
y= 9   .  b  .  i  .  .  .  .  .  h  .  .  .  .  .  e  
y=10   .  .  .  .  .  .  .  h  5  5  .  c  .  3  .  e  
y=11   .  .  .  .  .  .  .  5  5  5  .  .  .  .  .  e  
y=12   .  .  .  .  .  .  .  .  .  .  .  .  .  .  e  e  
y=13   .  .  .  .  .  .  .  .  .  .  .  .  .  .  e  e  
y=14   .  4  4  .  .  .  .  .  6  6  .  .  .  .  e  e  
y=15   .  .  f  .  .  .  .  .  6  6  .  .  .  .  e  a  ← 走得出去：(2,15)v (5,15)v (9,15)v (10,15)v (13,15)v (15,15)v
```

`.` 沒有事件的地面　`#` 四面都走不出去　`^v<>` 這一格往那個方向走得出地圖　`@` 宣告的進場錨點

## 離開這張圖的格子

⚠ **走得出去 ≠ 已經接上**：`external_exit` 沒宣告的那幾格，remake 會照 `wrap` 繞回對邊，走不到腳本安排的目的地。

| 格子 | 方向 | game pack 宣告 |
|---|---|---|
| `(0,0)` | 北 | `tilverton.sewers.north-00-to-guild` |
| `(4,0)` | 北 | `tilverton.sewers.north-04-to-guild` |
| `(9,0)` | 北 | `tilverton.sewers.north-09-wrap-west` |
| `(11,0)` | 北 | `tilverton.sewers.north-11-wrap-west` |
| `(15,0)` | 北 | `tilverton.sewers.north-15-wrap-west` |
| `(2,15)` | 南 | `tilverton.sewers.south-02-wrap-east` |
| `(5,15)` | 南 | `tilverton.sewers.south-05-wrap-east` |
| `(9,15)` | 南 | `tilverton.sewers.south-09-wrap-east` |
| `(10,15)` | 南 | `tilverton.sewers.south-10-to-hideout` |
| `(13,15)` | 南 | `tilverton.sewers.south-13-to-hideout` |
| `(15,15)` | 南 | `tilverton.sewers.south-15-to-hideout` |

## 每格事件

⚠ **有索引不等於站上去就會演**：處理常式自己可能還有守衛（一次性旗標、`RANDOM`、前置劇情、`SEARCH`）。實際演出來什麼看[`cell-sweep`](../../audit/cell-sweep.md)。

| 圖上的字 | 索引 | 格子 | 守衛 | 那一場的第一句 |
|---|---:|---|---|---|
| `1` | 1 | `(1,8)` | `AND 4C2B 01 7F79 / IF <> / EXIT / ADD 01 4C2B 4C2B` | YOU ARE AT A CHECKPOINT. |
| `2` | 2 | `(5,5)` | `AND 4C2B 01 7F79 / IF <> / EXIT / ADD 01 4C2B 4C2B` | YOU ARE AT A CHECKPOINT. |
| `3` | 3 | `(13,10)` | `AND 4C2B 02 7F79 / IF <> / EXIT` | HERE LIES THE SLAUGHTERED REMAINS OF A FIRE KNIFE |
| `4` | 4 | `(1,14)`、`(2,14)` | `AND 4C2B 08 7F79 / IF <> / EXIT / ADD 08 4C2B 4C2B` | A TERRIBLE STENCH ASSAULTS YOUR SENSES AS YOU ENTER |
| `5` | 5 | `(8,10)`、`(9,10)`、`(7,11)`、`(8,11)`、`(9,11)` | `COMPARE 4C0C 00 / IF <> / EXIT / SAVE 01 4C0C` | PILES OF EXCREMENT HAVE BEEN SHAPED INTO PYRAMIDS HERE. |
| `6` | 6 | `(8,14)`、`(9,14)`、`(8,15)`、`(9,15)` | `AND 4C2B 20 7F79 / IF <> / EXIT / ADD 20 4C2B 4C2B` | THE ROOM IS FILLED WITH FILTH, THOUGH MOST OF |
| `7` | 7 | `(11,2)`、`(12,2)`、`(11,3)`、`(12,3)` | `AND 4C2B 40 7F79 / IF <> / EXIT / ADD 40 4C2B 4C2B` | THE ROOM IS SWAMPY, AND YOU SINK DOWN TO YOUR |
| `8` | 8 | `(13,2)`、`(14,2)`、`(15,2)`、`(13,3)`、`(14,3)`、`(15,3)`、`(14,4)`、`(15,4)` | `AND 4C2B 80 7F79 / IF <> / EXIT / ADD 80 4C2B 4C2B` | AS YOU OPEN THE DOOR, HANDS REACH DOWN FROM ABOVE. |
| `9` | 9 | `(13,6)` | `SAVE 01 7F7B / GOTO 8A5C / SAVE 02 7F7B / AND 4C03 7F7A 7F79` | — |
| `a` | 10 | `(15,15)` | `SAVE 02 7F7B / AND 4C03 7F7A 7F79 / IF <> / EXIT` | — |
| `b` | 11 | `(5,7)`、`(1,9)` | `COMPARE C04D 03 / IF <> / EXIT / COMPARE C04B 01` | — |
| `c` | 12 | `(11,10)` | `SAVE 04 7F7A / AND 01 4C2B 7F79 / IF <> / EXIT` | — |
| `d` | 13 | `(5,3)`、`(1,6)`、`(13,8)` | `COMPARE C04B 01 / IF = / SAVE 01 7F7A / IF =` | — |
| `e` | 14 | `(14,6)`、`(15,6)`、`(13,7)`、`(14,7)`、`(15,7)`、`(14,8)`、`(15,8)`、`(15,9)`、`(15,10)`、`(15,11)`、`(14,12)`、`(15,12)`、`(14,13)`、`(15,13)`、`(14,14)`、`(15,14)`、`(14,15)` | `COMPARE 4C1F 01 / IF >= / GOTO 8B3E / SAVE 01 4C1F` | YOU ENTER THE HIDDEN CHAMBERS. |
| `f` | 15 | `(2,15)` | `AND 4C03 7F7A 7F79 / IF <> / EXIT / OR 4C03 7F7A 4C03` | YOU SEE A SCRAP OF PURPLE CLOTH CLINGING TO THE BOTTOM |
| `g` | 16 | `(10,3)` | `COMPARE 4C1E 01 / IF = / EXIT / SAVE 01 4C1E` | BURNT INTO THE WALL HERE IS THE SYMBOL OF A HAND WITH |
| `h` | 17 | `(9,9)`、`(7,10)` | `SAVE 00 4C0C / EXIT / COMPARE 4C0C 00 / IF <>` | PILES OF EXCREMENT HAVE BEEN SHAPED INTO PYRAMIDS HERE. |
| `i` | 18 | `(3,9)` | `COMPARE 4C10 01 / IF = / EXIT` | YOU SPOT SOMETHING FLAPPING ON THE CEILING. TO |
| `j` | 19 | `(13,5)` | `AND 04 4C2B 7F79 / IF <> / EXIT / ADD 04 4C2B 4C2B` | YOU HEAR A SOUND, SUDDENLY CUT OFF, TO THE SOUTH |
