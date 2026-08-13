# 第六百三十八輪：`overlay-12` 效果 callback 七支 —— 一對鏡得不完全的函式

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-12` 的 `01E6h`、`04E2h`、`1030h`、`11ACh`、`1222h`、
`126Dh`、`12A0h`。

## `126Dh` 與 `12A0h`：兩個 bit 對調，但守門的變數不同

```text
126Dh:  if DS:0A02Fh and 2 <> 0 then
            DS:0A02Ch := DS:0A02Ch + 2
        else if DS:0A02Fh and 1 <> 0 and DS:0A042h = 0 then
            DS:0A02Eh := byte(DS:0A02Eh × 2)

12A0h:  if DS:0A02Fh and 1 <> 0 then
            DS:0A02Ch := DS:0A02Ch + 2
        else if DS:0A02Fh and 2 <> 0 and DS:8CF7h = 0 then
            DS:0A02Eh := byte(DS:0A02Eh × 2)
```

兩支的結構與常數完全對稱——**除了守門的那個變數**：`126Dh` 看 `DS:0A042h`、
`12A0h` 看 `DS:8CF7h`。兩個位址差很遠，不是同一塊資料的相鄰欄位。

這種「幾乎鏡像但有一處不同」的寫法，通常代表兩件事之一：原始碼是複製貼上後改一
半，或那兩個變數本來就分屬不同子系統。**本輪不判斷是哪一種**，但 remake 照抄時
不能把它們合併成一支帶參數的函式。

`DS:0A02Eh × 2` 用 `shl ax, 1` 算完只存回 `al`——**乘 2 溢位時高位元靜默丟掉**。

## `01E6h`：兩個欄位各自夾在上限 60

```text
if arg_6^[19Bh] < 3Ah then arg_6^[19Bh] += 2 else arg_6^[19Bh] := 3Ch
if arg_6^[19Ch] < 3Ah then arg_6^[19Ch] += 2 else arg_6^[19Ch] := 3Ch
```

門檻是 `3Ah`（58）、上限是 `3Ch`（60）。`58` 與 `59` 都會被拉到 `60`，所以
**`59` 這個值加完會變成 `60` 而不是 `61`**，而且原本就 `≥ 58` 的一律直接設 60。
兩個欄位各做一次，寫法相同。

## 其餘四支

```text
04E2h:  if arg_6^[18Eh]^[2] <> 0 then
            備妥 'は沈黙させられた。'，<far 014Ah:0024h>(arg_6, 訊息, 1, 0Ah)
        arg_6^[18Eh]^[2] := 0
        arg_6^[18Eh]^[1] := 0            ← 兩個欄位都清，訊息只在 [2] 非 0 時出現

1030h:  if <sub_FCC>(DS:9594h) <> nil and <sub_FCC>(DS:9594h)^[5Ah] = 0 then
            <sub_F5B>(arg_6, arg_8, 64h)
        ← sub_FCC 用同樣的參數呼叫了兩次，結果沒有暫存

11ACh:  DS:0A03Dh := arg_6^[18Eh]^[0Ah]
        if DS:9594h^[11Ah] in {2, 0Ah} and (DS:9594h^[0DEh] and 7Fh) = 2 then
            DS:0A039h := DS:0A039h − 4

1222h:  v := (arg_2^[3] and 10h) div 10h          ← 取 bit 4，得到 0 或 1
        if arg_6^[198h] = v then
            <sub_8D>(arg_0, arg_2, arg_6)
        else
            dec DS:0A039h
            dec DS:0A02Ch
```

`11ACh` 的 `+11Ah` 是 `RACETYPE`（[spec 499](499-pc98-alignment-conditional-effects.md)），
兩個值 `2` 與 `0Ah` 才進入下一層檢查。`+0DEh` 先 `and 7Fh` 再比——**最高位是另一
個用途的旗標，比較前要先遮掉**。

`1222h` 取 bit 4 的寫法是 `and 10h` ＋ `cwd` ＋ `idiv 16`，編譯器把
`(x and 16) div 16` 直譯成有號除法。結果只會是 `0` 或 `1`。

## 明確不宣稱

- `DS:0A02Ch`／`DS:0A02Eh`／`DS:0A02Fh`／`DS:0A039h`／`DS:0A042h`／`DS:8CF7h`
  各自的身分。
- `+19Bh`／`+19Ch`／`+198h`／`+5Ah`／`+0DEh` 欄位的意義。
- `sub_FCC`／`sub_F5B`／`sub_8D`／`014Ah:0024h` 的行為。
