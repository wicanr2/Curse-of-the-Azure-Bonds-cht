# 1047 — 資料檔的分塊讀取：逐塊掃過去找 id，超過 36000 bytes 就中止

- 證據等級：`exact`（DOS 常駐段 235 條逐條讀完，`objdump` 路線見 spec 1046）
- 作法見 spec 783

## `dos START.EXE:16C3Eh`（`retf 12h`，559 bytes）

原本待解讀。呼叫端 `12AE1h`、`155D5h`。

```pascal
function 讀一塊(檔名A: 字串;      { bp+14h }
                檔名B: 字串;      { bp+10h }
                要的id: byte;     { bp+0Eh }
                長度出: 遠指標;   { bp+0Ah }
                指標出: 遠指標)   { bp+06h }
```

`retf 12h` ＝ 18 bytes ＝ 4 ＋ 4 ＋ 2 ＋ 4 ＋ 4，與五個參數對得上。

## 流程

```pascal
A := Copy(檔名A, 1, 50h);  B := Copy(檔名B, 1, 50h);   { 各上限 80 }
if not <sub_72D4h>(A, B, …, @檔) then begin            { 開不起來 }
    長度出^ := 0;  指標出^ := NIL;  exit(false)
end;
repeat
    BlockRead(檔, id,   1);        { ★ 1 byte：這一塊的 id }
    BlockRead(檔, 位移, 4);        { ★ 4 bytes：longint }
    BlockRead(檔, 長度出^, 2);     { ★ 2 bytes }
    BlockRead(檔, 大小,  2);       { ★ 2 bytes }
    if 大小 > 8CA0h then …錯誤，中止…;
    if id = 要的id then break;
    if FilePos ≥ FileSize then break;              { 掃完了 }
until false;
if id <> 要的id then begin
    長度出^ := 0;  指標出^ := NIL;  <sub_73C7h>(…);  exit(false)
end;
DS:5CF2h := NIL;
指標出^ := GetMem(長度出^);                        { ★ 配置 }
Seek(檔, 位置 ＋ 位移);
BlockRead(檔, DS:5CF2h^, …);
…<sub_7212h>(DS:5CF2h, 指標出^, @大小, 長度出^)…   { 形狀上是解壓 }
```

> ★★ **一塊的表頭是 9 bytes**：`id`(1) ＋ `位移`(4) ＋ 一個 word ＋ `大小`(2)。
> 讀不到要的 id 就往下一塊掃，掃到檔尾為止。
> ★★★ **`8CA0h` ＝ 36000 是硬上限**：超過就組錯誤訊息（`Str(大小, 6)` 接在後面）
> 並叫 `far 06EA:0000h` 中止。
> ⇒ **remake 的資料塊不能超過 36000 bytes**，否則原版讀不進去。

★ `DS:5CF2h` 是讀進來的原始塊指標，剛好落在物品類別表 `DS:5CF6h` 前面 4 bytes。
★ 最後把 `DS:5CF2h` 釋放掉（`sub_6B10h`），只留 `指標出^`
——**原始塊是暫存的，交出去的是處理過的結果**。

## ⚠ 常駐段的 objdump 路線有一個邊界

錯誤訊息的來源是 `push cs; mov di, 08D4h`，也就是**目前程式碼段的 `08D4h`**。
把它當成「IDA 線性 `108D4h`」去換算檔案位移，讀到的是 overlay stub 表（`CD 3F`），
不是字串。

> ⚠ **常駐段不是單一程式碼段**，`push cs` 的 CS 值要看 IDA 的段表才知道。
> `objdump` 路線可以逐條讀指令（spec 1046 已驗證進入點換算正確），
> 但**函式內的 `CS:` 資料參照沒有段表就解不出來**。
> 本規格因此不宣稱那句錯誤訊息的內容。

## 中文化

錯誤訊息（`CS:08D4h` ＋ `Str(大小, 6)`）沒讀到，見上。
兩個檔名參數各上限 80 bytes。

## 明確不宣稱

- 沒有宣稱 `CS:08D4h` 那句錯誤訊息的文字（理由見上）。
- 沒有宣稱 `sub_72D4h`（開檔）、`sub_73C7h`（關檔）、`sub_7212h`（處理／解壓）、
  `sub_6B2Dh`／`sub_6B10h`（配置／釋放）的介面。
- 沒有宣稱塊表頭第 6..7 個 byte（寫進 `長度出^` 的那個 word）與 `大小` 的關係。
- 沒有宣稱這裡讀的是不是 `.dax`（spec 892／1029／1033 的檔名組法）。
- 沒有宣稱兩個檔名參數為什麼要兩個。
