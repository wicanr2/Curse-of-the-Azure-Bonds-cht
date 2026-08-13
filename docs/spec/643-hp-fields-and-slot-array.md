# 第六百四十三輪：`+78h` 是 HP 上限、`+1Eh` 起是 84 格的槽位陣列

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-24` 的 `0BD5h`、`0E02h`、`1739h`。

## `+1A5h` 是目前 HP、`+78h` 是上限

`0BD5h` 把兩者直接比：

```text
if arg_6^[1A5h] < arg_6^[78h] then v := 6      ← 受傷
else                               v := 0Ah    ← full
if arg_0 <> 0 then                 v := 0Dh    ← 另一個狀態，覆蓋前兩者
<sub_12E5>(arg_6^[1A5h], @暫存, 0, v, arg_2, arg_4)     ← Str 的 byte 版
接上 ' '
<far 0418h:0D17h>()
```

`+1A5h` 由 [spec 623](623-killthedude-damage-message.md) 的死亡門檻判為目前 HP；
這裡它與 `+78h` 相比、小於就換一個顯示參數，所以 **`+78h` 是上限**。DOS 側的
`20FAh`（[spec 641](641-dos-field-offset-shift.md)）加血後夾在 `+78h`，是同一個結論
的另一條證據。`+78h` 兩平台相同（偏移表 14/14 票）。

`v` 是 `6`／`0Ah`／`0Dh` 三選一，跟著 `Str` 出來的數字一起傳給顯示 routine——
**受傷與滿血的 HP 數字用不同的顯示參數**（多半是顏色）。`arg_0` 非零時蓋掉前兩者。

`sub_12E5` 是 `Str` 的 byte 版（[spec 634](634-str-conversion-family.md)）。

## `+1Eh` 起是 84 格的槽位陣列

```text
1739h(值, 紀錄):
    i := 0
    while i <= 53h do
        if 紀錄^[1Eh + i] = 值 then
            紀錄^[1Eh + i] := 0
            i := 54h              ← 提前結束的寫法
        inc i
```

範圍 `+1Eh + 0..53h` ＝ **`+1Eh`..`+71h`，84 格**。

這與 [spec 624](624-ecl-special-address-space.md) 對上了：`STORESPECIALS` 把 ECL
位址 `7C20h`..`7C70h`（81 個連號）寫進 `紀錄^[1Eh + (addr − 1Fh)]`，也就是
`+1Fh`..`+6Fh`——**完全落在這 84 格之內**。兩邊各自獨立量到同一個陣列。

`i := 54h` 是 Pascal 沒有 `break` 的標準寫法：把迴圈變數設成超界值，`inc` 之後
條件就不成立。**只清掉第一個相符的格子**，後面的同值不動。

## `0E02h`：四個都試，任一成功就算成功

```text
result := 0
for i := 1 to 4 do
    if <sub_25CD>(arg_0, arg_2, byte[49D3h + i], @var_6) <> 0 then
        result := 1
return result
```

**沒有提前結束**——即使第一個就成功，剩下三個照跑。所以 `sub_25CD` 的副作用會發生
四次，不是一次。

索引由 `1` 到 `4`（不是 0..3），查的表在 `DS:49D3h`。

## 明確不宣稱

- `v` 的 `6`／`0Ah`／`0Dh` 實際對應什麼（顏色、屬性或別的）。
- `arg_0` 非零代表什麼狀態。
- `DS:49D3h` 表的內容與 `sub_25CD` 的行為。
- `+1Eh` 那 84 格各自放什麼（只知道 ECL 可以寫入其中 81 格）。
