# 861 — 戰鬥結束處理：清效果、復原狀態、發經驗值的總調度

- 證據等級：`exact`（DOS 側逐條讀完；PC-98 側 51 條運算元差異全是位址，
  **另有一處真的邏輯差異**）
- 作法見 spec 783

## `overlay-05:0053Ch` ↔ `overlay-05:0052Fh`（DOS 343 條／PC-98 341 條）

`retf` ＝ 無參數。狀態碼一律照 spec 825 的 `DS:0FE0h` 名稱表讀：
`0` Okay、`1` Animated、`3` Running、`4` Unconscious、`5` Dying。

```pascal
DS:8B57h := 0;  DS:4FC7h := 1;  DS:47EDh := 0;

{ ① 有沒有人在逃 }
p := 遠指標(DS:650Ah);
while (p <> NIL) and (p^[18Dh]^[13h] <> 1) do begin
    if p^[195h] = 3 then DS:47EDh := 1;
    p := p^[189h];
end;

{ ② 還有沒有人「沒真的倒下」 }
還有 := 0;  p := 遠指標(DS:650Ah);
while (p <> NIL) and (還有 = 0) do begin
    if (p^[196h] <> 0) or (p^[195h] in [3, 4, 5]) then 還有 := 1;
    p := p^[189h];
end;

if (DS:8B56h <> 0) or ((遠指標(DS:4F9Dh)^[5CCh] <> 0) and (還有 <> 0)) then
    DS:4FC7h := 0;

DS:8B5Ch := 0;  p := 遠指標(DS:650Ah);
if (DS:8B56h <> 0) and (DS:4FBCh <> 0) then goto 特殊;      { ← 兩平台不同，見下 }

{ ③ 逐一清效果、統計不分經驗的人 }
while (p <> NIL) and (p^[18Dh]^[13h] <> 1) do begin
    if (p^[195h] in [0, 1, 3]) and (p^[197h] = 0) and (p^[0F7h] < 80h) then
        DS:4FC7h := 0;                                  { 士氣已崩的還在場 }
    if p^[195h] in [0, 1] then begin
        DS:8B5Ch := 1;  DS:47EDh := 0;                  { 有人站著 → 要發經驗值 }
    end;
    if (p^[196h] = 0) or (p^[195h] = 1) then inc(DS:8B57h);
    for k := 1 to 13h do
        <far 143Dh+2>(p, byte[2A7h + k], NIL);          { 清 19 種效果 }
    p := p^[189h];
end;

{ ④ 發經驗值 }
if DS:8B5Ch <> 0 then begin
    <overlay-05:0039h>(@DS:8B58h);                      { spec 858：算每人份 }
    <overlay-05:0351h>(longint@DS:8B58h);               { spec 857：發下去 }
end;

p := 遠指標(DS:650Ah);
if DS:4FC7h <> 0 then goto 真的結束;

{ ⑤ 沒真的結束：復原狀態 }
while (p <> NIL) and (p^[18Dh]^[13h] <> 1) do begin
    if DS:47EDh <> 0 then begin                          { 有人在逃 }
        遠指標(DS:4F9Dh)^[58Eh] := 81h;
        if p^[195h] = 3 then begin
            p^[195h] := 0;  p^[196h] := 1;  p := p^[189h];
        end else begin
            DS:6506h := p;  p := p^[189h];  <far 0F0Dh+2>(1, 0);
        end;
    end else begin
        case p^[195h] of
          3: begin p^[195h] := 0;  p^[196h] := 1 end;    { Running → Okay }
          5: if 遠指標(DS:4F9Dh)^[5CCh] <> 0 then begin   { Dying }
                 p^[195h] := 0;  p^[196h] := 1;  p^[1A4h] := 1;
             end else
                 p^[195h] := 4;                          { → Unconscious }
          4: if p^[1A4h] > 0 then begin                  { Unconscious }
                 p^[195h] := 0;  p^[196h] := 1;
             end else if 遠指標(DS:4F9Dh)^[5CCh] <> 0 then begin
                 p^[195h] := 0;  p^[196h] := 1;  p^[1A4h] := 1;
             end;
        end;
        p := p^[189h];
    end;
end;
離開;

真的結束:
while (p <> NIL) and (p^[18Dh]^[13h] <> 1) do begin
    DS:6506h := p;  <far 0F0Dh+2>(1, 0);
    p := p^[189h];
end;
遠指標(DS:4F9Dh)^[67Ch] := 0;                            { 隊伍人數歸零 }
離開;

特殊:
p := 遠指標(DS:650Ah);
while p <> NIL do begin
    if (p^[196h] = 1) and (p^[195h] = 0) and (p^[197h] <> 1) then begin
        DS:8B5Ch := 1;
        <overlay-05:0039h>(@DS:8B58h);
        <overlay-05:0351h>(longint@DS:8B58h);
    end;
    p := p^[189h];
end;
p := 遠指標(DS:650Ah);
while p <> NIL do begin
    if p^[195h] in [0, 1] then p^[196h] := 1;
    if p^[195h] = 5 then p^[195h] := 4;                  { Dying → Unconscious }
    p := p^[189h];
end;
```

