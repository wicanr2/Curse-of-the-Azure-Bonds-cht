# 第二百七十五輪：PROGRAM external routine State adapter

狀態：READY

## Reference evidence

CoAB reference `ovr003.CMD_Program` 對四個 engine-level routine 有明確行為：

- `PROGRAM 0` 呼叫 start game menu。
- `PROGRAM 3` 設定 `party_killed` 並離開目前 ECL pass。
- `PROGRAM 8` 設定 `gameWon`、將每位隊員 HP 恢復到 MaxHP、狀態改為 okay，
  進入 start menu，並詢問是否在結束前存檔。
- `PROGRAM 9` 進入 encamp。

## Remake transaction

`State.applyECLProgram` 是選單 ECL 與戰鬥後 ECL continuation 共用的唯一 adapter：

- 0 清除 resumable ECL session 並返回標題。
- 3 保存可查詢的 party-killed flag，顯示全滅畫面，再由玩家返回標題。
- 8 保存 game-won flag，同步恢復 roster／fighter HP、清除 bleeding／downed visual，
  顯示「保存勝利進度／不保存」選單；保存仍經既有 frontend save-request boundary。
- 9 維持既有 CAMP menu transaction，不在進 CAMP 時擅自恢復 HP。

DOS 勝利流程會離開程式；重製版不強制終止 host process，而在相同存檔選擇後返回標題。
這項差異是明示的 frontend lifecycle policy。

## Regression

`internal/game/ecl_program_test.go` 分別鎖定 0、3、8、9；其中 PROGRAM 8 同時驗證
Character health、bleeding、combat Fighter HP／downed visual 與 save request。

```text
go test ./internal/game ./internal/ecl ./internal/party
```
