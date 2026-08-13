# 第五百九十二輪：ECL 的算術與比較

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`INTERPET`（overlay-02）。位址為 PC-98 overlay-local。

## `04h`～`07h`：四則運算共用一支（`01B7h`）

```text
READVAR(3)
v1 := ADDRESSVALUE(1)
v2 := ADDRESSVALUE(2)
dest := ADDFNC(high[3], low[3])
case DS:A891h of                          ← 重讀 opcode 分辨
    04h: r := v1 + v2                                        ADD
    05h: r := v2 - v1                                        SUBTRACT
    06h: r := v1 div v2 ; bank1^[67Eh] := v1 mod v2          DIVIDE
    07h: r := v1 * v2                                        MULTIPLY
STOREVALUE(dest, r)
```

四個細節，照抄才會對：

1. **`05h` 是 `operand2 − operand1`，不是 `operand1 − operand2`。**
   （`mov ax, v2` ／ `sub ax, v1`）順序寫反的話所有減法都會變號。
2. **`06h` 用無號 `div`，不是 `idiv`。** 負數會被當成大正數。
3. **`06h` 的餘數存進 bank 1 的 `+67Eh`。** 除法同時產出商與餘數，餘數不是
   丟掉而是放到一個固定位置給後續指令用。原版為此把 `div` 執行了**兩次**
   （一次取商、一次取餘），效率低但行為一致。
4. **`07h` 用無號 `mul` 且只取低 16 位**，溢位直接捨棄。

## `03h`＝`COMPARE`（`011Eh`）：字串與數值兩條路

```text
READVAR(2)
if operand_code[1] >= 80h or operand_code[2] >= 80h then     ← 任一是 packed text
    <far 0418:044A>(0, DS:A9DAh, @b1) ; s1 := b1
    <far 0418:044A>(0, DS:AADAh, @b2) ; s2 := b2
    <far 0062:008E>(s2, s1)                                  ← 字串比較
else
    <far 0062:0093>(ADDRESSVALUE(2), ADDRESSVALUE(1))        ← 數值比較
```

**同一個 opcode 依 operand 的 code 走完全不同的比較**。`DS:A9DAh` 與
`DS:AADAh` 是兩個 packed text 緩衝區，相差 `100h`——第一個就是 opcode `09h`／
`11h`／`12h` 用的那個（[spec 588](588-ecl-return-and-gosub-stack.md)、
[589](589-ecl-text-and-flag-handlers.md)）。

remake 若只實作數值比較，帶字串 operand 的 `COMPARE` 會拿緩衝區位址當數字比。

## 明確不宣稱

- `0062:008E`（字串比較）與 `0062:0093`（數值比較）把結果放到哪裡——多半是
  `DS:A88Ah` 那組旗標，但沒讀它們的本體之前不能說。
- `0418:044A` 的解碼細節。
- bank 1 `+67Eh` 之後被誰讀。
