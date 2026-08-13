# 第六百一十六輪：ECL 的兩種字串來源

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`ECL2`（overlay-07）。位址為 PC-98 overlay-local。

ECL 取得字串有兩條完全不同的路徑，這一輪把兩條都讀出來了。

## `FINDSTR(slot, n)`（`1364h`）：從 bytecode 逐 byte 讀

```text
DS:[A8DAh + slot*100h] := ''              ← 清空槽位
for i := 1 to n do
    ECL_PC := ECL_PC + 1
    DS:[A8DAh + slot*100h] :=
        DS:[A8DAh + slot*100h] + Chr(script[ECL_PC])
```

**字面字串直接嵌在 ECL bytecode 裡**，長度由呼叫端的 `n` 給出，內容是 `n` 個
連續 byte。每讀一個字元 `ECL_PC` 就前進一格，所以讀完之後 PC 正好停在字串
之後。

這是 packed text（operand code `80h`）之外的**第二種文字來源**，而且更直接——
沒有壓縮、沒有查表。中文化要處理的字面文字有一部分在這裡。

## `GETSTR(?, addr, slot)`（`13F7h`）：從 ECL 記憶體讀

```text
bank := <位址分類器>(addr)                ← overlay-07:0801h，spec 563
DS:[A8DAh + slot*100h] := ''
case bank of
    0: base := DS:7F05h ; off := 6A00h + i*2      ← 取 byte，走 word 陣列
    1: base := DS:7F09h ; off := 0800h + i*2
    2: base := DS:7F0Dh ; off := 0C00h + i*2
    3: base := DS:7F12h ; off := i                ← byte 陣列
while <該位置的值> <> 0 do
    s := s + Chr(<該位置的 byte>) ; i := i + 1
```

四個 bank 的基底與位移**與 [spec 563](563-ecl-memory-model-and-operand-resolution.md)
解出的記憶體模型完全一致**——那份是從讀寫路徑推的，這裡是字串路徑，兩個獨立
來源對上。

**字串以 0 結尾**（不是長度前綴），與 Turbo Pascal 自己的短字串相反。

### `addr = 7C00h` 是特例

bank 1 的分支裡，`addr` 剛好等於 `7C00h`（bank 1 的起始位址）時**不走陣列**，
改從 **`DS:9594h`（目前目標）的記錄**直接取字串。形狀是「角色自己的名字」。

remake 若把 `7C00h` 當成一般的 bank 1 位址讀，會拿到陣列第 0 項而不是目標的
名字。

## 兩種來源的共同去處

兩支都寫進 **`DS:A8DAh` 起、stride `100h` 的槽位**——與 `2Bh`（HORIZONTAL MENU，
[spec 605](605-ecl-horizontal-menu.md)）和 `29h`（PARLAY，
[spec 611](611-ecl-parlay.md)）用的是同一組緩衝區。所以「讀字串」與「顯示選項」
共用槽位，指令之間有隱含的先後依賴。

## 明確不宣稱

- `GETSTR` 第一個參數的用途（讀出來只用於決定 bank）。
- bank 4（具名特例）沒有對應分支——傳 bank 4 的位址會發生什麼未讀。
