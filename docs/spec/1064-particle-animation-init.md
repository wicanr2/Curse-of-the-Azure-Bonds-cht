# 1064 — 粒子動畫初始化：填 spec 1025 那張「執行期才填」的顏色斜坡表

- 證據等級：`exact`（DOS 側 382 條逐條讀完）
- 作法見 spec 783；動畫的每一幀見 spec 1025

## `dos overlay-18:000BCh`（`retf`）

原本待解讀。參數 `arg_0` 是一個遠指標，開頭 3 bytes 被複製出來當「輸入色」。

## ★★★ 關掉 spec 1025 的未定項：顏色斜坡表是這裡填的

spec 1025 記「那張顏色斜坡表在靜態映像裡是 `0FFh`，**執行期才填**」，
並把位置記成 `byte[4ACDh ＋ 列 × 5 ＋ 目前色階]`。這一支就是填表的人：

```pascal
DS:4AD2h := 0;
FillChar(DS:4AD3h, 0Fh, 1);
for 列 := 1 to 3 do begin
    if DS:4FE6h = 0 then begin                       { ★ CGA }
        if 輸入色[列] > 1 then begin
            byte[4ACEh ＋ 列 × 5] := 2;
            byte[4ACFh ＋ 列 × 5] := 3;
            byte[4AD0h ＋ 列 × 5] := 3;
            byte[4AD1h ＋ 列 × 5] := 3;
            byte[4AD2h ＋ 列 × 5] := 2;
        end;
    end else begin                                   { ★ EGA／Tandy }
        if 輸入色[列] > 1 then byte[4ACEh ＋ 列 × 5] := 0Fh;
        byte[4ACFh ＋ 列 × 5] := 0Fh;
        if 輸入色[列] > 1 then byte[4AD0h ＋ 列 × 5] := 輸入色[列] ＋ 8;
        byte[4AD1h ＋ 列 × 5] := 輸入色[列];
        byte[4AD2h ＋ 列 × 5] := 1;
    end;
end;
```

> ★★★ **`4ACEh ＋ 列 × 5` .. `4AD2h ＋ 列 × 5` 正好是
> spec 1025 的 `4ACDh ＋ 列 × 5 ＋ 色階`（色階 `1`..`5`）。**
> ⇒ spec 1025 的「沒有宣稱顏色斜坡表是誰填的」**關掉**。
>
> ★★★ **兩套值由 `DS:4FE6h`（顯示卡種類，spec 1046）分**：
> CGA 只有 `2`／`3` 兩個顏色可用，其他卡走 `0Fh`（白）→ `輸入色 ＋ 8`（亮色）
> → `輸入色` → `1` 的漸層。
> ⇒ **CGA 版的粒子只有兩階明暗，其他卡是四階漸層。**

★ `DS:4AD3h` 起 15 bytes 全填 `1`——形狀上是第 4 列以後的預設。

## ★★ 幀數與初速

```pascal
DS:4AD2h := max(DS:4AD2h, Random(14h) ＋ 19h);       { ★ 25..44 }
var_16 := Random(5) ＋ 5;                             { 5..9 }
var_18 := var_16 ＋ 0Fh;                              { ＋15 }
{ 兩次：浮點亂數 × 常數 → 累加 }
<0A54:111Bh>();                                       { 浮點 Random }
<0A54:0C5Ah>(…, cx ＝ 0DC83h, si ＝ 0CF80h, di ＝ 490Fh);
<0A54:0D7Eh>(…);
…
Random(0Ah) ＋ 18h;                                   { 24..33 }
```

★ `DS:4AD2h` 被拿來當**全域上限**（先跟 `Random(14h) ＋ 19h` 取大的）
——注意它同時是顏色斜坡表第 1 列的最後一格，**兩個用途共用同一個 byte**。
⚠ 本規格不判斷這是不是原作的疏漏。

★ `0A54:111Bh` ＝ 浮點亂數、`0A54:0C5Ah`／`0D7Eh` ＝ 浮點運算；
`490F CF80 DC83` 是一個 Turbo Pascal 6-byte real 常數
——與 spec 991 記的「方向用浮點算」是同一組常式。

## 中文化

本支沒有字串。

## 明確不宣稱

- 沒有宣稱那個 6-byte real 常數的數值。
- 沒有宣稱 `arg_0` 指的 3 bytes（每列一個「輸入色」）由誰給。
- 沒有宣稱 `DS:4AD2h` 一格兩用是不是原作疏漏。
- 沒有宣稱 `DS:4AD3h` 起那 15 bytes 的用途。
- 沒有宣稱 `Random(5) ＋ 5`／`Random(0Ah) ＋ 18h` 各自算的是什麼。
