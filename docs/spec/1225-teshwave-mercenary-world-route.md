# 1225：匕首瀑布至特什維夫的傭兵分支

狀態：`READY`（2026-08-27）

## 問題

spec 1224 後的最大未走叢集是 `ECL1/0x51:0x1004` 起 158 條。原版
世界路線表顯示，從匕首瀑布選擇特什維夫後走 `WILDERNESS`，會計算出
`4C9D=13` 並進入這支。事件先詢問是否調查向北小徑，再分成
直接攻擊或假扮傭兵；假扮成功後還有四道指令。

## 實作與證據

1. 測試由原版抵達 entry 進入匕首瀑布，後續只經
   `JOURNEY ON → TESHWAVE → WILDERNESS`；沒有直接寫 `4C9D` 或分派索引。
2. 六條玩家分支各用全新 state：忽略小徑、直接攻擊巡邏，以及假扮
   傭兵後的攻擊首領、通報匕首瀑布、進軍匕首瀑布、進攻特什維夫。
3. 結果斷言使用既有 stable ID：`teshwave.reach-dagger-falls`、
   `teshwave.monsters-infight`、`teshwave.sweep-down-melee`、
   `teshwave.army-melts-away` 與 `teshwave.both-routed`。
4. 戰鬥使用原版 `MON1CHA` 資料；強化單人只是內容盤點的有界工具，
   不是一般強度玩家隊伍的難度驗收。

## 結果

- `ECL1/0x51`：289／726（39.8%）→ **381／726（52.5%）**。
- 全作實跑：9,132／14,177（64.4%）→ **9,235／14,177（65.1%）**。
- 圖外執行：0。
- 原最大 `0x9004` 叢集已消失；新最大叢集從 `0x95B4` 起，68 條。
- 文字稽核維持 1,022 頁、unmatched 0；移除一條會污染後續 run 的
  短暫訊息規則後，matched 1,003、variable-insert 15、subroutine 4。

## 驗證與邊界

Docker、無網路環境：

- `TestRealDaggerFallsToTeshwaveMercenaryBranch` 六個子測試通過。
- `go test ./internal/game -run 'TestReal|TestTilvertonRoute' -count=1`：通過。
- 連續按鍵 session：通過。
- `cmd/ecl-run-coverage`：9,235／14,177，圖外 0。
- `cmd/ecl-text-coverage`：unmatched 0。

本規格證明六個事件結果由正常世界選單可達；不證明一般強度隊伍能打贏
這些戰鬥，也不把 65.1% 當作整體 remake 完成度。
