# 第七百一十七輪：範圍法術怎麼填目標陣列

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-22:370Eh`。

## 流程

```text
DS:6F9Dh := 1
if DS:6F97h = 40h then n := ROLLDICE(1, 3) × 2 + 1      ← 3、5 或 7
else                    n := <overlay-24 entry#36>(DS:6F97h)

if [4F99h]^[1CCh] = 0 then                              ← 只有這個情況才重選目標
    <overlay-31 entry#6>(DS:7559h, DS:755Ah, 2, 0FFh, 1, DS:6E92h)
    for k := 1 to DS:6E96h do
        idx := byte [6E94h + k×3]
        [7431h + k×4] := far [6D35h + idx×4]
    DS:7434h := DS:6E96h                                ← 連筆數一起換掉

<overlay-32 entry#13>(DS:7559h, DS:755Ah, 0, 8)
d := ROLLDAMAGEDICE(n, 6)                               ← n d6
<sub_F06h>(DS:6F97h, 0, 0, d, 9, 空字串（CS:370Dh））
```

## 選目標的那一段補完了 spec 699

spec 699 的 `overlay-24:274Ch` 呼叫的是同一支 `overlay-31 entry#6`（兩處都推
7 個 word，和它的 `retf 0Eh` 吻合），但那裡是**借用緩衝查完就還原**。這裡才是
正常用途：

```text
overlay-31 entry#6 → 填 DS:6E97h 起的 3 bytes 記錄（筆數 DS:6E96h）
                     每筆的 +0 是 DS:6D35h 那張 far 指標表的索引
本輪 370Eh        → 把那些指標搬進 DS:7431h 的目標陣列，並設定 DS:7434h
```

所以 `DS:6E97h` 那片是**選目標的中間結果**，`DS:7431h` 才是法術實際作用的清單。
spec 699 之所以要備份還原，正是因為這片是共用的。

## 三件要注意的

**`[4F99h]^[1CCh] <> 0` 時整段跳過。** 目標陣列維持呼叫前的內容——也就是沿用
上一條 opcode 留下的清單（可能含 `NIL`，spec 716）。所以同一支 opcode 在兩種
情境下的目標來源完全不同。

**筆數是整個換掉不是累加**（`DS:7434h := DS:6E96h`）。前一輪殘留在
`[7431h + k×4]` 而超出新筆數的那幾格不會被清，只是走不到。

**`DS:6F97h = 40h` 是寫死的特例**：骰 `1d3 × 2 + 1`，值域只有 `3`／`5`／`7`
（都是奇數）。其他情況才去查 `entry#36`。這個 `40h` 和 spec 708 `45B5h` 裡
`DS:6F95h := 40h`、spec 703 `4289h` 的哨兵 `40h` 是不同的東西——`DS:6F97h` 是
**分派用的法術／效果編號**（spec 721），不是施法者。

## 明確不宣稱

- `overlay-31 entry#6` 六個參數的意義（`2`、`0FFh`、`1` 三個常數）。
- `overlay-32 entry#13` 的行為。
- `DS:7559h`／`755Ah`、`DS:6F9Dh` 的語意。
- `[4F99h]^[1CCh]` 區分的是哪兩種情境。
