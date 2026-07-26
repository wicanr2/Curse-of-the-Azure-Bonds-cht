# 第九十輪：ITEMS base catalog

狀態：`READY`（限 ITEMS descriptor codec 與已知名稱 catalog）

## 證據

原始映像的 `ITEMS` member 是 2,050 bytes。公開 Gold Box file-format reference 說明它由 2-byte header 加上 16-byte base-item descriptor 組成；因此本映像可解成 128 筆 descriptor。每筆 descriptor 的 type 是其 table index，並保存 slot、雙尺寸傷害、rate-of-fire、AC adjustment、weapon type、range、class usability mask 與 ammunition type。

本輪 `monster.ParseBaseItems` 會嚴格驗證「2-byte header + 16-byte records」邊界，並以 `BaseItemCatalog.Lookup` 安全查詢。未知的最後兩個欄位仍保留為 raw bytes，不做猜測。

## 繁中化邊界

`ChineseName` 已擴充公開攻略列出的基本武器、護甲、彈藥、卷軸與常見魔法物品名稱；未出現在已驗證 catalog 的 type 仍顯示 `未翻譯物品(0xNN)`。這是名稱／資料層，不代表商店、裝備限制、法術效果或 inventory UI 已完成。

## 驗證

- `cmd/azure-bonds -base-items` 直接從本地原始 ZIP 讀取 `ITEMS`，目前實際 header bytes 是 `00 76`，以 little-endian 顯示為 `0x7600`，並報告 `descriptors=128`。
- `internal/monster/equipment_test.go` 覆蓋 type index、signed damage bonus、lookup 邊界與 malformed input。
- `go test ./...` 應通過。

格式交叉參考：[Curse of the Azure Bonds FAQ 的 Item File Format](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365#12.2.3)。
