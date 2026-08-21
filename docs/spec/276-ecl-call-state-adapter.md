# 第二百七十六輪：ECL CALL State／map adapter

狀態：READY

## Reference mapping

`ovr003.CMD_Call` 先以 unsigned word 計算 `operand - 0x7FFF`。原始 CoAB ECL image
觀察到的三個 raw operand 對應如下：

| raw ECL operand | dispatch value | reference routine |
|---|---:|---|
| `0x2E10` | `0xAE11` | 重算目前格的特殊碼；畫面髒了才重畫視窗、地點列與牆面碼（spec 1150）|
| `0xC01E` | `0x401F` | `ECL2.MOVEFORWARD` |
| `0xB200` | `0x3201` | 依 `DS:8B4Ch`（＝ ECL 格 `03DE`）播 10 或 11 號音效 |

`MovePositionForward` 只接受方向 0/2/4/6，將 16×16 map coordinate 前移並 wrap；
routine 本身沒有 collision check，之後才重算 wall/roof。因此 ECL forced movement
不得誤用玩家按鍵的 door／wall blocking transaction。

## Remake transaction

- `State.applyECLCallSignals` 依序消費 `RunResult.CallAddresses`。
- `0xC01E` 立即更新 persisted `DungeonX/Y`，四邊均有 wrap regression。
- `0x2E10` 只有在本次 session 從頭到尾未跨 block，且同 block 的
  `SAVE`／`SAVE TABLE` trace 證明 CALL 前新寫 `C04D` 作為 facing commit，
  才把同批實際新寫的 `C04B／C04C／C04D` 欄位投影至 State；未寫欄位維持
  原值。frontend 再以 one-shot request 重建 dungeon floor、wall stamps、
  wall type 與 roof state。這涵蓋完整與部分同 block 傳送，又不會把跨
  `NEWECL` 流程或無方向對話 scratch registers 誤當成玩家目的地。
- `0xC01E` 由 State 先完成 forced move，再由 frontend request 重繪。
- `0xB200` 播 reference 的 10 號音效（Step）。全 corpus 對 `03DE` 只有 15 次
  `SAVE 05 03DE`、沒有讀取，所以 11 號那一支走不到（spec 1150）。
- 未知 CALL 仍保留 ordered request，不猜 side effect。
- 一般 ECL menu flow 與 combat continuation 使用同一 State adapter。

## Real-image evidence

ECL3 block 16 graph 在 `+0x0193` 與 `+0x157B` 可達 `CALL 0x2E10`。real regression
由 entry `+0x0198` 執行 bounded loop，確認唯一 CALL 為 `0x2E10`，並可 exactly-once
轉入 State request。

ECL6 block `42h` terrain `08h` 在 `+1084h..+1090h` 直接寫
`C04B=0Bh／C04C=0Ah／C04D=02h`，隨後於 `+1096h` `CALL 2E10h`。
第 402 輪正常玩家路徑證明救援灌木誘餌後必須移到 `(11,10,S)`；獨立
State regression 鎖定這項同區塊 register projection，另以跨區塊交易回歸
證明舊 block 的 CALL／SAVE trace 不得覆蓋新 block 的出生點。

ECL6 block `42h` terrain `0Bh` 又在 `+13CFh` 只寫 `C04B=0Ah`、
`+13D5h` 寫 `C04D=0`，保留目前 `C04C` 後於 `+13DBh` 呼叫 `2E10h`。
第 403 輪因此改為 field-selective projection。ECL1 Filani 對話的反例會在
同一 CALL 前清 `C04B/C04C`，但不寫 `C04D`；回歸鎖定它不得移動玩家。

## Remaining boundary

`0x2E10` 的髒旗標在 remake 由立即 one-shot redraw 取代。原作是
`ECL2.STOREVALUE` 收到 `C04B`／`C04C`／`C04D` 就**當場**寫
`DS:720Fh`／`7210h`／`7211h` 並把 `8B68h` 設 1，`CALL` 只負責「髒了就重畫」；
本節這個「同 block、且要有新寫的 `C04D`」的判準是 remake 這一側的啟發式，由
第 402／403 輪的正常玩家路徑回歸鎖著。兩種模型在「對話把 `C04B`／`C04C` 當
暫存」時結果不同，要換成原作模型必須先解釋那些回歸為什麼過得去（spec 1150）。

```text
go test ./internal/game ./internal/ecl ./internal/party
```
