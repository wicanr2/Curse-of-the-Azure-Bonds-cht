# Gold Box ECL 可重用知識庫

## Entry smoke boundary

每個 ECL DAX block 的五個 initialization entries 是不同的 script entry，不應只測第五個 opening entry。`ecl.SmokeInitializationEntries` 以同一 bounded input sequence 分別執行它們，並保存 per-entry error，適合後續 Pool of Radiance、Secret of the Silver Blades 等 Gold Box 遊戲沿用。

`COMBAT`、`LOAD MONSTER`、`PROGRAM` 與 menu 都是 observable boundary；entry smoke report 出現 signal 只代表 VM 已讀到該 command，不代表 external routine、monster table、party memory 或 renderer side effect 已完成。

## Evidence discipline

當 real entry 在 operand 1 出現 `code 0x01` 時，不能直接把它當 literal monster count。應先反組 `SAVE／memory` operand semantics，再修改 `DecodeMonsterSpawn`／`DecodeMonsterSetup`；否則會把 ECL 的 runtime variable 誤解成固定 encounter。

目前 smoke evidence 已找到 ECL2 block 3 entry 3 的兩個 monster spawn 與 COMBAT，這是下一個 encounter vertical slice 的候選來源，但仍需用 `MON2CHA` 與完整 input／memory context 驗證後才能接入正常玩家流程。
