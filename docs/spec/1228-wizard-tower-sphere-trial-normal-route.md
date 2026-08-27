# 1228：巫師塔 Sphere Trial 正常玩家路徑

狀態：`READY`（2026-08-27）

## 問題

spec 1227 後最大的未走叢集是 `ECL5/0x33:0x11AF` 起 148 條。這是巫師塔
地形碼 `0x0C` 的 Sphere Trial：只能由一名法師進房，並以選單反覆集中精神。
正常戰役先前沒有走這條可選房間，因此靜態接線雖已存在，玩家路徑仍未驗證。

## 原版證據

- `ECL5/0x33:0x11AF` 先檢查地形碼 `7F79=0x0C` 與房間旗標 `4C09=0`，
  再顯示試煉告示並詢問是否進入。
- YES 分支以 `WHO "WHO WILL ENTER THE ROOM?"` 選人，並以 `7D00=1`
  限制挑戰者必須具有法師職業；不符者顯示 `THAT ONE CANNOT ENTER`。
- 合格角色進房後，腳本把 X 寫成 8、朝向寫成 2，顯示房間與規則文字；
  選單提供 `CONCENTRATE ON THE SPHERE YOURSELF` 與 `SURRENDER`。
- 成功路徑在 `0x1561..0x163D` 顯示紅袍巫師死亡及寶箱文字、寫入
  block-local `4C09=1`，並開啟寶物。投降會失去金錢與物品；挑戰者死亡則
  `DUMP` 該角色後回到房外。
- `GEO5/0x33` 的地形碼 `0x8C` 位於 `(7,2)`；進房落在 `(8,2)`。勝利後
  玩家由 `(8,2)` 向西穿過單向黑門離開，這是正常移動而非測試傳送。

## 實作

1. 正常戰役 observer 在告示等待頁後選 YES，於真正的 `WHO` 選單中尋找
   第一名具有法師職業的現有隊員，不硬寫角色索引。
2. 進房後依 stable ID 處理規則頁，反覆選擇 `wizard-tower.option.concentrate-on-sphere`。
3. 觀察同一 ECL session 的 `4C09=1`，並要求看見告示、房間、規則、巫師死亡與
   寶箱五組 stable ID。
4. 寶物選單走正常的 `TREASURE_EXIT`；回到地城模式後，由 `(8,2)` 呼叫
   `MoveDungeon(...,-1,0,6)` 向西離房。
5. 修正塔內樓梯走訪器：可選事件若把隊伍移走，不能再套用事件前算出的樓梯一步；
   改由實際落點重新規劃，避免測試工具製造不存在的移動。

## 結果

- `ECL5/0x33`：323／709（45.6%）→ **471／709（66.4%）**。
- 全作實跑：9,329／14,177（65.8%）→ **9,476／14,177（66.8%）**。
- 圖外執行：0。
- `0x11AF` 起 148 條的最大叢集已消失；新最大未走叢集是
  `ECL4/0x23:0x0054`（程式碼位址 `0x8054`），122 條。
- 文字稽核維持 1,022 頁、matched 1,003、unmatched 0、variable-insert 15、
  subroutine 4。

## 驗證與邊界

Docker、無網路環境：

- `TestRealNewGameRunsToTheEnding`：通過，包含 Sphere Trial 五組文字、勝利旗標、
  寶物選單與向西正常離房。
- `go test ./internal/game -run 'TestReal|TestTilvertonRoute' -count=1`：通過。
- `TestKeysDriveARealSessionFromTheTitle`：通過。
- `cmd/ecl-run-coverage`：9,476／14,177，圖外 0。
- `cmd/ecl-text-coverage`：unmatched 0。

本規格證明「法師進房、集中精神、勝利、取寶、走出房間」的合法玩家路徑；不宣稱
投降與挑戰者死亡兩條互斥結果已在同一正常戰役 session 實跑。
