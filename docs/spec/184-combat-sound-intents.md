# 第 184 輪：combat sound intents

狀態：`READY`（限目前已實作 combat actions）

`game.State` 現在以 renderer-neutral `SoundEvent` queue 發送一次性 sound intent：

- `CombatAct` 與敵方回合：每個 `AttackResult` 發出 hit 或 miss；`TargetHP <= 0` 另外發出 death。
- `CombatMove`：移動觸發的攻擊與免費反擊沿用同一組 attack result mapping。
- 已接入的 combat spell 成功施放發出 magic-hit intent。
- `Apply(ActionStart)` 與 wilderness `Move` 也使用同一個 queue；Ebiten adapter 每幀 consume 後轉成 `internal/sound` ID。

這樣的分層讓測試可以驗證音效順序而不需要 audio device，也避免 renderer 從繁中訊息反推規則結果。未實作的背景音樂、MIDI／AdLib、所有 ECL routine sound call 與音量選單仍不在本輪範圍。
