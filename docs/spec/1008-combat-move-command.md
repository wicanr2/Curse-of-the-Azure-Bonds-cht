# 1008 — 戰鬥移動指令：小鍵盤八方向、半點制的移動力、斜走貴一半

- 證據等級：`exact`（DOS 側 359 條逐條讀完；PC-98 對側 369 條，
  多出來的十條已逐條讀完並對照）
- 作法見 spec 783

## `dos overlay-08:00B06h` ↔ `pc98 overlay-08:00B20h`（`retf 0Ah`）

兩側原本都是待解讀。

### 簽章

`retf 0Ah` ＝ 10 bytes。從被呼叫者自己的堆疊框看（IDA 的 `arg_8` 其實是 `bp+0Ch`）：

```pascal
procedure 移動指令(角色: 遠指標;      { bp+0Ch..0Fh }
                   指令: byte;        { bp+0Ah }
               var 結果: byte);       { bp+06h..09h，遠指標 }
```

## ★★ 八個方向是小鍵盤掃描碼

```pascal
case 鍵 of
  48h: 方向 := 0;   { ↑     }
  49h: 方向 := 1;   { PgUp  ↗ }
  4Dh: 方向 := 2;   { →     }
  51h: 方向 := 3;   { PgDn  ↘ }
  50h: 方向 := 4;   { ↓     }
  4Fh: 方向 := 5;   { End   ↙ }
  4Bh: 方向 := 6;   { ←     }
  47h: 方向 := 7;   { Home  ↖ }
else   方向 := 8;   { ＝原地／無效 }
end;
```

★ **`48h`/`49h`/`4Dh`/`51h`/`50h`/`4Fh`/`4Bh`/`47h` 是 IBM PC 小鍵盤的擴充掃描碼**
（↑／PgUp／→／PgDn／↓／End／←／Home），而它們對應的 0..7 與 spec 999／991
的方向編碼（0 ＝ 上，順時針）**完全一致**。
方向表在建構端、消費端、幾何端之外，這裡是**輸入端**的第四個獨立確認。

## ★ 移動力是「半點」制

顯示用的字串是

```pascal
'Move/Attack, Move Left = ' ＋ Str(角色^[18Dh]^[6] div 2) ＋ ' '
```

而每一步的花費是

```pascal
成本 := byte[26A2h ＋ 地形碼 × 4] × (if odd(方向) then 3 else 2);
if byte[26A2h ＋ 地形碼 × 4] = 0FFh then 成本 := 0FFh;   { 不能走 }
```

> **`+18Dh`（戰鬥狀態）的 `+6` 是移動力，單位是「半點」——顯示時除以 2。
> 正交一步花地形成本的 2 倍，斜走花 3 倍，也就是斜走貴 50%。**

`DS:26A2h` 是 spec 994 的地磚查表（筆距 4，`0FFh` ＝ 不能放／不能走）。

離開時還有一道收尾：

```pascal
if 角色^[18Dh]^[6] < 2 then 角色^[18Dh]^[6] := 0;
```

★ **剩不到一整點就直接歸零**——半點制下「剩 0.5 點」不能走任何一步，
所以乾脆清掉。主迴圈的條件也是 `> 1`。

## 主迴圈

```pascal
原移動力 := 角色^[18Dh]^[6];
原朝向   := 角色^[18Dh]^[9];
x := <overlay-32 entry#15>(角色);   y := <overlay-32 entry#16>(角色);
結果 := 0;  方向 := 8;

while (角色^[18Dh]^[6] > 1) and (鍵 <> 0) and (鍵 <> 0Dh) do begin
    <清矩形>(18h, 27h, 18h, 18h);                  { resident 01A0:04CDh }

    if 鍵 = 20h then                                { ★ 空白鍵才開選單 }
        鍵 := <選單>(…, 'Move/Attack, Move Left = ' ＋ Str(移動力 div 2) ＋ ' ', …);

    if 鍵 = 0 then begin                            { 選單取消 }
        角色^[18Dh]^[6] := 原移動力;                { 還原 }
        <overlay-32 entry#10>(0, 0, <overlay-32 entry#18>(角色));
        結果 := ord(<overlay-32 entry#21>(角色, x, y, 0) = 0);
        <overlay-32 entry#13>(entry#15(角色), entry#16(角色), 0, 8);
        角色^[18Dh]^[9] := 原朝向;
        方向 := 8;
    end
    else 方向 := <上面那張掃描碼表>(鍵);

    if 方向 < 8 then begin
        <overlay-32 entry#14>(0, 0, 方向, 角色);
        <overlay-32 entry#19>(角色, 方向, @佔用者, @地形碼, @c, @d);

        if 佔用者 > 0 then                          { 那一格有人 → 攻擊 }
            <sub_EE9>(角色, 遠指標(DS:6D35h ＋ 佔用者 × 4), @結果)

        else if 地形碼 = 0 then begin               { ★ 走出戰場邊界 }
            顯示 'Flee:';
            case <是非選單>() of
              'Y': 結果 := <overlay-13 entry#7>(角色);   { 逃跑 }
              'N': 結果 := 0;
            end;
        end

        else if 成本 > 角色^[18Dh]^[6] then
            顯示 "can't go there"                   { <overlay-24 entry#19> }

        else begin
            <overlay-13 entry#6>(方向, 角色);       { 走一步 }
            if 角色^[196h] = 0 then 結果 := <overlay-24 entry#34>(角色)
            else begin
                if 角色^[18Dh]^[6] > 0 then <overlay-13 entry#5>(方向, 角色);
                if 角色^[196h] = 0 then 結果 := <overlay-24 entry#34>(角色);
                <overlay-23 entry#5>(1, 角色);
                if <overlay-24 entry#6>(角色) <> 0 then
                    結果 := <overlay-24 entry#34>(角色);
            end;
        end;
    end;

    if (鍵 <> 0) and (鍵 <> 0Dh) then 鍵 := 20h;    { ★ 下一圈自動回到選單 }
end;
```

