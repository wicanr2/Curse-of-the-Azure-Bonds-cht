# Gold Box 共用 state／城市 service 知識

## 世界地點值與旅途戰鬥 continuation

CoAB 的 world current-location byte 是作品資料值，不是連續的 remake enum。
目前真實流程已鎖定 Standing Stone=`4`、Essembra=`8`、Hap=`9`；其他 Gold Box
作品應保留自己的原始值與顯示名稱 mapping，不能按旅途順序自行遞增。

旅途上的 `COMBAT` 屬 ECL transaction 中斷點。勝利後必須恢復同一份 runtime，
才能執行抵達地點、寫入 world byte 與顯示 edge menu。複合英文 menu label 也可能
由多個文字片段組成，但仍只消耗一個 selection index；作品 adapter 可正規化顯示
文字，不可改 VM index。

一次性遭遇旗標可能在 `COMBAT` 前寫入。CoAB 的 Hap 黑龍事件以 `4CA2=1`
表示「已觸發接近事件」，不是「已獲勝」；戰鬥結果與續接 ownership 必須另行保存。

世界 dispatcher 的 `NEWECL` 也可能隱含 engine area transition。CoAB 從 Hap
村外進入 global block `0x31` 時，ECL 只發出 `LOAD PIECES 12,FF,FF`；Area 5、
dungeon mode 與對應 graphics namespace 是外層 dispatcher 的責任。共用 VM
不可猜 area，作品 adapter 則不能因缺少 `LOAD FILES` 而沿用上一區域。

SearchLocation 常以 `C04F & 0x7F` 將帶 roof high bit 的 terrain 轉成 dispatch
selector。Hap 的 `0x84` 民宅會寫 visited byte `4C02=1`、顯示 PICTURE 50，
並在同一 runtime 返回探索。`4BC9 > 14` 的 gate 已由原始 ECL 證實，但在
engine work byte 語意確認前不得命名成時間、機率或其他規則。

## 城市場所分層

Gold Box 城市應先由 `ModePlace` 保存 `INN／STORE／BAR／LEAVE` parent menu，再由各場所進入自己的 service state。場所 service 完成後以 `ModeEvent` 保存可翻譯訊息與 `eventReturnMode`，Enter 才回 parent menu；不可把荒野、城市 map 與場所事件混用成同一個返回值。

目前可沿用的 adapter 是：`INN` 的 safe HP restore、`STORE` 的 price-injected transaction、`BAR` 的 ordered `SetBarTales` sequence。各遊戲只需替換城市／ECL data loader，不應複製 renderer 或把原版 routine 的未知副作用猜進共用 state。

CAMP 入口與 REST 必須分離：`PROGRAM 9`／CAMP 只建立 camp command state，不能直接把 HP 設成 MaxHP；REST menu 再以遊戲提供的時間來源計算自然恢復。CoAB 目前採 24 小時一個可測試單位、每位受傷角色 +1 HP，並以 character ID 同步 combat projection。State 現已保存 reference 七-slot clock 並提供 elapsed-minute effect timeout；完整原版日曆規則與 random interruption 仍由 rules／ECL adapter 注入。

外部 `PROGRAM` routine 應集中在作品 State adapter，不應散落在一般選單與戰鬥後
continuation。CoAB 已確認 0=start menu、3=party killed、8=game won／全隊恢復／存檔詢問、
9=encamp。共用 VM 只發布 routine ID；桌面 frontend 可把 DOS 的 process exit 明確映射成
返回標題，同時保留勝利存檔選擇。

Map external-call adapter 要區分玩家移動與 script forced movement：前者檢查 GEO wall／
locked door，後者若 reference routine 只做 coordinate wrap，就必須先忠實移動再 refresh
wall/roof。State 保存座標與 ordered one-shot request，frontend 只使用已載入的 GEO／piece
assets 重建畫面；這能讓 headless rules test 不依賴 renderer。

Spell UI 也應保持三層：DOS／save adapter 保存 ordered slot IDs，verified catalog 只把有證據的 class／ID 映射成名稱，rules engine 才處理 CAST／MEMORIZE／SCRIBE 與消耗。CoAB 目前只核對一級牧師／魔法師前八個 ID；未知 slot 顯示 hex 比猜 global ordinal 更安全，後續 Golden Box 遊戲可替換同一個 catalog adapter。

`KnownSpells` 與 `SpellSlots` 必須分開保存：前者是角色已學會的 spell-book flags，後者是目前已記憶欄位。DOS parser、party Character 與 versioned JSON save 都應保留兩者；UI 只呈現數量與已核對名稱，不能因此推導可施法規則。

RuleBook 的 Magic Menu command 順序是 `CAST／MEMORIZE／SCRIBE／DISPLAY／REST／EXIT`。共用 state 可先實作 command routing；`DISPLAY` 只讀 roster，`REST` 呼叫作品專屬休息 service，而 CAST／MEMORIZE／SCRIBE 必須等待 spell target、capacity、time 與 interruption evidence，不應在 renderer 中猜測。

`MEMORIZE` 應使用兩階段 transaction：known-spell selection 只寫 pending state，`REST_START` 才 commit 到 memorized slots。這樣後續作品可以替換 capacity、每等級準備時間、部分成功與 random interruption，而不必改動 UI 或 DOS／save adapter。

RuleBook 的 preparation timing 可作共用 adapter 輸入：每 spell level 15 分鐘，另加最低準備時間（一、二級 4 小時，三、四級 6 小時，五級 8 小時）。目前 CoAB 只對已核對的一級 catalog 套用 4 小時加每個 spell 15 分鐘並取整小時；不要把這個 bounded rule 擴大成所有等級的完整時鐘模型。

戰鬥施法應拆成三層：game state 驗證並消耗 caster 的 memorized slot，combat core 套用已證實的 spell effect，renderer 只提供 action／target input。Magic Missile 是目前可沿用的例子：spell ID `7`、2–5 damage per missile、無 saving throw；不要用它的 damage path 代替治療、buff、save 或 area spell。

Cure Light Wounds 提供同一分層的 healing 例子：spell ID `3`、牧師 slot、combat core 1d8 封頂 MaxHP。target cursor 與完整 party ordering 尚未解出時，可用 stable roster／fighter order 的第一位受傷隊員作 adapter，但不能把這個 fallback 說成原版 target UI。

