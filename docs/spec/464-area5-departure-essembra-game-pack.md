# 第四百六十四輪：Area 5 離場與艾森布拉銜接資料化

狀態：`READY`

## 範圍

本輪把法師塔完成後到下一座城市的七個作品事件移入 CoAB game-pack：

- `area5.depart-akabar`
- `area5.dark-elf-gear-decays`
- `post-wizard.dracolich`
- `essembra.edge／places／branching-oak／outdoor-bar`

每條規則保存原始 ECL source fragments 與 en／zh-TW 訊息；Go State 不再知道
阿卡巴、黑暗精靈裝備、龍巫妖或艾森布拉的中文內容。

## 正常玩家路徑

第 463 輪長路徑在法師塔完成後由 ECL5 block `33h` 選擇離開，接續驗證：

1. 返回 block `30h`，阿卡巴離隊，roster 只剩原玩家角色。
2. 日光使 type `5Eh／60h／61h` 黑暗精靈裝備消失，普通 type `09h` 保留。
3. 返回世界 block `50h`，前往 Essembra trail，觸發龍巫妖復仇並完成戰鬥。
4. 抵達 `essembra.edge`，進城後顯示 places，進出 Branching Oak camp service
   與 outdoor bar，再正確返回同一城市選單與 edge。

這證明文字資料化未破壞 NPC／物品副作用、跨 ECL block handoff、戰鬥
continuation 或設施返回；不只驗證 matcher。

## 驗證與邊界

- 七條規則在 en／zh-TW 均有唯一 rule ID 與非空訊息。
- `gamepack` 與完整 `internal/game` 回歸通過，玩家路徑訊息由 stable ID 取得。
- Go 漢字 literal exact baseline 由 661 降至 655；`localization_debt`
  166→160，frontend 135、runtime 360 不變。七個事件中 outdoor bar 舊分支只
  回傳 locale key，原本沒有漢字 literal，因此事件數與下降數不同。

本輪不代表整個 Essembra、所有設施、酒館傳聞、龍巫妖能力或 Area 5 全部分支
已完成。
