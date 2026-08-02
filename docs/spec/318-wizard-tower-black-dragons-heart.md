# 第三百一十八輪：法師塔黑龍戰與龍心

狀態：`READY`

## 反組譯證據

- ECL5 block `0x33 +0x0452` 的普通四項 vertical menu：
  - index `0` `ATTACK DRAGONS` → `+0x04B9`；
  - index `2` `FLEE` → 同一 `+0x04B9`，所以此處無法逃離；
  - 規格 317 的 hostile PARLAY 也在 `+0x0602` GOTO 此處。
- `+0x04B9` 設定 PICTURE／monster `0x35`，龍群判定隊伍自證有罪；
  `+0x04F5 LOAD MONSTER 0x35,14,0x35` 後於 `+0x04FC COMBAT`。
- 戰後 `+0x04FD` 比較 engine work byte `7EC7 > 0x80` 時回到
  `+0x04EE` 重建同一戰鬥；否則 `+0x0508` 跳到 `+0x0676`。
  本輪保存 raw gate，不把 `7EC7` 猜成未證實的士氣、逃跑或戰況名稱。
- `+0x0676` 表示德拉坎德羅斯已在戰鬥中逃下樓，黑龍屍體散落屋頂。
  若 `4C61 != 1`，直接跳到安全屋頂；只有 `4C61==1` 才於 `+0x06C7`
  詢問是否取一顆龍心。
- 選 NO 直接到安全屋頂；選 YES 後：
  - `+0x06EC..+0x074D` 描述剖開黑龍時遭酸液淋濕，但成功取出心臟；
  - `+0x0752 DAMAGE 0xC0,3,4,3,1`：全隊承受 `3d4+3`、save type 1；
  - `+0x075D SAVE 1,[4C64]`；
  - `+0x0768` 顯示可安全守住房頂休息，PRINT RETURN 後 EXIT 回地城。

## 引擎與中文契約

- ATTACK、FLEE 與 hostile PARLAY 是三個來源、同一 script target；不得由
  State 看到 FLEE label 就提前顯示成功撤退。
- COMBAT continuation 必須保留 `7EC7` gate、`4C61` 條件、YES/NO selection、
  pending DAMAGE 與 `4C64` write，不能在勝利時直接跳到通用結果頁。
- `DAMAGE flags 0xC0` 是 whole-party saving-throw packet；這類不需 WHO／
  random target 的 request 應在 ECL continuation 自動解析。不能只自動處理
  `0xE0` no-save packets，否則龍心文字會播放但 HP 不變。
- 同一 `3d4+3` damage roll 依既有 reference adapter套用全隊；每名角色各自
  擲 save type 1。完整酸抗、物品保存位置與 `4C61` 的上游取得條件仍不可猜測。

## 驗收

- real-image regression 驗證 outer-menu FLEE selection 仍建立
  MON5 `0x35×14`，與 ATTACK DRAGONS 相同。
- State 長流程選 ATTACK DRAGONS，驗證 14 名原版黑龍、戰後屍體文字與
  `4C61==1` 龍心 YES/NO。
- 選 YES 後驗證全隊 HP 確實因 deterministic `3d4+3`／save 流程改變，
  `4C64==1`，再返回安全屋頂與 block `0x33` dungeon。
- 另以 NO 或 `4C61!=1` 證實跳過酸液與 `4C64` write。
- 正式 Ebiten／Xvfb direct-entry 使用 Area 5 CPIC namespace，產生
  [`wizard-tower-black-dragons.png`](../screenshots/wizard-tower-black-dragons.png)；
  640×480 畫面可見 14 隻原版黑龍小人與 compact 繁中戰鬥 HUD。

更新：第 451 輪已把龍群定罪、龍屍、龍心詢問與酸液取心文字，以及
`ATTACK DRAGONS` 選項移入 CoAB JSON；`4C61`、DAMAGE 與 `4C64` 仍由原 ECL
continuation 驅動。見 [`spec 451`](451-wizard-tower-branches-json.md)。
