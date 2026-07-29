# 第三百五十四輪：戰鬥動作時間軸

狀態：`IN PROGRESS`（共用時間軸與原版 arrow／generic spell projectile
已接入；完整逐影片 timing 與其他法術尚待驗證）

## 問題

現有規則會在單次輸入內立即完成命中、扣血、聲音與多名敵人的連續回合。
renderer 因而只看得到最終 state。弓箭沒有 projectile，`SoundMissile`
未被呼叫；Magic Missile 沒有 caster／flight／impact；一般受擊沒有動畫。
目前只有死亡骷髏 overlay 有明確時間相位。

## 共用引擎契約

新增 renderer-neutral `CombatVisualEvent`／action timeline：

1. `windup`：actor attack／cast pose；
2. `travel`：近戰接觸、箭矢或法術 projectile；
3. `impact`：hit、miss、save 或 area overlay；
4. `commit`：HP、訊息與對應聲音；
5. `death`：死亡 overlay／corpse；
6. `handoff`：游標、camera 與下一位 actor。

敵方 AI 改成逐 action pump；不可在一個 update callback 同步跑完整串。
作品 JSON／asset adapter 提供 projectile、色盤、素材與 duration，通用 engine
只保存 phase、source／target、路徑與 deterministic ordering。

## 第一階段驗收

- 一種近戰：attack pose、hit/miss、受擊與死亡。
- 一種弓箭：原版 projectile sprite、source-to-target 飛行、
  `SoundMissile` 與 impact。
- Magic Missile：施法姿勢、飛彈、命中與音效。
- deterministic timeline tests。
- DOSBox／公開影片各保存平台、時間碼、輸入、前後 frame 與 confidence；
  至少輸出 melee、bow、Magic Missile、kill 四段短 capture。

Fireball、Lightning Bolt、Cloudkill 等範圍法術要在此共同時間軸通過後再擴充，
不得以同一閃光代替不同法術。

## 2026-07-29 實作 checkpoint

- `combat.VisualEvent`／`FrameAt` 已建立固定 phase ordering 與 deterministic
  duration。
- Ebiten frontend 啟用 timeline 後會鎖住輸入；敵方 AI 每次只播放一個 action，
  handoff 完成才交給下一 actor。
- actor attack pose、箭／Magic Missile travel、impact、HP／文字、聲音、
  九相死亡 overlay 已按 phase 排序；`SoundMissile` 不再是未使用 selector。
- 規則層仍立即計算結果，但致死目標會在 renderer 中保存到 impact/death
  phase，避免攻擊一按下就消失。
- `-combat-visual-demo melee|bow|magic|magic-impact|fireball-travel|
  fireball-impact-1|fireball-impact-2|lightning-target-hit|
  lightning-line-continue|lightning-reflect|kill` 可輸出固定 phase 的 640×480
  Xvfb frame；Fireball oracle 仍經過 roster slot、BeginCombatCast 與
  CombatCast，不直接偽造 visual event。
- screenshot oracle 會凍結在指定的 timeline elapsed，不受 Ebiten／Xvfb
  啟動時間影響；弓箭 travel checkpoint 因而能穩定顯示箭身與箭頭。
- CPIC／COMSPR／SPRIT 遺失的 masked-blit alpha 由不透明 top-left chroma key
  恢復；骷髏與動畫不再顯示成整塊底色。
- COMSPR consumer audit 證實弓箭使用 `0x00/01/02` 與 attack/flip
  directions；Magic Missile 經共同施法入口使用 `0x05/0x85` 四格 travel，
  傷害 feedback 使用 `0x0A/0x8A` 四格 impact。renderer fallback 已移除。
- `-combat-visual-demo magic-impact` 新增固定命中 phase；五張 checkpoint
  都直接使用原始 DOS asset。
- DOS 公開影片 `wwYsij1wDC4` 的 `00:42:25.20–25.40` 逐格顯示
  Stinking Cloud 共用的青色 spell projectile，約 `25.50` 清除、`25.60`
  才出現文字與雲格；與 `0x05/0x85` consumer 路徑交叉吻合。
- 共用 engine `combat_visuals` schema 現可驗證 trigger／phase、逐格 frame
  及完整八方向 frame；CoAB game pack 宣告 arrow、Magic Missile travel
  與 magic-damage impact 的 COMSPR source、block、flip、scale 及原始
  `SysDelay` 參考值。frontend 不再保存這三組 title-specific block table。
- JSON 驅動後重新由 Docker／Xvfb 擷取 bow、magic、magic-impact 三張
  640×480 checkpoint，並由 frontend 與 game-pack regression 直接驗證
  圖片 key 是從 pack frame 產生。
- DOS 影片 `00:36:15.40–17.00` 證明 Fireball 是一次 generic spell
  travel，抵達後依序對每名範圍目標播放紅白 impact、傷害文字與 optional
  death；不是單一中心爆點或同步扣血。game pack 已加入 fireball
  travel／impact frame mapping。
- `VisualEvent.Impacts` 現以作品中立 contract 保存多目標順序；每名目標有
  獨立 impact／commit／optional death，renderer 會保留尚未輪到死亡的
  resolved combatant。單體箭與 Magic Missile 仍由 legacy fields 合成一個
  impact，原時間軸不變。
- Fireball `0x2F` 已接入正常玩家回合：記憶 slot、任意 32×16 combat tile
  cursor、友軍誤傷、一次 caster-level d6 共用 damage、每名 Spell save
  成功減半、slot 消耗與逐目標音畫 handoff。`sub_5F782`、
  `SpellEntry(0x2F)`、`DoSpellCastingWork` 與 `RollSavingThrow` 是規則證據。
- 三張 remake checkpoint 對應 travel、impact[0]、impact[1]，使用原版
  `0x05/0x85` 與 `0x0A/0x8A`。screenshot 在第三個完整 Draw 才讀回，
  避免 offscreen battlefield 尚未 flush 的半幀。
- 第 355 輪另以 `VisualEvent.Segments` 擴充 interleaved line travel；
  Lightning Bolt `0x33` 已有正常 memorized slot／tile cursor、weighted
  14-step path、地城牆反向、large footprint 去重、反彈重複命中、逐目標
  save／damage／impact 與 `sound_8` intent。DOS 影片
  `07:40:22.50–25.60` 證明 target hit 後繼續延伸；完整證據見 spec 355。

目前 projectile pixel source 與方向／frame ordering 已有 code-backed
evidence，但原版 `SysDelay(10/30/70)` 尚不能直接換算 wall-clock。完成本規格
仍需弓箭、Magic Missile、melee、kill 的 DOSBox／公開影片完整時間碼與短
capture，並把逐距離 cadence 寫成經影片驗證的 duration 規則。Fireball
仍缺 terrain-aware path reachability 與原 combatant-array tie order；
Lightning Bolt 的玩家 vertical slice 與影片 oracle 已建立，但牆角／多次反彈
仍缺 DOS runtime 動態畫面；Stinking Cloud／Cloudkill 的 persistent area
effect 仍須獨立建立，不能因共用投射物已出現就視為完整法術動畫。
