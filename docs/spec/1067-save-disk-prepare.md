# 1067 — 存檔前的磁碟準備：`'CURSE'` 目錄、磁碟機字母 < `'C'` 才問換片

- 證據等級：`exact`（DOS 側 381 條逐條讀完）
- 作法見 spec 783

## `dos overlay-16:00643h`（`retf`）

原本待解讀。三條路，由 **`DS:7584h`（角色來源，spec 1066）** 分：

| `DS:7584h` | 來源 | 這一支做什麼 |
|---|---|---|
| `0` | Curse（本作） | **要求 `'CURSE'` 這個目錄**，見下 |
| `1` | Pool of Radiance | 只建目錄／開檔，不檢查目錄名 |
| `2` | Hillsfar | （第三段） |

## ★★ `DS:7584h = 0`：路徑最後一段必須是 `'CURSE'`

```pascal
repeat
    路徑 := 磁碟機字母(DS:5BF1h) ＋ …;
    <far 0636:0257h>(路徑, @最後一段);
    if (最後一段 <> 'CURSE') and (UpCase(DS:5BF1h) < 'C') then begin
        顯示('Put save disk in ' ＋ 磁碟機字母 ＋ ':');
        continue                                   { ★ 換片後重試 }
    end;
    …
until 成功;
```

★ **`'CURSE'`（5 bytes）是存檔目錄名**——本作的存檔片要有這個目錄。
★ **`UpCase(磁碟機字母) < 'C'`** ⇒ 只有 A／B 槽（軟碟）才會叫玩家換片；
C 槽以上（硬碟）直接跳過。

## ★★ 開檔與錯誤處理

```pascal
路徑 := Copy(DS:5BF0h, 1, 長度 − 1);
<far 097F:0112h>(路徑, 10h);                       { 形狀上是 MkDir／ChDir }
錯誤 := DS:8C98h;                                   { ★ IOResult }
case 錯誤 of
    0:            成功;
    2, 12h:       …見下…;
    else          …顯示 'Unexpected error during save: ' ＋ Str(錯誤)…;
end;
```

★ 錯誤 `2`（檔案不存在）或 `12h`（18 ＝ 沒有更多檔案）時：
若磁碟機 < `'C'` 就顯示 `'Is your save disk in drive X:? '`
（`'Is your save disk in drive '` 27 ＋ 字母 ＋ `':? '` 3）並等 `'Y'`；
不是 `'Y'` 就改顯示 `'Put save disk in X:'` 再重試。

★ 兩個集合常數都在 `CS:061Dh`（`'Unexpected error during save: '` 之後），
用來判「錯誤碼要重試還是放棄」。

## 中文化

| DOS | 長度 | 建議中文 |
|---|---|---|
| `'CURSE'` | 5 | ⚠ **目錄名，不可翻** |
| `'Put save disk in '` ＋ 字母 ＋ `':'` | 17 | 「請放入存檔磁片到 」 |
| `'Is your save disk in drive '` ＋ 字母 ＋ `':? '` | 27 ＋ 3 | 「存檔磁片在 」＋字母＋「 槽嗎？」 |
| `'Unexpected error during save: '` ＋ 錯誤碼 | 30 | 「存檔時發生非預期錯誤：」 |

⚠ 三句都是**字母／數字接在中間或結尾**的組字，中文語序要重排。
⚠ `'CURSE'` 同時是遊戲名縮寫與目錄名——**翻了會找不到存檔片**。

## 明確不宣稱

- 沒有宣稱 `far 097F:0112h`（帶 `10h`）到底是 MkDir、ChDir 還是 FindFirst。
- 沒有宣稱 `far 0636:0257h`（取路徑最後一段）的介面。
- 沒有宣稱 `DS:8C98h` 之外還有沒有別的錯誤來源。
- 沒有宣稱 `DS:7584h = 2`（Hillsfar）那一段與前兩段的差別
  （本次只讀到它是第三個分支）。
- 沒有宣稱兩個集合常數的完整成員。
