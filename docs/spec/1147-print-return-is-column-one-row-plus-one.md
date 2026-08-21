# 1147 — `33h PRINT RETURN` 是「欄回 1、列 ＋1」，兩個分支做的事一樣

- 證據等級：`exact`（DOS `overlay-02:2CEAh`..`2D14h` 整支 14 條讀完）
- handler 位址取自 `docs/audit/ecl-opcode-handlers-dos.md`

## 全函式

```asm
2CED  inc   word ptr ds:4FB4h      ; PC 推進（每個 handler 都做，spec 1104）
2CF1  cmp   byte ptr ds:8B61h, 0
2CF6  jz    short loc_2D08
2CF8  mov   byte ptr ds:65A0h, 1   ; 欄 := 1
2CFD  inc   byte ptr ds:65A1h      ; 列 ＋1
2D01  mov   byte ptr ds:8B61h, 0   ; 清旗標
2D06  jmp   short loc_2D11
2D08  mov   byte ptr ds:65A0h, 1   ; ★ 跟上面一模一樣
2D0D  inc   byte ptr ds:65A1h
```

★ **兩個分支對游標做的事完全一樣。** `8B61h` 只決定要不要順手把它清成 0
（本來就是 0 的話不用清）。所以整條指令的可觀察效果是：

```pascal
DS:65A0h := 1;   { 欄回到 1 }
inc DS:65A1h;    { 列 ＋1 }
```

⇒ 它是**硬換行**，不是「把目前這一行結束掉」：連續兩條 `33h` 會多推一列，
也就是空一行。

## `65A0h` 是欄、`65A1h` 是列

spec 807 當時留了「沒有宣稱 `DS:65A0h` / `DS:65A1h` 是欄或列」。本輪關掉，
兩個獨立證人：

| 出處 | 用法 | 推得 |
|---|---|---|
| spec 826 | `DS:65A0h := DS:6506h^[0] ＋ 2`——拿**名字長度**算出來 | `65A0h` 是**欄** |
| spec 840 | 中段把行號重設成 `DS:65A1h ＋ 1`，逐行 `inc(行)` | `65A1h` 是**列** |

本規格的 `65A0h := 1` ／ `inc 65A1h` 與兩者一致：換行就是欄歸位、列前進。

## remake

`33h` 目前只把指令邊界記成 `PrintReturnCount`，**沒有游標模型**——版面由 UI
那一側決定。缺口因此不是「解讀」而是「表現層」：

⚠ **連續的 `33h` 應該產生空行**。remake 的文字管線目前不把它算成一列，所以
原作留白的地方會被擠掉。要補的是 UI 的行模型，不是 ECL VM。

## 明確不宣稱

- 沒有宣稱 `DS:8B61h` 是什麼旗標。本支只證明「它非 0 時會被清成 0」，
  而且**它不影響游標怎麼動**——兩個分支一樣。
- 沒有宣稱 `65A0h`／`65A1h` 的單位（字元格或像素），只宣稱哪個是欄、哪個是列。
