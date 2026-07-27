# 第二百五十三輪：CAMP MAGIC CAST／Cure Light Wounds

## 狀態

READY

## Contract

`CAMP → MAGIC → CAST` 現在提供三段繁中選單：

1. 選擇有已記憶法術的牧師／魔法師；
2. 選擇該角色的 memorized slot；
3. 對受傷且未死亡的隊員施放 Cure Light Wounds。

已核對的 Cure Light Wounds 會消耗第一個 matching spell slot，使用 deterministic seed
擲 `1d8`，HP 不超過 MaxHP，並以角色 ID 同步目前 combat fighter。沒有合法目標時不消耗
slot；未知法術只顯示繁中 boundary message，不會誤套 Cure path。

施法完成後回到 `MAGIC` menu，Enter 可繼續操作；這個流程也能讓後續 Gold Box 作品替換
spell catalog／target rule，而沿用 menu、slot transaction 與 roster projection 分層。

## 未完成邊界

SCRIBE、完整非戰鬥法術 catalog、高等級／多職業 slot capacity、施法時間與遭遇中斷仍需
各自的 RuleBook／ECL／DOS evidence；本輪不宣稱完成完整 AD&D spell engine。
