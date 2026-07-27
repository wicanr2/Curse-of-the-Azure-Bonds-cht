# Curse of the Azure Bonds 中文化 & remake 計畫

## 目標

在不依賴原始 DOS 執行環境的前提下，建立可驗證的 Golden Box 資料／腳本格式規格，之後以 Go + Ebiten 重製執行層，並加入繁體中文介面、訊息與手札內容。

## SDD 工作規則

1. 反組譯、檔案格式分析與執行觀察先記錄於 `docs/spec/`。
2. 規格只有在證據、未知項目與驗證方法寫清楚後，才能標示 `READY`。
3. 實作每輪都必須有可重現的測試或分析工具。
4. 每輪更新 Markdown、`CONTEXT.md`，commit 並 push 到 GitHub。
5. 原始遊戲映像與掃描手冊只作本地分析素材，不預設重新散布到 repository。

## 分階段

| 階段 | 產出 | 狀態 |
|---|---|---|
| 0. 基線與素材盤點 | 本計畫、素材清單、GitHub 基線 | 完成 |
| 1. DAX／EXE／ECL 格式分析 | `docs/spec/` 格式規格與樣本工具 | 進行中 |
| 2. ECL 最小直譯器 spike | 可讀取並追蹤一個最小場景 | 進行中 |
| 3. 遊戲狀態與 AD&D 規則 | 核心模型與相容性測試 | 進行中 |
| 4. 渲染、輸入、音效 | Go/Ebiten 可執行 prototype | 進行中 |
| 5. 中文化與手札 | 字串資源、字型、繁中內容 | 待開始 |
| 6. 整合與遊玩驗證 | DOS 對照測試與 release build | 待開始 |

## 第一輪驗收

- 完成原始 ZIP 的完整 manifest 與格式初步分類。
- 確認 `START.EXE`、`GAME.OVR`、`ECL*.DAX`、圖像／地城資料的大小與檔案標記。
- 建立第一版格式研究規格，明確區分已知、推測與待驗證項目。

## 第四輪驗收

- Go DAX container parser 能驗證 header、block boundary 與 RLE 解碼。
- Go locale catalog 能載入 `assets/locale/zh-TW.json` 的繁中資源並支援英文 fallback。
- CLI 能直接讀取原始 ZIP 的 ECL block metadata。
- ECL opcode、畫面、音效與完整劇情仍未完成，不得宣稱 remake 已完成。

## 第六輪驗收

- `internal/ecl.Trace` 能依 command arity 追蹤 decoded block。
- 未知 opcode／截斷 operand 會安全停止並保留已完成 trace。
- `go test ./...` 與原始 ECL CLI trace 已通過。

## 第七輪驗收

- Go ECL layer 能從 length-prefixed payload 解碼 6-bit packed text。
- 以原始 `ECL1.DAX` 的真實 payload 做 regression test。
- CLI 能列出英文原文候選，作為後續繁中翻譯資源輸入。

## 第八輪驗收

- `internal/game.State` 以 locale catalog 驅動繁中開場狀態轉移。
- 狀態核心有錯誤 action 與完整 opening flow 測試。

## 第九輪驗收

- Ebiten command 能編譯並使用 `internal/game.State`。
- 啟動畫面、輸入與繁中 catalog 已連通。
- 字型以外部路徑注入，避免將未確認授權的字型提交至 repo。

## 第十輪驗收

- Ebiten opening 由原始 `ECL1.DAX` block 初始化。
- 原始 marker 與繁中顯示文字在 state 層分離保存。
- parser、state 與 Ebiten compile verification 通過。

## 第十一輪驗收

- `GOTO/GOSUB` code targets 可轉換成 payload offsets。
- 靜態 graph 對未知／越界資料安全停止。
- ECL branch graph 有單元測試與原始 block CLI 驗收。

## 第十二輪驗收

- [x] 以公開 CoAB 重寫程式核對 ECL 初始化順序。
- [x] 加入五組 word-valued ECL 初始化入口的 bounded parser。
- [x] 修正已觀察 VM command table 的 arity metadata 並加入 regression test。
- [x] 正確消耗 length-prefixed compressed-string operand，並支援從指定 entry offset trace。
- [x] 以 bounded smoke analyzer 執行 ECL1–ECL6 全部實際 block 的五個 initialization entries，記錄 menu／COMBAT／PROGRAM／unsupported opcode。
- [ ] 對已支援的實際入口逐一與事件文字、完整 input sequence 對齊。

## 第十三輪驗收

- [x] 建立有步數上限的 ECL subset runner。
- [x] 實作 `SAVE/COMPARE/IF/GOTO/GOSUB/RETURN/PRINT` 與 packed text 輸出測試。
- [x] 未支援 opcode 以精確 payload offset 停止，未宣稱完整 VM。
- [x] 實作 bounded `ON GOTO/GOSUB`、menu input；完整 DOS memory model 仍保留 boundary。
- [x] 依 reference `CMD_PartyStrength`／`CMD_PartySurprise` 消費 ECL party-rule command，並由注入的 `PartyContext` 將 verified roster stats writeback 到 shared ECL memory；完整 AC scale／multi-class rule table 仍保留 boundary。
- [x] 依 reference `CMD_CheckParty` 接入 thief skill／movement／active-affect branches 與四欄 memory writeback；未知 selector／完整作品 scaling 仍保留 boundary。
- [x] 依 reference `CMD_Who` 完成 prompt、roster UI、selected-player writeback 與 resumable ECL transaction；NPC／temporary party side effects 仍保留 boundary。