CAST target selection 必須是 transaction state：開始選擇時不扣 slot，只有 Enter confirmation 才呼叫 effect adapter；Esc 必須完全取消。攻擊法術與治療法術應各自提供 target list，不能讓 enemy cursor ordinal 意外套到 party target。

Bless 是第一個可沿用的 buff transaction：`B` 開始、Enter 確認、Esc 取消；core 對未與存活怪物八方向相鄰的 party fighter 套用一次 `AttackBonus +1`，保存 `BlessRounds=6`，並以 status flag 防止重複疊加；每次開始戰鬥回合遞減，期滿撤回修正。低階 direct API 缺少位置資料時採「無法判定則不排除」fallback，後續 Golden Box 遊戲可替換更完整的 CombatMap／area adapter。

Curse 沿用同一 transaction 但 target side 相反：`C` 開始敵方目標選擇，Enter 才扣牧師 slot，Esc 完全取消；core 對未與存活 party fighter 八方向相鄰的 enemy 套用 `AttackBonus -1`、`CurseRounds=6`，每次新回合遞減並在期滿恢復。RuleBook 明確說 Curse 目標無 saving throw；range／area 的完整地圖選擇仍由後續 CombatMap／ECL adapter 提供。

Cause Light Wounds 是同一 target transaction 的傷害變體：`W` 開始敵方 touch target selection，Enter 才扣牧師 slot，Esc 取消；core 以 deterministic `1d8` 減少相鄰敵人 HP，無 saving throw，並沿用 Battle 的死亡／勝負 transition。Cure 與 Cause 都是 touch spell，不能共用 Bless／Curse 的 party-wide target fallback；缺少位置資料時才使用 direct API 的 bounded fallback。

Protection from Evil 是 conditional defense transaction：`P` 開始 party touch target selection（施法者可 self-target），Enter 才扣 slot，Esc 取消；fighter 保存 `ProtectedFromEvil` 與 `ProtectionEvilRounds=3×level`。攻擊解析只有在 attacker 明確 `Evil=true` 時才把受防護 target 的 AC 提高 2，不改 base AC。因 `MON*CHA` 尚未提供 alignment 證據，不能把所有怪物猜成 evil；saving throw +2、alignment import 與 dispel 由後續 rules／DOS adapter 注入。

Protection from Good 使用同一防護 transaction，但必須保留職業 spell identity：牧師表的 ID `7` 由 `G` 進入 party target，魔法師表的 ID `7` 由 `S` 進入 Magic Missile enemy target。fighter 保存 `ProtectedFromGood` 與 `ProtectionGoodRounds=3×level`，只有 attacker `Good=true` 時 AC +2；不要把 class-local ordinal 當全域 spell enum。未知 alignment 與 saving throw engine 仍不可猜測。

Encounter non-combat menu 也應保留兩階段 transaction：`FLEE` 直接進入可恢復事件，`PARLAY` 先保存 tactic token，再交由 script／reaction adapter 決定結果。CoAB 目前已接入五個 RuleBook tactic token 的繁中 menu，但不猜速度、追擊、speaker、reaction 或對話結果；後續 Golden Box 遊戲可共用 menu state，只替換 ECL／conversation data。

Combat MOVE 也採 action transaction：`M` 開始、方向鍵確認一格位置、Esc 取消；移入存活敵人格時由 Battle 直接產生一次 attack result，隊友格仍拒絕，正常移動才更新座標並消耗 party turn。地形／邊界／負重／free rear attack／facing 不應在 renderer 猜測，應由後續作品的 CombatMap 與 rules adapter 注入。

MOVE 的 adjacency transition 已可共用：Battle 保存舊座標，成功移動後對「舊位置相鄰、 新位置不相鄰」的存活 enemy 呼叫一次 free attack，再由 State 消耗 party turn。這只證明 trigger 與基本攻擊，不代表已解出背面 AC、facing、武器 reach 或地形規則。
移入敵格則是另一條 MOVE branch：回傳 `MoveResult.Attack` 並保留角色原座標，State 沿用一般攻擊訊息與勝負 transition；不可把敵格當成一般 occupancy error，也不可讓 renderer 自行推測佔格、reach 或 facing。
Combat VIEW 應是獨立 read-only transaction：保存 active party fighter ID，顯示可驗證的 HP／AC／attack summary，Enter／Esc 關閉且不消耗 turn。後續作品可替換 View Menu 欄位，但不能讓 renderer 直接修改戰鬥 state。
武器多次攻擊也應維持三層資料流：`ITEMS` raw RateOfFire（以二倍值保存）→ equipped fighter `AttacksPerTurn` → Battle attack sequence。目標倒下後由 game target adapter 換下一個存活目標；彈藥消耗、職業等級額外攻擊與 Aim／range 不可由單一 RateOfFire byte 臆測。
彈藥再拆成第四層：ITEMS raw `AmmunitionType` 與 inventory item type 是不同 namespace，必須由各遊戲資料層注入 mapping；`Character.ConsumeAmmunition` 在 Battle 前 atomic 扣除本回合 shots，mapping 缺失或不足時拒絕且不修改 inventory。後續 Gold Box 遊戲可重用 transaction，不共享未證實的 type 對應。
Combat DONE 是獨立的 no-attack action：驗證 party turn 後只遞增 turn index 並進入 enemy／next-party adapter，不應呼叫 attack、spell 或 ammo consumption。VIEW／MOVE／CAST 等 pending selection 必須先退出或拒絕 DONE，避免 renderer input 穿透。
MOVE 的格數應由 fighter 的 movement allowance transaction 管理：護甲 table 先給上限，每個 direction input 消耗一格，剩餘格數時不推進 party turn，耗盡才進 enemy／next-party adapter。負重、地形 cost、障礙與 FLEE 速度仍由各遊戲 CombatMap／rules adapter 注入。
Ranged attack 不能只看 `BaseItem.Range != 0`：RuleBook 的 adjacent missile prohibition 有 thrown exception，且 ITEMS raw Range 同時覆蓋兩者。應由 equipment adapter 保存明確 weapon profile；目前只辨識 41–47 missile group 與 dart type 9 exception，Battle 在有座標時共用 guard。

