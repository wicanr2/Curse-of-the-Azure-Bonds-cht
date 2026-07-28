# READY 341：獨立 Golden Box engine 與 JSON game pack

## Repo 邊界

- `wicanr2/golden-box-remake-engine`
  - 通用 Go runtime；
  - versioned JSON Schema；
  - ECL block／memory／party predicates；
  - ordered event actions、text rules、journal output 與 exactly-once。
- `Curse-of-the-Azure-Bonds-cht`
  - CoAB JSON game pack；
  - 原始 ECL／GEO 證據、翻譯、素材與 integration tests。

共用 engine 不得出現 `ALIAS`、`DRAGONBAIT`、`4C5B`、`7F12` 或 CoAB 敘事。
這些值只可出現在 game pack／範例資料。

## 第一個遷移事件

`gamepack/events/pit-of-moander.json` 宣告：

1. destination ECL block `81`；
2. memory `0x4C5B=255`、`0x7F12=1`；
3. Alias 存活／死亡 variant；
4. 依 script name 移除 Alias 與 Dragonbait；
5. 繁中告別與 pending world menu。

State 只建立通用 runtime snapshot，向 pack 查詢其宣告的 memory addresses，
再依 engine 回傳的 removed IDs／message／mode 更新 persistent projection。

## JSON text rule

同一 schema 以 `all_contains` 比對原始 ECL 文字，回傳 locale message 與可選
journal pages。散提爾堡入口的巡邏放行、城外提示、守衛盤問、伏筆、內城文字
及手札 32 均已移入 JSON。

## 驗收

- engine repo `go test ./...`；
- CoAB `go test ./gamepack ./internal/game`；
- real-image 長流程驗證離坑、前往散提爾堡、world value `12`、ECL4/GEO4
  block `0x20`、手札 32 與 `(2,0,S)` Dungeon entry；
- `rg` 證明 engine Go code沒有 CoAB 專屬名稱／旗標。
