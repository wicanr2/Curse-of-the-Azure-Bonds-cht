# 第 487 輪：建角與世界流程執行期文字資料契約

狀態：`READY`

## 範圍

本輪只處理「建角／讀入角色完成 → 荒野選單 → 世界地圖 → 城鎮／地城入口」
可見文字，不改 ECL、地圖、戰鬥或存檔規則。涵蓋隊伍完成提示、手札框、
一般按鍵提示、撤退結果、關閉遭遇圖片、哈普探索提示、三名 NPC 顯示名、
世界／暗影谷地圖提示、四個地點、城鎮場所選單與動態所在地提示。

## 資料契約

- Go 只保存 stable locale ID；缺鍵 fallback 使用同一 ID，不能再保存繁中副本。
- `place_prompt` 是帶一個所在地參數的正式格式字串；所在地名稱也由同一
  catalog 解析，renderer 不自行拼接中文語序。
- `npc_akabar`、`npc_alias`、`npc_dragonbait` 是顯示投影；NPC chapter／ID
  與角色資料仍是規則身分。
- 荒野選單繼續保留 `ENTER CITY／JOURNEY ON／CAMP` 原始 identity；顯示翻譯
  不參與控制流。

## 驗證

- `world_locale_test.go` 從正式 game pack 載入 catalog，逐鍵確認 23 個 ID
  均有翻譯，並驗證動態場所提示、世界地點投影及荒野選單。
- 撤退回歸以 `FLEE` 原始選項驅動，顯示 choice 刻意替換為 `DISPLAY`；期望
  訊息在測試當下由 `encounter_flee_done` 解析，不複製 JSON 中文。
- 聚焦 `./internal/game` 測試通過。
- 漢字稽核由 `runtime_ui_debt=212` 降為 `173`，本輪移除 39 筆；這不表示
  其餘執行期文字或完整中文化已完成。