攻擊 transaction 的順序也要跨 Gold Box 共用：先由 `Battle.ValidateAttack` 做不擲骰的 target／range preflight，再扣除本回合彈藥，最後才執行 Attack／AttackSequence。如此無效的相鄰 missile 攻擊不會消耗箭／弩矢，也不會改變 deterministic RNG；直接 `ResolveAttack` 仍可接受注入骰值做規則測試。

`AttacksPerTurn` 不能只在 party UI 套用：enemy turn 若已由 ITEMS／monster adapter 投影出大於 1 的值，也必須使用同一個 `Battle.AttackSequence`；零值或 1 維持單次攻擊。這只修正已知武器 profile 的回合套用，不代表已解出 enemy AI、彈藥或額外職業攻擊。

Monster spell 也沿用 raw-data → fighter → rules 三層：reference `PoolRadPlayer.field_33`
是 0x38 個 spell-list slot，`field_B5..B7` 是 magic-user level-use counts；CoAB 目前只
把已核對的 `0x0F Magic Missile` 接成一級單枚 2–5 damage，並在 enemy turn 成功施放後
消耗一次 level-1 use。其他 raw spell ID 只保存、不猜效果；AI priority、saving throw、
range 與完整 monster spell turn 必須各自取得證據。

`MON*SPC` 的資料層已另外完成：reference `load_mob` 以與 `MON*CHA` 相同的
chapter-local monster ID 讀取九-byte affect records，State 再將其複製到 enemy
fighter 的 `MonsterAffects`。這只證明 raw attachment 與生命週期邊界；隱形、加速、睡眠
等 effect 如何改變戰鬥，仍須逐種對照 reference routine 後才能投影。

目前第一個已核對的 monster effect projection 是 invisibility：reference
`CanHitTarget` 的 `CheckAffectsEffect(Type_16)` 會讓 `0x19`／`0x47` 攻擊骰 -4，combat
core 以目標 AC +4 表示，並只對 active raw records 生效；effect record 本身不被消耗。

Monster 的攻擊次數也不再完全依賴 synthetic default：reference `load_mob` 的
`field_A1` 已解析為 `Record.AttacksPerTurn`，再由 active Haste `0x27` 依
`AffectHaste` 加倍，`AffectSlow`（`0x2A`）減半，並保留目前 adapter 的至少一攻下限。
這仍未包含 movement half-actions、遠程彈藥與完整 `reclac_attacks` weapon profile。

敵方 physical turn 的 target selection 也必須獨立於 renderer。CoAB reference
`find_target` 先以 `BuildNearTargets(0xff, player)` 建立對立隊伍候選，再由 bounded
亂數選一個可見／可達目標；目前 remake 以 sorted fighter ID + Battle seeded RNG
重現「從存活 party 選目標」這一層，並在同一 enemy turn 固定該 target。牆面／pathfinding、
visibility、persistent `Action.target`、AI spell priority 與 guarding 仍由後續 rules／map
adapter 提供，不能把這個候選清單邊界說成完整 monster AI。

玩家輸入造成的 combat error 也應是可恢復 transaction：input adapter 將 `ValidateAttack`／彈藥／target selection 的 error 送到 localized message presenter，保留目前 Mode、turn、HP 與 inventory，不能直接結束 Ebiten game loop。`combat.ErrAdjacentMissileTarget` 可作為跨作品共用的規則錯誤識別；啟動／資料載入錯誤則仍可向上回報。

ECL `ADD NPC` 應採資料 signal boundary，但 CoAB 是兩 operands（ID、morale），不可因
command metadata 的 `1` 誤判只讀一個 operand。作品 adapter 再以目前 area 的
MON*CHA／SPC／ITM 執行 `load_mob`、最低空 icon slot、selected player 與
`(morale >> 1) + NPC_Base`。MON Player record 的 class_id 可能 stale；只有在
ClassLevel 恰有一個非零 slot 時才可於 NPC 專用 parser 推導，不能放寬普通 save importer。

同一 VM pass 可以先 PICTURE 再抵達 COMBAT。State 必須保存後續 combat transaction，
讓玩家看完圖片後開始 battle；若只依 result 欄位固定優先序提前 return，會永久吞掉戰鬥。

CoAB new-game dispatch 必須先分 demo 與玩家流程：`sub_29758` 在 `inDemo` 才選
block `0x52`，正常 `LastEclBlockId==0` 選 global block `0x01`。開始新流程要 fresh-reset
VM memory／PC／selection offsets；`NEWECL` 則必須保存它們。PICTURE 後續不只可能是
COMBAT，也可能是 menu，應保存通用 deferred result。

正式 initial entry 執行 `EXIT` 後要回到 engine world loop。CoAB 在 block `0x01` 先把
map X/Y/direction 寫入 `0xC04B..0xC04D`，因此 State adapter 應在 lifecycle completion
讀出 `(7,13,1)`；其中 direction 是 half-direction，renderer facing 為 `2`。`seg001`
證實 fresh campaign 的 `inDungeon=1`，所以應進入提爾佛頓 GEO1 DungeonMap，不是
wilderness map；不可用 renderer 自造的三選項取代這個 transition。

`CMD_LoadFiles` 的 C# local 命名容易造成 operand 反轉：`var_3=operand 1`、
`var_2=operand 2`、`var_1=operand 3`。Dungeon GEO selector 是 operand 1；outdoor
BIGPIC 判斷是 operand 3。跨作品 adapter 應保存 operand 原始順序，不能依 local 變數尾碼
猜索引。

五個 ECL entry 是獨立的 `RunEclVm` lifecycle invocation。前一次 EXIT 後再執行 per-turn
或 SearchLocation 時，要保留 shared memory、重新設定 PC 並清除 call stack；不能沿用
EXIT 後 PC，也不能 fresh-reset memory。CoAB dungeon 每次成功前進後會同步
`C04B..C04F` 再依序執行 entry 0／1。

Dungeon CAMP 也是 lifecycle transaction，不是 UI shortcut：先跑 PreCampCheck，讓作品
script 寫入 rest encounter period／percentage；只有 rest service 回報 interrupted 才跑
CampInterrupted。CAMP 子畫面必須記住來源 mode，EXIT 才能回到同一 dungeon，而不是回到
泛用 wilderness menu。中斷前已經過的時間可以提交，但不可套用完整休息的 healing／spell
memorization。

