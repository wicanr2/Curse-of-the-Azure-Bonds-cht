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
載入、頭像與身體的分塊合成、以及 `4FBAh`／`4FBBh` 那個「兩個都等於 4 就不重繪」
的旁路都在 UI 那一側，還沒逐項對過。

## 明確不宣稱

- `DS:4FBAh`／`DS:4FBBh` 由 ECL 格 `4BE6` 決定：`STOREVALUE` 發現 `4BE6` 換值時
  先 `4FBBh := 4FBAh`，再依新值 0／非 0 把 `4FBAh` 設成 3／4
  （DOS `overlay-07:0DA0h`，spec 1150）。所以「兩個都等於 4 就不重繪」＝
  「本來就是這種畫面、現在還是」。這一支各代表哪一種畫面仍未解讀。
- 沒有宣稱 `DS:8B62h`／`8B63h`／`8B65h`／`8B48h`／`8B49h` 各自的語意。
- 沒有宣稱 `DS:7580h`（大圖 0、關閉 1）代表什麼。
- 沒有宣稱 corpus 裡 `PICTURE 0FFh` 出現幾處——只確認它存在
  （`ECL3/0x10` 索引 39 的每格事件守衛就是 `PICTURE FF`，見 `docs/audit/cell-sweep.md`）。
