# 752 — 空程序（`55 89 E5 89 EC 5D CB`）

- 平台／模組：兩平台各 overlay，共 77 處
- 證據等級：`exact`（7 個 byte 全部比對）

## 位元組就是全部

```
55        push bp
89 E5     mov  bp, sp
89 EC     mov  sp, bp
5D        pop  bp
CB        retf
```

沒有配置 local、沒有參數（`retf` 無運算元）、沒有任何副作用。**這是一支空的
Pascal procedure**——原始碼裡多半是宣告了但 body 為空，或某個平台上被整段
`{$IFDEF}` 掉。

## 為什麼會被當成「還沒讀」

IDA 的匯出在 `89 EC` 這裡會漏掉一個 byte，剩下的 `EC` 解成 `in al, dx`（DX 埠
輸入），於是這 7 個 byte 在匯出裡長成：

```
push bp / mov bp, sp / in al, dx / pop bp / retf
```

看起來像「讀一個 I/O 埠然後丟掉」——**語意上並非明顯荒謬**，所以不會自己浮出
水面。有幾支甚至連 `mov bp, sp` 都被吃掉，變成 `push bp / in ax, 89h /
in al, dx / pop bp / retf`（IDA 還好心加註「DMA page register」）。這是 spec 736
記錄的同一種漏 byte，只是這次整支函式都被它吃掉了。

## 全庫盤點

以這 7 個 byte 為樣板掃過索引裡的全部函式：**77 處命中**，其中 61 處先前已判為
已解讀、7 處歸在邊界碎片、9 處仍是待解讀。本輪把這 9 支補上：

| 平台 | 模組 | 位址 |
|---|---|---|
| DOS | `overlay-01` | `0774h` |
| DOS | `overlay-09` | `1BC6h` |
| DOS | `overlay-11` | `08DCh` |
| DOS | `overlay-17` | `56BCh` |
| DOS | `overlay-23` | `25B6h` |
| DOS | `overlay-33` | `061Eh` |
| PC-98 | `overlay-03` | `03FCh` |
| PC-98 | `overlay-20` | `0EBCh` |
| PC-98 | `overlay-33` | `08C6h` |

## 對 remake 的意義

呼叫端仍然會呼叫它們，所以**不能直接刪掉**——會影響參數平衡的沒有（無參數），
但呼叫序列是行為的一部分。實作時保留為空函式即可。

## 明確不宣稱

- 沒有宣稱這 77 處在原始碼裡是同一支程序。位元組相同不代表同源；空 body 的
  編譯結果本來就都長一樣。
- 沒有宣稱它們「本來應該做什麼」。
