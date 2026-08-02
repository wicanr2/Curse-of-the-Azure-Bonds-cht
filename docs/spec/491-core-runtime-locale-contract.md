# 第 491 輪：核心互動執行期文字資料契約

狀態：`READY`

## 範圍

本輪資料化 WHO 選角、ECL PICTURE、戰鬥錯誤 fallback、紮營角色 HP 列、旅店
恢復、未知法術、ECL 字串輸入 prompt 與遊戲時間。技術性 ECL editor error
改為英文識別錯誤，不冒充玩家翻譯。

## 契約

- `selected_character`、`select_character`、`character_hp_choice`、`game_time` 與
  `event_picture` 由正式 catalog 提供。
- 未知法術一律經 `spell_unknown`，不在呼叫端另造「法術 0xNN」。
- WHO、camp target 等控制流程仍使用角色 ID／index；HP 顯示文字不參與選擇。
- 遊戲時間仍由 typed clock fields 組成，locale 只決定格式。

## 驗證

- 正式 catalog coverage 覆蓋本輪 key；戰鬥錯誤與時間測試改用動態 catalog
  期望值。
- 斷網 Docker：`go test -count=1 ./internal/game`。
- marker：`ROUND491_SAMPLED_EXIT=0`；漢字稽核 `94→77`。

這是資料邊界 milestone，不代表完整遊戲或完整中文化。
