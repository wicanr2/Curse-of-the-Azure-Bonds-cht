# 第六百四十四輪：resident 執行檔不適用 prologue 邊界

狀態：`READY`。等級：`exact`。日期：2026-08-14

## 動機

`PC98-GAME.EXE` 還有 113 支待解讀，是單一模組裡最多的。想沿用
`scripts/batch_small.py` 的做法（走 prologue 區間，不信 IDA 的 `size`），所以先為
resident 產一份 prologue 匯出。

## 先修好一個靜默的匯出缺口

`tools/ida/export_by_prologue.py` 原本只取 `ida_segment.get_first_seg()`。
raw overlay 只有一段，所以這個限制一直沒有浮現；但 `PC98-GAME.EXE` 是多段的 MZ
檔，**第一次跑只匯出 2 支函式**——輸出檔存在、格式正確、沒有任何錯誤訊息。

改成走完所有 segment（每支的終點取「下一個 prologue」與「所屬 segment 結尾」的
較小值）之後，匯出 200 支。

## 但 prologue 邊界不適合 resident

| | 數量 |
|---|---|
| IDA 函式清單（`small` 匯出） | 271 |
| prologue 區間匯出 | 200 |
| **兩者起點相同的** | **85** |

只有 85 支對得上。原因是 resident 裡有大量**不以 `55 89 e5`／`55 8b ec` 開頭**的
常式：

- **TPOV overlay stub**：body 只有 `INT 3Fh`（2 bytes），後接 3 bytes 描述子，
  以 5 bytes 為間隔連續排列。PC-98 側 50 支、DOS 側 49 支——這批在
  [spec 569](569-small-function-batch-reading.md) 已判為不阻塞。
- **手寫組語的 RTL**：Turbo Pascal 的 RTL 有不少常式直接 `push si` 或
  `mov ax, ...` 開頭，沒有標準 prologue。

overlay 是編譯器產生的 Pascal 程序，幾乎每支都有標準 prologue，所以那邊
prologue 邊界比 IDA 可靠；resident 不是這種東西。

## 結論：resident 走 IDA 邊界 ＋ 截斷檢查

讀 resident 時用 IDA 的函式範圍，再用
`scripts/verify_size_truncation.py` 確認最後一條是 `ret`／`jmp`。這支工具本來就是
為了抓「IDA 範圍在半途停住」而寫的，對 resident 一樣適用。

prologue 匯出仍然留著（`workplace/re-sweep/pc98/overlays/prologue/pc98-PC98-GAME.EXE.json`），
但**不可**當成 resident 的函式清單——它會漏掉那 186 支起點不同的常式。

## 明確不宣稱

- 兩平台 stub 數量差 1（50 對 49）的原因。
- 三個描述子 byte 的欄位切法。
