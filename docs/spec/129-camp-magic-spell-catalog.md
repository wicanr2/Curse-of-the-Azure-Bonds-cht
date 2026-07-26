# 第一百二十九輪：CAMP MAGIC bounded spell catalog

狀態：`READY`（限已核對的一級牧師／魔法師 spell names）

## 證據

RuleBook 的 MAGIC 章節區分 memory、spell book 與 scroll，並列出一級牧師與一級 magic-user spell table。既有 DOS player parser 已保存 ordered memorized byte IDs；本輪只把兩個已核對的一級 table 的前八個 one-based IDs 映射到繁中名稱。

牧師前八個順序為 Bless、Curse、Cure Light Wounds、Cause Light Wounds、Detect Magic、Protection from Evil、Protection from Good、Resist Cold；魔法師前八個順序為 Burning Hands、Charm Person、Detect Magic、Enlarge、Reduce、Friends、Magic Missile、Protection from Evil。

## 實作 contract

- `CAMP → MAGIC` 仍是只讀流程，不改變 spell slots。
- 已知 class／ID 顯示繁中法術名；未知 ID 保留 `0xNN`，避免把未解碼的 global spell table 猜成已知法術。
- `Cure Light Wounds` 仍以既有一級牧師 table ID `3` 供 FIX service 使用；catalog 與 cast／memorize rules 分離。
- spell book、MEMORIZE／CAST／SCRIBE mutation、施法消耗、saving throw 與完整高等級 catalog 仍待後續反組譯。

## 驗證

`TestCampMagicLocalizesVerifiedFirstLevelSpellNames` 覆蓋已知牧師 spell names 與未知 ID fallback；既有 `TestCampMenuMagicListsMemorizedSlots` 確認只讀與返回流程未改變。Docker `go test ./...` 已通過。
