# 第六百五十三輪：RLE 解壓縮器

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE:17DD5h`（86 bytes）。

## 演算法

```text
DS:SI := arg_8:arg_6          ← 來源
ES:DI := arg_4:arg_2          ← 目的
BX    := SI + arg_0           ← 來源結尾（arg_0 是長度）

SI := SI + 1                  ← 跳過第一個 byte
DH := [SI] ; SI := SI + 1     ← 預設填充 byte

repeat
    CH := 0
    CL := [SI] ; SI := SI + 1                 ← 控制 byte
    if (CL and 80h) = 0 then
        rep movsb                              ← 原樣複製 CL 個 byte
    else if (CL and 40h) <> 0 then
        CL := (CL and 3Fh) + 2
        用 DH 填 CL 個
    else
        CL := (CL and 3Fh) + 3
        AL := [SI] ; SI := SI + 1
        用 AL 填 CL 個
until SI >= BX
```

## 控制 byte 的三種形式

| bit 7 | bit 6 | 長度 | 內容 |
|---|---|---|---|
| `0` | — | `CL`（低 7 bit 直接用） | 之後的 `CL` 個 byte 原樣複製 |
| `1` | `1` | `(CL and 3Fh) + 2` | 重複 **`DH`**（串流開頭指定的預設 byte） |
| `1` | `0` | `(CL and 3Fh) + 3` | 重複**下一個 byte** |

兩種重複形式的**起始長度不同**：用預設 byte 的最短 2 個、用即時 byte 的最短 3 個。
這不是筆誤——`+2` 那條是 `inc cl` 兩次，`+3` 那條是三次，指令數就差在那裡。

長度上限都是 `3Fh + 起始值`，也就是 65 與 66。

## 串流開頭

前兩個 byte **不是資料**：第一個被跳過（`inc si` 之後才讀），第二個是 `DH`
（預設填充 byte）。控制資料從第三個 byte 開始。

第一個 byte 的用途本輪未確認——它被完全略過，沒有任何指令讀它。

## 結束條件是「來源用完」

`until SI >= BX`，`BX = 起點 + arg_0`。所以**呼叫端要給來源長度**，解壓端不知道
輸出應該多長，也不檢查目的緩衝的大小。輸出溢位不會被發現。

比較用 `jb`（**無號**）。

## 誰在用

[spec 652](652-disk-swap-loop-and-third-leadbyte.md) 的 `1790Dh` 是二選一的分派：

```text
if arg_C^^[0] = 0FFh then <sub_1C15D>(…)      ← 另一支，多半是原樣複製
else                      <17DD5h>(…)          ← RLE 解壓
```

所以 **`0FFh` 開頭代表「未壓縮」**，其餘走 RLE。而 `1790Dh` 傳給 `17DD5h` 的
`arg_C^^` 是完整的 dword（未壓縮那條則是 `arg_C^^ + 1`，跳過標記 byte）。

## 明確不宣稱

- 串流第一個 byte 的用途。
- `sub_1C15D` 是不是單純的 `Move`。
- 這個格式用在哪些資料（圖、地圖、文字都有可能）。
