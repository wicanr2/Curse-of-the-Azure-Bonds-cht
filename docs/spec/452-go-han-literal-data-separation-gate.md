# 452：Go 漢字字串與資料分離 gate

狀態：`READY`

## 問題

第 450–451 輪雖已把完整法師塔作品文字移入 JSON，專案其他歷史 Go 程式仍有
大量中文字串。只靠 agent 記憶或人工 `rg` 無法防止新功能再次把譯文塞回
State／frontend，也無法量化遷移是否真的前進。

## 實作

新增 `internal/sourceaudit` 與 `cmd/coab-audit`：

- 以 `go/parser`／AST 掃描 string literals，不掃註解或原始碼文字片段；
- 排除測試、研究工作區、JSON 資料與 nested engine；
- 以 path＋enclosing function＋完整 SHA-256 聚合相同 literal 的 occurrence；
- exact baseline 比對新增、移除、改字、搬動、重複數與分類漂移；
- baseline 只存 hash，不複製中文內容。

`internal/sourceaudit` 的 repository test 已納入 `go test ./internal/...` 正式 gate。
移除技術債後若不在同一變更更新下降後基線，測試也會失敗，避免舊 signature
日後悄悄回流。`-write-baseline` 是明確的機械更新操作，不會自動執行。

## 初始量化

| 分類 | 次數 |
|---|---:|
| 本地化／劇情 fallback 債 | 409 |
| Ebiten 前端 UI 債 | 164 |
| runtime／工具／未分類 UI 債 | 742 |
| 合計 | 1,315 |

共有 1,260 個獨立 signatures。這是誠實的初始技術債，不是 1,315 次永久豁免。
分類是排程 heuristic，不能拿來宣稱 runtime 字串一定作品中立。

## 驗證

- synthetic fixture 證明會掃 package／function string，忽略註解、測試、JSON
  與 `workplace/`。
- occurrence 由 2 降為 1 時 exact compare 同時回報 current／baseline drift，
  直到同一 commit 更新 reduced baseline。
- repository baseline test 與 `go run ./cmd/coab-audit` 均已通過。
- 正式 Docker gate 與 marker 記錄於 `CONTEXT.md`。

## 尚未完成

- 逐一人工把 742 筆 runtime 類再分為玩家 UI、技術診斷與真正作品中立 fallback。
- 將 409 筆本地化債按正常玩家章節逐批遷入 JSON；不能大量搬字卻不驗證 ECL
  source identity、Journal timing 與畫面取得同一份資料。
- 玩家可見字串最終應接近零 Go literal；技術診斷若保留，須建立理由明確的
  獨立分類，而非默認白名單。