## 第十四輪驗收

- [x] 實作 `COMPARE AND`、`GETTABLE` 與 `HORIZONTAL MENU` bounded semantics。
- [x] 以真實 ECL1 initial entry 回歸開場三段文字與三個 menu 選項。
- [x] 將原始 menu 選項接入 game state 與繁中 locale／Ebiten rendering。
- [ ] 將 deterministic menu selection 改成玩家輸入，並接通兩個開場事件分支。

## 第十五輪驗收

- [x] 加入 successive ECL menu selection injection 與 regression。
- [x] 將 Ebiten directional cursor／Enter 接入 game state 與 ECL selection。
- [x] 以實際 ECL1 驗證 selection 0／1 走不同分支。
- [ ] 實作 `VERTICAL MENU` 與後續事件 command subset。

## 第十六輪驗收

- [x] 實作 `VERTICAL MENU` header／prompt／variable options parsing。
- [x] 支援 vertical menu selection injection 與 regression。
- [x] 實際 ECL1 讀到城市場所 menu options。
- [x] 將 successive menu sequence 保存到 UI state；城市場所與離開事件仍待完成。

## 第十七輪驗收

- [x] interactive runner 在未提供 selection 時暫停於下一個 menu。
- [x] game state 保存 successive menu sequence 並接入 Ebiten。
- [x] 新增城市 menu／continue prompt 的繁中 locale。
- [ ] 建立城市場所事件 regression，完成第一個可玩的場所功能。

## 第十八輪驗收

- [x] 新增 interactive CLI 以重現 selection sequence。
- [x] 驗證 ECL1 `0,0,1` 城市選擇序列。
- [x] 將 Shadowdale／Ashabenford／Dagger Falls 接入繁中 locale 與 state mapping。
- [ ] 將地點選擇寫入 map state，接通第一個場所入口。

## 第十九輪驗收

- [x] 新增 `LocationWilderness`／`LocationShadowdale` map state。
- [x] 以 ECL1 `0,0,1,0` 驗證 Shadowdale map-entry menu。
- [x] 將 `WILDERNESS`／`EXIT` 接入繁中 locale。
- [x] 實作 Shadowdale 資料中立移動與場所 menu contract。

## 第二十輪驗收

- [x] 將 Shadowdale location name 接入 locale 與 Ebiten UI。
- [x] 保留 Shadowdale 原始名稱與 map-entry menu state。
- [x] 實作 Shadowdale 座標／移動與第一層場所 menu；原始 tile 與場所內部功能仍待完成。

## 第二十一輪驗收

- [x] 實作 `NEWECL` bounded signal。
- [x] 加入 target block ID regression。
- [x] 建立跨 ECL1–ECL6 DAX block session 與 global block namespace。

## 第二十二輪驗收

- [x] 建立 decoded ECL `BlockSession`。
- [x] 驗證 ECL1 block `0x50/0x51/0x52` initial entries。
- [x] 驗證 NEWECL target switch contract。
- [x] 將全 ECL session 接入 game runtime，保存 bounded 跨 block state。
- [x] 補齊 BlockSession 跨 NEWECL 的 LOAD FILES／PICTURE／SPELL／PROTECTION signal aggregation。
- [x] 將 ECL SPELL／PROTECTION signal 接到 State pending queue 與一次性 consume API。

## 第二十三輪驗收

- [x] game command 載入 ECL1 全部 decoded blocks。
- [x] `NewStateFromECLBlocks` 接入 BlockSession。
- [x] State selection sequence 與 session offset 接通。
- [x] 以真實 NEWECL transition 做回歸，保存 bounded 跨 block memory／call stack。

## 第二十四輪驗收

- [x] `BlockSession.RunInteractive` 接入 game runtime。
- [x] selection offset 與 bounded NEWECL switch 有 synthetic regression。
- [x] 真實 ECL1 JOURNEY ON 的 PICTURE event 先停住，按 Enter 後續抵達 COMBAT boundary。
- [x] 確認 ECL1 initial-entry graph 尚無 reachable NEWECL edge，保留此未知。
- [x] 從其他 event entries 找出 ECL4／ECL5 real NEWECL transition。

## 第二十五輪驗收

