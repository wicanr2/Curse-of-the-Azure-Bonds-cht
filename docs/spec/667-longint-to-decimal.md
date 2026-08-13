# 第六百六十七輪：32-bit 整數轉十進位字串

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `1B839h`、`1B756h`、`1AC19h`。

## `1B839h`：減法法轉十進位

```text
bx := AX（低位）, dx（高位）
if dx < 0 then
    (bx, dx) := −(bx, dx)          ← not/not/add 1/adc 0
    輸出 '-'（2Dh）
si := 123Bh                         ← 十的冪次表
cl := 9
重複:                                ← 跳過前導零
    if (dx:bx) >= cs:[si] 則跳出
    si := si + 4 ; cl := cl − 1
cl := cl + 1
重複 cl 次:
    al := '0'
    重複: al := al + 1 ; (dx:bx) := (dx:bx) − cs:[si] until 借位
    (dx:bx) := (dx:bx) + cs:[si]     ← 借過頭，加回來
    si := si + 4
    輸出 al
```

不用除法，**用「一直減到借位、再加回來」數出每一位**——16-bit 機器上比 32-bit
除法便宜。

`mov al, 2Fh` ＋ `inc al` 得到 `'0'`（`30h`）：先放 `'0' − 1` 再進迴圈加一，這樣
「至少加一次」就自然成立。

### 冪次表就在函式後面

`seg052` 的基底是 `1A650h`（該段第一支 `sub_1A650`），所以 `CS:123Bh` ＝ 線性
`1B88Bh`——**正好是這支函式的結尾**（`1B839h + 82 = 1B88Bh`），表緊接在程式碼後面。

實際讀出的十筆 32-bit 值：

| 索引 | 值 | |
|---|---|---|
| 0 | 1000000000 | 10⁹ |
| 1 | 100000000 | 10⁸ |
| … | … | … |
| 9 | 1 | 10⁰ |

**十筆全部吻合**，所以這支確定是 longint → 十進位字串，不是別的進位制。

負數處理是「取二補數後輸出 `'-'`」，所以 `−2147483648` 取補數後仍是自己——
這支對那個值會輸出錯誤結果（原作行為，remake 照抄時要知道）。

## `1B756h`：走 6-byte 一筆的表

```text
存下 (AX, BX, DX) 三個 word
(AX, BX, DX) := cs:[di], cs:[di+2], cs:[di+4]
重複 cx 次:
    (CX, SI, DI) := cs:[di], cs:[di+2], cs:[di+4]
    <sub_1AFE4h>()
    (CX, SI, DI) := 存下的三個 word
    <sub_1B0A7h>()
    di := di + 6
(CX, SI, DI) := (81h, 0, 0)
<sub_1AFE4h>()
```

**三個暫存器一組、每筆 6 bytes**——Turbo Pascal 的 `real` 就是 6 bytes，而
`CX:SI:DI` 正是它在暫存器裡的傳遞方式（[spec 622](622-character-money-block.md)
的 `ROBDOUGH` 用 `arg_4`／`arg_6`／`arg_8` 傳同樣的東西）。

收尾用 `(81h, 0, 0)` 呼叫一次——`real` 的第一個 byte 是指數，`81h` 對應
`2^(81h−81h)` ＝ 1.0 的指數位。

## `1AC19h`：正規化並夾住下限

```text
di := low(dword_23AF0h)
if di = 0 then
    di := di − word_23AF4h
    if 結果為 0 then
        si := 0
        di := high(dword_23AF0h) + 1000h
        return
else
    di := di − word_23AF4h
    if 借位 then di := 0
si := di and 0Fh                      ← 段內偏移
di := (di shr 4) + high(dword_23AF0h) ← 段落
if di > word_23AEEh then return
if di = word_23AEEh and si >= word_23AECh then return
(si, di) := (word_23AECh, word_23AEEh) ← 夾到下限
```

比較是**先比段落、相等才比偏移**的兩段式比較。夾住的下限是
`(word_23AECh, word_23AEEh)`，與 [spec 666](666-ems-page-frame-guard.md) 的
`1A9E1h`／`1ABE0h` 用的是同一組界線。

## 明確不宣稱

- `sub_1AFE4h`／`sub_1B0A7h` 各自對 6-byte real 做什麼運算。
- `word_23AF4h` 的角色與 `1000h` 這個常數的來歷。
