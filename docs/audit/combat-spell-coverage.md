# 玩家戰鬥法術 coverage audit

工具：`cmd/combat-spell-coverage-audit`

此 audit 只讀取正式 `gamepack/events/pit-of-moander.json` 的
`combat_player_spells`，再以 Go AST 檢查既有
`internal/game/combat_state.go` runtime callsite，並查詢 game-pack 已宣告的
`combat_visuals`。音效欄位則記錄各 `combatCast*` 函式實際呼叫的
`requestSound` intent。

它不新增法術 registry、不修改戰鬥規則、不解讀原版語意，也不把 JSON 宣告當成
完整實作。`missing`／`partial` 是待辦證據，不能直接解讀為原版缺少該效果；同樣，
`observed` 只代表目前 remake source callsite 與 binding 可被機器觀察，不代表
AD&D 公式、原版 timing、素材或音訊已 exact。

在可取得目前 engine module cache 的 Docker 工具鏈中執行：

```sh
go run ./cmd/combat-spell-coverage-audit
go run ./cmd/combat-spell-coverage-audit -strict
```

輸出是機器可讀 JSON；標準錯誤另輸出 `spell_count`、`covered` 與
`incomplete` 摘要。預期每一筆正式 stable spell ID 都有一列，並分開保存：

- `handler`：dispatch case、runtime function 與狀態。
- `visual`：pack phase、runtime visual kind 與狀態。
- `sound`：實際觀察到的 sound intent 與狀態；若由共用 `AdvanceCombatVisual`
  產生，會標為 `observed_shared`，不冒稱是該法術專屬 cue。
- `limitations`：不完整時的明確限制，避免誤宣稱完整法術。

此工具刻意放在獨立 `cmd/` 目錄，不加入 `scripts/`，避免觸發既有兩個
`main` 的 gate 問題。
