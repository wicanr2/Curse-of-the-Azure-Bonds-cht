# 第七百二十五輪：區域吐酸——同一支裡兩個迴圈，只有一個檢查 `NIL`

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-22:5AA8h`（`retf 0Ah`）。

## 流程

```text
DS:7746h := 0
if DS:758Dh = 0 then arg_2^[3] := 3                  ← 次數初始化成 3
if arg_2^[3] <= 0 then 返回                           ← 用完就不能再吐

DS:6F95h := 30h                                       ← 傷害旗標
x, y := entry#15/#16(arg_6)
<[DS:72A0h] 第 0 項>(@DS:7746h, 1, 3Dh)               ← 模式碼 3Dh（區域）
if DS:7746h <> 0 then <sub_175Bh>(x, y, DS:7559h, DS:755Ah, 1, 6)

cls := <overlay-24 entry#30>(arg_6)
for k := 1 to DS:7434h do                             ← 迴圈 A：⚠ 沒有 NIL 檢查
    if [7431h + k×4]^[197h] <> cls then DS:7746h := 0

if DS:7746h = 0 或 DS:7434h = 0 then 返回

<overlay-24 entry#20>(arg_6, '<名字> breathes acid'（CS:5A9Ah）, 0Ah, 1)
<overlay-24 entry#24>(12h)
<overlay-24 entry#25>(自己的 x,y, DS:7435h 的 x,y, 1, 1Eh)

for k := 1 to DS:7434h do                             ← 迴圈 B：有 NIL 檢查
    p := far [7431h + k×4]
    if p = NIL then 跳過
    s := <overlay-23 entry#8>(p, 3, 0)
    <overlay-23 entry#20>(p, arg_6^[78h], 2, s)

arg_2^[3] := arg_2^[3] − 1                            ← 扣一次
<overlay-24 entry#34>(arg_6)
```

## 兩個迴圈的檢查不一致

同一支函式、同一個陣列、隔了三十幾條指令：

```text
迴圈 A：les di, [di+7431h] ; mov al, es:[di+197h]        ← 直接取欄位
迴圈 B：mov ax, [di+7431h] ; or ax, [di+7433h] ; jz 跳過  ← 先檢查
```

所以「有沒有檢查 `NIL`」不是**函式**的性質，是**每一段迴圈各自的**性質
（spec 716 的掃描是以函式為單位，這一支會被算成「有檢查」而漏掉迴圈 A）。

迴圈 A 讀到 `NIL` 時取的是中斷向量表 `0000:0197h` 的內容，而它的作用是
「這一格跟我不同類就取消整次攻擊」——**一個殘留的 `NIL` 格子就足以讓吐酸不發動**。

## 全體同類才發動

`cls := entry#30(攻擊者)`，接著逐格比對 `+197h`。只要有一格不等於 `cls` 就把
`DS:7746h` 清成 0，整支返回。也就是**目標陣列裡混進任何一個不同類的單位，
攻擊就取消**——看起來是避免波及己方的保護，但本輪沒有讀 `entry#30`，
所以只記形狀不下定論。

## 次數

`arg_2^[3]` 是可用次數：`DS:758Dh = 0` 時重設成 3，每次成功發動扣 1，
歸零後整支立刻返回。**扣的動作在最後**，所以中途任何一個 `return` 都不會扣——
被取消的那些不消耗次數。

傷害參數和 spec 723 的單體吐酸一樣是 `arg_6^[78h]`（攻擊者 HP 上限）。

## 明確不宣稱

- `overlay-24 entry#30`／`entry#34` 的行為。
- `DS:758Dh`、`DS:6F95h := 30h` 的語意。
- 模式碼 `3Dh` 與 `41h`（單體版）的差別。
