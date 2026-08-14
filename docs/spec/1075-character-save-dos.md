# 1075 — DOS 角色存檔：先算所需空間再比 `DiskFree`，物品鏈 next 在 `+2Ah`、效果鏈 next 在 `+5`

- 證據等級：`exact`（DOS 側 547 條逐條讀完，無匯出破洞）
- 作法見 spec 783；讀取端見 spec 1038

## `dos overlay-16:00DEAh`（`retf 8`）

原本待解讀。範圍 `0DEAh`..`1383h`。這是 spec 1038 那支讀取函式的**寫入對側**。

```pascal
procedure 存角色(角色: 遠指標; 檔名: 字串);    { arg_4 ＝ 角色, arg_0 ＝ 檔名 }
```

（Turbo Pascal 由左至右推疊 ⇒ `arg_4`（`bp+0Ah`）是先宣告的那個。）

## ★★★ 一、先算出「這個角色要佔多少 bytes」

```pascal
名字 := 檔名;                       { 上限 8 ——★ DOS 檔名 8 字元 }
DS:7584h := 0;                      { 存檔格式 ＝ 現行 }
if not <near 0643h>() then exit;

需要 := 1A6h;                       { 422 ＝ 角色本體 }
節點 := 角色^[14Dh];                { 物品鏈 }
while 節點 <> NIL do begin 需要 := 需要 ＋ 3Fh; 節點 := 節點^[2Ah] end;
節點 := 角色^[0F2h];                { 效果鏈 }
while 節點 <> NIL do begin 需要 := 需要 ＋ 9;   節點 := 節點^[5]  end;
```

> ★★★ **關掉 spec 1038 的兩個未定項**：
> - **物品鏈節點（`3Fh` ＝ 63 bytes）的 next 在 `+2Ah`。**
>   這與 spec 1073 推出的 DOS 選單節點 next 位置相同——
>   **DOS 這一代的鏈結節點一律把 next 放在 `+2Ah`**。
> - **`角色^[0F2h]` 就是 `.fx` 效果鏈的鏈頭**（spec 1038 只記「一個尚未命名的
>   遠指標欄位」），節點 9 bytes，**next 在 `+5`**。
>
> ⚠ spec 1038 記的第三條鏈 `.spc` **本支沒有寫**。DOS 的存檔只產生
> 角色本體 ＋ `.swg` ＋ `.fx` 三個檔。`.spc` 由誰寫，本規格不宣稱。

## ★★ 二、磁碟空間檢查

```pascal
結果 := 'O';
可用 := DiskFree(UpCase(DS:5BF1h) − 40h);      { 097F:00DEh，32-bit }
if 需要 > 可用 then begin
    顯示("Can't save.  No room on this disk.", 0, 0Eh);     { 0542:0946h }
    行1 := 'Lose character? ';
    行2 := 'Ok  Try another disk';
    結果 := <overlay-26 entry#3>(行1, 行2, 0, 0, 0Fh, 0Ah, 0Dh);
end;
if 結果 <> 'O' then 回到最上面重做;
```

★ `UpCase(字母) − 40h` ⇒ `'A'` → 1、`'B'` → 2，正是 Borland `DiskFree` 的
磁碟機編號（0 ＝ 目前）。
★ 選項行 `'Ok  Try another disk'` 的熱鍵是 `O` 與 `T`（spec 1060：DOS 沒有熱鍵表，
直接掃選項文字找大寫字母）。

> ★★★ **`DS:5BF0h` 是存檔路徑字串，而 `DS:5BF1h` 是它的第一個字元
> ——也就是磁碟機代號。** 全檔一路用 `DS:5BF0h ＋ 名字 ＋ 副檔名` 組檔名，
> 只有算空間與提示句時單獨取 `DS:5BF1h`。
> ⇒ spec 1038 記的「主檔名（`DS:5BF0h` 那個字串 ＋ 目錄）」在這裡定案。

## ⚠ 三、副檔名：與 spec 1043 的對應表衝突

