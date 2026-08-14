# 1020 — 存檔畫面：十個槽、`savgamX.dat`、12,808 bytes 的空間檢查

- 證據等級：`exact`（DOS 側 532 條逐條讀完；PC-98 側 575 條——
  與 DOS 相同的核心逐條比對過，兩塊 PC-98 專屬區段另外逐條讀完）
- 作法見 spec 783

## `dos overlay-16:03EDDh` ↔ `pc98 overlay-16:05008h`（`retf`，不收參數）

兩側原本都是待解讀。

## ★★ 十個存檔槽 A..J

```pascal
repeat
    提示 := 'Save Which Game: ';
    選項 := 'A B C D E F G H I J';
    鍵 := <overlay-26 entry#3>(…, ord(DS:4FBAh = 2), …);
until 鍵 in [#0, 'A'..'J'];
if 鍵 = 0 then exit;                         { 取消 }
```

那個集合的 32 bytes 是 `01 00 …  fe 07 00 …` ⇒ `[#0, 'A'..'J']`，
用 `@Set@MemberOf$q4Byte`（`0A54:08D4h`）判斷——與 spec 1012 同一支 RTL helper，
**運算元一樣是集合點陣圖不是字串**。

★ 選單的第五個參數是 `ord(DS:4FBAh = 2)`——**畫面模式 2 時選單長得不一樣**
（`DS:4FBAh` ＝ 畫面模式，spec 892）。

## ★★ 檔名：`<路徑> ＋ 'savgam' ＋ <字母> ＋ '.dat'`

```pascal
檔名 := 字串(DS:5BF0h) ＋ 'savgam' ＋ Chr(鍵) ＋ '.dat';
```

`DS:5BF0h` 在靜態映像裡是 `0FFh`，**執行期才填**（路徑／磁碟機前綴）；
`DS:5BF1h` 是磁碟機代號（後面 `− 40h` 換成 1 ＝ A、2 ＝ B…）。

## ★★★ 存檔要 12,808 bytes

```pascal
if <far 0636:04FFh>(檔名) then 需要 := 0        { 檔案已存在 ⇒ 覆寫不用新空間 }
                          else 需要 := 3208h;   { ＝ 12808 }
需要 += <sub_3D38>();
if 需要 > DiskFree(DS:5BF1h − 40h) then begin   { far 097F:00DEh }
    顯示 "Can't save.  No room on this disk.";
    exit;
end;
```

> ★★ **一個存檔的固定大小是 `3208h` ＝ 12,808 bytes**，
> 再加上 `sub_3D38()` 算出來的可變部分（隊伍人數之類）。

## 開檔與錯誤處理

```pascal
Assign(f, 檔名);  Reset/Rewrite(f, 1);
if not (IOResult in [0, 2, 18]) then begin
    顯示 'Put save disk in ' ＋ Chr(DS:5BF1h) ＋ ':';
    …或…
    顯示 'Unexpected error during save: ' ＋ Str(IOResult);
    Close(f);
    exit;
end;
顯示 'Saving...Please Wait'（欄 0、列 18h、顏色 0Ah on 0）;
…開始把全域寫進檔案（`DS:4FA9h` 遊戲速度 → `+1F8h` 等）…
```

★ 可接受的 `IOResult` 是 **`[0, 2, 18]`**（0 成功、2 檔案不存在——新檔正常）。

## ★★★ PC-98 是另一套存檔管理

共用的核心（把全域寫進檔案那一大段，254 條）兩邊逐條相同，
但**前後的錯誤處理與槽位管理完全不同**：

| | DOS | PC-98 |
|---|---|---|
| 提示 | `'Save Which Game: '` ＋ `'A B C D E F G H I J'` | **沒有這兩個字串** |
| 空間檢查 | `3208h` ＋ `DiskFree` | **沒有** |
| 錯誤訊息 | `"Can't save.  No room on this disk."`／`'Put save disk in '` | 「ライトプロテクトをはずしてください。」（36 bytes） |
| 存檔失敗 | `'Unexpected error during save: '` | 「セーブ中にエラーが発生しました-」（32 bytes）＋ IOResult |
| 進度訊息 | `'Saving...Please Wait'`（20） | 「書きこみ中です」（14 bytes ＝ 7 全形字） |
| 額外檔案 | — | **`SaveList.EST`**、**`CHRDAT`** |

★★ **PC-98 多維護一個 `SaveList.EST` 的存檔目錄檔**：

```pascal
檔名 := 字串(DS:8BF6h) ＋ 字串(DS:0A23h) ＋ 'SaveList.EST';
Assign / Rewrite(f, 1);
byte[0A80Ah ＋ DS:0A670h] := 6;              { 失敗路徑寫常數 6 }
      { 成功路徑寫 DS:0BDF0h }
BlockWrite(f, DS:0A80Bh, 0Ah);               { ★ 10 bytes ＝ 十個槽各一個 byte }
Close(f);
```

★ 成功那一段前面還有一道 `if 槽字母 < 'K' then …`——**只有 A..J 才更新目錄**。

⚠ `mov di, offset loc_A23` 的運算元其實是 **`DS:0A23h`**，不是 overlay 內的程式碼；
和 spec 1006 的 `word_AF6` 同一類**假交叉參照**。

## 中文化

| 字串 | 長度 | 備註 |
|---|---|---|
| `'Save Which Game: '` | 17 | PC-98 沒有 |
| `'A B C D E F G H I J'` | 19 | 槽位標示，PC-98 沒有 |
| `"Can't save.  No room on this disk."` | 34 | PC-98 沒有 |
| `'Put save disk in '` ＋ `':'` | 17 ＋ 1 | 中間夾磁碟機字母 |
| `'Unexpected error during save: '` | 30 | 後面接數字 |
| `'Saving...Please Wait'` | 20 | PC-98 對應「書きこみ中です」7 字 |
| `'savgam'`／`'.dat'` | 6／4 | **檔名，不可翻** |

⚠ **`'savgam'` 與 `'.dat'` 是檔名的一部分，翻了會存不進去。**
⚠ 槽位字母 `A`..`J` 同時是**熱鍵集合**（`[#0, 'A'..'J']`）與**檔名的一部分**，
中文化不能把它們換成中文數字——會同時弄壞熱鍵與檔名。

## 明確不宣稱

- 沒有宣稱 `sub_3D38()` 算的是什麼可變空間。
- 沒有宣稱 `IOResult = 18` 為什麼算可接受。
- 沒有宣稱 `SaveList.EST` 那 10 個 byte 的值（`6` 與 `DS:0BDF0h`）代表什麼。
- 沒有宣稱 `'CHRDAT'`（PC-98 才有）在這一支的哪一段用到。
- 沒有宣稱寫進檔案的全域清單（共用核心那 254 條的細節）。
