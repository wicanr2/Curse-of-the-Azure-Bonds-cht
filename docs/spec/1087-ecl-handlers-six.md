# 1087 — 六個 ECL handler 的語意：`PROGRAM 8` 是通關開關、`TREASURE` 的 `operand8` 分三段

- 證據等級：`exact`（DOS 側六支合計 949 條，各自逐條讀完）
- 作法見 spec 783；指令名稱見 spec 1083；binding 見 spec 560

spec 560 把這六支的**指令碼綁定**釘好了，但註記都是「語意未解讀」。
本規格補上語意。指令名稱用 spec 1083 那張原作自己的助憶碼表。

## `dos overlay-02:00E13h`（35 條）＝ `2Ah` `GETTABLE`

```pascal
READVAR(3);
來源 := ADDFNC(high[1], low[1]);
偏移 := ADDRESSVALUE(2);
目的 := ADDFNC(high[2], low[2]);
STOREVALUE(目的, <ECL 記憶體讀取>(來源 ＋ 偏移));
```

⇒ **「從 ECL 記憶體的 `來源 ＋ 偏移` 讀一個 byte，寫進 `目的`」**——查表指令。

## `dos overlay-02:016D2h`（82 條）＝ `23h` `SURPRISE`

```pascal
READVAR(4);
for i := 1 to 4 do o[i] := ADDRESSVALUE(i);
甲 := o[4] ＋ 2 − o[1];
乙 := o[2] ＋ 2 − o[3];
r1 := ROLLDICE(1, 6);   r2 := ROLLDICE(1, 6);      { ★ 各擲一顆 d6，參數是（骰數, 面數），見 spec 1103 }
結果 := 0;
if r1 <= 甲 then begin
    if r2 <= 乙 then 結果 := 3 else 結果 := 1;
end;
if r2 <= 乙 then 結果 := 2;                        { ⚠ 會覆蓋上面的 1 }
STOREVALUE(2CBh, 結果);
```

> ★★★ **兩邊各擲一顆 d6 對上各自的門檻**，結果碼
> `0` ＝ 都沒突襲、`1` ＝ 只有一邊、`2` ＝ 只有另一邊、`3` ＝ 雙方都被突襲。
> ⚠ **`r1 <= 甲` 且 `r2 <= 乙` 時先寫 3，接著第二個 `if` 又把它改寫成 2**
> ——`3` 這個結果碼**永遠寫不出去**。形狀上是原作的邏輯瑕疵。
> ★ 結果固定寫到 ECL 位址 **`2CBh`**（不是 operand 指定的）。

## `dos overlay-02:01416h`（199 條）＝ `1Eh` `CHECKPARTY`

```pascal
READVAR(6);
if 第一個 operand 的 code = 1 then 值 := ADDFNC(high[1], low[1])
else                              值 := ADDRESSVALUE(1);
效果碼 := ADDRESSVALUE(2);
三個輸出位址 := ADDFNC(high[3..6], low[3..6]);
最小 := 0FFh;  最大 := 0;  總和 := 0;  人數 := 0;

case 值 − 7FFFh of
  8001h:                                        { ＝ operand1 為 0 }
      走隊伍鏈，只要有人 <overlay-24 entry#27>(效果碼) 成立就回報；
  0A5h..0ACh:
      欄位 := 角色^[0E9h ＋ (值 − 7FFFh − 0A4h)];
      走隊伍鏈累計最小／最大／總和／人數；平均 := 總和 div 人數;
  9Fh:
      欄位 := 角色^[1A5h];  同上;
end;
<near 13B6h>(bp);                               { 巢狀程序，把三個結果寫回 }
```

> ★★★ **`CHECKPARTY` 是「查全隊某個欄位的最小值、最大值、平均值」**，
> 三個結果由 operand 3..6 指定的位址收。
> ★ `0A5h`..`0ACh` 對到 `角色^[0E9h ＋ 1..8` 這一組八格的陣列；
> `9Fh` 對到 `角色^[1A5h]`。
> ⚠ 本規格不宣稱那八格與 `+1A5h` 是什麼欄位。

## ★★★ `dos overlay-02:030DDh`（104 條）＝ `38h` `PROGRAM`