```pascal
if 名字 = '' then 副檔名 := '.guy'      { CS:0D64h }
else               副檔名 := '.sav';    { CS:0D6Bh }
```

名字為空時另外走 `0636:0549h` ＋ `0636:0618h` 產生一個名字。

> ⚠⚠ **spec 1043 從 PC-98 的並列表推出「副檔名就是記錄長度」
> （`.guy` ＝ 423／`.cha` ＝ 285／`.sav` ＝ 188）。
> 但 DOS 這一支把 `1A6h` ＝ 422（現行格式，`DS:7584h := 0`）寫進 `.sav`。**
> 兩者不能同時成立。
>
> 本規格**只宣稱 DOS 的寫入端事實**：`.sav` 檔裡放的是現行長度的角色記錄。
> **不宣稱** spec 1043 的表對應錯在哪——那張表的三個值
> （423／285／188）與 spec 1038 的三種長度完全吻合，證據並不弱，
> 而且兩支表基底互相重疊（spec 1043 自己標了 ⚠）。
> 要定案得回頭重讀 `pc98 overlay-16:00614h` 的索引算式。
> 在那之前，**remake 不可以拿副檔名當格式判斷**。

## 四、覆寫確認與重新命名

```pascal
結果 := 'N';
while 結果 = 'N' do begin
    if 名字 <> '' then break;                          { 呼叫端指定名字就不問 }
    if not <0636:04FFh>(DS:5BF0h ＋ 名字2 ＋ 副檔名) then break;   { 檔案不存在 }
    結果 := <overlay-26 entry#6>('Overwrite ' ＋ 名字2 ＋ '? ', 0Fh, 0Ah, 0Eh);
    if 結果 = 'N' then begin
        名字2 := '';
        while 名字2 = '' do begin
            <0542:0722h>('New file name: ', 8, 0, 0Ah);   { ★ 上限 8 }
            名字2 := <輸入結果>;
        end;
    end;
end;
```

★ `overlay-26 entry#6` 是 spec 1060 的「問到合法按鍵為止」。
★ **新檔名的輸入上限也是 8 個字元**，與進門那個 `064Eh` 的上限一致。

## 五、I/O 錯誤處理

```pascal
if not (DS:8C98h in [0, 2]) then begin              { 集合常數 CS:0D70h ＝ [0,2] }
    if UpCase(DS:5BF1h) < 'C' then
        顯示('Put save disk in ' ＋ DS:5BF1h ＋ ':', 0, 0Eh)
    else begin
        顯示('Unexpected error during save: ' ＋ Str(錯誤碼), 0, 0Eh);
        Close(檔案);  exit;
    end;
end;
```

★ `CS:0D70h` 是 32 bytes 的**集合常數**，第一個 byte ＝ `05h` ⇒ 成員是 `{0, 2}`
（0 ＝ 成功、2 ＝ 檔案不存在，兩者在「準備建檔」的當下都算正常）。
⚠ 之前把 `CS:0D70h` 當成 5 bytes 的字串會得到「五個 NUL」這種讀不通的結果
——**遇到 `0A54:08D4h`（`@Set@MemberOf`）就要把運算元當 32 bytes 位元集合讀**。
★ **磁碟機是 A 或 B（軟碟）才提示換片**，C 以上直接當成非預期錯誤。

## 六、實際寫檔

```pascal
Assign(檔案, DS:5BF0h ＋ 名字2 ＋ 副檔名);  Rewrite(檔案, 1);
BlockWrite(檔案, 角色^, 1A6h, NIL);         Close(檔案);

<0636:0192h>(DS:5BF0h ＋ 名字2 ＋ '.swg');           { 形狀上是刪掉舊檔 }
if 角色^[14Dh] <> NIL then begin
    Assign(…'.swg');  Rewrite(檔案, 1);
    節點 := 角色^[14Dh];
    while 節點 <> NIL do begin
        BlockWrite(檔案, 節點^, 3Fh, NIL);  節點 := 節點^[2Ah];
    end;
    Close(檔案);
end;

<0636:0192h>(DS:5BF0h ＋ 名字2 ＋ '.fx');
if 角色^[0F2h] <> NIL then begin … 同上，每筆 9 bytes，next 在 +5 … end;
```

