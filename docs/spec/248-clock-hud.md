# 第二百四十八輪：reference game clock HUD

## 狀態

READY

## 證據

reference `ovr025.display_map_position_time` 顯示 `time_hour` 與由
`time_minutes_tens * 10 + time_minutes_ones` 組成的 `HH:MM`。`Area1.field_6A00`
將 raw offsets 對應為 `0x18E` ones、`0x190` tens、`0x192` hour、`0x194` day、
`0x196` reference `time_year`，而 `ovr021.timeScales` 對應最後三個欄位的
`24/30/12` 進位。這支持 remake 以 renderer-neutral display view 暴露
小時、十進位分鐘、日、月、年。

## 實作邊界

- `State.GameTimeDisplay` 保留 raw seven-slot clock，不修改存檔格式。
- `State.GameTimeText` 提供繁中 `時間：HH:MM　日期：...`，Ebiten 一般畫面與荒野地圖 HUD 使用同一字串。
- 日／月／年標籤是依 reference field 名稱與進位尺度建立的目前可驗證顯示；完整原版日曆規則與 time-triggered ECL 仍待追蹤。

## 驗證

- `TestGameTimeDisplayUsesReferenceArea1Mapping` 驗證 `5*10+4 = 54` 與 `13:54` 顯示。
- 核心 `internal/game`、`internal/area`、`internal/save`、`internal/ecl` 測試通過。
