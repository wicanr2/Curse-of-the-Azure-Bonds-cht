# 1236 — 戰鬥雲霧暫態生命週期與燃燒之手 QUICK 資料

狀態：`READY`（新遭遇清除雲霧失能、同場快照保留、燃燒之手原版 QUICK 欄位）

## 問題與根因

一般強度按鍵路線在 `ECL5/0x33` 黑暗精靈領主戰中，尚未施放惡臭之雲便出現
「因噁心而動彈不得」。`Battle` 的持續區域沒有跨場重用；真正滲漏的是
`syncPartyFromBattle()` 將整個 `combat.Fighter` 投影回持久 `s.party`，連同
明確標為戰鬥內、以行動次數消耗的 `CoughingTurns`／`HelplessTurns`。下一場
`StartCombat()` 又原樣載入，讓上一場未消耗完的失能進入新遭遇。

領主戰解鎖後，路線於 `ECL1/0x50` 的鷹馬戰觸發另一個失敗即關閉錯誤：法師
記憶四格燃燒之手（全域 spell ID 9），但 `combat_ai_spells` 從 7 跳到 15。
原版受版控屬性表 `DS:37DAh` 第 9 筆已提供完整欄位：`casting_time=1`、
`priority=2`、`cast_on=1`、`min_range=0`；缺的是 game-pack 宣告，不是 selector
行為。

## 修正

- `StartCombat()` 在**新遭遇入口**清除雙方 `CoughingTurns`／`HelplessTurns`。
- `combat.RestoreBattle()` 不走該入口；同一場戰鬥的存檔仍保存剩餘失能與雲區。
- `gamepack/pack/00-core.json` 依原版第 9 筆補上燃燒之手 QUICK metadata。
- 保留 spec 1213 的失敗即關閉：其他未宣告法術仍收回玩家控制，不靜默跳過。

## 驗證

- `TestStartCombatClearsPreviousEncounterCloudIncapacitation`：新遭遇雙方失能歸零。
- `TestBattleSnapshotKeepsSameEncounterCloudIncapacitation`：同場快照保留 1／4 回合。
- `TestEmbeddedPackValidatesAndOwnsZhentilText`：燃燒之手五欄逐項對表。
- 無強化 `TestKeysDriveARealSessionFromTheTitle`，`COAB_KEY_FRAMES=5000`：通過；
  600 格、6 種畫面、218 句、5 個 ECL 段、67 次快速戰鬥操作、全滅 0、原文
  fallback 0。先前第 1,936 幀的領主戰全滅與第 4,744 幀的 metadata 死環均消失。
  第 5,000 幀停在阿沙本福德服務選單輪巡，屬重放路線知識，不是產品停格。

後續 6,500 幀探針讓驗證器在整備完成後走可見的「離開／繼續旅程」，已推進到
`0x51`；首次走到原作拼成 `PRETEND TO BE MERCENAIES` 的選項並由 fallback 硬閘門
抓出。game-pack 現以原始拼字作匹配鍵，繁中顯示「假扮傭兵」，英文 catalog
校正顯示拼字；雙語 key 覆蓋測試通過。該探針隨後在匕首瀑布城內／城外循環，
未納入上述 5,000 幀通過基線。

這份長跑是 remake 正常路線證據，不宣稱原版逐回合戰術一致，也不宣稱一般強度
已由標題連續通關。
