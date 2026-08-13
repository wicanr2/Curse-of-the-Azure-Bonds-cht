# 第六百五十八輪：顏色位元互換表，與屬性的編解碼一對

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `19E98h`、`18FA3h`、`18FC5h`、`18EE0h`、`1917Ch`、
`191C2h`。

## `19E98h`：8 筆的位元互換表

```asm
mov  bx, 0FC0h
xlat byte ptr cs:[bx]
```

`seg050` 的基底是 `18EE0h`（該段第一支函式的位址），所以表在 **`19EA0h`**——
緊接在這支的 `retn` 之後。實際內容：

| 輸入 | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
|---|---|---|---|---|---|---|---|---|
| 輸出 | 0 | 1 | **4** | **5** | **2** | **3** | 6 | 7 |

寫成位元就清楚了：**bit 1 與 bit 2 互換，bit 0 不動**。

```text
in  000 001 010 011 100 101 110 111
out 000 001 100 101 010 011 110 111
```

PC-98 的顏色位元順序是 **G-R-B**，而一般習慣是 R-G-B——這張表就是那個互換。

而且它是**自反的**（`T[T[i]] = i`，已逐項驗算）。所以編碼與解碼可以共用同一張表，
不必準備兩份。

## `18FA3h` 與 `18FC5h`：互為反向

```text
18FA3h（打包 → PC-98 屬性）:
    colour := <19E98h>(AL and 7)
    other  := ((AL and 0F0h) shr 3) or 1
    return (colour shl 5) or other

18FC5h（PC-98 屬性 → 打包）:
    colour := <19E98h>((AL and 0E0h) shr 5)
    other  := (AL and 1Eh) shl 3
    return colour or other
```

`18FA3h` 把顏色放進 **bit 7..5**、其餘屬性放 bit 4..1，並且**一定設 bit 0**
（`or bl, 1`）。`18FC5h` 反過來取出，`and 1Eh` 只取 bit 4..1——**bit 0 在還原時被
丟掉**，所以往返一次不會完全還原：原本 bit 0 是什麼都會變成 1。

這與 [spec 647](647-pc98-palette-and-attribute.md) 的 `17D72h` 一致——那支也是
`(顏色 shl 5) or 1`，bit 0 恆為 1。

## `18EE0h`：兩組緩衝各做一次

```text
<sub_18F10>()
<sub_192D2>(DS:byte_2810Eh) ; <sub_1BA37>(DS:byte_2810Eh)
<sub_192D2>(DS:byte_2820Eh) ; <sub_1BA3C>(DS:byte_2820Eh)
```

兩塊緩衝相隔 `100h`（`2810Eh` 與 `2820Eh`），各自先過 `sub_192D2` 再交給**不同**的
routine（`1BA37h` 與 `1BA3Ch`，相差 5 bytes——多半是同一支的兩個入口）。

`byte_2820Eh` 也出現在 [spec 652](652-disk-swap-loop-and-third-leadbyte.md) 的換片
提示裡。

## `1917Ch` 與 `191C2h`

```text
1917Ch:  BH := byte_280C6h
         word_280E8h := word_280C8h        ← 存一份現值
         word_280EAh := word_280CAh
         <sub_19840>() ; <sub_19604>(word_280C8h)

191C2h:  <sub_195FC>()
         BH := byte_280C6h
         CL := low(word_280C8h) ; CH := DH ; DX := word_280CAh
         回傳 (CH <> DH) ? 1 : 0
```

`191C2h` 在 `mov ch, dh` 之後才 `mov dx, word_280CAh`，然後比較 `ch` 與 `dh`——
比的是**新舊兩個 `DH`**（一個來自 `sub_195FC` 的回傳、一個來自 `word_280CAh` 的
高位元組）。相等回 0、不等回 1。

## 明確不宣稱

- `sub_18F10`／`sub_192D2`／`sub_1BA37`／`sub_1BA3C`／`sub_19840`／`sub_19604`／
  `sub_195FC` 的行為。
- `word_280C8h`／`word_280CAh`／`byte_280C6h` 的用途。
- 屬性 bit 4..1 各自代表什麼。
