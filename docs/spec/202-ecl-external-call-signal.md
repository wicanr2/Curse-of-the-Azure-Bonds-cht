# 第二百零二輪：ECL external CALL signal

狀態：`READY`

## 原始 image 證據

ECL1–ECL6 的 `0x2D CALL` 線性掃描只觀察到非 code-segment word operand，
主要地址為 `0x2E10`，另有 `0xC01E`／`0xB200`。ECL3 block 16 的實際 entry
在 `CALL 0x2E10` 後立刻進入 `PRINTCLEAR`／`PRINT`／menu；它不是與 `GOTO`／
`GOSUB` 相同的 decoded payload target。

## Bounded contract

- `CALL` 讀取 word address operand，將 address 保存為 `RunResult.CallAddresses`；
- 呼叫視為 synchronous external routine，bounded VM 從下一個 ECL instruction
  繼續，因此不污染 ECL return stack；
- routine-specific side effect 由作品 game adapter 消費；第 276 輪已接入三個 observed
  CoAB address 的 forced movement、redraw request 與 default sound-a transaction；
- 非 word operand 仍回傳明確 error。

這個切片原先只讓 external call 後的中文事件 prefix 繼續執行；第 276 輪已補上
State／frontend adapter。`0xB200` 的音效編號由 `DS:8B4Ch`（＝ ECL 格 `03DE`）決定：
等於 `0Ah` 播 11、其餘播 10。CoAB 全 corpus 對 `03DE` 只有 15 次寫入、值一律是 5，
所以 11 那一支走不到（spec 1150）。
