# 第六百一十八輪：`INITECL` 與 ECL script 的檔頭

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-07:0317h`（362 bytes，134 條指令）。

## script 的前五個 operand 就是五個進入點

```text
INITECL:
    <清一批旗標>：DS:BDF4h, BDF6h, BDE4h, BDE5h := 0
    DS:A87Ah := 41h ('A') ; DS:A87Bh := 9
    DS:BDF9h := 1 ; FillChar(DS:BDDAh, 2, 0) ; DS:A895h := 8
    ECL_PC := 8000h                           ← bank 3 的起點
    DS:A9D8h := 0
    DS:A882h := nil ; DS:A886h := nil         ← 清空 GOSUB 堆疊
    FillChar(DS:A88Ah, 6, 0)                  ← 清空六個比較旗標
    bank1^[5C2h] := FFh
    bank1^[5A4h] := 0 ; bank1^[5A6h] := 0 ; bank0^[1CAh] := 0
    for i := 0 to 4 do
        READVAR(1)
        DS:[7F17h + i*2] := ADDFNC(high[1], low[1])
    <清 bank0 的 6A00h 起 20h 項>
    <清 bank1 的 800h + 7F79h*2 起 9 項>
    DS:7F11h := 0
```

**`ECL_PC` 從 `8000h` 開始，正好是 bank 3（script 緩衝區）的起點**
（[spec 563](563-ecl-memory-model-and-operand-resolution.md)），然後**連續讀
五個 operand**，依序存進 `DS:7F17h`／`7F19h`／`7F1Bh`／`7F1Dh`／`7F1Fh`。

所以 **ECL script 的檔頭格式是：五個 operand，各自編碼一個進入點位址**。
`READVAR(1)` 每次讀一組 `[code][low]([high])`，所以檔頭的長度取決於這五個
operand 的 code——**不是固定 10 bytes**。

這與 [spec 612](612-ecl-main-loop.md)／[613](613-ecl-scene-lifecycle.md) 讀出
的執行順序合起來，script 的生命週期就完整了：

| 進入點 | 存放 | 何時執行 |
|---:|---|---|
| 第 1 個 | `DS:7F17h` | 場景主迴圈每輪 |
| 第 2 個 | `DS:7F19h` | 地圖座標變動時 |
| 第 3 個 | `DS:7F1Bh` | `3237h` 第一段 |
| 第 4 個 | `DS:7F1Dh` | `3237h` 第二段（條件成立才執行） |
| 第 5 個 | `DS:7F1Fh` | 進場一次 |

**進入點在 script 裡的順序，與執行順序不同**——第 5 個最先跑。

## 初始化順帶確定的事

- **GOSUB 堆疊與六個比較旗標在 `INITECL` 就被清空**，所以每個 script 開始時
  堆疊必為空、六個旗標必為 0。`13h`（`RETURN`）在 script 第一條就執行的話，
  會直接走 `00h`（[spec 588](588-ecl-return-and-gosub-stack.md)）。
- `bank1^[5C2h] := FFh`——`0Eh`（顯示圖片，[spec 600](600-ecl-picture.md)）用
  它分支，初值 `FFh` 表示「走載入圖片那一路」。
- `DS:A87Ah := 41h`（`'A'`）與 `2Bh` 選單熱鍵從 `'A'` 起
  （[spec 605](605-ecl-horizontal-menu.md)）一致。

## 明確不宣稱

- `DS:A87Bh := 9`、`DS:A895h := 8`、`DS:BDF9h`、`DS:A9D8h`、`DS:7F11h` 的語意。
- 收尾清掉的兩塊（bank0 `6A00h` 起 `20h` 項、bank1 `800h + 7F79h*2` 起 9 項）
  是什麼。
