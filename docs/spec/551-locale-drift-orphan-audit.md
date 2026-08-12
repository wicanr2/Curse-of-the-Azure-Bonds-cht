# 第 551 輪：locale drift／orphan 稽核

狀態：`READY`（audit contract）

本輪新增 `internal/localeaudit` 與 `cmd/locale-drift-audit`。它只讀取資料並
輸出報告，不修改正式翻譯、State、renderer、主線測試或 README。

## 責任邊界

- `assets/locale/zh-TW.json` 是前端／runtime UI catalog，稽核產品層 Go 的
  `Text("stable_key", ...)` 呼叫是否都有對應 key。
- `gamepack/events/pit-of-moander.json` 的 `locales.en`／`locales.zh-TW` 是
  劇情、手札、事件、選項、戰鬥與作品資料的 content catalog；兩者必須 key 對稱。
- 兩個 catalog 不要求 key 集相等；跨層同名或不同名都不構成 drift。

## 會失敗的違約

1. 產品層 literal locale key 不存在於 assets catalog。
2. gamepack `en`／`zh-TW` key 不對稱。
3. `message_id` 找不到 en 或 zh-TW 值。
4. `option_rules` 缺少 `id`、`source`、`message_id`，或 stable ID／source 重複。

## 只報告、不失敗的項目

- assets catalog 中沒有被靜態 literal 呼叫找到的 key。
- gamepack locale 中沒有被 `message_id`／option binding 找到的 key。

這些 orphan 可能是動態 key、未接通的未來內容或保留資料，必須先人工確認，
不能因未被靜態掃描就自動刪除或讓 CI 失敗。

## 明確排除

- `*_test.go`、測試 oracle、原始 ECL 英文 token、`currentOriginalChoices`、
  工具命令訊息與 `internal/tooltext/messages/zh-TW.json` 不算玩家翻譯引用。
- 動態 `Text(key, fallback)` 的 key 不會被猜測；工具會記錄 dynamic call 數量，
  但不把 fallback 或原始 token 當成翻譯缺漏。

## 重現與驗收

在 Docker 中執行：

```text
go test ./internal/localeaudit ./cmd/locale-drift-audit
go run ./cmd/locale-drift-audit -root .
go run ./cmd/locale-drift-audit -root . -json
```

目前 audit 的 orphan 是資訊性輸出；只有 violation 才以非零狀態結束。
