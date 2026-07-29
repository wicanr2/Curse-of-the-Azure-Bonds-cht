# 第三百五十四輪：戰鬥動作時間軸

狀態：`IN PROGRESS`（共用時間軸與第一批 renderer adapter 已完成；原版
projectile 素材／逐影片 timing 尚待驗證）

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
- `-combat-visual-demo melee|bow|magic|kill` 可輸出四張固定 phase 的
  640×480 Xvfb frame。
- screenshot oracle 會凍結在指定的 timeline elapsed，不受 Ebiten／Xvfb
  啟動時間影響；弓箭 travel checkpoint 因而能穩定顯示箭身與箭頭。
- CPIC／COMSPR／SPRIT 遺失的 masked-blit alpha 由不透明 top-left chroma key
  恢復；骷髏與動畫不再顯示成整塊底色。

目前弓箭與 Magic Missile 是明確標示的 renderer fallback，不是原版 projectile
pixel-exact 素材。完成本規格仍需 DOSBox／影片時間碼、原始 projectile block
定位，以及將 fallback 替換為 title asset-pack 定義。
