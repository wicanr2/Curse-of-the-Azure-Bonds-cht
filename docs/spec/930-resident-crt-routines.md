# 930 — 常駐的 CRT／文字模式 routine 三支

- 證據等級：`exact`（逐條讀完）
- 位置：DOS `START.EXE` 的 `19AE0h`、`19F2Bh`、`1A050h`
- 作法見 spec 783

這三支是 Turbo Pascal `Crt` 單元的實作，沒有 Pascal 序言（`push bp`），
以組語手寫、`retn` 收尾。

## `19AE0h`（45 條）：顯示初始化

```pascal
mode := <int 10h AH=0Fh>();                 { sub_1A0B8 是 int 10h 的包裝 }
if (mode <> 7) and (mode > 3) then <設模式>(3);
<sub_19B97>();
attr := <int 10h AH=08h, BH=0>() shr 8;
byte_24E6A := attr and 7Fh;
byte_24E60 := attr and 7Fh;
byte_24E5B := 0;  byte_24E6B := 0;  byte_24E6C := 0;
byte_24E5A := 1;

{ 用 BIOS 計時器 0040:006Ch 校時 }
等到 es:[0040:006Ch] 變動;
<sub_19D6E>(cx = 0FFFFh);                   { 空轉計數 }
word_24E66 := (not 37h) div cx;             { 每毫秒的空轉次數 }

<int 21h AX=251Bh, DS:DX = cs:012Fh>();     { 掛 INT 1Bh（Ctrl-Break）handler }
```

**模式 7 是單色文字，被特別放行**；其餘大於 3 的模式一律拉回 3（80×25 彩色文字）。
校時那一段是 `Delay` 的校準——`word_24E66` 是「一個計時單位要空轉幾次」。

## `19F2Bh`（43 條）：輸出一個字元

```pascal
push bx, cx, dx, es;
<sub_19FA5>();                              { 取游標 }
case ch of
  07h: <int 10h AH=0Eh>(ch);                { BEL：交給 BIOS 響鈴 }
  08h: if dl <> word_24E62.lo then dec(dl); { BS }
  0Dh: dl := word_24E62.lo;                 { CR：回到左界 }
  0Ah: <sub_19F84>();                       { LF：捲一行 }
else
  <int 10h AH=09h>(ch, 屬性 byte_24E60, 次數 1);
  inc(dl);
  if dl > word_24E64.lo then begin          { 超過右界 }
      dl := word_24E62.lo;  <sub_19F84>();  { 折行 }
  end;
end;
<sub_19FAC>();                              { 寫回游標 }
pop es, dx, cx, bx;
```

`word_24E62` 與 `word_24E64` 是視窗的左上／右下角（各一個 word 裝兩個 byte），
`byte_24E60` 是目前的文字屬性。**BEL 走 BIOS，不走本檔的音樂引擎**（spec 929）。

## `1A050h`（56 條）：文字模式區塊搬移

```pascal
if si = di then 離開;
cx := di − si;                              { 長度 }
dl := byte_24E5D;   dh := byte_24E60;
{ 由 BIOS 資料區取硬體參數 }
di := (bh × byte[0040:004Ah] ＋ bx) × 2;    { 004Ah ＝ 每列欄數 }
dx := word[0040:0063h] ＋ 6;                { CRTC 狀態埠 }
段 := 0B800h;
if byte[0040:0049h] = 7 then 段 := 0B000h;  { 單色卡 }
…
```

`0040:0049h` 是目前顯示模式、`0040:004Ah` 是每列欄數、`0040:0063h` 是 CRTC 的
基底埠。`+6` 之後是**狀態暫存器**，用來等 retrace（避免 CGA 雪花）。
`0B800h` 彩色／`0B000h` 單色的分流是標準做法。

## 為什麼這三支值得記

remake 不需要重現它們，但它們**界定了原版的文字視窗模型**：
80×25、屬性單一 byte、視窗由 `word_24E62`／`word_24E64` 兩角決定、
折行與捲動都由 CRT 單元負責。spec 916／920 那些「欄、列」座標
就是落在這個座標系裡。

## 明確不宣稱

- 沒有宣稱 `sub_1A0B8`／`sub_19B97`／`sub_19D6E`／`sub_19F84`／`sub_19FA5`／
  `sub_19FAC` 的內部（形狀上分別是 int 10h 包裝、視窗初始化、空轉計數、
  捲一行、取游標、寫游標）。
- 沒有宣稱 `byte_24E5A`／`byte_24E5B`／`byte_24E5D`／`byte_24E6A`／`byte_24E6B`／
  `byte_24E6C` 各自的用途。
- 沒有宣稱 `1A050h` 尾段（`cld` 之後）的搬移細節——本規格只讀到參數準備完畢，
  其後是標準的 `rep movsw` 迴圈。
