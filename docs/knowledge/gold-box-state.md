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

## Tavern Tale boundary

Adventure Journal 的 Tavern Tale 是編號資料，不是任意酒館 flavor text。共用 UI 應一次消費一個 tale ID／文字，保存目前 sequence index，避免玩家一次看到尚未觸發的線索。真假、城市條件、買酒價格、重複規則與 random trigger 由 ECL／script adapter 提供；在證據不足時，`SetBarTales` 的 injected sequence 比硬編全域順序安全。

## 中文化注意

Tavern Tale 的繁中翻譯要保留角色名、地名與線索方向，不以 renderer 的 byte length 截斷中文。訊息顯示仍沿用 Unicode rune reveal；後續若接入完整 62 則，應維持 `bar_tale_<id>` 或獨立 catalog，並以來源編號做 regression。
