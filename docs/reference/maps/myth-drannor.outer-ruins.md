# myth-drannor.outer-ruins

由 `cmd/map-atlas` 產生，不要手改。`GEO6` 區塊 `0x42`；腳本 `ECL6/0x42`。

```text
       0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 
y= 0   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ← 走得出去：(2,0)^ (7,0)^ (12,0)^
y= 1   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  
y= 2   .  .  .  .  .  .  .  .  .  b  .  b  .  .  .  .  
y= 3   .  a  .  .  .  .  .  .  .  .  .  .  .  .  6  .  
y= 4   .  a  .  .  .  .  .  .  .  .  .  .  .  .  .  .  
y= 5   .  .  .  4  .  .  .  .  .  .  .  .  .  .  .  .  
y= 6   .  .  4  5  .  c  .  .  .  d  .  .  .  .  .  .  ← 走得出去：(0,6)<
y= 7   .  .  .  4  .  .  .  .  .  .  .  .  .  7  .  .  
y= 8   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  
y= 9   .  .  .  .  .  .  .  .  .  .  .  .  8  .  .  .  
y=10   .  .  .  .  .  .  .  .  .  8  .  .  .  .  .  .  
y=11   .  .  .  .  .  .  .  .  .  .  .  9  .  .  .  .  
y=12   .  1  .  .  .  .  7  .  .  .  .  .  .  .  .  .  ← 走得出去：(0,12)<
y=13   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  
y=14   .  .  3  2  .  .  .  .  .  .  .  .  .  .  .  .  
y=15   .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  
```

`.` 沒有事件的地面　`#` 四面都走不出去　`^v<>` 這一格往那個方向走得出地圖　`@` 宣告的進場錨點

## 離開這張圖的格子

⚠ **走得出去 ≠ 已經接上**：`external_exit` 沒宣告的那幾格，remake 會照 `wrap` 繞回對邊，走不到腳本安排的目的地。

| 格子 | 方向 | game pack 宣告 |
|---|---|---|
| `(2,0)` | 北 | **沒宣告** |
| `(7,0)` | 北 | **沒宣告** |
| `(12,0)` | 北 | **沒宣告** |
| `(0,6)` | 西 | **沒宣告** |
| `(0,12)` | 西 | **沒宣告** |

## 每格事件

⚠ **有索引不等於站上去就會演**：處理常式自己可能還有守衛（一次性旗標、`RANDOM`、前置劇情、`SEARCH`）。實際演出來什麼看[`cell-sweep`](../../audit/cell-sweep.md)。

| 圖上的字 | 索引 | 格子 | 守衛 | 那一場的第一句 |
|---|---:|---|---|---|
| `1` | 1 | `(1,12)` | `COMPARE AND 4CD0 00 4CD1 00 / IF <> / EXIT / SETUP MONSTER 43 00 43` | A RAKSHASA WITH MATTED FUR AND A DOUR EXPRESSION |
| `2` | 2 | `(3,14)` | `COMPARE 4CD1 01 / IF = / EXIT / SETUP MONSTER 45 00 45` | SOME MARGOYLES AND HELL HOUNDS GUARD THE ENTRANCE |
| `3` | 3 | `(2,14)` | `COMPARE 4CD2 01 / IF = / EXIT` | PILED WITHIN THIS BUILDING IS A LARGE ARRAY OF |
| `4` | 4 | `(3,5)`、`(2,6)`、`(3,7)` | `COMPARE 4CD3 01 / IF = / EXIT / SAVE 01 4CD3` | AHEAD, YOU SEE A MAN RUNNING, NEAR EXHUASTION, |
| `5` | 5 | `(3,6)` | `COMPARE 4CD4 01 / IF = / EXIT / SAVE 01 4CD4` | THE HOUNDS LEFT LITTLE THAT IS RECOGNIZABLE AS HUMAN. |
| `6` | 6 | `(14,3)` | `COMPARE AND 7ECA 01 4CD5 01 / IF <> / EXIT / SAVE 00 4CD5` | JUST AS THE DYING MAN DESCRIBED, YOU LOCATE A CACHE. |
| `7` | 7 | `(13,7)`、`(6,12)` | `COMPARE 4C06 01 / IF = / EXIT / SAVE 01 4C06` | NAMELESS SLIDES OUT OF THE SHADOWS, 'THIS DIRECT |
| `8` | 8 | `(12,9)`、`(9,10)` | `COMPARE 4CD6 01 / IF = / EXIT / SAVE 01 4CD6` | IN THE PLAZA AHEAD IS SOME DENSE BRUSH. A SMALL CHILD |
| `9` | 9 | `(11,11)` | `—` | THE BRUSH IS DENSE AND FILLED WITH RUBBLE. A FEW |
| `a` | 10 | `(1,3)`、`(1,4)` | `COMPARE 4CD7 01 / IF = / EXIT / SAVE 01 4CD7` | AS YOU STEP INTO THIS OPULENT ROOM, YOU SEE SEVERAL |
| `b` | 11 | `(9,2)`、`(11,2)` | `COMPARE 4CD8 01 / IF = / EXIT` | TWO MARGOYLES ARE TORTURING A SMALL ANIMAL IN A |
| `c` | 12 | `(5,6)` | `COMPARE 4CD9 01 / IF = / GOTO 95C1 / SAVE 01 4CD9` | A LONE MARGOYLE SKITTERS AWAY UNCOVERING A SEWER |
| `d` | 13 | `(9,6)` | `COMPARE 4CDA 01 / IF = / EXIT / SAVE 01 4CDA` | A RAKSHASA RESIDES HERE IN SPLENDOR |
