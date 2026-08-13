# 第五百八十八輪：ECL `RETURN` 與 GOSUB 堆疊

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`INTERPET`（overlay-02）。位址為 PC-98 overlay-local。

## opcode `13h` ＝ `RETURN`（`0AF9h`）

```text
ECL_PC := ECL_PC + 1                      ← 沒有 operand
if DS:A882h <> nil then
    ECL_PC   := node^[0]                  ← 取回返回位址
    next     := node^[2]
    Dispose(node, 6)
    DS:A882h := next
else
    <opcode 00h 的 handler>(sub_52)       ← 堆疊空
```

**GOSUB 堆疊是 6-byte 節點的單向鏈**，鏈頭 far pointer 存在 `DS:A882h`：

| 偏移 | 寬度 | 內容 |
|---:|---|---|
| `+0` | word | 返回用的 ECL PC |
| `+2` | far ptr | next |

節點大小 6 是 `Dispose` 的參數直接給的。

**堆疊空的 `RETURN` 不是錯誤，而是轉去執行 opcode `00h` 的 handler**
（`0052h`，分派表對得上）。remake 若在這裡報錯或忽略，行為會與原版不同。

## opcode `09h`（`02B8h`）：依 operand code 分兩條路

```text
READVAR(2)
addr := ADDFNC(high[2], low[2])
if operand_code[1] < 80h then
    <far 0062:006B>(ADDRESSVALUE(1), addr)
else
    <far 0062:0075>(DS:A9DAh, addr)       ← packed text 的緩衝區
```

`DS:A918h` 是 operand code 陣列的第 1 項（陣列起點 `DS:A917h`，
[spec 563](563-ecl-memory-model-and-operand-resolution.md)）。`>= 80h` 就是
packed text（code `80h`），這時傳的不是解出來的值，而是 `DS:A9DAh` 這個緩衝區
的位址。

**同一個 opcode 依 operand 的 code 走不同路徑**——只看 opcode 不看 code 會漏掉
一半。

## opcode `0Dh`（`0858h`）

```text
saved := DS:A2A8h
if bank1^[582h] > 0 then
    bank1^[582h] := bank1^[582h] - 1
    <far 0062:0048>(DS:BDDAh, bank1^[582h], DS:A894h, DS:A893h)
ECL_PC := ECL_PC + 1                      ← 沒有 operand
DS:A2A8h := saved                         ← 原樣還原
```

`DS:A2A8h` 在進入時存起來、離開時原樣寫回——被呼叫的常式會改動它。

## opcode `0Fh`（`0992h`）

```text
READVAR(2)
addr := ADDFNC(high[2], low[2])
r := <far 0418:1259>(0, 0Ah, @local)
<far 0062:0069>(r, addr)
<far 0064:003E>()
```

## 明確不宣稱

- opcode `09h`／`0Dh`／`0Fh` 在 ECL 指令集裡叫什麼。
- 誰把節點推進 GOSUB 堆疊（`ADDEFFECT` 那種 `New(6)` 的對應寫入端未讀）。
- `0418:1259`／`0062:0048`／`0062:0069`／`0062:006B`／`0062:0075` 的本體。
- `DS:A2A8h`／`DS:BDDAh`／bank1 `+582h` 的語意。
