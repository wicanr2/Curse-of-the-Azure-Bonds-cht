# 第一百九十八輪：跨 ECL chapter monster records

狀態：`READY`

## 問題

ECL session 已能跨 ECL1–ECL6 block namespace，但 State 原先只保存一張 `MON1CHA` map。當 event 轉到 ECL2 或其他 chapter 時，重複的 monster ID 可能查到錯誤的 record，導致錯誤名字、HP、AC 或 sprite block。

## Contract

- 啟動載入 `MON1CHA.DAX` 到 `MON6CHA.DAX`，每章保存獨立 map。
- State 只依目前 ECL block ID 做 chapter selection；VM 不知道 MON 檔名。
- 目前由 raw image 證實的 namespace mapping：ECL2 `0x00..0x0F`、ECL3 `0x10..0x1F`、ECL4 `0x20..0x2F`、ECL5 `0x30..0x3F`、ECL6 `0x40..0x4F`、ECL1 `0x50` 以上。
- 沒有 chapter table 時保留既有 default record map，維持 synthetic／舊呼叫端相容。

## Regression

`TestMonsterRecordsFollowCurrentECLChapter` 建立 current ECL block `3`，同一 monster ID 分別注入 ECL1／ECL2 records，確認 State 選取 ECL2。先前的 `TestRealECL2EncounterBuildsBattleFromMON2CHA` 則驗證實際 ECL2 encounter 可建立 Battle。

本輪 Docker 通過 `go test ./internal/game ./internal/monster ./internal/ecl`。

## Boundary

這完成的是 record lookup adapter，不等於 ECL2 已從 ECL1 正常玩家流程抵達，也不等於所有 chapter 的 variable operands、遭遇 script 或 MON*ITM／MON*SPC side effects 已完成。
