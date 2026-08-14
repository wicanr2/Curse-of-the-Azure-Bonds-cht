# 944 — PC-98 常駐的 RTL 五支：EMS 覆疊、檔案 I/O、文字視窗初始化

- 證據等級：`exact`（五支逐條讀完，來源是 `PC98-GAME.EXE.asm` 全文列表）
- 位置：`PC98-GAME.EXE` 的 `1A410h`、`1A4DAh`、`1BEE8h`、`1BFD3h`、`18FF5h`
- 作法見 spec 783

## `1A410h`（54 行，far）：`OvrInitEMS`

```pascal
if word_23AE4h = 0 then 結果 := −1                    { 覆疊檔沒開 }
else if not <sub_1A485h>() then 結果 := −5            { 0FFFBh }
else begin
    <sub_1A5F1h>();
    if <sub_1A49Bh>() 進位 then 結果 := −6            { 0FFFAh }
    else if <sub_1A4DAh>() 進位 then begin
        int 67h AH=45h, DX = word_23AE6h;             { EMS 釋放 handle }
        結果 := −4;                                    { 0FFFCh }
    end else begin
        int 21h AH=3Eh, BX = word_23AE4h;             { 關掉覆疊檔 }
        dword_28104h := cs:loc_1A588h;                { 換掉讀取常式 }
        dword_2810Ah := dword_23AFAh;                 { 存下舊向量 }
        dword_23AFAh := cs:loc_1A471h;
        結果 := 0;
    end;
end;
word_23AC6h := 結果;
```

★ 四個回傳值就是 **Turbo Pascal 的 `OvrResult` 錯誤碼**：

| 值 | Turbo Pascal 常數 |
|---|---|
| `0` | `ovrOk` |
| `−1` | `ovrError` |
| `−4`（`0FFFCh`） | `ovrIOError` |
| `−5`（`0FFFBh`） | `ovrNoEMSDriver` |
| `−6`（`0FFFAh`） | `ovrNoEMSMemory` |

所以 `word_23AC6h` ＝ **`OvrResult`**、`word_23AE4h` ＝ 覆疊檔的 DOS handle、
`word_23AE6h` ＝ EMS handle、`dword_23AFAh` ＝ 被接管的覆疊讀取向量。

**PC-98 版把整個 `GAME.OVR` 搬進 EMS**——成功之後檔案立刻關掉，
之後的覆疊載入都走記憶體。這是 DOS 版沒有的（DOS 側常駐裡沒有 `int 67h`）。

## `1A4DAh`（60 行）：把覆疊段搬進 EMS

```pascal
<sub_1A617h>();
int 67h AH=47h, DX = word_23AE6h;                     { 存 mapping context }

{ 先數一遍：從 word_23AD4h 起沿著 es:[0Eh] 走到 0，算出段數 cx }
{ 再逐段搬：對每一段設 es:[10h] := word_23ADCh、es:[16h]／es:[18h]，
  呼叫 dword_28104h（覆疊讀取常式），成功再叫 sub_1A544h }

int 67h AH=48h, DX = word_23AE6h;                     { 還原 mapping context }
<sub_1A634h>();
```

覆疊段的鏈用 **`+0Eh` 當 next**，`+10h`／`+16h`／`+18h` 是搬移時填的欄位。
搬移前後各存／還原一次 EMS mapping，所以**搬移期間可以任意切頁**。

⚠ `sp-analysis failed`（IDA 註記）——因為迴圈裡 `push cx` 與 `pop cx`
跨越了 `pop es` / `push es`，堆疊深度不是靜態常數。逐條讀是配平的。

## `1BEE8h`（64 行，far，`retf 6`）：`Assign` ＋ 開檔

雙入口：預設 `AH = 3Dh`（開啟，存取模式取自 `byte_23B0Eh`），
另一個入口 `AX = 3C00h`／`DX = 1`（建立）。

```pascal
f := [bp+8];                                          { File record 遠指標 }
if f^[2] = 0D7B0h then 直接往下                        { 已 Assign 未開 }
else if f^[2] = 0D7B3h then <sub_1BF69h>(f)            { 已開 → 先關 }
else begin word_23B08h := 66h;  離開 end;              { ★ 錯誤 102：未 Assign }

if f^[30h] <> 0 then begin                            { +30h 起是 ASCIZ 檔名 }
    int 21h AH=3Dh/3Ch, DS:DX = @f^[30h];
    if 進位 then begin word_23B08h := ax;  離開 end;
end else
    ax := dx;                                          { 空檔名 → 標準輸出入 }

f^[2] := 0D7B3h;                                       { 標記成「已開啟」 }
f^[0] := ax;                                           { DOS handle }
f^[4] := [bp+6];                                       { record size }
```

