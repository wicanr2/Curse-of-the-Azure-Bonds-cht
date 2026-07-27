# 第二百一十輪：ECL DAMAGE signal

狀態：`READY`（限五個 operand 的 raw signal 與 control-flow continuation）

## 證據

`internal/ecl/operand.go` 的 CoAB command table 將 `0x2E` 定為 `DAMAGE`、五個
operands。公開 CoAB reference 的 `CMD_Damage` 依序讀取：flags、dice count、dice
size、damage bonus、save flags；六個原始 ECL image 的 linear scan 也找到相同
五 operand framing，例如 ECL2 block `0x03` `+0x1599` 的
`0xA0, 1, 6, 1, 0x80`。

## Contract

- bounded VM 將五個可解析 numeric operand 保存為 `RunResult.DamageRequests`。
- VM 必須繼續到下一個 instruction，不在沒有 party／combat context 時自行擲骰或扣 HP。
- State／party adapter 後續必須依作品證據解讀 flags：目標數量／選擇、saving throw、
  damage bonus signedness、random target 與死亡流程不可由本規格臆測。
- 若 operand 是 packed string 或其他不可解析 value，runner 應在該 offset 明確報錯。

## 驗收

`internal/ecl/runtime_test.go` 以 `0x80,1,6,1,0x80` 驗證五欄保存與後續 PRINT；
六個 ECL image 的 scan output 保存於本輪工作紀錄，後續應加入可重現 real-image
prefix／reachable entry regression 後再擴充 party damage mutation。
