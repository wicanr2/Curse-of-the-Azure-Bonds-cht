# 第六百零二輪：`1Dh`（全隊戰力評分）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:12F7h`（325 bytes）。

```text
READVAR(1)
total := 0                                    ← byte
node := DS:9598h
while node <> nil do
    hp := node^[1A5h]                         ← 目前 HP
    a  := node^[19Bh] ; if a > 3Ch then a := a - 3Ch else a := 0
    b  := node^[19Ah] ; if b > 27h then b := b - 27h else b := 0
    lvl := <far 015D:0052>(node)
    c  := lvl * node^[116h] + node^[10Eh]
    d  := lvl * node^[111h] + node^[109h]
    total := total + (hp + a*5 + b*5 + c*8 + d*4) div 10
    node := node^[18Ah]
STOREVALUE(ADDFNC(operand 1), total)
```

## 公式的四個細節

1. **`+19Bh` 以 `3Ch`（60）為門檻、`+19Ah` 以 `27h`（39）為門檻**，低於門檻
   一律算 0，不會變成負值。`+19Ah` 是命中修正
   （[spec 577](577-attempttohit-and-effect-chain-walk.md)）。
2. `c` 與 `d` 各自是「等級 × 某欄位 ＋ 另一欄位」，權重 `8` 與 `4`（用
   `shl 3`／`shl 2`）。
3. **除以 10 是有號除法**（`cwd` ＋ `idiv`），且**每名角色各自除完才累加**——
   不是先加總再除，所以每人都有一次無條件捨去。
4. **`total` 是 byte**（`add [bp+var_6], al`），**超過 255 會環繞**。人多或
   等級高時這個值會繞回小數字。

`lvl` 由同一支常式（`015D:0052`）算兩次，兩次之間沒有任何狀態改變——與
`06h` 把 `div` 執行兩次、字串比較跑六次
（[spec 592](592-ecl-arithmetic-and-compare.md)、
[593](593-ecl-comparison-flags.md)）是同一種風格。

## 明確不宣稱

- `+19Bh`、`+109h`、`+10Eh`、`+111h`、`+116h` 各是什麼欄位。
- `015D:0052` 回傳的是不是等級（只確定它是 `char` 的函數、以 byte 參與乘法）。
- 這個分數之後被誰用。
