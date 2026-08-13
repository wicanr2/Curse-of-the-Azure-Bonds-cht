# 第六百一十九輪：`GETECL` —— script 檔怎麼載入

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-07:0499h`（189 bytes，84 條指令）。

```text
GETECL(n):
    FillChar(bank3, 1E41h, 0)                 ← 清空整個 script 緩衝區
    DS:8CF6h := 0
    repeat
        <顯示>「読みこみ中です」（`unk_481h`）
        Str(DS:8BF4h, s)
        name := 'ECL' + s + '.dax'            ← `unk_490h` ＋ `unk_494h`
        <讀檔>(n, @size, @buf)
    until size >= 2                            ← 讀到 2 bytes 以上才算成功
    Move(buf + 2, bank3, size − 2)            ← 跳過前 2 bytes
    out 0A6h, 0
```

## 三件事

1. **檔名是 `ECL<n>.dax`**——`'ECL'` 與 `'.dax'` 是函式前方的兩個字串常數，
   中間接一個由 `DS:8BF4h` 轉成的數字。DOS 側的字串掃描也掃到同樣的
   `.dax`（`overlay-02:1B35h`）。
2. **script 緩衝區是 `1E41h` ＝ 7,745 bytes**。這與
   [spec 563](563-ecl-memory-model-and-operand-resolution.md) 從位址分類器解出
   的 bank 3 範圍 `8000h..9E40h` **完全一致**（`9E40h − 8000h + 1 = 1E41h`）——
   兩個獨立來源對上。
3. **檔案的前 2 bytes 被跳過**，`Move` 從 `buf + 2` 起、長度 `size − 2`。所以
   `.dax` 檔開頭有 2 bytes 不屬於 script 本體；**script 的 `8000h` 對應的是
   檔案的 offset 2**。

remake 讀 `.dax` 時若不跳過那 2 bytes，所有位址都會差 2——包括
[spec 618](618-ecl-script-header.md) 讀出的五個進入點。

## 讀檔失敗會無限重試

`until size >= 2` 之前沒有其他跳出路徑：讀不到就回到迴圈開頭重新顯示
「読みこみ中です」再試一次。**沒有錯誤處理，也沒有重試上限。**

## 明確不宣稱

- `DS:8BF4h`（決定檔名數字）由誰設定。
- 被跳過的前 2 bytes 是什麼（長度？版本？）。
- `out 0A6h, 0` 在 PC-98 上的作用（`0A6h` 是中斷控制器區段，但用途未確認）。