```pascal
if DS:47E2h <> 0 then begin DS:6506h := DS:47E7h;  DS:47E2h := 0 end;
READVAR(1);
case ADDRESSVALUE(1) of
  0: begin <overlay-17 entry#5>();  <overlay-17 entry#1>();     { 重新初始化 ＋ 主選單 }
        if (DS:728Ah <> 'P') and (bank0^[1CCh] = 0) then
            <overlay-24 entry#37>();  end;                       { 重繪 }
  8: begin
        <overlay-18 entry#1>();                                  { ★ 結局過場，spec 1082 }
        DS:8B6Fh := 1;
        bank0^[3FAh] := 0FFh;
        bank1^[550h] := 0FFh;
        走隊伍鏈：目前 HP := 最大 HP、+195h := 0、+196h := 1;     { 全隊復活補滿 }
        <overlay-17 entry#1>();                                  { 主選單 }
        if 問("You've won. Save before quitting? ") = 'Y' then
            <overlay-16 entry#11>();                             { 存檔 }
        結束程式（06EA:0000h）;
     end;
  9: begin 存下 ECL PC(DS:4FB4h)；<overlay-15 entry#3>()；
           <near 3073h>()；還原 ECL PC；<near 0052h>() end;
  3: begin DS:4FC7h := 1;  <near 0052h>() end;
end;
```

> ★★★ **這一支把三個先前「不宣稱」的旗標一次關掉**：
> - **`DS:8B6Fh` ＝ 通關旗標**——spec 1084 記它讓訓練所**不收 1000 gp**。
> - **`bank0^[3FAh] ＝ 0FFh` ＝ 已通關**——spec 1085 在主選單用它印
>   「きみたちは勝利した。」，猜的方向對了，這裡是設定端。
> - **`bank1^[550h] ＝ 0FFh`**——spec 1084／1085 的「訓練所收哪些職業」遮罩
>   被設成全開 ⇒ **通關之後任何職業都能訓練**。
>
> ★★ 通關同時把**全隊復活並補滿血**（`+1A4h := +78h`、`+195h := 0`、`+196h := 1`），
> 然後回主選單讓玩家存檔——**存下來的是「通關後的隊伍」**，
> 這正是 Gold Box 系列「帶角色進下一款」的入口。
> ★ `DS:4FC7h` 就是 spec 1084 那個「這次訓練不收費」旗標之一，由 `PROGRAM 3` 打開。

## `dos overlay-02:00C15h`（131 條）＝ `21h` `LOAD FILES` ／ `37h` `LOAD PIECES`

**同一支處理兩個指令碼**，靠 `DS:75FFh`（目前指令碼，spec 1083 的同一格）分流：

```pascal
READVAR(3);  DS:47E3h := 1;
for i := 1 to 3 do o[i] := ADDRESSVALUE(i);

if DS:75FFh = 21h then begin                      { LOAD FILES }
    DS:47E5h := 1;
    if (o[3] <> 0FFh) and (o[3] <> 7Fh) and (bank0^[1CCh] <> 0) then begin
        bank0^[18Ah] := o[3];
        <overlay-30 entry#9>(o[3]);               { Load3DMap }
        bank1^[592h] := 0;
    end;
    if (o[1] <> 0FFh) and (bank0^[1CCh] = 0) and (DS:728Ah <> 'P') then
        <overlay-29 entry#9>('y');                { 載入 bigpic }
end else begin                                     { LOAD PIECES }
    DS:47E4h := 1;
    if o[3] = 7Fh then <overlay-30 entry#8>(1, 0)  { WALLDEF }
    else if (bank0^[1CEh] <> 0) and (bank0^[1D0h] <> 0) then begin
        if o[3] <> 0FFh then <overlay-30 entry#8>(1, o[3]);
        if o[1] <> 0FFh then <overlay-30 entry#8>(3, o[1]);
    end else
        for i := 1 to 3 do
            if o[i] <> 0FFh then <overlay-30 entry#8>(i, o[i])
            else begin word[DS:7210h ＋ i × 4] := 0FFFFh;
                       word[DS:7212h ＋ i × 4] := 0FFFFh end;
end;

if (DS:47E4h <> 0) and (DS:47E5h <> 0)
   and (DS:4FBBh = 3) and (DS:4FBAh <> 3) and (DS:8B6Eh <> 0) then begin
    <01A0:0136h>();  <overlay-24 entry#2>(目前角色);  <overlay-24 entry#38>();
end;
DS:8B6Eh := 0;
```

> ★★★ **`DS:7210h`／`DS:7212h` 就是 spec 1076 存進存檔的那三組牆面參數**
> ——這裡是**寫入端**，`0FFFFh` 代表「這一格不用」，正好對上
> spec 1074 記的初值與 spec 1072／1076 的 `> 0` 判斷。三份規格互相印證。
> ★ 兩個指令都會設一個「我做過了」的旗標（`DS:47E4h`／`DS:47E5h`），
> **兩個都做過**才觸發最後那次重繪。

## ★★★ `dos overlay-02:01B53h`（398 條）＝ `27h` `TREASURE`

