# 450：法師塔主線與手札 15 JSON 化

狀態：`READY`

## 目的

第 315 輪已由 ECL5 block `32h→33h`、GEO5 正常入口與使用者提供的
Adventurer's Journal 證明法師塔開場。本輪不改事件規則，而是移除 State
中同一段故事的作品文字硬編碼，讓原文辨識、英文、繁中與手札內容都由 CoAB
game-pack 的 stable ID 驅動。

## 資料契約

`gamepack/events/pit-of-moander.json` 新增：

- `wizard-tower.courtyard`；
- `wizard-tower.dracandros.arrival／freezes-party／attack-order／journal-15／bond-fades`；
- `wizard-tower.dragon-roof／dragon-steps-out／dragon-illusion`；
- `journal.15.1／journal.15.2`。

庭院有兩種原 ECL 組句，因此使用兩個不同 text-rule ID 指向同一 message ID；
其餘每個 pause 都保存最小必要 `all_contains` oracle。繁中與英文 locale 現各有
252 個相同 stable ID，沒有缺頁或只存在單一語言的 key。

## 遊戲內手札

手札 15 不是外部說明文件。只有 ECL 真正輸出
`FREEZE, BASE SLAYERS OF DRAGONKIND` 與 `JOURNAL ENTRY 15` 後，
`MatchText` 才會回傳兩個 `journal_message_ids`；State 隨即把兩頁加入
`JournalPages`，玩家可在遊戲內重讀。事件前不會預先解鎖，PDF 只作原文與
翻譯證據。

舊 `assets/locale/zh-TW.json` 的九段事件文字與兩頁手札複本已刪除；
`internal/game/state.go` 的九個英文句子 switch 與手札 15 特判也已刪除。
因此修改譯文只需更新 game-pack，不會再同時維護 Go 與第二份 locale JSON。

## 驗證

- `TestWizardTowerDracandrosStoryAndJournalAreGamePackDriven` 逐一以原 ECL
  fragments 驗證十個 text rules、九段繁中、兩頁手札及 en／zh-TW key parity。
- `TestFireKnifeLeaderStateVictoryReturnsToTilverton` 的既有正常玩家 vertical
  slice 會從 GEO5 進塔，驗證手札 15 事件前不存在、事件後恰有兩頁，再走到
  四項原版龍群選單；它沒有直接注入 Journal。
- 正式 Docker gate 與結果 marker 記錄於 `CONTEXT.md`。

## 尚未完成

- 四項龍群選單之後的分支文字仍有部分舊 State fallback，下一個相關
  milestone 應沿相同 stable-ID 模式遷移。
- 其他章節仍存在歷史硬編碼 ECL translation fallback；不可因本輪完成一段
  主線便宣稱全英文文本或 59 則手札已完整繁中化。
