# 第 355 輪：Lightning Bolt 折線、反彈與逐段演出

狀態：`READY`

## 問題

第 354 輪的 `VisualEvent` 可以表達單一路徑 projectile 與依序 impact，但
Lightning Bolt（閃電束）不是「飛到單一目標後結算」：

- 共同施法動畫到達玩家指定格後，電弧仍沿同一方向前進。
- 電弧途中可命中多名角色。
- 地城牆面會讓電弧反向折返，去程與回程可再次命中同一角色。
- 原版把每一段電弧 travel 與該段遇到的 damage/save 文字交錯播放。

若只把它塞進 Fireball 的中心＋impact model，規則、鏡頭與動畫順序都會錯。

## 證據

### 公開 reference（code-backed）

來源：

- <https://github.com/simeonpilgrim/coab/blob/9dc46f1d5911710fb2fcb7a9c0ec0ef74264d17c/engine/ovr023.cs>
- <https://github.com/simeonpilgrim/coab/blob/9dc46f1d5911710fb2fcb7a9c0ec0ef74264d17c/Classes/SteppingPath.cs>
- `ovr023.cs` 的 `sub_5FA44`、`DoElecDamage`、`DoActualElecDamage`、
  `SpellLightningBolt`。
- `Classes/Gbl.cs` 的 `SpellEntry(0x33)`。

可確認：

1. 全域 spell ID 是 `0x33`，magic-user level 3，save 類型為 Spell，
   成功時半傷。
2. `SpellLightningBolt` 只擲一次 `caster-level d6`；同一傷害基數供路徑上
   每名被命中者使用，各自執行 saving throw。
3. 玩家指定格先被結算一次，之後 `sub_5FA44(1, Spell, damage, 7)` 從該格
   沿「指定格－施法者格」方向延伸。
4. `SteppingPath.Step` 的正交步成本是 2、對角步成本是 3；玩家法術的
   初始 budget 是 `7*2=14`。
5. 每遇到新 combatant 就停下一小段、繪製 `draw_missile_attack(0x32, 4,
   current, previous)` 並結算傷害，再從該格繼續。
6. 連續 footprint 內仍是同一 player index 時不重複傷害；離開後若因反彈
   再次穿過，可以再次命中。
7. 只有 `inDungeon==1`、背景 `move_cost==0xFF` 且尚未處於反彈狀態時，
   該格使方向反轉；`groundTile==0` 則終止。
8. electrical path 載入 missile icon `0x13`。既有 COMSPR audit 對應
   DAX blocks `0x06/0x86`。

公開 reference 是可讀重建，不能單獨證明原始組語的每個控制流細節；上述規則
在本輪先標為 **code-backed strong inference**，並以 runtime 影片交叉驗證可見
行為。後續仍需用 `GAME.OVR` bytes／IDA 為高風險邊界補地址證據。

### DOS 遊戲影片（video-backed）

影片：<https://www.youtube.com/watch?v=wwYsij1wDC4>，DOS/EGA。

`07:40:22.50–25.60` 的逐幀順序：

1. `22.50`：青白電弧朝指定 dark elf mage 前進。
2. `22.70`：電弧與 dark elf mage 重疊。
3. 約 `23.10`：`MagicAttackDisplay(showMagicStars=false)` 的紅白 impact
   覆蓋目標，同時顯示 `STAN TAKES 10 POINTS OF DAMAGE FROM ELECTRICITY`；
   reference 證實它載入 icon `0x17`，即 COMSPR `0x0A/0x8A`，每格
   `SysDelay(70)`，並播放 `sound_3`。
4. 約 `23.60–24.50`：顯示 saving throw 結果。
5. `24.70`：同一電弧由命中格繼續向右前進。
6. `25.30`：電弧繼續到更遠格，之後才交棒給下一角色。

保存關鍵幀：

