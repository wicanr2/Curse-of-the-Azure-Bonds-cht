# 第一百七十六輪：locked-door menu

狀態：`READY`

## 證據與成果

reference `ovr015.locked_door` 在 detail 2 顯示 Bash、可用時 Pick、可用時 Knock、Exit；detail 3 顯示 Bash、Knock、Exit，Pick 明確 disabled。新增 `internal/dungeon.DoorMenuOptionsFor` 以 raw detail 與 party capability 過濾這三個 action，並將方向鍵撞到 detail 2/3 時的 dungeon preview 接成可操作 menu。

## 邊界

- preview menu 使用 `B/P/K/Escape` 作為目前 input adapter；成功後回到 movement，pick 失敗仍保留可選的其他 action。
- Pick 只在 detail 2 且有已驗證 open-lock skill 時顯示；Knock 只在隊伍有 memorized `0x1F` 時顯示；Bash 在 locked detail 2/3 顯示。
- 這是已驗證 menu capability 與 action dispatch，不宣稱原版完整文字視窗、鍵盤游標樣式、door graphics、時間／傷害副作用或 ECL 劇情 entry。

## Regression

`internal/dungeon/menu_test.go` 覆蓋 detail 2/3 與 party skill／spell capability 過濾；既有 P/K/B transaction tests 驗證 action side effects。
