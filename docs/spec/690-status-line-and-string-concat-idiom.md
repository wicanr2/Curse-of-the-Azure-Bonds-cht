# 第六百九十輪：狀態列，以及 Turbo Pascal 字串運算式的產碼形狀

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-24:2BAAh`（entry#38）。

## 三支 resident 字串程序組成一條鏈

`2BAAh` 裡這三支輪流出現數十次，形狀固定：

```text
<0A54:0634h>(@結果緩衝, 第一段)      ← 起頭，結果留在堆疊上
<0A54:06C1h>(下一段)                 ← 接一段，結果仍留在堆疊上
<0A54:06C1h>(再一段)
<0A54:064Eh>(@目的, 上限)            ← 收尾，把堆疊上的結果複製出去
```

中間結果**從頭到尾沒有落地**，一直以 far 指標的形式留在堆疊上。這正是
spec 687 觀察到的「`0A54:064Eh` 有時只被推入 3 個 word」的成因——少的那 2 個
就是前一環留下的結果指標。所以那不是特例，是 Borland 對
`s := a + b + c` 這種字串運算式的標準產碼。

推論：連續 `push` 的數量**不等於**被呼叫者的參數個數，只要前一條是這三支之一。
數參數要從鏈的起點算起。

## `2BAAh`：組出狀態列並畫到畫面上

```text
if DS:4FBAh = 3 then 直接返回

x := DS:720Fh；y := DS:7210h                       ← MAXRANGE 用的同一組座標
h := Str(bank0^[192h])；    if 長度 < 2 then h := '0' + h
m := Str(bank0^[190h] × 10 + bank0^[18Eh])；if 長度 < 2 then m := '0' + m

line := ''
if bank0^[1F6h] = 0 then
    line := Str(x) + ',' + Str(y) + ' '            ← 只有這個條件成立才顯示座標
line := line + 方向名[DS:7211h] + ' '              ← DS:2540h 起，一格 3 bytes
line := line + h + ':' + m
if DS:4FC1h <> 0        then line := line + '*'
if DS:4FBAh = 2         then line := line + ' camping'
elif (bank1^[594h] and 1) <> 0 then line := line + ' search'

<resident 01A0:04CDh>(0Fh, 26h, 0Fh, 11h)          ← 先清那一列
<resident 0542:0352h>(11h, 0Fh, 0Ah, 0, line)      ← 畫出來
```

補零的判斷是 `cmp [bp+var_5], 2` ＋ `jnb`——**比的是 Pascal 字串的長度位元組**，
不是字元值。

`DS:7211h` 就是 `MAXRANGE(dir, x, y)` 的第一個參數（spec 687），方向索引乘 3
去查 `DS:2540h`，所以方向名是三個 byte 一格的 Pascal 短字串（長度 1 ＋ 兩個
字元），也就是 `N`／`S`／`E`／`W` 這種兩字母以內的縮寫。方向編碼是 0／2／4／6
（spec 中 `MAXRANGE` 那筆），乘 3 之後是 0／6／12／18。

`line` 的緩衝上限是 `28h` ＝ 40 字元。

## 用到的字串常數

| 位址 | 內容 |
|---|---|
| `CS:2B8Fh` | `0` |
| `CS:2B91h` | `,` |
| `CS:2B93h` | （空白） |
| `CS:2B95h` | `:` |
| `CS:2B97h` | `*` |
| `CS:2B99h` | ` camping` |
| `CS:2BA2h` | ` search` |

中文化時這一列是**定寬 40 字元**，而 ` camping`／` search` 直接接在時間後面，
沒有補齊；換成中文會佔兩倍寬度，要先確認 40 的上限。

## 明確不宣稱

- `bank0^[192h]`／`[190h]`／`[18Eh]` 的單位。形狀（兩個補零到兩位、以 `:` 相接、
  第二個由 `×10 +` 組成）與時鐘一致，但沒有直接證據，本輪只記算式。
- `bank0^[1F6h]` 為何會讓座標整段不顯示。
- `DS:4FC1h` 的 `*` 代表什麼狀態。
- `DS:2540h` 那張方向名表的實際內容（在 resident 的 DS，不在本 overlay 裡）。
- `0542:0352h` 前兩個參數的確切語意。