★ 三個 `BlockWrite` 的第 4 個參數都推 `0, 0` ＝ NIL（不收寫入位元組數）。
★ **鏈為空時連檔案都不建**，但**舊檔一律先刪掉**——所以
「這一版沒有物品」與「還留著上一版的物品檔」不會混淆。

## Resident 對照（DOS）

| 位址 | 對應 |
|---|---|
| `0A54:181Ah` ／ `1851h` ／ `18C9h` | `Assign` ／ `Rewrite` ／ `Close` |
| `0A54:193Ah` ／ `1848h` | `BlockWrite` ／ `Reset` |
| `0A54:0634h` ／ `06C1h` ／ `064Eh` | 指派（留目的） ／ 串接（留目的） ／ 上限指派（全收） |
| `0A54:08D4h` ／ `1C62h` ／ `074Fh` | `@Set@MemberOf` ／ `UpCase` ／ 字元轉字串 |
| `097F:00DEh` | `DiskFree`（32-bit） |
| `0542:0946h` ／ `0722h` | 顯示訊息 ／ 讀入一行 |
| `0636:04FFh` ／ `0192h` ／ `0549h` ／ `0618h` | 檔案存在？ ／ 刪除 ／ 產生名字（兩支） |

## 中文化

| DOS | 位址 | 長度 | 建議 |
|---|---|---|---|
| `"Can't save.  No room on this disk."` | `CS:0D1Bh` | 34 | 「磁片空間不足，無法存檔。」 |
| `'Lose character? '` | `CS:0D3Eh` | 16 | 「要放棄這個角色嗎？」 |
| `'Ok  Try another disk'` | `CS:0D4Fh` | 20 | ⚠ **熱鍵 `O`／`T` 必須留在文字裡**（spec 1060） |
| `'Put save disk in '` | `CS:0D90h` | 17 | 「請放入存檔磁片到 」（spec 1038 同一句） |
| `'Unexpected error during save: '` | `CS:0DA4h` | 30 | 「存檔時發生非預期錯誤：」 |
| `'Overwrite '` ／ `'? '` | `CS:0DC3h`／`0DCEh` | 10／2 | 「要覆蓋 」／「 嗎？」 |
| `'New file name: '` | `CS:0DD1h` | 15 | 「新檔名：」 |
| `'.guy'`／`'.sav'`／`'.swg'`／`'.fx'`／`'.'`／`':'` | — | 4／4／4／3／1／1 | ⚠ **副檔名，一律不可翻** |

⚠ **檔名輸入上限是 8 個字元** ⇒ 中文全形**最多 4 字**，而且 DOS 檔名
不接受 Big5 高位元組落在保留字元上的組合。remake 建議直接沿用英數檔名。
⚠ `'Overwrite ' ＋ 名字 ＋ '? '` 是三段串接，中文要照顧語序
（「要覆蓋 <名字> 嗎？」）。

## 明確不宣稱

- 沒有宣稱 `near 0643h`（進門那道閘）判什麼。
- 沒有宣稱 `0636:0549h`／`0618h`（名字為空時自動產生檔名）怎麼產生。
- 沒有宣稱 `0636:0192h` 一定是「刪除」（回傳值沒被用，只能從位置推）。
- 沒有宣稱空間不足時選 `'O'` 之後為什麼還是往下寫
  （控制流是「回到最上面重做」只在 `結果 <> 'O'` 時發生）。
- 沒有宣稱 `.spc` 那條鏈由誰寫。
- 沒有宣稱 `.sav` 與 spec 1043 對應表的衝突該怎麼解（見上 ⚠⚠）。
- 沒有宣稱 PC-98 的對應函式是哪一支
  （`pc98 overlay-16:014C9h` 被 IDA 切成邊界碎片）。
