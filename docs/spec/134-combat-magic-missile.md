# 第 134 輪：combat Magic Missile

狀態：`READY`（限一級魔法師 Magic Missile）

## 證據與流程

RuleBook 明確記載 Magic Missile 無 saving throw，每枚造成 2–5 HP；1–2 級 1 枚，之後每兩級增加一枚，最高 11 級 6 枚。本輪加入 platform-neutral Battle spell result、seeded damage、State slot consumption 與 Ebiten `S` 操作：目前輪到魔法師時，選定敵人並施放已記憶的 spell ID `7`。

## 邊界

目前只實作 Magic Missile；spell slot 由 game state 消耗，效果由 combat core 套用。其他法術、saving throw／狀態效果、施法中斷與完整 action menu 仍待各自的反組譯證據，不把本輪效果擴張到其他 spell IDs。
