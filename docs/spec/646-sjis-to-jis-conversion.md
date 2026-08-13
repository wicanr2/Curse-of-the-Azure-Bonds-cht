# 第六百四十六輪：Shift-JIS → JIS 轉換，與字表查詢

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `17186h`、`17148h`、`171DFh`。

## `17186h`：Shift-JIS → JIS

輸入 `AX` ＝ Shift-JIS 碼（`AH` 首位元組、`AL` 次位元組），輸出同樣在 `AX`：

```asm
cmp  ah, 9Fh
jbe  短
sub  ah, 40h          ; 第二段（E0h..FCh）先拉回第一段的連續空間
短: sub  ah, 71h
add  ah, ah           ; ×2
inc  ah
cmp  al, 7Fh
jbe  短2
dec  al               ; 跳過 7Fh 這個洞
短2: cmp  al, 9Eh
jb   短3
sub  al, 7Dh
inc  ah               ; 後半區 → 區號加一
retn
短3: sub  al, 1Fh
retn
```

這是標準的 Shift-JIS → JIS（区点）換算。用三個已知碼點驗算：

| Shift-JIS | 字 | 算出的 JIS | 應為 |
|---|---|---|---|
| `8140h` | 全形空白 | `2121h` | `2121h` ✓ |
| `889Fh` | 亜 | `3021h` | `3021h` ✓ |
| `82A0h` | あ | `2422h` | `2422h` ✓ |

三個都對上，所以這支就是 SJIS→JIS，不是別的位元運算。

### 對中文化的意含

JIS 碼是 **PC-98 漢字 ROM 的索引**：換算出來的区点直接餵給字型讀取硬體。所以這支
與 [spec 645](645-pc98-text-layer-primitives.md) 的 `14342h`（首位元組判斷）合起來
就是整條日文顯示路徑的入口。

中文化**不能只改這兩支的常數**——漢字 ROM 裡沒有繁體字集，換算得再準也讀不到字。
必須把「碼點 → 字形」整條路徑換成自備字型：`14342h` 改判 Big5 首位元組，
`17186h` 換成「Big5 → 自備字庫索引」，而讀字型的那一段（`171DFh` 之後那條鏈）
改成從記憶體取字模。

## `17148h`：字表線性查詢

```text
n  := es:[bx]          ← 第一個 byte 是筆數
bx := bx + 1
i  := n
while i <> 0 do
    if es:[bx] = dx then return n − i + 1     ← 1 起算的序號
    bx := bx + 2
    i  := i − 1
return 0                                       ← 找不到
```

表格格式是「**一個 byte 的筆數 ＋ 連續的 word**」。回傳值 **1 起算**，`0` 專門
表示找不到——所以表最多 255 筆。

比較是 16-bit 整字比（`cmp ax, dx`），不是逐 byte。

## `171DFh`：把參數搬進全域再呼叫

```text
DS := seg042
cld
dword_16A19 := arg_0（far pointer）
byte_16C25  := arg_4
byte_16C24  := arg_6
word_16A1D  := arg_8 的 offset
word_16A1F  := arg_8 的 segment
word_16A21  := arg_C
<sub_16E70>()
```

七個參數全部搬進固定的全域位址，被呼叫端不吃參數。這是組語常式與 Pascal 之間的
常見介面寫法——`sub_16E70` 從全域讀，不從堆疊讀。

`byte_16C24` 與 `byte_16C25` **順序是反的**（`arg_4` 進 `16C25h`、`arg_6` 進
`16C24h`），照抄時要注意。

## 明確不宣稱

- `sub_16E70` 做什麼、那些全域各自的用途。
- `17148h` 查的是哪一張表（呼叫端傳 `ES:BX`）。
