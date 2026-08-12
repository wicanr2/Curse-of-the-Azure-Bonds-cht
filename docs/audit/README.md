# 原始碼資料分離稽核

本目錄保存機械產生、可由正式測試驗證的技術債基線。它不是豁免清單，也不是
允許 Go 保留固定數量中文的「額度」。

## ECL 全事件靜態清冊

[`ecl-event-catalog.json`](ecl-event-catalog.json) 與
[`ecl-event-catalog.md`](ecl-event-catalog.md) 由 `cmd/ecl-event-catalog` 從原始
`curseoftheazurebonds.zip` 重生。它保存六個 ECL DAX、25 個 block、125 個 lifecycle
entry、靜態可達 instruction／edge 與跨 effect-kind 候選。這是 parser／控制流 inventory，
不是完整 runtime side effects；限制與驗證見
[`spec 557`](../spec/557-ecl-event-catalog-and-ordered-effects-audit.md)。

JSON 對 packed text operand 只保存長度與 SHA-256，不複製原文 payload。

## Go 漢字字串基線

`go-han-literals-baseline.json` 由 `cmd/coab-audit` 使用 Go AST 產生，只掃正式
非測試 `.go` 字串 literal：

- 忽略註解、`*_test.go`、JSON／Markdown、`workplace/` 及 nested engine；
- 以 repository-relative path、函式、完整字串 SHA-256、出現次數及債務分類
  建立 exact multiset；
- 不把中文內容複製進基線；
- 新增、改字、搬動、增加副本或刪除後未更新基線，都會讓
  `TestRepositoryGoHanLiteralBaselineIsExact` 失敗。

目前分類是遷移排序用 heuristic：

- `localization_debt`：位於 localize／Journal bridge 的歷史字串；
- `frontend_ui_debt`：Ebiten command 的玩家可見 UI；
- `runtime_ui_debt`：其他 runtime、工具與尚未細分字串。

分類不代表後兩者可以永久留在 Go。最終目標是：本作劇情、人名、地名、物品、
法術、選項、手札與玩家可見 UI 都由 stable ID＋locale／game-pack 驅動；Go 只
保留 action、format contract、layout 與必要技術診斷。

每次遷移流程：

1. 先把正式文字移入 locale／game-pack，接通 stable ID 並測正常玩家路徑。
2. 刪除 Go fallback 及其他資料複本。
3. 在 Docker 內執行 `go run ./cmd/coab-audit -write-baseline`，只能接受數量下降
   或經審查的分類改善。
4. 執行 `go run ./cmd/coab-audit` 與正式套件 gate。
5. 在 READY spec／狀態表記錄前後數量；不可只更新基線掩蓋新增債務。

第 452 輪初始證據：1,260 個 signatures、1,315 次 occurrences，其中
`localization_debt=409`、`frontend_ui_debt=164`、`runtime_ui_debt=742`。
