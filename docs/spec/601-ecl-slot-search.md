# 第六百零一輪：`3Bh`（在全隊的 `+1Eh` 陣列裡找 id）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:2FDAh`（236 bytes）。

```text
READVAR(3)
id    := ADDRESSVALUE(1)
dest1 := ADDFNC(high[2], low[2])
dest2 := ADDFNC(high[3], low[3])
slot := 1 ; person := 0 ; found := false
node := DS:9598h
while node <> nil and not found do
    slot := 1                                 ← 每名角色重新從 1 開始
    repeat
        if node^[1Eh + slot] = id then found := true
        else if slot <= 64h then slot := slot + 1
        else overflow := true
    until found or overflow
    node := node^[18Ah]
    if node <> nil and not found then person := person + 1
if slot > 64h then slot := FFh                ← 全部找過都沒有
STOREVALUE(dest1, slot)
STOREVALUE(dest2, person)
```

## 角色記錄 `+1Eh`

是一個**以 1 為起點、上限 100（`64h`）格的位元組陣列**——迴圈用
`node^[1Eh + slot]` 取值，`slot` 從 1 起算、超過 `64h` 才停。所以實際佔用的是
`+1Fh`..`+82h`。

這與掛在 `+14Eh` 的物品鏈（[spec 596](596-ecl-party-item-sweep.md)）是**兩套
不同的東西**：一個是定長陣列、一個是動態鏈。

## 兩個輸出

| 輸出 | 內容 |
|---|---|
| operand 2 指定的位址 | 格號 `1..100`，**找不到是 `FFh`** |
| operand 3 指定的位址 | 第幾名角色，**0-based** |

`person` 只在「還有下一個角色而且還沒找到」時才加一，所以找到時它正好是命中
角色的 0-based 序號。

`FFh` 再次當哨兵（`0Eh` 的關閉、`21h`／`37h` 的「operand 無值」）。

## 明確不宣稱

- `+1Eh` 那個陣列裝的是什麼。
- 找不到時 `person` 的值（迴圈跑完後它是「角色數 − 1」，但那是副作用不是設計）。
