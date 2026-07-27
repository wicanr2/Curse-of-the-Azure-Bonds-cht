# 第一百九十四輪：真實 ECL PICTURE continuation

狀態：`READY`

## 觀察

真實 `ECL1.DAX` 的 `JOURNEY ON` 路徑會先發出 `PICTURE` request。這個 request 已由 State 轉成繁中事件畫面，因此下一個 player input 必須是事件畫面的 Enter，而不是直接把下一個 ECL menu choice 再送進 `Select`。

## Contract

流程固定為：

`JOURNEY ON` → `ModeEvent / OriginalEvent=PICTURE` → `Continue()` → 恢復荒野／ECL menu → 下一個 choice → `COMBAT` boundary。

`Continue()` 負責清除 `PictureRequested`、`PictureBlock`、BIGPIC 與 scene overlay flags；它不假設 picture 已經完成任何戰鬥或地圖副作用。

## Regression

`TestRealECLJourneyReachesBattleWithLoadedParty` 使用 repo 原始 image 的 `ECL1.DAX`、`MON1CHA.DAX`，先選 JOURNEY ON，確認停在 PICTURE，再呼叫 `Continue()` 模擬 Enter，最後驗證 `ModeEvent / OriginalEvent=COMBAT`。這保護真實 ECL selection sequence 不會因 renderer event boundary 被跳過或重複消費。

Docker 內 non-Ebiten internal packages 全部通過；Ebiten command packages 的完整 build 仍需要容器提供 ALSA／X11 開發標頭。

## 跨作品知識

Golden Box 後續作品可沿用此 input boundary，但各作品仍需以自己的 ECL／picture catalog 證明何時 request picture、何時恢復 script；不可由單一 `PICTURE` opcode 推測自動 continuation。
