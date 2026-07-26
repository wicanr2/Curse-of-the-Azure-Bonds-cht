# 第五十三輪：可恢復的 ECL menu runtime

狀態：`READY`（限 bounded menu pause／resume execution context）

## 已確認行為

ECL interactive runner 在缺少下一個 selection 時會停在目前 command，不再從事件入口重跑。可恢復的 context 包含：

- payload program counter；
- `GOSUB`／`ON GOSUB` return stack；
- numeric memory 與 string memory；
- `COMPARE` 產生的六種比較旗標。

`BlockSession` 為每個 decoded block 保存此 context，並保存累積 selection offset。State 以累積 selection sequence 呼叫 session 時，下一次輸入只會交給目前暫停的 menu；既有 memory side effect 與 call stack 會保留。

## 驗證

- synthetic horizontal menu：第一次呼叫停在 menu，第二次呼叫從同一個 PC 接受 selection，接著正常抵達 `EXIT`。
- `RuntimeState` regression 確認 menu destination memory 在 resume 後仍存在。
- `BlockSession` regression 確認兩次呼叫使用 cumulative selections 時不會重播第一個 menu。
- `go test -vet=off ./...` 通過。

## 邊界與未完成項目

本輪沒有宣稱完整 VM：unknown opcode、DOS memory layout、外部 `PROGRAM` routine、跨 ECL block 的完整 memory／call stack transfer，以及 event 完成後的原版 continuation 仍待反組與驗證。此 context 目前以 block 為保存單位；`NEWECL` target 會切換到 target block 的獨立 runtime context。
