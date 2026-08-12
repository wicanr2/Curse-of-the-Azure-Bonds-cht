# 第五百五十二輪：README 截圖證據 manifest

狀態：`READY`（截圖檔案完整性與證據索引；不代表 UI fidelity 或整作通關）

本輪只建立獨立的 `docs/screenshots/manifest.json` 與 validator，沒有修改
renderer、README、正常主線測試或事件 game-pack。manifest 將現有五張 README
代表圖的檔名、640×480 尺寸、SHA-256、生成模式、證據等級與對應 spec 集中保存。

## 驗證契約

- 現有圖片必須存在、可由 PNG header 解讀，且實際尺寸符合 manifest 與 640×480
  canvas。
- SHA-256 必須是 64 個十六進位字元，且每次 validator 執行都重新計算；檔案被
  替換或修改時 fail-closed。
- `evidence_level` 只能使用 AGENTS.md 的
  `exact`、`nearby`、`material-exact/layout-reconstructed`、`layout-only`、
  `hypothesis`、`unknown`。這些等級描述證據範圍，不把圖稱為完整原版畫面。
- planned 項目不得假造檔名或 SHA-256；目前只記錄正常玩家路徑、缺口描述與
  相關 spec，並以 `unknown` 表示尚無畫面證據。
- validator 是檔案／尺寸／雜湊／schema 級 gate，不會把靜態截圖升格為完整
  戰鬥、完整地圖或正常通關證據。

## 目前明確缺口

manifest 保留以下 planned 項目，不能解讀為已完成：正常 `VIEW` 角色資訊頁、
`AREA` 地圖、世界旅行 `overland` map，以及法術施放／飛行／命中／後續效果的
正常戰鬥關鍵幀。這些仍需由正常玩家路徑重拍，並依畫面狀態、座標、seed、theme
及證據等級補入，不得直接把既有 vertical slice 截圖代替。

## 重現與驗收

在 repository 根目錄執行：

```text
go run ./cmd/screenshot-audit
go test ./internal/screenshotmanifest
git diff --check
```

本輪實際驗收使用 Docker、`--network none`、有界 CPU／記憶體／PID，並以暫存測試
資料確認缺檔、尺寸錯、未知等級與雜湊漂移都會被拒絕。