★ **`DS:6D35h` 是戰鬥員遠指標表**（spec 805／851 同一張），
`佔用者 × 4` 取出目標——**走進有人的格子就是攻擊**，這就是選單標題
`'Move/Attack'` 的來由。

★ `角色^[196h]` 非 0 才走後半段（第二次移動、效果結算、`entry#6` 的再檢查），
形狀上是「還能行動」旗標（spec 576 的 `STANDUP` 也寫這一格）。

## ⚠ PC-98 多做兩件事

DOS 359 條、PC-98 369 條，多出來的十條全部集中在兩處：

**一、開頭多叫一次十筆欄位初始化**

```asm
0B27  call far ptr 164h:2Fh      ; overlay-26 entry#3 @0858h（spec 755）
```

那一支把 `DS:0A338h ＋ i × 7`（i = 1..10）的記錄清成範本，
並把 `DS:0A334h ＋ i` 全設成空白（`20h`）。

**二、選單前把十個熱鍵全設成 `0Dh`**

```asm
mov di, offset 06A6h  / push ds / push di      ; 來源：DS:06A6h
mov di, 0A335h        / push ds / push di      ; 目的：DS:0A334h ＋ 1
mov ax, 0Ah           / push ax                ; 10 bytes
call far ptr 0A65h:262h                        ; rep movsb（spec 990）
```

`DS:06A6h` 起的十個 byte 靜態值是 **`0D 0D 0D 0D 0D 0D 0D 0D 0D 0D`**。

> ★★ **PC-98 把這個選單的十個熱鍵全部覆蓋成 `0Dh`（Enter），等於關掉字母熱鍵。**
> DOS 完全沒有這兩段——它靠選項行裡的大寫字母標熱鍵（spec 977／978）。

**中文化必須跟 PC-98 這一版**：Big5 的次位元組落在 `41h..5Ah`，
用大寫字母當熱鍵會把漢字的第二個 byte 誤判成熱鍵（spec 993）。
熱鍵表在 `DS:0A334h ＋ i`，是逐格可改的。

## 其他兩平台差異

| | DOS | PC-98 |
|---|---|---|
| 戰鬥狀態欄位 | 角色 `+18Dh` | 角色 **`+18Eh`** |
| 選單 routine | `overlay-26 entry#3` | `overlay-26 entry#5` |
| 選單參數常數 | `0Ah, 0Ah, 0Fh, 1, 0` | `0Ah, 0, 0, 1, 0` |
| 選項緩衝 | `DS:75DBh` | `DS:0A86Dh` |
| 清矩形 | `resident 01A0:04CDh` | `resident 019E:06B5h` |

⚠ `+18Dh` ↔ `+18Eh` 是 spec 641 記過的同一類**一個 byte 的欄位位移**，
**引用前各自確認**。

## 中文化

四個字串都在 overlay code 段：

| 字串 | 長度 | 說明 |
|---|---|---|
| `'Move/Attack, Move Left = '` | 25 | 後面接數字與一個空白 |
| `'Flee:'` | 5 | 是非選單的提示 |
| `"can't go there"` | 14 | 走不動的訊息 |

⚠ `'Move/Attack, Move Left = '` 之後**接的是執行期組出來的數字**，
中文譯法要讓數字留在句尾（例如「移動／攻擊，剩餘移動力 ＝ 」），
不能把數字塞到句中。

## 明確不宣稱

- 沒有宣稱 `overlay-32` 的 `entry#15`／`16`／`18`／`19`／`21` 各回什麼
  （形狀上 15／16 是取座標，19 一次回四個值）。
- 沒有宣稱 `sub_EE9`（同模組）怎麼打。
- 沒有宣稱 `overlay-13` 的 `entry#5`／`6`／`7` 內部行為
  （`entry#7` 由 `'Flee:'` 的 Y 觸發，形狀上是逃跑）。
- 沒有宣稱 `entry#19` 回的第三、第四個值（`@c`／`@d`）是什麼，本支沒有讀。
- 沒有宣稱 PC-98 為什麼選 `0Dh` 而不是別的鍵當那十格的值。
