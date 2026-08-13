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

## `2Ah`：索引讀取，與 `35h` 對稱

```text
READVAR(3)
base  := ADDFNC(high[1], low[1])
index := ADDRESSVALUE(2)
dest  := ADDFNC(high[2], low[2])          ← 同一個 operand 再取一次 raw word
v := <far 0062:0070>(base + index)        ← 讀出來
STOREVALUE(dest, v)
```

`ECL[dest] := ECL[base + index]`——`35h` 是索引寫入，這支是索引讀取。

operand 2 被用了兩次：一次經 `ADDRESSVALUE` 取值當索引，一次取原始 word 當
目的位址。這印證了「`ADDRESSVALUE` 的呼叫次數與 arity 無關」
（[spec 564](564-ecl-operand-decoding-and-arity-validation.md)）——**要驗
arity 一律看 `READVAR` 的參數**。

宣稱的 arity 是 3，但**第三個 operand 在 body 裡沒有被用到**。

## `3Dh`（`2E8Ch`）：無 operand 的重繪

```text
ECL_PC := ECL_PC + 1
<far 019E:014A>() ; <far 019E:0268>()
<far 014A:002A>(DS:9594h) ; <far 014A:00DE>()
<far 0176:0025>(DS:A2CDh, 3, 3, 3)
<far 014A:00DE>()
DS:BE00h := 0 ; DS:BDF2h := 1
```

## 明確不宣稱

- `0A65:0CBEh`／`0A65:0CC2h` 這兩個緊接在 `Random` 之後的 RTL 呼叫做什麼
  （型別轉換？）——只確定結果最後被當成 byte 存進去。
