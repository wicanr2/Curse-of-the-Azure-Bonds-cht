# 第五百九十八輪：`36h`（ADD NPC）與 `+0F7h` 的 NPC 標記

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:2F5Fh`（123 bytes）。

```text
if bank1^[67Ch] > 7 then return               ← 名額已滿就整條指令不做事
READVAR(2)
a := ADDRESSVALUE(1)
<far 00DC:004D>(a)
b := ADDRESSVALUE(2)
b := (b div 2) or 80h                         ← 有號除法，再把最高位設起來
DS:9594h^[0F7h] := b
<far 014A:0043>(DS:9594h) ; <far 014A:002A>(DS:9594h)
```

## 三件事

1. **上限是 7**（`bank1^[67Ch] > 7` 就直接返回）。不是「加不進去回傳失敗」，
   而是**整條 ECL 指令什麼都不做**，也不設任何旗標——呼叫端無從得知。
2. **`+0F7h` 的最高位是 NPC 標記。** 寫進去的值一定 `or 80h`。這解釋了
   `3Eh`（[spec 596](596-ecl-party-item-sweep.md)）掃描可行動成員時要求
   `node^[0F7h] = 0`——那是在排除 NPC；也解釋了 `REMOVEFX`
   （[spec 576](576-adnd-strength-encoding-and-effect-removal.md)）檢查
   `char^[0F7h] = 0B3h`（`10110011b`，最高位為 1）。
3. `b div 2` 是**有號**除法（`cwd` ＋ `idiv`），與四則運算的 `06h` 用無號
   `div`（[spec 592](592-ecl-arithmetic-and-compare.md)）不同。同一個直譯器
   裡兩種都有，不能一概而論。

## 再次確認與 `SKIP` 的不一致

這支的第一個動作是 `READVAR(2)`，而 `ECL2` 的 `SKIP` 表把 `36h` 歸為 arity 1
（[spec 591](591-skip-arity-crosscheck.md)）。**兩邊的證據都是直接讀出來的**，
不是解析推論——`36h` 的不一致確定成立。

## 明確不宣稱

- bank 1 `+67Ch` 的確切語意（隊伍人數？NPC 人數？）。
- `00DC:004D` 與 `+0F7h` 低 7 位的意義。
