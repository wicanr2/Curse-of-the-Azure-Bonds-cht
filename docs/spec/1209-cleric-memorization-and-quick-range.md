# 1209 — 牧師記憶候選、休息亂數與 QUICK 近戰距離

狀態：`READY`（限一級牧師候選、休息嘗試亂數生命週期與 QUICK 近戰移動）

## 已證實差異與修正

- 原作法術表與 spec 809 證實一級牧師有可記憶容量；`combat_player_spells`
  亦列出 1..8 八支一級牧師法術。新建牧師的 DOS `KnownSpells`／法術書旗標為空，
  remake 先前卻把它當成牧師候選清單，造成 `MEMORIZE` 只有「完成／取消」。現在
  牧師依原作法術表取得其可施放環級的完整職業候選；法師仍受法術書限制。
- 休息中斷先前每次都以同一 `eclSeed` 重建 RNG，第一次中斷後重試必然重演。
  現在休息嘗試使用獨立、可由 `SetECLSeed` 重播且逐次推進的 deterministic stream。
- QUICK 把玩家角色委派給 AI 後，只有敵方會在不在武器距離內時接近目標；玩家
  QUICK 角色因此可能隔空近戰。現在兩方共用相同的距離檢查與接近動作。

## 正常玩家路徑控制組

未撐隊伍的按鍵重放已走完：建立戰士／牧師／法師／遊俠／聖武士／盜賊各一名，
保留個人資金並用正式選單把六人訓練到二級；逐人購買、合法整備皮甲、盾牌與
職業可用武器，接著進入 CAMP → MAGIC → MEMORIZE，牧師可選八支法術並
暫存「祝福」，再選 REST。開場區域的 ECL 將休息遭遇設成每 1 小時、100%，所以
這次正常休息必定中斷，祝福不會寫回；重放不得用無限重試繞過這條規則。

後續原版角色樣本證明，這個「二級隊伍」控制組並不是原作完成建角後的正常起始
狀態：`BOB.GUY` 已是 25,000 XP、五級戰士、5 HD。舊結論把 NEWCHAR 中途的
「職業槽先寫 1」誤當成最終存檔等級，因而製造出不合理的皇家衛兵敗戰。訂正與
正常玩家路徑證據見 spec 1210；本規格只保留牧師記憶、休息亂數與 QUICK 距離結論。

## 回歸

- `TestCampMagicNewClericCanMemorizeFromClericalList`
- `TestRestInterruptionAdvancesItsDeterministicStream`
- `TestCombatQuickPartyApproachesBeforeMakingAMeleeAttack`
- `TestKeysDriveARealSessionFromTheTitle`（`COAB_KEY_BOOST=0` 聚焦重放）

## 尚未宣稱

- 不再把一級／二級測試隊伍的敗戰解讀成原版正常起始難度；最終起始等級見 spec 1210。
- 沒有把開場強制休息遭遇改成安全休息，也沒有用測試旗標灌入法術槽。
- 法師建角後法術書旗標在此路徑仍需另立窄任務追蹤；本輪不以牧師規則掩蓋法師資料。
