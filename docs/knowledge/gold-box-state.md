# Gold Box 共用 state／城市 service 知識

## 城市場所分層

Gold Box 城市應先由 `ModePlace` 保存 `INN／STORE／BAR／LEAVE` parent menu，再由各場所進入自己的 service state。場所 service 完成後以 `ModeEvent` 保存可翻譯訊息與 `eventReturnMode`，Enter 才回 parent menu；不可把荒野、城市 map 與場所事件混用成同一個返回值。

目前可沿用的 adapter 是：`INN` 的 safe HP restore、`STORE` 的 price-injected transaction、`BAR` 的 ordered `SetBarTales` sequence。各遊戲只需替換城市／ECL data loader，不應複製 renderer 或把原版 routine 的未知副作用猜進共用 state。

CAMP 入口與 REST 必須分離：`PROGRAM 9`／CAMP 只建立 camp command state，不能直接把 HP 設成 MaxHP；REST menu 再以遊戲提供的時間來源計算自然恢復。CoAB 目前採 24 小時一個可測試單位、每位受傷角色 +1 HP，並以 character ID 同步 combat projection。memorize duration、clock、safe location 與 random interruption 應由後續作品的 rules／ECL adapter 注入。

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

玩家輸入造成的 combat error 也應是可恢復 transaction：input adapter 將 `ValidateAttack`／彈藥／target selection 的 error 送到 localized message presenter，保留目前 Mode、turn、HP 與 inventory，不能直接結束 Ebiten game loop。`combat.ErrAdjacentMissileTarget` 可作為跨作品共用的規則錯誤識別；啟動／資料載入錯誤則仍可向上回報。

ECL `ADD NPC` 應採資料 signal boundary：runner 保存 operand 的 NPC ID 並繼續至下一個 command，game adapter 再依作品 NPC table 決定是否建立角色／對話／隊伍 side effect。`NPCIDs` 可跨 Gold Box runner 重用，但不能把 ID signal 當成已完成 NPC record lookup。

`LOAD PIECES` 也應先保存三個 selector，再由作品的 file／map adapter 解讀：目前 ECL2 `1,2,3` 與跨 ECL 章節的 repeated observations 只證明 operand shape，不足以把欄位硬編成某一個 floor、wall 或 tile 檔案。共用 runner 可沿用 `LoadPiecesRequested`／`LoadPieces` signal，後續 Gold Box 遊戲替換實際 map loader。

State 層不能丟掉已驗證的 ECL signal：`LoadPiecesRequested` 應像 `LOAD FILES` 一樣進入一次性 `ConsumeLoadPiecesRequest()`，renderer／map adapter 再決定如何解讀，避免 VM 直接依賴 ZIP 檔名，也保留後續作品替換地圖格式的空間。

## Tavern Tale boundary

Adventure Journal 的 Tavern Tale 是編號資料，不是任意酒館 flavor text。共用 UI 應一次消費一個 tale ID／文字，保存目前 sequence index，避免玩家一次看到尚未觸發的線索。真假、城市條件、買酒價格、重複規則與 random trigger 由 ECL／script adapter 提供；在證據不足時，`SetBarTales` 的 injected sequence 比硬編全域順序安全。

## 中文化注意

Tavern Tale 的繁中翻譯要保留角色名、地名與線索方向，不以 renderer 的 byte length 截斷中文。訊息顯示仍沿用 Unicode rune reveal；後續若接入完整 62 則，應維持 `bar_tale_<id>` 或獨立 catalog，並以來源編號做 regression。
