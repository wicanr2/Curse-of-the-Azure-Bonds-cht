# 第四百九十六輪：PC-98 Quick Lightning Bolt 折線目標交接

狀態：`READY`（只關閉 CoAB 的 Quick `Lightning Bolt (33h)` bounded line-target
slice；原版候選鏈排序、隨機選擇、牆角反彈與完整 Quick AI 仍未完成）

## 本輪結論

第 493–495 輪已將 PC-98 overlay 09 的 Quick selector 交給 Sleep、Fireball、
Stinking Cloud 與 Cloudkill 的既有效果 pipeline。本輪接通 Lightning Bolt：

1. 正式 game-pack record 的 `spell_id=51 (33h)` 是 `priority=6`、`cast_on=1`、
   `min_range=0`、`casting_time=3`。Quick selector 選中後，adapter 要求目前
   line-terrain projection 可用，從穩定 living-enemy order 找第一個有合法戰鬥格
   的敵人 point。
2. point 以同一個 `BeginPendingPointSpellAction` 保存；`CastingTime/3` 形成非零
   延遲，scheduler 續跑後交給既有 `CastReflectingLineSpell`。因此共同 travel、
   沿線逐格命中、共用傷害池、save／電擊抗性、折線 segment、音效 intent 與
   法術格交易仍由正常 Lightning runtime 負責。
3. 缺少 line terrain、合法敵人位置或 point 時 fail-closed；不改成 Fireball、
   近戰或沒有 visual 的立即傷害。

這是可玩的 Quick Lightning 折線戰鬥進度，不是原版 target pointer projection、
candidate tie／random、牆面反射或所有 Quick 法術的逐指令還原。

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

## 原始控制流與 line evidence

- overlay 09 local `03D3h..04C9h` 讀 priority、`CastOn`、`MinRange`；`33h`
  的 `min_range=0` 不進非零 MinRange SCAN helper。local `04CCh..0627h`
  仍會由角色 record 的 target pointer chain 建立候選，local `072Bh..0754h`
  將選定 spell 交給 `CASTCOMBATSPELL`。dispatch／branch 是 `exact control
  flow`；candidate pointer 到戰鬥格的完整投影與 tie／random 仍是
  `unknown`。
- spec 355 的公開 code-backed reference 與 DOS/EGA 影片已證明玩家 Lightning
  不是單一目標傷害：指定格先結算，電弧沿「指定格－施法者格」方向繼續，途中
  可命中多名角色、離開 footprint 後可再次命中；14 weighted budget、正交 2／
  對角 3 與逐段 travel／impact 順序是 `strong inference`，尚待 PC-98 原始
  bytes 對所有牆角分支閉合。
- `combat_visual_test.go` 的手動玩家 regression 已以既有 runtime 驗證
  `33h`：同一折線可命中 `near`／`far`、保存 `Segments`、電擊保護不造成傷害、
  travel／impact 音效依時間軸出現。這支持「Quick 應交給 line pipeline」，不
  單獨證明 PC-98 Quick 的候選選擇順序。

## Remake 實作邊界

- 新增 `quickLineSpellTarget`，只依目前 `line-terrain` 與有位置的存活敵人
  建立 bounded point；fighter order 由 Battle 的 stable ID order 提供，是
  `strong inference` 暫時政策，不是原版 exact。
- `tryQuickSpell` 只在這個 point predicate 通過時讓 `33h` 成為 Quick candidate。
  選中後以 `BeginCombatCast` 設定同一 point transaction；raw `casting_time=3`
  先走 pending action，並在正常 scheduler 續跑時呼叫
  `CombatCastWithTerrain(LightningBoltSpellID, combatLineTerrain)`。
- engine 沒有新增 CoAB 法術、角色或地圖硬編碼；共用 line spell／visual contract
  不變。Quick adapter 不會自行重算傷害、反彈、save、音效或 segments。
- 完整 PC-98 target pointer projection、area／line safety、牆面反射細節、
  NPC Quick Lightning 與原版 palette／wall-clock 仍保留在後續工作清單。

## 驗證

- `TestCombatAltMQuickLightningBoltUsesLineTargetAndPendingDelay` 使用正式
  game-pack、ALT+M／Quick、12×12 line terrain、兩名同一直線敵人與持續 PRNG；
  先確認 action 保存 line point、非零 delay、法術格未消耗，再以
  `CombatManualControl`／正常 scheduler 續跑到 `lightning_bolt` line visual，
  驗證兩個 ordered impact、segments 與最後 slot consumption。
- 同一 focused Docker gate 亦通過 Quick Sleep、Fireball、Stinking Cloud、
  Cloudkill、Bless pending，以及手動 Lightning／毒雲回歸。
- Docker／Xvfb 正式 gate
  `go test -count=1 -p 2 ./cmd/... ./gamepack ./internal/...` 通過，
  `go run ./cmd/coab-audit -root .` 回報 `total=0`；本輪 marker 為
  `ROUND496_FORMAL_EXIT=0`。

## 後續缺口

1. 以 PC-98／DOSBox 固定戰場關閉 Quick target pointer chain 的排序、
   `1..7` random helper、牆角反射與 line safety。
2. 補齊 Quick spell 的中斷／存檔／延遲演出，並逐項對照弓箭、Lightning、
   Fireball、Magic Missile、持續雲霧的 travel／impact／death／音效節奏。
3. 完成全敵方 AI、完整 ECL 玩家路徑、全部地圖／文字／音樂音效與跨平台發行；
   本規格不能支撐「完整 remake 已完成」的聲明。
