# 執行期獨立 PNG 圖像

此目錄由 `cmd/runtime-image-export` 從本機原版資料一次性匯出；release 與 remake
執行期只讀這裡的 PNG 與 `manifest.json`，不再為下列圖像解碼原版 DAX：

- `tiles/`：48 張世界／區域地圖磚；
- `combat/`：65 張 `DUNGCOM`／`WILDCOM`／`RANDCOM` 戰場地形；
- `symbols/`：1,625 張 `8X8D1..6` 原生 8×8 符號，包含 AREA、第一人稱共用符號，
  以及下一階段重建牆片會使用的符號；
- `sky/`：3 張 `SKY` 圖像。

重生方式（須依根目錄 `AGENTS.md` 在 Docker 內執行）：

```text
go run ./cmd/runtime-image-export curseoftheazurebonds.zip assets/runtime-images
```

檔名與 manifest 同時保留原始來源、block、item；排列順序由 manifest 決定，runtime
不得靠目錄列舉順序猜索引。原版 ZIP 不屬於 release 資產，也不得加入 Git。

`manifest.json` 另保存 `WALLDEF2..6` 的 20 個 block；每筆仍帶原始來源與 block，
資料長度必須是 780 bytes（5×156）整數倍。runtime 由這些 JSON bytes 與 symbols PNG
重建 `PieceSet`，不再讀 `WALLDEF*.DAX` 或 `8X8D*.DAX`。
