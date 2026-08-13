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

物品節點目前確定兩個欄位：`+52h` next、`+56h` 類型 id。

`DS:9598h`（角色鏈頭）與 `DS:9594h`（目前目標，
[spec 595](595-ecl-target-selection-and-effect-query.md)）是相鄰的兩個 far
pointer，一個指鏈頭、一個指當前選中的成員。

## 明確不宣稱

- 物品節點的大小與其餘欄位。
- `014A:0075`（移除）與 `014A:0043`（每名角色的收尾）的本體。
