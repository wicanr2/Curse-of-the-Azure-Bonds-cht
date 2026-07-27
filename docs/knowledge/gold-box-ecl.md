# Gold Box ECL 可重用知識庫

## Entry smoke boundary

每個 ECL DAX block 的五個 initialization entries 是不同的 script entry，不應只測第五個 opening entry。`ecl.SmokeInitializationEntries` 以同一 bounded input sequence 分別執行它們，並保存 per-entry error，適合後續 Pool of Radiance、Secret of the Silver Blades 等 Gold Box 遊戲沿用。

`COMBAT`、`LOAD MONSTER`、`PROGRAM` 與 menu 都是 observable boundary；entry smoke report 出現 signal 只代表 VM 已讀到該 command，不代表 external routine、monster table、party memory 或 renderer side effect 已完成。

## Evidence discipline

當 real entry 在 operand 1 出現 `code 0x01` 時，不能直接把它當 literal monster count。應先反組 `SAVE／memory` operand semantics，再修改 `DecodeMonsterSpawn`／`DecodeMonsterSetup`；否則會把 ECL 的 runtime variable 誤解成固定 encounter。

目前 smoke evidence 已找到 ECL2 block 3 entry 3 的兩個 monster spawn 與 COMBAT，這是下一個 encounter vertical slice 的候選來源，但仍需用 `MON2CHA` 與完整 input／memory context 驗證後才能接入正常玩家流程。

## Monster AC adapter

真實 `MON2CHA` 的 `ArmorClass` byte 使用 50..60 的 packed representation：`60 - raw` 才是 signed combat AC；ECL2 FIRE KNIFE raw `59` 因此是 AC `1`。`monster.CombatArmorClass` 只在這個已觀察範圍轉換，較小 synthetic／已解碼值保持原值。這個 adapter 可由後續 Gold Box 遊戲重用，但仍需各作品自己的 record evidence。

ECL2 block 3 entry 3 現在已通過 raw ECL2／MON2CHA → `game.StartEncounter` regression，證明 encounter data bridge 能建立 playable Battle；它仍是 direct entry slice，不宣稱一般玩家流程已自動抵達。

## Chapter-local monster tables

`MON1CHA`–`MON6CHA` 的 monster ID 不能直接合併成一張 map；State 依 observed global ECL namespace 分流：`0x00..0x0F`→ECL2、`0x10..0x1F`→ECL3、`0x20..0x2F`→ECL4、`0x30..0x3F`→ECL5、`0x40..0x4F`→ECL6、`0x50..`→ECL1。這是 loader／State adapter 的責任，不應放進 bounded ECL VM。

若作品的 block namespace 或 MON ID 規則不同，後續 Gold Box 遊戲必須注入自己的 mapping；不能只因 CoAB 的 chapter ranges 相鄰，就把它當成通用 DOS 常數。
