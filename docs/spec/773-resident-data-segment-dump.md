# 773 — 倒出常駐資料段，四張表一次解開

- 工具：`tools/ida/dump_data_segments.py`、`scripts/dseg_tables.py`
- 輸出：`workplace/re-sweep/<平台>/dseg/`、`docs/audit/resident-data-tables.md`

## 為什麼要做

先前的判讀裡累積了一整串「位址知道、內容不知道」的欄位：法術名稱表
（`DS:27BDh`）、金錢欄位名稱（`DS:0F93h`）、方向 dx／dy 表（`DS:2694h` /
`DS:269Dh`）、雲霧形狀索引（`DS:27D8h`）、法術屬性表（`DS:37DAh`）。它們全都
住在**常駐執行檔的資料段**，而 `scripts/scan_pascal_strings.py` 只掃 overlay 的
`.bin`，一條都掃不到。

`dump_data_segments.py` 用 IDAPython 把 `START.EXE` 與 `PC98-GAME.EXE` 的
`DATA` / `BSS` / `CONST` 段原封不動寫成檔案，同時輸出段落清單（名稱、class、
起訖、selector）。**`DS:xxxx` 直接就是 dump 檔的位移**——DOS 的 `dseg` 在
flat `1C1C0h`..`250C0h`（36608 bytes），段基底對齊在段首。

| 平台 | 段 | 大小 |
|---|---|---|
| DOS `START.EXE` | `dseg` | 36608 bytes |
| PC-98 `PC98-GAME.EXE` | `dseg` | 49360 bytes |

## 解開的四張表（完整內容在 `docs/audit/resident-data-tables.md`）

### `DS:0F93h` 金錢欄位名稱（7 筆，每筆 `0Bh` bytes，0 起算）

`Copper`、`Silver`、`Electrum`、`Gold`、`Platinum`、`Gems`、`Jewelry`。

這一組把好幾支函式接了起來：spec 772 的七欄金錢顯示、spec 764 的
`overlay-21:0A03h`（從角色 `+0FBh` 的七個 word 扣掉、累加到全域）、
spec 757 的 `overlay-21:019Eh`——後者寫的是 `參數 div 5` → 索引 4
（Platinum）、`參數 mod 5` → 索引 3（Gold），**正是 5 金幣換 1 白金幣**。

### `DS:2694h` / `DS:269Dh` 方向位移（各 9 bytes）

| 方向 | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|---|
| dx | 0 | 1 | 1 | 1 | 0 | −1 | −1 | −1 | 0 |
| dy | −1 | −1 | 0 | 1 | 1 | 1 | 0 | −1 | 0 |

**0 是正上方，順時針一圈到 7；索引 8 是 `(0, 0)`——原地**。這解釋了 spec 765 的
`(方向 + 4) mod 8 ＝ 正對面`，也解釋了為什麼表是 9 個而不是 8 個。

### `DS:27D8h` 雲霧形狀（值是上表的方向索引）

- 4 格版（`byte[27D7h + k]`，k = 1..4）：`[8, 2, 3, 4]` ＝ 原地、右、右下、下
  → **2 × 2 方塊**
- 9 格版（`byte[27DBh + k]`，k = 1..9）：`[8, 0, 1, 2, 3, 4, 5, 6, 7]` ＝ 原地
  加八個鄰格 → **3 × 3 方塊**

spec 748 當時寫「4 格與 9 格是位元組事實，形狀不是」，現在形狀有了。

### `DS:27BDh` 法術名稱（每筆 `29h` ＝ 41 bytes，**1 起算**，共 56 筆）

`1 Bless` … `56 Restoration`，第 57 格開始全是 0。完整清單在
`docs/audit/resident-data-tables.md`。

### `DS:37DAh` 法術屬性（每筆 16 bytes，索引與名稱表相同）

56 列全部排進了同一份文件。可以直接核對的兩點：

- `+2 = 0FFh` 的五筆是 `Cause Light Wounds`、`Shocking Grasp`、
  `Cause Blindness`、`Cause Disease`、`Bestow Curse`——全部是**接觸型的傷害／
  詛咒法術**，與既有「`+2 = 0FFh` 是哨兵」的結論一致。
- `+7`（spec 719 的目標模式）：`Bless` 與 `Prayer` 是 `04`（全隊）、
  `Cure Light Wounds` 是 `02`（選一個）、`Detect Magic` 與 `Find Traps` 是
  `01`（自己）、`Fireball` 是 `00`（戰鬥限定）。四種模式各有實例對上。

## 兩處先前的過度解讀，已更正

`overlay-21:0F5Bh` 的指令是

```asm
mov ax, [di+6F70h]
or  ax, [di+6F72h]
jz  ...
```

——那是**把兩個 word 或起來測非零**，不是載入遠指標測 NIL。spec 756 當時寫成
「8 個 4-byte 遠指標……檢查是否為 NIL」，spec 764 進一步據此說
「`DS:6F70h` 出現三種互相衝突的型別用法」。**兩處都不成立**：那塊區域從頭到尾
都是 8 格 32-bit 的量，三支函式（測非零、寫商餘、累加）用法一致。

spec 756、757、764 的正文已改寫成正確版本；本節是集中的推翻紀錄。

## 明確不宣稱

- 沒有宣稱法術屬性表 16 個欄位的完整語意。本輪只核對了 `+2` 的哨兵與 `+7` 的
  目標模式兩欄，其餘照排位元組。
- 沒有宣稱 `DS:6F70h` 第 8 格（`6F8Ch`）是什麼；前 7 格與金錢名稱表對得上。
- 沒有宣稱 PC-98 的對應表位置。PC-98 的 `dseg` 已經倒出來了，但表基底要各自從
  該平台的指令反推，**不能用 DOS 的位址加減**（spec 755 已記錄兩平台 resident
  之間沒有固定位移）。
- 未初始化的 `BSS` 位元組在 dump 裡讀出來是 0，那是 dump 的產物不是程式碼寫的
  值——判讀執行期內容不能只看這份檔案。
