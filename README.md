# Curse of the Azure Bonds 中文化／Remake

這是 SSI《Curse of the Azure Bonds》（青色枷的詛咒）的反組譯、繁體中文化與 remake 研究專案。目前是**可執行的初步 prototype**，不是完整重製版；GitHub 上的每輪提交都保留可測試的成果與驗證邊界。

## 目前成果

以下圖片由原始 `curseoftheazurebonds.zip`，透過專案目前的 DAX／GFX／GEO parser 離線產生，證明圖像資料管線已經接通：

![TILES.DAX 原始圖塊 gallery](docs/screenshots/tiles-gallery.png)

![GEO2 原始 16×16 wall geometry](docs/screenshots/geo-geometry.png)

![原版規則生成的 50×25 wilderness floor 局部](docs/screenshots/wilderness-floor.png)

![GEO2 wall/door 組合出的 dungeon floor slice](docs/screenshots/dungeon-floor.png)

![原始 CPIC 戰鬥小人與效果 sprite sheet](docs/screenshots/combat-sprites.png)

目前已完成的垂直切片包括：

- DAX 容器／RLE、ECL bounded VM trace 與跨 ECL1–ECL6 block context。
- 繁中開場、暗影谷／阿沙本福德／匕首瀑布城市 routing、荒野／場所狀態、角色建立、可恢復的 remake game JSON 存檔，以及可操作戰鬥 prototype。
- `TILES.DAX`／`8X8D*.DAX` indexed pictures、`WALLDEF*.DAX`、EGA16 palette 與 `GEO2–GEO6` geometry parser。
- 原版 50×25 wilderness floor 生成規則、background entry → tile index mapping，以及依 movement cost 的荒野移動。
- GEO2 wall／door fields → dungeon background composition → TILES pixel art 的可見 slice（`D` 預覽）。
- dungeon table／chair decoration 已依 GEO `terrain & 0x40` 與原版 seeded dice pass 接入。
- Ebiten 原始 tile gallery、GEO wall viewport 與依 GEO wall bytes 驗證的游標移動。
- 已從 `CPIC1.DAX`–`CPIC6.DAX` 抽出 156 張透明背景戰鬥小人 PNG；完整索引在 [`assets/sprites/README.md`](assets/sprites/README.md)。
- Ebiten 戰鬥畫面已載入 repo 內 CPIC PNG，並依 ECL monster `IconBlock` 顯示敵方小人；無對應 block 時有 deterministic fallback。
- `SPRIT1.DAX`–`SPRIT6.DAX` 的 frame stream 也已解析並抽出 138 個逐幀 PNG，manifest 同時記錄 delay／尺寸／座標。
- 戰鬥 renderer 會依 ECL `SETUP MONSTER` 的 SPRIT block 與原始 delay 循環播放逐幀 PNG，缺圖時退回 CPIC 靜態圖。
- 戰鬥畫面已依當前 active character 套用 CombatMap camera transform；較大的 reference placement 座標會先轉成 viewport 座標再繪製。
- SPRIT manifest 的 frame `x/y` placement 也已接入戰鬥 renderer，播放時會依原始 frame canvas offset 顯示。
- PIC1–PIC6 的 PIC/FINAL-style XOR frame delta 也已解碼並抽出 152 張 PNG；SPRIT 與 PIC 兩種 payload 語意在 parser 中明確分流。
- ECL `PICTURE` request 已接到繁中事件畫面：game state 保存 block、Ebiten 播放對應 PIC frames，Enter 可返回原流程。
- PICTURE block `>= 0x78` 已分流到 BIGPIC 靜態大圖；目前從 BIGPIC1／2／6 抽出 4 張原始大圖並在事件畫面置中顯示。
- 一般場景人物的 `HEAD2–6`／`BODY2–6` 也已抽出並依 reference body `y+5` 合成 30 張 PNG，後續城鎮／事件 renderer 可直接載入。
- PICTURE 的 Area2 head sentinel 分支也已接入：有 head block 時改顯示 HEAD/BODY scene composite，無 head block 時維持 PIC／BIGPIC。
- Area2 `HeadBlockId @ 0x5C2` 已接入 binary codec；載入 raw area 後會自動驅動上述 HEAD/BODY 分支。
- 戰鬥畫面已改用 tile-derived formation placement，並建立 reference 八方向 delta contract；真實 CombatMap position／camera data 仍待解碼，但 active-character camera transform 已接入 renderer。
- `combat.Fighter`／game battle state 已保存 CombatMap position／size；外部真實座標優先，缺少時才使用 deterministic formation fallback。
- 已封裝 reference 的 encounter team origin／facing：`combat.EncounterTeamStart`；實際 `mapDirection`、occupancy 與候選格排序仍待 Area／Player record 解碼。
- reference `try_place_combatant` 的 position formula 已建立可測試 adapter，待 team／occupancy inputs 解碼後即可取代 fallback。
- 已從 `CHEAD.DAX`＋`CBODY.DAX` 合成六組 normal／attack party combat icon，Ebiten party fighter 會依 fighter icon state 顯示小人；合成、透明、方向 flip 規則與跨 Gold Box 知識整理在 [`docs/knowledge/gold-box-graphics.md`](docs/knowledge/gold-box-graphics.md)。
- 新建角色的玩家 icon default 已依原作 race switch 建立：矮人／侏儒／半身人 small，其餘 normal；head／weapon 初值為 block 0。
- Area1／Area2 已知欄位已有 `0x800` bytes binary round-trip codec，未知 bytes 會保留。
- 原始 `ITEMS` 已解析為 128 筆 base-item descriptor；`cmd/azure-bonds -base-items` 可列出裝備欄位／傷害／可用職業與目前繁中名稱 catalog。
- 已新增 `Character.FighterWithEquipment`：已知 `ITEMS` descriptor 的 readied 武器／護甲可投影到戰鬥 fighter；舊 party JSON 與未帶 equipment 的角色行為不變。
- party inventory 已有 `EquipItem`／`UnequipItem` contract，會驗證 class usability、雙手／副手衝突與最多兩枚戒指。
- `RemoveItem` 已支援 Count stack decrement、readied protection 與 cursed equipment lock，供後續商店／treasure mutation 使用。
- `UseConsumable` 已支援卷軸 stack、藥水單次移除與魔杖 charge decrement，回傳繁中化 UI／後續法術 engine 可用的 effect signal。
- ECL `SPELL`／`PROTECTION` 已由 bounded VM 回傳 `SpellSearches`／`ProtectionRequests` signal；實際 party spell-slot lookup 與效果 engine 仍待接入。
- party roster 已有 ordered `SpellSlots` first-match resolver，且 game bootstrap 會載入原始 `ITEMS`，讓角色建立／party load 的 readied equipment 影響 fighter projection。
- 已加入 bounded DOS player record spell parser：`.SAV`／`.GUY` 的 memorized slots 與 known-spell flags 可接到 `Character.ApplyDOSSpellRecord`；完整 DOS save/import container 尚未完成。
- 已加入 `ParseDOSPlayerRecord`：可將已解壓的單職業 `.SAV`／`.GUY` 核心欄位（姓名、能力、HP、等級、head／weapon／icon_id／size、金幣與法術）投影到 party／戰鬥；`.SWG` inventory、`.FX` effects 與多職業仍待完成。
- 已加入 `.SWG` inventory 匯入：連續 `0x3F` item records 可接到 `DOSPlayerRecord.Inventory`／`Character.Equipment`，readied 基本裝備可沿用既有 fighter projection；pointer resolution 與 `.FX` effects 仍待完成。
- 已加入 `.FX` effects 匯入：連續 9-byte effects 可保存到 `DOSPlayerRecord.Effects`／`Character.Effects`，並提供常見效果繁中名稱；effect gameplay tick／解除仍待完成。
- `.FX` duration／strength 欄位已依原始格式修正：16-bit 分鐘與 `255=永久`，並提供 `AdvanceEffects` duration tick；effect-specific gameplay 仍由後續 rules layer 處理。
- 新增 `ParseDOSPlayerFiles`：將必要的 `.SAV/.GUY` 與可選 `.FX/.SWG` sidecars 組成可用的 party `Character`，並保存 gold/gems/jewelry；`SAVGAM?.DAT` container 尚未解析。
- CLI 可用 `-import-character -character-record <file> [-character-effects <file>] [-character-inventory <file>] -out-party <json>` 將原版角色匯入 remake party JSON；不會修改原始檔案。
- `cmd/azure-bonds-game` 也支援 `-dos-character-record`（及 optional `.FX/.SWG`）直接以原版單一角色啟動 remake；`-party-load` 與此模式互斥。
- imported active Bless／Curse／Blind／Bestow Curse／friendly Prayer effects 會投影到 fighter attack／AC（可確認的修正為 +1、-1、Blind -4/+4 AC、Bestow Curse -4、Prayer +1）；需要目標或戰鬥 phase 的 effects 仍待 rules layer。
- 城市 `INN` 已接成安全休息場所：恢復 party roster 與畫面 fighter 的 HP，並以繁中訊息返回場所選單；CAMP 的 SAVE／VIEW／MAGIC／ALTER／FIX 與 BAR Tavern Tale 窄 service 已逐步接入，買酒價格／完整酒館 trigger 與 CAMP 時間／中斷規則仍待完成。
- 已建立商店 Buy／Sell／ID 的 party transaction contract：價格由後續 shop stock 提供，ID fee 為 200 GP；完整 money pool 與 Shop Menu UI 仍待接入。
- 城市 `STORE` 已接入繁中 Shop Menu（購買／查看／取出／集中／分配／估價／離開）；尚未載入 stock 的 action 會明確提示並可返回選單。
- 已接入 injected shop offers 與 party money pool：可集中／提取／平均分配金幣，並由 pool 購買指定 offer；價格仍由城市／ECL data 提供。
- `STORE → 購買` 現在會列出繁中商品與 GP 價格，選取後扣 pool 金幣並加入未裝備物品；目前 active shop character 預設為第一位。
- `STORE → 查看` 現在會列出角色 HP／金幣與繁中裝備摘要，選取後可返回 Shop Menu。
- `STORE → 取出金幣` 現在可選角色與 1／10／100／全部金額，更新 party pool 與角色金幣後返回 Shop Menu。
- `STORE → 估價` 現在可選角色與寶石／珠寶，接受外部注入報價後清除財寶並將 GP 加入 party pool。
- APPRAISE 現在會先顯示「接受／拒絕／返回」確認；拒絕報價會保留財寶與 party pool。
- 荒野 `CAMP` 現在會進入繁中 `SAVE／VIEW／MAGIC／REST／ALTER／FIX／EXIT` 選單；`REST` 可返回 CAMP Menu，`EXIT` 返回荒野選單，`ALTER → ORDER` 可重排隊伍順序，`ALTER → DROP` 具二次確認並同步移除角色；尚未反組譯完成的 routine 仍保留明確 placeholder 邊界。
- `ALTER → PICS` 現在可切換怪物遭遇圖片與動畫；圖片關閉會略過事件圖片 renderer，動畫關閉會使用事件／戰鬥動畫首幀。
- `ALTER → SPEED` 現在可調整 1–5 級訊息速度，Ebiten 事件訊息會依設定逐字顯示繁中內容。
- `ALTER → ICON` 現在可選擇已抽出的 CHEAD／CBODY 頭部與身體圖層，並同步角色與戰鬥畫面小人。
- `CAMP → VIEW` 現在可選角色查看職業、HP、金幣、寶石、珠寶與裝備摘要，並可返回 CAMP Menu。
- 已接入已裝備弓／飛鏢的 RuleBook 多次攻擊：ITEMS RateOfFire raw `4/6` 分別投影為每回合 2/3 次攻擊；目標倒下時會依 target cursor 改攻下一個存活敵人。
- 已建立彈藥 transaction：保存武器 raw `AmmunitionType`，由資料層注入 raw code→inventory type mapping 後，CombatAct 會 atomic 扣除本回合箭／弩矢／飛鏢數量；未注入 mapping 時不臆測對應。
- 戰鬥中按 `D` 可執行 RuleBook `DONE`，不攻擊、不消耗彈藥，直接結束目前角色回合並進入敵方／下一位隊友回合。
- `CAMP → MAGIC` 現在提供原版已證實的 `CAST／MEMORIZE／SCRIBE／DISPLAY／REST／EXIT` command menu；`DISPLAY` 可查看各角色已記憶法術，`MEMORIZE` 可選取 known spells 並在 `REST → START` 寫回記憶欄位，`REST` 回用 CAMP 休息服務，施法／抄錄與完整 slot／時間規則仍待接入。
- `CAMP → SAVE` 現在會透過 state request 寫入 configured versioned remake party save，並顯示成功／錯誤訊息；原版 `SAVGAM?.DAT` slot／area container 仍待反組譯。
- `CAMP → FIX` 現在會依已記憶的 Cure Light Wounds 對受傷隊員施放固定 `1d8` 治療，並同步 roster／戰鬥 HP；戰鬥中 S／H／C／W／P／G 會先進入施法目標選擇，左右鍵切換、Enter 確認、Esc 取消，B 進入 Bless 無目標確認，再分別施放 Magic Missile／Cure Light Wounds／Curse／Cause Light Wounds／Protection from Evil／Protection from Good；牧師與魔法師的職業分表 spell ID `7` 會正確分流。
- ECL encounter menu 的 `FLEE` 現在會進入繁中撤退事件並返回荒野；`PARLAY` 會提供 `傲慢／狡猾／謙卑／友善／威嚇` 五種談判策略。戰鬥中 `V` 可開啟不消耗回合的繁中角色檢視。怪物速度、追擊、speaker／reaction 與完整對話 script 仍待反組譯。
- 戰鬥按 `M` 可進入 MOVE，方向鍵移動當前角色一格；目前已同步 CombatMap 座標與 occupancy，移入敵方格會觸發攻擊、離開敵人鄰接範圍會觸發免費反擊；地形、負重與完整 facing 仍待反組譯。
- MOVE 已依 RuleBook 接入護甲移動上限：皮甲 12 格、中／重甲依表限制 9 或 6 格；方向鍵會逐格扣除 allowance，負重、地形與 FLEE 邊界仍待反組譯。
- 戰鬥已接入 missile 近身限制：已辨識的弓／弩／投石索不可攻擊相鄰目標，飛鏢保留 RuleBook 的 thrown exception；完整射程與 line-of-sight 仍待反組譯。
- 攻擊已加入不擲骰的 preflight：無效的相鄰 missile 攻擊會在彈藥 transaction 前拒絕，不消耗箭／弩矢。
- 敵方若有已驗證的多次攻擊 profile，也會沿用相同的 RateOfFire attack sequence。
- 玩家戰鬥輸入若違反射程／彈藥／目標規則，會顯示繁中錯誤並留在戰鬥畫面，不會結束遊戲主迴圈。
- ECL `ADD NPC` 現在會保存 NPC ID signal 並繼續執行；NPC 資料表與加入隊伍的完整 side effect 仍由後續 adapter 接入。
- ECL `LOAD PIECES` 現在會保存三個 map-piece selectors 並繼續執行；State request 會由 `WALLDEF{area}`／`8X8D{area}` raw adapter 消費，完整地城／牆面／碰撞副作用仍待完成。
- `LOAD PIECES` 現在會依反組譯證據載入 `WALLDEF{area}.DAX`／`8X8D{area}.DAX` selector，套用三組 global symbol offset，並在 dungeon preview 顯示素材 adapter 已就緒；牆面拼圖與完整 3D renderer 仍待完成。
- dungeon preview 現在會從目前 GEO wall 找出一組 reference 3D viewport layout，顯示原始 8×8D wall stamp sample；完整方向遍歷、遮擋與 camera 仍待完成。
- dungeon preview 現在會依 party facing 執行 Far／Mid／Near GEO wall traversal，展開有順序的 8×8D wall stamps；dungeon context 已套用 reference 16×16 coordinate wrap，sky／roof、door、遮擋與 camera 仍待完成。
- dungeon preview 方向鍵現在會依 GEO 雙側 wall collision（含 wrapped edge）移動 map position，Q/E 依 reference 八方向順序轉動 facing，並重建 floor／Far/Mid/Near wall view；正式 Area camera、scroll、movement cost 與 encounter 仍待完成。
- remake game save version 4 現在會保存 dungeon preview 的 `(x,y)`、八方向 facing 與 reference map wall cache；v1/v2/v3 舊版 save 可載入並安全回到相容預設，F9／啟動載入後會重建 floor 與 wall view。
- `CAMP → REST` 現在提供 `REST ADD SUBTRACT EXIT`，每 24 小時不間斷休息自然恢復 1 HP；一級法術記憶會先檢查「4 小時最低準備 + 每個法術 15 分鐘」，不足時保留 pending 選擇，完整高等級時間與遭遇中斷仍待反組譯。
- `城市 → BAR` 現在可逐則閱讀前六則繁中 Tavern Tale，按 Enter 回到酒館再離開返回場所選單；買酒價格、城市條件與完整 ECL tale trigger 仍待反組譯。內容整理見 [`docs/manual/tavern-tales-zh-TW.md`](docs/manual/tavern-tales-zh-TW.md)。

