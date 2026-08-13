# 第五百九十六輪：`40h` 與物品鏈的結構

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:369Fh`。

## `40h`：對全隊移除某種物品

```text
READVAR(1)
id := ADDRESSVALUE(1)
node := DS:9598h                              ← 角色鏈頭
while node <> nil do
    item := node^[14Eh]                       ← 這名角色的物品鏈頭
    while item <> nil do
        next := item^[52h]
        if id = item^[56h] then <far 014A:0075>(item, node)   ← 移除
        item := next                          ← 先取 next 再移除，鏈不會斷
    <far 014A:0043>(node)                     ← 每名角色處理完的收尾
    node := node^[18Ah]                       ← 下一名角色
```

**先取 `next` 再移除**——移除會釋放節點，順序寫反就會讀到已釋放的記憶體。

## 兩條鏈的結構

| 鏈 | 鏈頭 | next 偏移 |
|---|---|---|
| 角色 | `DS:9598h` | 節點 `+18Ah` |
| 物品（每名角色一條） | 角色 `+14Eh` | 節點 `+52h` |

角色鏈的 `+18Ah` 與 `DS:9598h` 先前已知（`sub_269` 遍歷 effect 時走同一條，
[spec 577](577-attempttohit-and-effect-chain-walk.md)），這裡確認**同一條鏈同
時掛著 effect 與物品**，只是入口偏移不同：

| 角色記錄偏移 | 內容 | 出處 |
|---|---|---|
| `+0F2h` | effect 鏈頭（9-byte 節點） | [spec 578](578-effect-node-list.md) |
| `+14Eh` | 物品鏈頭 | 本輪 |
| `+18Ah` | 下一名角色 | [spec 577](577-attempttohit-and-effect-chain-walk.md) |

物品節點是 **`67h`＝103 bytes**（`1Ch` 的 `Dispose(node, 67h)` 直接給的），
目前確定兩個欄位：`+52h` next、`+56h` 類型 id。

`DS:9598h`（角色鏈頭）與 `DS:9594h`（目前目標，
[spec 595](595-ecl-target-selection-and-effect-query.md)）是相鄰的兩個 far
pointer，一個指鏈頭、一個指當前選中的成員。

## `3Eh`（`35AFh`）：掃全隊看還有沒有人能行動

```text
ECL_PC := ECL_PC + 1                      ← 沒有 operand
<far 00E9:002F>(1, 0)
DS:789Dh := DS:9594h                      ← 記住目前目標
<far 0418:14C7>(0Fh, 26h, 1, 11h) ; <far 014A:002A>(DS:9594h)
DS:7F34h := 1                             ← 先假設「沒有」
node := DS:9598h
while node <> nil and DS:7F34h <> 0 do
    if (node^[196h] = 0 or node^[196h] = 4) and node^[0F7h] = 0 then
        DS:7F34h := 0                     ← 找到一個就停
    node := node^[18Ah]
```

狀態 `0` 是可行動、`4` 是加血後回復的狀態
（[spec 579](579-character-status-fields.md)）。迴圈條件帶著 `DS:7F34h <> 0`，
所以**找到第一個就提早結束**。

`DS:7F34h` 的語意因此是「全隊都不能行動」——1 代表沒人可動。

## `1Ch`（`1294h`）：清空一整條物品鏈

```text
ECL_PC := ECL_PC + 1                      ← 沒有 operand
DS:789Ch := 0 ; DS:BDFBh := 0 ; DS:A895h := 8
FillChar(DS:A00Ah, 1Ch, 0)                ← 清 28 bytes
while DS:A026h <> nil do
    next := node^[52h]
    Dispose(node, 67h)                    ← 節點 103 bytes
    DS:A026h := next
```

`DS:A026h` 是另一條物品鏈的鏈頭（**不是**掛在角色身上的那條），next 偏移
同樣是 `+52h`，所以兩者是同一種節點。節點大小 `67h` 由 `Dispose` 的參數直接
給出。

## 明確不宣稱

- 物品節點 103 bytes 裡除了 `+52h`／`+56h` 之外的欄位。
- `DS:A026h` 這條鏈裝的是什麼（地上的物品？遭遇的戰利品？）。
- `DS:A00Ah` 那 28 bytes 與 `DS:A895h := 8` 的語意。
- `014A:0075`（移除）與 `014A:0043`（每名角色的收尾）的本體。
