# 第 134 輪：combat Magic Missile

狀態：`SUPERSEDED`（spell identity 由 spec 423 取代；效果邊界仍有效）

## 證據與流程

RuleBook 明確記載 Magic Missile 無 saving throw，每枚造成 2–5 HP；1–2 級 1 枚，之後每兩級增加一枚，最高 11 級 6 枚。本輪加入 platform-neutral Battle spell result、seeded damage、State slot consumption 與 Ebiten `S` 操作。當時把玩家 spell ID 寫成 `7` 是錯誤的 class-local 推論；PC-98 IDA 與原始 spell record 已在 spec 423 證明全域 ID 是 `0Fh`。

## 邊界

目前只實作 Magic Missile；spell slot 由 game state 消耗，效果由 combat core 套用。其他法術、saving throw／狀態效果、施法中斷與完整 action menu 仍待各自的反組譯證據，不把本輪效果擴張到其他 spell IDs。
