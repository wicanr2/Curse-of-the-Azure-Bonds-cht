# 第五百九十三輪：ECL 的六個比較旗標

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-07:19FEh`（`ECL2` stub `0093h`），數值比較。

## 六個條件全部解出

```text
數值比較(a, b):                            ← COMPARE 傳入時 a = v2、b = v1
    FillChar(DS:A88Ah, 6, 0)               ← 先把六個旗標清零
    if v1 =  v2 then DS:A88Ah := 1
    if v1 <> v2 then DS:A88Bh := 1
    if v1 <  v2 then DS:A88Ch := 1         ← 無號比較（jnb）
    if v1 >  v2 then DS:A88Dh := 1         ← 無號（jbe）
    if v1 <= v2 then DS:A88Eh := 1         ← 無號（ja）
    if v1 >= v2 then DS:A88Fh := 1         ← 無號（jb）
```

配上 `16h`～`1Bh`「旗標為 0 就跳過下一條指令」
（[spec 591](591-skip-arity-crosscheck.md)），ECL 的條件式完整了：

| opcode | 旗標 | 條件 |
|---:|---|---|
| `16h` | `DS:A88Ah` | `v1 = v2` |
| `17h` | `DS:A88Bh` | `v1 <> v2` |
| `18h` | `DS:A88Ch` | `v1 < v2` |
| `19h` | `DS:A88Dh` | `v1 > v2` |
| `1Ah` | `DS:A88Eh` | `v1 <= v2` |
| `1Bh` | `DS:A88Fh` | `v1 >= v2` |

**六個比較全是無號的**（`jnb`／`jbe`／`ja`／`jb`）。ECL 的值是 16 位元，
`FFFFh` 在這裡比 `0001h` 大，不是 `-1`。remake 用有號比較會在跨越
`8000h` 的值上得到相反的結果。

比較常式**進來就把六個旗標一起清零**，所以每次 `COMPARE` 都是乾淨的起點；
六個旗標同時被設定，`COMPARE` 一次就決定了後續六種分支全部的答案。

## 與 `14h` 的分工

`14h`（[spec 590](590-ecl-control-flow-if-and-branch.md)）也清六個旗標，但
**只設 `A88Ah`／`A88Bh`**：它比較的是**兩對** word（`v1:v2` 與 `v3:v4`），
兩對都相等才設 `A88Ah`，否則設 `A88Bh`。

所以 `18h`～`1Bh`（大小比較）**只有在 `03h`（`COMPARE`）之後才有意義**；
接在 `14h` 之後的話那四個旗標永遠是 0，也就是永遠跳過。這解答了 spec 590
留下的「`A88Ch`..`A88Fh` 由誰設定」。

## 明確不宣稱

- 字串比較（`0062:008E`，PC-98 `overlay-07:1928h`）是不是也設同一組旗標。
