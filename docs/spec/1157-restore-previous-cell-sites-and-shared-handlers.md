# 1157 — 那 23 處逐處分類：全部是真實移動；以及兩處被 IDA 切錯的共用 handler

- 證據等級：`exact`（23 處逐處列出守衛；`0DA4h` 與 `0DF3h` 逐條讀出）
  ＋ 一次已還原的實跑探針
- 分歧的量測見 [spec 1155](1155-redraw-call-coordinate-divergence.md)
- `7EC7h` 的語意見 [spec 1156](1156-post-combat-outcome-cell-and-execution-order.md)

## 分類：兩類，都是真實移動

[spec 1155](1155-redraw-call-coordinate-divergence.md) 量到 23 處
「只寫 `C04B`／`C04C`、remake 因此跳過」。逐處讀完之後只有兩種：

| 類 | 處數 | 形狀 |
|---|---:|---|
| 明確移動 | 8 | `SAVE <常數> C04B/C04C`（傳送）或 `ADD`／`SUBTRACT 1`（走一步）|
| 退回上一格 | 15 | `SAVE 4BF0 C04B; SAVE 4BF1 C04C` |

**兩類都是真實的隊伍移動**——後者的來源是「這一步移動之前的座標」
（[spec 1155](1155-redraw-call-coordinate-divergence.md)），退回去也是移動。
⇒ `C04D` 條件擋掉的是**玩家看得見的位移**，不是暫存值。

## 15 處退回的守衛各不相同

先前以為它們都掛在戰後結果碼上。逐處讀完之後不是：

| 位置 | 守衛 | 情境 |
|---|---|---|
| `ECL2/0x01:1444h` | `COMPARE 4C10 03` | 「THEY SEND YOU BACK. YOU MOVE AWAY.」|
| `ECL2/0x02:00C2h` | `COMPARE 7ED5 00` | 接在 `CALL C01Eh`（`MOVEFORWARD`）與 `NEWECL 03` 之後 |
| `ECL2/0x02:0CF1h` | （窗口內無）| 「THEY GROWL LOUDLY AS YOU ESCAPE.」|
| `ECL2/0x03:1B6Ch` | `COMPARE 4C06 01` | — |
| `ECL2/0x04:160Ch` | `COMPARE 7EC7 80h` | 戰後沒打贏 |
| `ECL3/0x10:0177h` | `COMPARE 4C44 FFh` | — |
| `ECL3/0x11:0589h` | `COMPARE 7F79 06` | — |
| `ECL4/0x20:1237h` | （窗口內無）| 「THE CITY HAS GONE SILENT.」|
| `ECL4/0x22:11F0h` | （窗口內無）| 氣孢子那一段 |
| `ECL4/0x23:0043h` | `COMPARE 4CE4 FFh` | 接 `NEWECL 20` |
| `ECL4/0x25:151Bh` | 接在 `COMBAT` 之後 | — |
| `ECL5/0x31:0E9Fh`、`ECL5/0x32:176Ch`、`ECL6/0x42:17ADh` | `COMPARE 7F79 01` | — |
| `ECL5/0x33:1BD6h`、`ECL6/0x40:097Fh` | （窗口內無）| — |
| `ECL6/0x43:14B8h` | `COMPARE 7F82 01` | — |

⇒ 只有一處用 `7EC7h`。**「等 `7EC7h` 有 producer」不是這 23 處的共同前提。**

## ★★ `160Ch` 是從六個地方跳進去的共用處理

在 `ECL2/0x04` 的位元組裡搜目的地 `960Ch`（＝偏移 `160Ch`），命中**六處**：
`02D0h`、`0ABFh`、`0AD9h`、`0CFCh`、`0E6Eh`、`15FDh`。

⇒ `160Ch` 是「把隊伍退回上一格」的**共用出口**，`15FDh` 那條戰後跳轉只是其中之一。
先前寫「那一處只有戰後才走得到」是只看了一個入口。

## ★★★ 火刃巢穴那三條路線測試為什麼會紅

拿掉 `C04D` 條件之後，`TestRealNewGame*` 三條在第 4 步失敗。實跑探針
（已還原）指出投影發生在 `ECL2/0x04:161Ah`——也就是 `160Ch` 那個退格處理
之後的 `CALL`，把 `(5,2)` 蓋回隊伍的 `(4,2)`。

同一次探針也證實 `COMPARE 7EC7 80h` 讀到的是 `left=0 right=80h`（位址解得對、
值是 0），所以走進去的**不是** `15FDh` 那條戰後跳轉，而是另外五個入口之一。

⇒ **那三條路線是照著「被 `C04D` 條件壓掉之後」的行為寫出來的**：路線註解說
「intentionally crosses the blade barrier and frozen-room cells」，而原作在那些
格子上會把隊伍推回去。要換成忠實模型，得連同重新推導那三條路線一起做。

## ★★ 兩處被 IDA 切錯的共用 handler

| opcode | audit 表原本 | 真正的範圍 | 差 |
|---|---|---|---|
| `21h`／`37h` | `0C15h`..`0D4Ah`（104）| `0C15h`..`0DA3h`（131）| spec 1153 |
| `2Fh`／`30h` | `0DA4h`..`0DBFh`（11）| `0DA4h`..`0E12h`（43）| 本規格 |

`2Fh`／`30h` 也是**共用一支**，在 `0DD8h` 重讀 `DS:75FFh` 分辨自己被誰呼叫
——與 `21h`／`37h` 同一個形狀（[spec 587](587-ecl-handler-21-37-shared.md)）。
audit 表原本連 `2Fh` 那一列都沒有，兩列都已補上，並在表頭寫了邊界警告。

## ★★ `AND`／`OR` 之後那六個比較格子的方向

原作在寫回目的地之前呼叫 `compare_variables(0, 結果)`——**`0` 是左運算元**：

```text
0DF3  xor ax, ax
0DF5  push ax          ← 左 ＝ 0
0DF6  mov al, [bp-7]   ← 結果
0DFB  push ax          ← 右 ＝ 結果
0DFC  call compare_variables
```

對照 `03h COMPARE`（`overlay-02:011Eh`）的收尾是 `push op1` 再 `push op2`，
所以 `AND`／`OR` 那一次的左右**確實是反過來**的：排序格子是 `0 op 結果`。

remake 的程式碼本來就是這樣寫的；錯的是它上面那行註解（寫成
`compare_variables(result, 0)`）。註解已改正。

⚠ **照那行註解「修正」程式碼會壞掉，而且測不出來**：全 corpus 174 處
`AND`／`OR` 後面**沒有一處**接排序型 `IF`（125 處是 `IF <>`，其餘不接 `IF`），
所以把四個排序格子對調之後所有測試照樣全綠。只有讀 `0DF3h` 那三條指令
才分得出方向。

## 明確不宣稱

- 沒有宣稱那 15 處守衛（`4C10`／`7ED5`／`4C06`／`4C44`／`7F79`／`4CE4`／`7F82`）各是什麼。
- 沒有宣稱火刃巢穴那三條路線在忠實模型下**正確**的走法是什麼。
- 沒有宣稱 `160Ch` 那六個入口各自的觸發條件。
- 沒有宣稱 audit 表其餘各列的邊界都是對的——只驗過這兩處。
