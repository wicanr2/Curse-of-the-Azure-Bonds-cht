# 1242 — 高階牧師法術 QUICK 資料接線

狀態：`READY`

## 問題

第 1241 輪讓五級牧師能依原版 `SpellCastCount[3][5]` 準備三級法術後，正常按鍵
長跑第一次把 `Prayer`（全域 spell ID `42`）帶進 QUICK 戰鬥。玩家施法 handler
已存在，但 `combat_ai_spells` 沒有這一筆，因此第 1213 輪的失敗即關閉交接會正確
收回手動控制；測試驅動器再次按 QUICK 後則重複撞上同一筆缺口。

## 原版資料

`gamepack/rules/spell-table.json` 是由常駐資料段產生的 100 筆原版主法術表。

Prayer 的 raw record 為 `00030000000100040004310206050000`，已解碼欄位為：

- `ai_priority = 5`
- `requires_target_check = false`，因此 `cast_on = 0`
- `scan_radius = 0`，因此 `min_range = 0`
- `casting_time_segments = 6`

第一份診斷紀錄補完 Prayer 後仍失敗，進一步顯示同一負載中的 Hold Person（全域
spell ID `23`）也缺宣告。其 raw record 是 `06020004003406000104340105060100`，
對應 `ai_priority=6`、`requires_target_check=true`、`scan_radius=0` 與
`casting_time_segments=5`。這一輪只把上述既有資料接到 `combat_ai_spells`，不新增
或推測兩支法術的效果。

## 驗證

- `gamepack.TestPack` 逐欄釘住 Prayer 與 Hold Person QUICK metadata。
- 正常按鍵路徑以合法五級牧師負載準備五格 `Hold Person` 與一格 `Prayer`，用來驗證
  二、三級記憶槽確實能進入正式戰鬥路徑；這是測試控制樣本，不是尚待使用者決定的
  快速建角預設法術書。

## 延遲施法收尾（2026-08-27）

正常強度 trace 在 `ECL0x33` 暴露另一個跨層缺口：牧師的 pending Bless 動作被
scheduler 再次選中時，title roster 已沒有該記憶槽。舊 `resolvePendingSpell`
重新呼叫初始 `BeginCombatCast`，得到 `Bless is unavailable` 後直接返回；Battle
裡的同一 pending action 沒被取走，因此每一幀無限重試。

修正契約：pending 動作已通過初始合法性檢查；若完成時記憶槽已不存在，視為施法
中斷，原子取走 point／target action、不得套用效果，並推進 scheduler。這不改變
原版「被中斷會失去法術」的契約，只補上跨層狀態不一致時的失敗即關閉收尾。

代表性驗證：

- `TestPendingBlessWithoutMemorizedSlotFinishesInsteadOfRetryingForever` 重現缺槽但
  pending action 尚存的狀態，確認動作清零且 Bless 不生效。
- pending 中斷、手動 Bless 延遲、QUICK Hold Person／Prayer 五項精準回歸通過。
- `go test ./gamepack ./internal/game ./cmd/azure-bonds-game -count=1` 通過。
- 一般強度重放不再停在同一 Bless：2,455 幀走到 251 格／8 段／199 句、fallback 0，
  最後是一次真實全滅重開。手動戰鬥對照同樣在該段全滅，故下一缺口歸為戰術／
  loadout，而不是 pending scheduler。
- 依 2026-08-27 使用者決定，一般強度遇到真實全滅即停止並保留報表，不再要求
  該 attempt 通關；5,000 幀上限＋`COAB_KEY_REQUIRE_WIN=1` 的同一樣本已通過。
