# 第509輪：PC-98 per-action target 與 QUICK 清除邊界

狀態：`READY`（只關閉 typed action-target projection 與 QUICK 同隊清除）

本規格不是完整敵方 AI、移動、逃跑、守備或原始 far pointer layout。它只把已在
第421／422輪關閉的 PC-98 Action writer／consumer 接到可重用 engine 與 CoAB
game-pack，避免延遲法術 target 與每回合 action target 共用欄位而互相清除。

## 輸入與位址基準

- `GAME.EXE` SHA-256：
  `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`
- `GAME.OVR` SHA-256：
  `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a`
- 工具：IDA Pro 9.4 Docker 副本、PC-98 16-bit overlay-local 位址；原始檔與
  IDA database 唯讀。證據沿用 `docs/spec/421-pc98-combat-quick-guard-bandage-speed.md`
  與 `docs/spec/422-pc98-combat-all-quick-interrupt.md`，不改寫原始名稱或位址。

## 證據與推論等級

| 原始定位 | 觀察 | 等級 |
|---|---|---|
| overlay 24 local `2A5Bh..2AADh`，resident `014A:00CA` | clear-actions 清理 Action 的 delay、pending 欄位、guarding 與另一個欄位 | `exact` |
| overlay 8 local `1375h..140Fh` | Quick setter 寫入 quick flag；目前 target 非空且 target `combat_team` 與自己相同時清除 target far pointer | `exact` |
| overlay 8 local `0677h..06B6h` | ALT+Q 依 TeamList 對每個 combatant 呼叫同一 Quick setter | `exact` |
| stable ID `ActionTargetID` 對 raw far pointer 的對應 | 以可保存的作品中立識別取代指標，未宣稱 raw layout | `strong inference` |

`Action+06h` 的原始語意仍未知；本輪只將已由 Quick target branch 證明的 target
pointer 另存為 typed projection，不把欄位 offset 命名成未證實規則。

## Remake contract

engine `combat/action.State` 現分離：

- `ActionTargetID`：目前回合 action 的 opaque stable target；
- `TargetID`：延遲單體法術的 handoff target；
- `TargetX／TargetY／HasTargetPoint`：延遲格點法術的 handoff target。

`TakeTargetedSpell`、`TakePointSpell` 與 `InterruptSpell` 只清除法術 transaction，
不可誤清 `ActionTargetID`。完整 `Clear`、死亡、Guard 或明確 action 結束才清除整筆
Action。CoAB `Battle.SetActionTarget` 只驗證 stable fighter identity；隊伍語意由
adapter 判定。

game-pack `combat_action_rules` 新增 `clear_same_team_on_quick`。CoAB 正式資料設為
`true`；State 的 QUICK／ALT+Q 讀取該欄位後呼叫 Battle policy。政策開啟時只清同隊
target，敵隊 target 保留；沒有正式 pack 的 synthetic 狀態保留既有驗證過的預設，
不影響正式 JSON 路徑。

## 驗證

- engine `combat/action`：action target 與 delayed spell target 可同時存在，取出
  法術後 action target 仍在，空 ID 失敗即關閉。
- engine `pack`／schema：`clear_same_team_on_quick` 由 JSON 載入並保留。
- CoAB `internal/combat`：QUICK 清除同隊 target、保留敵隊 target，且 policy 關閉
  時可保留同隊 target。
- CoAB `internal/game`：正式 game pack → State QUICK → Battle policy 的同隊清除
  regression 通過；既有延遲施法、戰鬥與資料分離測試維持通過。

## 未完成邊界

- 原始 Action far pointer 的配置、save bytes 與同格多 target 排序仍未知。
- 完整 movement／flee／guard producer、敵方 AI priority、persistent target 的
  產生與取代規則仍需 PC-98 bytes／runtime／影片逐項閉合。
- 本輪沒有宣稱完整戰鬥或完整 CoAB 通關。
