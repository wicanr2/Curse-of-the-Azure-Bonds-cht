# 玩家戰鬥法術 coverage audit

工具：`cmd/combat-spell-coverage-audit`；逐支法術的台帳在
[`combat-spell-coverage-ledger.md`](combat-spell-coverage-ledger.md)
（`-output` 產生，不要手改）。

**分母是原作那 100 筆，不是 game pack 宣告過的那幾支。**
用宣告數當分母會得到「12 支裡 3 支完成」這種與玩家無關的比例；台帳的摘要因此
從表的筆數往下扣：占位 13 筆玩家取不到、`+0Bh = 0` 的 8 支只能紮營施放，
剩下 **79 支**才是戰鬥法術的分母（法術表本身見 spec 1111）。

此 audit 讀取 game pack 的 `combat_player_spells`，再以 Go AST 檢查既有
`internal/game/combat_state.go` runtime callsite，並查詢 game-pack 已宣告的
`combat_visuals`。音效欄位則記錄各 `combatCast*` 函式實際呼叫的
`requestSound` intent。

它不新增法術 registry、不修改戰鬥規則、不解讀原版語意，也不把 JSON 宣告當成
完整實作。`missing`／`partial` 是待辦證據，不能直接解讀為原版缺少該效果；同樣，
`observed` 只代表目前 remake source callsite 與 binding 可被機器觀察，不代表
AD&D 公式、原版 timing、素材或音訊已 exact。

在可取得目前 engine module cache 的 Docker 工具鏈中執行：

```sh
tools/go.sh run ./cmd/combat-spell-coverage-audit
tools/go.sh run ./cmd/combat-spell-coverage-audit -strict
tools/go.sh run ./cmd/combat-spell-coverage-audit -quiet \
    -output docs/audit/combat-spell-coverage-ledger.md
```

stdout 是機器可讀 JSON（`-quiet` 關掉），`-output` 產生逐支法術的 Markdown
台帳；標準錯誤一行摘要同時給出兩個數字：宣告過的幾支完成，以及**原作可戰鬥
施放的 79 支裡宣告了幾支**。每一筆正式 stable spell ID 都有一列，並分開保存：

- `handler`：dispatch case、runtime function 與狀態。
- `visual`：pack phase、runtime visual kind 與狀態。
- `sound`：實際觀察到的 sound intent 與狀態；若由共用 `AdvanceCombatVisual`
  產生，會標為 `observed_shared`，不冒稱是該法術專屬 cue。
- `limitations`：不完整時的明確限制，避免誤宣稱完整法術。

此工具刻意放在獨立 `cmd/` 目錄，不加入 `scripts/`，避免觸發既有兩個
`main` 的 gate 問題。
