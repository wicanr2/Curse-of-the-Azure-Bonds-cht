# 第 492 輪：診斷工具訊息資料分離

狀態：READY
日期：2026-08-09

## 結論

本輪把 CoAB 反組譯、PC-98 媒體稽核與 DOS 角色匯入工具中最後一批可編輯繁中
訊息移到 `internal/tooltext/messages/zh-TW.json`。Go 原始碼只保留穩定訊息 ID、
格式參數與輸出位置；工具行為、原始 bytes、位址與稽核報告欄位沒有改動。

`internal/tooltext` 使用 `go:embed` 載入工具專用 catalog。格式化錯誤另外保留
`%w` 的包裝行為，沒有把錯誤文字當成控制流或資料識別。PC-98 Sound BIOS 的
`SoundBIOSService.arguments` 仍是報告輸出欄位，但內容由同一份 catalog 產生，
不再在音訊 bridge 的 Go table 複製中文。

## 證據與邊界

- `cmd/coab-audit` 的 Go AST exact baseline 從 77 次 occurrence 降為 0；
  `docs/audit/go-han-literals-baseline.json` 是空 findings，而不是新增豁免。
- 原始中文移出的範圍是工具 help、錯誤訊息與 Sound BIOS 參數說明，不是遊戲
  劇情、規則、戰鬥數值或 ECL opcode 語意。
- JSON 的訊息 ID 不等於原版 bytes 證據；反組譯結論仍須依各 READY spec 的
  原始檔雜湊、位址空間、IDA／runtime 證據與推論等級判讀。
- 本輪是「引擎／資料分離」與工具鏈品質里程碑，不代表完整 ECL、完整戰鬥、
  全地圖、全翻譯、音效或開場到結局的單一路徑已完成。

## 驗證

在 Docker、網路關閉並使用工作樹內 engine replacement 執行：

```text
ROUND492_SOURCE_AUDIT=0
ROUND492_FORMAL_EXIT=0
go test -count=1 ./cmd/... ./gamepack ./internal/...  # Docker／Xvfb
```

所有 `cmd/...`、`gamepack` 與 `internal/...` 套件通過；Ebiten 套件以 Docker 內
Xvfb 啟動。`go test ./...` 仍保留既有 `scripts/` 多個獨立 `main()` 的結構限制，
不以該命令取代正式套件 gate。

## 可重用規則

後續 Gold Box 工具若有玩家可見或研究者可見的可編輯文字，應使用穩定 ID＋locale
catalog；Go 只保留格式契約、bytes／位址與控制流。作品專用原文、翻譯與來源
仍留在 game-pack 或各作品 repo，不升入共用 engine。此模式也可提供《冬之魔》
與未來《Wasteland》工具鏈參考，但不能因此共用它們的 native 格式或劇情資料。
