# 第四十輪：遊戲內冒險手札與 CAMP state

狀態：READY（繁中資料呈現與 CAMP 控制邊界）

## 本輪證據

- 新增 `docs/manual/curse-of-the-azure-bonds-zh-TW.md`，整理故事起點、操作、地點、戰鬥提示與研究來源。
- 新增 `docs/history.md`，保存 SSI／Gold Box／本專案重製方向的繁中歷史筆記。
- `assets/locale/zh-TW.json` 新增冒險手札與 CAMP 繁中字串。
- `game.State.OpenJournal/CloseJournal` 與 Ebiten `J`／`Esc` input 接通；戰鬥中禁止開啟手札。
- ECL `PROGRAM 9` 現在呼叫 `State.Camp`，顯示紮營事件並可由 Enter 返回荒野；`CampCount` 與原始 `PROGRAM 9` marker 有 regression。

## 邊界與未完成項目

- CAMP 的實際 HP／法術恢復、遭遇中斷、守夜與 memory side effects 尚未還原。
- 手札目前是第一篇背景與遊玩摘要；完整 Adventurer's Journal 條目、謎語／翻譯輪與所有劇情提示仍待整理。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./...
```
