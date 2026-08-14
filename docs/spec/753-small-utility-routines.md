# 753 — 一批小工具程序

- 平台／模組：見各節
- 證據等級：`exact`（每支的原始 bytes 都逐條核對過；IDA 匯出在其中數支漏 byte，
  以 `.bin` 為準）

## `overlay-31:0007h`（兩平台，位元組相同）— `Sign`

```pascal
function Sign(x: integer): integer;   { retf 2，結果在 AX }
begin
    if x < 0 then Sign := -1
    else if x > 0 then Sign := 1
    else Sign := 0;
end;
```

比較是**有號**的（`jge` / `jle`）。DOS 與 PC-98 的 45 個 byte 完全相同。

## `overlay-25:139Ah`（DOS）／`14A2h`（PC-98）— 記錄型別判斷

```pascal
function 是某類(p: 遠指標): boolean;  { retf 4 }
begin 是某類 := (p^[74h] = 7) end;
```

回傳 0 或 1。兩平台除位址外相同。沒有宣稱 `+74h` 是什麼欄位、7 是什麼類別。

## `overlay-30:016Bh`（兩平台）— 四個 byte 的 setter

```pascal
procedure 設定(a, b, c, d: byte);     { retf 8，四個 word 參數只取低位 byte }
```

依「第一個宣告的參數在最高位址」讀：`a`→`DS:720Ah`、`b`→`720Bh`、`c`→`720Ch`、
`d`→`720Dh`（PC-98 對應 `DS:0A2A4h`..`0A2A7h`）。四個 byte 連續，是同一個
4-byte 記錄。

## `overlay-27:0007h` 與 `0032h`（兩平台）— 一組取值 ＋ 一組繪製

`0007h`（`retf`，無參數）：

```pascal
q := DS:4F99h^[342h];        { 遠指標的欄位，內容當 DS 內的位移用 }
DS:8C78h := byte[q + 0A48h];  { 零延伸成 word }
DS:8C7Ah := byte[q + 0A68h];
```

兩個來源相差 `20h`。PC-98 是 `DS:7F05h^[342h]`、`+160Ch` / `+162Ch`（同樣相差
`20h`），結果放 `DS:0BE04h` / `0BE06h`。

`0032h`（`retf`，無參數）：

```pascal
0297h:0EFFh(DS:8C78h, DS:8C7Ah, 遠指標@DS:65CAh);
0297h:1110h(DS:8C78h, DS:8C7Ah, 遠指標@DS:8C7Ch);
```

參數形狀由推入順序解出：前兩個 word 各自獨立，最後兩個 word 是「先高後低」，
所以是一個存在 `65CAh` / `8C7Ch` 的 4-byte 遠指標。PC-98 換成 `2A8h:098Fh` 與
`2A8h:0A86h`，位址對應 `0BE04h`／`0BE06h`／`9660h`／`0BE08h`。

## `overlay-29:0881h`（DOS）／`0777h`（PC-98）— 兩平台真的不一樣

DOS：

```pascal
01A0h:0252h();
0297h:1110h(1, 1, 遠指標@DS:728Ch);
```

PC-98：

```pascal
019Eh:0384h();
02A8h:0A86h(1, 1, 遠指標@DS:0A327h);
02A8h:10D5h(@DS:0A327h);          { DOS 沒有這一步 }
```

差異已對原始 bytes 確認，不是匯出誤差：DOS 版在第二個呼叫之後直接就是
`89 EC 5D CB`。`0297h:1110h` 與 `02A8h:0A86h` 是同一支的兩平台版本（`overlay-27`
也成對呼叫它們），參數形狀一致：`(word, word, 遠指標)`。

## `overlay-00:0051h`（DOS）— 安裝回呼

```asm
mov ax, 25h ; mov dx, 39h
mov ds:47C4h, ax ; mov ds:47C6h, dx
```

把遠指標 `0039h:0025h` 寫進 `DS:47C4h`。`0039h:0025h` 是 VROOMM stub，
`(25h − 20h) / 5 = 1` → **overlay-00 entry#1 `@0017h`**。與 spec 749 的
`DS:72A0h` 同一種「可換掉的掛鉤」寫法。

## `PC98-overlay-03:000Ch` — 永遠回 1

配置 `19Ah`（410）bytes 的 local，一個都沒用到，直接
`mov byte [bp-1], 1; mov al, [bp-1]` 回傳。`retf` 不收參數。

## `START.EXE:170ECh`（DOS）— 清空鍵盤緩衝區

```pascal
while KeyPressed do 丟掉(讀一個鍵);
```

`KEYPRESSED` 由 Borland 的符號還原確認；迴圈裡用 `push cs` ＋ `call near
sub_16FADh` 呼叫讀鍵程序，結果存進一個之後不再讀的 local。

## `START.EXE:1AAD5h`（DOS）— 向下配置 8 bytes 的邊界檢查

近呼叫（`retn`），以 CF 表示失敗：

```
di := word[2097Eh] − 8
if di = 0 then 失敗
si := (di shr 4) + word[20980h]        { 正規化成節數再加上段基底 }
if si <= word[2097Ch] then 失敗
word[2097Eh] := di                      { 成功才寫回 }
```

`2097Eh` / `20980h` 是一組 offset ＋ segment，`2097Ch` 是下界。形狀是「從高位址
往下切 8 個 byte，切完不能低於下界」。減法沒有處理借位：若 `word[2097Eh] < 8`
會迴繞成很大的值，而 `di = 0` 這個判斷只擋得住剛好等於 8 的情形。**沒有宣稱這
是 bug**——Turbo Pascal 的堆疊／堆積指標在正規化下是否可能小於 8，本輪沒有查。

## 明確不宣稱

- 沒有宣稱 `0297h:0EFFh`、`0297h:1110h`、`01A0h:0252h`、`02A8h:10D5h` 是什麼。
- 沒有宣稱 `DS:8C78h` / `8C7Ah` 那對值代表什麼（只知道是兩個零延伸的 byte，
  來源相差 `20h`）。
- 沒有宣稱 `START.EXE:1AAD5h` 是 RTL 的哪一支。它沒有 Borland 還原出的名稱。
