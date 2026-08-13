# 第六百六十四輪：overlay manager 如何改寫 stub

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `1A2E6h`、`1A31Ch`、`1A29Ah`、`1A04Bh`。

## 兩支互為反向的自我修改

[spec 644](644-resident-reading-strategy.md) 記過 resident 裡有 50 支 body 只是
`INT 3Fh` 的 stub。這兩支就是改寫它們的地方：

```text
1A2E6h（載入後：stub → 直接跳）:
    if es:[20h] = 0EAh then return          ← 已經改過
    if es:[2] <> 0 then <sub_1A3B7>(es:[10h], es)
    bx := es:[10h]                           ← 載入後的 segment
    cx := es:[0Ch]                           ← 入口數
    di := 20h
    重複 cx 次:
        dx := es:[di+2]                      ← 原本 stub 裡的 offset
        寫入 0EAh, dx, bx                     ← EA off seg，共 5 bytes
        （stosb ＋ stosw ＋ stosw，di 自動前進 5）

1A31Ch（卸載時：直接跳 → stub）:
    if es:[20h] = 0CDh then return           ← 已經是 stub
    <sub_1A3B7>(es, es:[10h], 0)
    es:[2] := 0
    cx := es:[0Ch]
    di := 20h
    重複 cx 次:
        dx := es:[di+1]                      ← 現在的 offset（EA 之後那個 word）
        寫入 3FCDh, dx                        ← CD 3F ＋ offset
        …
```

`0EAh` 是 **far JMP** 的 opcode（`EA off seg`，5 bytes）；`3FCDh` 以 little-endian
寫出就是 `CD 3F` ＝ **`INT 3Fh`**（2 bytes）。所以：

- overlay **在記憶體時**，stub 被改寫成直接的 far jump——呼叫不經過中斷。
- overlay **被換出時**，改回 `INT 3Fh`——下次呼叫觸發中斷，manager 再載入。

判斷「要不要改」是**直接看第一個 byte 是 `0EAh` 還是 `0CDh`**，不是另外記旗標。

兩邊讀 offset 的位置差一（`es:[di+2]` 對 `es:[di+1]`），正是因為 `EA` 是 1 byte
opcode ＋ 4 bytes 運算元、而 `CD 3F` 是 2 bytes opcode ＋ 2 bytes。

## overlay 描述子的欄位

由這兩支與 `1A29Ah` 一起讀出：

| 位移 | 內容 |
|---|---|
| `+2` | 已載入的標記／計數（`1A31Ch` 卸載時清 0） |
| `+0Ch` | **入口數**（stub 的個數） |
| `+10h` | 載入後的 segment |
| `+14h` | 下一個 overlay（串鏈） |
| `+20h` 起 | **stub 區**，每筆 5 bytes |

`+20h` 每筆 5 bytes 與 [spec 644](644-resident-reading-strategy.md) 觀察到的
「`CD 3F` ＋ 3 bytes 描述子，以 5 bytes 為間隔」完全吻合——那 3 bytes 就是
`EA` 形式要用的 offset ＋ 一個 byte。

## `1A29Ah`：釋放到夠為止

```text
if word_23AE2h = 0 then { word_23ADEh := word_23ADCh ; return }
if es:[10h] >= word_23ADEh then return
cx := 1
重複:
    走 es:[14h] 串鏈到底，把 word_23AE2h 換成最後一個
    word_23ADEh := word_23AE0h
    把節點接回串鏈頭
    <sub_1A3FCh>()
    word_23ADEh := word_23ADEh − 回傳值
    <sub_1A34Fh>()
until --cx = 0
```

`cx` 在迴圈裡被 `push`／`pop` 保存，`loop` 只跑 `cx` 次——而 `cx` 進迴圈前是 1。
內層的 `inc cx` 在另一個小迴圈裡累加，**外層實際跑幾次取決於內層數到多少**。

## `1A04Bh`：設定前的四道檢查

```text
if word_23AE4h <> 0 then return 0FFFFh
if word_23AE2h <> 0 then return 0FFFFh
if (word_23AEAh − word_23AEEh) or word_23AECh <> 0 then return 0FFFFh
addr := <sub_1A0FFh>(參數的 far pointer)
if addr < word_23AD6h then return 0FFFFh
addr := addr + word_23ADCh
if 進位 then return 0FFFDh
if addr > word_23B04h^[2] then return 0FFFDh
word_23AE0h := word_23AEAh := word_23AEEh := addr
return 0
```

兩種錯誤碼分工明確：**`0FFFFh` ＝ 狀態不允許**（已經設定過、還有 overlay 在用），
**`0FFFDh` ＝ 位址超出範圍**（含加法溢位）。加完才檢查上界，而溢位單獨用進位抓。

## 明確不宣稱

- `sub_1A0FFh`／`sub_1A3B7h`／`sub_1A3FCh`／`sub_1A34Fh` 的行為。
- `word_23AD6h`／`word_23ADCh`／`word_23ADEh`／`word_23AE0h`／`word_23B04h` 各自的
  角色（只知道是緩衝範圍與目前配置點）。
- overlay 描述子 `+4`..`+0Bh` 的欄位。
