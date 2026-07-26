# 第 135 輪：combat Cure Light Wounds

狀態：`READY`（限一級牧師戰鬥治療）

## 證據與流程

RuleBook 將 Cure Light Wounds 列為 `Both`，並明確記載治療 1–8 HP。本輪接入：輪到有已記憶 spell ID `3` 的牧師時，按 `H` 選擇第一位受傷隊員，消耗一個 slot，套用 seeded 1d8 並封頂 MaxHP，之後沿用既有敵方回合流程。

## 邊界

目前沒有完整 target cursor、施法動畫、saving throw（此法術不需要）、施法中斷或其他治療法術；target selection 以 roster／fighter 穩定排序的第一位受傷隊員作 bounded adapter。其他 spell IDs 不共用此 healing path。
