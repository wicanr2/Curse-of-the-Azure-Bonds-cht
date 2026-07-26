# 第 132 輪：CAMP MEMORIZE → REST writeback

狀態：`READY`（限 known-spell selection、暫存與 REST 寫回）

## 證據與流程

RuleBook 說明：法術從 `MEMORIZE` 選取後，不會立刻進入記憶；隊伍必須使用 `REST`，休息足夠時間後才完成記憶。本輪實作：

1. `CAMP → MAGIC → MEMORIZE` 選角色。
2. 從 `KnownSpells` 選取／取消法術，並以 bounded capacity 防止超選。
3. 選擇暫存後返回 Magic command menu。
4. `REST → START` 才將 pending selection 寫回 `SpellSlots`。

## 邊界

目前 capacity 對已匯入角色沿用觀察到的 `SpellSlots` 數量；沒有現有 slot 的一級牧師／魔法師使用一個保守的一級預設。完整等級／多職業／各法術等級 slot table、每 spell level 的準備時間、遭遇中斷與部分記憶結果仍待資料與規則層接入；本輪不宣稱完整 AD&D spell engine。
