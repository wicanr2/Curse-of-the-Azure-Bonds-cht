# 第 130 輪：Known Spells preservation

狀態：`READY`（限 DOS player record → party/save → CAMP MAGIC 資料保存）

## 證據與邊界

已反組譯／解析的 DOS player spell record 同時有兩組資料：

- memorized spell IDs：目前已記憶、可供現有 `SpellSlots` 顯示與窄 FIX 流程使用。
- known-spell flags：角色已學會的法術，不等於目前已記憶的欄位。

本輪將 `KnownSpells` 從 DOS parser 投影到 `party.Character`，加入 JSON save 欄位，並在 `CAMP → MAGIC` 顯示「已記憶／可用」數量。未知 ID 仍保留資料，不猜測其名稱或規則。

## 尚未宣稱完成

這不是完整的 MEMORIZE、CAST、SCRIBE、法術消耗、職業等級限制或休息後重置。後續 Golden Box 遊戲可沿用此資料分層，只替換各作品已驗證的 spell catalog 與 rules adapter。
