# 1045 — 地城探索主迴圈（DOS）：走一步、跑 ECL、回頭再走

- 證據等級：`exact`（DOS 側 205 條逐條讀完）
- 作法見 spec 783

## `dos overlay-02:03772h`（`retf`）

原本待解讀。整支沒有字串。

## 進場設定

```pascal
DS:47E7h := DS:6506h;                 { 記住目前角色 }
DS:7580h := 1;  DS:8B64h := 0;
DS:47E1h..47E5h := 0;                 { 五個旗標 }
DS:8B6Eh := 1;
DS:4FBAh := 4;                        { ★ 畫面模式 ＝ 4（spec 1042 的門選單也要 4）}
if bank0^[1E4h] = 0 then begin
    DS:8B6Eh := 0;
    DS:8B5Eh := (DS:4FBCh <> 0) ? 'R'(52h) : 1;
    <overlay-24 entry#2>(DS:6506h);   { 畫隊伍名單 }
end else
    DS:8B5Eh := bank0^[1E4h];
if bank0^[1CCh] = 0 then DS:4FBAh := 3;
```

★ `bank0^[1E4h]` ＝ **上一個 block／scene 編號**（spec 1097 §一，六份規格交叉驗證）：
非 0 就沿用，0 才重新初始化。
★ `DS:4FBAh` ＝ 畫面模式（spec 592 起的既有結論）：地城是 `4`，
`bank0^[1CCh] = 0` 時降成 `3`。

## 主迴圈

```pascal
repeat
    鍵 := <overlay-14 entry#2>();              { 走一步／收鍵 }
    DS:47E7h := DS:6506h;
    while (bank1^[594h] > 1) or (UpCase(鍵) = 'E') do begin
        if DS:4FC7h <> 0 then break;
        if UpCase(鍵) = 'E' then begin
            <overlay-15 entry#3>();  <sub_3073>();  break
        end;
        DS:8B5Fh := bank1^[594h] and 1;
        bank1^[594h] := 1;
        DS:7580h := 1;
        <overlay-28 entry#1>();                { ★ 跑 ECL }
        <sub_3621>();
        if DS:47E1h <> 0 then <sub_3691>();
        bank1^[594h] := DS:8B5Fh;              { ★ 還原 }
        if DS:4FC7h = 0 then begin
            鍵 := <overlay-14 entry#2>();
            DS:47E7h := DS:6506h;
        end;
    end;
    …
    bank0^[1E0h] := DS:720Fh;  bank0^[1E2h] := DS:7210h;   { 記下座標 }
    <overlay-14 entry#1>();
    <overlay-28 entry#1>();
    if (bank0^[1E0h] <> DS:720Fh) or (bank0^[1E2h] <> DS:7210h) then
        <far 0713:0020h>();                    { ★ 座標被 ECL 改過才叫 }
    DS:8B62h := 0;  DS:8B63h := 1;
    <sub_3621>();
    if DS:47E1h <> 0 then <sub_3691>();
until DS:4FC7h <> 0;
DS:4FC7h := 0;
```

★★ **ECL 執行的入口是 `overlay-28 entry#1`**，主迴圈裡出現三次。
★★ `bank1^[594h]` 進 ECL 前被暫存成 `and 1`、設 1、跑完再還原
——形狀上是「這一格還有沒有事件」的計數，ECL 執行期間先壓成 1。
★★★ **座標比對**：`bank0^[1E0h]`／`1E2h` 存進 ECL 前的座標，
跑完拿 `DS:720Fh`／`7210h`（spec 1022／1026／1042 的同一組）比對，
**只有被 ECL 搬動過才叫 `0713:0020h`**。
★ `DS:8B62h := 0`、`DS:8B63h := 1` 這兩格就是 spec 1039／1041 決定選單樣式的那兩個旗標。
★ `DS:4FC7h` 是離開旗標，出迴圈後清 0。

## 另一條路：`DS:4FBCh <> 0`

```pascal
while (DS:650Ah, DS:650Ch) = (0, 0) do
    <overlay-17 entry#3>(1, 1);                { 隊伍是空的就一直叫 }
```

★ 形狀上是「還沒有隊伍就先去組隊」。

## 中文化

本支沒有字串。

## 明確不宣稱

- 沒有宣稱 `sub_3621`／`sub_3691`／`sub_3073` 三支本模組函式做什麼。
- 沒有宣稱 `DS:47E1h`..`47E9h` 那組旗標與指標各自的語意。
- 沒有宣稱 `DS:8B5Eh` 的值域（`bank0^[1E4h]` ＝ 上一個 block／scene 編號，見 spec 1097 §一）。
- 沒有宣稱 `far 0713:0020h` 做什麼。
- 沒有宣稱 `'E'` 鍵走的 `overlay-15 entry#3` 是哪個畫面。
★ `bank1^[594h]` ＝ **這一格還有沒有事件的計數**（`> 1` 才跑 ECL，跑完還原成 `and 1`；
spec 1095 的 `COMBAT` 收尾做同一個動作，spec 1097 §三）。
