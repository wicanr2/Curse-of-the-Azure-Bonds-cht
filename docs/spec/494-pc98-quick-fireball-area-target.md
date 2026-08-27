# 第四百九十四輪：PC-98 Quick Fireball 範圍落點與延遲交接

狀態：`READY`（只關閉 CoAB 的 Quick `Fireball (2Fh)` bounded slice；其他
Quick Area 法術與原版候選排序仍未完成）

## 本輪結論

第 493 輪已把 PC-98 overlay 09 的非零 `MinRange` 掃描流程接到 Quick
`Sleep`。本輪沿用同一份非破壞性證據，接通已有完整 CoAB 範圍效果、法術格
交易與視覺時間軸的 `Fireball (2Fh)`：

1. Quick selector 抽到 Fireball 時，adapter 以 game-pack 的 `min_range=3`
   和目前 `TACTICALMAP`，逐一以存活敵人的戰鬥格建立 `SCAN` bounded candidate。
2. 沒有地形投影、敵人格或合法 `SCAN` candidate 時，Fireball 仍回報
   unsuitable／error，不改成近戰或另一個法術。
3. 合法中心保存到同一個 point target transaction。Fireball 的 raw
   `CastingTime=03h` 經 `/3` 形成非零 pending action，先保留法術格與中心；
   scheduler 重新選到該角色後，才進既有 Fireball 的逐目標 travel／impact／
   damage／death visual pipeline 並消耗法術格。

這是可玩的 Quick Fireball 戰鬥進度，不代表 PC-98 的 target linked-list、
`1..7` random helper 或所有 Quick Area 法術已逐指令還原。Fireball 的
area-safety predicate 已由 spec 1239 接通；其他範圍法術仍不在此聲明內。

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

## 原始控制流與法術資料

- overlay 09 local `02D3h..03D0h` 建立非零 `MinRange` 的 SCAN candidate，
  依 actor record 的 `+198h` 準備方向／arc 參數，走目前戰場 object list，
  比較候選隊伍與角色狀態，再讀 spell record 的 `+61BCh/+61BDh` 並呼叫
  效果／豁免 predicate。這些 bytes 與 branch 是 `exact control flow`；
  offset 的完整欄位名稱仍只標為 `strong inference`。
- local `03D3h..04C9h` 讀 priority、`CastOn`、`MinRange`，只有非零
  `MinRange` 才進入上述 helper。`Fireball (2Fh)` 的 game-pack record
  保存 `min_range=3`、`casting_time=3`；raw table 的欄位位置與 `/3`
  規則由 READY spec 425 保存，資料值本身為 `exact`。
- local `04CCh..0627h` 由角色 record 的 target pointer chain 建立候選，
  再以候選位置重呼叫 suitability；local `072Bh..0754h` 將選定 spell
  交給 `CASTCOMBATSPELL`。候選 linked-list 的完整 tie／random 行為仍是
  `unknown`，不可用 stable fighter order 宣稱逐指令相同。

## Remake 邊界

- CoAB adapter 將原本 Sleep 專用 helper 抽成
  `quickAreaSpellTarget(caster, spellID, minRange)`，但 engine
  `combat/quickspell` 仍只保存 title-neutral selector，不認識 Fireball、
  CoAB 地圖或角色資料。
- 目前只有 Sleep 與 Fireball 使用這個非零 `MinRange` bounded predicate；
  Lightning Bolt、Stinking Cloud、Cloudkill 仍 fail-closed。中心選擇保留
  stable fighter order，這是 `strong inference` 的暫時 adapter policy。
- Fireball 仍重用既有 `combatCastFireball`：法術區域的半徑、saving throw、
  火焰抗性、逐目標 visual impact、傷害文字與死亡 overlay 不在本輪重寫。
  Quick 只負責把原版候選中心安全地交給同一條正式 pipeline。
- 原始 runtime 尚未提供本輪 Quick Fireball 固定戰場的逐鍵／逐幀 oracle，
  因此 palette、sound cadence、target tie 與所有 NPC Quick 行為仍不得標為
  `exact`。友軍範圍安全的原版控制流與 remake 回歸另見 spec 1239。

## 驗證

- `TestCombatAltMQuickFireballUsesAreaCenterAndPendingDelay` 使用正式 CoAB
  game-pack、ALT+M／Quick、11×1 `TACTICALMAP`、遠離施法者的敵人與持續
  combat PRNG。測試先確認 Fireball action 保留 point 與 pending delay、
  法術格尚未消耗，再以 `CombatManualControl`／正常 scheduler 續跑到
  Fireball area visual，確認中心、敵方 impact 與最後 slot consumption。
- 同一 focused Docker gate 另外通過 Quick Sleep、Quick Bless pending 與
  手動 Fireball 多目標回歸；這些只支持受影響的 combat slice，不支持完整遊戲
  可通關聲明。Docker／Xvfb 正式套件 gate
  `go test -count=1 -p 2 ./cmd/... ./gamepack ./internal/...` 及
  `go run ./cmd/coab-audit -root .` 亦通過，marker 為
  `ROUND494_FORMAL_EXIT=0`、`total=0`。

## 後續缺口

1. 以 PC-98／DOSBox 固定戰場關閉 Quick target linked-list 的排序與
   `1..7` random helper。
2. 逐一閉合 Lightning Bolt 的 line target、
   Stinking Cloud／Cloudkill 的 persistent area、save／中斷與延遲演出。
3. 完成全敵方 AI、完整 ECL 玩家路徑與 DOS／PC-98／remake 的動態戰鬥逐幀
   對照。
