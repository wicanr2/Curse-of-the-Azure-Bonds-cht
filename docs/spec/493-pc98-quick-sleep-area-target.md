# 第四百九十三輪：PC-98 Quick Sleep 區域落點交易

狀態：`READY`（只關閉 CoAB 的 Quick `Sleep (15h)` bounded slice；其他
`MinRange>0` 法術仍 fail-closed）

## 本輪結論

PC-98 overlay 09 的 Quick suitability 對 `MinRange` 非零的法術，不是以單一
距離數字判斷。它會先以目前戰鬥位置建立 `SCAN` 候選，逐項檢查候選的隊伍與
角色狀態，再讀全域法術 record 並呼叫效果／豁免 predicate。Quick target helper
另會巡覽目前角色的候選 linked list，對每個候選重新呼叫同一 suitability，最後
把合法候選交給 `CASTCOMBATSPELL`。

本輪只把已經有完整 CoAB 目標、效果、法術格交易與演出的 `Sleep (15h)` 接入：

1. Quick selector 抽到 `Sleep` 時，adapter 以 game-pack 的 `min_range` 取得
   掃描範圍。
2. 以候選敵人的戰鬥格作為原版 area source center，透過目前
   `TACTICALMAP` 建立 `BuildLegacyAreaScanTargetIDs`；沒有合法 SCAN 候選就
   回傳 unsuitable，不能改成物理攻擊或另一個法術。
3. 合法落點保存進同一個 point target transaction。`Sleep` 目前 raw
   `CastingTime=1`，`/3=0`，因此立即進入既有 `CastSleepOrdered`、魔抗、
   `35h` effect、TWINKLE、聲音與法術格消耗；若日後 game-pack record 證明
   非零 delay，仍走同一 `BeginPendingPointSpellAction` handoff。

這是遊戲本體的可玩戰鬥進度，不代表所有 Quick Area 法術或完整敵方 AI 已完成。

## 非破壞性輸入與位址空間

| 輸入 | SHA-256 | 用途 | 推論等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | resident symbols／全域資料 | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay control | `exact` |
| overlay 09 | `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20` | Quick selector／suitability | `exact` |
| overlay 09 IDA report | `3f7f6160e9aac8140ecf320ae45f34572b6d77a39bc7b5f1eb6c8f417c30b36c` | 本輪連續指令輸出 | `exact bytes／control flow` |

IDA Pro 9.4 只讀取 `workplace/ida406/pc98-overlays/overlay-09.bin`，先複製到
`/tmp/coab493-ida` 再分析；報告由既有
`scripts/ida/pc98_quick_magic_audit.idc` 產生。原始 binary、既有 database、
Borland symbol 與位址均未改寫。報告內的 `local=xxxx` 是 overlay-local
位址，不可與 resident linear address 或 file offset 混用。

## 原始控制流證據

- `02D3h..03D0h`：建立 Quick `MinRange` 候選檢查；依 actor record 的
  `+198h` 準備方向／arc 參數，呼叫 resident／overlay 的戰場掃描，逐一走
  `DS:9F30h` 的候選數與 object pointer，先比較候選隊伍，再讀
  `spell_id×10h + 61BCh/61BDh` 並呼叫效果／豁免 predicate。這些 bytes 與
  branch 是 `exact control flow`；`+198h`、`61BCh`、`61BDh` 的完整欄位名稱
  仍保留原始 offset，語意只標為 `strong inference`。
- `03D3h..04C9h`：依 record priority／`CastOn`／`MinRange` 分派；
  `MinRange` 非零才呼叫上述 helper。`Sleep (15h)` 的 game-pack record
  `min_range=1` 與手動 Sleep 已證明的 `SCAN range=1` 相互一致，但 Quick
  helper 的所有 target tie／random 細節尚未由 runtime 關閉，標為
  `strong inference`。
- `04CCh..0627h`：由角色 record 的 target pointer chain 建立候選，檢查
  visibility／狀態欄位，再以候選位置呼叫 suitability；`072Bh..0754h`
  將選定 spell 交給 `CASTCOMBATSPELL`。linked-list 的確切排序與
  `1..7` random helper 的完整用途仍是 `unknown`，不可假稱 remake 已逐指令
  等價。

## Remake 邊界

- `quickSleepAreaTarget` 只屬 CoAB adapter；engine `combat/quickspell` 不認識
  Sleep、CoAB 地圖或角色。
- target center 使用目前 stable fighter order，並保留上游 `SCAN` 回傳的
  object／fighter 順序；沒有位置、完整 IDLIST、地形或合法候選時 fail-closed。
- `tryQuickSpell` 只有 `Sleep` 取得這個 predicate；Fireball、Lightning Bolt、
  Stinking Cloud、Cloudkill 的 Quick Area 仍明確回報 unresolved predicate，
  不會偷偷改成普通攻擊。
- 成功後重用手動 Sleep 的正式 pipeline，所以 effect duration、受傷喚醒、
  自然到期、active combat save 與 TWINKLE 回歸仍由 specs 440–446 管理。

## 驗證

- `TestCombatAltMQuickSleepUsesAreaCenterAndConsumesSlot` 使用正式 CoAB
  game-pack、持續戰鬥 PRNG、ALT+M／Quick、TACTICALMAP 與遠離施法者的敵人；
  只有以敵人格為中心的 SCAN 能產生 `TWINKLE`，並驗證 effect、impact point
  與 slot consumption。
- Docker 斷網 focused gate：
  `TestCombatAltMQuickSleepUsesAreaCenterAndConsumesSlot`、既有
  `TestCombatAltMQuickSleep`／`TestCombatCastSleep` 均通過。
- PC-98 磁片本輪只完成 Docker DOSBox-X 啟動／標題畫面擷取；尚未完成
  Quick Area 戰鬥的原機逐鍵／逐幀 oracle，因此本輪不把 target tie、palette、
  sound timing 或所有 Quick spell 宣稱為 `exact`。

## 後續缺口

1. 用 PC-98／DOSBox 固定戰場關閉 Quick target linked-list 的排序與 random
   `1..7` helper。
2. 逐法術接通 Fireball、Lightning Bolt、Stinking Cloud、Cloudkill 的
   area safety、target pointer、save／中斷與 delayed visual handoff。
3. 完成全敵方 AI、完整 ECL 玩家路徑與原版／remake 動態戰鬥逐幀對照。
