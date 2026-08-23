# 1148 — `0Eh PICTURE`：`0FFh` 是關閉，開啟再分頭像合成與大圖

- 證據等級：`exact`（DOS `overlay-02:0841h`..`092Bh` 整支 77 條讀完）
- handler 位址取自 `docs/audit/ecl-opcode-handlers-dos.md`

## 三層分岔

```pascal
n := ADDRESSVALUE(1);
if n = 0FFh then goto 關閉;                      { 085Ah }

DS:8B49h := 1;  DS:8B62h := 1;
if bank1^[5C2h] = 0FFh then begin                { 0871h；5C2h ＝ ECL 格 7EE1h }
    DS:8B63h := 1;
    if n >= 78h then begin                       { 087Fh：大圖 }
        sub_17DD(n); sub_17E2(); DS:7580h := 0;
    end else begin                               { 一般圖 }
        sub_17C4(DS:722Ch, 0, n);
        sub_17B5(DS:7232h, DS:7234h, 1, 3, 3);
    end;
end else
    <loc_6F1+2>(n, bank1^[5C2h]);                { ★ 頭像 ＋ 身體合成 }
goto 尾;

關閉:                                             { 08E9h }
if not ((DS:4FBBh = 4) and (DS:4FBAh = 4)) then
    if (DS:8B62h <> 0) or (DS:8B65h <> 0) then begin
        DS:7580h := 1;  <loc_1774+1>();          { 重繪 }
        DS:8B62h := 0;  DS:8B65h := 0;  DS:8B63h := 1;
    end;
尾:
DS:8B48h := 0;  DS:8B49h := 0;
```

## 三個判準

| 判準 | 位址 | 意義 |
|---|---|---|
| `n = 0FFh` | `085Ah` | **關閉**目前的圖 |
| `bank1^[5C2h] <> 0FFh` | `0871h` | 有指定頭像 ⇒ 走**頭像＋身體合成**那一支 |
| `n >= 78h` | `087Fh` | **大圖**（另一組載入與繪製常式）|

`bank1^[5C2h]` ＝ ECL 格 `7EE1h`（換算見 spec 1146），腳本用 `SAVE ?? 7EE1` 指定
頭像，用 `SAVE FF 7EE1` 取消——`ECL2/0x01` 提爾佛頓的招牌那一段就是這個用法。

## remake

`0FFh` 先前是**什麼都不做**；現在會發出 `PictureCloseRequested`。開啟那一支的
三件事（區塊、大圖旗標、頭像 block）本來就對得上。

一般圖那一支的兩條呼叫是 `LOADSEQUENCE(圖名, n, 0, @DS:722Ch)` 與
`SHOWPORTRAIT(3, 3, 1, 序列[1])`——`PIC*.DAX` 的每個 block 都是**多格序列**，
`PICTURE` 只畫第 1 格，之後由 `MENUS` 的自走驅動器或腳本的 `2Dh CALL 6803h`
往下推（spec 1150）。

⚠ 仍是 `partial`：`PictureRequested`／`PictureCloseRequested` 是**訊號**，實際的
載入、頭像與身體的分塊合成都在 UI 那一側，還沒逐項對過。旁路的語意見下一節，
但 remake 這一側還沒建那個模型。

## `4FBAh`／`4FBBh` 是「前後兩次的畫面模式」

`STOREVALUE` 發現 ECL 格 `4BE6` 換值時先 `4FBBh := 4FBAh`，再依新值把 `4FBAh`
設成 3（新值 0）或 4（新值非 0）（DOS `overlay-07:0DA0h`，spec 1150）。而
**`4BE6h` 就是第一人稱（3D 地圖）模式旗標**：非零才載 3D 地圖，為零改走大圖
（spec 1096／1181，引擎側同一格是 `bank0^[1CCh]`）。

⇒ `4FBAh` ＝ 現在的模式、`4FBBh` ＝ 上一次的模式，而

| 值 | 畫面 |
|---:|---|
| 3 | 非第一人稱（大圖那一類）|
| 4 | 第一人稱（3D 地圖）|

所以關閉那一支的旁路 `not ((4FBBh = 4) and (4FBAh = 4))` 讀作：**前後都還在
第一人稱就不重繪**。模式沒變而且本來就是 3D 視窗時，關掉圖片不需要重畫整個
畫面——主迴圈本來就會把 3D 視窗畫回去。模式**變過**（3→4 或 4→3）或現在是
大圖，才走重繪。

⚠ 旁路連 `8B62h`／`8B65h` 的清除一起跳過（那兩格與 `7580h := 1` 都在 `if` 裡），
所以走旁路時「圖還開著」這個狀態**留著**。

## 明確不宣稱

- 沒有宣稱 `DS:8B62h`／`8B63h`／`8B65h`／`8B48h`／`8B49h` 各自的語意。
- 沒有宣稱 `DS:7580h`（大圖 0、關閉 1）代表什麼。
- 沒有宣稱 corpus 裡 `PICTURE 0FFh` 出現幾處——只確認它存在
  （`ECL3/0x10` 索引 39 的每格事件守衛就是 `PICTURE FF`，見 `docs/audit/cell-sweep.md`）。
