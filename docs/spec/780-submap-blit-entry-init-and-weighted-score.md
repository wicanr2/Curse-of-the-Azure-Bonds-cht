# 780 — 區域地圖的分塊貼圖、清單項目初始化、依代碼加權的累計

- 證據等級：`exact`（以補洞後的匯出逐條讀完，見 spec 761）

## `overlay-30:0448h`（DOS）— 依形狀表把一塊區域畫出來

`retf 8`：

```pascal
欄 := byte[0AD8h + 形狀];
下界 := byte[0AE2h + 形狀] + y0 − 1;
右界 := byte[0AECh + 形狀] + x0 − 1;
for x := x0 to 右界 do begin
    for y := y0 to 下界 do begin
        if (0 <= x <= 0Ah) and (0 <= y <= 0Ah) then begin
            v := DS:7202h^[(列 − 1) * 9Ch + 欄];
            if v > 0 then <020Bh:002Fh>(y + 2, x + 2, v, 1, 1);
        end;
        inc(欄);
    end;
end;
```

- `DS:0AD8h`、`DS:0AE2h`、`DS:0AECh` 是**三張各 10 bytes 的平行表**（起始欄、
  高度、寬度），以第一個參數當索引，所以「形狀」最多 10 種。
- 資料在 `DS:7202h` 指向的陣列，**每列 `9Ch`（156）bytes**，列索引 1 起算。
- 螢幕範圍檢查是 `0..0Ah`（0..10）——**超出就跳過該格，不畫也不中斷**。
- `欄` 在**內圈**遞增，所以資料是逐列（y 方向）連續存放的。
- 傳給繪圖程序的座標各 `+3 − 1`（＝ `+2`）。

## PC-98 `overlay-16:402Ah` — 清單項目初始化

`retf 4`。記錄陣列基底 `DS:0A648h`、**每筆 `29h`（41）bytes**，索引來自
`參數^[0]`：

```pascal
i := 參數^[0];
記錄[i].長度 := 14h;                            { 20 }
FillChar(記錄[i].內容, 記錄[i].長度, ' ');       { 20 個半形空白 }
本模組 45C9h(0, @記錄[i], @本地緩衝);
StoreString(→ 記錄[i], 28h);
記錄[i][+09h] := 87h;
記錄[i][+0Ah] := i + 3Fh;
if byte[0A80Ah + i] <> 5 then
    Move(@[byte[0A80Ah + i] * 15h + 6812h],
         @記錄[i][+0Dh],
         byte[byte[0A80Ah + i] * 15h + 6812h]);
```

- 每筆的 `+0` 是 Pascal 短字串的長度 byte，初始化成 **20 個空白**——所以項目
  一開始是固定寬度的空白欄位。
- `+09h` 固定填 `87h`、`+0Ah` 填 `i + 3Fh`（依索引遞增的識別碼）。
- `DS:6812h` 起是**每筆 `15h`（21）bytes** 的另一張表，由 `DS:0A80Ah + i` 選；
  值為 `5` 時整段跳過。搬的長度取自該筆的第一個 byte（也是 Pascal 字串長度）。

**中文化要注意**：`14h` 個空白是寫死的欄寬，全形字會讓實際顯示寬度變成兩倍。

## PC-98 `overlay-23:17A5h` — 依代碼加權累計

`retf 6`，第二個參數是 SS 相對的記錄位址：

```pascal
上限 := byte[6F4Dh + 等級];
if 上限 <= n then n := 上限 − 1;                       { 無號 }
if 等級 = 4 then
    if (rec^[8]^[0E6h] = 0) or (rec^[8]^[115h] = rec^[8]^[0E6h]) then inc(n);
c := ss:[rec − 1];
if 等級 in {2, 3, 4} then begin
    if      0Fh <= c <= 13h then ss:[rec − 0Ah] += n * (c − 0Eh)   { 1..5 }
    else if c = 14h         then ss:[rec − 0Ah] += n * 5
    else if 15h <= c <= 17h then ss:[rec − 0Ah] += n * 6
    else if 18h <= c <= 19h then ss:[rec − 0Ah] += n * 7;
end else begin
    if      c > 0Fh then ss:[rec − 0Ah] += n * 2
    else if c = 0Fh then ss:[rec − 0Ah] += n
    else if c < 4   then ss:[rec − 0Ah] += n * (−2)
    else if c < 7   then ss:[rec − 0Ah] += n * (−1);
end;
```

- 乘法是**有號**的（`imul`），所以 `c < 7` 的兩段會讓累計值**減少**。
- 累計目標 `ss:[rec − 0Ah]` 是 **byte** 加法，沒有上下限檢查。
- `等級 in {2,3,4}` 與其他走完全不同的兩張權重表；`等級 = 4` 另外還有一個
  「條件成立就 `n` 加一」的加成。
- 上限表在 `DS:6F4Dh`，以等級索引；**夾制的寫法是 `n := 上限 − 1` 而不是
  `n := 上限`**，所以實際可用值比表上的數字少一。
- `c` 落在 `07h`..`0Eh` 時（非 2/3/4 等級那條路）**四個條件都不成立**，累計值
  不變。

## 明確不宣稱

- 沒有宣稱 `DS:7202h` 那個 156-byte 一列的陣列裝的是什麼。
- 沒有宣稱 `020Bh:002Fh` 畫的是什麼。
- 沒有宣稱 `記錄[i][+09h] = 87h` 與 `+0Ah = i + 3Fh` 的用途。
- 沒有宣稱 `overlay-23:17A5h` 的「等級」「代碼 `c`」「累計值」各是什麼；只確定
  權重表與有號乘法。
