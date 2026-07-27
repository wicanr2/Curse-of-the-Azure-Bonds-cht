# 第二百七十六輪：ECL CALL State／map adapter

狀態：READY

## Reference mapping

`ovr003.CMD_Call` 先以 unsigned word 計算 `operand - 0x7FFF`。原始 CoAB ECL image
觀察到的三個 raw operand 對應如下：

| raw ECL operand | dispatch value | reference routine |
|---|---:|---|
| `0x2E10` | `0xAE11` | 重算目前 wall/roof，必要時 redraw view／時間 HUD |
| `0xC01E` | `0x401F` | `MovePositionForward` |
| `0xB200` | `0x3201` | 依 `word_1EE76` 播放 sound A/B，其他值預設 A |

`MovePositionForward` 只接受方向 0/2/4/6，將 16×16 map coordinate 前移並 wrap；
routine 本身沒有 collision check，之後才重算 wall/roof。因此 ECL forced movement
不得誤用玩家按鍵的 door／wall blocking transaction。

## Remake transaction

- `State.applyECLCallSignals` 依序消費 `RunResult.CallAddresses`。
- `0xC01E` 立即更新 persisted `DungeonX/Y`，四邊均有 wrap regression。
- `0x2E10` 與 `0xC01E` 由 frontend one-shot request 重新建立 dungeon floor、
  wall stamps、wall type 與 roof state。
- `0xB200` 目前播放 reference default sound A（selector 10／Step）。
- 未知 CALL 仍保留 ordered request，不猜 side effect。
- 一般 ECL menu flow 與 combat continuation 使用同一 State adapter。

## Real-image evidence

ECL3 block 16 graph 在 `+0x0193` 與 `+0x157B` 可達 `CALL 0x2E10`。real regression
由 entry `+0x0198` 執行 bounded loop，確認唯一 CALL 為 `0x2E10`，並可 exactly-once
轉入 State request。

## Remaining boundary

`word_1EE76` 尚未投影到 State，因此 `0xB200` 的值 10 → sound B 分支未實作；
`0x2E10` 的 DOS dirty/redraw flags 在 remake 由立即 one-shot redraw 取代。

```text
go test ./internal/game ./internal/ecl ./internal/party
```