```pascal
READVAR(8);
for i := 1 to 6 do dword[DS:6F70h ＋ i × 4] := ADDRESSVALUE(i ＋ 1);
n := ADDRESSVALUE(8);
```

`n` 分三段：

| `n` | 行為 |
|---|---|
| **< `80h`** | 從 **`'ITEM' ＋ Str(DS:5BEEh) ＋ '.dax'`** 讀進第 `n` 塊，<br>每 `3Fh` ＝ **63 bytes** 一筆（＝物品節點，spec 1038）<br>逐筆 `GetMem(3Fh)` ＋ `Move` 串進 `DS:6F8Ch` 鏈（next 在 `+2Ah`，spec 1075） |
| **`0FFh`** | 什麼都不做 |
| **≥ `80h`** | 隨機產生 **`n − 80h`** 件寶物（見下） |

★ 檔案讀不到就印 `'Unable to find item file'` 並**直接結束程式**（`06EA:0000h`）。

### 隨機寶物的 d100 分段

```pascal
r1 := ROLLDICE(1, 100);
if r1 in [1..3Ch] then begin                       { 1..60 }
    r2 := ROLLDICE(1, 100);
    if (r2 in [1..2Fh]) or (r2 in [32h..3Bh]) then
        類型 := (r2 = 2Dh) ? 3Bh : r2               { ★ 45 這一格改判成 59 }
    else if r2 in [3Ch..5Ah] then begin            { 60..90 }
        r3 := ROLLDICE(1, 100);
        case r3 of 1..4: 24h;  5..7: 23h;  8: 22h;  9: 25h;  0Ah: 26h end;
    end
    else if r2 in [5Bh..5Eh] then 類型 := 49h
    else if r2 in [5Fh..61h] then 類型 := 5Dh
    else if r2 in [62h..64h] then 類型 := 4Dh
    else                          類型 := 3Bh;
end
else if r1 in [3Dh..55h] then 類型 := 3Dh          { 61..85 }
else if r1 in [56h..5Ch] then 類型 := 3Eh          { 86..92 }
else if r1 in [5Bh..62h] then begin                { ⚠ 與上一段重疊 }
    r3 := ROLLDICE(1, 100);
    case r3 of 1..9: 47h;  0Ah: 54h;  0Bh..0Fh: 4Fh end;
end
else if r1 in [63h..64h] then 類型 := 3Bh;         { 99..100 }
<overlay-21 entry#18>(類型, …);                    { 造出這件物品 }
…GetMem(3Fh) ＋ Move 串進鏈…
```

⚠ **`56h..5Ch` 與 `5Bh..62h` 兩段重疊**（91、92 兩點）——先判的那段贏，
所以 `5Bh..62h` 實際只涵蓋 93..98。形狀上是原作的表格瑕疵。
⚠ 本規格不宣稱那些類型碼（`22h`..`5Dh`）各自是哪一種物品。

## 中文化

| DOS | 位址 | 長度 | 建議 |
|---|---|---|---|
| `'Unable to find item file'` | `overlay-02 CS:1B3Ah` | 24 | 「找不到物品檔案」（發生時會直接結束程式） |
| `"You've won. Save before quitting? "` | `overlay-02 CS:30BAh` | 34 | 「你們獲勝了。結束前要存檔嗎？」（熱鍵 `Y`） |
| `'ITEM'`／`'.dax'` | `CS:1B30h`／`1B35h` | 4／4 | ⚠ **檔名，不可翻** |

## 明確不宣稱

- 沒有宣稱 `CHECKPARTY` 用的 `角色^[0E9h ＋ i]`（八格）與 `角色^[1A5h]` 是什麼。
- 沒有宣稱 `TREASURE` 那些物品類型碼各自對應哪一件物品。
- 沒有宣稱 `DS:6F70h ＋ i × 4` 那六個 32-bit operand 交給誰用。
- 沒有宣稱 `DS:47E2h`／`47E3h`／`47E4h`／`47E5h`／`47E7h` 由誰讀。
- 沒有宣稱 `DS:8B6Eh`（`LOAD FILES`／`LOAD PIECES` 收尾判斷的那一格）是什麼。
- 沒有宣稱 `near 13B6h`（`CHECKPARTY` 寫回結果的巢狀程序）的細節。
- 沒有宣稱 `near 3073h`／`near 0052h`（`PROGRAM 9`／`3` 呼叫的兩支）做什麼。
- 沒有宣稱 `SURPRISE` 的兩個門檻 `o[4] ＋ 2 − o[1]` 與 `o[2] ＋ 2 − o[3]`
  在規則書上對應什麼。
- 沒有宣稱 PC-98 側這六支是否相同。
