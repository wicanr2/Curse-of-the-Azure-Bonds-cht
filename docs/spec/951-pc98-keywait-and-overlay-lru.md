# 951 — PC-98 的等按鍵（沒有 STING、多三個切換鍵）與覆疊 LRU

- 證據等級：`exact`（兩支逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `18036h`、`1A1E9h`
- 作法見 spec 783

## ★ `18036h`（101 行，far）：等一個按鍵

DOS 的對應是 spec 936 的 `16FADh`。**兩邊差很多。**

```pascal
鍵 := <sub_192A0h>();

if 鍵 = 13h then begin                          { Ctrl-S：音效 }
    if byte_24DE8h <> 2 then begin
        byte_24DE9h := byte_24DE8h;
        <sub_18930h>(word_20ACAh);
        byte_24DE8h := 2;
    end
    else if byte_24DE9h in [0, 1] then begin    { 集合常數 byte_18016h }
        byte_24DE8h := byte_24DE9h;
        <sub_18930h>(word_20ACCh);
    end;
end;

if 鍵 = 0Fh then begin                          { ★ PC-98 專屬 }
    byte_24E73h := byte_24E73h xor 1;
    <sub_18AA7h>();
end;

if 鍵 = 02h then begin                          { ★ PC-98 專屬 }
    byte_24E7Eh := not byte_24E7Eh;
    byte_24E80h := 1;
end;

if 鍵 = 16h then begin                          { ★ PC-98 專屬 }
    byte_24E82h := not byte_24E82h;
    <sub_18194h>(byte_24E82h);
end;

if 鍵 <> 0 then
    while <sub_19293h>() do 鍵 := <sub_192A0h>();   { 吃掉多餘按鍵 }
結果 := 鍵;
```

### 與 DOS 版的差異

| | DOS（spec 936） | PC-98（本支） |
|---|---|---|
| Ctrl-S 音效切換 | 有（`byte_21DAAh`／`byte_21DABh`） | 有（`byte_24DE8h`／`byte_24DE9h`） |
| 合法音效裝置 | 集合 `byte_16F7Dh` ＝ `[0, 1]` | 集合 `byte_18016h` ＝ **`db 3, 1Fh dup(0)`，同樣是 `[0, 1]`** |
| **`STING` 命令列後門** | **有**（Ctrl-D／Ctrl-Z／Ctrl-C、`debug.txt`） | **完全沒有** |
| 非阻塞模式 | 有（`byte_2117Ch`） | 沒有這一段 |
| **`0Fh`／`02h`／`16h` 三個切換鍵** | 沒有 | **有** |
| 吃掉多餘按鍵 | 有 | 有 |

**PC-98 移植時把開發者後門整段拿掉，換上三個玩家可用的切換鍵。**
三個鍵各自 toggle 一個全域並叫一支處理常式（`02h` 那個只設旗標，
由別處去讀 `byte_24E80h`）。

## `1A1E9h`（92 行）：覆疊段的載入與淘汰

```pascal
inc(word_23AC8h);                               { 載入請求計數 }
if es^[10h] <> 0 then begin                     { 已經在記憶體裡 }
    es^[12h] := 1;                              { 重設「最近用過」 }
    goto 收尾;
end;

inc(word_23ACAh);                               { 實際載入計數 }
需要 := (es^[0Ah] ＋ 0Fh) shr 4 ＋ <sub_1A3FCh>();
<sub_1A3E1h>();
while 需要 > 0 do begin                          { ★ 騰空間 }
    <sub_1A29Ah>();
    節點 := word_23AE2h;
    word_23AE2h := 節點^[14h];                   { 走鏈 }
    if 節點^[12h] = 0 then begin                 { 沒被用過 → 丟掉 }
        <sub_1A31Ch>();
        節點^[10h] := 0;
        騰出 := <sub_1A3FCh>();
    end else begin                               { 用過 → 降一級，留著 }
        dec(節點^[12h]);
        <sub_1A34Fh>();  <sub_1A397h>();
        騰出 := 0;
    end;
    需要 := 需要 − 騰出;
end;

es^[10h] := word_23ADEh;                         { 標記成已載入 }
if <dword_28104h>(es) <> 0 then goto <sub_19EB0h>;   { 讀取失敗 }
<sub_1A397h>();

收尾:
<sub_1A2E6h>();
<sub_1A3E1h>();
{ 從 word_23AE2h 起再走一遍，累計大小到 word_23ACEh 為止，
  對 +12h ＝ 0 的節點叫 <sub_1A31Ch> }
```

★ **`+12h` 是 second-chance（clock）演算法的參照位元**：
命中時設成 1，淘汰掃描碰到非 0 就減 1 並跳過，碰到 0 才真的丟。
`+10h` 是「這一段在不在記憶體」、`+14h` 是 next、`+0Ah` 是段大小。

`word_23AC8h`／`word_23ACAh` 是**請求次數與實際載入次數**——
兩者相除就是覆疊快取的命中率，原版自己有在數。

## 明確不宣稱

- 沒有宣稱 `0Fh`／`02h`／`16h` 三個鍵各切換什麼功能。
- 沒有宣稱 `<sub_18930h>`／`<sub_18AA7h>`／`<sub_18194h>`／`<sub_192A0h>`／
  `<sub_19293h>` 的內部。
- 沒有宣稱 `word_23ACEh`（收尾迴圈的上界）與 `word_23ADEh` 的意義。
- 沒有宣稱 `<sub_1A29Ah>`／`<sub_1A2E6h>`／`<sub_1A31Ch>`／`<sub_1A34Fh>`／
  `<sub_1A397h>`／`<sub_1A3E1h>`／`<sub_1A3FCh>` 各做什麼。
