# 第四百九十五輪：PC-98 Quick 毒雲術／死亡雲範圍交接

狀態：`READY`（只關閉 CoAB 的 Quick Stinking Cloud／Cloudkill bounded slice；
原版候選鏈排序、隨機選擇與完整 Quick AI 仍未完成）

## 本輪結論

第 493、494 輪已將 PC-98 overlay 09 的 Quick `MinRange` 流程接到 Sleep 與
Fireball。本輪沿同一份證據，接通兩個已存在正式效果 pipeline 的毒雲法術：

1. `Stinking Cloud (22h)` 的 game-pack `min_range=1`。Quick selector 抽中後，
   以目前 `TACTICALMAP`、敵人戰鬥格與 `SCAN` bounded candidate 建立雲霧中心，
   立即重用既有 2×2 persistent-area、毒素豁免／咳嗽／無助、覆蓋與地面恢復
   pipeline。
2. `Cloudkill (5Bh)` 的 game-pack `min_range=0`。Quick selector 抽中後，從有
   合法戰鬥格且通過目前 line-terrain projection 的存活敵人建立 bounded point；
   raw `CastingTime=05h` 經 `/3` 形成 pending point action，scheduler 續跑後
   重用既有 3×3 persistent-area、低 Hit Dice 直接死亡／豁免、持續區域與施法
   中斷 pipeline。
3. 缺少必要地形投影、戰鬥位置或合法 candidate 時仍 fail-closed；不改成近戰、
   另一個法術或猜測性的 UI fallback。

這是可玩的 Quick 毒雲戰鬥進度，不是 PC-98 target linked-list、candidate
tie／random helper、完整 area-safety predicate 或完整 Quick AI 的逐指令還原。

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

## 原始控制流與資料值

- overlay 09 local `02D3h..03D0h` 建立非零 `MinRange` 的 SCAN candidate，
  依 actor record 的 `+198h` 準備方向／arc 參數，走目前戰場 object list，
  比較候選隊伍與角色狀態，再讀 spell record 的 `+61BCh/+61BDh` 並呼叫
  效果／豁免 predicate。這些 bytes 與 branch 是 `exact control flow`；offset
  的完整欄位名稱仍只標為 `strong inference`。
- local `03D3h..04C9h` 讀 priority、`CastOn`、`MinRange`，只有非零
  `MinRange` 才進入上述 helper。正式 game-pack 保存：Stinking Cloud `22h`
  為 `min_range=1`、`casting_time=2`；Cloudkill `5Bh` 為 `min_range=0`、
  `casting_time=5`。這些資料值與 `/3` delay 規則是 `exact`（raw table 欄位
  位置與轉換規格見 READY spec 425）。
- local `04CCh..0627h` 由角色 record 的 target pointer chain 建立候選，
  再以候選位置重呼叫 suitability；local `072Bh..0754h` 將選定 spell 交給
  `CASTCOMBATSPELL`。候選 linked-list 的完整 tie／random 行為仍是
  `unknown`，不可用 stable fighter order 宣稱逐指令相同。
- 本輪的 Cloudkill `MinRange=0` 分支，是針對現有遊戲資料與既有 line-target
  contract 的 bounded adapter；它不是證明 PC-98 helper 完全不做候選篩選。
  「存活敵人順序中的第一個合法戰鬥格」只屬 `strong inference` 暫時政策。

## Remake 實作邊界

- CoAB 的 `quickAreaSpellTarget(caster, spellID, minRange)` 現統一處理
  Sleep、Fireball、Stinking Cloud、Cloudkill；作品中立 engine 仍只負責
  Quick selector，不知道 CoAB 法術、地圖或劇情資料。
- Stinking Cloud 走 `TACTICALMAP／BuildLegacyAreaScanTargetIDs`，再由正常
  point transaction 呼叫 `CombatCastWithTerrain`；其 2×2 persistent effect、
  poison save、咳嗽／無助與地面生命週期由既有 runtime 負責，不在本輪複製一套
  Quick 專用規則。
- Cloudkill 先以 line-terrain 的合法敵人格建立 point，再以同一個
  `BeginPendingPointSpellAction` 保存施法者、法術、中心與延遲；中斷、3×3 區域、
  4 HD 直接死亡與 slot transaction 都由既有手動施法路徑處理。
- Fireball、Stinking Cloud、Cloudkill 的範圍中心仍是 bounded adapter policy；
  不得把目前穩定 fighter order、測試 seed 或現有 renderer timing 寫成原版
  exact。Lightning Bolt 尚未由 Quick area adapter 接通，仍應維持 fail-closed／
  unsupported 邊界。

## 驗證

- `TestCombatAltMQuickStinkingCloudUsesAreaCenterAndPersistentArea` 使用正式
  game-pack、ALT+M／Quick、12×12 `TACTICALMAP`、有位置敵人與持續 PRNG；確認
  Quick 產生 `stinking_cloud` area visual、正確中心、單一 2×2 persistent area、
  impact 與法術格消耗。
- `TestCombatAltMQuickCloudkillUsesAreaCenterAndPendingDelay` 使用正式
  game-pack、line-terrain、7 級施法者與 4 HD 敵人；先確認 point action 保存
  中心與非零 delay、法術格尚未消耗，再交回正常 scheduler，確認 `cloudkill`
  area visual、低 HD 直接死亡與最後 slot transaction。
- 同一 focused Docker gate 亦通過 Quick Sleep、Quick Fireball、Quick Bless
  pending、手動 Stinking Cloud、手動 Cloudkill 與 Fireball 多目標回歸。
- Docker／Xvfb 正式 gate
  `go test -count=1 -p 2 ./cmd/... ./gamepack ./internal/...` 通過，
  `go run ./cmd/coab-audit -root .` 回報 `total=0`；本輪 marker 為
  `ROUND495_FORMAL_EXIT=0`。

## 後續缺口

1. 以 PC-98／DOSBox 固定戰場關閉 Quick target linked-list 的排序、`1..7`
   random helper 與完整 candidate safety predicate。
2. 完成 Lightning Bolt line target、所有 Quick 法術的中斷／存檔／延遲演出，
   以及弓箭、法術飛行、命中、死亡、音效與回合節奏的原版動態對照。
3. 完成全敵方 AI、完整 ECL 玩家路徑、全部地圖／文字／音樂音效與跨平台發行；
   本規格不能支撐「完整 remake 已完成」的聲明。
