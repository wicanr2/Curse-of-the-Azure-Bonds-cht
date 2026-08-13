# 第六百四十二輪：`overlay-12` 傷害修飾五支 —— 用迴圈寫的「減 N 但不低於下限」

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-12` 的 `2237h`、`2274h`、`22C6h`、`24B6h`、`257Eh`。

## `24B6h`：減 N 次，每次都夾一遍下限

```text
if DS:0A02Fh and 1 <> 0 then
    n := DS:0A032h
    for i := 1 to n do
        dec DS:0A02Eh
        if DS:0A02Eh < DS:0A032h then
            DS:0A02Eh := DS:0A032h
```

減的次數與下限**是同一個變數** `DS:0A032h`。而且夾限在迴圈**裡面**，每減一次就
夾一次——結果等同 `max(傷害 − n, n)`，但寫法是 `n` 次迴圈。

`n = 0` 時整段跳過（`cmp 1, n` ＋ `ja`）。比較是**無號**（`jnb`）。

remake 若改寫成一行算式，行為相同；但若把下限與次數當成兩個獨立參數就會偏離
原作——它們在這裡綁在一起。

## 傷害減半的三種判準

`DS:0A02Eh`（傷害值，[spec 640](640-save-for-half-and-damage-global.md)）在
`overlay-12` 裡有三支各自減半，條件都不一樣：

| 函式 | 條件 | 比較 |
|---|---|---|
| `133Fh`（[spec 639](639-overlay12-poison-and-race-gates.md)）| `p^[5Ah] < 3` | 有號 `jge` |
| `257Eh`（本輪）| `p^[5Ah] > 0` | 有號 `jle` |
| `1A34h`（[spec 640](640-save-for-half-and-damage-global.md)）| 豁免 ＋ 免疫表 | — |

前兩支查的是**同一個欄位** `p^[5Ah]`（`p` 都來自 `<sub_FCC>(DS:9594h)`），門檻卻
不同——`< 3` 與 `> 0` 有重疊（`1`、`2` 兩個值都成立）。兩支都會減半的情況存在。

除法一律是 `cwd` ＋ `idiv 2`（有號）。

## 其餘三支

```text
2237h:  if (DS:0A02Fh and 1) = 0 and (DS:0A02Fh and 10h) = 0 then
            <sub_1437>(arg_8, arg_6, 66h, ROLLDICE(3, 6), 0FFh, 1)

2274h:  if <far 014Ah:00A7h>(arg_6, arg_8, 62h, @var_4) = 0
           and <far 014Ah:00A7h>(arg_6, arg_8, 3Bh, @var_4) = 0 then
            <sub_1437>(arg_8, arg_6, 3Bh, 3, 0FFh, 1)

22C6h:  if <far 0176h:0473h>(arg_6, arg_6^[78h]) = 0 then
            <sub_3C>(arg_6, 66h, arg_2^[3], 1)
```

`2274h` 連查兩個效果 id（`62h` 與 `3Bh`），**兩個都沒有**才施加 `3Bh`——避免重複
套用同一個效果，而且順帶排除了 `62h`。

`2237h` 與 `2274h` 都呼叫 `sub_1437`，倒數第二個參數固定 `0FFh`、最後一個固定
`1`；差別在效果 id（`66h` 對 `3Bh`）與強度（`3d6` 對常數 `3`）。

`22C6h` 的 `sub_3C` 呼叫形狀與 [spec 639](639-overlay12-poison-and-race-gates.md)
的 `1973h`、`13F1h` 一致（`(far pointer, id, byte, 旗標)`），是第三個呼叫端。

## 明確不宣稱

- `DS:0A032h`（次數兼下限）、`p^[5Ah]`、`+78h` 的意義。
- 效果 id `3Bh`／`62h`／`66h` 的語意。
- `sub_1437`／`sub_FCC`／`0176h:0473h` 的行為。
