# 938 — 顯示模式 2 的字模繪製：Tandy 四半頁 16 色

- 證據等級：`exact`（逐條讀完）
- 位置：DOS `START.EXE` 的 `16246h`
- 作法見 spec 783

## `16246h`（131 條）—— `retn 0Ah`

`byte_211A6h = 2` 時由 spec 937 的分派層呼叫，簽章與
`16147h`（模式 0，spec 935）相同：`(索引, 前景, 背景, 列, 欄)`。

```pascal
di := 列 × 140h ＋ 欄 × 4;
dl := 前景;   dh := 背景;               { ★ 不過 DS:256Dh 轉換表 }
bx := 0;   ah := 1;

for 組 := 1 to 2 do begin                { 2 × 4 半頁 ＝ 8 條掃描線 }
    for 半頁 in [0B800h, 0BA00h, 0BC00h, 0BE00h] do begin
        es := 半頁;   di := di ＋ 3;
        for j := 1 to 4 do begin         { 每列 4 bytes }
            al := 0;
            for k := 1 to 2 do begin     { 每 byte 2 個像素 }
                if byte[6598h ＋ bx] and ah <> 0 then al := al or dh
                else                                   al := al or dl;
                rol(dl, 4);  rol(dh, 4);  rol(ah, 1);
            end;
            es:[di] := al;   dec(di);
        end;
        di := di ＋ 4;   inc(bx);
    end;
    di := di ＋ 0A1h;
end;
```

## 四個半頁、每像素 4 bits

| | 模式 0（`16147h`，spec 935） | 模式 2（本支） |
|---|---|---|
| 半頁 | `0B800h`／`0BA00h`（2 個） | `0B800h`／`0BA00h`／`0BC00h`／`0BE00h`（**4 個**） |
| 每 byte 像素數 | 4 | **2** |
| 每像素 bits | 2（四色） | **4（16 色）** |
| 每列 bytes | 2 | **4** |
| 欄位移 | `欄 × 2` | **`欄 × 4`** |
| 顏色 | 過 `DS:256Dh` 轉換表 | **直接用** |

四個交錯半頁是 **Tandy／PCjr 320×200×16 色**的版面。
`rol dl, 4` 一次轉 4 bit，正好把下一個像素的顏色送到低半位元組。

`add di, 3` 之後往回寫（`dec di`）——與模式 0 同樣是**由右往左填一列**。

這也解釋 spec 934 的「`byte_211A6h ≠ 0` 時每列位元組加倍」：
4 bits／像素是 2 bits／像素的兩倍。

## 兩個 Tandy 的證據互相印證

spec 929 讀到常駐支援 **Tandy SN76496 音效**（`out 0C0h`），
本支則是 **Tandy 16 色圖形**。兩者合起來，這份 DOS 版
確實把 Tandy／PCjr 當成一個完整的目標平台，不只是「順手支援音效」。

## 明確不宣稱

- 沒有宣稱模式 1（`sub_161E6h`）是哪一種硬體。
- 沒有宣稱 `0A1h` 這個列間位移怎麼來（形狀上是回到下一組掃描線的起點）。
- 沒有宣稱前景與背景在模式 2 為什麼不需要顏色轉換。
