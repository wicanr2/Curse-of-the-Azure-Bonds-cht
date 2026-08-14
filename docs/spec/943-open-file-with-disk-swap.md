# 943 — 開檔：兩個路徑輪流試，軟碟就提示換片

- 證據等級：`exact`（逐條讀完；字串取自 `START.EXE.asm`）
- 位置：DOS `START.EXE` 的 `16645h`
- 作法見 spec 783

## `16645h`（219 條）—— `retf 0Eh`

`(檔案變數 arg_0, 不提示 arg_4, 換片提示前綴 arg_6, 檔名 arg_A)`。
回傳 `al`：成功／失敗。

| 字串 | 內容 |
|---|---|
| `asc_16623` | `':'` |
| `aCouldnTFind` | `"Couldn't find "` |
| `aCheckInstall` | `". Check install."` |

```pascal
名   := 限長(arg_A, 50h);
提示 := 限長(arg_6, 50h);
FSplit(名, 目錄, 主檔名, 副檔名);
if 目錄 = '' then 目錄 := 限長(byte_210B2h, 43h);      { 預設路徑 }

repeat
    成功 := <sub_1685Fh>(目錄 ＋ 主檔名 ＋ 副檔名);

    if (not 成功) and (目錄 = byte_210B2h) then begin
        { ★ 交換兩個路徑全域 }
        暫 := byte_210B2h;
        byte_210B2h := byte_21102h;
        byte_21102h := 暫;
        目錄 := 限長(byte_210B2h, 43h);
        成功 := <sub_1685Fh>(目錄 ＋ 主檔名 ＋ 副檔名);
    end;

    if (not 成功) and (arg_4 = 0) then begin
        if 目錄[1] < 'C' then                          { ★ 磁碟機代號 }
            <sub_16498h>(提示 ＋ byte_21152h ＋ ':', 0Eh)
        else
            <sub_16498h>("Couldn't find " ＋ 主檔名 ＋ 副檔名
                         ＋ ". Check install.", 0Eh);
    end;
until 成功 or (arg_4 <> 0);

if 成功 then begin
    Assign(arg_0, 目錄 ＋ 主檔名 ＋ 副檔名);
    Reset(arg_0, 1);                                    { record size 1 }
end;
結果 := 成功;
```

## ★ 兩個路徑輪流，而且會**互換**

`byte_210B2h` 與 `byte_21102h` 是兩個 80 bytes 的路徑字串
（兩者相距 `50h`，都由 `sub_1236Ch` 初始化）。
第一次找不到就把兩者**對調**，用新的再試一次。

對調是**永久的**——下一次呼叫會先試上次成功的那一個。
這是雙磁碟機（或安裝目錄 ＋ 光碟目錄）情境下的簡易快取：
**玩家換過一次片之後，後續的檔案會優先從同一片找**。

## ★ 用磁碟機代號決定提示哪一種訊息

```
cmp [bp+var_137], 43h      ; var_137 是目錄字串的第一個字元
jnb  → 走 "Couldn't find …"
     → 否則走換片提示
```

`43h` ＝ `'C'`。**第一個字元小於 `'C'`（也就是 `A:` 或 `B:`）當成軟碟**，
提示 `<呼叫端給的前綴> ＋ <磁碟機代號 byte_21152h> ＋ ':'`；
`C:` 以上當成硬碟，提示 `"Couldn't find <檔名>. Check install."`。

兩種訊息都用 spec 932 的 `<sub_16498h>` 印在第 24 列、等一次按鍵。

## `arg_4` ＝ 不要提示

`arg_4 <> 0` 時**既不提示也不重試**，找不到就直接回失敗。
spec 939 的資源開啟就是走這一條（它自己處理失敗）。

## 中文化

`"Couldn't find "`、`". Check install."` 是**找不到檔案時的致命提示**，
與 spec 926 的 `'Unable to load '` 分屬兩層：
本支會等玩家換片重試，spec 926 那一支直接結束程式。

換片提示的前綴由呼叫端提供，本支只負責接上磁碟機代號與 `':'`。

## 明確不宣稱

- 沒有宣稱 `<sub_1685Fh>`（試開檔）的內部。
- 沒有宣稱 `byte_21152h` 那個字元由誰算出（`sub_1236Ch` 寫的，尚未解讀）。
- 沒有宣稱兩個路徑字串的初值。
