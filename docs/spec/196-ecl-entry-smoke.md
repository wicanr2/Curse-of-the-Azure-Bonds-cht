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
- 第二百七十三輪重跑時，ECL1–ECL6 共 25 blocks／125 entries 已全數抵達正常 bounded boundary，沒有 unsupported opcode；早期 `0x2D` CALL、`0x2F` AND 與 variable monster operands 的停止點均已由後續 evidence-backed semantics 取代。

## Boundary

這是 corpus prefix gate，不是完整事件文字對齊或完整玩家流程；menu 的所有選擇、random branches、PROGRAM side effects 與跨事件持久狀態仍需各自 real-flow regression。
