# 第六百七十一輪：DOS 側的顯示頁切換、40×25 範圍檢查，與一組死掉的下限判斷

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `START.EXE` 的 `15493h`、`15731h`、`134F4h`。

## `15493h`：切換 BIOS 顯示頁

```text
byte_2118Bh := 5                          ← Registers.AH
byte_2118Ah := arg_0                       ← Registers.AL
<@INTR$q4BYTEm9REGISTERS>(10h, @byte_2118Ah)
word_2119Eh := (arg_0 shl 9) + 0A000h
```

呼叫的是 Turbo Pascal 的 `Intr(IntNo, Regs)`（IDA 由 Borland 簽章還原出名稱）。
`Registers` 記錄的 `AX` 在偏移 0，所以 `byte_2118Ah` 是 `AL`、`byte_2118Bh` 是
`AH` ——`AH = 5` 是 BIOS 的「選擇顯示頁」。

`shl 9` ＝ ×512 **段落** ＝ 8 KB，所以 `word_2119Eh` ＝ `A000h + 頁碼 × 8 KB`
的段位址。這支同時做兩件事：叫 BIOS 換頁，**並自己算好那一頁的段位址存起來**。

## `15731h`：40×25 的範圍檢查

```asm
cmp  [bp+arg_6], 0
jb   短：不做          ← ★ 無號比較，永遠不成立
cmp  [bp+arg_6], 27h
ja   短：不做
cmp  [bp+arg_4], 0
jb   短：不做          ← ★ 同上
cmp  [bp+arg_4], 18h
jbe  短：做
```

上限是 `27h`（39）與 `18h`（24）——**40 欄 × 25 列**。

兩個下限檢查是**死碼**：`jb`（低於）用的是無號比較，而任何無號值都不小於 0。
原始碼多半寫的是 `if (x >= 0) and (x <= 39)`，編譯器沒有把恆真的那半消掉。

**這不是判讀不確定，是原作就有的冗餘**——remake 直接寫上限即可，行為完全相同。

通過後呼叫 `<sub_15606>(1, 20h, arg_0, arg_2, arg_2, arg_4, arg_6)`：`arg_2`
**被推入兩次**。

## `134F4h`：交換兩個 nibble

```text
var := (arg_0 shl 4)                       ← 只保留低 byte
var := var + ((arg_0 and 0F0h) shr 4)
return var
```

把一個 byte 的高低 4 bit 對調。用 `add` 而不是 `or`——兩半不重疊所以結果相同。

`shl` 之後存進 byte 變數，高位自然被截掉，不需要額外遮罩。

## 明確不宣稱

- `sub_15606` 的行為，以及為何 `arg_2` 要推兩次。
- `word_2119Eh` 算出的段位址由誰使用。
- 40×25 是整個畫面還是某個視窗。
