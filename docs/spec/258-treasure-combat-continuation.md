# 第 258 輪：TREASURE／COMBAT continuation 順序

狀態：`READY`（限同一 ECL result 的 loot、戰鬥與勝利後恢復順序）

## 已完成

- 同一段 ECL result 同時發出 `TREASURE` 與 `COMBAT` 時，不會先進 loot menu 而跳過戰鬥。
- deterministic／random treasure 先解析並保留 pending item，接著建立 combat。
- party victory 後，若 ECL session 還有 continuation，先執行原始 PRINT／MENU／NEWECL；沒有下一個 control-flow boundary 時才開啟繁中 loot menu。
- headless adapter 尚未載入 ITEM DAX 時保留 raw request，仍讓 COMBAT continuation 繼續。

## 回歸

現有 real ECL JOURNEY→COMBAT regression 維持通過；State loot menu regression 也確認 pickup 後回到 event state，不會遺留 wilderness menu。
