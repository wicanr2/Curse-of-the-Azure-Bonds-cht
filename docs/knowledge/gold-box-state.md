# Gold Box 共用 state／城市 service 知識

## 城市場所分層

Gold Box 城市應先由 `ModePlace` 保存 `INN／STORE／BAR／LEAVE` parent menu，再由各場所進入自己的 service state。場所 service 完成後以 `ModeEvent` 保存可翻譯訊息與 `eventReturnMode`，Enter 才回 parent menu；不可把荒野、城市 map 與場所事件混用成同一個返回值。

目前可沿用的 adapter 是：`INN` 的 safe HP restore、`STORE` 的 price-injected transaction、`BAR` 的 ordered `SetBarTales` sequence。各遊戲只需替換城市／ECL data loader，不應複製 renderer 或把原版 routine 的未知副作用猜進共用 state。

CAMP 入口與 REST 必須分離：`PROGRAM 9`／CAMP 只建立 camp command state，不能直接把 HP 設成 MaxHP；REST menu 再以遊戲提供的時間來源計算自然恢復。CoAB 目前採 24 小時一個可測試單位、每位受傷角色 +1 HP，並以 character ID 同步 combat projection。memorize duration、clock、safe location 與 random interruption 應由後續作品的 rules／ECL adapter 注入。

Spell UI 也應保持三層：DOS／save adapter 保存 ordered slot IDs，verified catalog 只把有證據的 class／ID 映射成名稱，rules engine 才處理 CAST／MEMORIZE／SCRIBE 與消耗。CoAB 目前只核對一級牧師／魔法師前八個 ID；未知 slot 顯示 hex 比猜 global ordinal 更安全，後續 Golden Box 遊戲可替換同一個 catalog adapter。

`KnownSpells` 與 `SpellSlots` 必須分開保存：前者是角色已學會的 spell-book flags，後者是目前已記憶欄位。DOS parser、party Character 與 versioned JSON save 都應保留兩者；UI 只呈現數量與已核對名稱，不能因此推導可施法規則。

RuleBook 的 Magic Menu command 順序是 `CAST／MEMORIZE／SCRIBE／DISPLAY／REST／EXIT`。共用 state 可先實作 command routing；`DISPLAY` 只讀 roster，`REST` 呼叫作品專屬休息 service，而 CAST／MEMORIZE／SCRIBE 必須等待 spell target、capacity、time 與 interruption evidence，不應在 renderer 中猜測。

`MEMORIZE` 應使用兩階段 transaction：known-spell selection 只寫 pending state，`REST_START` 才 commit 到 memorized slots。這樣後續作品可以替換 capacity、每等級準備時間、部分成功與 random interruption，而不必改動 UI 或 DOS／save adapter。

## Tavern Tale boundary

Adventure Journal 的 Tavern Tale 是編號資料，不是任意酒館 flavor text。共用 UI 應一次消費一個 tale ID／文字，保存目前 sequence index，避免玩家一次看到尚未觸發的線索。真假、城市條件、買酒價格、重複規則與 random trigger 由 ECL／script adapter 提供；在證據不足時，`SetBarTales` 的 injected sequence 比硬編全域順序安全。

## 中文化注意

Tavern Tale 的繁中翻譯要保留角色名、地名與線索方向，不以 renderer 的 byte length 截斷中文。訊息顯示仍沿用 Unicode rune reveal；後續若接入完整 62 則，應維持 `bar_tale_<id>` 或獨立 catalog，並以來源編號做 regression。
