# 第三百一十六輪：法師塔攻擊德拉坎德羅斯

狀態：`READY`

## 反組譯證據

- ECL5 block `0x33 +0x0452` 的四項 vertical menu 以零起算 selection
  `1` 進入 `ATTACK WIZARD` 分支 `+0x050C`；不得把顯示文字當成
  engine-level `COMBAT` action。
- 分支先執行 PICTURE `0x35`。龍群表示這是人類之間的爭端，隨即飛離；
  德拉坎德羅斯則呼叫部隊保護自己，趁一支巡邏隊衝上前時逃下樓梯。
- `+0x05BB` 起的遭遇資料為：
  - `SETUP MONSTER 0x34,0,0x34`；
  - 寫入 `7EC6=100` 後 `CLEAR MONSTERS`；
  - `LOAD MONSTER 0x34` 伊弗利特一名；
  - `LOAD MONSTER 0x31` 黑暗精靈戰士兩名；
  - `LOAD MONSTER 0x32` 黑暗精靈法師一名；
  - `+0x05E2 COMBAT`。
- 戰鬥勝利後同一 resumable ECL 由 `+0x05E3` 跳到 `+0x0768`，顯示隊伍
  已能守住屋頂並安全休息；下一個 PRINT RETURN 後於 `+0x07A2 EXIT`，
  返回 block `0x33` 的塔頂地城探索。

## 引擎與中文契約

- 戰鬥 roster 必須由 MON5 `0x31/0x32/0x34` 原始 records 建立；沿用既有
  「黑暗精靈戰士／黑暗精靈法師／伊弗利特」翻譯與原版 sprite block。
- `CLEAR MONSTERS` 清除的是本次 encounter build list，不得刪除玩家隊伍；
  `SETUP MONSTER` 的 animation/icon request 與 `LOAD MONSTER` 的角色 record
  namespace 必須分開保存。
- COMBAT 是 VM 可恢復 boundary。勝利後不得先跳回 wilderness 或重跑
  block initial entry，必須從原 PC 顯示安全屋頂文字，再由 EXIT 回地城。
- 原圖在 640×480 畫布採 nearest-neighbor 整數放大；繁中敘事使用 24px，
  戰鬥 HUD 使用 compact 16px。

## 驗收

- 從規格 315 的真實四項選單選 index `1`，驗證龍群離去與德拉坎德羅斯逃跑。
- 驗證戰場恰有一名伊弗利特、兩名黑暗精靈戰士、一名黑暗精靈法師，
  且 sprite blocks 分別為 `0x34/0x31/0x32`。
- 以 deterministic battle 完成勝利，驗證安全屋頂繁中訊息，Continue 後回到
  `ModeDungeon`、block `0x33`，而非荒野或 block initial entry。
- `-wizard-tower-battle` 經正式 Ebiten／Xvfb 重現戰場，640×480 畫面保存為
  [`wizard-tower-battle.png`](../screenshots/wizard-tower-battle.png)；中文 HUD
  與原版伊弗利特／黑暗精靈小人同時可見。
