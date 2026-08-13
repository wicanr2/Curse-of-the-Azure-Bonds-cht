# 第六百零六輪：`24h`（重繪畫面的分派）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:1820h`（684 bytes，206 條指令）。無 operand。

```text
ECL_PC := ECL_PC + 1
if DS:BDFBh <> 0 or DS:BDE8h <> 0 then goto second
if bank1^[6D8h] = 1 then
    bank1^[6D8h] := 0 ; ⟪005E:002F⟫ ; ⟪005E:0025⟫
else if bank1^[5C4h] = 1 then
    bank1^[5C4h] := 0 ; ⟪0050:002A⟫ ; ⟪0050:0025⟫
else
    ⟪0057:002A⟫ ; ⟪0057:0025⟫
if bank0^[1CCh] = 0 then ⟪016D:003E⟫ else ⟪017C:0057⟫
goto tail

second:
    n := ⟪0062:003E⟫(DS:A2ABh, DS:A2AAh, DS:A2A9h)     ← 由地圖座標算
    if n > bank1^[582h] then bank1^[582h] := n          ← 取較大者
    ⟪006F:0025⟫ ; ⟪0057:002A⟫ ; ⟪0057:0025⟫
    if bank0^[1CCh] = 0 then
        ⟪016D:003E⟫
        if DS:A325h <> 50h then ⟪0176:004D⟫(79h)
    else ⟪017C:0057⟫
    if DS:BDE8h <> 0 then DS:BDE8h := 0
```

⟪…⟫ 標示的每一個動作都包在同一組前後包夾裡：

```text
if DS:8B5Ah <> 0 then <far 0893:0000>(DS:4838h)
<動作>
if DS:8B5Ah  = 0 then <far 0893:010Dh>()
```

這個包夾在這支函式裡出現 **12 次**，形狀是「動作前後暫停／恢復某個常駐服務」。
`DS:8B5Ah` 決定包夾往哪個方向作用——非 0 時做前置、為 0 時做後置，**兩者不會
同時發生**。

## 與 `0Ch` 的關係

`second` 分支算 `n` 的方式與 `0Ch`（[spec 599](599-ecl-select-member-and-0c.md)）
完全相同（同一支 `0062:003E`、同樣三個虛擬地圖暫存器），差別是 `0Ch` 對
`bank1^[582h]` 取**較小者**、這裡取**較大者**。

## 明確不宣稱

- `DS:8B5Ah` 與 `0893:0000`／`0893:010Dh` 是什麼服務。
- 三種畫面模式（`6D8h`／`5C4h`／預設）各自對應什麼。
- `bank0^[1CCh]` 的語意（它在 `21h`／`37h` 也出現過，
  [spec 587](587-ecl-handler-21-37-shared.md)）。
