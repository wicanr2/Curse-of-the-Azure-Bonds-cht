# 第五百七十七輪：命中判定與 effect 鏈遍歷

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`EFFECTS`（overlay-23）。兩平台已配對，判讀同時適用 DOS。

## `ATTEMPTTOHIT`：完整的命中式

PC-98 `122Ch`／DOS `123Fh`，10 bytes 參數。

```text
ATTEMPTTOHIT(need, attacker_far, target_far) -> boolean
    <前置>(target)                       ← sub_1552
    DS:A039h := ROLLDICE(1, 20)          ← d20，結果存進全域
    if DS:A039h <= 1 then return false   ← 自然 1（含 0）必失手
    if DS:A039h  = 20 then DS:A039h := 100
    CHECKFX(0Ah, target)                 ← 可改寫 DS:A039h
    CHECKFX(10h, attacker)               ← 可改寫 DS:A039h
    if target^[198h] = 0 then v := bank1^[6E2h]
    else                       v := bank1^[6E0h]
    if DS:A039h < 0 then return false    ← CHECKFX 可能把它壓成負值
    return (DS:A039h + target^[19Ah] + v) >= need
```

要點：

- **骰值放在全域 `DS:A039h`，而不是區域變數**——兩個 `CHECKFX` 就是靠這點
  改寫骰值（法術效果調整命中）。remake 若把骰值留在區域變數，這條路就斷了。
- 自然 20 換成 **100**（不是 +20），確保任何修正下都命中。
- 自然 1 的判定是 `<= 1`，把 `CHECKFX` 之前可能出現的 0 也一起吃掉。
- `CHECKFX` 之後再檢查一次 `< 0`，這是**第二道失手閘門**。
- `v` 取自 bank 1 的 `6E0h`／`6E2h`，由 `target^[198h]` 是否為 0 選擇。
  兩個位址差 2，形狀像「同一張表的兩欄」。

`DS:A039h` 是 byte，比較用 `cmp byte ptr`＋`jle`／`jl`，取值用 `cbw`
**符號延伸**——所以它確實可以是負的。

## `sub_269`：effect 鏈遍歷（161 個呼叫者）

PC-98 `0269h`／DOS `0269h`，6 bytes 參數。這是整個 `EFFECTS` 單元被呼叫最多
的函式。

```text
sub_269(id, subject_far)
    found := false
    node  := DS:9598h                        ← 鏈頭 far pointer
    if <查找>(@ctx, id, subject) then found := true
    else if id in [15h, 2Dh, 2Eh, 31h] then
        Move(DS:9F31h, @save, 0D8h)          ← 借用暫存區前先備份
        while node <> nil and not found do
            if <查找>(@ctx, id, node) then
                if DS:7F27h <> 5 then found := true
                else
                    n := if id = 31h then 6 else 1
                    <收集>(node, n, ...) → DS:9F2Ch 起、stride 3 的陣列，
                                            項數在 DS:9F30h
                    for i := 1 to DS:9F30h do
                        if DS:[9F2Eh + i*3] = <取屬性>(subject) then
                            found := true
            node := node^[18Ah]              ← 下一個節點
        Move(@save, DS:9F31h, 0D8h)          ← 還原
    if found then CALLEFFECT(0, ctx…, subject, id)
```

三件確定的事：

1. **`unk_249` 是 Turbo Pascal 的 `set of byte` 常數**——函式前方 32 bytes
   ＝ 256 bits，由 RTL 的 set-in 呼叫（`0A65:08E4`）測試。兩平台的 32 bytes
   **逐位元組相同**，成員是 `{15h, 2Dh, 2Eh, 31h}`。只有這四個 effect 需要
   走完整條鏈，其餘只查 subject 自己。
2. `31h` 在函式裡另有特判（`n := 6`，其餘為 1），與它在集合內一致。
3. `DS:9F31h` 起的 `0D8h` bytes 是**共用暫存區**，遍歷期間會被覆寫，所以
   進迴圈前備份、離開前還原。`DS:9F2Ch`（stride 3 的陣列）與 `DS:9F30h`
   （項數）就落在它前面。

## 明確不宣稱

- `<查找>`（far `014A:00A7`）、`<收集>`（`0189:xxxx` 那組）的本體。
- `DS:9598h` 這條鏈的節點格式（只確定 `+18Ah` 是 next 指標）。
- `DS:7F27h = 5` 代表什麼狀態。
- `{15h, 2Dh, 2Eh, 31h}` 這四個 effect 分別是什麼。
- `ATTEMPTTOHIT` 的 `need` 由誰算出（呼叫端未讀）。
