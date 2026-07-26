# 第一輪素材與格式盤點

狀態：`DRAFT`

## 來源

- `curseoftheazurebonds.zip`：原版 DOS 映像。
- `Curse-of-the-Azure-Bonds_*`：英文手冊、冒險者手札、線索書與參考卡。
- `珍020-青色枷的詛咒.rar`：中文掃描手冊候選素材。

## 映像分類（依檔名與大小的初步推論）

| 類別 | 樣本 | 初步用途 | 證據狀態 |
|---|---|---|---|
| 啟動程式 | `START.EXE` | DOS 啟動與載入 | 已由 MZ 標記確認，載入流程待反組譯 |
| overlay／共用程式 | `GAME.OVR` | 主遊戲程式或 overlay | 檔名與大小支持，格式待驗證 |
| 事件腳本 | `ECL1.DAX`–`ECL6.DAX` | 章節／區域事件腳本 | 按序命名且大小集中，內容格式待驗證 |
| 地城資料 | `GEO*.DAX`, `WALLDEF*.DAX`, `TILES.DAX` | 地圖、牆面、地城圖塊 | 檔名語意支持，資料佈局待驗證 |
| 圖片 | `PIC*.DAX`, `CPIC*.DAX`, `BIGPIC*.DAX`, `TITLE.DAX` | 場景、角色或標題圖 | 檔名語意支持，像素格式待驗證 |
| 角色／怪物 | `HEAD*.DAX`, `BODY*.DAX`, `MON*` | 角色、怪物與物品 | 檔名語意支持，索引格式待驗證 |
| 其他資料 | `ITEMS`, `*.BAT`, `COPYCURS.EXE` | 物品表、DOS 工具與批次流程 | 原始檔案存在，語意待驗證 |

## 已知檔案標記

- `START.EXE` 開頭為 DOS MZ header（`4D 5A`）。
- `GAME.OVR` 開頭為 `54 50`（ASCII `TP`），不是 DOS MZ header。
- 六個 ECL 樣本的前兩位元組依序為：`ECL1=1B 00`、`ECL2=24 00`、`ECL3=24 00`、`ECL4=2D 00`、`ECL5=2D 00`、`ECL6=24 00`。因此前兩位元組不是目前可確認的跨檔案 magic；可能是 record 長度、版本或資料本身。
- `GAME.OVR` 與各 `.DAX` 不應假定是標準可執行檔或標準影像格式；需要從位元組樣本與交叉引用驗證。

## 第一輪可重現測量

由 `scripts/inspect_image.py curseoftheazurebonds.zip` 產生。工具會依檔名分類，但分類只代表研究假設，不代表格式已解碼。

| 檔案 | 大小 | 前綴 | 初步分類 |
|---|---:|---|---|
| `START.EXE` | 69,968 bytes | `4D 5A` | executable |
| `GAME.OVR` | 272,218 bytes | `54 50` | overlay candidate |
| `ECL1.DAX`–`ECL6.DAX` | 16,530–27,437 bytes | 不一致 | event-script candidates |
| `ITEMS` | 2,050 bytes | `00 76` | table candidate |

## 待驗證問題

1. 每個 DAX 是否共用同一個索引／壓縮容器，或只是副檔名相同的不同格式？
2. ECL 是否含有可辨識的 script header、entry point table、字串區與 opcode stream？
3. `GAME.OVR` 是 DOS overlay、原始 flat binary，還是資料與程式混合檔？
4. 原版文字是否採用 code page 437、客製字型索引或壓縮字串？

## 驗證方法

- 以固定的 ZIP manifest 保存檔名、大小、mtime、SHA-256。
- 以 `file`／十六進位樣本／DOS 反組譯檢查 executable header。
- 對 DAX 做跨檔案 magic、長度欄位、索引單調性與重複結構掃描。
- 在規格能描述欄位與錯誤行為前，不標示 `READY`，也不實作引擎解析器。
