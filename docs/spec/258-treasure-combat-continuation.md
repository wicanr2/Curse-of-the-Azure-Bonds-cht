# 第 258 輪：TREASURE／COMBAT continuation 順序

狀態：`READY`（限同一 ECL result 的 loot、戰鬥與勝利後恢復順序）

## 已完成

- 同一段 ECL result 同時發出 `TREASURE` 與 `COMBAT` 時，不會先進 loot menu 而跳過戰鬥。
- 有實際 monster spawns 時，deterministic／random treasure 以 raw request 掛在
  encounter 上，不在戰鬥前解析或顯示。
- party victory 後才解析 encounter reward 並開啟 loot menu；收取完畢或選擇
  「暫不收下／繼續」後，以 `treasureResumeECL` 從 COMBAT 下一條 ECL
  instruction 恢復 PRINT／MENU／NEWECL。這避免跨多個戰後 PICTURE pause 的
  劇情讓 pending loot 永久滯留。
- headless adapter 尚未載入 ITEM DAX 時保留 raw request，仍讓 COMBAT continuation 繼續。

## 回歸

現有 real ECL JOURNEY→COMBAT regression 維持通過；火刀首領 State regression
確認勝利 loot 後可依序跑完手札 54／53、夢境、`NEWECL 0x50`，最後恢復
Tilverton edge menu。