## 戰後鏈整條接上

```
overlay-05:0053Ch （本支）
    ├─ 清 19 種效果（DS:2A7h 表）
    ├─ 數出不分經驗的人 → DS:8B57h
    ├─ overlay-05:0039h（spec 858）→ 每人份 = 總值 div (人數 − DS:8B57h)
    └─ overlay-05:0351h（spec 857）→ 逐人加上 +127h
```

spec 858 的「明確不宣稱」原本記著「沒有宣稱 `DS:8B57h` 是什麼」——
**本支就是唯一寫它的地方**：初值 0，對每個 `+196h = 0`（不能行動）
或 `+195h = 1`（Animated）的成員加 1。所以那個除數是
**「還能分經驗值的人數」**，被召喚物與倒下的人不參與均分。

## `+18Dh^[13h] = 1` ＝ 隊伍到此為止

四個迴圈裡有三個都以 `p^[18Dh]^[13h] = 1` 中止，第四個（特殊分支）沒有。
`DS:650Ah` 那條鏈同時串著隊伍與怪物（spec 858 靠 `+197h` 分邊），
本支則靠戰鬥狀態記錄的 `+13h` 當**分界標記**——
**這是 22 bytes 戰鬥狀態裡 `+13h` 的第一個用途**。

## `DS:2A7h`（PC-98 `DS:5BFh`）＝ 戰鬥結束要清掉的 19 個效果碼

1 起算，**兩平台逐 byte 相同**：

```
03 0B 0D 15 17 1B 23 28 33 34 35 3A 5B 88 89 8B 8E 90 1F
```

`88h`／`8Bh`／`90h` 與 spec 846／847／851 讀到的效果碼同一組。
表的後面（`2BBh` 起）就是 spec 837 的 AI 方向表 `DS:2B6h` 的資料區，
兩張表相鄰但不重疊。

## `+1A4h` 與 `DS:4F9Dh^[5CCh]`：瀕死不會真死的設定

`5CCh` ≠ 0 時，`Dying`(5) 與 `Unconscious`(4) 都直接復原成 `Okay` 並把
`+1A4h` 設 1；`5CCh` ＝ 0 時 `Dying` 只降級成 `Unconscious`。
`+1A4h` 已經是 1 的人下次無條件復原。
形狀上是**「這場戰鬥不會真的死人」的設定**，本規格不宣稱它由哪個選項控制。

## ⚠ 兩平台真的不一樣：進入特殊分支的條件

這是全支唯一的助憶碼差異：

| | 條件 |
|---|---|
| DOS | `(DS:8B56h <> 0) and (DS:4FBCh <> 0)` |
| PC-98 | `DS:0BDE8h <> 0`（＝ DOS 的 `8B56h`）**單一條件** |

PC-98 把 DOS 的第二個旗標 `DS:4FBCh` 換成同一個 `8B56h`，於是第一個測試變成
多餘、被編譯器消掉，少了兩條指令（341 對 343）。
**這不是反組譯錯位，是 PC-98 真的改了判斷**（spec 783 的第 d 類）。
remake 要選一邊；本規格不判斷哪一邊是對的。

## ⚠ 特殊分支裡的發經驗值在迴圈內

特殊分支對**每一個**符合條件的存活成員都呼叫一次
`0039h` ＋ `0351h`，所以 N 個存活者就發 N 次。
`0039h` 在 `DS:8B56h <> 0` 時回的是固定值（`遠指標(DS:6506h)^[0E5h] × 100`，
spec 858），所以總量是 N 倍。**形狀可疑但兩平台一致**，
不是反組譯問題；remake 要照抄的話得先用實機確認。

## 明確不宣稱

- 沒有宣稱 `far 143Dh+2` 怎麼清效果。最後推的兩個 0 是**一個 NIL 遠指標**，
  所以簽章是 `(?, 碼, ?)` 三個參數；spec 862 同一支傳的是
  `(效果節點, 碼, 角色)`，**第一與第三個參數的角色對不起來**，兩支都不定名。
- 沒有宣稱 `far 0F0Dh+2(1, 0)` 做什麼（配 `DS:6506h` 指定對象，
  形狀上是「處理這一個人」）。
- 沒有宣稱 `DS:8B56h`／`DS:4FBCh`／`DS:4F9Dh^[5CCh]`／`^[58Eh] := 81h` 的意義。
- 沒有宣稱 19 個效果碼各是什麼效果。
- 沒有宣稱 `DS:4FC7h`（「戰鬥真的結束」）與 `DS:47EDh`（「有人在逃」）由誰讀。
