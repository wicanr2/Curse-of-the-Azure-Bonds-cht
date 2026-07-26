# 第三輪：DOS loader 與 GAME.OVR

狀態：`DRAFT`

## `START.EXE` MZ 實證

由原始映像解析 DOS MZ header：

| 欄位 | 值 |
|---|---:|
| 檔案大小 | 69,968 bytes |
| pages / last page bytes | 137 / 336 |
| relocation count | 483 |
| header paragraphs | 123（1,968 bytes） |
| entry CS:IP | `0000:002F` |
| entry file offset | `0x07DF` |
| initial SS:SP | `150C:2710` |
| overlay number | 0 |

檔案大小與 `137 × 512 − (512 − 336)` 完全相等，MZ image 邊界可重現。

## loader 字串實證

`START.EXE` 含有下列明碼字串，直接支持啟動流程包含：

- 顯示 `Curse of the Azure Bonds v1.0` 與 `Play Demo`。
- 載入 `GAME.OVR`，找不到時顯示 `Overlay error, program abort!`。
- 依序處理 `*.DAX`，並在需要時顯示 `Please insert overlay disk.`。
- 設定 CGA/EGA/Tandy 顯示卡、PC/Tandy/靜音音效與 save path。
- 找不到資料檔時顯示 `Couldn't find ... Check install.`。

## `GAME.OVR` 實證

- 檔案大小：272,218 bytes。
- 開頭 bytes：`54 50 4F 56`（ASCII `TPOV`），不是 MZ executable header。
- 內含標題、製作人員、戰鬥選單、法術／治療訊息與遊戲中的提示文字。
- 因此目前最保守的模型是：`START.EXE` 是 DOS 啟動／overlay 管理程式，`GAME.OVR` 是由它載入的非 MZ 主程式或 overlay payload；「可直接 EXEC 的第二個 EXE」尚未證明。

## 待驗證

1. `START.EXE` 的 overlay loader 是 DOS `EXEC`、自訂段落載入，或兩者混合？
2. `GAME.OVR` 的 `TPOV` 是版本標記、作者標記還是資料格式 magic？
3. `GAME.OVR` 的 16-bit machine code 範圍、字串池與資料表如何分界？
4. 哪些 DAX 由 loader 載入，哪些由主程式自行讀取？

下一輪將對 `START.EXE` 的 relocation／DOS interrupt 呼叫與 `GAME.OVR` 的入口候選做反組譯取樣；在確認載入邊界前，不會把 OVR 當作可執行格式實作。
