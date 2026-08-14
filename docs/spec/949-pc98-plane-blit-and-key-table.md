# 949 — PC-98 的三平面貼圖，與按鍵設定表的展開

- 證據等級：`exact`（兩支逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `13E22h`、`1705Ah`
- 作法見 spec 783

## ★ `13E22h`（81 行，`retn 0Eh`）：三平面的表面貼圖

`(目的表面 arg_0, 來源表面 arg_4, 張號 arg_8, 列 arg_A, 欄 arg_C)`。

```pascal
高   := 目的^[0];
寬   := 目的^[2];
每平面 := 高 × 寬;

di := 目的 ＋ 17h ＋ (列 shl 3) × 目的^[2] ＋ 欄;      { 列一單位 ＝ 8 掃描線 }
列差 := 目的^[2] − 來源^[2];
每列字數 := 來源^[2] div 2;
si := 來源 ＋ 17h ＋ 來源^[11h] × 張號;

for 平面 := 1 to 3 do begin                            { ★ 三個平面 }
    for y := 1 to 來源^[0] do begin
        for x := 1 to 每列字數 do begin
            word(es:di) := word(ds:si);
            si := si ＋ 2;   di := di ＋ 2;
        end;
        di := di ＋ 列差;
    end;
    di := 平面起點 ＋ 每平面;
    平面起點 := di;
end;
```

**目的表面裡三個平面連續存放，每個平面 `高 × 寬` bytes**；
來源則是連續讀下去（`si` 不重設），所以來源的一張圖也是三個平面連著。

`17h` 標頭與 `+0`／`+2`／`+11h` 三個欄位與 spec 934／947 一致。
列的單位同樣是 **8 條掃描線**（`列 shl 3`），與 spec 937／940／942 一致。

PC-98 的圖形是 GDC 的平面式版面，這一支一次搬三個平面。
（PC-98 的 16 色是 4 個平面 B／R／G／I，本支只動三個——
本規格不宣稱第四個平面由誰處理。）

⚠ 每列以 **word** 為單位複製（`每列字數 := 寬 div 2`），
所以**寬是奇數時最後一個 byte 不會被搬**。本支沒有處理奇數寬。

## `1705Ah`（79 行）：把按鍵設定字串展開成一張 word 表

```pascal
FillChar(ds:28A4h, byte_1ED34h 個 word, 0);            { 先清空 }
si := 28A4h;
來源 := 遠指標(dword_1EB2Dh);
n := 來源^[0];                                          { 長度前綴 }
i := 1;
while n > 0 do begin
    c := 來源^[i];  inc(i);  dec(n);
    if <sub_17172h>(c) 進位 then begin
        word(ds:si) := c;                               { 單 byte，高位補 0 }
    end else begin
        ah := 來源^[i];                                 { 取第二個 byte }
        k := <sub_17148h>(ax, ds:2AADh);                { 在表裡找 }
        word(ds:si) := word[2AE3h ＋ (k − 1) × 2];      { ★ 查 word 表 }
        if n > 0 then begin dec(n);  inc(i) end;
    end;
    si := si ＋ 2;
end;
```

`dword_1EB2Dh` 指向一個**長度前綴的位元組串**（設定或按鍵定義），
逐項展開成 `ds:28A4h` 起的 word 陣列，長度上限 `byte_1ED34h`。
兩個 byte 的項目要先在 `ds:2AADh` 那張表裡查出編號，
再用編號索引 `ds:2AE3h` 的 word 表取值。

這張展開後的表就是 spec 945 的 `16E70h` 按鍵分派迴圈
所用的 `遠指標(dword_16A19h)`／`ds:2AA6h` 那一類——
**按鍵對應是資料驅動的，不是寫死在程式碼裡**。

## 明確不宣稱

- 沒有宣稱第四個圖形平面（I）由誰處理。
- 沒有宣稱 `<sub_17172h>`（判單／雙 byte）與 `<sub_17148h>`（查表）的內部。
- 沒有宣稱 `ds:2AADh`／`ds:2AE3h` 兩張表的內容與筆數。
- 沒有宣稱 `dword_1EB2Dh` 指向的資料是按鍵定義還是別的設定。
