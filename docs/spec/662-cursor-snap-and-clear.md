# 第六百六十二輪：全形字的游標吸附、畫面清除，與 VRAM 位移計算

狀態：`READY`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `19B00h`、`19ACEh`、`19B4Fh`、`19AA2h`。

## `19B00h`：游標落在右半格就左移

```text
if byte_280D0h = 0 then return                  ← 功能關閉
if dl = low(word_280C8h) then return            ← 已在左界
es := word_280D4h
bx := (dh × byte_280DAh + dl) × 2               ← 不是 80 欄時再 × 2
ax := es:[bx]                                    ← 目前這一格
if ah = 0 then return                            ← 高位元組 0 ＝ 半形
if ah 或 al 的 bit 7 沒設 then return
（是全形的某一半）
ax := es:[bx−2]                                  ← 看前一格
if ah = 0 或 ah 的 bit 7 已設 或 al 的 bit 7 已設 then return
dec dl                                           ← 游標左移一格
```

全形字佔兩格，右半格的低位元組 bit 7 是 1（[spec 648](648-pc98-text-draw-core.md)）。
這支檢查「目前格是全形的一半、而前一格是它的左半」，成立就把游標往左移一格——
**避免游標停在一個字的中間**。

中文化沿用同樣的排版（一個字兩格）時，這個吸附邏輯照樣需要；但**判斷條件綁在
bit 7 的編碼上**，若中文版改用別的存法就要跟著改。

`byte_280D0h` 是這個功能的開關。

## `19ACEh`：清成空白 ＋ 白色

```text
cx := (low(word_280CAh) + 1) × byte_280E5h       ← 格數
es := word_280D4h ; bx := word_280D6h ; di := 0
重複 cx 次:
    es:[bx+di] := 0E1h                            ← 屬性
    ES:DI := 0020h ; di := di + 2                 ← 字元（stosw）
```

`0E1h` ＝ `(7 shl 5) or 1`——顏色 7、bit 0 設起，與
[spec 658](658-color-bit-swap-and-attribute-codec.md) 的編碼一致。字元填
`0020h`（半形空白的 word 形式）。

屬性寫在 `bx + di`、字元寫在 `di`，所以 **`word_280D6h` 是兩個平面的位移差**。

`byte_280E5h` 是 `19AA2h` 從 BIOS 工作區 `0000:0712h` 抄下來的列數上限
（[spec 661](661-cursor-and-bios-work-area.md)）。

## `19B4Fh`：算位移並送出

```text
dx := (dh × byte_280DAh + dl) × 2                ← 不是 80 欄時再 × 2
word_280D8h := dx
if word_280EDh <> 0 then dx := dx + 1000h        ← 換一塊
ah := 13h ; <sub_1977B>()
```

位移算式與 `19B00h`、[spec 659](659-console-raw-mode-and-textrec.md) 的 `192D2h`
三處相同。`+1000h` 是條件性的區塊切換（`1000h` ＝ 4096 bytes ＝ 2048 格）。

`word_280D8h` **先存未加 `1000h` 的值**，加了之後只用於這次呼叫。

## `19AA2h`：初始化游標狀態

```text
<19A10h>()                                        ← 開游標
<19A8Bh>()                                        ← 從 BIOS 讀回位置
word_280FBh := word_280FDh := high(word_280F9h) shl 8
byte_280E5h := 0000:0712h                         ← 抄一份列數上限
```

`word_280FBh` 與 `word_280FDh` 都設成「目前列 ＋ 欄 0」——兩個變數初值相同。

## 明確不宣稱

- `word_280D4h`／`word_280D6h`／`word_280EDh`／`byte_280D0h`／`word_280C8h` 的完整語意。
- `sub_1977B` 與 `ah = 13h` 的作用。
- `word_280FBh`／`word_280FDh` 分別代表什麼。
