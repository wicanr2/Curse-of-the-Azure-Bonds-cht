# 第六百二十輪：`CHARSPEED` —— 速度欄位與加速／減速

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-07:20C5h`（205 bytes，76 條指令）。

```text
CHARSPEED(out_min, out_max):
    node := DS:9598h
    out_min^ := node^[1A6h] ; out_max^ := node^[1A6h]   ← 兩者都先取第一人
    while node <> nil do
        v := node^[1A6h]
        if <查效果>(27h, node) then      v := v * 2       ← shl 1
        else if <查效果>(2Ah, node) then v := v div 2     ← 有號 idiv 2
        if v > out_max^ then out_max^ := v
        if v < out_min^ then out_min^ := v
        node := node^[18Ah]
```

## `+1A6h` 是速度

`1Eh`（全隊屬性統計，[spec 603](603-ecl-party-stat-aggregate.md)）的選擇器
`9Fh` 取的也是這個欄位——當時只知道「是某個屬性」，這一輪從加速／減速的處理
確認它是**速度**。

同時它也是角色記錄的**最後一個 byte**（記錄長 `1A7h`＝423，
[spec 604](604-ecl-spawn-monsters.md)）。

## 兩個 effect id

| effect | 效果 | 運算 |
|---:|---|---|
| `27h` | 加速 | `v × 2`（`shl 1`） |
| `2Ah` | 減速 | `v ÷ 2`（**有號** `idiv`） |

**兩者互斥**：`27h` 命中就不檢查 `2Ah`。所以同時中了加速與減速的角色，
**只有加速生效**——不是互相抵消。

`27h` 與 `2Ah` 都出現在 `CHECKFX` 的 timing `12h` 清單裡
（[spec 584](584-checkfx-timing-table.md)：`27h, 2Ah, 3Ah`），三者是同一組。

## 初值取第一人，不是 0 或 FFh

`out_min` 與 `out_max` 都先設成鏈頭那個角色的**原始**速度（**未經加速／減速
調整**），之後才進迴圈重新算含調整的值。所以第一人若中了減速，
`out_min` 的初值會比它調整後的實際值大——但迴圈第一輪就會把它蓋掉，**沒有
實際影響**。

隊伍為空（`DS:9598h = nil`）時直接返回，兩個輸出維持…**未定義**——初值取自
`nil^[1A6h]`，那是對空指標解參照。原版沒有防護。

## 明確不宣稱

- `<查效果>`（`014A:00A7`）在這裡的完整語意（同一支在 `sub_269`、`3Fh` 也用）。
- 速度的單位與後續怎麼用。
