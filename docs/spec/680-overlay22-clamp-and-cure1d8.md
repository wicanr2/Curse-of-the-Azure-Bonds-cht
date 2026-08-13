# 第六百八十輪：座標夾限、物品鏈掃描，與 `1d8` 的治療

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-22` 的 `1897h`、`09EEh`、`1FB6h`。

## `1897h`：兩個座標各自夾限

```text
if arg_8^ > 31h then arg_8^ := 31h
else if arg_8^ < 0 then arg_8^ := 0
if arg_4^ > 18h then arg_4^ := 18h
else if arg_4^ < 0 then arg_4^ := 0
```

用的是 **`jle`／`jge`（有號）**，所以負值真的會被夾到 0——與
[spec 671](671-dos-video-page-and-dead-bounds.md) 的 `15731h` 對照：那支用無號
`jb`，下限判斷因此是死碼。**同一個專案裡兩種寫法都有**，一個真的檢查、一個沒有。

上限 `31h`（49）與 `18h`（24）——**50 × 25**。與 `15731h` 的 40 × 25 不同，是另一個
座標系。

參數是**指標**（`les di, [bp+arg_8]` 之後直接改 `es:[di]`），所以夾限的結果寫回
呼叫端。

## `09EEh`：清陣列再掃物品鏈

```text
i := 0
重複:
    di := i × 4
    [di+7CCEh] := 0 ; [di+7CD0h] := 0      ← far pointer 陣列，清空
    if i = 30h then 跳出
    i := i + 1
DS:7D92h := 0
p := DS:9594h^[14Eh]                        ← 角色的物品鏈頭
while p <> nil do
    if <sub_64E8>(p) <> 0 then <sub_90E>(arg_0)
    p := p^[52h]
```

陣列在 `DS:7CCEh`，每筆 4 bytes（far pointer）。迴圈是**後測**（先做再比），
`i` 走 `0..30h`，所以清了 **49 筆**（`0..48`），不是 48 筆。

物品鏈的 next 在 `+52h`（[spec 621](621-ecl2-robstuff.md)）。條件成立時呼叫的
`sub_90E` **只吃 `arg_0`，不吃目前這個物品** ——所以它作用的對象另有來源
（多半是前一步寫進 `DS:7CCEh` 陣列或 `DS:0A388h` 的東西）。

`DS:0A388h`／`DS:0A38Ah` 是走訪用的暫存 far pointer。

## `1FB6h`：`1d8` 的治療

```text
if DS:0A520h = 0 then return                ← 沒有選定目標就不做
t := DS:0A521h                               ← 目標陣列第 1 筆（spec 630）
if <far 013Eh:008Eh>(t, ROLLDICE(1, 8), 0) <> 0 then
    <far 014Ah:054Fh>(t)
```

`013Eh:008Eh` 與 [spec 627](627-spell-cure-family.md) 的 `434Dh`／`5B2Bh` 是同一支
（那邊是 `2d4+2`），這裡的量是 **`1d8`**——正是 AD&D 的 **Cure Light Wounds**。

加上先前定出的 `2d8+1`（`4416h`）與 `3d8+3`（`46CAh`），**cure 系列的三個標準級距
都齊了**：

| 量 | AD&D | 位置 |
|---|---|---|
| `1d8` | Cure Light Wounds | `1FB6h`（本輪） |
| `2d8+1` | Cure Serious Wounds | `4416h` |
| `3d8+3` | Cure Critical Wounds | `46CAh` |

`2d4+2`（`434Dh`／`5B2Bh`／`43DAh`）不在這三級裡，是另一個東西。

先檢查 `DS:0A520h`（選定目標數）再動作——沒選到目標時**連骰都不擲**。

## 明確不宣稱

- `DS:7CCEh` 陣列與 `DS:7D92h` 的用途。
- `sub_64E8`／`sub_90E`／`014Ah:054Fh` 的行為。
- `1897h` 的 50 × 25 是哪一個畫面的座標系。
