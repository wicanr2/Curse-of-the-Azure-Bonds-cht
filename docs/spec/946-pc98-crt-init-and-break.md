# 946 — PC-98 的 Crt 初始化與 Ctrl-Break 處理

- 證據等級：`exact`（兩支逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `18F10h`、`19085h`
- 作法見 spec 783

## ⚠ 先講一個判讀陷阱：IDA 的 `int 18h` 註解是錯的

`19085h` 裡有兩處 `int 18h`，IDA 自動加上的註解是

```
int 18h ; TRANSFER TO ROM BASIC
        ; causes transfer to ROM-based BASIC (IBM-PC)
```

**那是 IBM PC 的中斷表。PC-98 的 `int 18h` 是 CRT／鍵盤 BIOS**
（`AH=00h` 讀鍵、`AH=01h` 查按鍵狀態、`AH=0Ah`..`0Ch` 是畫面控制）。
本支用的 `AH=1` ＋ 檢查 `BH` 正是**查鍵盤緩衝區有沒有東西**，
和「跳進 ROM BASIC」毫無關係。

> 引用 PC-98 反組譯裡的 `int` 註解之前，先確認 IDA 用的是哪一張表。
> 這與 spec 928 記的 stub 撞號同類：**自洽但錯**。

## `18F10h`（59 行）：Crt 單元初始化

```pascal
byte_280D0h := 1;                          { 開啟雙位元組處理，spec 945 }
word_280D2h := 4Ch;
byte_280C1h := 0;  byte_280E1h := 0;  byte_280E2h := 0;  byte_28102h := 0;
cs:byte_193C8h := 0;                        { ⚠ 寫進程式碼段 }
byte_280C0h := 1;  byte_280FFh := 1;  byte_28100h := 1;  byte_28101h := 1;

int 21h AX=2506h, DS:DX = cs:0196h;        { 掛 INT 06h handler }

dword_280DCh := dword_23AFAh;              { 存下覆疊向量，spec 944 }

byte_280E5h := <sub_19AA2h>();             { 列數 − 1 }
<sub_18FE4h>(<BIOS AH=0Bh>() and 7);
word_280EDh := 0;
<sub_18FF5h>();                             { 視窗參數，spec 944 }
cs:byte_18FA2h := 0;                        { ⚠ 又一個 code segment 旗標 }

word_280D4h := 0A000h;                      { ★ 文字 VRAM 段 }
word_280D6h := 2000h;                       { ★ 屬性平面的偏移 }
探一次 [word_280D4h : word_280D8h ＋ word_280D6h];   { 讀一個 byte 確認可存取 }

byte_280E0h := <sub_18FC5h>();
byte_280C6h := 同值;  byte_280F7h := 同值;  byte_280F8h := 同值;
```

★ **PC-98 的文字畫面在 `A000:0000`，屬性平面在 `+2000h`**。
兩者以同一個位移索引，一個放字碼、一個放屬性——
與 IBM PC 的「字碼與屬性交錯」完全不同。
spec 944 的 `18FF5h` 算出的 `word_280D8h` 就是這個共用位移。

## `19085h`（91 行，far，`retf 2`）：Ctrl-Break 處理

```pascal
if byte_280E2h = 0 then 離開;              { 沒有待處理的中斷 }
byte_280E2h := 0;
if byte_280C0h = 0 then begin 結果 := 3;  離開 end;

repeat until int 21h AH=6, DL=0FFh 沒有字元;   { 清空輸入 }
repeat
    if int 18h AH=1 回報有鍵 then int 18h AH=0;  { ★ PC-98 鍵盤 BIOS，把鍵吃掉 }
until 沒有鍵;

送('^');  送('C');                          { 印出 ^C }
<sub_1948Ch>();  <sub_19AC2h>();
int 23h;                                    { DOS 的 Ctrl-C 離開位址 }

r := <sub_19A69h>();
if (r <> 14h) and (r <> 19h) then <sub_19112h>();

word_280F9h := 0;  word_280D8h := 0;  byte_280D0h := 0;
r2 := <sub_19ACEh>(ss:[bx+4]);
if r2 <= 3 then byte_280D0h := 1;           { 重新開啟雙位元組處理 }
if r2 and 1 = 0 then byte_280E5h := 18h     { 24 列 }
else                byte_280E5h := 13h;     { 19 列 }
<sub_18FE4h>();  <sub_18FF5h>();  <sub_19A59h>();
byte_280C6h := byte_280E0h;
```

**`int 23h` 之後畫面模式可能被改掉**，所以整段結尾把視窗參數重算一次
（`18FE4h` ＋ `18FF5h`），並依 `<sub_19ACEh>` 回報的模式決定
列數是 **24（`18h`）還是 19（`13h`）**。

`byte_280D0h`（雙位元組開關，spec 945）在處理期間先關掉、
確認模式 ≤ 3 才打開——**PC-98 的高解析模式才支援全形**。

## 明確不宣稱

- 沒有宣稱 `cs:byte_193C8h`／`cs:byte_18FA2h` 這兩個 code segment 旗標
  的用途（與 spec 945 的 `cs:byte_19D01h` 是同一種寫法）。
- 沒有宣稱掛在 INT 06h 的 `cs:0196h` 做什麼。
- 沒有宣稱 `<sub_18FC5h>`／`<sub_18FE4h>`／`<sub_19A59h>`／`<sub_19A69h>`／
  `<sub_19AA2h>`／`<sub_19AC2h>`／`<sub_19ACEh>`／`<sub_19112h>` 的內部。
- 沒有宣稱 `19085h` 的回傳值 `3` 代表什麼。