- `docs/reference/original-dos/combat-lightning-target-travel-074022500.png`
- `docs/reference/original-dos/combat-lightning-target-hit-074022700.png`
- `docs/reference/original-dos/combat-lightning-damage-impact-074023100.png`
- `docs/reference/original-dos/combat-lightning-line-continue-074024700.png`
- `docs/reference/original-dos/combat-lightning-line-end-074025300.png`

這段影片直接證明「travel 到目標 → damage/save → 從目標繼續 travel」，
但沒有發生牆面反彈；反彈目前仍由 executable/reference 與官方規則說明支撐。

## 作品中立契約

### 規則

共用 combat package 新增：

- `LineTerrain func(x, y int) LineCell`
- `LineCell{Valid, Reflect}`
- `TraceReflectingLine(origin, target, weightedBudget, terrain)`
- `Battle.CastReflectingLineSpell(...)`

契約只描述：

- Bresenham-like 格線步進；
- 正交 2／對角 3 的 weighted budget；
- 無效格終止；
- 可反射格把方向反向；
- 完整 fighter footprint 的 encounter order；
- 離開 footprint 後才允許同一 fighter 再次命中。

共用引擎不得知道 CoAB spell ID、`DUNGCOM`、`move_cost=0xFF` 或
`COMSPR 0x13`。CoAB adapter 負責把目前戰鬥地形轉成 `LineCell`。

### 視覺

`VisualEvent` 新增 ordered `Segments`。每段保存：

- `From`／`To`
- 可選 `ImpactIndex`

時間軸依序播放：

`windup → common travel → segment travel → optional impact/commit/death → … → handoff`

Fireball、弓箭與 Magic Missile 未提供 `Segments` 時維持既有行為。作品素材
仍由 JSON `combat_visuals` 宣告：

- `lightning_bolt/travel`：共同施法 projectile `05/85`
- `lightning_bolt/line`：電弧 `06/86`
- `lightning_bolt/impact`：一般 magic damage burst `0A/8A`

共同施法入口會送出原始 `sound_8` intent；目前 recovered PC resource table
沒有對應 WAV，因此 sound catalog 保留 selector 但不虛構音檔。每次 damage
impact 另使用已有 `sound_3`／`magic_hit.wav`。

## CoAB 接線

- 新增 `LightningBoltSpellID = 0x33` 與繁中顯示「閃電束」。
- 正常玩家路徑要求角色確實 memorized `0x33`，由戰鬥 CAST 熱鍵進入
  32×16 tile cursor，Enter 後才消耗 slot。
- 玩家指定格不得等同施法者格。
- 傷害與 save 在 `Battle` 一次完成；前端依 segment/impact 順序呈現。
- 地城 adapter：不存在／地圖外為 invalid；`MoveCost==0xFF` 為 reflect。
- 野外 adapter：不存在／地圖外為 invalid，不反彈。

## 驗收

- [ ] 正交、對角、斜率步進與 14-point budget 測試。
- [ ] invalid cell 終止與 dungeon wall 反向測試。
- [ ] 多目標順序、large footprint 單次命中、反彈重複命中測試。
- [ ] 共用傷害骰、逐目標 Spell save 半傷與 slot rollback 測試。
- [ ] 正常玩家 tile cursor 路徑測試。
- [ ] JSON schema／COMSPR `06/86` mapping 測試。
- [ ] deterministic travel、target hit、continued line／reflection 畫面。
- [ ] Docker/Xvfb 正式套件測試與 `git diff --check`。

## 尚未證明

- 原版每段 `GameDelay` 的精確 wall-clock 長度。
- 反彈時 `path_a.steps += 8` 的近距離 self-protection 分支在所有角度的完整
  原始組語語意。
- 牆角、多次反彈與 camera handoff 的 DOS 逐幀畫面。
- 敵方 `throws lightning` 的雙 16d6／range 10 特殊能力已由 spec 415 接線；
  原版目標候選與動態影片時間碼仍未完成。

這些項目不能因本輪完成玩家 Lightning Bolt vertical slice 而寫成「完整法術
系統」或「完整戰鬥」。