地城移動的 reference ordering是「玩家輸入／移動與門處理 → per-turn → SearchLocation」；
每次 invocation 前，作品 adapter 必須從目前 GEO cell／facing 同步 map registers。CoAB
的 `C04D` 是 half-direction，`C04E` 是面前 wall type，`C04F` 是完整 terrain/x2 byte
（包含 high-bit flags），不能只傳座標或把 x2 預先遮罩。SearchLocation 自己決定 mask 與
dispatch table，frontend 不應另做地點 switch。

繁中 remake 不應把 320×240 當作最終排版限制。可重用的視覺策略是把 logical canvas
擴到至少 640×480：原版 tile／sprite／PIC 只以 nearest-neighbor 整數倍放大，保留像素
輪廓；Unicode 中文則在放大後的畫布直接用約 24px 的 TTF／OTF／TTC 重繪。圖片縮放與
文字 rasterization 必須是兩條 pipeline，才不會得到模糊圖片或被 8×8 英文字格限制的中文。
較密集的短選單可使用約 16×15 的中文字級，敘事與標題則以約 24×24 為主；兩者都應
直接在 640×480 層 rasterize，而不是先畫進 320×240 再放大。事件 PICTURE 最好使用
獨立 layout，避免沿用探索 HUD 後讓 3× HEAD／BODY 圖覆蓋日期或提示。

`LOAD PIECES` 先保存三個 selector，再由作品的 file／map adapter 解讀：ECL2 `1,2,3` 與跨 ECL 章節的 repeated observations，加上 reference `LoadWalldef`，已足以把 selectors 接到 `WALLDEF{area}` 的 symbol set 1/2/3 與對應 `8X8D` block。這只證明 raw piece catalog 的載入 boundary，不等於完成 floor／wall／tile renderer；共用 runner 仍可沿用 `LoadPiecesRequested`／`LoadPieces` signal，後續 Gold Box 遊戲替換作品專屬 map adapter。

State 層不能丟掉已驗證的 ECL signal：`LoadPiecesRequested` 應像 `LOAD FILES` 一樣進入一次性 `ConsumeLoadPiecesRequest()`，renderer／map adapter 再決定如何解讀，避免 VM 直接依賴 ZIP 檔名，也保留後續作品替換地圖格式的空間。

## Save 與 dungeon camera

可恢復的 remake save 應把 party／Area state 與 renderer-driving dungeon camera 分開命名。CoAB version 3 的 `dungeon_x`／`dungeon_y`／`dungeon_direction` 保存 16×16 map 座標與八方向 facing；第 280 輪已由正式 new-game ECL 證實它對應原版 party map registers。依 reference `seg001.Init`，舊版 save 或越界值回到 `(7,13,0)`；其他 Gold Box 作品仍需自己的 Area／ECL 證據。

第 181 輪已由 `ovr017.SaveGame/loadSaveGame` 證實 `SAVGAM?.DAT` 固定前綴：`game_area`、Area1／Area2、runtime state、ECL memory、5-byte map state、game states、三組 block/set pair、party count 與 8 筆 `0x29` CHRDAT name records。`internal/save.SAVGAMContainer` 以 raw bytes 保留未知欄位並可 round-trip；其後個別 CHRDAT player files、slot 命名與 file side effects 仍由作品專屬 adapter 處理，不能只因 prefix codec 存在就宣稱完成完整 DOS save。

第 182 輪將這個 prefix 接到 `game.State.LoadSAVGAMPrefix`／`SaveSAVGAMPrefix`：Area codec 更新已知 Area1／Area2 欄位，map segment 更新 signed map position、facing 與 wall cache，未知 memory／records 由 raw container 原樣保留。這個 adapter 可以作為後續 Gold Box 共用入口，但不應從 CHRDAT name record 猜測角色能力、裝備或 spell pointer；那些仍需個別 player record evidence。

第 185 輪依 `ovr017` 的實際命名把 prefix 與 player sidecars 接成 load path：`savgam{a..j}.dat` → `CHRDAT{A..J}{1..6}.sav`，同 basename 的 `.swg`／`.fx` optional。這條路徑重用既有 raw player parser 並能進入中文 remake；寫回 `Player.StructSize`、刪除原始 player 檔與 CAMP multi-file atomic save 仍需獨立 adapter，不能把 load success 當成完整 save compatibility。

第 186 輪新增 `SaveSAVGAMSlot`。它以載入時保留的 raw `.sav` 為底，依已證實 offset patch HP、能力、金幣／寶石／珠寶、icon、memorized／known spells 與 thief skills；`.swg` 以 `0x3F` item encoder、`.fx` 以 9-byte effect encoder 重建，未知 `.sav` bytes 不被清零。prefix 的 party refs 改為 `CHRDAT{slot}{1..6}`，所有檔案先寫入 sibling staging directory，再逐檔替換。這是可重現的 staged writer，不是已完成原版刪檔、跨檔案 atomic commit 或多職業完整序列化；後續 Gold Box 可沿用 raw-preserving patch contract，再替換作品專屬 player offsets。

第 187 輪將這個 writer 接到 Ebiten workflow：`-savgam-dir/-savgam-slot` 載入成功後，F5 與 CAMP SAVE 共用 `saveCurrentGame`，寫回同一個 slot；未載入 SAVGAM 時維持 remake JSON。這個選擇留在 platform adapter，不讓 `State` 依賴鍵盤或檔案路徑，後續 Gold Box 可沿用 workflow contract。

第 188 輪補上縮編清理：`SaveSAVGAMSlot` 只列舉目前 key 的 prefix 與六組 `CHRDAT` sidecar，先移至受限 backup directory，再安裝新 bundle；失敗會移除已安裝項目並還原已備份檔。這解決 party drop/reorder 後舊角色檔污染同一 slot 的問題，但不等於已解碼多職業或全部原始 player serializer。

第 189 輪確認 state ownership：combat `Fighter` 是回合中的 mutable projection，`partyRoster` 才是 CAMP／save 的持久來源。`finishCombat` 透過 fighter ID 將 HP／MaxHP 寫回 roster，但不從 fighter 覆蓋角色名稱、裝備、財寶、法術或 DOS raw-derived fields；這個 raw-preserving sync contract 可供其他 Gold Box 戰鬥 adapter 沿用。

