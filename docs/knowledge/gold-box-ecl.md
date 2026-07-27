# Gold Box ECL 可重用知識庫

## Entry smoke boundary

每個 ECL DAX block 的五個 initialization entries 是不同的 script entry，不應只測第五個 opening entry。`ecl.SmokeInitializationEntries` 以同一 bounded input sequence 分別執行它們，並保存 per-entry error，適合後續 Pool of Radiance、Secret of the Silver Blades 等 Gold Box 遊戲沿用。

CoAB 目前的 corpus gate 固定原始 ECL1–ECL6 共 25 blocks／125 initialization entries，要求每個入口不再出現 unsupported-opcode error。這證明「目前可達 prefix」已收斂，不能外推成所有 menu choice／random branch 都覆蓋；其他 Gold Box 作品應以自己的 member／block count 建立同類 gate，而不是沿用 CoAB 的 25／125 常數。

`COMBAT`、`LOAD MONSTER`、`PROGRAM` 與 menu 都是 observable boundary；entry smoke report 出現 signal 只代表 VM 已讀到該 command，不代表 external routine、monster table、party memory 或 renderer side effect 已完成。

初始化 entry 的 block selection屬於 engine lifecycle，不可從內容猜用途。CoAB reference
證明 block `0x52` 是 `inDemo`，正常 new game 是 global block `0x01`；兩者雖都能輸出
青色枷背景文字，party side effect完全不同。fresh new-game reset與 NEWECL continuation
也必須是兩個 API。

## EXIT 是 lifecycle signal

Opcode `0x00 EXIT` 不只是 runner 正常返回。作品 adapter 必須能區分「真的執行 EXIT」、
menu／WHO input pause、步數上限與 unsupported stop。CoAB block `0x01` 的 initial EXIT
會把控制權交回 `sub_29758` world loop；若只看到 `WaitingForMenu=false` 就回到泛用 menu，
會憑空改寫原版流程。共享 VM memory 可提供 script-written registers，但位址的作品語意
應留在 work adapter，不放進共用 ECL interpreter。

同一 block 的五個 entry 不是五個可互換的開場。Reference `vm_init_ecl` 依序讀取
per-turn、SearchLocation、PreCampCheck、CampInterrupted、initial。`RunEntry` 類 API
必須保存 shared memory、把 PC 設到指定 entry 並清空上一 invocation 的 call stack；
menu pause 才是從既有 PC resume。EXIT 也要提交該 invocation 的 memory writes。

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

## Code memory 與 lifecycle ABI

`vm_init_ecl` 依序載入五個 word entry：per-turn、SearchLocation、PreCampCheck、
CampInterrupted、initial。它們是 engine ABI；initial 是 index 4，不能依文字內容猜測。
新的 lifecycle invocation 會從指定 PC 開始，但 ECL／Area／player memory 仍有各自生命週期。

`0x8000..0x9DFF` 是目前 ECL block 的 byte-addressable code memory，不是空白 shared word
map。GETTABLE 可直接把腳本尾端 bytes 當 dispatch table；NEWECL Switch 必須替換這個
window，同時保留 `0x4Bxx` Area、`0x7Axx` party structure、`0x7Cxx` player 等外部區域。
測試若要驗證跨 block shared memory，也不可把 `0x9000` 誤當持久 scratch address。

AND／OR 除了寫回 byte result，還會執行 `compare_variables(result, 0)`。許多 event-bit
utility 隨即使用 IF，沒有額外 COMPARE；遺漏這個 side effect 會讓 opcode corpus 看似
可跑，實際地點事件卻全部被錯誤跳過。

SearchLocation 是 GEO 與 script 的橋接：CoAB 由目前 cell／facing 產生
`C04B..C04F`，ECL 再以 `C04F & 0x7F → GETTABLE → ON GOTO` 派送事件。共用 VM只實作
memory 與 control flow；各作品 adapter 負責座標、half-direction、wall 與 terrain 語意。

SearchLocation 的 plot state 必須留在 resumable ECL memory。Tilverton 城門 `(1,0)`
已證實同一座標第一次只顯示封路警告；在 Weaponers 與 Filani flags 都成立且警告已發生
後，第二次才進入 PICTURE 11／Royal Guard COMBAT。戰後 YES surrender 再經 jail、
PICTURE 2 與 `NEWECL 0x02`，最後由 script 將 map registers 寫成 `(1,12,0)`。這種
「場所條件 → combat engine boundary → 多個 pause → chapter switch」不能拆成 renderer
旗標，也不能在 combat victory 後重建 fresh runtime，否則會遺失投降與章節轉場。

