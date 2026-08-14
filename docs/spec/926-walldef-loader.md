# 926 — 載入 `WALLDEF<磁碟>.dax`：搬進牆面圖庫並重定索引

- 證據等級：`exact`（兩平台都逐條讀完；相似度 0.938）
- 作法見 spec 783

## `overlay-30:0104Ch`（DOS 285 條）↔ `overlay-30:00F8Fh`（PC-98 268 條）

`retf 4` ＝ `(arg_0, arg_2)` 兩個 byte。

| | DOS |
|---|---|
| 檔名 | `'WALLDEF'` |
| 副檔名 | `'.dax'` |
| 失敗訊息 | `'Unable to load '` ＋ `arg_2` ＋ `' from WALLDEF'` ＋ 磁碟編號 ＋ `'.'` |

PC-98 的失敗訊息只有一個字串常數（`unk_F61`）加上一個 `'.'`，**不組合磁碟編號**。

```pascal
if (arg_0 < 1) or (arg_0 > 3) then 離開;             { 無號 }

Str(DS:5BEEh, 0, @磁碟, 1);                          { 磁碟編號，spec 892 預設 1 }
檔名 := 'WALLDEF' ＋ 磁碟;
副檔 := '.dax';
<0636h:08DEh>(@緩衝, @大小, arg_2);                  { 載入 }

if (大小 = 0) or (大小 div 30Ch ＋ arg_0 > 4) then begin
    <0297h:0107h>();
    WriteLn(DS:8DB2h, 'Unable to load ', arg_2,
            ' from WALLDEF' ＋ 磁碟 ＋ '.');
    <06EAh:010Dh>();                                 { 等按鍵 }
    <0297h:0107h>();
    <06EAh:0000h>();                                 { ★ 直接結束程式 }
end;

差 := word[2680h ＋ arg_0 × 2] − word[2682h];

{ 逐塊 30Ch bytes 搬進圖庫 }
偏 := 0;   n := 1;
repeat
    Move(緩衝 ＋ 偏,
         遠指標(DS:7202h) ＋ (arg_0 ＋ n − 2) × 30Ch,
         30Ch);
    偏 := 偏 ＋ 30Ch;   inc(n);
until 偏 >= 大小;                                     { 無號 }
FreeMem(@緩衝, 大小);                                 { ★ PC-98 沒有這一步 }

塊數 := n − 1;
for i := 1 to 塊數 do begin
    idx := arg_0 ＋ i − 1;
    if idx in [1, 2, 3] then begin                    { 32 bytes 的集合常數 }
        word[7210h ＋ idx × 4] := 0FFFFh;
        word[7212h ＋ idx × 4] := 0FFFFh;

        for b := 1 to 5 do                            { 每塊 5 個 9Ch bytes 的子塊 }
            for f := 0 to 9Bh do begin
                q := 遠指標(DS:7202h) ＋ (idx − 1) × 30Ch ＋ (b − 1) × 9Ch ＋ f;
                if q^ >= DS:268Ah then q^ := q^ ＋ lo(差);
            end;

        if 塊數 > 1 then <020Bh:002Ah>(idx, arg_2 × 0Ah ＋ i)
        else               <020Bh:002Ah>(idx, arg_2);
    end;
end;

word[7210h ＋ arg_0 × 4] := arg_2;
word[7212h ＋ arg_0 × 4] := arg_0;
```

## 圖庫的分塊

`遠指標(DS:7202h)` 是牆面圖庫，**每塊 `30Ch` ＝ 780 bytes，每塊再分成
5 個 `9Ch` ＝ 156 bytes 的子塊**（5 × 156 ＝ 780 ✓）。
塊編號 `idx` 從 1 起算，實際位址是 `(idx − 1) × 30Ch`——
搬運時寫成 `idx × 30Ch ＋ 0FCF4h`（`0FCF4h` ＝ −`30Ch`），
重定索引時寫成 `idx × 30Ch ＋ b × 9Ch − 3A8h`，兩者化簡後一致。

上限 4 塊（`大小 div 30Ch ＋ arg_0 > 4` 就失敗）。

## 重定索引

子塊裡每一個 byte，**只要 `>= DS:268Ah` 就加上 `差`**，
`差 := word[2680h ＋ arg_0 × 2] − word[2682h]`。

`DS:2680h` 是一張 word 表（索引 `arg_0`），`DS:2682h` 是它的第 1 筆。
所以 `arg_0 = 1` 時差是 0，其餘是相對於第 1 筆的位移。
`DS:268Ah` 是門檻——**低於門檻的 byte 不動**（形狀上是共用的圖塊，
門檻以上才是這一組專屬的）。

## 32 bytes 的集合常數 ＝ `[1, 2, 3]`

`unk_102Ch` 的第一個 byte 是 `0Eh` ＝ `0000 1110`，其餘 31 bytes 全 0，
所以集合是 **`[1, 2, 3]`**——與函式開頭 `1 ≤ arg_0 ≤ 3` 的守衛一致。
用 `0A54h:08D4h`（PC-98 `0A65h:08E4h`）測試，ZF 回報。

## ⚠ 載入失敗 ＝ 結束程式

沒有回退路徑：印訊息、等一次按鍵、`<06EAh:0000h>`。
`06EAh:0000h` 與 spec 924 紮營選單的「Quit TO DOS」是**同一個離開入口**。

## ⚠ PC-98 沒有那一次 `FreeMem`

DOS 在搬完之後 `FreeMem(@緩衝, 大小)`；PC-98 在**同一個位置**換成

```
mov al, 0
out 0A6h, al
```

（形狀上是 PC-98 的繪圖頁選擇 port）。**PC-98 在本支裡沒有釋放那塊緩衝。**

PC-98 的堆疊也小得多（`sub sp, 20h` vs DOS `sub sp, 118h`），
因為它不組合那幾個長字串。

## 中文化

`'Unable to load '` 與 `' from WALLDEF'` 是**載入失敗的致命訊息**，
出現時遊戲會直接結束。`'WALLDEF'` 與 `'.dax'` 是檔名，**不可翻譯**。

## 明確不宣稱

- 沒有宣稱 `arg_0`（1..3）與 `arg_2` 各代表什麼。
- 沒有宣稱 `DS:2680h` 那張表與 `DS:268Ah` 門檻的實際內容。
- 沒有宣稱 `word[7210h ＋ idx × 4]`／`word[7212h ＋ idx × 4]` 那兩張表的用途
  （先寫 `0FFFFh` 再在結尾寫回實值，形狀上是「這塊正在重建」的標記）。
- 沒有宣稱 `<020Bh:002Ah>`／`<0636h:08DEh>`／`<0297h:0107h>` 的內部行為。
- 沒有宣稱 PC-98 是否在別處釋放那塊緩衝（本支沒有）。