第 190 輪補上 event continuation contract：戰鬥結果按 Enter 不直接沿用上一個 ECL menu，而是重建 `ENTER CITY／JOURNEY ON／CAMP` 的荒野主選單；同一 helper 也供離開城市使用。這保持 renderer/input 與 ECL script 分離，後續完整戰鬥後 script 可替換 eventReturn handler。

第 191 輪把 shop transaction contract 的 SELL 接到繁中 UI。`ItemRecord.Value` 是目前唯一已解碼的 item value source；出售只移除非 readied、非 cursed、正值物品並將 value 放入 party pool，其他 Gold Box 可重用 transaction，但必須注入作品專屬價格／鑑定規則。

第 192 輪把既有 `party.PayIdentifyFee` 接到繁中 Shop Menu。200 GP fee 由 character gold 扣除，raw `HiddenNameFlags` 保留不改；共用 Gold Box adapter 可以沿用 fee／selection transaction，但各作品仍需提供 magic item result table，不能由 type byte 臆測。

第 193 輪補上 `BlockSession` 的跨 `NEWECL` signal aggregation contract。ECL block 是可中斷／可轉移的 bounded VM 執行單位，因此 `LOAD FILES`、`PICTURE`、`SPELL`、`PROTECTION` 不能只留在單一 `RunSubset` 結果；session 會保留一次性 request，合併 spell／protection list，並把資料交給作品的 map／picture／party adapter。這個 signal boundary 可供後續 Gold Box 遊戲沿用，但不代表已完成各作品的資源 lookup 或法術效果。

第 194 輪確認事件畫面是 input boundary，不是自動跳過的文字訊息：真實 ECL1 `JOURNEY ON` 先發出 `PICTURE`，State 停在 `ModeEvent`，Ebiten 的 Enter 對應 `Continue()` 清除 picture state，再恢復荒野／ECL 選擇。後續 Gold Box 遊戲可沿用「request → render → explicit continue → resume」順序，避免畫面尚未閱讀就消耗下一個 ECL menu selection。

第 195 輪把 ECL `SPELL`／`PROTECTION` 從 VM signal 接到 State pending queue。State 只保存原始 spell ID、slot／character address 與 protection address，透過一次性 `ConsumeSpellSearches()`／`ConsumeProtectionRequests()` 交給作品專屬 party／rules adapter；這延續 Gold Box 共用的 signal boundary，不把未知 address 直接寫入 party memory，也不把 request 誤稱為已完成施法。

第 196 輪建立可跨作品重用的 ECL entry smoke contract：`SmokeInitializationEntries` 逐一執行每個 block 的五個 `vm_init_ecl` command-set entry，保留 entry address、bounded PC、steps、menu、COMBAT、spawn、PROGRAM 與 per-entry error。實際 ECL1–ECL6 image 已跑過全 block；ECL2 block 3 entry 3 觀察到兩個 monster spawn 並抵達 COMBAT，其他章節也明確暴露 `0x2D`／`0x2F` 等尚未支援 opcode。這是反組譯證據與 triage 工具，不等於完整劇情或 VM 已完成。

Facing rotation 也應留在 state adapter，不要讓 renderer 自己改 byte：CoAB 目前以 `N,NE,E,SE,S,SW,W,NW` 的 `±2` delta 實作 Q/E 90 度轉向，wall traversal 直接讀 state。其他 Gold Box 遊戲可替換輸入／轉向規則，但應保留 normalized 0..7、wrap 與 save validation。

GEO 座標要把 strict 與 wrapped context 分成兩個 API。reference dungeon `getMap_XXX`／`MovePositionForward` 會在 16×16 邊緣 wrap，但 ECL block 0／10 有 invalid-coordinate 特例；因此 CoAB 的 `CanMoveWrapped`／`TraverseWallViewWrapped` 只由 dungeon preview 呼叫，不能把 wrap 偷渡到 wilderness、Area loader 或所有 ECL block。

原版 map save segment 的 `mapWallType`／`mapWallRoof` 是可重算的 cache，不是另一份地圖真相：前者由目前 facing 的 GEO wall type 得出，後者由 current cell 的 GEO `x2`／`Terrain` 得出。共用 save adapter 可以保存它們以維持 byte-level state，但每個作品仍應在位置／方向變更後重算，不能只信任載入的 cache 來判斷碰撞。

Dungeon background 也應分成「Area palette input」與「GEO roof selection」：CoAB 已解碼 Area1 `0x1FA/0x1FC` 的 outdoor／indoor sky words，並依 `mapWallRoof & 0x80` 選擇 reference sky palette。後續 Gold Box 遊戲可沿用這個 adapter，但不要把 sky colour index 當作 terrain、door 或 roof geometry。

Door state 也應先保存 raw signal 再交給 rules service：reference `WallDoorFlagsGet` 的 no-wall default `1` 與 walled `x3` detail 是不同語意，不能只用非零判定「有門」。CoAB 目前只顯示 `WallDoorFlags` evidence；後續作品可共用 flag adapter，再注入 lock／bash／pick／knock skill 與 mutation rules。

Dungeon movement 的 door branch 也要和 generic wall collision 分層：CoAB `CanMoveDungeonWrapped` 只放行無 wall 或 detail `1` 的 unlocked door，detail `2/3` 保持 blocked；`UnlockDoorWrapped` 只做 reference 的雙側 raw mutation。上層作品 service 必須先完成 skill／dice／spell transaction 才能呼叫它，不能讓 renderer 直接解鎖。

DOS player import 應保留 thief skill array 的 ordinal，不只保存 class：CoAB `ThiefSkills[1]` 對應 reference `open_locks`，透過 `Character.OpenLocksSkill()` 提供 rules input。其他 Gold Box 遊戲可重用欄位與 JSON contract，但不得用 local class 或 Dexterity 重新推算未經證實的百分比。

Door action resolver 可沿用同一個跨作品 contract：`PickLock` 依 marching order 為每位隊員消費一次 d100，再以健康狀態與已驗證 open-lock skill 判定 `roll <= skill`；成功才交給地圖 service 做雙側 unlock，失敗仍消耗本次 pick opportunity。Knock 應以 reference spell ID `0x1F` 搜尋並消耗第一個 memorized slot。這個 resolver 不應自行碰 renderer 或把 bash 規則混進來。

