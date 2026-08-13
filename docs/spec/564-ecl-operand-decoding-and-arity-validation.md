# 第五百六十四輪：`READVAR` operand 解碼與 arity 驗證

狀態：`READY`（限 operand 編碼格式、`READVAR` 語意與本輪的 arity 對照結果）。
日期：2026-08-14

## 結論先行

`READVAR(n)` 是 ECL 的 **operand 解碼器**：它從 ECL PC 往後解 `n` 個 operand
進第 563 輪記錄的三個平行陣列。因此——

**`internal/ecl` 宣稱的 arity 與原版一致：64 個 opcode 中 62 個完全吻合**
（PC-98；DOS 60 個，差的兩個是 IDA 函式邊界問題，見下）。不吻合的 4 個全是
已知的變長指令，而且本輪連它們的**固定前綴長度**也解出來了。

先前第 562 輪用 `ADDRESSVALUE` 的呼叫次數當訊號，只有 35/64 吻合。那個訊號
本身是錯的：`ADDRESSVALUE(i)` 是「取用第 i 個已解好的 operand」，同一個
operand 可以被取用零次或多次，與 arity 無關。該輪的表已由本輪取代。

## `READVAR` 的行為（PC-98 `overlay-07:008Eh`／DOS `0034h`）

```text
READVAR(n):
  for i = 1 .. n:                    ← 索引從 1 開始
      code[i] = script[PC + 1]
      low[i]  = script[PC + 2]
      PC += 2
      if code[i] in {1, 2, 3}:
          PC += 1
          high[i] = script[PC]
      if code[i] == 80h:             ← packed text，長度可變
          （另一條路徑）
```

- ECL PC 是 `DS:7F21h`（兩平台同一個角色，PC-98 實測位址）。
- script 緩衝區就是第 563 輪的 bank 3：`les di, ds:7F12h` 後以 `-8000h` 取址。
- **operand 陣列索引從 1 開始**。這解釋了為什麼 `CALL`（arity 1）的 handler
  直接讀 `hi[1]`／`lo[1]`（PC-98 `DS:A998h`／`DS:A958h`）。
- operand 的位元組佈局是 `[code][low]`，code 為 `1`／`2`／`3` 時再加一個
  `[high]`。這與既有 parser 的 framing 一致。

## arity 對照結果

產物：[`docs/audit/ecl-handler-operand-audit.md`](../audit/ecl-handler-operand-audit.md)
（由 `scripts/ecl_handler_operand_audit.py` 產生）。

| 平台 | 吻合 | 不吻合 |
|---|---:|---|
| PC-98 | 62／64 | `15h`、`2Bh` |
| DOS | 60／64 | `15h`、`25h`、`26h`、`2Bh` |

arity 0 的 16 個指令完全不呼叫 `READVAR`——這也算吻合，不是缺漏。

### 四個變長指令的固定前綴

| opcode | 指令 | 固定前綴 | 備註 |
|---|---|---:|---|
| `15h` | VERTICAL MENU | 3 | 兩平台一致 |
| `25h`／`26h` | ON GOTO／ON GOSUB | 2 | 共用 handler |
| `2Bh` | HORIZONTAL MENU | 2 | 兩平台一致 |

變長的作法是固定的：

```text
READVAR(固定前綴)
  → ADDRESSVALUE(某個 operand) 取得後續數量 N
  → dec PC          ← 回退一格
  → READVAR(N)
```

`KnownCommands` 把這四個宣告成 `0` 並由 parser 特別處理，因此**不是錯誤**；
但固定前綴長度是原版事實，remake 的 parser 應該對得上這三個數字。

### 兩平台差異的成因

`25h`／`26h` 在 PC-98 側顯示吻合、DOS 側顯示不吻合，看起來像平台差異，其實
不是：PC-98 的該 handler 在 IDA 裡只有 **3 bytes**（函式邊界建錯），所以掃不到
它的 `READVAR` 呼叫，是假陰性。DOS 那份是 149 bytes，結果可信。

**規則：拿 IDA 的函式邊界當範圍時，先看 size 是否合理。** 個位數 bytes 的
「函式」幾乎一定是邊界建錯，不是真的短函式。第 563 輪的 DOS `STOREVALUE`
也踩過同一個坑。

## 工具在本輪修掉的兩個錯誤

兩個都會讓表看起來很有結論但其實是錯的：

- **呼叫歸屬用區間猜測**。原本用「排序後取前一個 handler 起點」把呼叫歸給
  handler，但 handler 在位址上並不連續，後面函式的呼叫會被算到前一個頭上
  （讓 `3Dh`／`3Eh` 這種 arity 0 的指令冒出 `READVAR`）。改用 IDA 的實際
  函式 chunk 範圍。
- **只比對 stub offset、沒比對 segment**。`014A:002A`（overlay-24）被當成
  `0062:002A`（ECL2 的 `READVAR`）。這正是「相同數字不代表相同物件」，
  修正後 `3Dh`／`3Eh`／`24h` 三個假不吻合全部消失。

## 這份規格明確不宣稱

- **`code == 80h`（packed text）的完整長度規則**。`READVAR` 對它另走一條路徑，
  本輪只確認存在，未讀完。
- **`ADDRESSVALUE` 對 `01h/03h/80h` 呼叫的記憶體讀取 routine**（PC-98
  `sub_FFC`）的 bank 路由是否與 `STOREVALUE` 對稱。
- **變長指令第二段 `READVAR(N)` 的 N 來自哪一個 operand**。已確認 N 來自
  `ADDRESSVALUE` 的回傳值，但是第幾個 operand、以及 menu 項目的後續格式，
  尚未逐一讀完。
- **`KnownCommands` 的指令名稱**。本輪驗證的是 operand 數量，不是語意；
  `DUMP`、`CLEAR BOX` 這些名字仍是既有文件的說法。
