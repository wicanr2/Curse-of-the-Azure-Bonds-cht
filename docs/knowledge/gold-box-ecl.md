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

ECL1 block `0x50` payload `+0x5B5` 的 `NEWECL 0x03` 已由原始 image regression 證實會切到 ECL2 block `3`。`BlockSession` 應先套用 target，再讓 target entry 自己 bounded stop；target 後的 unsupported opcode 不能回退成 source block，也不能清空共享 runtime context。

## Variable monster descriptors

ECL3／ECL4／ECL6 的 real-image entry smoke 證實，`LOAD MONSTER`／`SETUP MONSTER`
不一定把 ID、數量、sprite 寫成 literal；常見形式是 `code 0x01` memory operand，
由同一段 entry 先用 `SAVE`、`AND`／`OR` 初始化，再在遭遇 command 讀回。bounded
runner 現在用 runtime memory resolve 這些 numeric descriptor，並限制結果在 byte
範圍，避免把未初始化或過大的 word 當成 monster ID。

這個 adapter 讓 real ECL3 block 17／18 與 ECL4 block 33／37 的 smoke entry
抵達 `COMBAT` 並產生 spawn signal；它仍不代表已完成所有 `CALL`、monster table
side effect、party memory 或完整劇情流程。

## External CALL boundary

ECL `0x2D CALL` 的 operand 集合在 ECL1–ECL6 raw image 收斂到非 code-segment
address，主要為 `0x2E10`，另見 `0xC01E`／`0xB200`。它與 `GOTO`／`GOSUB` 的
payload target 不同；real ECL3 opening 在 CALL 後會繼續 `PRINTCLEAR`、文字與
menu。bounded VM 現在保存 `RunResult.CallAddresses`，並從下一個 instruction
繼續，讓後續中文事件可以被觀察；真正 routine 的 DOS memory、UI、sound 或
combat side effect 仍由後續 adapter 實作。

State 現在會把已驗證的 ECL3 Yulash smoke text 經 zh-TW catalog 顯示為繁中訊息，
未知 segment 保留原文，且 raw `RunResult.Text` 不變，方便後續逐段翻譯與對照原版。

後續 raw-image runs 又確認 ECL3 的邪教徒／受傷牧師事件、戰火摧毀的城市片段，
以及 ECL4 的小型魔法商店片段；State 會在 `WaitingForMenu` 前保存合併後訊息，
所以 `PRESS BUTTON OR RETURN TO CONTINUE.` pause 不會吃掉中文事件內容。

ECL3 block 16 entry 4 的 `PRINT RETURN`（`0x33`）現在也會保留 signal 並繼續到
menu；這證實 text-window boundary 與 ECL control-flow boundary 可以分開處理。

ECL5 block 48 的 `LOAD CHARACTER` 後續 inventory sequence 也已被拆成可重用
 signals：`FIND ITEM` 保存查詢 ID，`DESTROY ITEMS` 保存待消耗 ID；real entry
 已從原本 `0x0A`／`0x32`／`0x40` stops 推進到 `NEWECL` boundary。State 現在會把
 `DESTROY ITEMS` 的 verified IDs 套用到 persistent party roster；`Character` 的
 ECL destroy adapter 會刪除所有相同 type 的 item units，包括已 readied record。
 `FIND ITEM` 仍只保存 query signal，因為 compare flag／item namespace 尚未由
 原始 party memory 完整解出。這使 VM signal 與作品專屬 party state mutation
 維持可跨 Gold Box 重用的邊界。

`DAMAGE` 也已建立可跨作品重用的 raw request boundary。公開 CoAB reference 證實
五個 operand 順序是 `flags, dice_count, dice_size, damage_bonus, save_flags`；VM
保存 `DamageRequest` 並繼續 cursor。flags 的 party target／saving throw／random
選擇與 HP mutation 必須由作品 adapter 解讀，不可把它直接當成 combat attack。