- [x] 掃描全部 ECL 初始化 entries。
- [x] 定位 ECL4 block 0x25 `+0x022B` real NEWECL。
- [x] CLI entry-level regression 驗證 target 0x50。
- [x] 驗證 ECL5 block 0x30 `+0x0098` 的第二條 real NEWECL。
- [x] 將 Shadowdale `WILDERNESS/EXIT` 接成資料中立的 map entry／movement slice。
- [x] 將 `INN/STORE/BAR/LEAVE` 接成 Shadowdale `ModePlace` 與繁中 state event contract。
- [x] 將 `TREASURE` 8-operand framing 接入 bounded no-op，讓場所 trace 可安全繼續。
- [x] 將 `COMBAT` 接入 ECL／session／game state request signal。
- [x] 建立可注入骰點的 party／enemy combat core、initiative、AC、攻擊與傷害。
- [x] 解析 ECL `SETUP MONSTER`／`LOAD MONSTER` descriptor，驗證 Shadowdale encounter sequence。
- [x] 解碼 `MON*CHA` 固定 record offsets，建立 raw stats 到 `combat.Fighter` adapter。
- [x] 解碼 `MON*ITM` 63-byte 與 `MON*SPC` 9-byte raw records。
- [x] 依 reference `PoolRadPlayer.field_33`／`field_B5..B7` 保存 MON*CHA monster spell slots，並接入第一個已核對的 monster Magic Missile `0x0F`；其他 spell／AI／MON*SPC effect 仍保留 boundary。
- [x] 依 reference 將 `MON1SPC`–`MON6SPC` 以 chapter-local monster ID 掛到 enemy fighter raw `MonsterAffects`；效果 gameplay projection 仍保留 boundary。
- [x] 依 reference `CanHitTarget` 將 active monster invisibility `0x19`／`0x47` 投影為 combat target AC +4；其他 effect kinds 仍保留 boundary。
- [x] 依 reference `MON*CHA[0xA1]`／`AffectHaste`／`AffectSlow` 接入 monster base attacks count 與 active Haste `0x27`／Slow `0x2A` 倍率；movement timing 仍保留 boundary。
- [x] 依 reference `Player.IsHeld` 接入 enemy held turn skip 與 held-target guaranteed hit（`0x1F`／`0x33`／`0x34`／`0x35`）；cure／save／effect removal 仍保留 boundary。
- [x] 為已觀察 monster item/effect IDs 加入繁中顯示與未知 fallback。
- [x] 將真實 ECL spawn sequence 合併 MON*CHA，建立 24 個 enemy fighters 並在 COMBAT 邊界停止。
- [x] 建立可操作 party／enemy Battle state、回合攻擊、勝負轉移與 Ebiten 繁中戰鬥畫面。
- [x] 將 bounded ECL `COMBAT` result 的 encounter descriptors 與 `MON*CHA` records 接到 Battle，並提供 ECL1 direct-entry 驗證入口。
- [x] 對 `PROGRAM 0/3/8/9` 建立外部 routine boundary signal，避免把 CAMP／勝利／死亡流程錯誤當作 ECL no-op。
- [x] 加入繁中遊戲內冒險手札、中文歷史筆記，以及 `PROGRAM 9` 到 CAMP state 的可測試控制邊界。
- [x] 將冒險手札摘要做成八頁可翻閱的繁中遊戲內資料，接通 J／方向鍵／Esc 導航。
- [x] 保存 party roster、同步戰鬥 HP，並建立 CAMP 後 HP 恢復的 state boundary。
- [x] 建立七種玩家種族、六種基本職業、能力值與 1–6 人 roster validation。
- [x] 接通繁中角色建立 starter UI，完成後將角色投影並保存到 `State.SetParty`。
- [x] 接通 Unicode／繁中自訂角色姓名輸入與 State 保存。
- [x] 接通六項能力值的繁中編輯、3–18 bounds 與職業最低值 validation。
- [x] 加入六項能力值的 3d6 擲骰／重擲，並保留 seed regression。
- [x] 建立版本化 remake party JSON，接通 F5／F9 與啟動載入；原版 DOS save/import 仍待反組。
- [x] 將 ECL `COMBAT` 在有 party／MON*CHA records 時接到可操作 Battle；缺少資料時保留安全 boundary。
- [x] 以真實 ECL2 block 3 entry 3／MON2CHA records 建立可操作 encounter regression，並提供跨章節 `-encounter-monster-member` 入口。
- [x] 將正常 State COMBAT boundary 的 MON*CHA lookup 依 ECL1–ECL6 chapter namespace 分流，不再只使用 MON1CHA。
- [x] 以原始 ECL1 block `0x50`／ECL2 block `3` 驗證跨 DAX `NEWECL` transition，並保留 target 後續 bounded stop。
- [x] 以真實 ECL1 block 0x51 的 `JOURNEY ON → STORE` 建立 COMBAT boundary regression，確認缺 descriptor 時不虛構 Battle。
- [x] 實作 ECL `RANDOM` 與 State 可注入 seed，保留 deterministic regression。
- [x] 實作 ECL `ENCOUNTER MENU` operand framing、selection pause 與 memory action mapping。
- [x] 保存 ECL menu pause 的 PC、memory、比較旗標、call stack，並以 cumulative selections resume。
- [x] 將 bounded runtime context（memory／flags／call stack）帶過 `NEWECL` target block，並建立跨 block regression。
- [x] 解碼 `GEO2–GEO6` 的 16×16 geometry planes，建立原始 map cell data layer。
- [x] 解碼 `TILES.DAX`／`8X8D*.DAX` indexed pictures 與 `WALLDEF*.DAX` records。
- [x] 建立 reference EGA16 palette 與 indexed picture → RGBA adapter。
- [x] 將 `TILES.DAX` indexed pictures 接入 Ebiten，建立繁中原始 tile gallery preview。
- [x] 將原始 `GEO2` geometry block 接入 Ebiten raw geometry viewport。
- [x] 以 GEO wall fields 建立 shared `CanMove` navigation contract 與 viewport cursor regression。
- [x] 建立 reference `BackGroundTiles` 74 筆 metadata layer，保留 floor construction 的待實作邊界。
- [x] 建立根目錄 README 與由原始 DAX parser 可重現的 TILES／GEO PNG 截圖證據。
- [x] 還原 50×25 wilderness floor construction，接通 background entry → TILES index 與 movement cost。
- [x] 還原 GEO dungeon floor 四段 tile composition，接通 D 預覽與可重現 dungeon screenshot。
- [x] 接回 dungeon `terrain & 0x40` table／chair decoration pass 與 seeded regression。
- [x] 建立 GEO2–GEO6 全 16 block 的原始 map catalog，並接通 set/block preview selector。
- [x] 將 ECL `LOAD FILES` 第三 operand 接到 State pending GEO map request 與 renderer catalog。
- [x] 建立 Area1／Area2 `inDungeon`、`game_area` 與 LOAD FILES branch contract。
- [x] 建立 Area1／Area2 已知欄位 binary codec，並保留未知 bytes。
- [x] 將 remake F5/F9 擴充為可恢復 Area、location、mode 與 map 座標的版本 2 game save。
- [x] 將 remake F5/F9 擴充為 version 3 dungeon 3D preview 位置／方向 save state，並保留 v1/v2 舊檔安全 defaults。
- [x] 將 dungeon preview Q/E 接成 reference 八方向 facing rotation，重建 Far／Mid／Near wall view 並覆蓋 wrap regression。
- [x] 將 reference dungeon 16×16 coordinate wrap 接成明確 wrapped GEO／wall traversal API，讓 preview 跨邊界移動並保留 strict API。
- [x] 將 reference 5-byte map state 的 `mapWallType/mapWallRoof` 接入 version 4 remake save，並由 wrapped GEO refresh 重算。
- [x] 解碼 Area1 `0x1FA/0x1FC` indoor/outdoor sky colour，依 mapWallRoof high bit 接入 dungeon preview sky layer。
- [x] 依 reference `WallDoorFlagsGet` 接入 GEO no-wall default／walled x3 detail adapter 與 preview evidence。
- [x] 依 reference `MapSetDoorUnlocked`／`TryStepForward` 接入 unlocked detail `1` 的 dungeon movement 與雙側 raw unlock adapter。
- [x] 從 DOS player record `0xEA–0xF1` 保存八項 thief skills，並提供 `open_locks` party adapter。
- [x] 將 opening city menu routing 到暗影谷、阿沙本福德、匕首瀑布，並使用對應 wilderness city flags。
- [x] 從 CPIC1–CPIC6 抽出戰鬥小人 PNG、manifest 與可重現 sprite sheet。
- [x] 將 CPIC block mapping 與戰鬥場景 Ebiten renderer 接通。
- [x] 解碼 `SPRIT*.DAX` frame stream 並抽出逐幀 PNG。
- [x] 將 SPRIT animation timing 與 ECL `SETUP MONSTER` block 接入 renderer。
- [x] 解析 CHEAD／CBODY masked layers，依 reference `MergeIcon` 規則產生 party combat icon 並接入 renderer。
- [x] 將玩家 icon state、normal／attack layer 與 direction flip 接入 renderer boundary。
- [x] 反組並接入新建角色 head／weapon block 0 與 race-based icon size defaults。
- [x] 將 SPRIT frame x/y position metadata 寫入 manifest 並接入 combat renderer。
- [x] 解碼 PIC1–PIC6 的 PIC/FINAL first-frame XOR delta，並加入共用 animation parser regression。
- [x] 將 ECL `PICTURE` block request 接到可恢復繁中 event screen 與 PIC playback。
- [x] 將 ECL PICTURE `>=0x78` 分流到 BIGPIC unmasked static picture extraction／event renderer。
- [x] 抽取 HEAD/BODY scene layers，依 reference body y+5 合成並接入共用 asset loader。
- [x] 將 Area2 HeadBlockId sentinel 接入 ECL PICTURE 的 HEAD/BODY scene branch。
- [x] 將 Area2 `HeadBlockId @ 0x5C2` 接入 codec 與 game state sync。
- [x] 建立 combat 八方向 delta／tile placement contract，並接入目前 Ebiten formation renderer。
- [x] 將 CombatMap position／size 接入 Fighter、StartCombat 與 renderer boundary。
- [x] 反組譯並封裝 reference `try_place_combatant` position formula。
- [x] 保存 DOS player record `icon_id` metadata 到 party／combat projection。
- [x] 將城市 `INN` 接成安全休息、HP restore 與繁中返回場所流程。
- [x] 建立商店 Buy／Sell／ID 的 price-injected party transaction contract。
- [x] 將城市 `STORE` 接成繁中 Shop Menu command state，保留 stock／money-pool 邊界。
- [x] 建立 injected shop stock、party money pool、POOL／TAKE／SHARE 與 BUY state API。
- [x] 將 BUY 接成繁中商品清單、價格顯示、pool 扣款與 inventory 更新。
- [x] 將 SELL 接入繁中 Shop Menu：角色／物品選擇、item `Value` 入 pool、readied／cursed protection 與 party projection。
- [x] 將 ID 接入繁中 Shop Menu：角色／物品選擇與已確認 200 GP fee；保留 identification result data boundary。
- [x] 將 VIEW 接成繁中角色 HP／金幣／裝備摘要與返回 Shop Menu。
- [x] 將 TAKE 接成繁中角色／金額選單、pool 扣款與角色 gold 更新。
- [x] 將 APPRAISE 接成繁中角色／寶石選單、injected offer 與 pool 入帳。
- [x] 補上 APPRAISE 報價的接受／拒絕／返回 confirmation branch。
- [x] 建立繁中 CAMP Menu（SAVE／VIEW／MAGIC／REST／ALTER／FIX／EXIT），接入 REST／EXIT state boundary。
- [x] 將 CAMP VIEW 接成角色選單、繁中角色摘要與可返回 CAMP Menu 的只讀流程。
- [x] 將 CAMP MAGIC 接成角色選單、已記憶 spell-slot 查看與已核對的一級法術名稱；保留 spell rules／mutation boundary。
- [x] 將 CAMP MAGIC 接成 RuleBook 證實的 `CAST／MEMORIZE／SCRIBE／DISPLAY／REST／EXIT` command routing；保留 CAST／MEMORIZE／SCRIBE rules boundary。
- [x] 將 CAMP MEMORIZE 接成 known-spell selection、pending state 與 REST_START 的 SpellSlots writeback；保留完整 capacity／時間／中斷規則 boundary。
- [x] 對目前已核對的一級法術加入 RuleBook 證實的最低準備時間檢查；不足休息時保留 pending selection。
- [x] 將 RuleBook 證實的一級 Magic Missile 接入 Battle、slot consumption、繁中 Ebiten S 鍵與 deterministic damage。
- [x] 將 RuleBook 證實的一級 Cure Light Wounds 接入 Battle、slot consumption、受傷隊員 bounded target、繁中 Ebiten H 鍵與 deterministic healing。
- [x] 將已接入戰鬥法術改為 CAST target-selection state，接入左右切換、Enter 確認與 Esc 取消。
- [x] 將 RuleBook 證實的 Bless 接入戰鬥、牧師 slot、B／Enter confirmation 與 party attack bonus；保留鄰近怪物與 duration boundary。
- [x] 將 Bless 的 RuleBook `6r` duration 與 CombatMap 八方向相鄰排除接入 Battle；缺少位置資料時保留 bounded fallback。
- [x] 將 RuleBook Curse（spell ID `2`）接入牧師 C／敵方 target selection、slot transaction、相鄰排除與 6 回合攻擊 debuff。
- [x] 將 RuleBook Cause Light Wounds（spell ID `4`）接入牧師 W／touch target selection、slot transaction 與 deterministic 1d8 damage。
- [x] 將 RuleBook Protection from Evil（spell ID `6`）接入牧師 P／party touch target、conditional evil-attacker AC +2 與 `3r/lvl` duration；保留 saving throw／alignment import boundary。
- [x] 修正 class-local spell ID collision，並將 RuleBook Protection from Good（牧師 ID `7`）接入 G／party target／conditional good-attacker AC +2；Magic Missile（魔法師 ID `7`）保留 S 分流。
- [x] 將 RuleBook encounter `FLEE` 接成繁中可恢復事件，`PARLAY` 接成五種 tactic menu；保留速度／追擊／reaction／script boundary。
- [x] 將 RuleBook Combat Menu `MOVE` 接成 M／方向鍵單格移動、occupancy-safe position mutation 與 party turn consumption；保留地形／free attack／facing boundary。
- [x] 將 RuleBook MOVE 的離開鄰接範圍 free attack 接入 Battle／State；保留背面 AC、facing、reach 與地形 boundary。
- [x] 將 RuleBook MOVE 的移入敵格攻擊接入 Battle／State；保留地形、邊界、負重、facing、reach 與動畫 boundary。
- [x] 將 RuleBook active-character centered camera 接入 CombatMap／State／Ebiten renderer；保留 viewport、scroll animation 與真實 Area camera boundary。
- [x] 將 RuleBook Combat Menu `VIEW` 接成繁中 read-only fighter summary、V／Enter／Esc 與不消耗回合的 State transaction；保留完整 View Menu／物品／交易 boundary。
- [x] 將 RuleBook／ITEMS RateOfFire 接成 equipped fighter 多次攻擊與目標倒下後的下一目標 transaction；保留彈藥、等級額外攻擊、Aim／range boundary。
- [x] 將 ITEMS raw AmmunitionType 投影到 fighter，建立注入式彈藥 mapping、atomic stack consumption 與 CombatAct preflight；保留特殊彈藥、裝填與 ranged line-of-sight boundary。
- [x] 將 RuleBook Combat Menu `DONE` 接成 D／繁中 no-attack turn transaction，重用 enemy／next-party advancement；保留 hold／delay 與其他 command boundary。
- [x] 將 RuleBook armor movement table 接入 `Fighter.MovementAllowance` 與多格 MOVE turn transaction；保留負重、地形、邊界與 FLEE speed boundary。
- [x] 將 RuleBook missile adjacent prohibition 與 dart thrown exception 接入 weapon profile／Battle guard；保留完整 Range、line-of-sight、Aim 與其他 thrown weapon boundary。
- [x] 將 missile adjacency guard 接到 ammunition transaction 前的 attack preflight，避免無效攻擊消耗彈藥；保留多目標 ranged transaction boundary。
- [x] 將已投影的 `Fighter.AttacksPerTurn` 套用到 enemy turn；保留 enemy AI、彈藥與額外職業攻擊 boundary。
- [x] 將玩家 combat action error 接成繁中可恢復訊息，避免非法輸入結束 Ebiten game loop；保留資料／啟動錯誤向上回報 boundary。
- [x] 將 ECL `ADD NPC` 接成 NPC ID signal，讓實際 ECL1 block 0x52 能安全走到 EXIT；保留 NPC table／party side effect boundary。
- [x] 將 ECL `LOAD PIECES` 接成三 selector signal，保留地城 floor／wall／tile file side effect boundary。
- [x] 將 `LOAD PIECES` signal 保存為 State 一次性 request，保留 renderer／map adapter boundary。
- [x] 依 reference `LoadWalldef` 將 `LOAD PIECES` selector 接到 WALLDEF／8X8D raw piece catalog，保留牆面 renderer boundary。
- [x] 依 reference `WallDefBlock.Offset` 將 WALLDEF graphic IDs 映射到 global 8×8D symbol item，提供 bounded `WallSymbol` lookup。
- [x] 依 reference `draw_3D_8x8_titles` 建立十組 wall viewport layout metadata、`WallStamp` 與 dungeon preview sample。
- [x] 依 reference `Draw3dWorldFar/Mid/Near` 建立 Far／Mid／Near GEO wall traversal 與 ordered `WallLayoutCall`。
- [x] 將 dungeon preview position 接到 GEO-safe movement，移動後重建 floor 與 Far／Mid／Near wall stamps。
- [x] 將 CAMP SAVE 接成 state request 與 Ebiten party-save adapter；保留原版 SAVGAM container boundary。
- [x] 將 CAMP ALTER／ORDER 接成兩階段角色重排，並同步 party roster 與 combat fighter 順序。
- [x] 將 CAMP ALTER／DROP 接成二次確認的永久角色移除，並同步 party roster 與 combat fighter。
- [x] 將 CAMP ALTER／PICS 接成怪物圖片／動畫 runtime preference，並接到事件／戰鬥 renderer。
- [x] 將 CAMP ALTER／SPEED 接成 1–5 級訊息速度，並接到 Ebiten Unicode message reveal。
- [x] 將 CAMP ALTER／ICON 接成已驗證 CHEAD／CBODY block 選擇，並同步 roster／combat fighter icon。
- [x] 將 CAMP FIX 接成已驗證 Cure Light Wounds spell-slot、`1d8` healing 與 roster／combat fighter HP sync boundary。
- [x] 將城市 BAR 接成 ordered Tavern Tale menu、繁中內容與城市場所返回 boundary；買酒價格／ECL trigger 保留 data adapter。
- [x] 將 CAMP REST 接成 `REST ADD SUBTRACT EXIT`、24 小時自然 HP recovery 與 CAMP menu boundary；spell memorization／中斷保留 adapter。
- [x] 解碼 DOS player runtime CHEAD／CBODY icon slot mapping 與 small `+0x40` namespace；真實 CombatMap position／camera 與 direction-specific placement仍待完成。
- [x] 將 CHEAD／CBODY attack `+0x80` namespace 接入 on-demand combat icon composition；direction-specific placement／recolor／runtime cache 仍待完成。
- [x] 依 reference `HalfDirToIso` 將 combat map direction 接到 party／enemy IconDirection 與 flip adapter；完整 Area/ECL direction source 與 placement 仍待完成。
- [x] 解析原始 `ITEMS` base-item descriptor table，建立可重用 catalog 與已知繁中名稱。
- [x] 將已知 `ITEMS` readied 武器／護甲效果投影到 party fighter；charges、魔法效果與商店仍待完成。
- [x] 建立 party equipment class mask、slot collision、雙手武器與雙戒指 transaction contract。
- [x] 建立 inventory Count stack／readied protection／cursed lock mutation contract。
- [x] 解碼 scroll／potion／wand properties，建立 consumable use signal 與 charge mutation。
- [x] 將 ECL `SPELL`／`PROTECTION` operand 接成 bounded runtime signal。
- [x] 建立 remake party ordered spell-slot resolver，並將 ITEMS catalog 接入 character creation／party load。
- [x] 解析公開 DOS player record 的 memorized／known spell 欄位，並接到 party spell-slot adapter。
- [x] 將 DOS known-spell flags 保存到 Character／party save，並在 CAMP MAGIC 顯示已記憶／可用數量。
- [x] 解析公開 DOS player record 的單職業核心欄位，接到 party／combat projection。
- [x] 將已保存的 DOS open-locks skill 接成注入 d100 的 pick-lock transaction 與 Knock `0x1F` slot 消耗核心；door UI／bash 仍待完成。
- [x] 將 pick-lock／Knock 接到 dungeon preview P/K action adapter，成功後呼叫 GEO 雙側 unlock；完整 locked-door menu／bash 仍待完成。
- [x] 依 reference strength／exceptional tables 將 bash-door dice resolver 接到 preview B；完整 locked-door menu 與撞門 side effects 仍待完成。
- [x] 依 reference `locked_door` 將 detail 2/3 action capabilities 接到方向鍵阻擋後的 preview menu；完整 DOS 視窗／door graphics／劇情 entry 仍待完成。
- [x] 依 reference `seg001.Init` 修正 dungeon 預設／save fallback 為 `(7,13,0)`；`InitAgain` restart direction 與完整 SAVGAM context 仍待完成。
- [x] 解碼 DOS `.SWG` item records，接到 party equipment projection。
- [x] 解碼 DOS `.FX` 9-byte effects，接到 party effect preservation 與繁中名稱。
- [x] 修正 `.FX` duration／strength 欄位並建立 finite/permanent duration tick adapter。
- [x] 依 reference `ovr021.step_game_time` 接入七-slot raw clock、slot scaling 與 party／active battle effect timeout；age writeback、REST interruption 與完整 calendar 仍保留 boundary。
- [x] 將 `REST_START` 接到 slot-1 game-time advancement（每小時 60 分鐘），在自然回血前到期 finite effects；random interruption／safe location／spell-learning side effects 仍保留 boundary。
- [x] 建立 `.SAV/.GUY` + optional `.FX/.SWG` DOS player sidecar bundle importer。
- [x] 提供 DOS character bundle → versioned remake party JSON 的可重現 CLI。
- [x] 將 DOS character bundle 接入 Ebiten remake startup bridge。
- [x] 將 imported active Bless／Curse effects 投影到 combat fighter attack bonus。
- [x] 將 active Blind／Bestow Curse／friendly Prayer effects 投影到 fighter attack／AC。
- [x] 依 `ovr017.SaveGame/loadSaveGame` 建立 `SAVGAM` 固定前綴 raw codec 與 round-trip regression；完整 slot／Area 欄位／player-file side effects 仍待接入。
- [x] 將 `SAVGAM` fixed prefix 接到 State 的 Area／map load／export adapter，保留未知 raw segments；尚未取代 remake JSON。
- [x] 依 reference `seg044`／`Resource.resx` 抽出 9 個 PC WAV，建立共用 sound selector catalog 與 Ebiten playback adapter。
- [x] 將 State sound intent 接到武器命中／未命中、擊倒、免費反擊與已實作法術；背景音樂與完整 ECL sound calls 仍待接入。
- [x] 依 reference `SaveGame/loadSaveGame` 將 SAVGAM slot prefix 與 `CHRDAT*.sav`／optional `.fx/.swg` 接到 State loader，並提供 `-savgam-dir/-savgam-slot` 啟動入口。
- [x] 將已完成的 `SAVGAM` slot load path 延伸到已證實 Player fields／`.swg`／`.fx` writeback，使用 sibling staging directory 後逐檔替換；保留未知 `.sav` bytes。
- [x] 將 loaded SAVGAM slot writer 接到 Ebiten F5 與 CAMP SAVE；一般 remake JSON save path 維持不變。
- [x] 將戰鬥結束的 battle HP 同步回 party roster，供 CAMP／remake save／SAVGAM writeback 使用。
- [x] 將戰鬥結果 continuation 接回荒野主選單，清除 stale ECL／戰鬥前 choices。
- [x] 依已知 `SaveGame` side effect 清理該 slot 的 stale `CHRDAT*.sav/.swg/.fx`，並以 backup/rollback 保護 staged replacement。
- [ ] 補齊原版 player delete／rename semantics、多職業與未知欄位證據，及完整 player serialization。
- [ ] 從完整玩家流程抵達該 entry，驗證跨 block context 的完整劇情 continuation。
- [x] 依 CoAB reference 與 ECL2 block 3 real scan 建立 `DAMAGE` 五 operand raw signal；target／saving throw／HP mutation 保留 party adapter boundary。
- [x] 將 ECL `DAMAGE` signal 接入 State pending queue 與 exactly-once consume；保留 selected-character／save-throw／HP mutation boundary。
- [x] 保存 DOS player `saveVerse` `0xDF–0xE3` 到 Character／JSON／writeback adapter。
- [x] 依 CoAB reference 接入 ECL `DAMAGE` selected／whole-party branches、注入骰點、save resolution 與 roster／fighter HP sync；random-target／death continuation 保留 boundary。
- [x] 依 reference `CanHitTarget` 接入 ECL `DAMAGE` random target count、target order、natural 1／20 hit resolver 與 State HP sync；AC／affect projection／death continuation 保留 boundary。
- [x] 保存 DOS player `field_186 @ 0x186` signed saving bonus，接入 ECL DAMAGE save resolution 與 record writeback。
- [x] 建立 State default hit resolver，投影 fighter／equipment AC 並套用已證實的 invisibility `0x19`／`0x47` -4、action-delay-aware blink `0x25` 與 displace `0x59` consumed-bit；effect mutation 具 transactional rollback，其他 affect／death continuation 仍待補證。
- [x] 依 reference `damage_player` 將 ECL DAMAGE exact-zero／overkill／animated cases 投影到 Character.HealthStatus 與 DamageOutcome；完整 Death routine、bleeding 與 party win/loss continuation 仍待接入。
- [x] 將 active combat 的 ECL DAMAGE HP 經 `Battle.SetHitPoints` 同步並重新計算 party／enemy status，status 結束時接既有 `finishCombat`；完整 Death routine、bleeding 與 effect removal 仍待接入。
- [x] 依 reference `RemoveCombatAffects` 清理 active-combat 倒下角色的 19 個 combat-only effect kind；blink／invisibility 保留，`CheckAffectsEffect(Death)` 與 bleeding 仍待接入。
- [x] 依 reference `CheckAffectsEffect(Death)` 接入 affect_63 recovery、Bleeding、troll_fire_or_acid damage-flag gate／TrollRegen `0x66`，explicit dragon-slayer `0x4B` target context，以及 HP=0 時的 combat position removal、renderer-neutral `DeathOverlay` signal、team `DownedCorpse`／`Tile_DownPlayer=0x1F`、`COMSPR 0x8B`／`0x19` skull sprites、可治療倒下隊員的 Cure Light Wounds、explicit `CombatHealAllowed`→`RestoreCombatant` stand-up、9-cycle `DeathOverlayFrame` lifecycle、enemy post-flash render removal、`CombatAction` cleanup 與 NewBattle 初始倒下正規化；其他 Death routine 仍待接入。
- [x] 依 reference `find_target`／`BuildNearTargets` 建立 seeded enemy target-selection contract：敵方回合從存活 party 選擇目標，同一回合多次攻擊固定 target；visibility／pathfinding／AI spell priority／guarding 保留 boundary。
- [x] 依 reference `CMD_EclClock` 修正 ECL CLOCK 為兩 operand，跨 BlockSession 聚合並接到 State `AdvanceGameTime`；完整 memory-backed clock values／time-triggered event table 仍待驗證。
- [x] 將七-slot game clock／age cycles 接入 remake JSON save version 5，保留 versions 1–4 相容載入；DOS SAVGAM Area1 raw clock offset 仍待獨立驗證。
- [x] 依 reference Area1 `0x18C..0x198` 將七個 raw clock words 接回 Area codec、SAVGAM load 與 State synchronization；完整原版日曆規則仍待追蹤。
- [x] 依 reference `display_map_position_time` 將七-slot clock 接成 renderer-neutral 繁中 `HH:MM`／日曆 HUD，並顯示於一般畫面與荒野地圖；完整原版日曆規則仍待追蹤。
- [x] 依 reference `Player.age @ 0x76` 與 `NormalizeClock` 接入 DOS age import／writeback、slot-6 overflow 年齡增加與 regression；Pool/Rad `0x30` 與 age-based ability modifiers 仍待獨立驗證。
- [x] 依 reference `StatValue.AgeEffects` 建立五段 race bracket／六項 ability delta 的明確 `WithAgeEffects` adapter；避免對已含 age-adjusted stats 的 DOS import 重複套用，creation UI／class limits 仍待接線。
- [x] 依 reference `race_ages`／`ovr018` 建立 single-class starting-age base+dice resolver；多職業與完整 creation UI 仍待接線。
- [x] 將 starting-age／age ability effect 接入 `State.AddCreationCharacter` copy transaction；多職業與 alignment UI 仍待接線。
- [x] 將已驗證的 22 個 single-class race/class 組合接入角色建立選單，並加入五列捲動顯示；多職業與完整原版建立流程仍待反組譯。
- [x] 讓 ECL `COMBAT` 保存 next-PC，並在 party victory 後續跑同一個 resumable ECL session；menu／picture／NEWECL continuation 已有 synthetic／real regression，完整各 block routine 仍待反組譯。
- [x] 將 `CAMP → MAGIC → CAST` 接入施法者／memorized slot／受傷目標選單與 Cure Light Wounds `1d8` transaction；SCRIBE、其他法術與完整 slot／時間規則仍待接入。
- [x] 將 `CAMP → ALTER → RENAME` 接入 15-byte DOS name editor、roster／fighter projection 與 SAVGAM raw-preserving name writeback；Big5 transcoding、多職業與完整 delete semantics 仍待接入。
- [x] 依 `ovr003.CMD_LoadCharacter` 解碼 ECL `LOAD CHARACTER` 的 1-based selector／bit-7 flag，接回 State persistent roster 的 selected player；完整 external string／party-summary side effects 仍待反組譯。
- [x] 依 `vm_CopyStringFromMemory` 將 selected player name 投影到 resumable ECL `0x7C00` string slot，讓 `LOAD CHARACTER → COMPARE/IF` 姓名分支可執行；其他 DOS memory-string regions 仍待驗證。
- [x] 依 `CMD_FindItem` 將全隊 raw item type 投影到 ECL party context，設定 `=`／`<>` compare flags，並讓同-run `DESTROY ITEMS` 更新 working query view。
- [x] 依 `CMD_FindSpecial` 保存 resumable selected-player index，讓 LOAD CHARACTER／WHO 後續查 selected member active affect 並設定 `=`／`<>` compare flags。
- [x] 將原始 ECL1–ECL6 的 25 blocks／125 initialization entries 建成無 unsupported-opcode 的 real-image corpus gate，並驗證 ECL5 日光腐朽的 real FIND ITEM found branch。
- [x] 依 `CMD_Dump`／`FreeCurrentPlayer` 實作 selected member 離隊、working-party更新、fallback selection 與 State roster／fighter同步；鎖定 real ECL5 Akabar DUMP。
- [x] 依 `CMD_Program` 將 0/3/8/9 接成共用 State external-routine adapter，涵蓋 start menu、party killed、game won 全隊恢復／存檔選擇與 CAMP，並由戰鬥後 continuation 共用。