CoAB 現在把該 contract 接到 dungeon preview：P 僅對 detail 2 撬鎖，K 對 detail 2/3 消耗 Knock 並解鎖。State 持有獨立 dungeon RNG seed，Ebiten 只負責輸入、訊息與 GEO mutation；其他 Gold Box 遊戲可重用 action service，但必須替換 menu／map context，不應把 preview 快捷鍵當成原版完整 dungeon loop。

撞門規則不能簡化成 Strength 百分比：CoAB reference 以 `Str.full`／`Str00.cur` 選擇 die size 與 threshold，detail 3 另有 unpickable door table；`bash_door()` 甚至在 Strength 18、exceptional 0–50 先成功再額外擲骰。共用 parser 應保存 DOS full／exceptional 欄位，resolver 接受骰子注入，並讓 map service 在成功後才做雙側 unlock。

Locked door menu 的 capability 應在 raw detail 之上計算：detail 2 可列 Pick，detail 3 不列 Pick；Knock 必須先存在 `0x1F` memorized slot，Bash 對兩種 locked detail 都列出。preview 的方向鍵阻擋可重用這個 menu service，但各 Gold Box 遊戲仍需提供自己的文字／游標與劇情 entry。

Party icon import 也要分 raw slot 與實際 DAX block：DOS `icon_size=1` 不是把 raw ID 當成 normal，而是 CHEADT／CBODYT 的 `raw+0x40` namespace；`icon_size=2` 保持 raw ID。共用 party projection 可保存 raw bytes，renderer adapter 再呼叫 normalized mapping，讓後續 Gold Box 遊戲替換其 CHEAD/CBODY file family。

攻擊姿態再沿用同一個 block contract 加 `0x80`：reference `LoadIcons(normal_id, normal_id+0x80)`，因此 small icon 的攻擊 block 會是 `(raw+0x40)+0x80`。方向只在 `>3` 時選水平翻轉副本；不可把攻擊姿態誤當成另一套方向 ID。

Combat icon facing 應使用 reference `HalfDirToIso={7,2,3,6}`，不是直接把 map direction 當 screen direction；party 用 `HalfDirToIso[mapDirection/2]`，enemy 加 4 modulo 8。這個四組方向同時供 placement 與 CHEAD/CBODY flip adapter 使用。

## Tavern Tale boundary

Adventure Journal 的 Tavern Tale 是編號資料，不是任意酒館 flavor text。共用 UI 應一次消費一個 tale ID／文字，保存目前 sequence index，避免玩家一次看到尚未觸發的線索。真假、城市條件、買酒價格、重複規則與 random trigger 由 ECL／script adapter 提供；在證據不足時，`SetBarTales` 的 injected sequence 比硬編全域順序安全。

## ECL DAMAGE state boundary

`DAMAGE` 和 `SPELL`／`PROTECTION` 一樣，先由 VM 產生資料 signal，再由 State 保存
一次性 pending request。CoAB 的 `State.ConsumeDamageRequests()` 會保留五個 raw
operand 的順序；現在 DOS `saveVerse` `0xDF–0xE3` 已保存到 Character，並可由
`ResolvePendingECLDamage` 以注入骰點處理 selected／whole-party branches，並由
`ResolvePendingECLDamageWithHitResolver` 處理 random target；DOS `field_186` 已納入
save resolution，`ResolvePendingECLDamageWithDefaultHitResolver` 則提供 fighter／
equipment AC、invisibility、action-delay-aware blink 與 displace consumed-bit 的
共用投影；context variant 可由戰鬥回合傳入 delay／round，並將 FX effect-data bit
寫回 working roster，並在 resolver error 時 rollback。其他 affect 與死亡 continuation
仍應由作品 adapter 補上；目前 DamageOutcome 已保存 exact-zero／overkill 的 health
state；`CheckAffectsEffect(Death)` 尚未接入 ECL queue。active combat 的 ECL queue 現已
透過 `Battle.SetHitPoints` 接入 win/loss transition，並在倒下時呼叫
`Character.RemoveCombatAffects`；`CheckAffectsEffect(Death)` 與完整 combatant removal
仍待接入；Battle HP=0 時已清除 `HasCombatPosition` 並發出 `DeathOverlay` signal；
Ebiten 以死亡座標 anchor 顯示目前的繁中「倒下」overlay；Battle 會清除 per-fighter
`CombatAction`，team party 另保存 `DownedCorpse` 對應 `Tile_DownPlayer=0x1F`；Cure Light
Wounds 可治療可復原的 corpse 但不恢復 position；只有明確 `CombatHealAllowed` 的
affect_63 recovery 會用保存座標呼叫 `RestoreCombatant` 站起。State current turn 也會清除施法／移動／檢視
selection。State `ResolveDeathEffects` 現可在 caller 明確提供 damage flags／combat-heal
條件時 transactionally 套用 affect_63 recovery 與 TrollRegen；未知 target side effect
不會因缺少資料而猜測；`ResolveDragonSlayer` 另要求 caller 明確提供 target monster
kind、strength damage bonus 與 d12 roller。

## 中文化注意

Held effects 也已接到 enemy turn：reference `Player.IsHeld` 的 helpless、snake charm、
paralyze、sleep（`0x1F`／`0x33`／`0x34`／`0x35`）會讓怪物跳過整個回合；`AttackTarget01`
的 `target.IsHeld()` 例外則讓攻擊 held target 必定命中。這只處理 combat action boundary，
不代表已完成解除、豁免、持續時間或治療流程。

## Game-time adapter

`ovr021.step_game_time` 的時間槽不是單純 combat round：slot 2 以 10 分鐘換算，
更高槽位依 `{10,10,6,24,30,12,0x100}` 級聯進位。`State.AdvanceGameTime` 現在保留
raw clock、套用同一 elapsed-minute 計算，並讓 party `.FX` 與 active battle raw effects
共同到期；`Strength=0xFF` 永久 effect 不被移除。slot-6 overflow 暫存為 age cycles，
尚未寫回 DOS player age 欄位。

REST 現在使用同一 adapter：`REST_START` 將 requested hours 轉成每小時 60 個 slot-1
minutes，先跑 effect timeout，再執行既有每 24 小時 +1 HP 的 bounded natural healing。
這保留了原版「時間推進與 effect expiry 先發生」的順序；中斷、safe location、spell
learning 與完整 rest encounter table 仍不能由此窄 slice 推論。

