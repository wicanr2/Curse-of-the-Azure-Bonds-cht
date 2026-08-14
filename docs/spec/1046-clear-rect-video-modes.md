# 1046 — 清除矩形區域：`DS:4FE6h` 是顯示卡種類，CGA／EGA／Tandy 三條路

- 證據等級：`exact`（DOS 常駐段 205 條逐條讀完）
- 作法見 spec 783

## `dos START.EXE:11ECDh`（`retf 8`，472 bytes）

原本待解讀，呼叫端有三處（`11A00h`、`11CA5h`、`11DAAh`）。

> ★ **常駐段沒有「補洞匯出」，`annotated_dump.py` 只吃 overlay。**
> 本次改用 `objdump -D -b binary -m i8086`，檔案位移由既有結論換算：
> `file = ea − 10000h ＋ MZ 標頭長度`（DOS 是 1968）。
> `11ECDh → 267Dh`，開頭 `55 89 e5` ＝ `push bp; mov bp,sp` 對上，
> 換算正確。**常駐段的待解讀函式從此可以逐條讀。**

```pascal
procedure 清矩形(左: byte;   { bp+0Ch }
                 上: byte;   { bp+0Ah }
                 右: byte;   { bp+08h }
                 下: byte);  { bp+06h }
```

## ★★★ 三條路由 `DS:4FE6h` 分派

```pascal
case DS:4FE6h of
  0: …CGA…;
  1: …交給 far 0542:00BDh…;
  2: …Tandy／PCjr…;
end;
```

### 模式 0（`DS:4FE6h = 0`）

```pascal
for 列 := 上 × 4 to 下 × 4 ＋ 3 do begin
    FillChar(B800h:(列 × 50h ＋ 左 × 2), (右 − 左 ＋ 1) × 2, 0);
    FillChar(BA00h:(列 × 50h ＋ 左 × 2), (右 − 左 ＋ 1) × 2, 0);
end;
```

★ 每列 `50h` ＝ **80 bytes**、每個字元欄 **2 bytes**、
兩個交錯 bank `B800h`／`BA00h`、每個字元列 **4 條掃描列**
⇒ **CGA 320×200 4 色**（8 條掃描列 ÷ 2 bank ＝ 4）。

### 模式 1

```pascal
<far 0542:00BDh>(左, 上, 右, 下, 0, 0);
<far 0542:00BDh>(左, 上, 右, 下, 1, 0);
```

★ 呼叫兩次，只差第五個參數 `0`／`1`——形狀上是**兩個平面／兩個 page**。

### 模式 2（`DS:4FE6h = 2`）

```pascal
for 列 := 上 × 2 to 下 × 2 ＋ 1 do begin
    FillChar(B800h:(列 × 0A0h ＋ 左 × 4), (右 − 左 ＋ 1) × 4, 0);
    FillChar(BA00h:…, …, 0);
    FillChar(BC00h:…, …, 0);
    FillChar(BE00h:…, …, 0);
end;
```

★ 每列 `0A0h` ＝ **160 bytes**、每個字元欄 **4 bytes**、
**四個** 交錯 bank `B800h`／`BA00h`／`BC00h`／`BE00h`、每個字元列 **2 條掃描列**
⇒ **Tandy 1000／PCjr 320×200 16 色**（4 bank × 2 ＝ 8 條掃描列）。

> ★★★ **`DS:4FE6h` 是顯示卡種類：`0` ＝ CGA、`1` ＝ 走 `0542:00BDh` 的那一種
> （EGA／VGA）、`2` ＝ Tandy／PCjr。**
> spec 1037 看到「`byte_4FE6 = 0` 時顏色編輯被關掉、印 `'Sorry, not in CGA'`」
> ——現在知道那個判斷就是**在問「是不是 CGA」**。
>
> ★ 三條路的幾何一致：字元格一律 **8 × 8 像素**
> （CGA 2 bytes × 4 列 × 2 bank、Tandy 4 bytes × 2 列 × 4 bank）。

## 中文化

本支沒有字串。★ 但**字元格是 8 × 8**，這是中文化畫布規劃的硬約束：
全形字要 16 × 16 就得佔**兩格寬、兩格高**，
而清矩形的參數是**字元座標**（不是像素），呼叫端給的矩形都會跟著放大一倍。

## 明確不宣稱

- 沒有宣稱 `DS:4FE6h` 由誰、依什麼偵測結果設定。
- 沒有宣稱 `far 0542:00BDh` 的介面（只知道六個參數與最後兩個的值）。
- 沒有宣稱模式 1 是 EGA 還是 VGA（只確定不是 CGA 也不是 Tandy）。
- 沒有宣稱三個呼叫端（`11A00h`／`11CA5h`／`11DAAh`）各自清的是哪一塊。
