# 935 — CGA 8×8 字元格的直接繪製，與「刪掉所有出現的子字串」

- 證據等級：`exact`（逐條讀完）
- 位置：DOS `START.EXE` 的 `16147h`、`168A9h`
- 作法見 spec 783

## `16147h`（78 條）：把一個 8×8 字模畫進 CGA 畫面 —— `retn 0Ah`

```pascal
di := 列 × 140h ＋ 欄 × 2;
dl := byte[256Dh ＋ 前景];              { ★ 同 spec 932 的顏色轉換表 }
dh := byte[256Dh ＋ 背景];
bx := 0;   ah := 1;

for 對 := 1 to 4 do begin               { 8 條掃描線 ＝ 4 偶 ＋ 4 奇 }
    es := 0B800h;                        { 偶數列 }
    for j := 1 to 2 do begin             { 每列 2 bytes ＝ 8 像素 }
        inc(di);
        al := 0;
        for k := 1 to 4 do begin         { 每 byte 4 個像素 }
            if byte[6598h ＋ bx] and ah <> 0 then al := al or dl
            else                                   al := al or dh;
            rol(dl, 2);  rol(dh, 2);  rol(ah, 1);
        end;
        es:[di] := al;   dec(di);
    end;
    inc(bx);   inc(di);

    es := 0BA00h;                        { 奇數列 }
    { 同樣的兩個 byte }
    inc(bx);
    di := di ＋ 51h;
end;
```

### 這是 CGA 320×200 四色

- **`0B800h` 與 `0BA00h` 兩個交錯的半頁**——CGA 圖形模式把偶數列與奇數列
  分開放，相差 `2000h`。
- **每個 byte 4 個像素**（`for k := 1 to 4`），也就是 **2 bits／像素 ＝ 四色**。
  `rol dl, 1` 做兩次就是把 2 bit 的顏色轉到下一個像素的位置。
- `140h` ＝ 320 ＝ **4 列 × 80 bytes**，正好是一個字元格（8 掃描線）在
  單一半頁裡佔的距離。
- 字模來自 **`DS:6598h`**，`bx` 每條掃描線遞增一次，8 條線用 8 bytes。

**前景與背景都要過 `DS:256Dh` 轉換表**——與 spec 932 的字元輸出、
spec 934 的顯示模式旗標是同一套配色機制。

`1618Eh` 的 `db 90h` 與 `161C3h` 的 `align 2` 是對齊填充，不是程式碼。

## `168A9h`（82 條）：刪掉所有出現的子字串 —— `retf 8`

```pascal
刪子字串(要刪的 arg_0, 主字串 arg_4, 結果 arg_8);

s := 限長(arg_4, 0FFh);
t := 限長(arg_0, 0FFh);
n := length(s) − length(t) ＋ 1;
for i := 1 to n do
    if Copy(s, i, length(t)) = t then
        Delete(s, i, length(t));
arg_8^ := 限長(s, 0FFh);
```

用的是 Turbo Pascal 的 `Copy`／`Delete`／字串比較 RTL
（`@Copy$qm6Stringt17Integert3`、`@Delete$qm6String7Integert2`、
`@$bsub$qm6Stringt1`）。

⚠ **上界 `n` 在迴圈開始前算好一次，刪除之後不重算**，而 `i` 也照樣往前。
所以：

- 刪掉一段之後，`i` 不會退回去，**緊接著形成的新出現不會被刪**
  （例如從 `'aabb'` 刪 `'ab'` 只會刪一次）。
- `i` 會掃到超過縮短後的字串尾端，但 `Copy` 在越界時回傳較短的字串、
  比對不中，所以**不會出錯，只是白跑**。

## 明確不宣稱

- 沒有宣稱 `16147h` 的 `arg_0`（本支沒有讀它）。
- 沒有宣稱 `DS:6598h` 那 8 bytes 由誰填、字模從哪來。
- 沒有宣稱 `DS:256Dh` 轉換表的內容。
- 沒有宣稱 `168A9h` 由誰呼叫、拿來刪什麼。
