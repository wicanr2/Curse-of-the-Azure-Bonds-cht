# 第五百八十六輪：ECL handler `31h`／`33h`／`34h`

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`INTERPET`（overlay-02）。位址為 PC-98 overlay-local。

## `31h`（`2E06h`，38 bytes）

```text
ECL_PC := ECL_PC + 1                  ← 沒有 operand
if DS:BDF7h <> 0 then
    DS:A66Ch := 1
    <far 0172:0025>()
    DS:BDF7h := 0
    DS:BDF4h := 0
```

**無 operand**（`READVAR` 沒被呼叫，PC 只前進一格）。條件不成立時整條指令
什麼都不做。

## `33h`（`2E61h`，43 bytes）

```text
ECL_PC := ECL_PC + 1                  ← 沒有 operand
if DS:BDF3h <> 0 then DS:BDF3h := 0
DS:9636h := 1
DS:9637h := DS:9637h + 1
```

兩個分支都做最後兩件事，差別只在要不要清 `DS:BDF3h`。

`DS:9637h` 就是 `PUTDAMAGE` 拿來當訊息停留時間的那個全域
（[spec 581](581-putdamage-pipeline.md)：`delay := DS:9637h + 1`）。所以這條
ECL 指令的效果是**把後續訊息的停留時間加長一格**。

## `34h`（`2E2Ch`，53 bytes）

```text
READVAR(2)
a := ADDRESSVALUE(1)
b := ADDRESSVALUE(2)
<far 0105:002A>(a, b)
```

兩個 operand 解出後原樣交給外部常式，handler 自己不做判斷。

## `21h`／`37h`（`0C81h`）：IDA 邊界完全不對

兩個 opcode 共用這支 handler。IDA 標的 size 是 53 bytes，但實際的 body 至少
延伸到 `0CFE h` 之後——`0CB6h` 的 `jnz` 跳回 `0C9A h` 形成 `for i := 1 to 3`
的迴圈，迴圈結束後還有一整段（`DS:789Bh := 1`、對 `[bp-3]` 的 `FFh`／`7Fh`
判斷、讀 bank 0 的 `+1CCh`…），而且 `0CB8h..0CBEh` 這 7 bytes **IDA 根本沒
認成指令**。

已讀到的部分：

```text
READVAR(3)
DS:7899h := 1
for i := 1 to 3 do local[i] := ADDRESSVALUE(i)
DS:789Bh := 1
if local[1] <> FFh and local[1] <> 7Fh and bank0^[1CCh] <> 0 then …
```

**這支標記為`待解讀`**，要用位址範圍重讀並補上未解碼的位元組才算完整。

## 明確不宣稱

- `DS:BDF7h`／`DS:BDF4h`／`DS:BDF3h`／`DS:A66Ch`／`DS:9636h` 的語意。
- `0172:0025`／`0105:002A` 的本體。
- `31h`／`33h`／`34h` 在 ECL 指令集裡叫什麼名字。
