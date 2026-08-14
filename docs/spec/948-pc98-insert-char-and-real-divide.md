# 948 — PC-98 常駐兩支：全形感知的字元插入、6-byte real 除法

- 證據等級：`exact`（兩支逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `19BF0h`、`1B1ACh`
- 作法見 spec 783

## ★ `19BF0h`（83 行）：在輸入緩衝裡蓋掉一個字元

暫存器約定：`al` ＝ 要寫的字元、`es:di` ＝ 緩衝、`bx` ＝ 游標位置、
`si` ＝ 目前長度、`dx` ＝ 緩衝上限。

```pascal
if byte_280E3h <> 0 then goto 直接寫;          { 不做全形判定 }

if (byte_280D0h <> 0) and <sub_1977Eh>(要寫的字元) then begin
    { 要寫的是全形前導 }
    if bx ＋ 1 >= dx then begin                { ★ 放不下 }
        byte_280E1h := 0;
        <sub_192A0h>();                        { 錯誤／嗶聲 }
        離開;
    end;
    if si <= bx then goto 直接寫;              { 在字串尾端，直接寫 }
    { 游標下原本的字：前導 → 不用挪；否則要挪 }
    ah := 要寫的字元;
    al := 緩衝[bx];
    if <sub_1977Eh>(al) then goto 還原並寫;
    if not <sub_1977Eh>(緩衝[bx＋1]) then goto 還原並寫;
    goto 左移一格;
end else begin
    if bx = dx then 離開;                       { 滿了 }
    if si <= bx then goto 直接寫;
    ah := 要寫的字元;
    al := 緩衝[bx];
    if byte_280D0h = 0 then goto 還原並寫;
    if not <sub_1977Eh>(al) then goto 還原並寫;
    { 落到「左移一格」}
end;

左移一格:                                       { ★ 蓋掉全形字的前半 → 整段左移 }
    cx := si − bx − 2;
    while cx > 0 do begin 緩衝[bx] := 緩衝[bx＋1];  inc(bx);  dec(cx) end;
    dec(si);

還原並寫:  al := ah;
直接寫:
    緩衝[bx] := al;
    inc(bx);
    <sub_19493h>(al);                           { 送到畫面 }
    if bx > si then si := bx;                   { 長度往外長 }
```

**與 spec 945 的退格是一對**。核心規則：

- 要寫的字元本身是全形前導時，**緩衝至少要再容得下一個 byte**，
  否則叫 `<sub_192A0h>` 拒絕。
- 游標下面原本是一個**完整的全形字**（前導 ＋ 後續）時，
  蓋掉它的前半會留下孤兒後續，所以**把後面整段左移一格、長度減 1**。
- `byte_280E3h` 或 `byte_280D0h` 為 0 時整套判定都跳過（純單位元組模式）。

繁中版可以照抄這個結構；`<sub_1977Eh>` 換成 Big5 的前導判定即可。

## `1B1ACh`（79 行）：6-byte real 除法

Turbo Pascal 的 `Real` 是 **1 byte 指數 ＋ 40 bits 尾數 ＋ 1 bit 符號**。

```
al = 0 → 除數為零，跳 loc_1B1A5（錯誤路徑）
符號：dx ← 兩個運算元符號的 xor（只留 bit 15）
指數：dl:dh ← 被除數指數 − 除數指數
尾數：兩邊都補上隱含的最高位（or …, 8000h）

迴圈（al 從 2 起、dx 起始 1）：
    比較 bp:bx:ah 與 di:si:ch（48 bits），夠減就減
    rcl dx,1 把商位推進去；進位表示這一段商滿了
    左移被除數一位（shl ah,1 / rcl bx,1 / rcl bp,1），溢位就再減一次
    dx 滿了就 push 起來、換下一段（第三段用 dl := 40h 只收 2 bits）

收尾：ax := dx shl 6；把三段商 not 之後組成尾數
      指數 cx ＋ 8080h，跳 loc_1B18Ah 做正規化與符號套用
```

`8080h` 是**兩個 `80h` 偏置**（被除數的減去除數的之後要加回一個偏置，
再加上 Turbo Pascal `Real` 的指數偏置）。

⚠ IDA 註記 `sp-analysis failed`——迴圈裡的 `push dx` 與收尾的 `pop`
不在同一條路徑上，堆疊深度不是靜態常數。逐條讀是配平的
（三段商各 `push` 一次、收尾 `pop bx` ＋ `pop dx` ＋ `pop cx` ＋ `pop bp`）。

## 明確不宣稱

- 沒有宣稱 `<sub_1977Eh>` 判前導的範圍、`<sub_19493h>` 送字元的細節、
  `<sub_192A0h>` 是嗶聲還是別的。
- 沒有宣稱 `byte_280E3h` 與 `byte_280D0h` 的差別（兩者都能關掉全形判定）。
- 沒有宣稱 `loc_1B18Ah`（正規化收尾）與 `loc_1B1A5h`（除以零）的內容。
