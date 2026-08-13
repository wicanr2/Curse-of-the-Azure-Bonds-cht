# 第六百四十四輪：resident 裡的 99 支 TPOV overlay stub

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 50 支、DOS `START.EXE` 49 支。

## 形狀

IDA 在兩個 resident 執行檔裡各認出一批「函式」，body 只有兩個 byte：

```asm
sub_10320       proc near
                int     3Fh             ; Overlay manager interrupt
sub_10320       endp                    ; (Microsoft LINK.EXE, Borland TLINK VROOMM)
; ---------------------------------------------------------------------------
                db 55h, 2 dup(0), 0CDh, 3Fh, 1Bh, 2 dup(0), 0CDh, 3Fh
```

`INT 3Fh` 之後緊接**三個 byte 的描述子**，然後就是下一支 stub 的 `CD 3F`——
以 5 bytes 為間隔連續排列。這是 Borland TLINK 的 **VROOMM** overlay 機制：呼叫
被 overlay 化的 unit 裡的程序時，編譯器產生的是這個 stub，overlay manager 在執行
時把它換成真正的 far call。

兩平台的數量幾乎一樣：

| 平台 | 檔案 | 匯出函式數 | `INT 3Fh` stub |
|---|---|---|---|
| PC-98 | `PC98-GAME.EXE` | 271 | **50** |
| DOS | `START.EXE` | 325 | **49** |

全部 stub 的 code body 長度都是 **2 bytes**，沒有例外（長度分佈只有一個值）。

## 分類為「不阻塞」

這些 stub **不實作任何遊戲規則**——它們是連結器產生的派送機制，行為完全由
overlay manager 決定。remake 不會有對應物（Go 沒有 overlay），所以列為不阻塞而
不是待解讀。

描述子那三個 byte 指向哪個 overlay 的哪個程序，本輪**沒有解析**。真要對應回
unit／entry，`workplace/re-sweep/*/ovr-manifest.json` 已經有 TPOV control block
的 entry 表（[spec 562](562-tpov-control-block-entry-index.md)），走那條路比反推
描述子直接。

## 明確不宣稱

- 三個描述子 byte 的欄位切法。
- 每一支 stub 對應哪一個 overlay unit 的哪一個程序。
- 兩平台數量差 1（50 對 49）的原因。
