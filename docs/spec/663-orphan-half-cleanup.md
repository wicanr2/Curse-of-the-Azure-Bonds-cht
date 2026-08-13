# 第六百六十三輪：覆寫全形字時清掉落單的另一半

狀態：`READY`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `19B7Dh`、`19B9Ch`、`19BBCh`、`19C71h`、`19D02h`。

## 一對前後守衛

寫一個字元進某一格之前後各叫一次，處理「那一格原本是全形字的一半」的情況：

```text
19B7Dh（看前一格）:
    if low(word_280E8h) = 0 then return          ← 已在左界
    ax := es:[di−2]
    if ah = 0 then return                         ← 半形，不用管
    if ah 的 bit 7 已設 then return
    if al 的 bit 7 已設 then return               ← 前一格是右半，不是左半
    es:[di−2] := 0020h                            ← 前一格是左半 → 清成空白

19B9Ch（看目前這一格）:
    if low(word_280EAh) = 50h then return         ← 已在右界（80 欄）
    ax := es:[di]
    if ah = 0 then return
    if ah 或 al 的 bit 7 已設 then es:[di] := 0020h
```

全形字佔兩格、右半格的低位元組 bit 7 為 1（[spec 648](648-pc98-text-draw-core.md)）。
**寫進右半格會讓左半格落單，寫進左半格會讓右半格落單**——這兩支就是把落單的那一半
清成空白，不留下半個字。

中文化沿用同樣排版時這個清理照樣需要，而且與
[spec 662](662-cursor-snap-and-clear.md) 的游標吸附是同一組問題的兩面：吸附避免
**停在**字中間，清理處理**寫進**字中間。

## `19BBCh`：在游標處放一個空白

```text
word_280E8h := word_280EAh := word_280F9h        ← 兩個都設成目前游標
es := word_280D4h ; di := word_280D8h ; bx := word_280D6h
es:[bx+di] := <18FA3h>(byte_280C6h)              ← 屬性（打包 → PC-98 格式）
ax := 0020h
cld
<19B7Dh>()                                        ← 先處理前一格
stosw                                             ← 寫入空白
<19B9Ch>()                                        ← 再處理後一格
```

順序是「先清前、再寫、後清後」。屬性經 `18FA3h`
（[spec 658](658-color-bit-swap-and-attribute-codec.md)）換算，所以 `byte_280C6h`
存的是**打包格式**而不是 PC-98 原生屬性。

## `19C71h`：讀一格，全形讀兩個

```text
byte_280E3h := 0
al := es:[bx+di]
if byte_280D0h <> 0 and <sub_1977Eh>(al) 判定為前導 then
    <sub_19493>(al) ; bx := bx + 1               ← 先送第一個 byte
al := es:[bx+di]
<sub_19493>(al) ; bx := bx + 1                    ← 再送（第二個或唯一一個）
```

`byte_280D0h` 又是那個雙位元組總開關（[spec 662](662-cursor-snap-and-clear.md) 的
游標吸附也看它）。關掉時**一律只送一個 byte**。

`sub_1977Eh` 用進位回答（`jnb` ＝ 不是前導就跳過），與
[spec 574](574-pc98-shiftjis-and-text-vram.md) 的 `169D9h` 同一種慣例。

## `19D02h`：游標形狀

```text
byte_280FFh := al
if al 的 bit 0 = 0 then
    <sub_1977Bh>(AH = 12h)
else
    <sub_1977Bh>(AH = 10h, AL = al shr 1)
    <sub_1977Bh>(AH = 11h)
```

`bit 0` 決定走哪條路：偶數送一個命令，奇數送兩個（其中第一個帶 `al shr 1` 當參數）。
`11h` 與 [spec 661](661-cursor-and-bios-work-area.md) 的 `19A10h` 用的是同一個
功能碼。

## 明確不宣稱

- `sub_1977Bh`／`sub_1977Eh`／`sub_19493` 的行為與功能碼 `10h`／`11h`／`12h` 的意義。
- `byte_280E3h`／`byte_280FFh`／`word_280E8h`／`word_280EAh` 的完整語意。
