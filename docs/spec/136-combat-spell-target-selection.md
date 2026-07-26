# 第 136 輪：combat spell target selection

狀態：`READY`（限已接入兩個法術的 CAST target step）

## 證據與流程

RuleBook 的 CAST 流程是先列出 memorized spells，再指定 target。本輪將 `S／H` 改為先進入 target-selection state：左右鍵切換目標，Enter 確認，Esc 取消；Magic Missile 使用敵方清單，Cure Light Wounds 使用我方清單。未確認前不消耗 spell slot。

## 邊界

目前 target list 只涵蓋 Magic Missile 與 Cure Light Wounds；Cure 的 UI 可選隊友，無 UI 的直接 API 仍以第一位受傷隊員作 deterministic fallback。完整 CAST menu、spell-specific range／area、施法動畫與其他法術 target rules 仍待接入。
