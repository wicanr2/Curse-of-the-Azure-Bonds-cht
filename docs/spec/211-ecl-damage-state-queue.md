# 第二百一十一輪：ECL DAMAGE State queue

狀態：`READY`（限 State pending queue／一次性 consume，不宣稱 HP mutation）

## Contract

- `State.Select` 執行 ECL 後，將 `RunResult.DamageRequests` 保存至 State pending queue。
- `ConsumeDamageRequests()` 回傳原始順序並清空 queue，避免跨 event／menu pause 遺失
  script damage request。
- State 不在本輪自行選 party target、擲 damage dice、判 saving throw 或寫回 HP；這些
  必須等 selected-character address、DOS save-throw projection 與作品 rules adapter
  一起驗證。

## Evidence boundary

公開 CoAB `CMD_Damage` 已確認五欄 operand 順序；DOS `saveVerse` 現已保存到
`Character`，但在本規格建立時尚未具備 selected-character memory mapping。固定扣除
第一位角色 HP 會把 adapter fallback 誤稱成原版語意，因此先採與
`SPELL`／`PROTECTION` 相同的 signal-to-State queue 分層；後續窄化 resolver 已另列於
spec 212。

## 驗收

`internal/game/state_test.go` 驗證 request transfer 與 exactly-once consume；下一輪應
以真實 ECL damage entry 接入注入式 target／dice／save resolver，再加入 roster HP
writeback 與戰鬥／死亡 continuation regression。
