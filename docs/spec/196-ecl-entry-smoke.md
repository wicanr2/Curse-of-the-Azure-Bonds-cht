# 第一百九十六輪：ECL initialization entry smoke analysis

狀態：`READY`

## 目的

`vm_init_ecl` 會載入五個 command-set entry，但原有工具只能列地址或從單一入口 trace。這不足以盤點 ECL1–ECL6 的真實入口，也容易把一個 unsupported routine 誤當成整個 block 無法執行。

## Contract

`ecl.SmokeInitializationEntries` 對每個 block：

- 讀取五個 word-valued entry address，轉成 payload start；
- 以相同 bounded step limit／selection sequence 執行每個入口；
- 每個 entry 個別保存 `RunResult` 與 `error`，不因單一錯誤丟掉其他入口結果；
- 不執行外部 PROGRAM side effect，也不把未知 opcode 當作成功。

CLI 入口：

```text
go run ./cmd/azure-bonds -member ECL1.DAX -entry-smoke
```

## Real image evidence

本輪在原始 image 對 `ECL1.DAX` 至 `ECL6.DAX` 的所有 decoded blocks 執行 smoke：

- ECL1 的 opening entries 可停在 menu；ECL1 block 0x52 entry 4 可安全走到 `ADD NPC` boundary。
- ECL2 block 3 entry 3 觀察到 `COMBAT` 與兩個 `LOAD MONSTER` spawn，提供下一輪真實可玩 encounter 的候選 entry。
- ECL3／ECL4／ECL5／ECL6 入口結果保留了多個 `0x2D`／`0x2F` unsupported opcode，以及 operand 非 literal 的 `LOAD／SETUP MONSTER`，明確指出需要補 operand semantics，而不是猜測敵人資料。

## Boundary

這輪完成的是 evidence and triage tool，不是完整 ECL VM、完整事件文字對齊或玩家流程。後續只有在 opcode／operand 證據收斂後，才能把 smoke entry 接成正式劇情或 Battle entry。