remake JSON save version 5 現在保存七個 raw clock slots 與 age-cycle overflow，
`State.SavePartyFile`／`LoadPartyFile` 會保留時間進度；versions 1–4 仍可載入並使用零時鐘。
remake JSON 與 DOS SAVGAM Area1 clock 現已在已定位 raw offsets 層分開保存；其他未知
Area1 bytes 仍不會被 remake JSON 欄位誤宣稱為原版格式。

更新：reference Area1 的七個 clock words 位於 `0x18C..0x198`，不是 unknown bytes；
`area.State.GameTime`、Area1 codec、SAVGAM load 與 State `AdvanceGameTime` 現在共用這個
raw order。其他未定位 Area1 bytes 仍維持 preservation boundary。

Reference `NormalizeClock` 在 slot 6 overflow 時會對每個 `Player.age` 加一。CoAB normal
player record 的 signed age `0x76` 已由 parser／`Character`／`PatchDOSPlayerRecord` 保存，
而 `State.AdvanceGameTime` 會同步 party roster 年齡並以 int16 saturation 防止 wrap；Pool/Rad
record 的 `0x30` 仍是另一個 importer contract。

Reference map-position UI 只直接顯示 `time_hour` 與十進位 `HH:MM`；remake 的
`GameTimeDisplay`／`GameTimeText` 另外把已由 `timeScales` 支持的日／月／年欄位以繁中
HUD 暴露。這是可重用的 renderer-neutral adapter，不代表已完成所有作品的日曆或時間事件。

Age modifier 是 new-character generation rule，不是載入既有 DOS player 後每次重算的
runtime effect。共用 adapter `Abilities.WithAgeEffects` 保存五段 race brackets 與六項
delta；作品 UI 先提供 age/base abilities，再由 creation rules 呼叫，避免 imported
record double-count。

新角色年齡另有 `race_ages[race][class]` 生成表，不應用 age bracket threshold 反推。共用
`StartingAgeSpecFor`／`RollStartingAge` 保存 base age、dice count／size 與 unsupported
combination error；後續作品只需替換 race/class table，即可重用 deterministic creation
resolver。

角色建立 UI 現已把這個 resolver 與既有 `Character.Validate` 真正接線：22 個 single-class
race/class 組合可由選單選取，加入隊伍時才對 copied character 套用 age roll／age effects。
多職業／dual-class 與原版完整 create/modify/drop menu 仍是獨立 boundary；half-orc 已接入
cleric／fighter／thief 單職業生成。外部 AD&D 1e age table 與 reference class-index mapping
已重新核對，half-orc fighter 採 `13+1d4`；原先將此值誤讀成 ranger slot 的 boundary 已移除。

目前 `State.AddCreationCharacter` 是 age transaction 的作品 adapter：template preview 不變，
copy 入 roster 時才生成 age、套用 ability delta，再給 copied ID。這保留 UI 編輯可重複預覽，
也避免對同一 template 累積 modifiers；完整原版 race/class selection 仍可沿用同一 ordering。

### Multi-class rules projection（CoAB round 262）

`Character.HasClass` 是跨 Gold Box 可重用的 compatibility seam：若 `ClassLevels[8]` 有
raw level metadata，規則層依 slot 判定職業；舊資料沒有 metadata 時才回到 primary class。
目前已接入 equipment class-mask、CAMP MAGIC 與 combat spell eligibility。THAC0、HP growth、
高等級 spell capacity 與 UI label 仍須各作品依 reference class-level table 擴充，不能把
primary-class fallback 當成完整 AD&D multi-class engine。

Tavern Tale 的繁中翻譯要保留角色名、地名與線索方向，不以 renderer 的 byte length 截斷中文。訊息顯示仍沿用 Unicode rune reveal；後續若接入完整 62 則，應維持 `bar_tale_<id>` 或獨立 catalog，並以來源編號做 regression。

### ECL selected-player bridge（CoAB round 269）

`WHO` 是 UI pause／resume transaction；`LOAD CHARACTER` 則是 script 直接指定角色的
1-based selector。State 將其低 7 bits 映射到 `partyRoster`，與 WHO 共用 selected-player
ID；無效 selector 只留下 not-found flag，不破壞上一個有效選擇。bit 7 目前只保存為
restore／redraw metadata，未把尚未完整反組譯的 DOS global cleanup 假設成已完成。這個
「VM decoded request → persistent roster adapter」是其他 Golden Box 作品可重用的接點。

selected-player identity 不只供 UI 顯示：`FIND SPECIAL` 會立即查該角色的 active affect。
因此 VM 的 selected index 必須和 memory／string／compare state 一樣可恢復；WHO pause
前不能先猜角色，resume 後才提交 UI index，LOAD CHARACTER 無效 selector 也不能清掉
上一個有效角色。作品 State 只需提供 active effect IDs，VM 負責 reference `=`／`<>` 分支。

script-controlled party departure 必須和 player-facing drop 分開：ECL `DUMP` 可移除最後一人，
並依原 index選前一位／新第一位；ALTER DROP 則保護玩家避免留下空隊伍。State adapter用
stable character ID同步移除 fighter，save layer稍後處理 stale DOS sidecars。VM 先更新
working party，確保同一 ECL run 的 FIND ITEM／FIND SPECIAL 不會看到已離隊角色。
BlockSession 應持有 public context 的 deep copy並跨 NEWECL重用；每個 block重新 clone會讓
角色復活，直接改 caller slice則會在 State正式消費 DumpRequest前洩漏 mutation。

### Terrain menu branches（CoAB round 298）

SearchLocation 的 terrain selector 是事件入口，不是 UI event ID。原始 menu option
順序必須保留給 ECL selection index；翻成較寬的中文標籤時只替換 State display
字串。像火刀據點 `0x99 → selector 0x19` 的 WAIT 分支，regression 應同時鎖定
原始 index、沒有 `DAMAGE` signal，以及後續 plot text，避免 renderer 因中文字寬度
重新排序選項而改變劇情。

