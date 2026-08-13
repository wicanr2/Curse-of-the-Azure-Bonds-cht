# 第六百一十五輪：`FINDGUY` 與 `MAXRANGE`

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`ECL2`（overlay-07）。位址為 PC-98 overlay-local。

## `FINDGUY(ptr)`（`085Ch`）

```text
node := DS:9598h ; index := 0
while node <> nil do
    if node = ptr then return index           ← 0-based 序號
    index := index + 1 ; node := node^[18Ah]
return index                                  ← 找不到時回傳「總數」
```

**找不到不是回 `FFh`，而是回傳鏈的長度。** 呼叫端若不另外檢查，會把「不存在」
當成「最後一個之後那一格」。

與 `0Ah`（選第 N 名隊員，[spec 599](599-ecl-select-member-and-0c.md)）互為反向：
一個由序號取指標、一個由指標取序號，兩邊的序號都是 0-based。

## `MAXRANGE(dir, x, y)`（`05BDh`）

```text
if bank0^[1CCh] = 0 then
    bank1^[582h] := 2 ; return 2              ← 非地城模式固定 2
steps := 0 ; result := 0
while steps < 2 do
    if <far 017C:0034>(dir, x, y) then break  ← 該格走不過去
    steps := steps + 1 ; result := steps
    case dir of
        0: y := y − 1
        2: x := x + 1
        4: x := x + 1
        6: y := y − 1
return result
```

兩個要點：

1. **射程上限固定 2 格**，寫死在 `cmp [bp+var_3], 2`。
2. **方向編碼是 `0`／`2`／`4`／`6`**——間隔 2，不是 `0..3`。這與虛擬地圖暫存器
   `C04Bh`..`C04Fh`（[spec 563](563-ecl-memory-model-and-operand-resolution.md)）
   的用法要一起看：ECL 層看到的方向值必須是偶數。

`bank0^[1CCh] = 0`（非地城）時**完全跳過檢查**直接回 2，並順手把
`bank1^[582h]` 也設成 2——這個 bank 1 欄位在 `0Ch`／`0Dh`／`24h`／`29h` 都出現過
（[spec 599](599-ecl-select-member-and-0c.md)、
[606](606-ecl-redraw-dispatch.md)、[611](611-ecl-parlay.md)），語意一致：
**它是「可用步數／次數」**。

## 明確不宣稱

- `dir` 的 `0`／`2`／`4`／`6` 各對應哪個方位（只知道 `0` 與 `6` 動 `y`、
  `2` 與 `4` 動 `x`，而且 `2` 與 `4` 都是 `x + 1`）。
- `017C:0034`（可通行判定）的本體。
