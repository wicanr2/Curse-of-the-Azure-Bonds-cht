# 966 — PC-98 的堆積錯誤：`ヒープエラーが発生しました。`

- 證據等級：`exact`（逐條讀完，訊息由 `PC98-GAME.EXE.asm` 的位元組解碼）
- 位置：`PC98-GAME.EXE` 的 `18407h`
- 作法見 spec 783

## `18407h`（145 行，far，`retf 6`）

`(要求大小 arg_0, 檔案 arg_2)`。由 `seg041:01BDh` 與 `seg041:0234h` 呼叫。

```pascal
原始大小 := 要求大小;
if (要求大小 mod 10h) <> 0 then 要求大小 := 要求大小 ＋ 10h;
要求大小 := 要求大小 div 10h;                     { ★ 換算成段數 }

Assign(f, '');                                     { 空檔名 ＝ 主控台 }
Rewrite(f);
Write(f, #1Bh);                                    { ESC }
Write(f, '[32m');                                  { ★ ANSI 綠色 }
WriteLn(f);

if <sub_188E3h>(要求大小, 檔案) = 0 then begin     { ★ 配置失敗 }
    Write(f, 'ヒープエラーが発生しました。');
    WriteLn(f);
    Write(f, word_280B6h);                         { 載入計數 }
    Write(f, '_');
    Write(f, word_280B2h);                         { 主版本 }
    Write(f, ':0__');
    Write(f, 要求大小);                            { 段數 }
    WriteLn(f);
    Close(f);
    <sub_1A726h>(1);                               { ★ Runtime error 1，不返回 }
end;

BlockRead(檔案 ＋ 2, var_6, 2);
inc(word_280B6h);
if word_280B2h < var_6 then begin
    word_280B2h := var_6;
    word_280B4h := 0;
end;
Close(f);
```

## ★ 錯誤訊息

```
byte_183E5  db 1Ch, 83h,71h, 81h,5Bh, 83h,76h, 83h,47h, 83h,89h, 81h,5Bh,
                 82h,0AAh, 94h,0ADh, 90h,0B6h, 82h,0B5h, 82h,0DCh,
                 82h,0B5h, 82h,0BDh, 81h,42h
```

長度 `1Ch` ＝ 28 bytes ＝ **14 個全形字**，Shift-JIS 解出來是

> **`ヒープエラーが発生しました。`**（發生了堆積錯誤。）

後面接的診斷資訊格式是 `<載入計數>_<主版本>:0__<段數>`。

⚠ `:0__` 裡的 `0` 是**寫死的常數**，不是 `word_280B4h`（次版本）——
所以次版本永遠印成 0。

## 訊息走主控台，不走遊戲畫面

`Assign(f, '')` 的空檔名在 Turbo Pascal 裡代表**標準輸出**，
前面還先送了 `ESC [32m`（ANSI 綠色）。
**這是繞過遊戲的文字層直接印到主控台**——
所以出錯時畫面上會殘留遊戲的內容，訊息疊在上面。

之後呼叫 `<sub_1A726h>(1)`，也就是 spec 955 的
**`Runtime error 1`**（Turbo Pascal 的 `ovrError`／一般錯誤）。

## 與 spec 965 的關係

`18233h`（spec 965）與本支是**同一件事的兩個版本**：
都做「對齊大小 → 配置 → 讀標頭 → 更新版本號 → 計數」。
差別是：

| | `18233h`（spec 965） | `18407h`（本支） |
|---|---|---|
| 對齊粒度 | `40h`（64 bytes），再 `shl 2` | **`10h`（16 bytes ＝ 一個段）** |
| 配置函式 | `<sub_18883h>` | **`<sub_188E3h>`** |
| 失敗處理 | **沒有**（結果回給呼叫端） | **印訊息並 `Runtime error 1`** |
| 版本更新 | 主相同更新次／主較小更新兩者 | **只更新主，次直接歸 0** |

## 明確不宣稱

- 沒有宣稱 `<sub_188E3h>`／`<sub_1B9C7h>`／`<sub_1BA41h>`／`<sub_1BD76h>`／
  `<sub_1BDDBh>`／`<sub_1BE71h>`／`<sub_1BD32h>`／`<sub_1BD13h>`／
  `<sub_1BA90h>`／`<sub_1C15Dh>` 的內部（都是 Turbo Pascal 的 Text I/O 常式）。
- 沒有宣稱 `word_280B2h`／`word_280B4h` 是不是版本號（同 spec 965 的保留）。
- 沒有宣稱 DOS 版有沒有對應的訊息。
