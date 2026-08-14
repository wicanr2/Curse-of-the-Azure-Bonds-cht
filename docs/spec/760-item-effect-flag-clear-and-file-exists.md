# 760 — 清掉物品三個效果槽的旗標位元，兩平台的「檔案在不在」不同做法

- 證據等級：`exact`（逐條讀完）

## `overlay-15:01CBh`（兩平台）— 清掉三個效果槽的高位位元

`retf 4`，參數是角色遠指標：

```pascal
p := 角色^[14Dh];                        { 物品鏈頭；PC-98 是 +14Eh }
while p <> NIL do begin
    if 本模組 11DEh(p) then
        for i := 1 to 3 do
            p^[3Bh + i] := p^[3Bh + i] and 7Fh;
    p := p^[2Ah];                        { PC-98 是 +52h }
end;
```

`3Bh + 1..3` ＝ `+3Ch`／`+3Dh`／`+3Eh`，正是物品節點的**三個效果槽**（既有結論：
低 7 位是效果 ID、最高位是旗標）。所以這一支的作用就是**把三個槽的旗標位元
一次清掉，效果 ID 保留**。

PC-98 的對應位移是 `+64h`／`+65h`／`+66h`、鏈結在 `+52h`——與 spec 756 在
`overlay-26` 觀察到的「PC-98 物品／選單節點重新排過」一致。

呼叫端是 `overlay-15:023Eh`（spec 755），對隊伍每個成員各做一次。

## `overlay-13:040B8h`（PC-98）— 與 DOS 同構但最後一個參數不是常數

DOS 的 `overlay-13:4162h`（spec 759）最後呼叫
`<overlay-24 entry#25>(x1, y1, x2, y2, 1, 1Eh)`；PC-98 這一支同樣的位置放的是
`DS:7F16h * 7`（有號乘法），不是常數。其餘結構逐條相同（取兩個戰鬥員的 x／y
再畫）。這是兩平台真實的行為差異，不是匯出誤差。

## `overlay-16:3DFFh`（PC-98）— DOS `351Ch` 的對應

流程與 spec 756 的 DOS `overlay-16:351Ch` 相同：`DS:7F09h^[67Ch] > 7` 就整支不
做；`GetMem` 之後填 `+126h`、組出 `CS:3DFAh` 的 `'CPIC'` 字串，最後叫
`(參數, p^[143h])`。

**唯一的差別是配置大小**：DOS 是 `1A6h`（422），PC-98 是 `1A7h`（423），
**多一個 byte**。移植時多留的那一格要照抄。

## `overlay-29:045Ah`（PC-98）— 逐項釋放後清空

`retf 4`：

```pascal
if p^[0] = 0 then 離開;
DS:0A9D9h := p^[0];                       { 項數，有號判斷 >= 1 }
repeat
    02A8h:10D5h(p + DS:0A9D9h * 8 − 1);
    if DS:0A9D9h <> 1 then dec(DS:0A9D9h);
until DS:0A9D9h = 1;
FillChar(p^, 43h, 0);
DS:0A313h := 0;
DS:0A325h := 0FFh;
```

由高索引往低索引處理，元素間距 **8 bytes**、起點是 `p + 8i − 1`。整塊清掉
`43h`（67）bytes。計數器用的是全域 `DS:0A9D9h` 而不是 local——**這支不可
重入**。

`02A8h:10D5h` 在 spec 753 的 `overlay-29:0777h` 也出現過（收一個遠指標），
兩處用法一致。

## `overlay-04:0051h`（PC-98）— 一句 Y/N 詢問

`retf 4`：

```pascal
StoreString(參數字串, 緩衝, 0FFh);
014Ah:0084h(遠指標@DS:9594h, @緩衝, 0, 0);
組字串(CS:0034h = 'それでもかけてもらいますか？');
結果 := 0164h:0048h(0Dh, 0Ah, 0Fh);
014Ah:0089h();
```

`CS:0034h` 是 28 bytes 的 Shift-JIS 字串，意思是「還是要施在你身上嗎？」。
`0164h:0048h` 收三個常數（`0Dh`、`0Ah`、`0Fh`）並回一個 byte，是這類詢問的輸入
程序。`014Ah:0084h` 與 `014Ah:0089h` 成對，像是開窗與關窗。

**中文化要點**：這句是 overlay 內嵌的 Shift-JIS 常數，不在任何字串表裡。

## 兩平台判斷「檔案在不在」的做法不同

DOS `START.EXE:1685Fh`（spec 757）用 `FindFirst` ＋ `DosError`。

PC-98 `PC98-GAME.EXE:17656h`（`retf 4`）改用開檔：

```pascal
StoreString(參數字串, 緩衝, 50h);
word_23B08 := 0;                 { InOutRes }
Assign(f, 緩衝);
Reset(f, 1);                     { 記錄長度 1 }
if word_23B08 = 0 then begin Close(f);  結果 := true end
else 結果 := false;
word_23B08 := 0;                 { 離開前再清一次 }
```

行為差異是實際存在的：`FindFirst` 找得到唯讀／被別人開啟的檔案，`Reset` 在
共享模式下可能失敗。remake 要對齊哪一版得先決定。

## 明確不宣稱

- 沒有宣稱 `overlay-15:11DEh` 的判斷條件是什麼（只知道回真才清旗標）。
- 沒有宣稱 `02A8h:10D5h`、`014Ah:0084h`／`0089h`、`0164h:0048h` 是什麼。
- 沒有宣稱 `DS:7F16h * 7` 算的是什麼。
- 沒有宣稱 `'CPIC'` 是檔名還是資源標籤（同 spec 756）。
