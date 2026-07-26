# Curse of the Azure Bonds 中文化 & remake 計畫

## 目標

在不依賴原始 DOS 執行環境的前提下，建立可驗證的 Golden Box 資料／腳本格式規格，之後以 Go + Ebiten 重製執行層，並加入繁體中文介面、訊息與手札內容。

## SDD 工作規則

1. 反組譯、檔案格式分析與執行觀察先記錄於 `docs/spec/`。
2. 規格只有在證據、未知項目與驗證方法寫清楚後，才能標示 `READY`。
3. 實作每輪都必須有可重現的測試或分析工具。
4. 每輪更新 Markdown、`CONTEXT.md`，commit 並 push 到 GitHub。
5. 原始遊戲映像與掃描手冊只作本地分析素材，不預設重新散布到 repository。

## 分階段

| 階段 | 產出 | 狀態 |
|---|---|---|
| 0. 基線與素材盤點 | 本計畫、素材清單、GitHub 基線 | 完成 |
| 1. DAX／EXE／ECL 格式分析 | `docs/spec/` 格式規格與樣本工具 | 進行中 |
| 2. ECL 最小直譯器 spike | 可讀取並追蹤一個最小場景 | 進行中 |
| 3. 遊戲狀態與 AD&D 規則 | 核心模型與相容性測試 | 進行中 |
| 4. 渲染、輸入、音效 | Go/Ebiten 可執行 prototype | 進行中 |
| 5. 中文化與手札 | 字串資源、字型、繁中內容 | 待開始 |
| 6. 整合與遊玩驗證 | DOS 對照測試與 release build | 待開始 |

## 第一輪驗收

- 完成原始 ZIP 的完整 manifest 與格式初步分類。
- 確認 `START.EXE`、`GAME.OVR`、`ECL*.DAX`、圖像／地城資料的大小與檔案標記。
- 建立第一版格式研究規格，明確區分已知、推測與待驗證項目。

## 第四輪驗收

- Go DAX container parser 能驗證 header、block boundary 與 RLE 解碼。
- Go locale catalog 能載入 `assets/locale/zh-TW.json` 的繁中資源並支援英文 fallback。
- CLI 能直接讀取原始 ZIP 的 ECL block metadata。
- ECL opcode、畫面、音效與完整劇情仍未完成，不得宣稱 remake 已完成。

## 第六輪驗收

- `internal/ecl.Trace` 能依 command arity 追蹤 decoded block。
- 未知 opcode／截斷 operand 會安全停止並保留已完成 trace。
- `go test ./...` 與原始 ECL CLI trace 已通過。

## 第七輪驗收

- Go ECL layer 能從 length-prefixed payload 解碼 6-bit packed text。
- 以原始 `ECL1.DAX` 的真實 payload 做 regression test。
- CLI 能列出英文原文候選，作為後續繁中翻譯資源輸入。

## 第八輪驗收

- `internal/game.State` 以 locale catalog 驅動繁中開場狀態轉移。
- 狀態核心有錯誤 action 與完整 opening flow 測試。

## 第九輪驗收

- Ebiten command 能編譯並使用 `internal/game.State`。
- 啟動畫面、輸入與繁中 catalog 已連通。
- 字型以外部路徑注入，避免將未確認授權的字型提交至 repo。

## 第十輪驗收

- Ebiten opening 由原始 `ECL1.DAX` block 初始化。
- 原始 marker 與繁中顯示文字在 state 層分離保存。
- parser、state 與 Ebiten compile verification 通過。

## 第十一輪驗收

- `GOTO/GOSUB` code targets 可轉換成 payload offsets。
- 靜態 graph 對未知／越界資料安全停止。
- ECL branch graph 有單元測試與原始 block CLI 驗收。

## 第十二輪驗收

- [x] 以公開 CoAB 重寫程式核對 ECL 初始化順序。
- [x] 加入五組 word-valued ECL 初始化入口的 bounded parser。
- [x] 修正已觀察 VM command table 的 arity metadata 並加入 regression test。
- [x] 正確消耗 length-prefixed compressed-string operand，並支援從指定 entry offset trace。
- [ ] 對全部實際 ECL block 驗證五個入口，並與事件文字對齊。

## 第十三輪驗收

- [x] 建立有步數上限的 ECL subset runner。
- [x] 實作 `SAVE/COMPARE/IF/GOTO/GOSUB/RETURN/PRINT` 與 packed text 輸出測試。
- [x] 未支援 opcode 以精確 payload offset 停止，未宣稱完整 VM。
- [ ] 實作並驗證 `ON GOTO/GOSUB`、menu input 與完整 memory model。