全隊環境陷阱可由 `DAMAGE flags & 0xE0 == 0xE0` 安全地在作品 State 自動提交：
`0x80` 表示非 random-target、`0x40` 表示 whole party、`0x20` 表示 auto damage。
這一型不需要 WHO、saving throw 或 CanHitTarget，且一個 damage roll 套用到所有
隊員。其他 flag 組合不可順便自動處理，必須保留 pending request 給各自 adapter。

#### ECL2 block 4 selector map

SearchLocation entry `+0x00E1` 在 `+0x00F5` 以 `C04F & 0x003F` 做 33-way
dispatch。火刀據點後半可重用的定位如下：

| Selector | Handler | State | Room |
|---|---:|---|---|
| `0x19` | `+0x0C2A` | `4CFE & 0x20` | 旋轉刀刃；ENTER 於 `+0x0D16` 送出 `0xE0,8d8,0,0` |
| `0x1A` | `+0x0DA2` | `4CFE & 0x40` | 被定身的火刀；撤退／審問／殺死，審問解鎖手札 26 |
| `0x1B` | `+0x0F17` | `4C10`、`4CFE & 0x80` | 火刀高層辦公室；多階段文件、手札 9 與 treasure |
| `0x1C` | `+0x101A` | `4C11=1` | 走廊的奇怪煙味 |
| `0x1D` | `+0x105F` | `4C12=1` | 異常整齊、由隱形僕人復原的臥室 |
| `0x1E` | `+0x111C` | `4C13=1` | 焚毀圖書館、焦屍紙張與手札 29 |
| `0x1F` | `+0x120D` | `4C14=1` | 被烈焰摧毀的實驗室 |
| `0x20` | `+0x1279` | `4C15=1` | 「待復活／待埋葬」兩排覆屍 |

`0x1C–0x20` 多採 `COMPARE visited → EXIT → SAVE 1 → PRINT → 共用 Continue`
模板；`0x1B` 是多階段狀態，不可壓成單一 visited boolean。刀刃事件則在選單前
先設定 `4CFE|=0x20`，所以 ENTER、WAIT、RETREAT 都會消耗事件；ENTER 的順序是
主選單 → 撕裂提示 Continue → DAMAGE → 消散 Continue → `SPRITE OFF/CALL
0x2E10/EXIT`。

#### Explicit dungeon SEARCH

SEARCH 不是每步自動搜尋。作品 UI 在玩家按 `S` 時暫時寫 `7ECA=1`，只執行
SearchLocation entry，結束後清回 0。火刀辦公室 `0x1B` 展示標準多階段模式：
首次普通進入將 `4C10` 從 0 設成 1；只有 `4C10==1 && 7ECA==1` 才搜索抽屜，
接著先設 `4C10=2` 與 `4CFE|=0x80`，再顯示文件、手札與寶物。不能把 SEARCH
簡化成 renderer 額外文字，也不能在每次 movement 永久維持 `7ECA=1`。

`CLEARMONSTERS → TREASURE → COMBAT` 是原引擎 treasure-service dispatch。
若 `TREASURE` 已產生可領取物，State 必須先開 treasure UI，而不是因看見
`CombatRequested` 就建立戰場；其 return mode 也要保存來源，地城搜索所得財寶
結束後應回 ModeDungeon。`ItemBlock=0x80+n` 表示 n 件隨機物品，而非一般 ITEM
block ID。

selectors `0x1C–0x20` 證實常見的一次性房間模板可跨 Gold Box 作品重用：
`COMPARE flag,1 → IF= EXIT → SAVE 1 → PRINT/GOSUB Continue → dungeon return`。
但 pause 數仍由 script 決定；例如焚毀圖書館先描述焦屍，再於第二個 pause 取得
紙張與手札 29，renderer 不能把多段文字合成後提前解鎖 journal。作品層只負責
依原始 text signal 在正確 continuation append 中文手札。

#### Encounter-owned treasure

`TREASURE → COMBAT` 不能只按 opcode 相鄰關係判成商店／財寶服務。ECL2 block 4
火刀首領先建立 21 名 monster spawns，再建立 treasure packet，最後送出 COMBAT；
該 packet 是勝利獎勵，戰前必須保持 raw pending。可重用的判別規則是：

- `CombatRequested && len(MonsterSpawns)>0`：遭遇戰，TREASURE 延後至勝利。
- `CombatRequested && len(MonsterSpawns)==0`：引擎服務 boundary，可立即解析財寶。

戰後敘事也可能跨多次 PICTURE／Continue 才 `NEWECL`；因此 pending reward 的
ownership 應跟著 ECL session，而不是跟著單一 renderer 畫面。實作上在勝利後
解析 reward，loot menu 關閉時再 resume 保存的 ECL engine boundary；不可先
resume 多段故事、期待任意後續 EXIT 恰好撿回 pending loot。火刀首領線另證實
劇情旗標有嚴格時序：`4CFF`（解除火刀枷印）→ `4C2A`（皇家事件）→ 四段夢境 →
`7F12`（跨章 bond progression），不可在戰鬥結束瞬間一次寫完。

#### World destination state

ECL1 block `0x50` 的 route selection 以 `4C9C` 保存目的地，world dispatcher
再將它寫入 `4C9B`（current location）與 `4CA1`。State 不可依「新遊戲時第四個
selection」硬編碼城市；章節 continuation 已累積許多 Continue/menu choices。
可重用 adapter 應優先讀 dispatcher bytes：CoAB 的 `1/2/3` 分別投影為
Shadowdale/Ashabenford/Dagger Falls，再以舊 sequence 僅作 synthetic fallback。

loot 選擇「暫不收下」代表離開剩餘物品，必須清掉 pending item list；否則下一個
沒有 TREASURE 的 encounter 勝利時仍會誤開上一戰 loot。火刀首領→Tilver's Gap
端到端流程已用這個跨戰鬥 regression 鎖定 ownership。

world location values 不只三座早期城市。`4C9B=4` 是 Standing Stone；作品層應用
明確 value→Location table 投影，而不是把超過 3 的值忽略、也不能與 remake
內部 `LocationTilverton` enum 數字混為一談。

`PROGRAM` ID 也不是全域單義。Ashabenford HALL 再次證實 PROGRAM 0 必須和
玩家剛選的場所 context 配對為 training service；只有其他 start-menu context
才回標題。可重用 VM 應輸出 routine boundary，語意由作品 State 的 caller context
判斷。
