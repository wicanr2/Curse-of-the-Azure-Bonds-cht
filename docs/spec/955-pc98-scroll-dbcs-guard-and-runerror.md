# 955 — PC-98 的捲動（會保護被切半的全形字）與 `Runtime error` 終止

- 證據等級：`exact`（兩支逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `198AEh`、`1A726h`
- 作法見 spec 783

## ★ `198AEh`（108 行）：文字畫面的區塊搬移

```pascal
if (bx = 0) or (cx = 0) then 離開;                { 寬或高為 0 }
if byte_280DAh <> 50h then begin ax := ax shl 1;  dx := dx shl 1 end;  { 40 欄 }

si := <sub_199F5h>(來源座標);
di := <sub_199F5h>(目的座標);
bx := (50h − dx) shl 1;                           { 每列搬完要跳過的距離 }
ds := es := word_280D4h;                          { 0A000h，字碼平面 }

for y := 1 to 高 do begin
    { ★ 先檢查這一列的第一格與最後一格 }
    v := word(ds:si);
    if (hi(v) <> 0) and (hi(v) 或 lo(v) 的 bit 7 立著) then word(ds:si) := 20h;
    v := word(ds:[si ＋ (寬 − 1) × 2]);
    if 同樣的條件 then word(…) := 20h;

    <sub_19B7Dh>();  rep movsw;  <sub_19B9Ch>();
    si := si ＋ bx;   di := di ＋ bx;
end;

ds := es := word_280D4h ＋ 200h;                  { ＝ 0A200h，屬性平面 }
for y := 1 to 高 do begin
    for x := 1 to 寬 do begin movsb;  inc si;  inc di end;
    si := si ＋ bx;   di := di ＋ bx;
end;
```

### ★ 全形字被切半時改成空白

搬移的矩形邊界可能**剛好落在一個全形字的中間**。
本支在搬每一列之前檢查**第一格與最後一格**：只要那一格的
bit 15（右半標記，spec 953）或字碼的 bit 7 顯示它是全形字的一部分，
就先把該格改成空白 `20h`，**再搬**。

於是被切開的半個字不會以亂碼的形式出現在新位置——
這是 spec 945／948／952／953 之外**第五個**認得雙位元組的環節。

屬性平面在 `word_280D4h ＋ 200h` ＝ `0A200h`，與 spec 950 的寫法一致；
同樣是 `movsb ＋ inc` 隔一個搬一個。

## `1A726h`（108 行，far）：`Runtime error` 與程式終止

```pascal
word_23AFEh := 錯誤碼;

if (cx or bx) <> 0 then begin                     { 有出錯位址 }
    { 走 word_23AE2h 的覆疊鏈（next 在 +14h），比對 +10h 找出是哪一段 }
    bx := 該段的段位址 − word_23B04h − 10h;       { 換算成覆疊內的相對位址 }
end;
word_23B00h := cx;   word_23B02h := bx;

if dword_23AFAh <> NIL then begin                 { ExitProc 鏈 }
    dword_23AFAh := NIL;   word_23B08h := 0;      { InOutRes 清 0 }
    以 cs:0110h 當回位址，retf 跳進 ExitProc;
end;

{ ── 真的要結束了 ── }
<sub_1BA90h>(ds:0BE7Eh);   <sub_1BA90h>(ds:0BF7Eh);
di := 0C07Eh;   si := cs:01E1h;   cx := 13h;
for i := 1 to 13h do begin                        { ★ 還原 19 個中斷向量 }
    al := cs:[si];  inc(si);
    int 21h AH=25h, DS:DX = [di];
    di := di ＋ 4;
end;

if (word_23B00h or word_23B02h) <> 0 then begin
    <sub_1A7E8h>(1F4h);                           { 'Runtime error ' }
    <sub_1A7F6h>(word_23AFEh);                    { 錯誤碼（十進位） }
    <sub_1A7E8h>(203h);                           { ' at ' }
    <sub_1A810h>(word_23B02h);                    { 段（十六進位） }
    <sub_1A82Ah>(':');
    <sub_1A810h>(word_23B00h);                    { 偏移 }
    <sub_1A7E8h>(208h);                           { 換行 }
end;

int 21h AH=4Ch, AL = 錯誤碼;
```

★ **玩家看到的 `Runtime error NNN at XXXX:YYYY` 就是這裡印的。**
出錯位址會先**換算成覆疊段內的相對位址**（減 `word_23B04h` 與 `10h`），
所以那個 `XXXX` 不是絕對段址——**對照反組譯時要記得加回去**。

還原向量的表在 `cs:01E1h`，19 筆中斷編號，對應的舊向量存在 `ds:0C07Eh` 起、
每筆 4 bytes。程式一啟動就把這些接管走了。

`1F4h`／`203h`／`208h` 是訊息表裡的三個位移（前綴、` at `、結尾）。

⚠ IDA 註記 `sp-analysis failed`——因為 ExitProc 那條路用
`push cs / push ax / push es / push bx / retf` 手動組出一個遠呼叫，
堆疊不是靜態平衡的。逐條讀是正確的。

## 明確不宣稱

- 沒有宣稱 `<sub_199F5h>`（座標 → 位移）、`<sub_19B7Dh>`／`<sub_19B9Ch>`
  （搬移前後的鉤子）、`<sub_1BA90h>`、`<sub_1A7E8h>`／`<sub_1A7F6h>`／
  `<sub_1A810h>`／`<sub_1A82Ah>` 的內部。
- 沒有宣稱 `cs:01E1h` 那 19 個中斷編號是哪些。
- 沒有宣稱 `word_23B04h` 的意義（只知道換算相對位址時要減掉它）。
