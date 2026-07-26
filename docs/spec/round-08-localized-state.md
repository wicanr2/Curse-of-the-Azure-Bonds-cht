# 第八輪：繁中開場狀態核心

狀態：`DRAFT`

`internal/game.State` 是第一個平台無關的 remake runtime 狀態：

```text
Title --Start--> Wilderness --EnterCity--> Event
                         \--JourneyOn--> Event
```

所有顯示文字透過 `internal/locale.Catalog` 取得，因此繁中 catalog 已真正接入狀態轉移，而不是只有孤立 JSON。

## 邊界

這是已知開場選擇的可測試骨架，不宣稱已還原城市地圖、ECL 條件、戰鬥或存檔。下一步要把 `Event` 的 transition 綁到 ECL trace 的 branch target，並由 Ebiten input adapter 驅動。
