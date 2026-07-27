# 第一百九十五輪：ECL spell signal state bridge

狀態：`READY`

## 問題

bounded ECL runner 已能解碼 `SPELL` 與 `PROTECTION`，`party.Roster` 也已有 ordered spell lookup；但 State selection path 原先只處理 map、picture、combat 與 program signals，會直接丟掉法術 request。

## Contract

- `State` 在每次 ECL selection result 後保存 `SpellSearches` 與 `ProtectionRequests`。
- queue 保留跨 picture／menu pause 的 request 順序。
- `ConsumeSpellSearches()` 與 `ConsumeProtectionRequests()` 會複製並清空各自 queue，保證一次性傳遞。
- State 不寫入 ECL runtime memory、不消耗 party slot，也不由位址猜測 spell effect；這些由作品專屬 adapter 處理。

## Regression

`TestECLSpellSignalsTransferToStateOnce` 驗證 spell ID、slot address、character address 與 protection address 完整保留，並確認第二次 consume 為空。ECL runner 本身的 signal decode 由 `TestRunSubsetEmitsSpellAndProtectionSignals` 覆蓋。

本輪 Docker 通過 `go test ./internal/game ./internal/ecl`。

## 跨作品重用

後續 Golden Box 遊戲可以重用 pending queue 與 ordered resolver 的邊界，但必須依各作品的 player record、ECL memory layout 與 rules 證據實作 slot 消耗／效果；不能把 queue 存在視為完整 spell engine。
