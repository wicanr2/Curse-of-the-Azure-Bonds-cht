# 931 — 常駐的堆積管理器與 Crt 緩衝輸出

- 證據等級：`exact`（逐條讀完）
- 位置：DOS `START.EXE` 的 `19FB3h`、`1A374h`、`1A986h`、`1AA22h`
- 作法見 spec 783

四支都沒有 Pascal 序言，以組語手寫、`retn` 收尾。

## `19FB3h`（73 條）：Crt 的緩衝輸出

```pascal
基準列欄 := word[0040:0050h];              { BIOS 游標位置 }
si := di;   bx := dx;
repeat                                      { cx 個字元 }
    ch := es:[di];
    case ch of
      07h: begin 沖出(); <int 10h AX=0E07h>(); end;   { BEL }
      08h: begin 沖出(); if dl <> 左界 then dec(dl) end;
      0Ah: begin 沖出(); 捲一行() end;
      0Dh: begin 沖出(); dl := 左界 end;
    else
        inc(di);  inc(dl);
        if dl > 右界 then begin
            沖出();  捲一行();  dl := 左界;
        end;
    end;
until 迴圈結束;
沖出();

word[0040:0050h] := dx;                     { 寫回 BIOS 游標 }
偏移 := dh × byte[0040:004Ah] ＋ dl;
{ 直接寫 CRTC：0Eh／0Fh 是游標位置的高／低位 }
port := word[0040:0063h];
out port, 0Eh;  out port+1, hi(偏移);
out port, 0Fh;  out port+1, lo(偏移);
```

「沖出」是 `sub_1A050`（spec 930 的區塊搬移）——**普通字元先累積成一段，
碰到控制字元或行尾才一次搬進顯示記憶體**，所以整行輸出只有一次 `rep movsw`。

游標位置**同時寫 BIOS 資料區與 CRTC 暫存器**，不走 `int 10h AH=02h`。
`jmp short $+2` 出現四次，是 8088 時代的 I/O 延遲慣用法（不是死碼）。

## 堆積管理器：三支

Turbo Pascal 6 的堆積用 **`seg:para` 兩個 word 表示位址**，所以到處看得到
這個正規化樣式：

```
add ax, cx        ; para
cmp ax, 10h
jb  短路
sub ax, 10h
inc dx            ; segment 進位
```

全域變數（由用法反推）：

| 位址 | 角色 |
|---|---|
| `word_2096Ah` | 堆積起點 |
| `word_2096Ch` | 目前配置指標（段） |
| `word_2096Eh` | 堆積上限 |
| `word_20970h` | 已配置區塊鏈的頭（節點 `+14h` 是 next） |
| `dword_2097Eh` | free list 的遠指標 |
| `word_2097Ah`／`word_2097Ch` | 目前配置指標（para／段） |
| `word_20976h`／`word_20978h` | 下界 |

free list 的每一筆是 **8 bytes**：`+0`／`+2` 是起點的 para／段，
`+4`／`+6` 是終點的 para／段。

### `1A374h`（74 條）：釋放並回收

走 `word_20970h` 那條鏈，把**位置高於目前配置指標**的區塊逐一釋放
（`sub_1A4E3h`），直到要求的大小騰出來為止；接著把鏈上剩下的區塊
往下搬（`sub_1A495h`）並修正 `word_2096Ch`。
最後把新區塊接到鏈尾——`bx := 47B0h` 起，沿著 `+14h` 走到 next 為 0 才接上。

### `1A986h`（62 條）：從 free list 取一段

掃 `dword_2097Eh` 的每一筆，找**大小足夠**（`+6:+4` 減 `+2:+0` ≥ 要求）的一筆；
找到就從那一筆的起點切走，切完剛好用光就叫 `sub_1AAF9h` 把該筆移除。
free list 走完還沒找到，就從 `word_2097Ah`／`word_2097Ch` 那個
「配置指標」往上切，切之前先確認不會超過 `sub_1AB0Eh` 給出的上界。

回傳用 **CF**：`clc` ＝ 成功。

### `1AA22h`（66 條）：把一段併回 free list

`ax = 0` 直接返回。先確認要還的區間落在 `[word_20976h:word_20978h,
word_2097Ah:word_2097Ch]` 之內，否則 `stc` 失敗返回。

然後掃 free list，**與相鄰的項合併**：起點接得上就把起點往前延、
終點接得上就把終點往後延，合併掉的那一筆用 `sub_1AAF9h` 移除。
若還的正好是配置指標的下緣，就直接把指標往回退（不佔 free list 一筆）；
否則用 `sub_1AAD5h` 取一筆新的 free list 位置寫進去。

## 為什麼記這四支

它們解釋了 `GetMem`／`FreeMem`（`0A54h:0329h`／`0A54h:0364h`）背後的行為：

- **堆積是 LIFO 傾向的**：`1A374h` 釋放時會把高位址的區塊一起丟掉。
  所以 spec 849 的角色解構子那種「先抄 next 再釋放」的寫法是必要的。
- **free list 會合併相鄰區塊**，所以碎片化不會無限惡化。
- 位址是 `seg:para`，**`FreeMem` 傳的大小必須與 `GetMem` 一致**——
  spec 832／849 那些 `3Fh`／`67h`／`2Eh`／`56h` 的節點大小因此是硬性的。

## 明確不宣稱

- 沒有宣稱 `sub_1A42Ch`／`sub_1A438h`／`sub_1A495h`／`sub_1A4E3h`／`sub_1AAD5h`／
  `sub_1AAF9h`／`sub_1AB0Eh`／`sub_1AB57h` 的內部。
- 沒有宣稱 `bx := 47B0h` 那個起點是什麼結構。
- 沒有宣稱七個全域變數的原始 Pascal 名稱（角色由用法反推）。
