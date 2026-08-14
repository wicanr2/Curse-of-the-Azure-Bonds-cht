# 第七百三十輪：豁免決定的是「套哪一個」，不是「有沒有」

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-22:4DA0h`。

## 流程

```text
p := DS:7435h
DS:6F95h := 40h

if <overlay-23 entry#8>(p, 4, 0) = 0 then            ← 一般讀成「豁免失敗」
    d := <sub_E11h>(56h)
    <PUTEFFECT>(p, 1Bh, d, 0, 0, 0, 0, '<名字> is clumsy'（CS:4D8Ch）)
    if <overlay-24 entry#27>(p, 1Bh, @node) then
        <overlay-23 entry#2>(1Bh, p, 0, 0)
else
    d := <sub_E11h>(56h)
    <PUTEFFECT>(p, 2Ah, d, 0, 0, 0, 0, '<名字> is slowed'（CS:4D96h）)
    if <overlay-24 entry#27>(p, 2Ah, @node) then
        <overlay-23 entry#2>(2Ah, p, 0, 0)

<sub_F06h>(DS:6F97h, 0, 0, 1, 0, '<名字> is clumsy'（CS:4D8Ch）)   ← ⚠ 無條件
```

**兩條路都會套效果**，差別只在編號（`1Bh` 對 `2Ah`）與訊息。這和 spec 722 的
凝視（豁免成功就完全沒事）、spec 708 的即死（豁免成功改成傷害）都不同——
本作的豁免有三種收場方式，不能假設「豁免成功 ＝ 沒事」。

持續時間兩條路都是 `sub_E11h(56h)`，同一個值。

## 最後那句訊息永遠是 `is clumsy`

函式尾端無條件用 `CS:4D8Ch`（`is clumsy`）呼叫 `sub_F06h`，**即使剛才走的是
`is slowed` 那條**。兩個 `mov di, offset` 的位址一個是 `4D8Ch`、一個是 `4D96h`，
只差 10——很像是複製上一段時忘了改。

本輪不宣稱這是 bug：`sub_F06h` 的第四個參數在這裡是 `1`（spec 701 的致傷是傷害
量、spec 709 的 `24CDh` 是 `0`），有可能那個模式下訊息根本不顯示。要判定得先
讀 `sub_F06h`。

## `2Ah` 這個編號的張力

| 出處 | 用法 |
|---|---|
| `3EE2h`（spec 716） | `entry#16(目標, 2Ah)` 為 **0** 時才顯示 `is Speedy` |
| 本輪 `4DA0h` | 套上 `2Ah`，訊息 `is slowed` |

一個編號同時和「加速」與「減速」有關。最省事的解釋是 `2Ah` 代表
**「速度已經被改過」**，而方向由套用它的法術決定，`3EE2h` 則是「已經被改過就
不再疊加」。本輪沒有讀 `entry#16`／`entry#21` 的內部，所以只記下這個張力，
不下結論。

## 明確不宣稱

- `1Bh`／`2Ah` 的正式語意與兩者的關係。
- `sub_F06h` 第四個參數 `1` 的意義。
- `entry#8` 回傳 0／非 0 到底是不是「豁免失敗／成功」（各處都這樣用，但沒有
  直接讀過那支）。
