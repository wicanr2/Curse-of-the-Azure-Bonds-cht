# 第五百九十四輪：ECL 的 `RANDOM` 與索引寫入

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`INTERPET`（overlay-02）。位址為 PC-98 overlay-local。

## `08h`＝`RANDOM`（`025Ah`）

```text
READVAR(2)
n := ADDRESSVALUE(1)                      ← byte
if n < FFh then n := n + 1
dest := ADDFNC(high[2], low[2])
r := Random(n)                            ← RTL 的 @Random（0A65:1155）
STOREVALUE(dest, r)
```

**`n` 會先加一**（除非已經是 `FFh`），所以 ECL 寫 `RANDOM(6)` 得到的是
`Random(7)`＝`0..6`，**含上界**。少了這個加一，remake 的骰子範圍會少一格。

`Random` 就是 [spec 575](575-random-core-and-pc98-vram.md) 解出的那支：
`(RandSeed shr 16) mod n`。

## `35h`：目的位址是算出來的

```text
READVAR(3)
value := ADDRESSVALUE(1)
base  := ADDFNC(high[2], low[2])
index := ADDRESSVALUE(3)
STOREVALUE(base + index, value)
```

**第一個 operand 是值，第二＋第三個相加才是目的位址**——這是陣列寫入
（`ECL[base + index] := value`）。

⚠ 參數順序很容易寫反。`STOREVALUE` 的呼叫慣例是 `(位址, 值)`，而 `35h` 在
push 時第一個推的是**和**（位址）、第二個是 `ADDRESSVALUE(1)`（值）；`04h`
系列則是第一個推 `ADDFNC(operand 3)`（位址）、第二個推運算結果（值）。
兩者的 operand 編號分工不同，但 `STOREVALUE` 的參數順序一致。

## 明確不宣稱

- `0A65:0CBEh`／`0A65:0CC2h` 這兩個緊接在 `Random` 之後的 RTL 呼叫做什麼
  （型別轉換？）——只確定結果最後被當成 byte 存進去。
