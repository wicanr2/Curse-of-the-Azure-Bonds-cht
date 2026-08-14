# 898 — 3D 視圖的四塊牆面：依左右前方的狀態選圖磚

- 證據等級：`exact`（DOS 側逐條讀完；PC-98 側只差一條 `xor ah, ah`）
- 作法見 spec 783

## `overlay-10:006ECh` ↔ `overlay-10:006E6h`（DOS 186 條／PC-98 185 條）

`overlay-10` ＝ COMPREP。`retf 2` ＝ 一個**父程序的框架指標**（巢狀程序），
父框架的 `−1`／`−2` 是一組座標。

```pascal
左 := <本模組 0378h>(2, 父[−2] − 1, 父[−1]);
右 := <本模組 0378h>(0, 父[−2],     父[−1] ＋ 1);
兩邊都空 := (左 = 0) and (右 = 0);

for i := 1 to 4 do
    case i of
      1: if      DS:4A0Eh = 0 then A := (左 = 1) ? 04h : 16h
         else if DS:4A0Eh = 3 then A := 0Fh
         else if DS:4A0Eh = 1 then A := 05h;

      2: if DS:4A0Eh = 0 then begin
             if      左 = 0 then B := 16h
             else if 左 = 3 then B := (DS:4A10h = 0) and (右 <> 0) ? 18h : 01h
             else if 左 = 1 then
                 if DS:4A10h <> 0 then B := 03h
                 else B := (右 <> 0) ? 0Bh : 07h;
         end else begin
             if      DS:4A10h <> 0 then B := 09h
             else if 右 <> 0       then B := 05h
             else if 兩邊都空       then B := 11h
             else                       B := 13h;
         end;

      3: if      DS:4A0Eh = 0 then C := 16h
         else if DS:4A0Eh = 3 then C := 10h
         else if DS:4A0Eh = 1 then C := 0Ah;

      4: if DS:4A0Eh = 0 then begin
             if      左 = 0        then D := 16h
             else if DS:4A10h <> 0 then D := 04h
             else if 右 = 0        then D := 08h
             else                       D := 0Ch;
         end else begin
             if      DS:4A10h <> 0 then D := 0Eh
             else if 右 = 0        then D := 17h
             else                       D := 0Ah;
         end;
    end;

<本模組 0048h>(A, 0, 5);
<本模組 0048h>(B, 0, 6);
<本模組 0048h>(C, 1, 5);
<本模組 0048h>(D, 1, 6);
```

## 四個位置 × 一張圖磚編號

最後四個呼叫的後兩個參數是固定的 `(0,5)`／`(0,6)`／`(1,5)`／`(1,6)`——
**畫面上四個固定位置**，每個位置依上面算出來的編號選一張圖磚。
形狀上是 3D 地城視圖裡「左近／左遠／右近／右遠」四塊牆面。

用到的編號集合：`01h 03h 04h 05h 07h 08h 09h 0Ah 0Bh 0Ch 0Eh 0Fh 10h 11h 13h 16h 17h 18h`
——**`16h`（22）出現在四個位置的「什麼都沒有」分支**，
形狀上是空白／全黑那一張。

## 兩個全域是主要分歧點

`DS:4A0Eh`（PC-98 `DS:7AC4h`）值 0／1／3 三選一，
`DS:4A10h`（PC-98 `DS:7AC6h`）只判是不是 0。
四個位置的分支結構都以 `DS:4A0Eh = 0` 為第一層。
**本規格不宣稱這兩格是什麼**（形狀上是「正前方是什麼」與「有沒有門」）。

`本模組 0378h(方向, y, x)` 回 0／1／3 之類的小整數，
與 `DS:4A0Eh` 用同一組值域——形狀上是同一種「這一格是什麼」的查詢。

## PC-98：只差一條 `xor ah, ah`

DOS 把迴圈變數零擴充後 `cmp ax, n`，PC-98 直接 `cmp al, n`。
其餘 13 條是兩個全域的位址。**沒有邏輯差異。**

## 明確不宣稱

- 沒有宣稱 18 個圖磚編號各是什麼圖。
- 沒有宣稱 `本模組 0048h(編號, a, b)` 怎麼畫。
- 沒有宣稱 `DS:4A0Eh`／`DS:4A10h` 的名稱與由誰設定。
- 沒有宣稱父框架 `−1`／`−2` 哪個是 x 哪個是 y。
