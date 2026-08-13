# 第六百四十九輪：日文輸入緩衝 —— 濁點合成與空白補齊

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `16ECEh`、`16F2Bh`、`16F81h`。

## 緩衝區

`DS:28A4h` 起是輸入緩衝，**每字一個 word**（存 Shift-JIS 碼）。相關全域：

| 位址 | 用途 |
|---|---|
| `byte_1ED34` | 緩衝容量（顯示與掃描的上限） |
| `byte_1EB33` | 目前已輸入字數 |
| `dword_1EB2D` | 指向「目前長度」的 far pointer |
| `word_1EB31` | 顯示起點在文字 VRAM 的位移 |

## `16F2Bh`：插入一個字，可能與前一字合成

```text
if es:[dword_1EB2D]^ >= byte_1ED34 then return       ← 滿了就不收
si := 28A4h + byte_1EB33 × 2
if <1711Fh>(bl, 2AAAh) <> 0        ← byte 版表查詢
   and byte_1EB33 <> 0
   and [si−1] = 0 then                                ← 前一字還沒有附加記號
    bl 暫存到 bh
    bl := [si−2]                                      ← 取前一字
    if <17148h>(bl, 2AE2h) <> 0 then                  ← 查第二張表
        [si−1] := 原本的 bl                            ← 寫進前一字的高位元組
    else return
else
    [si] := bl
    inc byte_1EB33                                    ← 只有這條路徑會增加字數
inc es:[dword_1EB2D]^
```

兩張表在 `DS:2AAAh` 與 `DS:2AE2h`。`2AAAh` 用 **byte 版**查（`1711Fh`）、
`2AE2h` 用 **word 版**查（`17148h`，[spec 646](646-sjis-to-jis-conversion.md)）——
兩支是同一種表格式的兩個寬度版本，見 [spec 650](650-input-table-lookup-pair.md)。

合成成立時**字數不變**——新輸入的 byte 被寫進**前一個字的高位元組**（`[si−1]`），
把兩個 byte 併成一個 Shift-JIS 碼。這正是日文輸入把濁點／半濁點併進前一個假名的
做法。

不合成時才 `[si] := bl` 並把字數加一。

`es:[dword_1EB2D]^` 兩條路徑**都會加一**，所以它數的是「輸入過的 byte 數」，與
`byte_1EB33`（字數）不同。

## `16F81h`：把緩衝畫到文字 VRAM

```text
<sub_16FEC>()
bx := 29A4h                            ← 注意不是 28A4h
cl := byte_1ED34                       ← 畫滿整個容量
ES := 0A000h ; DI := word_1EB31
<sub_16FCF>()
while cl <> 0 do
    ax := [bx] 的兩個 byte（高低對調取出）
    if ax = 0 then ax := 8140h         ← 空格子補全形空白
    <17186h>()                         ← SJIS → JIS
    ah := ah − 20h
    暫存 ax
    左半格 := (ah, al and 7Fh) 交換後 stosw
    取回 ax
    右半格 := (ah, al or 80h) 交換後 stosw
    cl := cl − 1
```

**空格子補 `8140h`（全形空白）**，所以畫出來一定是滿排的固定寬度欄位，不是隨字數
變短。迴圈跑的是 `byte_1ED34`（容量）而不是目前字數。

寫入慣例與 [spec 648](648-pc98-text-draw-core.md) 的 `16966h` 相同：同一個 JIS 碼
寫進相鄰兩格，左半 `and 7Fh`、右半 `or 80h`，区號先減 `20h`。

`bx` 起點是 `29A4h`，比輸入緩衝的 `28A4h` 高 `100h`——**是另一塊區域**，不是同一
個緩衝。本輪沒有確認兩者的關係。

## `16ECEh`：去掉尾端空格後的長度

```text
ax := byte_1ED34 × 2
bx := 28A4h + ax
cl := byte_1ED34
while cl <> 0 do
    bx := bx − 2
    if [bx] <> 0 then break
    cl := cl − 1
return cl
```

從緩衝**尾端往回掃**，跳過為 0 的格子，回傳最後一個非 0 格的 1 起算序號。全部
為 0 時回傳 0。

## 明確不宣稱

- `DS:2AAAh`／`DS:2AE2h` 兩張表的內容（只確定格式與查法）。
- `29A4h` 與輸入緩衝 `28A4h` 的關係。
- `sub_1711F`／`sub_16FEC`／`sub_16FCF` 的行為。
- 合成規則對應的是濁點、半濁點還是兩者（只確定「把新 byte 併進前一字的高位元組」）。