★ **Turbo Pascal 的 `File` record 版面**：`+0` handle、`+2` 狀態魔數
（`0D7B0h` ＝ 已 Assign、`0D7B3h` ＝ 已開啟）、`+4` record size、
`+30h` 起 ASCIZ 檔名。`word_23B08h` ＝ **`InOutRes`**（`66h` ＝ 102 ＝
「檔案未 Assign」）。

## `1BFD3h`（66 行，far，`retf 0Eh`）：`BlockRead` ／ `BlockWrite`

同樣是雙入口：`BL = 3Fh`／`CX = 64h`（讀，錯誤碼 100）與
`BL = 40h`／`CX = 65h`（寫，錯誤碼 101）。

```pascal
f := [bp+10h];
if not <sub_1BF8Eh>(f) then 結果 := 0
else begin
    n := [bp+0Ah];                                     { 要幾筆 }
    if n <> 0 then begin
        cx := n × f^[4];                               { × record size }
        int 21h AH=BL, BX = f^[0], DS:DX = 緩衝;
        if 進位 then begin word_23B08h := ax;  結果 := 0 end
        else ax := ax div f^[4];                       { 換回筆數 }
    end;
    if [bp+6] <> NIL then [bp+6]^ := ax;               { 可選的 Result 參數 }
    if (沒有 Result 參數) and (ax <> n) then
        word_23B08h := cx;                             { ★ 100／101 }
end;
```

★ **沒有傳 `Result` 參數時，讀寫筆數不足就設 `InOutRes := 100／101`**
（`ovrIOError` 之外的 `Disk read error`／`Disk write error`）。
傳了 `Result` 就只回報筆數、不設錯誤——這正是 Turbo Pascal
`BlockRead(F, Buf, Count)` 與 `BlockRead(F, Buf, Count, Result)` 的差別。

## `18FF5h`（48 行，near）：文字視窗參數初始化

```pascal
byte_280E5h  := <BIOS AH=3>();                         { 列數 − 1 }
word_280CAh.hi := byte_280E5h;
word_280F3h.hi := byte_280E5h;
word_280F5h.hi := byte_280E5h;
byte_280DBh  := byte_280E5h ＋ 1;                      { 列數 }

word_280C4h  := <BIOS AH=0Bh>() and 7;
if <那個回傳值> and 2 <> 0 then 欄數 := 50h else 欄數 := 28h;
byte_280DAh  := 欄數;                                  { 80 或 40 }
word_280CAh.lo := 欄數 − 1;
word_280F3h.lo := 欄數 − 1;
word_280F5h.lo := 欄數 − 1;

游標位移 := (word_280F9h.hi × 欄數 ＋ word_280F9h.lo) shl 1;
if 欄數 <> 50h then 游標位移 := 游標位移 shl 1;
word_280D8h := 游標位移;

word_280C8h := 0;  word_280EFh := 0;  word_280F1h := 0;
byte_280E4h := 0;  byte_280E3h := 0;  byte_280C3h := 0;
byte_280C2h := 1;
```

`word_280CAh`／`word_280F3h`／`word_280F5h` 三個 word 各裝
**（欄, 列）兩個 byte 的右下角**——就是 Turbo Pascal `Crt` 的
`WindMax` 那一組。40 欄時位移要再乘 2，**因為 PC-98 的文字 VRAM
在 40 欄模式下每格佔兩倍寬度**。

## 明確不宣稱

- 沒有宣稱 `<sub_1A485h>`／`<sub_1A49Bh>`／`<sub_1A544h>`／`<sub_1A5F1h>`／
  `<sub_1A617h>`／`<sub_1A634h>`／`<sub_1BF69h>`／`<sub_1BF8Eh>`／`<sub_1977Bh>`
  的內部。
- 沒有宣稱 `loc_1A471h`／`loc_1A588h` 兩段被掛上去的常式做什麼。
- 沒有宣稱 `word_23AD4h`／`word_23ADCh`／`word_23B04h` 的角色。
- 沒有宣稱 DOS 版是否也有 EMS 覆疊（常駐裡沒有 `int 67h`，但沒有全檔比對）。
