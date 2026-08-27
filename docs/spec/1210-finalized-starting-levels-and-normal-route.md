# 1210 — 最終起始等級、累計 HP 與正常路線訂正

狀態：`READY`（限完成建角後的起始等級、HP 投影與按鍵重放分支）

## 原版證據

DOS 原版樣本 `docs/reference/original-dos/save-samples/BOB.GUY` 的 SHA-256 是
`2fec4c3f0373e9d03a59b312fde75072f9f279d09b81aaa0e56170084769710e`。既有
角色匯入器解出：25,000 XP、300 PP、戰士槽等級 5、`HitDice=5`、HP 46、
`BaseMaxHP=36`、`AttackAbility=44`。這直接推翻「完成建角後仍是一級」的舊解讀。

spec 1101 的職業槽寫入 1 是 NEWCHAR 的中途狀態；最終角色還會依原作起始 XP
套用同一份訓練門檻。remake 現在於快速／引導建角完成時，將各有效職業槽提升到
25,000 XP 支援的等級：戰士、牧師、魔法師、遊俠、聖武士為 5，盜賊為 6。

## 實作界線

- `RollCreationHitPoints` 保留中途一級 helper 的語意。
- `RollStartingHitPoints` 依最終職業等級逐級累計 hit dice、體質加值、遊俠首級
  骰數與封頂後固定 HP；同種子可重播。
- 快速建角的 HP 種子是 remake 的可重現近似，不宣稱逐骰等同原版 RNG。
- 法師升級時以第一個合法候選補入法術書，是快速建角的確定性替代；不宣稱等同
  原版互動選法術的每一步。

## 正常按鍵路徑

未使用 boost、由標題建立六職隊伍並透過正式商店／紮營選單整備後，皇家衛兵戰
以 QUICK 在第 243 幀取勝：五名隊員仍可戰、合計 139 HP，敵人歸零。短程 1,800 幀
回歸在訂正天花板支線後走過 350 格、5 種畫面、131 句玩家文字、6 次 QUICK，
全滅重開 0、落回原文 0。

完整 12,000 幀基線走過 549 格、4 個 ECL 段、141 句玩家文字、6 次 QUICK，
全滅重開 0、落回原文 0；最後仍在 0x03／0x04 間活動，因此尚未宣稱一般強度通關。

追蹤亦證明原路線在下水道「只有盜賊爬得上」事件選「是」後，會進入
「冒險告一段落」標題頁；那不是世界地圖進度。正常通關重放改選「否」。若玩家
仍選「是」，共用角色選擇器會依上一頁明示選盜賊，不再由通用卡住策略亂選戰士。

## 回歸

- `TestQuickCreationFinalizesOriginalStartingLevels`
- `TestStartingHitPointsAccumulatesFinalFighterLevels`
- `TestStartingHitPointsLevelOneMatchesCreationHelper`
- `TestStartingHitPointsIsDeterministic`
- `TestKeysDriveARealSessionFromTheTitle`（`COAB_KEY_BOOST=0`）

## 尚未宣稱

- 原版 BOB 的 46 HP 是單一存檔觀測，不是所有戰士都必須得到的固定值。
- 1,800 幀只證明正常開場戰鬥與連續路線不再被錯誤結束支線切斷，不等於完整通關。
