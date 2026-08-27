# 1226：匕首瀑布與特什維夫的航運事件

狀態：`READY`（2026-08-27）

## 問題

spec 1225 後的最大未走叢集是 `ECL1/0x51:0x15B4` 起 68 條。這不是
城內事件，而是世界旅行的航運計數器：同方向第一次航行會被另一艘船
逼上岸交戰，第二次同方向航行時海盜會在認出隊伍後逃走。兩個方向
使用不同計數器（`4C88`／`4C8A`），所以「搭兩次船」不等於命中第二次
同向分支。

## 實作與證據

1. 從匕首瀑布的原版抵達 entry 開始，以選單走
   `JOURNEY ON → TESHWAVE → BY BOAT`。
2. 第一趟驗證 `hillsfar.boat-forced-ashore`、等待頁、原版怪物戰鬥與
   `hillsfar.captain-impressed`。
3. 以正常選單搭反向船回匕首瀑布；這趟使用另一個計數器，不會
   錯誤觸發順向的第二次事件。
4. 再次正向搭船到特什維夫，驗證 `hillsfar.pirates-flee` 與抵達選單。
5. 戰鬥仍用原版 `MON1CHA`；強化隊員只縮短內容盤點，不是難度證據。

## 結果

- `ECL1/0x51`：381／726（52.5%）→ **430／726（59.2%）**。
- 全作實跑：9,235／14,177（65.1%）→ **9,285／14,177（65.5%）**。
- 圖外執行：0。
- `0x15B4` 叢集已消失；新最大未蓋叢集是 `ECL1/0x50:0x13F9`，84 條。

## 驗證與邊界

Docker、無網路環境：

- `TestRealDaggerFallsTeshwaveBoatEncounters`：通過。
- `go test ./internal/game -run 'TestReal|TestTilvertonRoute' -count=1`：通過。
- 連續按鍵 session：通過。
- `cmd/ecl-run-coverage`：9,285／14,177，圖外 0。
- `cmd/ecl-text-coverage`：1,022 頁、unmatched 0。

本規格證明航運計數器的玩家可見分支與抵達交接；不證明河流上下游文字的
每個路線組合，也不把 65.5% 當作整體 remake 完成度。
