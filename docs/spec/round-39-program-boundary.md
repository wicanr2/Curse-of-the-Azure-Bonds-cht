# 第三十九輪：PROGRAM 外部 routine 邊界

狀態：READY（VM 控制轉移邊界；State side effects 見第 275 輪）

## 本輪證據

- 參考重寫程式的 `CMD_Program` 顯示 `PROGRAM 0/3/8/9` 會執行 engine-level routine，並結束目前 ECL VM pass；`PROGRAM 9` 會進入紮營流程。
- `ecl.RunResult` 保留 observed `ProgramIDs`，並以 `ProgramExit` 表示 bounded runner 在外部 routine 邊界停止。
- 真實 ECL1 block `0x51` initial flow，從 payload `+0x014B` 以 selection `2`（CAMP）執行 34 steps 後在 `PROGRAM 9` 的 `+0x05DB` 停止；不再錯誤地重複跑場所 menu。
- synthetic regression 驗證 `PROGRAM 9` 只消耗自身指令並回傳 external boundary。

## 後續進度與邊界

- 第 40、117 輪已將 `PROGRAM 9` 接入 CAMP state；第 275 輪再將
  `PROGRAM 0/3/8` 接入 start menu、party-killed 與 game-won transaction。
- DOS 原版直接結束 process 的部分，在桌面重製版被明確轉譯為返回標題；這是 frontend
  policy，不宣稱與 DOS process lifecycle 相同。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./internal/ecl ./internal/game
go run ./cmd/azure-bonds -block-id 81 -run-subset -run-start 331 -select 2
```
