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

## opcode `00h`（`0052h`）：結束 script

`RETURN` 在堆疊空時轉過來的就是這一支，理由在它的內容裡——**它會把整個
GOSUB 堆疊清空**。

```text
if DS:BDF6h <> 0 then <far 019E:014A>()
if DS:7898h <> 0 then
    DS:9594h := DS:789Dh                  ← 換一個 far pointer
    DS:7898h := 0
FillChar(DS:BDDAh, 2, 0)
DS:BDF4h := 0 ; DS:BDF6h := 0 ; DS:7896h := 1
ECL_PC := ECL_PC + 1
while DS:A882h <> nil do                  ← 逐一釋放整條鏈
    next := node^[2] ; Dispose(node, 6) ; DS:A882h := next
DS:9637h := 11h                           ← 訊息停留時間設成 11h
DS:9636h := 1
DS:BDF2h := 0
```

`DS:9637h` 是 `PUTDAMAGE` 的訊息停留時間全域
（[spec 581](581-putdamage-pipeline.md)），這裡設成 `11h`——而 opcode `33h`
是把它加一（[spec 586](586-ecl-handlers-31-33-34.md)）。所以 script 結束時
會把停留時間**重設**成固定值，不是累加。

## opcode `02h`（`0107h`）＝ `GOSUB`

```text
READVAR(1)
<far 0062:0098>(1)
```

handler 本身只有 11 條指令——**推堆疊的動作在 `ECL2` 裡**（`0062:0098` 是
`ECL2` 控制區塊的 stub offset），對應 Borland 符號表裡的
`SETUPGOSUBSTACK`。所以 `13h`（`RETURN`）彈出、`ECL2` 推入，兩端分屬不同
unit。

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
