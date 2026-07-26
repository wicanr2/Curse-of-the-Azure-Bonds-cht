# 第 131 輪：CAMP MAGIC command menu

狀態：`READY`（限已由 RuleBook 證實的 command routing 與 DISPLAY／REST 邊界）

## 證據

CoAB RuleBook 的 Magic Menu 明列：`CAST MEMORIZE SCRIBE DISPLAY REST EXIT`。本輪將 `CAMP → MAGIC` 從直接列角色改為此 command menu；`DISPLAY` 進入既有角色法術摘要，`REST` 回用 CAMP REST service。

## 邊界

`CAST`、`MEMORIZE`、`SCRIBE` 目前只顯示待接入事件，不虛構目標選擇、spell slot capacity、施法效果、抄錄時間或中斷規則。DOS memorized slots、known-spell flags 與已核對的一級名稱仍由既有 adapter 提供。後續 Golden Box 遊戲可沿用 command state，再注入各作品 rules service。
