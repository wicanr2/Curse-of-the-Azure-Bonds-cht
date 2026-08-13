# 第六百六十輪：第四種 SJIS→JIS —— 拆法不同、結果相同、多了守衛

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `19791h`、`195B6h`、`197BDh`。

## `19791h` 與 `17186h` 是等價的

[spec 646](646-sjis-to-jis-conversion.md) 記過三份**逐位元組完全相同**的
SJIS→JIS（`17186h`／`17DA9h`／`169EDh`）。`19791h` 是第四份，但**位元組不同**：

```asm
cmp  ah, 80h
jz   短：原樣返回              ← 守衛一
cmp  ah, 0A0h
jnb  短2
sub  ah, 70h                   ; 81h..9Fh
jmp  短3
短2: cmp  ah, 0F0h
jnb  短：原樣返回              ← 守衛二
sub  ah, 0B0h                  ; A0h..EFh
短3: or   al, al
jns  短4
dec  al                        ; al >= 80h
短4: add  ah, ah                ; ×2（沒有 +1）
cmp  al, 9Eh
jb   短5
sub  al, 5Eh
jmp  短6
短5: dec  ah
短6: sub  al, 1Fh               ← 兩條路徑都會走到
```

關鍵在 `短6`：`sub al, 5Eh` 之後**接著**執行 `sub al, 1Fh`，合計 `−7Dh`——與
`17186h` 的 `sub al, 7Dh` 相同。而區號那邊，`17186h` 算 `(ah − 71h) × 2 + 1`、
這支算 `(ah − 70h) × 2` 再視情況 `− 1`，兩者恆等（`(X+1)×2 − 1 = X×2 + 1`）。

**逐碼點驗算**：把兩支各自實作一遍，跑遍 `81h..9Fh` ＋ `E0h..FCh` 的所有合法
Shift-JIS 碼點（高位元組 × 低位元組），**不一致的有 0 筆**。

## 差別只在守衛

`19791h` 兩處會**原樣返回不做轉換**：

| 高位元組 | `17186h` | `19791h` |
|---|---|---|
| `80h` | 照算（得到垃圾） | **不轉換** |
| `F0h..FCh` | 照算 | **不轉換** |

`F0h..FCh` 正是 [spec 648](648-pc98-text-draw-core.md) 提到的兩族首位元組判準差異
所在（`FCh` 對 `F7h`）。這支的守衛切在 `F0h`，**又是第三個界線**。

三個界線並存：`FCh`（`14342h`／`177BDh`）、`F7h`（`169D9h`）、`F0h`（本支）。
remake 要合併時得先決定照哪一個，而不是假設它們一致。

## `197BDh`：畫之前先做有號範圍檢查

```text
if dl < cl then return          ← 有號 jl
if dh < ch then return
if al <> 0 then
    word_280E8h := cx ; word_280EAh := dx
    <sub_19825>() ; <sub_198AE>()
    ch := dh − al + 1
<sub_19840>()
```

兩個比較都是 **`jl`（有號）**，所以座標是有號量，負值會被擋掉而不是當成大正數。
`(cl, ch)` 與 `(dl, dh)` 是矩形的兩個角。

`al = 0` 時跳過中間那段，直接走 `sub_19840`。

## `195B6h`：捲動時的邊界調整

```text
<sub_19A69>()
al := dh
<sub_19A76>()
if dh <> al then
    if byte_280DBh = high(word_280CAh) + 1 then
        dec dh ; high(word_280CAh) := dh
    byte_280DBh := dh + 1
inc dh
if dh > high(word_280CAh) then
    dec dh
    <sub_197BD>(1, byte_280C6h, word_280C8h, word_280CAh)
```

`byte_280DBh` 與 `word_280CAh` 的高位元組是一對「目前列」與「界線」，兩者相差 1
時才調整。最後那個比較是**無號**（`jbe`），與 `197BDh` 的有號比較不同——同一組座標
在兩支裡用不同的有號性。

## 明確不宣稱

- `sub_19A69`／`sub_19A76`／`sub_19825`／`sub_198AE`／`sub_19840` 的行為。
- `byte_280DBh`／`word_280C8h`／`word_280CAh` 的確切語意。
- `19791h` 的守衛是刻意排除造字區，還是與其他兩族各自獨立演化的結果。
