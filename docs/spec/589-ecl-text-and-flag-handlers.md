# 第五百八十九輪：ECL 的文字輸出與六個旗標指令

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`INTERPET`（overlay-02）。位址為 PC-98 overlay-local。

## `16h`～`1Bh`：六個 opcode、六個旗標、一支 handler

`0BAEh`（120 bytes）被六個 opcode 共用，結構完全對稱：

```text
ECL_PC := ECL_PC + 1                      ← 沒有 operand
op := DS:A891h                            ← 再讀一次 dispatcher 的 opcode
if DS:[A88Ah + (op - 16h)] = 0 then <far 0062:00B1>()
```

| opcode | 旗標位址 |
|---:|---|
| `16h` | `DS:A88Ah` |
| `17h` | `DS:A88Bh` |
| `18h` | `DS:A88Ch` |
| `19h` | `DS:A88Dh` |
| `1Ah` | `DS:A88Eh` |
| `1Bh` | `DS:A88Fh` |

六個分支呼叫的是**同一支**常式（`0062:00B1`），差別只在檢查哪一個旗標。
`DS:A88Ah..A88Fh` 是六個連續的旗標位元組。

分辨方式與 `21h`／`37h` 一樣——重讀 `DS:A891h`
（[spec 587](587-ecl-handler-21-37-shared.md)）。

## `11h`／`12h`：顯示文字

`0A5Dh`（156 bytes）：

```text
READVAR(1)
DS:BDF8h := 0
DS:7F36h := 1
if operand_code[1] < 80h then                    ← 不是 packed text
    <far 014A:004D>(ADDRESSVALUE(1), @buf)       ← 取出字串
    StrCopy(DS:A9DAh, buf, FFh)                  ← 放進共用文字緩衝區
if DS:A891h = 11h then
    <far 0418:0E6A>(1, 11h, 26h, 16h, 0Ah, 0, 0, DS:A9DAh)
else                                             ← 12h
    DS:9637h := 11h ; DS:9636h := 1              ← 先重設訊息停留時間
    <far 0418:0E6A>(1, 11h, 26h, 16h, 0Ah, 0, 1, DS:A9DAh)
DS:7F36h := 0
```

兩者的差別只有兩點：**倒數第二個參數 `0` 對 `1`**，以及 `12h` 會先把
`DS:9637h` 重設為 `11h`。其餘完全相同。

`DS:A9DAh` 是共用的文字緩衝區，opcode `09h`
（[spec 588](588-ecl-return-and-gosub-stack.md)）在 packed text 分支傳的也是
它。**operand code `>= 80h` 時不做取字串與複製**——因為 packed text 早就被
解到那個緩衝區裡了。

`(1, 11h, 26h, 16h, 0Ah)` 這組固定值在兩個分支相同，形狀像視窗的邊界與樣式。

## 明確不宣稱

- `16h`～`1Bh` 那六個旗標各代表什麼、`0062:00B1` 做什麼。
- `0418:0E6A` 的參數意義（座標？樣式？）。
- `DS:BDF8h`／`DS:7F36h` 的語意。