執行遊戲需要原始素材與可顯示繁中的 TTF／OTF 字型：

```sh
go test ./...
go run ./cmd/azure-bonds -base-items
go run ./cmd/azure-bonds-game -font /path/to/chinese-font.ttf
# 例：選擇原始 GEO3 block 0x10 作為目前 map preview
go run ./cmd/azure-bonds-game -font /path/to/chinese-font.ttf -geo-set 3 -geo-block 0x10
# 重新由本地原始 ZIP 產生 sprites 與 README 截圖
go run ./scripts
```

遊戲內快捷鍵：`Enter` 開始、`C` 建立角色、`J` 冒險手札、`T` 圖塊預覽、`G` GEO 預覽、`D` dungeon floor 預覽、`F5/F9` 儲存／載入 remake game。

## 尚未完成

完整 ECL opcode／routine、三城市各自的副本與城鎮 floor／tile mapping、完整場所與劇情、AD&D 全規則、音效音樂，以及原版 DOS save/import 仍在反組譯與實作中。戰鬥小人素材、CHEAD/CBODY party icon、SPRIT frame timing 與 frame offset 已接入目前 Ebiten combat slice，但方向-specific placement、八方向 placement 與完整戰鬥 UI 仍未完成；設定 `Area.InDungeon` 後，ECL `LOAD FILES` 能驅動 GEO map preview。現有 remake save 已能恢復已實作的 game state，現在也包含 dungeon preview 位置／方向；原版完整 save slot／game-area loader 與所有 file side effects 仍未完成。

更多證據與規格請見 [`CONTEXT.md`](CONTEXT.md)、[`docs/spec/`](docs/spec/)、[`docs/manual/`](docs/manual/)、[`docs/knowledge/`](docs/knowledge/) 與 [`docs/history.md`](docs/history.md)。
