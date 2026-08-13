# 第六百九十五輪：物品鏈的節點、配置與釋放

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-24` 的 `16E6h`、`15FFh`。

## 三支 RTL 的身分

兩支函式對稱地用到同一組 resident 程序，參數形狀直接對上 Turbo Pascal 的宣告：

| 位址 | 呼叫形狀 | 身分 |
|---|---|---|
| `0A54:0329h` | `(@指標變數, 3Fh)` | `GetMem(var P; Size)` |
| `0A54:0364h` | `(@指標變數, 3Fh)` | `FreeMem(var P; Size)` |
| `0A54:1ABDh` | `(來源, 目的, 3Fh)` | `Move(Src, Dst, Count)` |

兩支都傳**指標變數的位址**（`lea` 而非 `les` ＋ `push es`），這是 `var` 參數的
標記，也是分辨 `GetMem`／`FreeMem` 與一般函式的依據。

## 節點與鏈

```text
節點大小 3Fh ＝ 63 bytes
+2Ah      far 指標，指向下一個節點（兩個 word）
鏈頭      角色紀錄的 +14Dh（far 指標）
```

`+2Ah` 之外的 62 個 byte 本輪沒有拆。

## `16E6h(角色, 來源)`：複製一份掛到鏈尾

```text
GetMem(node, 3Fh)
Move(來源^, node^, 3Fh)
node^[2Ah] := NIL                       ← 複製過來的 next 一定要蓋掉
if 角色^[14Dh] = NIL then
    角色^[14Dh] := node
else
    q := 角色^[14Dh]
    while q^[2Ah] <> NIL do q := q^[2Ah]
    q^[2Ah] := node
retf 8
```

掛的是**複本**不是來源本身，所以呼叫端的來源可以是堆疊上的暫存。
接尾是每次都從頭線性走到底，`n` 個物品就是 `O(n)`。

`GetMem` 的結果**沒有檢查**。

## `15FFh(角色, 物品)`：從鏈上摘掉並釋放

```text
if 物品 = NIL then
    顯示 'Nil Item pointer...'；返回
q := 角色^[14Dh]
if q = 物品 then
    角色^[14Dh] := 物品^[2Ah]
else
    while q <> NIL 且 q^[2Ah] <> 物品 do q := q^[2Ah]
    if q = NIL then
        顯示 "Tried to Lose item & couldn't find it!"；返回
    q^[2Ah] := 物品^[2Ah]
物品^[2Ah] := NIL
FreeMem(物品, 3Fh)
retf 8
```

兩個錯誤分支都只是**顯示訊息然後正常返回**，不是中止；呼叫端拿不到失敗訊號。

## 兩則開發用訊息留在零售版裡

| 位址 | 內容 |
|---|---|
| `CS:15C4h` | `Nil Item pointer...` |
| `CS:15D8h` | `Tried to Lose item & couldn't find it!` |

寫法（省略號、`&`、驚嘆號、直接講指標）都不是給玩家看的。中文化時這兩則要不要
翻是個決定：翻了會讓玩家以為是遊戲內容，不翻則保留原樣比較誠實。

## 參數位置的判讀

兩支都是 `retf 8`，也就是兩個 far 指標。實際位移是 `[bp+6]` 與 `[bp+0Ah]`
（IDA 的 `arg_0`／`arg_2`／`arg_4` 標籤在這裡對不上，要看 `8b46xx`／`c47exx`
的位移位元組）。Pascal 先推的參數在高位址，所以 `[bp+0Ah]` 是第一個參數
（角色）、`[bp+6]` 是第二個（物品）。

## 明確不宣稱

- 節點 63 bytes 裡除了 `+2Ah` 之外的內容。
- `+14Dh` 是不是唯一一條物品鏈（角色身上可能還有別的鏈）。
- `0542:0946h`（顯示錯誤訊息那支）的行為。
