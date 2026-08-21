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

⚠ 「明確移動」實際是 **12 處**：這裡的 23 處是只認 `09h SAVE` 當寫入路徑時
量出來的，四處 `GETTABLE` 的假門傳送被漏到「沒有座標寫入」那一類。逐處清單見
[spec 1159](1159-storevalue-is-the-only-write-path.md)。

**兩類都是真實的隊伍移動**——後者的來源是「這一步移動之前的座標」
（[spec 1155](1155-redraw-call-coordinate-divergence.md)），退回去也是移動。
⇒ `C04D` 條件擋掉的是**玩家看得見的位移**，不是暫存值——條件已於
[spec 1158](1158-hap-village-extent-and-refused-edges.md) 拿掉。

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

## ★★★ `160Ch` 是「退回上一格」的共用出口，六個入口有一半是選單

`cmd/ecl-window -into 0x160C` 列出六處，兩類各一半：

**一、戰後跳轉**（三處）：

| 位置 | 守衛 |
|---|---|
| `02CCh` | `COMPARE 7EC7 80h` ＋ `IF >` |
| `0AD5h` | `COMBAT` ⇒ `COMPARE 7EC7 80h` ＋ **`IF =`** |
| `15F9h` | `COMBAT` ⇒ `COMPARE 7EC7 80h` ＋ `IF >` |

**二、`ON GOTO` 的選單目的地**（三處，玩家自己選的）：

| 分派 | 選單 | 選項 → 目的地 |
|---|---|---|
| `0AB3h` | 刺客（`0AA0h`）| `ATTACK`→`0ABFh`、**`LEAVE`→`160Ch`** |
| `0CEDh` | 刀刃屏障（`0CCCh`）| `ENTER THE BLADES`→`0CFCh`、`WAIT`→`0D21h`、**`RETREAT`→`160Ch`** |
| `0E65h` | 冰凍房間（`0E47h`）| **`RETREAT`→`160Ch`**、`INTERROGATE`→`0E74h`、`KILL`→`0EE3h` |

⇒ **`160Ch` 主要是「玩家選擇退出」的共用出口**，戰後跳轉只是其中一種。
選單選項的文字用 `-raw` 讀出來：`80` 之後那個位元組是**位元組數**不是字數，
6 位元組 ＝ 8 個字元（`DecodePackedText` 六位元一字）。

## ★★★ 火刃巢穴那三條路線測試為什麼曾經是綠的

實跑追蹤（探針已還原）給出完整路徑：

```text
0E47  HORIZONTAL MENU     ← 「…A NUMBER OF PEOPLE FROZEN IN POSITIONS OF
                             BATTLE… WHAT DO YOU DO?」
0E65  ON GOTO             ← 玩家選的那一項
160C  SAVE 4BF0 C04B      ← 退回上一格
1613  SAVE 4BF1 C04C
161A  CALL 2E10
161E  EXIT
```

冰凍房間的第 0 個選項是 `RETREAT`，而那三條路線的 `continueHideoutEvent`
一律送 `Select(0)`。原作會把隊伍推回上一格；當時的 `C04D` 條件把那個位移壓掉，
路線才「剛好」走得通。路線註解自己寫著「intentionally crosses the blade
barrier and frozen-room cells」——意圖是穿過去，所以要明確挑
`INTERROGATE`，不是把座標期望值改成失敗值。

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
- 沒有宣稱 `160Ch` 除了 `-into` 列出的六處以外還有沒有別的入口——位元組掃描
  另有三處 `0C 96` 命中（`02D0h`／`0ABFh`／`0AD9h`）沒有逐處確認是指令還是資料。
- 沒有宣稱 audit 表其餘各列的邊界都是對的——只驗過這兩處。
