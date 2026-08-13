# 第六百四十七輪：PC-98 調色盤設定、屬性平面填色，以及兩組重複的常式

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `18194h`、`181C7h`、`17CE9h`、`17D72h`、`17DA9h`、
`180F7h`。

## `181C7h`：寫 16 色類比調色盤

```asm
mov  ax, cs
mov  ds, ax
mov  cx, 10h              ; 16 色
迴圈:
dec  bx
mov  ax, bx
out  0A8h, al             ; 調色盤索引
mov  al, [si] ; out 0AAh, al ; inc si    ; 綠
mov  al, [si] ; out 0ACh, al ; inc si    ; 紅
mov  al, [si] ; out 0AEh, al ; inc si    ; 藍
loop 迴圈
```

`A8h`／`AAh`／`ACh`／`AEh` 是 PC-98 的類比調色盤埠：先寫索引，再依序寫綠、紅、藍
（**PC-98 的順序是 G-R-B，不是 R-G-B**）。每色 4 bit，所以資料值都在 `0..0Fh`。

索引從 `0Fh` 遞減到 `0`（`dec bx` 在迴圈開頭），每筆讀 3 個 byte，資料在 **`CS:SI`**。

## `18194h`：選一組調色盤並切換模式

```text
out 6Ah, 1                      ; 開啟類比模式
si := 304h
if arg_0 = 0 then si := 334h    ; 兩組表二選一
<181C7h>()                      ; 寫入調色盤
int 18h  (AH=42h, CH=80h)       ; 設定顯示模式
out 68h, 8
int 18h  (AH=40h)               ; 顯示開
int 18h  (AH=12h)
```

兩張表在 `seg045` 的 `304h` 與 `334h`——**相鄰**（`304h + 16 × 3 = 334h`），各
48 bytes。`arg_0` 非零取前者、為零取後者。

`seg045` 的基底是 `17E30h`，所以兩張表的線性位址是 `18134h` 與 `18164h`。實際色值
本輪**沒有逐 byte 抄出來**（只確認資料範圍是 `0..0Fh`）。

## `17D72h`：填屬性平面

```asm
mov  ax, 0A200h           ; 屬性平面
mov  es, ax
mov  al, [bp+0Ch]
mov  cl, 5
shl  al, cl               ; 顏色放進 bit 7..5
or   al, 1                ; 一定會設的 bit 0
test byte ptr [bp+0Ah], 0FFh
jz   短
or   al, 4                ; 條件性的 bit 2
短: rep stosw
```

`0A200h` 是 PC-98 文字畫面的**屬性平面**（字元碼平面在 `0A000h`，見
[spec 636](636-pc98-text-vram-and-disk-bios.md)）。屬性 byte ＝
`(顏色 shl 5) or 1`，另一個參數非零時再 `or 4`。

`rep stosw` 依 `CX` 與 `DI` 填——兩者由呼叫端（`17CE9h`）先算好。

## 兩組完全重複的常式

| 位址 | 與誰位元組完全相同 | 內容 |
|---|---|---|
| `17DA9h` | **`17186h`** | Shift-JIS → JIS（[spec 646](646-sjis-to-jis-conversion.md)）|
| `17CE9h` | **`1692Dh`** | 文字 VRAM 定址（[spec 645](645-pc98-text-layer-primitives.md)）|

兩組各自逐位元組相同（`17CE9h` 與 `1692Dh` 都是 57 bytes）。差別只在它們**呼叫的
下游不同**：`1692Dh` 叫 `sub_16966`／`sub_169B6`，`17CE9h` 叫 `sub_17D22`／
`sub_17D72`。

也就是同一段顯示程式碼被編譯／連結了兩份，各配一組下游常式。remake 不必複製這個
重複，但**兩條路徑的下游行為要各自確認**，不能假設它們也一樣。

## `180F7h`：處理到沒有為止

```text
repeat
    if <sub_19293>() = 0 then break
    <sub_18036>()
forever
```

`sub_19293` 回 0 就結束，否則呼叫 `sub_18036` 再回頭。`sub_18036` 的回傳值存進
區域變數後**沒有再讀**——是死存。

## 明確不宣稱

- 兩張調色盤表的實際色值。
- 屬性 byte 的 `bit 0`／`bit 2` 對應什麼效果。
- `int 18h` 各功能碼與 `out 68h, 8` 的作用。
- `sub_19293`／`sub_18036`／`sub_17D22`／`sub_16966`／`sub_169B6` 的行為。
