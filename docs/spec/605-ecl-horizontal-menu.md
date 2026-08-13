# 第六百零五輪：`2Bh`（HORIZONTAL MENU）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:10C2h`（466 bytes，180 條指令）。

## 變長指令的實作對上了

```text
<初始化>()
READVAR(2)                                ← 固定前綴
dest := ADDFNC(high[1], low[1])
n    := ADDRESSVALUE(2)                   ← 選項數
ECL_PC := ECL_PC − 1                      ← 回退一格
READVAR(n)                                ← 再讀 n 個 operand
```

與 [spec 564](564-ecl-operand-decoding-and-arity-validation.md) 由 `READVAR`
參數推出的變長模式**完全一致**——那是從解碼器反推的，這裡是 handler 本體的
直接證據。

## `n = 1` 是特例

只有一個選項時不畫選單，改顯示
**「リターン・キーを押してください」**（`unk_10A3h`，「請按 Enter 鍵」），
樣式參數也不同（`0` 而不是 `0Fh`／`0Ah`）。

## 選項文字與熱鍵

```text
for i := 1 to n do
    <解 packed text>(1, DS:[A8DAh + i*100h], @buf)     ← 每個選項一個 256 bytes 槽位
for i := 1 to n do
    Move(DS:[A8DBh + i*100h], DS:[A339h + i*7], DS:[A338h + i*7])
    DS:[A334h + i] := i + 40h                          ← 熱鍵
```

- **選項文字陣列在 `DS:A8DAh`，stride `100h`（256 bytes）**，索引從 1 起算。
- **熱鍵是 `i + 40h`**：第 1 個選項是 `41h`＝`'A'`，第 2 個 `'B'`……
- `DS:A339h` 起是 stride 7 的短標籤陣列，長度取自 `DS:A338h + i*7`。

## 明確不宣稱

- `0062:0084`（選單本體）的按鍵處理與回傳值範圍。
- `DS:BDF4h`／`BDF5h` 如何決定樣式參數 `var_3C`。
- 收尾的 `0418:14C7(18h, 27h, 18h, 0)` 是清除哪一塊。
