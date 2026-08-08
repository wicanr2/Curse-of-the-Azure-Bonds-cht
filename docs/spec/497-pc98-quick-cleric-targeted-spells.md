# 第四百九十七輪：PC-98 Quick 牧師指定目標法術

狀態：`READY`（只關閉 CoAB Quick 的 Curse、Cause Light Wounds、Protection
from Evil／Good bounded targeted slices；原版候選 pointer chain、優先序與
完整敵方 Quick AI 仍未完成）

## 本輪結論

正式 game-pack `combat_ai_spells` 將以下牧師法術提供給 PC-98 Quick selector：

| spell ID | 法術 | priority | cast_on | min_range | casting_time |
|---:|---|---:|---:|---:|---:|
| `02h` | Curse | 3 | 1 | 0 | 10 |
| `04h` | Cause Light Wounds | 2 | 1 | 0 | 5 |
| `06h` | Protection from Evil | 1 | 0 | 0 | 4 |
| `07h` | Protection from Good | 1 | 1 | 0 | 4 |

本輪將四個 selector 分支接到既有手動法術 runtime：

1. Curse 使用目前存活敵人列表的第一個 stable entry。
2. Cause Light Wounds 使用既有鄰接敵人 contract；沒有鄰接敵人時維持
   unsuitable／fail-closed，不把遠方敵人當成近戰法術目標。
3. Protection from Evil／Good 使用既有鄰接／自身 party target contract。
4. 四者都將 target ID 送進 `BeginPendingTargetedSpellAction`，由
   `CastingTime/3` 的 pending scheduler 續跑，再呼叫原有 Curse／Cause／
   Protection effect writer、文字／音效與法術格 transaction。

這是可玩的 Quick 牧師指定目標進度，不是 PC-98 object pointer 的候選順序、
random helper 或 `cast_on` 欄位完整規則的逐指令還原。

## 非破壞性輸入與位址空間

| 輸入 | SHA-256 | 用途 | 推論等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | resident symbols／全域資料 | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay control | `exact` |
| overlay 09 | `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20` | Quick selector／suitability | `exact` |
| overlay 09 IDA report | `3f7f6160e9aac8140ecf320ae45f34572b6d77a39bc7b5f1eb6c8f417c30b36c` | 連續指令與控制流輸出 | `exact bytes／control flow` |

IDA Pro 9.4 使用既有 `ida-pro-9.4-ver2:uidfix-v1` image 的有效 default
entrypoint。分析前將唯讀輸入複製至 `/tmp/coab493-ida`；原始 binary、既有
database、Borland symbol 與位址均未改寫。報告中的 `local=xxxx` 是
overlay-local 位址，不能與 resident linear address、segment:offset 或 file
offset 混用。

## 原始資料與現有 runtime 證據

- overlay 09 local `03D3h..04C9h` 讀 priority、`CastOn`、`MinRange`，再由
  `04CCh..0627h` 的角色 target pointer chain 建立候選，local `072Bh..0754h`
  交給 `CASTCOMBATSPELL`。這些 branch／handoff 是 `exact control flow`；
  target pointer 到作品中立 stable ID 的映射、候選排序／random 是 `unknown`。
- `docs/spec/425-pc98-quick-bless-casting-delay.md` 的 `/3` delay contract
  證明 raw casting time 10／5／4 分別形成非零 pending action；本輪資料值與
  pending transaction 為 `exact` remake contract，不能反推原版完整 target AI。
- CoAB 手動 runtime 已有 `CombatSpellTargets`、`causeLightWoundsTargets`、
  `protectionFromEvilTargets`、`protectionFromGoodTargets` 與四個 effect writer；
  這些現有 target／效果 boundary 是本輪重用的第一級本機證據。其 stable ID
  第一筆政策只支持 bounded remake continuity，不支持原版逐指令順序。

## Remake 實作邊界

- 新增 `quickTargetedSpellTarget`，依 spell ID 選取現有作品中立 target
  contract；它不新增 Curse、角色名稱、地圖座標或繁中字串到 engine。
- `tryQuickSpell` 只在合法 target 與相應 `CombatCanCast...` gate 同時通過時，
  讓四個法術成為 Quick candidate。選中後重新以 `CombatSpellTargets` 找回
  target index，確保 pending target ID 與 effect writer 使用同一筆資料。
- pending resolution 保持同一個 `Block／combat action` 生命週期；受到正傷害、
  被死亡效果中斷或 target 消失時，沿既有 transaction rollback／clear 邊界，
  不複製一套 Quick 專用 slot removal。
- `cast_on` 仍只作 game-pack selector metadata；本輪沒有把它直接命名成
  「敵方／我方」語意。真正 target side 來自既有 spell ID adapter 與效果
  consumer，避免 compact 後把欄位名稱誤當規則證據。

## 驗證

- `TestCombatAltMQuickCurseUsesPendingEnemyTarget` 驗證 Curse 保存敵方 stable
  target，續跑後寫入 cursed／attack modifier／duration，最後才消耗 slot。
- `TestCombatAltMQuickCauseLightWoundsUsesPendingAdjacentTarget` 同時放置
  鄰接與遠方敵人，驗證只保存鄰接 target，續跑後造成傷害且遠方敵人不變。
- `TestCombatAltMQuickProtectionSpellsUsePendingPartyTarget` 以 table-driven
  stable spell IDs 驗證兩種 Protection 都保存自身 party target，續跑後寫入
  對應 active effect 與 duration，最後消耗 slot。
- 同一 focused Docker gate 亦通過 Quick Lightning、Sleep、Fireball、Stinking
  Cloud、Cloudkill、Bless pending 與手動 combat regressions。
- Docker／Xvfb 正式 gate
  `go test -count=1 -p 2 ./cmd/... ./gamepack ./internal/...` 通過，
  `go run ./cmd/coab-audit -root .` 回報 `total=0`；本輪 marker 為
  `ROUND497_FORMAL_EXIT=0`。

## 後續缺口

1. 以 PC-98／DOSBox 關閉 Quick object pointer 的 target priority、tie／random
   與完整 `cast_on` consumer 語意；stable ID 第一筆不可升格為 exact。
2. 完成敵方 Quick cleric／magic-user AI 的 target safety、抗性、死亡／中斷與
   動態演出，並與原版逐幀對照。
3. 完成完整 ECL 玩家路徑、全地圖、全規則、全翻譯、音樂音效、原版 save 與
   發行包；本規格不能支撐「完整 remake 已完成」的聲明。
