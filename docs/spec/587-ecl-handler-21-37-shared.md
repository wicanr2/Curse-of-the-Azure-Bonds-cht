# 第五百八十七輪：共用 handler 怎麼分辨自己被哪個 opcode 呼叫

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:0C81h..0E0Dh`（**396 bytes**，IDA 標的是 53）。

## 共用 handler 的分辨方式

opcode `21h` 與 `37h` 走同一支 handler。它靠**再讀一次 dispatcher 的 opcode
全域**來分辨：

```text
0CB8:  cmp byte ptr ds:A891h, 21h
0CBD:  jnz 0D22h                      ← 不是 21h 就走 37h 那一半
```

`DS:A891h` 正是 dispatcher 取 opcode 的位址（DOS 為 `DS:75FFh`，
[spec 560](560-ecl-opcode-dispatch-table.md)）。**opcode 值在整條 handler
執行期間都還留在那裡**，共用 handler 因此不需要額外參數。

分派表裡標「多個 opcode 指向同一位址」的每一支，都要檢查有沒有這種二次判讀
——只實作其中一個 opcode 的行為會讓另一個靜靜地做錯事。

## 內容

```text
READVAR(3)
DS:7899h := 1
for i := 1 to 3 do v[i] := ADDRESSVALUE(i)      ← v 存在 [bp-3]、[bp-2]、[bp-1]

if opcode = 21h then
    DS:789Bh := 1
    if v[1] <> FFh and v[1] <> 7Fh and bank0^[1CCh] <> 0 then
        bank0^[18Ah] := v[1]
        <far 017C:004D>(v[1])
        bank1^[592h] := 0
    if v[3] <> FFh and bank0^[1CCh] = 0 and DS:A325h <> 50h then
        <far 0176:004D>(79h)
else                                             ← opcode = 37h
    DS:789Ah := 1
    if v[1] = 7Fh then
        <far 017C:0048>(1, 0)
    else if bank0^[1CEh] <> 0 and bank0^[1D0h] <> 0 then
        if v[1] <> FFh then <far 017C:0048>(1, v[1])
        if v[3] <> FFh then <far 017C:0048>(3, v[3])
    else
        for i := 1 to 3 do
            if v[i] <> FFh then <far 017C:0048>(i, v[i])
            else                DS:[A2AAh + i*4] := FFFFh

if DS:789Ah <> 0 and DS:789Bh <> 0 and DS:7F28h = 3 then
    if DS:7F27h <> 3 and DS:BE00h <> 0 then
        <far 019E:014A>() ; <far 014A:002A>(DS:9594h) ; <far 014A:00DE>()
    DS:BE00h := 0
```

`FFh` 與 `7Fh` 是兩個哨兵值：`FFh` 代表「這個 operand 沒有值」，`7Fh` 另有
含義（`21h` 分支把它和 `FFh` 一起排除，`37h` 分支則對它做特別處理）。

`DS:789Ah`／`DS:789Bh` 是兩個各由一半設定的旗標，**兩個都為真才會執行結尾那
段收尾**——也就是說 `21h` 與 `37h` 必須都執行過一次，收尾才會發生。

## IDA 的邊界差了 7 倍

IDA 標 53 bytes，實際 396。兩處位元組**完全沒被認成指令**：

| 範圍 | 內容 |
|---|---|
| `0CB6h..0CBEh` | `jnz`（迴圈回跳）＋ **上面那條 opcode 二次判讀** |
| `0D22h..0D2Eh` | `37h` 分支的開頭（`DS:789Ah := 1` 與 `7Fh` 判斷） |

**漏掉的正好是這支函式最關鍵的兩段。** 靠 IDA 的函式清單讀會得到一支
「只做 `READVAR(3)` 然後填三個變數」的空殼。

`scripts/show.py --whole` 現在用原始位元組找下一個 `55 89 e5` prologue 來定
邊界，不再相信 IDA 的 size，也不用「IDA 認的下一個函式起點」（IDA 把
`0CBFh` 也當成一支函式，那還在同一支裡面）。

已複驗 `00E8h`／`2E06h`／`2E2Ch`／`2E61h` 四支：真實邊界與 IDA 的 size 一致，
先前的判讀完整。

## 明確不宣稱

- `21h`／`37h` 在 ECL 指令集裡叫什麼。
- `7Fh` 這個哨兵的含義。
- `DS:7899h`／`789Ah`／`789Bh`／`A325h`／`BE00h`／`7F28h` 的語意。
- `017C:004D`／`017C:0048`／`0176:004D` 的本體。
