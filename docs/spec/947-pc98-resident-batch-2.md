# 947 — PC-98 常駐四支：長整數除法、表面釋放、記憶體池回收、清文字矩形

- 證據等級：`exact`（四支逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `1A8E9h`、`13B55h`、`18683h`、`19840h`
- 作法見 spec 783

## `1A8E9h`（76 行，far）：32 位元有號除法

`bx:cx ÷ dx:ax`，商在 `dx:ax`、餘數在 `bx:cx`。

```
if (cx or bx) = 0 then begin ax := 0C8h;  jmp <sub_1A721h> end;   { ★ 除以 0 }
記下兩個運算元的正負，各自取絕對值；
xor 出 33 次的 restoring division 迴圈（bp := 21h ＝ 33）：
    rcl cx,1 / rcl bx,1 / sub cx,si / sbb bx,di
    不夠減就加回去，cmc 之後 rcl ax,1 / rcl dx,1
依原本的正負號把商與餘數各自取回負值。
```

`0C8h` ＝ **200 ＝ Turbo Pascal 的「Division by zero」執行期錯誤碼**，
`<sub_1A721h>` 是執行期錯誤處理常式。

這一支就是專案裡到處看到的 `0A65h:0299h`／`0A54h:0294h`
（**商在 `dx:ax`、餘數在 `bx:cx`**）那個約定的實作——spec 913／925
用到的除法語意由此確認。

## `13B55h`（72 行，far，`retf 4`）：釋放一個表面

```pascal
if 表面指標^ = NIL then 離開;
s := 表面指標^;
甲 := s^[8] × s^[11h];                     { 張數 × 每張位元組 }
乙 := s^[0] × s^[2] × s^[8];               { 高 × 寬 × 張數 }
if (s^[2] > 16h) and (s^[13h] <> NIL) then
    FreeMem(@s^[13h], 乙);                 { 遮罩緩衝 }
FreeMem(表面指標, 甲 ＋ 17h ＋ 乙);
表面指標^ := NIL;
```

表面標頭 `17h` bytes 與 `+0`／`+2`／`+8`／`+11h`／`+13h` 五個欄位
與 spec 934（DOS 的配置端）逐格對應。

**配置端是 spec 968 的 `139DDh`**：PC-98 的表面本來就是
`主資料 ＋ 17h ＋ 遮罩` 一次要齊，遮罩 `+13h` 指進同一塊的後段。
所以本支把兩段都算進 `FreeMem` 是正確的、與配置端一致。
不一致的是 **DOS 版**——spec 934 的 `GetMem(總量 ＋ 17h)` 另外配遮罩。
本支多的那道 `寬 > 16h` 對應 spec 968 四個「遮罩不接在主資料後面」的特例。

## `18683h`（67 行）：把區塊還回池子

```pascal
if <sub_187ECh>() 進位 then 結果 := 0FFFFh
else begin
    <sub_1883Eh>(word_2418Ah);             { 從池 A 摘下 }
    word_2418Ch := word_2418Ch − MCB^[0];  { 扣掉大小 }
    尾 := 區塊段 − 1 ＋ MCB^[0] ＋ 1;
    if 尾 = word_24186h then begin         { ★ 剛好接在池尾 }
        word_24186h := 區塊段 − 1;          { 直接把池邊界往回縮 }
        word_24188h := word_24188h ＋ MCB^[0] ＋ 1;
        MCB^[8] := 0;
    end else begin
        <sub_18810h>(word_2418Eh);         { 掛到池 B }
        word_24190h := word_24190h ＋ MCB^[0];
        MCB^[8] := 0;
    end;
    結果 := 0;
end;
```

與 spec 945 的 `18612h`（搬進池）是一對。**區塊剛好貼著池尾時直接縮邊界**，
不放進可回收鏈——這是簡單的 bump allocator 反向操作。
`MCB^[8]` 的魔數 `818Eh`（spec 945）在這裡被清成 0。

## `19840h`（67 行）：清一個文字矩形

```pascal
di := (列 × byte_280DAh ＋ 欄) × 2;
寬 := (右 − 左 ＋ 1);   { dx，經 dx := dx − cx ＋ 101h 後拆成 dh/dl }
si := (byte_280DAh − 寬) × 2;              { 每列剩下要跳過的距離 }
if byte_280DAh <> 50h then begin di 、寬、si 全部再 shl 1 end;   { 40 欄模式 }

es := word_280D4h;                         { A000h，字碼平面 }
for y := 1 to 高 do begin
    <sub_19B7Dh>();  rep stosw (' ');  <sub_19B9Ch>();
    di := di ＋ si;
end;

di := di ＋ word_280D6h;                   { ★ 跳到屬性平面（＋2000h） }
<sub_18FA3h>(bh);
for y := 1 to 高 do begin
    for x := 1 to 寬 do begin stosb;  inc di end;   { 屬性一格佔兩個 byte，只寫低位 }
    di := di ＋ si;
end;
```

★ **字碼平面用 `stosw` 一次兩個 byte，屬性平面用 `stosb ＋ inc di` 隔一個寫一個**
——PC-98 的屬性平面每格也佔兩個 byte，但只有低位有意義。

`word_280D6h` ＝ `2000h` 的偏移在這裡實際被用到，印證 spec 946。
`byte_280DAh`（欄數 80 或 40，spec 944）決定 40 欄模式下所有位移要再加倍。

## 明確不宣稱

- 沒有宣稱 `<sub_1A721h>`（執行期錯誤）、`<sub_187ECh>`、`<sub_1883Eh>`、
  `<sub_18810h>`、`<sub_19B7Dh>`、`<sub_19B9Ch>`、`<sub_18FA3h>`、`<sub_18392h>`
  的內部。

- 沒有宣稱 spec 968 那四個特例（含 `寬 > 16h`）對應哪幾種圖庫。