`ROB (0x28)` 是三 operand party mutation，不是單純文字「付款」：scope 為 selected
member／all party；第二 operand 是損失百分比，各 coin 以
`floor(count × (100-loss)/100)` 獨立縮減；第三 operand 是逐 item 的 `1d100` 偷竊率。
reference 在每件 item 判定前依重量累積調低 chance：`weight>24` 減 50、
`weight>255` 減 90，最低為零。Gems／Jewelry 不在 `Money.ScaleAll` 的 coin 範圍。
因此 VM 應發出含原始 scope/percent/chance 的 request，作品 party adapter 才處理 typed
money 與 inventory；不能把 ROB 硬編成固定 GP 費用。

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
 `FIND ITEM` 現已由注入的 party inventory context 解析全隊 raw item type 並設定
 `=`／`<>` compare flags；缺 context 的 trace 仍保留 unresolved signal。這使 VM
 query／working view 與作品專屬 persistent party state mutation
 維持可跨 Gold Box 重用的邊界。

real-image regression 另從 ECL5 block `0x30:+0x0014` 注入含 item type `0x5E` 的 party context，確認第一個 FIND ITEM found branch 會抵達原文含 `SUNLIGHT` 的裝備腐朽事件；這把 synthetic compare test 與真實 ECL operand／branch framing 連在一起。

`DAMAGE` 也已建立可跨作品重用的 raw request boundary。公開 CoAB reference 證實
五個 operand 順序是 `flags, dice_count, dice_size, damage_bonus, save_flags`；VM
保存 `DamageRequest` 並繼續 cursor。DOS player `saveVerse` raw bytes 現在由 party
adapter 保存；selected／whole-party branches 與 random target 可注入 dice／save／hit
resolver 後寫回 roster HP。DOS `field_186` saving bonus 已由 party adapter 保存並
納入 save threshold；State 的 default hit resolver 也會投影 fighter／equipment AC，
並套用已證實的 invisibility `0x19`／`0x47` -4 attack roll 與 action-delay-aware
blink `0x25`。State context variant 可傳入目前 action delay／combat round；displace
`0x59` 會依 FX effect-data 第一 byte 的 `0x10` consumed bit 實作首次 miss 與後續
命中；State resolver 會在 transaction working roster deep-copy effects，避免失敗
request 洩漏 consumed bit。Damage adapter 也保存 reference 的 OK／animated／unconscious／
dying／dead health state；active combat 時 State 會透過 Battle bridge 重新計算
party／enemy win state 並走既有 finish continuation。角色倒下時也會套用 reference
19-kind `RemoveCombatAffects` table；blink／invisibility 依 reference 不在該清單中而
保留。已解出的 Death adapter 另處理 affect_63 recovery、bleeding 與 troll fire/acid
gate；dragon-slayer `0x4B` 已由 explicit dragon target／strength bonus context 投影
1d12×3+4 damage 與 +2 attack bonus。Battle 對倒下 fighter 也會清除
`HasCombatPosition`，對應 reference CombatMap size 0；同一個 Battle bridge 會發出
renderer-neutral `DeathOverlay` signal，讓 Ebiten 在保留死亡座標 anchor 的位置顯示目前
繁中「倒下」overlay。`seg001.Init` 的 mapping 已證實 `combat_icons[24].Attack` 是
`COMSPR 0x8B`、`combat_icons[25].Normal` 是 `COMSPR 0x19`；Ebiten 以 100ms phase
交替顯示兩張 derived sprite。Battle 也將 per-fighter `CombatAction` 的 delay、move、
spell ID、guarding 清零；若倒下者是 State current turn，State 會同步清除施法／移動／
檢視 selection。team party 倒下者另保存 `DownedCorpse=true`，對應原版
`Tile_DownPlayer=0x1F`；Cure Light Wounds 可依 `heal_player` boundary 選到
unconscious／dying corpse，普通治療只清除 skull flash、保留 corpse 與無 position 狀態，
直到後續 combat-heal／placement contract 明確讓角色站起。`NewBattle` 對 save／encounter
初始 HP=0 fighter 也套用相同正規化，因此不會進入 turn 或佔用碰撞格。明確
`CombatHealAllowed` 的 affect_63 recovery 會以死亡時 CombatX/Y 呼叫
`Battle.RestoreCombatant`，清除 `DownedCorpse` 並重新放回 position；普通 Cure 不會站起。
renderer-neutral `DeathOverlayFrame` 現已覆蓋完整 9-cycle timing，Ebiten 九次後 party
轉為 corpse marker、enemy 完全移除名稱／HP render；其他 Death routine 仍保留 boundary。
