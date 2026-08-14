# 787 — Bresenham 走線記錄、地圖邊界的第二種政策、掃描結果的複製

- 證據等級：`exact`（DOS 側逐條讀完；PC-98 側的運算元差異已逐條列出）
- 作法見 spec 783

## `overlay-31:019Dh`（entry#2，58 條，**差異 0 條**）— Bresenham 的初始化

兩平台**逐位元組相同**（一條運算元差異都沒有）。`retf 4`，參數是一個記錄的
遠指標：

```pascal
r^[0Eh] := r^[00h];                    { 目前 x := 起點 x }
r^[10h] := r^[02h];                    { 目前 y := 起點 y }
r^[0Ah] := abs(r^[04h] − r^[00h]);     { |dx| }
r^[0Ch] := abs(r^[06h] − r^[02h]);     { |dy| }
r^[12h] := Sign(r^[04h] − r^[00h]);    { x 方向：−1 / 0 / +1 }
r^[14h] := Sign(r^[06h] − r^[02h]);    { y 方向 }
r^[08h] := 0;                          { 誤差累加器 }
r^[16h] := 0;                          { 結束旗標（byte）}
```

`Sign` 就是 spec 753 讀過的 `overlay-31:0007h`。整個記錄的版面因此定了下來：

| 位移 | 型別 | 內容 |
|---|---|---|
| `+00h` / `+02h` | word | 起點 x / y |
| `+04h` / `+06h` | word | 終點 x / y |
| `+08h` | word | 誤差累加器 |
| `+0Ah` / `+0Ch` | word | `|dx|` / `|dy|` |
| `+0Eh` / `+10h` | word | 目前 x / y |
| `+12h` / `+14h` | word | x / y 的步進方向 |
| `+16h` | byte | 結束旗標 |

這是**標準的 Bresenham 走線狀態**，共 `17h`（23）bytes。取絕對值用的是
`or ax,ax` ＋ `jns` ＋ `neg`（有號）。

## `overlay-14:078Eh`（entry#12，55 條，差異 15 條）— 邊界夾制（與環繞並存）

```pascal
x := 有號(DS:720Fh);  y := 有號(DS:7210h);  d := DS:7211h;
DS:4F9Dh^[5AAh] := 0;
if <017Fh:0039h>(x, y, d) then begin
    x := x + byte[2694h + d];
    y := y + byte[269Dh + d];
    if x > 0Fh then begin DS:720Fh := 0Fh;  DS:4F9Dh^[5AAh] := 1 end;
    if x < 0   then begin DS:720Fh := 0;    DS:4F9Dh^[5AAh] := 1 end;
    if y > 0Fh then begin DS:7210h := 0Fh;  DS:4F9Dh^[5AAh] := 1 end;
    if y < 0   then begin DS:7210h := 0;    DS:4F9Dh^[5AAh] := 1 end;
end;
```

兩件要照抄的事：

- **算出來的新座標只有在越界時才寫回去**（而且寫的是夾制值）。留在範圍內的
  情形，`x` / `y` 這兩個 local 算完就丟——實際移動是 spec 770 的
  `overlay-14:083Ch` 做的。
- 同一個模組裡**兩種邊界政策並存**：這一支**夾制**並把 `DS:4F9Dh^[5AAh]`
  設成 1，`083Ch` 則是**無條件環繞**。remake 不能挑一種統一。

四個判斷是四個獨立的 `if`（不是 `else if`），所以理論上 x 與 y 可以各自被夾
一次，旗標會被重複設成 1。

## `overlay-13:38AAh`（entry#41，60 條，差異 9 條）— 把掃描結果複製出來

```pascal
n := <掃描>(緩衝(DS:6E92h), <取某值>(x, y), 0FFh, 7Fh,
            <取 y>(x, y), <取 x>(x, y));
ss:[rec − 0DAh] := DS:6E96h;                       { 結果筆數 }
for i := 1 to ss:[rec − 0DAh] do
    Move(DS:6E94h + i * 3, ss:[rec − 0DBh + i * 3], 3);
```

用的是 spec 777 記錄過的同一組全域：`DS:6E96h` 是筆數、`DS:6E94h` 起是**每筆
3 bytes** 的結果陣列（1 起算）。這一支把整批搬進呼叫端的 SS 相對記錄——
**筆數放在 `rec − 0DAh`，資料從 `rec − 0D8h` 起**（同樣每筆 3 bytes、1 起算）。

`<取 x>` / `<取 y>` 是 spec 759 認過的 `overlay-32 entry#15 / #16`（戰鬥員的
世界座標）。

搬完之後**呼叫端的陣列大小沒有任何檢查**——筆數由掃描程序決定。

## 明確不宣稱

- 沒有宣稱 Bresenham 記錄是用來走什麼（視線？投射物？）。
- 沒有宣稱 `017Fh:0039h` 判的是什麼（形狀上是「這一步走不走得動」）。
- 沒有宣稱 `DS:4F9Dh^[5AAh]` 這個旗標由誰讀。
- 沒有宣稱掃描程序那兩個常數 `0FFh` / `7Fh` 的意義。
