# 第 1–347 輪：早期工作序列與成果

從 ECL/DAX 取樣到第 347 輪原版 demo oracle 的累積工作項與輪次紀錄。

> 本檔由 2026-08-13 的 `CONTEXT.md` 分冊而來，內容逐行保留原文。
> 歷史敘述以當時證據為準；與目前 worktree、READY spec 或
> `docs/knowledge/coab-re-coverage-matrix.md` 衝突時，以後者為準。

## 下一步

1. 建立可重現的映像 manifest／樣本檢查工具。
2. 對 `ECL*.DAX`、`GAME.OVR` 做十六進位與反組譯取樣，找出共通 header／索引。
3. 依證據更新 `docs/spec/`，未驗證內容維持 `DRAFT`。
4. 建立 ECL block 的欄位／事件邊界測試，才評估是否能把 ECL 規格標為 `READY`。
5. 將 ECL decoded payload 分成事件 header、控制資料與文字，建立最小場景 trace。
6. 以 operand framing 為基礎加入未知 opcode 安全停止的 trace walker。
7. 對齊 ECL branch targets 與原版場景入口，建立最小可執行事件狀態。
8. 對全部 ECL 建立去重原文 catalog，開始逐項接入繁中翻譯。
9. 建立 Ebiten input／render adapter，先視覺化 opening state。
10. 將 ECL event trace 與 Ebiten event screen 接起來，加入第一個可驗證劇情場景。
11. 建立 ECL branch target graph，將 opening marker 後的選項與事件序列接入 state。
12. 對 ECL graph 的 entry points 做原版事件文字對齊，建立第一個完整 event screen。
13. 用 `EntryPoints` 對實際 ECL1–ECL6 做入口 regression，再逐步加入可執行 VM command subset。
14. 將 `ON GOTO/GOSUB`、選單與 bounded memory model 以 regression 驗證後接入遊戲 state；完整 DOS memory model 仍保留 boundary。
15. 將 menu selection 變成 Ebiten input／runner action，完成 Enter City／Journey On 第一個事件分支。
16. 實作 `VERTICAL MENU` 的可觀測選項與 input，再擴展第一個城市事件。
17. 將 successive menu sequence 保存為 game event state，完成城市場所選擇與離開分支。
18. 建立城市場所的 ECL event regression，接入第一個可玩的場景功能。
19. 將城市選擇寫入 map state，完成第一個地點（SHADOWDALE）的場所入口。
20. 接通 Shadowdale `WILDERNESS/EXIT` 後的移動／場所 menu。
21. 建立第一個場所的 map／ECL event state。
22. 建立 ECL1–ECL6 DAX block session，驗證 NEWECL target。
23. 將 BlockSession 串入 game runtime，保存跨 block event state。
24. 建立 real NEWECL transition regression，補齊跨 block memory／call stack。
25. 掃描各 ECL event entry，找出可達 NEWECL transition 並建立實際 regression。
26. 將 ECL4 real transition 接入可導航的 event entry，補齊跨 block memory／call stack。
27. 解碼原始地圖 tile／座標規則，並接入 ECL1 block 0x51 的完整場所 command path。
28. 實作場所內部的角色、交易、休息、情報與 AD&D 規則。
29. 解碼 TREASURE table 與 party inventory 效果，補真實 event regression。
30. 建立 party／enemy combat model，接入 AD&D 回合與戰鬥 UI。
31. 將 ECL `LOAD MONSTER`／`SETUP MONSTER` 接到 combat fighter 與 battle map。
32. 解碼 `MON1CHA` 等 monster records，建立 ECL-to-combat adapter。
33. 接入 `MON*ITM`／`MON*SPC`，將 ECL spawn sequence 建立成 Battle。
34. 建立 item type／name-number catalog，合併 monster equipment/effects 並建立 Battle setup。
35. 擴充完整 item catalog，整合 locale 與 ECL-to-Battle equipment/effects。
36. 建立 party fighter 與 Battle，接入 encounter equipment/effects 與戰場 UI。
37. 將 ECL `COMBAT` result 的 spawn descriptors／MON*CHA records 自動接入 `State.StartCombat`。
38. 由完整 opening／城市／地圖流程抵達 ECL1 encounter，移除 debug party 依賴。
39. 整理完整 Adventure Journal 條目、接入 CAMP 恢復／中斷規則與 party state。
40. 將 Journal Entry／Tavern Tale 做成資料驅動的已讀條目與劇情觸發。
41. 反組 party creation/import 與完整 CAMP 恢復／中斷規則。
42. 將 `internal/party` 接入繁中角色建立 UI 與 State.SetParty。
43. 補自訂姓名／能力值、裝備、XP／等級與 party save/import。
44. 接能力值逐項分配與規則化的角色完成流程。
45. 補擲骰／重擲點數池、性別／年齡與 alignment，接角色存檔。
46. 補裝備選擇、XP／等級與 party save/import。
第 89 輪功能／文件 commit：`d389002`，已推送至 GitHub `main`。依 `ovr011.PlaceCombatants` 將遭遇距離與八方向 `mapDirection` 的 team origin／facing group 封裝為 `combat.EncounterTeamStart`，刻意未假設尚未解出的 occupancy／candidate ordering。

第 90 輪功能／文件 commit：`0151599`，已推送至 GitHub `main`。原始 `ITEMS` 解析與 128-type 繁中 catalog 已由實際 ZIP 驗證。

第 91 輪功能／文件 commit：待本輪提交。新增 `monster.ItemRecord.Effect`、packed AC／base damage adapter 與 `party.Character.FighterWithEquipment`；只套用 readied 基本武器／護甲，charges、Affects／法術、雙持、彈藥與完整 DOS inventory 仍明確保留未完成。完成 Docker 測試後提交。

第 91 輪功能／文件 commit：`ebc8715`，已推送至 GitHub `main`。ITEMS readied 武器／護甲效果已接入 party fighter，舊 JSON 行為保持相容。

第 92 輪功能／文件 commit：`fe5edbd`，已推送至 GitHub `main`。party class／slot transaction 已完成，stack、cursed 與完整 inventory mutation 待後續處理。

第 93 輪功能／文件 commit：`5cdb800`，已推送至 GitHub `main`。inventory stack mutation 與 cursed readied lock 已完成。

第 94 輪功能／文件 commit：`5834646`，已推送至 GitHub `main`。scroll／potion／wand consumable data signal 與 inventory mutation 已完成。

第 95 輪功能／文件 commit：`a604bcc`，已推送至 GitHub `main`。ECL SPELL／PROTECTION bounded signals 與 regression 已完成。

第 96 輪功能／文件 commit：`0c1a122`，已推送至 GitHub `main`。新增 `Character.SpellSlots`、`Roster.FindSpell`、`State.ResolveSpellSearch`，並由 Ebiten bootstrap 載入原始 `ITEMS`，使 creation／party load 的 readied equipment 進入 fighter projection；原始 DOS spell offsets 與 ECL memory writeback 仍未完成。

第 97 輪功能／文件 commit：`2280184`，已推送至 GitHub `main`。新增 bounded DOS player spell record parser、truncated-record guard、known／memorized tests 與可重用知識庫規格。

第 98 輪功能／文件 commit：`43b26d0`，已推送至 GitHub `main`。新增 `ParseDOSPlayerRecord`，解析公開 `.SAV/.GUY` 單職業核心欄位，並保留原始 HP、icon、金幣與 spell 到 party projection；`.SWG` inventory／`.FX` effects／多職業與完整 DOS save container 仍未完成。

第 99 輪功能／文件 commit：`938d2a7`，已推送至 GitHub `main`。新增 DOS player item/effects pointer preservation、`.SWG` `0x3F` item stream adapter 與 party equipment projection；`.FX` effects、pointer address-space 與完整 save container 仍未完成。

第一百輪功能／文件 commit：`5bcea79`，已推送至 GitHub `main`。新增 DOS `.FX` 9-byte effects stream adapter、Character preservation 與常見效果繁中名稱；effect gameplay tick／解除與完整 save container 仍未完成。

第一百零一輪功能／文件 commit：`2179dd6`，已推送至 GitHub `main`。修正 `.FX` 16-bit duration／strength 欄位語意，新增 finite/permanent duration tick 與 party adapter；effect-specific gameplay 與 CAMP／戰鬥時間接線仍未完成。

第一百零二輪功能／文件 commit：`0332a09`，已推送至 GitHub `main`。新增 `ParseDOSPlayerFiles` sidecar bundle importer，並將 gold/gems/jewelry 保存到 `Character`；`SAVGAM?.DAT` party／area container 仍未解析。

第一百零三輪功能／文件 commit：`6497c18`，已推送至 GitHub `main`。新增 `cmd/azure-bonds -import-character`，把已證實的 `.SAV/.GUY` + optional `.FX/.SWG` bundle 輸出成 versioned remake party JSON。

第一百零四輪功能／文件 commit：`d3bc6e9`，已推送至 GitHub `main`。新增 `cmd/azure-bonds-game -dos-character-record` startup bridge，直接載入單一原版角色 sidecar bundle；完整六人 SAVGAM party／area container 仍未解析。

第一百零五輪功能／文件 commit：`f57ba3e`，已推送至 GitHub `main`。將 active Bless／Curse effect 投影為 fighter attack +1／-1，其他 target/phase effects 保留未套用。

第一百零六輪功能／文件 commit：`253fdfb`，已推送至 GitHub `main`。新增 active Blind／Bestow Curse／friendly Prayer fighter projection；Haste、Protection、Mirror Image 與完整 target/action rules 仍未完成。

第一百零七輪功能／文件 commit：`c3e93f3`，已推送至 GitHub `main`。保存 DOS player `icon_id @ 0x143` 到 party／combat metadata；runtime icon slot allocation、CombatMap position／camera 與完整戰鬥流程仍未完成。

第一百零八輪功能／文件 commit：`75ae586`，已推送至 GitHub `main`。城市 `INN` 會恢復 party roster／fighter HP 並以繁中訊息返回場所選單；完整 CAMP、商店與酒館情報仍未完成。

第一百零九輪功能／文件 commit：`755d83c`，已推送至 GitHub `main`。建立 price-injected Buy／Sell／200 GP ID party transaction；完整 money pool、shop stock 與 Shop Menu UI 仍未完成。

第一百一十輪功能／文件 commit：`e8ddd14`，已推送至 GitHub `main`。城市 STORE 已接入繁中 BUY／VIEW／TAKE／POOL／SHARE／APPRAISE／EXIT menu，未知 stock／money-pool action 保留明確 boundary。

第一百一十一輪功能／文件 commit：`8ec1f86`，已推送至 GitHub `main`。接入 injected shop offers、party money pool、POOL／TAKE／SHARE 與 pool-funded BUY；實際 item selection UI、VIEW／APPRAISE 仍未完成。

第一百一十二輪功能／文件 commit：`981d01d`，已推送至 GitHub `main`。BUY 會顯示繁中商品／價格、扣除 party pool、加入 active character inventory 並返回 Shop Menu；VIEW／TAKE 數量／APPRAISE 仍未完成。

第一百一十三輪功能／文件 commit：`502b53e`，已推送至 GitHub `main`。VIEW 會列出角色 HP／金幣／繁中裝備摘要並返回 Shop Menu；完整 ALTER／ID／APPRAISE 與角色選擇 side effects 仍未完成。

第一百一十四輪功能／文件 commit：`5ef7203`，已推送至 GitHub `main`。TAKE 會選角色與 1／10／100／全部金額，更新 pool／角色 gold 並返回 Shop Menu；任意數字輸入與 APPRAISE 仍未完成。

第一百一十五輪功能／文件 commit：`6d69011`，已推送至 GitHub `main`。APPRAISE 會選角色與 gems／jewelry，接受 injected offer 後清空財寶、將 GP 加入 pool 並返回 Shop Menu；拒絕報價分支仍未完成。

第一百一十六輪功能／文件 commit：`8cdde8b`，已推送至 GitHub `main`。APPRAISE 新增接受／拒絕／返回確認；只有接受才清除財寶並入帳 pool。

第一百一十七輪功能／文件 commit：`993ced5`，已推送至 GitHub `main`。建立繁中 CAMP Menu，接入 REST return 與 EXIT 荒野返回；SAVE／VIEW／MAGIC／ALTER／FIX 保留明確 placeholder boundary。

第一百一十八輪功能／文件 commit：`b7c3f44`，已推送至 GitHub `main`。CAMP VIEW 新增角色選單、只讀繁中摘要與返回 CAMP Menu；未識別物品與 ALTER side effects 仍不猜測。

第一百一十九輪功能／文件 commit：`e4734d0`，已推送至 GitHub `main`。CAMP MAGIC 新增角色選單與已記憶 spell-slot ID 查看；當時完整 spell catalog、prepare／cast／recovery rules 尚未接入；目前已另加 bounded 一級法術名稱 catalog。

第一百二十輪功能／文件 commit：`e99977e`，已推送至 GitHub `main`。CAMP SAVE 透過一次性 state request 接到 Ebiten versioned party save；原版 SAVGAM slot／area container 尚未解析。

第一百二十一輪功能／文件 commit：`29284fb`，已推送至 GitHub `main`。CAMP ALTER／ORDER 新增兩階段角色重排，並同步 party roster 與 combat fighter；DROP／SPEED／ICON／PICS／FIX 仍未完成。

第一百二十二輪功能／文件 commit：`7c73ab3`，已推送至 GitHub `main`。CAMP ALTER／DROP 新增二次確認的永久角色移除，並同步 party roster／combat fighter；SPEED／ICON／PICS／FIX 仍未完成。

第一百二十三輪功能／文件 commit：`f235e96`，已推送至 GitHub `main`。CAMP ALTER／PICS 新增怪物圖片／動畫 runtime toggle，並接到事件／戰鬥 renderer；SPEED／ICON／FIX 仍未完成。

第一百二十四輪功能／文件 commit：`cf49211`，已推送至 GitHub `main`。CAMP ALTER／SPEED 新增 1–5 級訊息速度與 Ebiten Unicode message reveal；ICON／FIX 仍未完成。

第一百二十五輪功能／文件 commit：`8ae1d64`，已推送至 GitHub `main`。CAMP ALTER／ICON 新增已驗證 CHEAD／CBODY block 選擇，並同步 party／combat fighter icon；FIX 仍未完成。

第一百二十六輪功能／文件 commit：`2de8b01`，已推送至 GitHub `main`。CAMP FIX 依已記憶的 Cure Light Wounds（目前由一級牧師表順序映射為 ID `3`）以 deterministic `1d8` 治療受傷 roster，並同步 combat fighter HP；spell catalog、時間推進與中斷規則仍待反組譯。

第一百二十七輪功能／文件 commit：`c585d1b`，已推送至 GitHub `main`。城市 BAR 已接 ordered Tavern Tale menu、前六則繁中整理與城市場所返回；買酒價格、城市條件、完整 62 則內容與 ECL trigger 仍待反組譯。

第一百二十八輪功能／文件 commit：`c82919e`，已推送至 GitHub `main`。CAMP／PROGRAM 9 現在只開啟 CAMP Menu；REST 接入 `ADD／SUBTRACT／EXIT` 與每 24 小時自然恢復 1 HP，並同步 roster／fighter。法術記憶、遊戲時鐘與遭遇中斷仍待反組譯。

第一百二十九輪功能／文件 commit：`ce44a6d`，已推送至 GitHub `main`。CAMP MAGIC 現在將已核對的一級牧師／魔法師前八個 spell IDs 顯示為繁中名稱，未知 ID 保留 hex；完整 spell catalog、CAST／MEMORIZE／SCRIBE 與 recovery rules 仍待反組譯。

第一百三十輪功能／文件 commit：`a885b9f`，已推送至 GitHub `main`。DOS known-spell flags 現在保存到 `Character.KnownSpells`／party save，並在 CAMP MAGIC 顯示已記憶／可用數量；CAST／MEMORIZE／SCRIBE 與消耗規則仍未完成。

第一百三十一輪功能／文件 commit：`8ba67b5`，已推送至 GitHub `main`。依 RuleBook 將 CAMP MAGIC 接成 `CAST／MEMORIZE／SCRIBE／DISPLAY／REST／EXIT` command menu；DISPLAY／REST 已接入，CAST／MEMORIZE／SCRIBE 仍保留 rules boundary。

第一百三十二輪功能／文件 commit：`312f9b4`，已推送至 GitHub `main`。MEMORIZE 可從 KnownSpells 選法術，先保存 pending state，REST_START 才寫回 SpellSlots；完整 capacity、準備時間、遭遇中斷與 CAST／SCRIBE 仍未完成。

第一百三十三輪功能／文件 commit：`3b61c33`，已推送至 GitHub `main`。對已核對的一級法術套用 RuleBook 的最低準備時間檢查；不足休息時保留 pending selection，不猜測高等級與遭遇中斷結果。

第一百三十四輪功能／文件 commit：`2d7af62`，已推送至 GitHub `main`。戰鬥接入 RuleBook 證實的 Magic Missile（spell ID `7`）、slot consumption、2–5 damage per missile、level scaling 與 Ebiten S 鍵；其他 spell effects 仍未完成。

第一百三十五輪功能／文件 commit：`2a7b7c7`，已推送至 GitHub `main`。戰鬥接入 RuleBook 證實的 Cure Light Wounds（spell ID `3`）、牧師 slot、1d8 封頂治療與 Ebiten H 鍵；完整 target cursor、施法中斷與其他 spell effects 仍未完成。

第一百三十六輪功能／文件 commit：`25c0019`，已推送至 GitHub `main`。S／H 先進入施法目標選擇，左右切換、Enter 確認、Esc 取消；Magic Missile／Cure Light Wounds 分別使用敵方／我方 target list。

第一百三十七輪功能／文件 commits：`89b7620`、`a7aaca3`，已推送至 GitHub `main`。已接入 RuleBook Bless（spell ID `1`）的 B／Enter／Esc 無目標施法 transaction；成功消耗牧師 slot，隊伍攻擊加值效果與後續 adjacency／duration 收斂於下一輪。

第一百三十八輪功能／文件 commits：`64214fe`、`3090264`，已推送至 GitHub `main`。Bless 已依 RuleBook `6r` 接入 6 回合 duration，並依 CombatMap 八方向相鄰排除鄰近存活怪物的隊友；缺少位置資料時採 bounded 不排除 fallback。

第一百三十九輪功能／文件 commits：`82e2c64`、`9279fc2`，已推送至 GitHub `main`。已接入 RuleBook Curse（spell ID `2`）的 C／敵方目標選擇／Enter confirmation；未與我方八方向相鄰的敵人攻擊加值降低 1，持續 6 回合後恢復，並保留無 position direct API fallback。

第一百四十輪功能／文件 commits：`a2d968b`、`dabafa2`，已推送至 GitHub `main`。已接入 RuleBook Cause Light Wounds（spell ID `4`）的 W／敵方 touch target selection／Enter confirmation；相鄰敵人承受 deterministic 1d8 damage、無 saving throw，並保留無 position direct API fallback。

第一百四十一輪功能／文件 commits：`ee6d6d1`、`883321b`，已推送至 GitHub `main`。已接入 RuleBook Protection from Evil（spell ID `6`）的 P／party touch target／Enter confirmation；明確標記 `Evil=true` 的攻擊者對受防護目標只獲得 AC +2 門檻，duration 為 `3×caster level` 回合；saving throw、alignment import 與 dispel 保留 boundary。

第一百四十二輪功能／文件 commits：`cd8942a`、`9e6e660`，已推送至 GitHub `main`。已修正 class-local spell ID collision，牧師 ID `7` 由 G 施放 Protection from Good，魔法師 ID `7` 由 S 施放 Magic Missile；明確 `Good=true` 攻擊者才觸發 AC +2，duration 為 `3×caster level` 回合。

第一百四十三輪功能／文件 commits：`c7fac1f`、`2f078a1`，已推送至 GitHub `main`。ECL encounter 的 FLEE 進入繁中可恢復撤退事件；PARLAY 提供 HAUGHTY／SLY／MEEK／NICE／ABUSIVE 五種 tactic，選擇後返回荒野事件。怪物速度、追擊、speaker／reaction 與完整 conversation script 保留 boundary。

第一百四十四輪功能／文件 commits：`7849d19`、`4a26a36`，已推送至 GitHub `main`。戰鬥按 M 進入 MOVE，方向鍵單格移動當前 party fighter，Battle 驗證 occupancy 後更新 CombatMap 座標並消耗回合；地形、負重、進入敵格、facing 與完整離場規則仍保留 boundary。

第一百四十五輪功能／文件 commits：`1fab06e`、`b8c5616`，已推送至 GitHub `main`。MOVE 成功後若角色離開敵人鄰接範圍，Battle 對該移動者觸發存活 enemy free attack，State 顯示繁中反擊訊息並沿用勝負 transition；背面 AC／facing／reach／地形仍保留 boundary。

第一百四十六輪功能／文件 commit：`77e4245`，已推送至 GitHub `main`。MOVE 移入存活敵人格時回傳既有 attack transaction，保留 party fighter 原座標；隊友格仍拒絕，離開敵人鄰接範圍的 free attack 仍維持既有 branch。新增 RuleBook 規格、繁中 README／共用 state knowledge 與 game/combat tests。

第一百四十七輪功能／文件 commit：`960cffb`，已推送至 GitHub `main`。依 RuleBook active character centered camera，新增可重用 `CombatCamera`、State active fighter read API、Ebiten 座標轉換、測試與 graphics knowledge 更新；viewport 尺寸、scroll animation、地圖遮擋與真實 Area camera 仍保留 boundary。

第一百四十八輪功能／文件 commit：`223fc04`，已推送至 GitHub `main`。Combat Menu `VIEW` 以 `V` 開啟繁中 read-only fighter summary，Enter／Esc 關閉且不消耗回合；新增 State／renderer tests、READY 規格與共用 state knowledge。完整 View Menu、物品／交易與 Combat FLEE 的速度／追擊規則仍保留 boundary。

第一百四十九輪功能／文件 commit：`418fe09`，已推送至 GitHub `main`。依 RuleBook 與 ITEMS RateOfFire，將已裝備武器 raw `4/6` 投影為每回合 2/3 次攻擊，接入 Battle sequence、目標倒下後換下一存活敵人、繁中摘要與測試；彈藥消耗已於第 150 輪接入注入式 adapter，職業等級額外攻擊、Aim／range 與 back stab 仍保留 boundary。

第一百五十輪功能／文件 commit：`02a161a`，已推送至 GitHub `main`。保存武器 raw AmmunitionType，建立 raw code→inventory type mapping 的注入式 atomic consumption，CombatAct 在 attack 前 preflight 本回合 shots；mapping 缺失／彈藥不足不修改 inventory。新增 party／game tests、READY 規格與共用 state knowledge。

第一百五十一輪功能／文件 commit：`d4aef0b`，已推送至 GitHub `main`。Combat Menu `DONE` 以 `D` 結束目前 party turn，不攻擊、不消耗彈藥，重用 enemy／next-party advancement；新增繁中提示、測試、READY 規格與共用 state knowledge。hold／delay 與其他 Combat Menu command 仍保留 boundary。

第一百五十二輪功能／文件 commit：`900db82`，已推送至 GitHub `main`。依 RuleBook Armor List 將 armor type `50–58` 的 12／9／6 movement allowance 投影到 fighter，MOVE 每個方向鍵逐格扣除，剩餘格數不推進回合；新增 UI 提示、party／game tests、READY 規格與共用 state knowledge。負重、地形、障礙、邊界與 FLEE speed 仍保留 boundary。

第一百五十三輪功能／文件 commit：`e6971c8`，已推送至 GitHub `main`。依 RuleBook missile adjacent prohibition 與 dart thrown exception，將 ITEMS weapon group profile 投影到 fighter，Battle 在有 CombatMap 座標時拒絕 missile 近身攻擊；新增 tests、READY 規格與共用 state knowledge。完整 Range、line-of-sight、Aim cursor 與其他 thrown weapon 仍保留 boundary。

第一百五十四輪功能／文件 commit：`e0d4e31`，已推送至 GitHub `main`。修正第 153 輪 guard 與第 150 輪 ammunition atomic contract 的順序，Battle 提供不擲骰 `ValidateAttack`，CombatAct 在扣除彈藥前先拒絕無效相鄰 missile attack；新增 regression、READY 規格與共用 state knowledge。完整 ranged multi-target transaction、Range、line-of-sight 與其他 thrown weapon 仍保留 boundary。

第一百五十五輪功能／文件 commit：`07634ab`，已推送至 GitHub `main`。將已投影的 `Fighter.AttacksPerTurn` 套用到 enemy turn，讓敵方也使用 deterministic `AttackSequence` 與繁中多次攻擊摘要；新增 regression、READY 規格與共用 state knowledge。enemy AI、彈藥、Aim／line-of-sight 與額外職業攻擊仍保留 boundary。

第一百五十六輪功能／文件 commit：`531f892`，已推送至 GitHub `main`。將玩家 combat input error 接成 `ReportCombatError`／繁中訊息，尤其讓相鄰 missile／彈藥／目標錯誤留在戰鬥畫面而不結束 Ebiten game loop；新增 error sentinel、輸入攔截、regression、READY 規格與共用 state knowledge。完整 error catalog、ranged rules 與資料／啟動錯誤仍保留 boundary。

第一百五十七輪曾將 `ADD NPC (0x36)` 誤判為單 operand，使 block 0x52 的 morale
operand code `0x00` 被當成 EXIT；第 277 輪已推翻並修正。正確流程是連續加入
`0x55/0x58/0x5A`、播放完整 demo 展示序列並抵達 COMBAT；第 278 輪確認正常 new game
改走 global block `0x01`。

第一百五十八輪功能／文件 commit：`f9164e4`，已推送至 GitHub `main`。依跨 ECL 實際掃描將 `LOAD PIECES (0x37)` 接成三 selector signal，讓 ECL2 block 0x01 等實際 entry 不再因 opcode 停止；新增 synthetic／ECL1／ECL2 regression、CLI／BlockSession propagation、READY 規格與共用 state knowledge。地城 floor、wall、tile、碰撞與 camera side effect 仍保留 boundary。

第一百五十九輪功能／文件 commit：`eec71dd`，已推送至 GitHub `main`。將 `LoadPiecesRequested` 從 ECL runner 接到 game State 一次性 `ConsumeLoadPiecesRequest()`，與 GEO `LOAD FILES` request 對齊；新增 state regression、READY 規格與共用 state knowledge。當時尚未接入 map-piece file adapter；完整地城副作用仍保留 boundary。

第一百六十輪功能／文件 commit：`7ba3006`，已推送至 GitHub `main`。依公開 CoAB `LoadWalldef` reference 將 State 的 `LOAD PIECES` request 接到 `WALLDEF{area}`／`8X8D{area}` raw `PieceSet` catalog；補上單／多 WALLDEF record selector regression、原始 ZIP area 2 regression、dungeon preview 載入狀態與共用 graphics knowledge。WALLDEF row／column 的牆面拼圖、0x7F 特殊分支、碰撞與完整 3D renderer 仍保留 boundary。

第一百六十一輪功能／文件 commit：`1da9e05`，已推送至 GitHub `main`。依 reference `WallDefBlock.Offset` 將 WALLDEF graphic IDs 套用 dungeon global bases `0x2E／0x74／0xBA`，新增 `PieceSet.WallSymbol` bounded cell-to-8×8D lookup、offset regression、READY 規格與共用 graphics knowledge；九種 3D viewport layout、GEO 深度遍歷與完整 renderer 仍保留 boundary。

第一百六十二輪功能／文件 commit：`4b88cf6`，已推送至 GitHub `main`。依 reference `draw_3D_8x8_titles` 建立十組 `idxOffset／rowCount／colCount` wall layout，輸出 `WallStamp` 並接到 dungeon preview 的原始 8×8D sample；補上 layout regression、READY 規格與共用 graphics knowledge。`Draw3dWorldFar/Mid/Near` 方向遍歷、遮擋、sky／roof、door 與 camera 仍保留 boundary。

第一百六十三輪功能／文件 commit：`519ec4f`，已推送至 GitHub `main`。依 reference `Draw3dWorldFar/Mid/Near` 將 party direction 與 GEO wall fields 接成 ordered `WallLayoutCall` traversal，dungeon preview 改用 Far／Mid／Near 全段 wall stamps；補上 depth／座標 regression、READY 規格與共用 graphics knowledge。GEO wrap、sky／roof、door、遮擋與 camera 仍保留 boundary。

第一百六十四輪功能／文件 commit：`b890ffc`，已推送至 GitHub `main`。dungeon preview 保存 `(dungeonX,dungeonY)`，方向鍵透過 GEO 雙側 wall contract 移動，成功後重建 floor 與 Far／Mid／Near wall stamps；新增 position／camera slice 規格與共用 graphics knowledge。Area／save 真實座標、direction、movement cost、encounter、wrap 與 scroll animation 仍保留 boundary。

第一百六十五輪功能 commit：`6f530a9`，已推送至 GitHub `main`。將 dungeon preview 的 `(x,y,direction)` 接入 remake game save version 3；v1/v2 舊檔與越界值回到安全預設，F9／啟動載入後重建 floor／wall view。新增 save codec／game adapter round-trip regression、READY 規格與共用 state knowledge。原版 DOS `SAVGAM?.DAT` container 與 Area 真實座標寫回仍保留 boundary。

第一百六十六輪功能 commit：`e45cd03`，已推送至 GitHub `main`。將 dungeon preview Q/E 接成 reference 八方向 facing rotation（`±2`），轉向後重建 Far／Mid／Near wall view；新增 state wrap regression、READY 規格與共用 state knowledge。Area 真實 `mapDirection`、轉向時間、sky／roof／door overlay 與完整 3D viewport 仍保留 boundary。

第一百六十七輪功能 commit：`9324a03`，已推送至 GitHub `main`。依 reference `getMap_XXX`／`MovePositionForward` 將 dungeon 16×16 coordinate wrap 接成明確 `geo` wrapped API、wrapped Far／Mid／Near traversal 與 preview 跨邊界 movement；strict API 保留給不允許 wrap 的 context。原版 `mapWallType/mapWallRoof` 五-byte save segment、ECL 例外判斷與完整 3D overlay 仍保留 boundary。

第一百六十八輪功能 commit：`5cc1b42`，已推送至 GitHub `main`。依 reference `ovr017` 5-byte map segment 將 `mapWallType`／`mapWallRoof` 接入 remake save version 4；wrapped GEO refresh 會在位置／方向變更後重算，v3 舊檔 cache 回到 0。新增 save codec／game adapter round-trip 與 v3 compatibility regression、READY 規格及共用 state knowledge。完整 `SAVGAM?.DAT` container、slot、Area／ECL memory 與 player records 仍保留 boundary。

第一百六十九輪功能 commit：`f714494`，已推送至 GitHub `main`。解碼 Area1 `0x1FA/0x1FC` outdoor／indoor sky colour，依 `mapWallRoof > 0x7F` 接入 dungeon preview EGA sky background；新增 Area codec regression、READY 規格與共用 state knowledge。完整 `Draw3dWorldBackground`、roof geometry、door overlay 與原版 save container 仍保留 boundary。

第一百七十輪功能 commit：`a8c050e`，已推送至 GitHub `main`。依 reference `WallDoorFlagsGet` 接入 GEO wrapped no-wall default `1`／walled `x3` detail API，並在 dungeon preview 顯示目前 facing 的 door/detail evidence；新增 exact GEO regression、READY 規格與共用 state knowledge。開門、解鎖、撬門、撞門與 door symbol overlay 仍保留 boundary。

第一百七十一輪功能 commit：`1d292c0`，已推送至 GitHub `main`。依 reference `TryStepForward`／`MapSetDoorUnlocked` 接入 detail `1` unlocked doorway movement 與雙側 `UnlockDoorWrapped` raw mutation；detail `2/3` 保持阻擋。新增 GEO movement／mutation regression、READY 規格與共用 state knowledge。完整 bash／pick／knock menu、技能／骰點／法術消耗與 door graphics 仍保留 boundary。

第一百七十二輪功能 commit：`a6a1622`，已推送至 GitHub `main`。依 reference `Player.thief_skills` 將 DOS record `0xEA–0xF1` 保存到 `DOSPlayerRecord`／`Character`／JSON，提供 `OpenLocksSkill()` index 1 adapter；新增 synthetic parser regression、READY 規格與共用 party/state knowledge。thief skill 重算、pick-lock dice、door menu、bash／knock 與完整 DOS save container 仍保留 boundary。

第一百七十三輪功能 commit：`939b258`，已推送至 GitHub `main`。依 reference `pick_lock()`／`Spells.knock` 建立 `internal/dungeon` 的 injected d100 pick-lock resolver 與 Knock `0x1F` first-slot consume；新增逐位 roll、健康狀態、inclusive roll、失敗消耗嘗試與 spell removal regression、READY 規格與共用 dungeon knowledge。door mutation、完整 door menu、bash 與 thief skill 重算仍保留 boundary。

第一百七十四輪功能 commit：`ec2d0bb`，已推送至 GitHub `main`。將 pick-lock／Knock 接入 Ebiten dungeon preview 的 P/K action adapter；P 只允許 detail 2，K 允許 detail 2/3，成功後呼叫 GEO 雙側 unlock，新增 seeded State action regression 與 READY 規格。完整 locked-door menu、bash、door graphics 與劇情 integration 仍保留 boundary。

第一百七十五輪功能 commit：`db51598`，已推送至 GitHub `main`。依 reference `bash_door()` 保存 DOS `Str.full`／`Str00.cur`，建立 detail 2/3 的 strength／exceptional die resolver，並接入 dungeon preview `B` 撞門與 GEO 雙側 unlock；新增 bash table、extra-roll 與 DOS import regression、READY 規格。完整 locked-door menu、side effects、door graphics 與劇情 integration 仍保留 boundary。

第一百七十六輪功能 commit：`a53744b`，已推送至 GitHub `main`。依 reference `locked_door` 建立 detail 2/3 的 Bash／Pick／Knock／Exit capability resolver，並將方向鍵撞上上鎖門時接到 preview menu；新增 menu capability regression 與 READY 規格。完整 DOS 視窗樣式、door graphics、時間／傷害 side effects 與劇情 integration 仍保留 boundary。

第一百七十七輪功能 commit：`b86c485`，已推送至 GitHub `main`。依 reference `seg001.Init` 將 State、save／preview fallback 與 startup dungeon default 從 `(8,8,0)` 修正為 `(7,13,0)`，並清理已過時的 README／knowledge assertions；新增 READY 規格。`InitAgain` direction 2、完整 SAVGAM context 與劇情 integration 仍保留 boundary。

第一百七十八輪功能 commit：`cd2c0d9`，已推送至 GitHub `main`。依 reference `LoadPlayerCombatIcon`／`chead_cbody_comspr_icon` 將 DOS `icon_size=1` 的 CHEAD/CBODY raw slot 映射到 `+0x40`，載入 extracted raw layers 並在 renderer on-demand 合成；新增 small／normal icon regression、READY 規格與共用 party knowledge。direction-specific placement、recolor、animation 與完整 CombatIcon runtime 仍保留 boundary。

第一百七十九輪功能 commit：`2c3426c`，已推送至 GitHub `main`。依 reference `CombatIcon.LoadIcons` 將 attack layer 映射到 normal block `+0x80`，並接入 CHEAD／CBODY on-demand attack composition；新增 attack block regression、READY 規格與共用 icon knowledge。direction-specific placement、recolor 與完整 CombatIcon runtime cache 仍保留 boundary。

第一百八十輪功能 commit：`9f7c476`，已推送至 GitHub `main`。依 reference `SetupCombatActions`／`HalfDirToIso` 將 map direction 映射到 party／enemy `IconDirection`，接入 StartCombat 與水平 flip adapter；新增 placement／StartEncounter regression、READY 規格與共用 combat knowledge。完整 Area/ECL direction source、CombatMap placement、recolor 與 runtime cache 仍保留 boundary。

第一百八十一輪功能 commit：`bed7e56`，依 reference `ovr017.SaveGame/loadSaveGame` 建立 `SAVGAM?.DAT` 固定前綴 raw codec：保存 game area、Area1／Area2、runtime／ECL raw bytes、5-byte map state、game states、三組 block/set pair、party count 與 8 筆固定 CHRDAT name records；新增 strict-size validation、trailing player-file suffix boundary regression、READY 規格與共用 Gold Box save knowledge。完整 slot、Area 欄位解碼、個別 player files 與 file side effects 仍保留 boundary。

第一百八十二輪功能 commit：`216f7b2`，將 `SAVGAM` fixed prefix 接到 `game.State.LoadSAVGAMPrefix`／`SaveSAVGAMPrefix`；依 Area codec 更新已知 Area／GEO/map 欄位，保留未知 runtime／ECL／raw records，並以 signed map position、facing、wall cache 建立 State regression。此為 prefix load/export adapter，不取代 F5 remake JSON，也不宣稱已完成個別 CHRDAT player files、slot 選擇與 multi-file side effects。

第一百八十三輪功能 commit：`aa04200`，依 reference `seg044.SoundInit/PlaySound` 與 `Main/Resource.resx` 保存 9 個 PC WAV，建立 `internal/sound` selector catalog、WAV decode regression 與 Ebiten playback adapter；title start、荒野移動、dungeon preview 移動已播放對應原版音效，並加入 `-sound-dir`。完整戰鬥 sound calls、背景音樂、MIDI／AdLib 與音量設定仍保留 boundary。

第一百八十四輪功能 commit：`e0d871f`，建立 renderer-neutral `game.SoundEvent` queue，將武器命中／未命中、擊倒、移動時攻擊／免費反擊與已實作法術的 sound intent 接到 Ebiten `internal/sound` player；title start 與 wilderness movement 也改由 State queue 發送。新增 one-shot event order regression、READY 規格並清理上一輪「完整戰鬥音效尚未接入」的過時斷言；背景音樂、MIDI／AdLib、音量設定與所有 ECL sound calls 仍保留 boundary。

第一百八十五輪功能 commit：`9186a08`，依 `ovr017.SaveGame/loadSaveGame` 的實際檔名與 side-effect 順序，新增 `State.LoadSAVGAMSlot`：載入 `savgamA..J.dat`、`CHRDAT{slot}{1..6}.sav` 與 optional `.swg/.fx`，重用既有 DOS player parsers 建立 party／fighter 並進入繁中 wilderness；新增 `-savgam-dir/-savgam-slot` 啟動入口、synthetic slot regression、READY 規格與 state knowledge。Player.StructSize writeback、原始檔刪除與 CAMP multi-file save transaction 仍保留 boundary。

第一百八十六輪功能 commit：`9e64347`，依同一組 `ovr017` 證據新增 `PatchDOSPlayerRecord`、`.swg/.fx` encoders 與 `State.SaveSAVGAMSlot`。已載入角色的已證實欄位可回寫，未知 `.sav` bytes 保留；輸出先進 staging directory 再逐檔替換，仍保留原版刪檔、多職業、未知 sidecar 與 CAMP multi-file atomic transaction boundary。`go test ./...` 與兩個 CLI build 已於 Docker 通過。

第一百八十七輪功能 commit：`b40587c`，將 loaded SAVGAM slot 接到 Ebiten F5 與 CAMP SAVE；`-savgam-dir/-savgam-slot` 模式寫回同一 slot，一般模式維持 remake JSON。新增 workflow 規格與 README／PLAN／state knowledge 更新；原版刪檔、多職業、未知 sidecar 與跨檔案 atomic transaction 仍保留 boundary。

第一百八十八輪功能 commit：`ce606ad`，依已知 `SaveGame` side effect，將 SAVGAM slot 的 prefix 與 `CHRDAT{key}{1..6}` 檔案先移入 backup，再替換 staged bundle；隊伍縮編的 stale player sidecars 會被清理，替換失敗可 rollback。新增 stale-file regression 與 READY 規格；多職業、未知欄位與完整 player serialization 仍保留 boundary。

第一百八十九輪功能 commit：`84def37`，修正戰鬥結束時只同步 renderer-facing party、未同步持久 `partyRoster` 的狀態問題；現在 HP／MaxHP 會依 fighter ID 回寫 roster，供 CAMP 與兩種 save path 使用，並新增 regression／READY 規格與 state knowledge。

第一百九十輪功能 commit：`7b33384`，修正戰鬥結果按 Enter 後保留 stale ECL choices 的主流程問題；新增 `restoreWildernessMenu`，統一返回繁中荒野主選單，並加入 continuation regression／READY 規格與 state knowledge。

第一百九十一輪功能 commit：`a9806e9`，將城市商店 SELL 接入繁中 Shop Menu，依已解碼 item `Value` 將非 readied／非 cursed 物品售出並將 GP 放入 party pool；新增 menu／transaction regression、繁中 locale、READY 規格與共用 shop knowledge。城市 stock／ID／鑑定 routine 仍保留 boundary。

第一百九十二輪功能 commit：`befdc61`，將既有 `PayIdentifyFee` 接入繁中 Shop Menu，完成角色／物品選擇與 200 GP ID transaction，保留 `HiddenNameFlags` 與未解碼 magic result；新增 regression、locale、READY 規格與 shop knowledge。

第一百九十三輪功能 commit：`00682ff`，修正 `BlockSession` 跨 `NEWECL` 遺失 `LOAD FILES`／`PICTURE`／`SPELL`／`PROTECTION` 結果的問題；新增跨三個 synthetic ECL block 的 signal regression。`go test ./internal/ecl` 已於 Docker 通過；完整 `go test ./...` 仍受容器缺少 ALSA／X11 headers 及既存 game integration failure 影響。

第一百九十四輪功能 commit：`630d1b3`，修正真實 ECL1 JOURNEY ON integration regression：PICTURE 已是明確的繁中事件畫面，測試現在先驗證 request，再以 `Continue()` 模擬 Enter，最後確認流程抵達 COMBAT boundary。Docker non-Ebiten internal packages 全部通過。

第一百九十五輪功能 commits：`ad676f2`、`12a0fd7`，將 ECL `SPELL`／`PROTECTION` 結果接到 State pending queue，新增一次性 consume API，並驗證真實 `State.Select` wiring；State 保留原始 signal 順序／位址，不猜測未知 party memory side effect。`go test ./internal/game ./internal/ecl` 已於 Docker 通過。

第一百九十六輪功能 commit：`35fffaa`，新增 `SmokeInitializationEntries` 與 `cmd/azure-bonds -entry-smoke`，逐一 bounded 執行 ECL1–ECL6 全部 block 的五個 initialization entries；實際 image smoke run 已記錄 menu／COMBAT／monster spawn／unsupported opcode per entry，並新增 READY 規格與 ECL knowledge。

第一百九十七輪功能 commit：`d1327af`，以真實 ECL2 block 3 entry 3 與 `MON2CHA.DAX` 建立 playable Battle regression；修正 `MON*CHA` 50..60 packed ArmorClass 的 `60-raw` adapter，並新增 `-encounter-monster-member` 支援跨章節 direct encounter。ECL2 direct entry 已於 Docker 通過，正常玩家流程仍待完整 ECL continuation。

第一百九十八輪功能 commit：`860f7c4`，遊戲啟動載入 `MON1CHA`–`MON6CHA`，State 依 ECL global block namespace 選擇 chapter-local monster table；新增 ECL2 chapter selection regression。`go test ./internal/game ./internal/monster ./internal/ecl` 已於 Docker 通過。

第一百九十九輪功能 commit：`cb70681`，以原始 ECL1 block `0x50` payload `+0x5B5` 驗證 `NEWECL 0x03` 會切換到 ECL2 block `3`，新增 global session regression；target 後續 unsupported routine 仍保留 bounded stop boundary。

第二百輪功能 commit：`f822c89`，依既有 ECL command table／operand contract 將 `0x2F AND` 與 `0x30 OR` 接入 bounded 16-bit memory destination semantics，新增 regression 與 READY 規格；另建立可供後續 Gold Box 作品沿用的 [`gold-box-ecl-command-set.md`](docs/knowledge/gold-box-ecl-command-set.md) 指令集知識庫。ECL1–ECL6 smoke 已遇到的 `0x2D CALL` 仍維持 unsupported，待確認 external dispatch／return context 後再實作。

第二百零一輪功能 commit：`a04f6d6`，以原始 ECL3／ECL4／ECL6 smoke 的 `code 0x01` monster operands 為證據，將 `LOAD MONSTER`／`SETUP MONSTER` 接到 bounded runtime memory resolution，加入 byte-range validation 與 variable descriptor regression；ECL3 block 17／18、ECL4 block 33／37 real entries 已抵達 COMBAT／spawn boundary。完整 `CALL`／external routine、party memory 與玩家流程仍保留 boundary。

第二百零二輪功能 commit：`c45888e`，以 ECL1–ECL6 raw scan 的 `0x2E10`／`0xC01E`／`0xB200` 非 code-segment CALL operands 與 ECL3 opening 的 CALL→PRINT／menu sequence 為證據，新增 `RunResult.CallAddresses` external dispatch signal；bounded VM 從 CALL 後續 instruction 繼續，ECL3 block 16／17／18／21 smoke 已越過原本 `0x2D` stop。真正 engine routine side effect 仍保留 boundary。

第二百零三輪功能 commit：`76a1fa4`，將已由 ECL3 block 16 entry 4 raw image 驗證的 Yulash smoke text 接入 zh-TW locale 與 State event message；未知 ECL text 原樣保留，raw runner 結果不變，新增 localization regression。完整 ECL 對話翻譯與其他作品文字仍需逐段反組譯／翻譯。

第二百零四輪功能 commit：`eb4ab29`，依 ECL3／ECL4 raw event runs 新增邪教徒／受傷牧師、戰火城市與小型魔法商店的 zh-TW catalog mapping；State 將 ECL text message 提前保存到 menu pause 前，新增 unknown fallback regression 與 READY 規格。完整事件分支、CALL routine side effect 與全部 ECL 文字仍保留 boundary。

第二百零五輪功能 commit：`11ea665`，依 ECL3 block 16 entry 4 的 raw `PRINT RETURN`→後續 menu sequence，新增 `RunResult.PrintReturnCount` 與 session aggregation；真實 entry 已由原本 `0x33` stop 推進至 `menu=true`，新增 bounded regression／READY 規格。DOS text-window layout 與完整後續事件仍保留 boundary。

第二百零六至二百零八輪功能 commit：`3dd0645`，依 ECL5 block 48 raw trace／scan 將 `LOAD CHARACTER`、`FIND ITEM`、`DESTROY ITEMS` 接成 bounded signals，保留 character address、inventory query／destroy IDs 並繼續 control flow；real entry 已由 `0x0A`／`0x32`／`0x40` stops 推進到 `NEWECL` boundary。party ownership、compare result、實際 item deletion 與完整事件分支仍保留 boundary。

第二百零九輪功能 commit：`74adb2f`，依 ECL5 sunlight event 的 raw evidence 建立
`DESTROY ITEMS` → persistent party roster adapter；新增可刪除 readied／stacked item
units 的 `Character.DestroyItemType`，並以 party／State regression 驗證。`FIND ITEM`
仍維持 query-only signal，compare result、完整 item namespace 與事件分支仍保留
boundary。Docker 已通過 `internal/party`、`internal/game`、`internal/ecl`、
`internal/locale` 測試。

第二百一十輪功能 commit：`ed4f162`，依公開 CoAB reference `CMD_Damage` 與六個
ECL raw scan，將 `DAMAGE` 五欄 `flags／dice_count／dice_size／damage_bonus／save_flags`
保存為 `RunResult.DamageRequests`，並接入 `BlockSession` aggregation；新增 synthetic
continuation 與真實 ECL2 block 3 `+0x1599` operand regression、READY spec 與 command-set
knowledge。Docker 已通過 `internal/ecl`、`internal/game`、`internal/party`、
`internal/locale`；target／saving throw／random roll／HP mutation 仍保留 party adapter
boundary。

第二百四十七輪重大里程碑：依 reference `Area1.field_6A00_Get/Set` 確認七個 game-time
words 位於 `0x18C..0x198`，接入 `area.State.GameTime`、Area1 binary codec、SAVGAM prefix
load、`State.SetAreaState`、`AdvanceGameTime` 與 remake save synchronization；新增 Area1
raw round-trip 與 State mirror regression、READY spec、README／PLAN／Gold Box state knowledge。
完整 calendar UI 與其他 unknown Area1 fields 仍保留 boundary。

第二百四十五輪重大里程碑：依 reference `race_ages`／`ovr018` 補上 single-class
`StartingAgeSpecFor` 與 deterministic `RollStartingAge`，明確映射六 race、六 supported
classes 的 base age／dice count／size，unsupported 組合回傳錯誤。新增 human fighter
range regression、READY spec 與 README／PLAN／Gold Box state knowledge；目前 starter
templates 與完整角色建立 UI 尚未自動套用，multi-class／half-orc 仍保留 boundary。

第二百四十六輪重大里程碑：將 `RollStartingAge`／`WithAgeEffects` 接入
`State.AddCreationCharacter`：保留可編輯 template 不變，加入隊伍時對 copied character 生成
deterministic age、套用六項 ability effects，再寫入 roster；新增 real creation regression。
完整 race/class／alignment 建立選單、多職業與 half-orc 仍保留 boundary。

第二百一十一輪功能 commit：`3068c2b`，將 `RunResult.DamageRequests` 接入 State
pending queue／`ConsumeDamageRequests()` exactly-once API，避免事件／選單 pause 遺失
script damage effect；新增 READY spec 與共用 State knowledge。由於 remake 尚未保存
原版五類 `saveVerse` 與 selected-character memory mapping，本輪不猜固定 HP mutation。
Docker 已通過 `internal/game`、`internal/ecl`、`internal/party`、`internal/locale`。

第二百一十二輪功能 commit：`387b9fb`，依 reference player `saveVerse` `0xDF–0xE3`
與 `CMD_Damage` flags，保存五類 saving throws 到 DOS parser／Character／JSON／record
writeback；新增 selected／whole-party DAMAGE resolver、natural 1／20、注入骰點、
transactional roster HP 與 stable-ID fighter sync。random-target／`CanHitTarget`、
affect save bonus 與死亡 continuation 仍保留 boundary。Docker 已通過
`internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十三輪功能 commit：`7250778`，依 reference `CanHitTarget` 接入 ECL DAMAGE
random-target branch：低八位 target count、party-size random selection、raw hit bonus、
natural 1／20 與注入式 hit resolver；State 提供 resolver variant 並 transactional sync
roster／fighter HP。新增 party／game regressions 與 READY spec；AC／invisibility affect、
save-effect bonus 與死亡 continuation 仍保留 boundary。Docker 已通過
`internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十四輪功能 commit：`442b51a`，依 reference `CanHitTarget`／`RollSavingThrow`
補上 DOS player `field_186 @ 0x186` signed saving bonus 的 parser／Character／JSON／
writeback，並接入 ECL DAMAGE save threshold；State 新增 default hit resolver，使用
fighter／equipment projected AC 與已證實的 invisibility `0x19`／`0x47` -4 attack roll。
blink／displace／其他 `CheckAffectsEffect` 與死亡 continuation 仍保留 boundary。
Docker 已通過 `internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十四輪文件 commit：`efa1056`，更新 README、PLAN、ECL／State 共用知識庫與
212／213 READY specs，明確區分已完成的 save bonus／AC／invisibility adapter 與後續
作品專屬效果邊界。

第二百一十五輪功能 commit：`b1a3298`，依 reference `CheckAffectsEffect(Type_16)`／
`AffectBlink` 修正 `CanHitTarget` 的 natural-20 順序：先放大為 100，再套用 effects；
active blink `0x25` 在 `actions.delay == 0` 時可將 attack roll 設為 -1。新增
`ECLHitContext` 與 State context adapter，讓戰鬥回合能傳入 action delay；displace 的
persistent affect-data bit、其他 effects 與 death continuation 仍保留 boundary。
Docker 已通過 `internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十五輪文件 commit：`0da9742`，更新 ECL／State 共用知識庫、READY spec、
README 與 PLAN，清除 blink 已完成後的過時描述。

第二百一十六輪功能 commit：`4cfa81e`，依 reference `AffectDisplace` 將 ECL `DAMAGE`
命中投影延伸到 displace `0x59`：FX effect-data 第一 byte 的 `0x10` consumed bit
使首次攻擊 miss、後續攻擊可命中；combat round 0 且 attack roll 為 0 時清除此 bit。
第二個功能 commit：`d4f4a51`，State working roster deep-copy effects，確保多筆 DAMAGE
transaction 在後續 request error 時 rollback displace bit，不污染 live roster。Docker
已通過 `internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十六輪文件 commit：`6a5f2ed`，更新 ECL AC／effects READY spec、Gold Box ECL／
State 知識庫、README 與 PLAN，記錄 displace data mapping 與 transactional rollback。

第二百一十七輪功能 commit：`8dc0c1e`，依 reference `damage_player` 將 ECL DAMAGE
傷害結果投影到可向後相容的 `Character.HealthStatus`／`DamageOutcome.Health`：exact zero
為 unconscious、1..9 overkill 為 dying、10+ overkill 為 dead，animated exact zero
亦為 dead；非 OK／animated 狀態 HP 寫回 0。DOS 固定 player record 未被臆測新增欄位。
Docker 已通過 `internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十七輪文件 commit：`9ad9f79`，更新 ECL DAMAGE READY spec、Gold Box ECL／
State 知識庫、README 與 PLAN，區分已完成的 health-state projection 與尚未接入的
`CheckAffectsEffect(Death)`、bleeding、combatant removal、party win/loss continuation。

第二百一十八輪功能 commit：`d8825a4`，新增 `combat.Battle.SetHitPoints` external bridge；
State active-combat ECL DAMAGE resolution 會把 roster HP 同步到 Battle fighter、重新
計算 party／enemy status，並在 status 結束時走既有 `finishCombat` continuation。新增
active party defeat regression；完整 `CheckAffectsEffect(Death)`、bleeding、effect removal
仍保留 boundary。Docker 已通過 `internal/combat`、`internal/game`、`internal/party`、
`internal/ecl`、`internal/locale`。

第二百一十八輪文件 commit：`e6f8e2b`，更新 ECL death continuation READY spec、Gold Box
ECL／State 知識庫、README 與 PLAN，清除 active combat win/loss 已完成後的過時描述。

第二百一十九輪功能 commit：`37d2678`，依 reference `RemoveCombatAffects` 建立
`Character.RemoveCombatAffects` 的 19-kind cleanup table，並在 active-combat ECL damage
角色倒下時接入 State；blink `0x25`／invisibility `0x19`／`0x47` 因不在 reference
清單中保留。新增 party cleanup 與 State death regression；`CheckAffectsEffect(Death)`、
bleeding、完整 combatant removal 仍保留 boundary。Docker 已通過 `internal/combat`、
`internal/game`、`internal/party`、`internal/ecl`、`internal/locale`。

第二百一十九輪文件 commit：`b31e551`，更新 death/effect READY spec、Gold Box ECL／
State 知識庫、README 與 PLAN，記錄 cleanup table 與尚未解出的 Death side effects。

第二百二十輪功能 commit：`eb24633`，依 reference `CheckAffectsEffect(Death)`／
`sub_3BEE8`／`AffectTrollFireOrAcid` 建立 `DeathEffectContext` 與
`Character.ApplyDeathEffects`：Bleeding 保存 overkill；affect_63 對 dying／unconscious
在明確 combat-heal 條件下恢復並建立永久 affect_5F；troll effect `0x64` 只在已知非
火／酸 damage flags 時以 3d6 建立 TrollRegen `0x66`。`State.ResolveDeathEffects` 以
deep-copy transaction 接入 roster／Battle sync；未猜測 ECL DAMAGE 缺少的 damage type。
Docker 已通過 `internal/combat`、`internal/game`、`internal/party`、`internal/ecl`、
`internal/locale`。

第二百二十輪文件 commit：`800c064`，更新 Death side-effect READY spec、Gold Box ECL／
State 知識庫、README 與 PLAN，明確保留 dragon-slayer target side effect 與其他未知
Death routine boundary。

第二百二十一輪功能 commit：`a0738a9`，依 reference `AffectDragonSlayer` 建立
`Character.ResolveDragonSlayer` 與 `State.ResolveDragonSlayer`：只有 explicit
`MonsterTypeDragon` target 才以 injected d12 計算 `1d12*3 + 4 + strength damage bonus`
並回傳 attack roll `+2`；非龍目標不觸發。target kind／Strength bonus 不從 ECL DAMAGE
五 operands 猜測。Docker 已通過 `internal/combat`、`internal/game`、`internal/party`、
`internal/ecl`、`internal/locale`。

第二百二十一輪文件 commit：`bd0b6da`，更新 Death effect READY spec、Gold Box ECL／
State 知識庫、README 與 PLAN，記錄 dragon-slayer explicit target contract。

第二百二十二輪功能 commit：`5abf8be`，依 reference `RemoveFromCombat`／`CombatantKilled`
將 Battle fighter HP=0 的 combat position 清為 `HasCombatPosition=false`，對應原版
`CombatMap[player_index].size=0`；既有 Battle win/loss 與 finishCombat continuation
保持不變。新增 State active-combat regression；skull overlay、actions clear、完整 map
redraw 與其他 Death routine 仍保留 boundary。Docker 已通過 `internal/combat`、
`internal/game`、`internal/party`、`internal/ecl`、`internal/locale`。

第二百二十二輪文件 commit：`4353c21`，更新 Death／combatant removal READY spec、
Gold Box ECL／State 知識庫、README 與 PLAN，記錄 position removal 與剩餘 renderer side
effects。

第二百二十三輪功能 commit：`eea573d`，將 reference `CombatantKilled` 的死亡視覺需求
建立為 renderer-neutral `Fighter.DeathOverlay` signal；Battle 在 HP=0 時同時保留死亡時
CombatX/Y anchor、清除 `HasCombatPosition`，治療時清除 overlay signal。Ebiten combat
renderer 已在該 anchor 畫出可見的繁中「倒下」overlay；exact `combat_icons[24]/[25]`
skull asset 尚未因 CPIC/COMSPR byte-family 證據不足而硬編。新增 Battle regression。
Docker Go 1.23 已通過 `internal/combat`、`internal/game`、`internal/party`、`internal/ecl`、
`internal/locale`；`cmd/azure-bonds-game` 編譯仍受容器缺 ALSA/X11 headers／Ebiten backend
限制，並非本輪核心測試失敗。

第二百二十三輪文件 commit：`bd31db2`，更新 Death／combatant removal READY spec、Gold Box
ECL／State 可重用知識庫、README 與 PLAN，記錄 DeathOverlay contract、目前 renderer fallback
與 exact skull 素材證據邊界。

第二百二十四輪重大里程碑 commit：`258fde2`，追證 `seg001.Init` 的 COMSPR icon
initialization：`combat_icons[24].GetIcon(Attack, 0)` 對應 `COMSPR` block `0x8B`，
`combat_icons[25].GetIcon(Normal, 0)` 對應 `COMSPR` block `0x19`。Ebiten 現在載入
COMSPR derived PNG，依 DeathOverlay signal 在死亡座標以 100ms phase 交替顯示原版
skull／blank overlay；更新 graphics／ECL／README／PLAN 知識庫。Docker Go 1.23 已通過
`internal/combat`、`internal/game`、`internal/party`、`internal/ecl`、`internal/locale`。

第二百二十五輪重大里程碑 commit：`f2372a7`，依 reference `Action.Clear` 將死亡後的
per-fighter `CombatAction`（delay、move、spell ID、guarding）建立成共用資料 contract，
Battle 在 HP=0 時清零；State 若倒下者正是 current turn，也清除施法、移動、檢視與 target
selection。新增 combat／State regressions，並更新 ECL／State READY spec、README、PLAN、
Gold Box knowledge base。Docker Go 1.23 通過 `internal/combat`、`internal/game`、
`internal/party`、`internal/ecl`、`internal/locale`。

第二百二十六輪重大里程碑 commit：`51ae23b`，補上 save／encounter 初始 HP=0 fighter
的 CombatantKilled boundary：`NewBattle` 立即清除 `HasCombatPosition` 與
`CombatAction`、發出 `DeathOverlay`，且 `StartRound` 不把倒下者放入 turns。新增
initially-downed regression，確認倒下角色不會佔用碰撞格；更新 README、PLAN 與 Gold Box
ECL knowledge。Docker Go 1.23 通過 `internal/combat`、`internal/game`、`internal/party`、
`internal/ecl`、`internal/locale`。

第二百二十七輪重大里程碑 commit：`e6aa0a2`，依 reference `Tile_DownPlayer=0x1F`／
`downedPlayers` 將 `Fighter.DownedCorpse` 與死亡 flash 分離：team party HP=0 角色保留
死亡座標但不佔用 CombatMap position；Ebiten 在 skull flash 後以繁中「倒下」corpse marker
顯示。Cure Light Wounds target routing 現可選到非 dead 的 unconscious／dying corpse，
普通治療同步 roster HP、清除 DeathOverlay，但不讓角色恢復戰鬥格；新增 combat／State
regressions。Docker Go 1.23 通過 `internal/combat`、`internal/game`、`internal/party`、
`internal/ecl`、`internal/locale`。

第二百二十八輪重大里程碑 commit：`a907601`，依 reference `combat_heal` 建立
`Battle.RestoreCombatant(fighterID, position)` explicit stand-up contract：只有
`DeathEffectContext.CombatHealAllowed` 的 affect_63 recovery，在 HP 恢復為 OK 後才以保存的
CombatX/Y 清除 `DownedCorpse`、恢復 `HasCombatPosition`；普通 Cure Light Wounds 仍只加 HP
並讓 corpse 留在原地。新增 Battle／State placement regressions，更新 ECL／State spec、
README、PLAN 與 Gold Box knowledge。Docker Go 1.23 通過 `internal/combat`、`internal/game`、
`internal/party`、`internal/ecl`、`internal/locale`。

第二百二十九輪重大里程碑 commit：`1041099`，建立 renderer-neutral
`combat.DeathOverlayFrame`：以 100ms cadence 交替 `COMSPR 0x8B`／`0x19` 九次後結束
flash。Ebiten 保存每個 fighter 的 flash start time；party 隨後顯示 `DownedCorpse`，enemy
停止繪製死亡小人，治療時清理 lifecycle state。新增 9-cycle core regression，更新 Death／
graphics／README／PLAN 知識庫。Docker Go 1.23 通過 `internal/combat`、`internal/game`、
`internal/party`、`internal/ecl`、`internal/locale`。

第二百三十輪重大里程碑 commit：`d40f877`，依 reference `RemoveFromCombat` 完成
Ebiten render lifecycle：enemy 的 DeathOverlay 九次 phase 結束後完全停止繪製名稱／HP，
並從戰場畫面退出；team party 的 DownedCorpse 則保留原座標與繁中「倒下」marker。更新
ECL／graphics／README／PLAN 知識庫。Docker Go 1.23 通過 `internal/combat`、`internal/game`、
`internal/party`、`internal/ecl`、`internal/locale`，`git diff --check` 通過。

第二百三十一輪重大里程碑 commit：`d5c6f8c`，依 reference `find_target`／
`BuildNearTargets` 建立 `Battle.SelectCombatTarget`：enemy turn 從 sorted、存活的 party
candidate 以 seeded RNG 選擇目標，同一回合的 multi-attack 維持同一 target，不再固定攻擊
party[0]。新增 combat／State deterministic regressions、READY spec 與 Gold Box state
knowledge；visibility／pathfinding／persistent Action.target／AI spell priority／guarding
仍保留明確 boundary。Docker Go 1.23 通過 `internal/combat`、`internal/game`、
`internal/party`、`internal/ecl`、`internal/locale`，`git diff --check` 通過。

第二百三十二輪重大里程碑 commit：`c5c0cc5`，以 headless Xvfb 啟動目前 Ebiten
`-encounter` direct-entry，擷取 640×400 實際繁中戰鬥畫面 `docs/screenshots/combat-game.png`，
並更新 README gallery、sprites manifest 與 live screenshot READY spec。畫面證明目前
renderer 可顯示繁中戰鬥訊息、party／enemy 小人、HP 與操作提示；明確標示這仍是
direct-entry vertical slice，不宣稱完整玩家流程。重新產生 parser screenshot，
`git diff --check` 通過。

第二百三十三輪重大里程碑 commit：`58668fa`，依 reference `PoolRadPlayer.field_33`／
`field_B5..B7` 保存 MON*CHA spell-list slots 與 magic-user level-use counts 到
`combat.Fighter`。敵方回合現在會先嘗試已核對的 Magic Missile `0x0F`：一級單枚、
2–5 damage、成功後 atomic 消耗 level-1 use，失敗或無可用次數才回到 physical attack。
新增 monster parser／combat／State regressions、READY spec、README／PLAN／Gold Box state
knowledge；其他 monster spells、MON*SPC effects、AI priority／range／saving throw 仍保留
明確 boundary。Docker Go 1.23 核心測試與 `git diff --check` 通過。

第二百三十四輪重大里程碑 commit：`7ece3a3`，依 reference `load_mob` 將
`MON1SPC`–`MON6SPC` 以 chapter-local monster ID 載入，並把九-byte raw affect records
以 copy 掛到 enemy fighter 的 `MonsterAffects`。新增 `BuildEnemiesWithAffects`、State
chapter table adapter、CLI loader、copy-isolation regression、READY spec 與 README／PLAN／
Gold Box state knowledge。隱形／加速／睡眠等效果的戰鬥語意仍未猜測。Docker Go 1.23
完整 `go test ./...` 與 `git diff --check` 通過。

第二百三十五輪重大里程碑 commit：`f703d80`，依 reference
`CanHitTarget`／`CheckAffectsEffect(Type_16)` 將 active monster affect `0x19`／`0x47`
投影為 combat target AC +4；inactive effect 不影響命中，raw record 不被消耗。新增
combat exact-boundary regression、READY spec 與 README／PLAN／Gold Box state knowledge。
其他 `MON*SPC` effect kinds 仍保留逐項證據 boundary。Docker Go 1.23 核心測試與
`git diff --check` 通過。

第二百三十六輪重大里程碑：依 reference `load_mob` 的 `field_A1` 解析
`MON*CHA[0xA1]` 為 `Record.AttacksPerTurn`，並依 `AffectHaste` 將 active
`MON*SPC` affect `0x27` 加倍接到 enemy fighter 的每回合攻擊次數。新增 raw offset／
active-inactive Haste regressions、READY spec 與 README／PLAN／Gold Box state knowledge。
完整 Docker Go 1.23 `go test ./...` 與 `git diff --check` 通過。

第二百三十六輪追加規則：依 reference `AffectSlow` 將 active `MON*SPC` affect `0x2A`
套用為每回合攻擊次數減半，並與 Haste 組合測試；目前 adapter 保留至少一攻下限。
movement half-actions、遠程彈藥與完整 weapon profile 仍保留 boundary。

第二百三十七輪重大里程碑：依 reference `Player.IsHeld`／`AttackTarget01` 接入 active
`MON*SPC` helpless／snake charm／paralyze／sleep（`0x1F`／`0x33`／`0x34`／`0x35`）。
Held enemy 會在 State enemy turn 跳過 physical／spell action；held target 由 combat core
套用 guaranteed-hit 例外，raw effect 不消耗。新增 combat、monster、State regressions、
READY spec 與 README／PLAN／Gold Box state knowledge。解除／豁免／持續時間／治療仍保留
boundary；Docker Go 1.23 核心測試與 `git diff --check` 通過。

第二百三十八輪修正：enemy 單次 physical attack 的中文訊息不再固定使用 `party[0]`，
而是依 `AttackResult.TargetID` 重新查找實際命中的隊員；當第一位 party fighter 已倒下、
第二位成為 target 時新增 regression，避免戰鬥規則與畫面文字分離。Docker Go 1.23
`internal/game`／`internal/combat`／`internal/monster` 測試與 `git diff --check` 通過。

第二百三十九輪重大里程碑：依 reference `ovr021.step_game_time`／`CheckAffectsTimingOut`
建立 `State.AdvanceGameTime` 七-slot raw clock adapter，採用
`{10,10,6,24,30,12,0x100}` 級聯進位與 slot→elapsed-minute conversion；party `.FX` 與
active battle raw effects 共用 timeout transaction，`Strength=0xFF` 永久 effect 保留，
slot-6 overflow 以 age cycles 保存。新增 clock normalization、slot-2 十分鐘換算、finite/
permanent party／battle effect regressions、READY spec 與 README／PLAN／Gold Box state
knowledge。完整 Docker Go 1.23 `go test ./...` 與 `git diff --check` 通過；REST interruption、
calendar UI、DOS age writeback 與完整 time-triggered ECL 仍保留 boundary。

第二百四十輪重大里程碑：依 reference REST loop 的 `step_game_time(1,5)`，將
`REST_START` 接到 `AdvanceGameTimeHours`，每 requested hour 推進 60 個 slot-1 minutes，
先處理 finite effect timeout，再執行既有每 24 小時 +1 HP natural healing。新增 REST
clock／effect-order regression、READY spec 與 README／PLAN／Gold Box state knowledge。
完整 Docker Go 1.23 `go test ./...` 與 `git diff --check` 通過；rest interruption、safe
location、spell-learning side effects 與完整 rest encounter table 仍保留 boundary。

第二百四十一輪重大里程碑：依 reference `CMD_EclClock`／`vm_LoadCmdSets(2)` 確認
`ECL CLOCK (0x34)` 必須吃 `timeStep`、`timeSlot` 兩個 operand；修正 command metadata，
新增 `ClockRequest` 與跨 `BlockSession` aggregation，並由 State 呼叫既有七-slot
`AdvanceGameTime`，讓 ECL 與 REST 共用 finite effect timeout。新增 ECL／State regression、
READY spec 與共用 Gold Box command-set knowledge。核心 `internal/ecl`／`internal/game`
測試通過；完整 `go test ./...` 在目前容器因缺 ALSA/X11 headers 的 `cmd/azure-bonds-game`
與 `internal/sound` build dependency 失敗，非本輪 Go logic failure。memory-backed operand
值、time-triggered event table 與完整玩家流程仍保留 boundary。

第二百四十二輪重大里程碑：將七-slot game clock 與 age-cycle overflow 接入 remake JSON
save version 5；`State.SavePartyFile`／`LoadPartyFile` 現在會保存並恢復時間進度，versions
1–4 仍可載入並使用零時鐘。新增 save round-trip regression、READY spec、README／PLAN／
Gold Box state knowledge。DOS SAVGAM Area1 clock raw offset、calendar UI 與完整 time-triggered
event table 仍保留 boundary。

第二百四十三輪重大里程碑：依 reference `Player.age @ 0x76`、Pool/Rad `age @ 0x30` 與
`NormalizeClock`，將 normal DOS player `.SAV/.GUY` age 接入 parser、`Character`、slot-6
overflow writeback 與 `PatchDOSPlayerRecord`；每次 slot-6 overflow 讓 party roster 每人加一歲，
並以 int16 saturation 防止 wrap。新增 parser／writeback／game clock regressions、READY spec、
README／PLAN／Gold Box state knowledge。Pool/Rad importer、age-based ability modifiers、
完整 DOS age UI 與多職業序列化仍保留 boundary。

第二百四十四輪重大里程碑：依 reference `StatValue.AgeEffects`／`Limits.RaceAgeBrackets`
建立 `Abilities.WithAgeEffects`，保存 dwarf／elf／gnome／half-elf／halfling／human 五段
age brackets 與六項 ability deltas，新增 human bracket regression。確認這是新角色生成
規則；既有 DOS player record 已含結果，故沒有把 helper 隱式接到 import／Fighter，避免
double-count。creation age UI、完整 race/class limits 與 Pool/Rad `0x30` importer 仍保留
boundary。

第二百四十八輪重大里程碑：依 reference `display_map_position_time` 與 Area1 raw field
mapping，建立 `State.GameTimeDisplay`／`GameTimeText` renderer-neutral 繁中 clock HUD；
一般畫面與荒野地圖現在顯示 `HH:MM` 及依七-slot scale 推出的日／月／年欄位。新增 READY
spec、回歸測試並更新 README／PLAN／Gold Box state knowledge；完整原版日曆規則與
time-triggered ECL 仍保留 boundary。

第二百四十九輪重大里程碑：將已驗證的 race/class restrictions、class minimums 與
`race_ages` single-class table 接入角色建立選單，從三個固定模板擴充為 22 個可驗證組合；
加入隊伍仍使用 copied-character age／ability transaction，Ebiten 以五列捲動顯示。新增
creation option regression、READY spec 與 README／PLAN／Gold Box state knowledge；
多職業、half-orc 與原版完整 create/modify/drop menu 仍保留 boundary。

第二百五十輪重大里程碑：依 reference `Gbl.RaceClasses`、`race_ages`、`Limits` 與 DOS
raw race `6`，新增 `RaceHalfOrc`、半獸人 age effects、DOS parser 與繁中建立選單的
cleric／fighter／thief 選項；依 `ClassId` index 修正 fighter `13+1d4` age。角色建立
目前共有 22 個已驗證 single-class 選項。新增 READY spec、parser／rules regression 與
README／PLAN／knowledge 更新；half-orc multi-class 與完整原版建立流程仍保留 boundary。

第二百五十一輪存檔／知識庫里程碑：整理 `SAVGAM?.DAT`、`CHRDAT{slot}{1..6}.SAV/.GUY`
與 optional `.FX/.SWG` sidecar 的可驗證邊界，確認角色 age 位於 `0x76..0x77` signed
little-endian，並加入 race `0x74`、class `0x75` 與 shared ECL flag mapping 文件；同時
完成 ECL `SAVE TABLE (0x35)` indexed write 與 regression。這些 raw-preserving contract
可供後續 Gold Box 作品沿用；完整 multi-class serializer 與未知 sidecar schema 仍保留 boundary。

第二百五十二輪主流程里程碑：修正 ECL `COMBAT (0x24)` 在 bounded VM 只回傳 next PC、卻
未保存 resumable runtime state 的問題；現在真實／synthetic battle victory 會接回同一個
ECL session，繼續 menu／picture／NEWECL／PROGRAM 9 結果。新增跨戰鬥 continuation regression、
READY spec 與共用 Gold Box ECL knowledge；完整各 block engine routine 與劇情 side effects
仍保留 boundary。

第二百五十三輪法術流程里程碑：將 `CAMP → MAGIC → CAST` 從 placeholder 改為可操作的
施法者／memorized slot／受傷目標三段選單；已核對的 Cure Light Wounds 會消耗 slot、擲
`1d8`、同步 roster／fighter HP，並以繁中訊息回到 MAGIC。SCRIBE、未知 spell rules、
高等級／多職業 slot 與完整施法時間仍保留 boundary。

第二百五十四輪玩家流程里程碑：新增 `CAMP → ALTER → RENAME`，以最多 15 bytes 的中文
input boundary 更新角色名稱、roster 與 fighter display name；SAVGAM writer 會保留角色
ID／sidecar basename 與未知 bytes，只 patch DOS name field。原版 code-page transcoding、
多職業 serializer 與完整 delete semantics 仍保留 boundary。

第二百六十二輪 multi-class rules 里程碑：新增 `Character.HasClass`，以保存的
`ClassLevels[8]` 判斷角色實際擁有的職業，舊 JSON／缺少 metadata 時退回 primary class。
`CanEquip`、CAMP MAGIC 與 combat cleric／magic-user gates 已使用此判定，Protection from
Good 也不再被 primary-class projection 誤導。新增 READY spec 與 Gold Box 可重用邊界文件；
THAC0、生命骰、高等級 spell capacity、部分 UI label 與完整 multi-class serializer 仍待
逐欄反組譯驗證。

第二百六十三輪工具里程碑：將已驗證的 DOS player age `0x76..0x77` 接成
`cmd/azure-bonds -set-age` 安全修改流程，要求明確 `-out-record`，不覆寫來源檔並保留
未知 bytes。新增 READY spec 與 CLI 文件；完整 SAVGAM slot replacement、sidecar atomic
transaction 與原版 player delete semantics 仍維持既有 boundary。

第二百六十四輪 ECL engine 里程碑：依 `ovr003.CMD_PartyStrength`／`CMD_PartySurprise`
接入 `PARTYSTRENGTH (0x1D)` 與 `PARTY SURPRISE (0x22)` 的 word destination request，
並讓 bounded VM／BlockSession 正常繼續與聚合。新增 READY spec、synthetic regression 與
共用 ECL command knowledge；實際 party stat calculation、AC scale、multi-class level
來源與 surprise result 仍待 State adapter 逐欄接線。

第二百六十五輪 ECL／State 里程碑：新增可注入的 `ecl.PartyContext`，由 State 使用目前
roster／fighter projection 解析 `PARTYSTRENGTH` 與 `PARTY SURPRISE`，將結果寫回 shared
ECL memory，並在 `NEWECL`／menu continuation 間保留。新增 context-resolved regression 與
READY spec；reference AC internal scale、完整 multi-class level／THAC0 table 仍保留邊界。

第二百六十六輪 ECL／State 里程碑：依 `ovr003.CMD_CheckParty` 接入 normalized selector
`0xA5..0xAC` thief skills、`0x9F` movement 與 `8001` active-affect branches；四個結果
destination 由 PartyContext 寫回 shared ECL memory，新增 READY spec 與 min／max／average
regression。未知 selector、NPC／temporary party 語意與各作品 scaling 仍保留 boundary。

第二百六十七輪 ECL engine 里程碑：依 `ovr003.CMD_Who` 接入 `WHO (0x39)` prompt request，
跨 `NEWECL` 聚合並繼續 cursor；新增 READY spec、command knowledge 與 no-prompt regression。
這輪刻意保留 State roster UI／selected-player transaction，沒有自動替玩家選第一位角色。

第二百六十八輪 State／ECL 里程碑：將 `WHO` 做成真正可恢復的角色選擇 transaction；
interactive VM 會停在 WHO，State 顯示繁中 roster，選擇後透過同一個 BlockSession resume，
並保存 selected player ID。新增 READY spec 與 State regression；selected player 對所有
DOS global routine 的完整 side effects、NPC／temporary party 語意仍保留 boundary。

第二百六十九輪 ECL／State 里程碑：依 `ovr003.CMD_LoadCharacter` 將 `LOAD CHARACTER`
從 raw address signal 擴充為 decoded `LoadCharacterRequest`，保留低 7 bits 的 1-based
player selector 與 bit 7 restore/redraw metadata；State 已映射到 persistent `partyRoster`
並與 WHO 共用 selected player ID，無效 selector 有明確 not-found regression。完整
`FreeCurrentPlayer`、party summary redraw、external string context 與 NPC 語意仍保留
boundary。新增 READY spec、command knowledge 與可跨 Golden Box 重用的 VM→roster adapter。

第二百七十輪 ECL string-memory 里程碑：依 `ovr008.vm_CopyStringFromMemory` 的明確
`0x7C00 == SelectedPlayer.name` 特例，將 `PartyMemberContext.Name` 接入 resumable
`RuntimeState.Strings`。現在 `LOAD CHARACTER` 後的 `0x81` string operand 可經
`COMPARE → IF/GOTO` 走姓名分支，新增互斥 success/failure regression、READY spec、README
與共用 command knowledge。原先「完整 external string context 未完成」已收斂為
「`0x7C00` 姓名已完成，其餘 DOS memory regions 仍待逐區驗證」。

第二百七十一輪 ECL inventory-query 里程碑：依 `ovr003.CMD_FindItem` 將
`PartyMemberContext.ItemTypes` 接入全隊 item-record 查詢；resolved `FIND ITEM` 現會
清空 compare flags並設定 `=`／`<>`，可直接驅動原始 `IF/GOTO`。同一 VM run 的
`DESTROY ITEMS` 也會更新 working inventory view，後續查詢不再看到已毀 type；persistent
roster mutation仍由 State 負責。新增 found／not-found／find-destroy-find regressions、
READY spec、README 與共用 Golden Box ECL knowledge，並移除舊「尚未設定 compare」斷言。

第二百七十二輪 ECL selected-affect 里程碑：確認 opcode `0x3F` 是 `FIND SPECIAL`、
`0x3D` 是 CLEAR BOX；新增 `RuntimeState.SelectedPlayerIndex/Set`，讓 LOAD CHARACTER 的
1-based selector 與 WHO 的 0-based UI selection 更新同一份可恢復 selected identity。
`FIND SPECIAL` 現在查 selected member 的 active effects、回傳 resolved request 並設定
`=`／`<>`。新增 LOAD CHARACTER branch 與 WHO pause/resume 第二角色 regressions、READY
spec、README、State／command knowledge；缺 context 或尚未選角仍維持 unresolved。

第二百七十三輪 real-image verification 里程碑：重跑原始 ECL1–ECL6 共 25 decoded blocks／125 個 initialization entries，全部以正常 EXIT、menu、COMBAT、PROGRAM boundary 或 NEWECL 返回，unsupported-opcode error 為零。新增 corpus regression，若 member／block count、entry framing 或 bounded semantics 退化會指出精確 member／block／entry／PC。另以 ECL5 block `0x30:+0x0014` 與含 item `0x5E` 的 PartyContext 驗證真實 FIND ITEM found branch 抵達 `SUNLIGHT` 裝備腐朽文字。更新 READY spec、README、ECL knowledge，並移除第 196 輪仍記載 `0x2D/0x2F` unsupported stop 的過時斷言。

第二百七十四輪 ECL party-departure 里程碑：依 `CMD_Dump`／`FreeCurrentPlayer` 確認
`DUMP (0x3E)` 會移除 selected TeamList member、釋放 icon、減 party size，並選前一位／
新第一位；空隊伍則清除 selection。VM 新增 ordered `DumpRequest` 與 mutable working
PartyContext，讓後續 inventory／affect／party-rule query看見移除後狀態；State同步移除
persistent roster與同 ID fighter，且不套 ALTER DROP 的 last-member guard。新增中間／
最後角色 regressions，並鎖定 real ECL5 block `0x30:+0x020E` 的 Akabar DUMP opcode。
補充 cross-NEWECL regression：BlockSession 使用 caller PartyContext 的 deep copy作為同一
session mutable working party，target block可見已離隊結果，而呼叫端 context保持不變。

第二百七十五輪 ECL／State 里程碑：依 `ovr003.CMD_Program` 將外部 routine 0/3/8/9
集中到 `State.applyECLProgram`。一般事件與戰鬥後 ECL continuation 現在共用 start-menu、
party-killed、game-won／全隊 HP 與健康恢復／存檔詢問，以及 CAMP transaction。DOS 勝利
後 process exit 在桌面重製版明確映射為返回標題；新增四 routine regression、READY spec、
README 與可跨 Gold Box 沿用的 VM／作品 adapter 分層知識。

第二百七十六輪 ECL／map 里程碑：反組 `CMD_Call` 的 `operand - 0x7FFF` dispatch，
將原始 image observed `0x2E10/0xC01E/0xB200` 對應到 redraw、forced
`MovePositionForward` 與 sound A/B。State 現會讓 `0xC01E` 依 0/2/4/6 方向在 16×16
座標 wrap，frontend exactly-once 重建 dungeon floor／wall／roof；`0xB200` 先重現
reference default sound A。新增 ECL3 block 16 real CALL、四方向 wrap、ordered request
regressions與 READY spec；`word_1EE76 == 10` sound-B transient 仍保留 evidence boundary。

第二百七十七輪 demo／NPC 里程碑：依 `CMD_AddNPC.vm_LoadCmdSets(2)` 修正 NPC ID＋
morale framing，新增 `NPCRequest` 與 `DELAY` timing signal；真實 ECL1 block 0x52
現執行 53 steps，加入 RUSTLE／CYNTHIA／GRENDEL、輸出 11 段青色枷展示文字、聚合
`CALL 0x6803`／DELAY，最後抵達 COMBAT。State 依 chapter-local MON*CHA／SPC／ITM
建立 persistent NPC Character／fighter、control morale 與最低空 icon slot；NPC 專用
parser以唯一 ClassLevel修正 stale class_id，普通 save import仍嚴格。PICTURE 後 deferred
combat transaction也已接通，11 段文字逐行翻成繁中。第 278 輪依 `sub_29758` 確認
0x52 僅供 `inDemo`，正常玩家流程不可加入這三名 NPC。

第二百七十八輪正式 new-game 里程碑：`FinishCharacterCreation` 在 production ECL
session 現會 fresh reset 到 global block `0x01`（ECL2），而非人造荒野 menu。真實 entry
載入 FILES `1,2,FF`／PIECES `1,2,3`，依序顯示「小房間醒來、裝備與記憶消失」及
PIC 0x0A「持劍手臂出現奇異圖紋、全隊相同印記」兩段繁中。State 將 picture deferred
boundary擴充到 menu，Ebiten圖片下方顯示三行漸顯文字；real regression從角色建立完成
鎖定 block identity、兩段文字、picture、menu與 pieces。配合繁中文字較大，Ebiten
邏輯畫布也由 640×400 擴為 640×480；88px PIC／人物圖以 nearest-neighbor 3×、BIGPIC
以 2× 整數像素放大，文字則以 24px 高解析字型重繪，下方保留三行訊息與獨立 Enter 提示列。

第二百七十九輪 commit `6384eef` 曾正確找到 `+0x1CBB EXIT` 與
`C04B/C04C/C04D = 7/13/1`，但因 remake Area 零值誤稱為 wilderness `ModeMap`。
第 280 輪依 `seg001.Init/InitAgain` 的 `inDungeon=1` 推翻該 adapter：正式起點是提爾佛頓
GEO2 block 1 的 DungeonMap，script half-direction 1 對應 renderer 東向 2。

第二百八十輪未提交成果同時修正 `CMD_LoadFiles` operand 次序：operand 1 是 dungeon
GEO selector，operand 3 才是 outdoor BIGPIC。BlockSession 新增五-entry lifecycle API，
EXIT 保存 shared memory writes；正式 dungeon 成功前進會同步 `C04B..C04F`，依序執行
per-turn／SearchLocation，並把文字、PICTURE、menu、combat signal 接回 State。
Ebiten 正式流程會自動顯示既有 GEO／WALLDEF／8X8D 3D renderer，不再要求按 D。
另外確認 `ovr011.SetupWildernessFloor` 的 50×25 buffer 是野外遭遇 combat floor，
不是世界地圖；README、PLAN、spec 63／67 與共用知識庫已清除舊斷言。正式地城 UI
已依 640×480／24px 中文重排左右圖像區與分行狀態列，font loader 補 TTC collection；
`-opening` 走過真實序幕後以 Xvfb 擷取 `docs/screenshots/tilverton-opening.png` 並更新 README。

第二百八十一輪成果：依 `TryEncamp` 接通 ModeDungeon 的 `E → PreCampCheck(entry2)
→ CAMP → optional CampInterrupted(entry3)`。真實 block 1 在 `(7,13)` 寫 rest encounter
`0/0`，`x<5 || y<13` 寫 `1/100`；unsafe 24h rest 只推進 1 小時即中斷，不套完整 healing
／memorization，執行皇家巡邏繁中事件，Continue 後回原 dungeon。CAMP EXIT 以
`campReturnMode` 返回 3D view；一般 640×480 event text 改成 24px 五行換行。新增 READY
spec 281 與 real-image regression。

第二百八十二輪成果：修正 VM 長期缺少的 ECL byte code-memory mapping；reference
`0x8000..0x9DFF` 現在隨 block load／NEWECL switch 重載，GETTABLE 可讀腳本內 dispatch
table，code window 外的 shared memory 仍保留。AND／OR 也依 `CMD_AndOr` 補上
`compare_variables(result,0)` side effect。正式序幕後由 GEO2 `(7,13)` 往西抵達
`(6,13)`，SearchLocation selector `0x86` 進入 Windlord's Inn PICTURE 3 與兩段繁中事件；
事件引用 Journal Entry 31 時才將使用者提供 PDF 的中文全文加入遊戲內手札，最後返回原
地城格。PICTURE opcode 同步保存稍後會被 script 清除的 HeadBlockId，讓 HEAD3／BODY3
原始人物素材正確顯示；手札改為 24px、22 字寬七行排版。新增 READY spec 282、
synthetic VM gates、real-image regression 與 640×480 `tilverton-inn.png` README 截圖。

第二百八十三輪成果：反組 `CMD_Rob (0x28)` 的三 operands 與 reference
`RobMoney/RobItems`，VM 現發出 selected/all-party、loss percent、item chance request；
State 對 Copper／Silver／Electrum／Gold／Platinum 逐欄向下取整，並依 inventory 順序、
重量 `>24/-50`、`>255/-90` 與 deterministic `1d100` 處理物品。DOS player
`0xFB..0x108` 七個 money/treasure words 全部進入 typed parse/project/writeback。
正式 Tilverton GEO `(6,5)` selector `0x8A` 現顯示 HEAD5／BODY5 賢者菲拉妮；
「是 → 如實相告」執行真實 `ROB 1,50,0`，解鎖使用者提供 Journal Entry 38 的三頁繁中
全文，經兩個 Continue 返回原格。新增 READY spec 283、ECL/save knowledge、
`-filani` 重現入口與 640×480 `tilverton-filani.png` README 截圖。

第二百八十四輪成果：反組 `ovr003.CMD_Combat` 與 `ovr007.CityShop`，確認
`COMBAT (0x24)` 會在無怪物的 normal context 依 Area2 `EnterShop／EnterTemple`
分派 engine service。VM 現輸出一次性 shop／temple signal並保存 next-PC；State
以同一結果的 TREASURE／ITEM block 建立商品，套用原版 shift 計價、角色五種硬幣
優先與 pooled-money fallback，購買 clone 不耗盡庫存。正式 Tilverton `(2,12)`
selector `0x84` 已完成 Weaponers PICTURE 4、YES／NO、ITEM2 block 5、購買與離店後
`MAY YOU ALWAYS STRIKE TRUE.` continuation，最後返回原格；General Store 的舊
「COMBAT」測試斷言也已修正為 CityShop。新增 READY spec 284、共用 command-set
知識、`-weapon-shop` 重現入口與 640×480 `tilverton-weaponers.png` README 截圖。

第二百八十五輪成果：依 ECL2 real scan 確認 GEO2 `(0,7)` terrain `0x92` 是剛德祭壇，
會聚合 PICTURE 6（HEAD2 `9`／BODY2 `6`）與 EnterTemple service。State 現先保留
PICTURE boundary，Enter 後進入神殿；`ovr005.temple_shop/temple_heal` 的十種治療、
固定價格、`1d8／2d8+1／3d8+3／Heal`、blind／disease／poison／curse／stone／raise
effect transaction 與 typed-coin payment 已接入，離開後 resume ECL 並返回原格。
Raise Dead 的 Constitution／多職業 max-HP penalty 保留明確 boundary。另因這個事件
首次證明 HEAD／BODY selectors 不同，新增可擴張 masked scene compositor，BODY `y+5`
再覆蓋 HEAD，修正舊預產同號圖造成的缺圖／裁頭。新增 READY spec 285、
`-temple` 入口與 640×480 `tilverton-gond-temple.png` README 截圖。

第二百八十六輪成果：ECL2 GEO2 `(5,2)` terrain `0x8C` 的 Hall of Training 已由
PICTURE 4／中文詢問接到場所限定 `PROGRAM 0`。依 `ovr018.train_player` 保存 DOS
`Player.exp @ 0x127` dword，加入 cleric／fighter／paladin／ranger／magic-user／thief
經驗門檻、1000 GP 角色付款、擲兩次 hit die 取高、Constitution 與 class-level／HP
成長。一般 `PROGRAM 0` 返回標題的既有語意不變。新增 READY spec 286、`-training`
重現入口與 640×480 README 截圖；高等級固定 HP、種族上限、完整多職業 CON 與升級
選法術當時維持 boundary；前三項已於第 287 輪補齊，目前只保留升級選法術與
dual-class HP gate。

第二百八十七輪成果：回讀公開 CoAB reference `sub_509E0`／`get_con_hp_adj`／
`Limits.RaceClassLimit`，補上六職業 hit-dice 上限後的固定 `+1/+2/+3 HP`、只對
未達 hit-dice 上限職業計算的 Constitution、多職業除數與 dwarf／elf／gnome／
half-elf／halfling 職業等級上限。正式角色建立整合測試現從真實 ECL2／GEO2
`(5,2)` 跑過 PICTURE 4、中文詢問、YES、場所限定 PROGRAM 0、角色確認、扣 1000 GP、
升級／HP 成長，再離開返回同一格。dual-class HP gate 已於第 288 輪完成，目前只保留
升級選法術 boundary。

第二百八十八輪成果：補回 DOS Player `HitDice @ 0xE5` 與既有
`multiclassLevel @ 0xE6` 的 Character／JSON／raw patch round-trip。訓練升級後會像
`ReclacClassBonuses` 以 active class level 更新 HitDice；若尚未超過 dual-class
舊職業等級，仍扣款並升級但不增加 HP，超過後恢復一般 HP 成長。新增抑制／恢復
regression 與 READY spec 288。升級選法術經 reference 確認還需要
`spellCastCount[class, spellLevel]` 篩選；此模型與選單已於第 289 輪接續完成。

第二百八十九輪成果：定位 DOS Player `spellCastCount[3,5] @ 0x12D..0x13B`，加入
Character／JSON／raw patch round-trip。訓練依 `MU_spell_lvl_learn` 與 ranger
`unk_1A758` 重算容量，再用 CoAB spell class／level metadata 排除超過 5 級、容量為零、
monster／cleric 與已知法術。magic-user 升級及 ranger 新等級大於 8 會顯示不可取消的
繁中法術選單，選一個加入 KnownSpells；9 級遊俠 regression 同時鎖定 druid 與
magic-user 候選。

第二百九十輪成果：GEO2 `(6,10)` terrain `0x88` 的真實提爾佛頓酒館已接通
PICTURE 4／HEAD4／BODY4、酒館動作與四種飲料選單。`LEMONADE → YES` 會走過紫色
腰帶女子及側邊騷動，調查後找到華麗火焰形匕首；Adventure Journal Entry 17 原本只有
插圖，因此遊戲內手札使用忠實的中文圖像描述，不杜撰額外劇情。real-image regression
鎖定完整分支、手札解鎖、EXIT 返回同格及 stale choice 清理。新增 `-tavern` 重現入口與
`tilverton-tavern.png`；事件畫面改成獨立 640×480 layout，原始圖 3× nearest-neighbor、
中文 24px 直接重繪，修正探索 HUD 與人物圖重疊。共用知識庫同步記錄 16×15／24×24
中文字級與圖片、文字分離 pipeline。

第二百九十一輪成果：定位 GEO2 `(1,10)` terrain `0x8F` 為剛德神殿高階祭司。
真實 ECL2 流程顯示 PICTURE 6／HEAD6／BODY6，YES 分支施展 Remove Curse 並記錄
Journal Entry 19。使用者提供的 Adventure Journal 掃描 PDF 證實完整內容為青色枷
發光、射出藍焰並令眾人劇痛，祭司因神力遭抵抗而停止；此中文全文只在事件發生後
解鎖。real-image regression 鎖定問答、兩段 press-button pause 與同格返回；新增
`-high-priest` 及 640×480 README 實機圖。事件 caption 由 34 個英文字元假設改成
每行 22 個 Unicode 字元，24px 中文不再超出右邊界。Gold Box command-set 知識庫也
修正 CALL 的過期「未實作」斷言，新增五層 opcode 支援矩陣、signal exactly-once
時序與 ECL／engine memory ownership 表。

第二百九十二輪成果：正式新遊戲 session 已跑通 Tilverton `(1,0)` 皇家馬車主線。
Weaponers、Filani 與第一次城門警告是原 ECL memory 條件；第二次進入才顯示 PICTURE 11，
國王聲音使青色枷發光並強迫隊伍攻擊。整合測試載入 `MON2CHA.DAX`，要求建立單一
test hero 加五名 Royal Guard 的 active battle，不再接受「戰鬥規則尚未完成」占位；
以真實 combat actions 勝利後，續跑紅袍人劫走假國王、YES／NO 投降、牢房、
PICTURE 2／HEAD2／BODY2 盜賊救援與 Thieves' Guild 描述，最後確認 `NEWECL 0x02`
及 `(1,12,0)` map registers。新增 `-carriage` 正式條件 bootstrap、READY spec 292、
完整繁中敘事與 640×480 `tilverton-carriage.png` README 實機圖。共用 ECL 知識庫新增
「location state → combat boundary → pauses → chapter switch」不得 fresh-reset 的契約。

第二百九十三輪成果：反組 ECL2 block 2 `+0x046B..+0x04BC`，確認原版以
`LOAD CHARACTER 10..13` 逐名讀 selected-player `in_combat @ +0x100`，再寫
`combat_team/quick_fight @ +0x10C = 0x80`，讓四名 THIEF 成為我方 AI 友軍。
VM 現能投影 selected TeamList player-window、跨 pause 保存 team writes，並把
單一 15 人 spawn 拆成 4 名我方與 11 名敵方；混合陣營 AI 會攻擊相反 side，
不再停成四個玩家回合。正式新遊戲 regression 已由 Weaponers、Filani、皇家馬車、
投降與牢房一路抵達公會，驗證 hero + 4 allied THIEF 對 2 FIRE KNIFE + 11 THIEF，
勝利後顯示繁中遺言並解鎖只有地圖圖像的 Journal Entry 4。ECL block 2 的 local
`(1,12,N)` 與 GEO2 combined `(9,3,S)` 已有雙向 renderer adapter。戰鬥 HUD 改為
640×480 專用畫面：24×24 原始小人 nearest-neighbor 2×，隊伍色條／選取框及下方
24px 中文名稱與 HP，十八組文字不再互相重疊。新增 `-guildmaster`、READY spec 293、
知識庫與 `tilverton-guildmaster-battle.png` README 實機圖。

第二百九十四輪成果：公會戰 ECL `PartyMask` 產生的四名 QuickFight THIEF 現標記為
temporary allies，戰鬥結束後連同屍體從 active party projection 清除，避免污染下一場
犬舍戰與 `PARTYSTRENGTH`。正式新遊戲 regression 已繼續走過抱豎琴半身人、
1 FIRE KNIFE＋依隊伍強度縮放的 FIGHTING DOG、猴籠、奧莉芙・拉斯凱托訪客簿及
綠色黏液門，全部加入繁中。依 reference 邊界移動與 ECL2 block 2 entry 0，
renderer 現只回報 passable boundary attempt，由 State 寫 `0x7ED5=1` 後照常執行
`CALL 0xC01E → NEWECL 3`，真實流程已進入提爾佛頓下水道而非硬指定 block。
README／knowledge base／READY spec 294 同步確立 640×480 logical canvas：
原始像素素材 nearest-neighbor 整數放大，繁中以 24px（緊湊欄位可用 16×15）
獨立高解析重繪。

第二百九十五輪成果：ECL2 block 2 的下水道出口現會先把 combined GEO
`(10,15,S)` 寫入 source registers，避免 local guild X 被錯誤減十成負數；`NEWECL 3`
後再讀回 target `C04B/C04C/C04D`，正式落在 GEO2 block 3 `(0,1,S)`。block 3
initial entry 的惡臭、黏液、低天花板與濕滑戰鬥環境已繁中化。real-image regression
接著抵達 `(1,8)` terrain `0x81` 的火刀檢查哨，拒絕投降、驗證五名
MON2 FIRE KNIFE、實際打贏戰鬥，再由同一 resumable ECL 顯示藏起屍體的繁中
continuation。新增 `-sewers` 全故事重現入口、READY spec 295；subagent 的 block 3
唯讀盤點也整理出五入口 lifecycle ABI、`C04F&0x3F → ON GOTO` terrain dispatch、
camp entry 分工與主要 encounters，已收斂進共用 Gold Box ECL 知識庫。

第二百九十六輪成果：正式新遊戲 regression 在五名火刀檢查哨戰後繼續前往
GEO2 block 3 `(13,10)` terrain `0x83`，跑過遭屠滅的檢查哨與迷斯卓諾騎士出場、
青色刺青質問，以及 FIRE KNIVES／PRINCESS NACACIA／NO ONE 三項效忠 menu。
繁中 display labels 保持原始 menu index；選「娜卡西亞公主」後，騎士提示別殺
拿戰鎚的牧師並放行。最後 Continue 會落實 ECL first-visit／friend state，返回
ModeDungeon；重訪同 terrain 已驗證不重播。新增 READY spec 296，知識庫補上
multi-pause dialogue 必須保存 pending PC、plot mutation 可能延後到最後 Continue、
localized label 不可取代 script key 的共用契約。

第二百九十七輪成果（第 516 輪勘誤後的歷史紀錄）：追蹤 ECL2 block 3 entry 0 後確認，下水道邊界除了
`0x7ED5` 還會先檢查 engine movement sentinel `0x7EC9`；公會轉場留下的 `0xFF`
若未在新步伐清除，會取消 exit attempt，接著把 E2 格誤派成 Otyugh 房間。
`RunDungeonExitLifecycle` 現在先同步 combined GEO、清除 stale sentinel、再交回
原始 ECL。當時的 regression 是從騎士分支以 script 直接寫入 `(8,15,S)`，再執行
`CALL 0xC01E → Y=0 → X-2 → NEWECL 4`，因此只證明 direct-entry 的 GEO2 block 4
`(6,1,S)` 初始化，不證明正常玩家已走到 E2；該正常路徑宣稱已由第 516 輪撤回。
target initial entry 的 `LOAD FILES 4,2,0xFF`、`LOAD PIECES 1,2,4` 與
`YOU ARE ENTERING THE HIDEOUT` 均在同一 session 聚合，入口文字已繁中化。
新增 READY spec 297，Gold Box 知識庫補上 boundary work flag 與 movement sentinel
是兩個不同 lifecycle 狀態的契約。

第二百九十八輪成果：盤點 ECL2 block 4 的 `C04F&0x3F` dispatch，確認 GEO2
terrain `0x99` 對應 selector `0x19` 的旋轉刀刃屏障。真實 ECL regression 鎖定
`ENTER THE BLADES / WAIT / RETREAT` 原始順序，並驗證 WAIT 不產生 DAMAGE、
刀刃減速消散的 continuation。State 新增「闖入刀刃／等待／撤退」、機關描述與
消散結果繁中；READY spec 298、README 與 Gold Box 共用知識庫同步記錄
640×480 畫布、24px 閱讀字級、16×15 緊湊字級及原始像素整數放大契約。

第二百九十九輪成果：補完刀刃屏障的危險分支。原始 ENTER index 0 先顯示
`THE BLADES TEAR INTO YOU`，下一個 press-button continuation 才送出
`DAMAGE flags=0xE0, dice=8d8, bonus=0, saveFlags=0`，最後與 WAIT 匯流到刀刃
消散。State 現只自動提交這種 whole-party auto-damage packet，以 seeded dice
對所有隊員套用同一傷害並同步 persistent roster／renderer fighter HP；選角、豁免
與 random-hit 形式維持既有 pending boundary。真實 ECL、State 兩人隊伍 100→62 HP
與 exactly-once consume 均有 regression，新增 READY spec 299。

第三百輪成果：接通 ECL2 block 4 selector `0x1A`／GEO terrain `0x9A` 的定身房。
原始 `RETREAT / INTERROGATE / KILL` 順序已繁中為「撤退／審問／殺死」而不改
script index；審問會先繳械逐漸恢復行動的火刀、取得情報並解鎖手札 26，殺死分支
則忠實顯示趁定身尚未解除時屠殺。手札中文依使用者 Adventurer's Journal 核對：
入侵牧師為營救南方首領房囚犯而施展定身，最後在此房被制伏。`4CFE & 0x40`
在選單前設定，因此三分支均消耗事件；真實 ECL 與 playable State regression 已
鎖定返回地城、手札不提前洩漏及重訪不重播。新增 READY spec 300。

第三百零一輪成果：補上原版地城 `SEARCH` 操作；640×480 地城按 `S` 時只在
SearchLocation invocation 期間設定 `7ECA=1`。火刀辦公室 GEO2 block 4
`(14,11)`／terrain `0x9B` 首訪描述房間並令 `4C10:0→1`，普通重訪無事；搜索才令
`4C10:1→2`、設定 `4CFE&0x80`，找到花梨木書桌文件並解鎖手札 9。手札中文依
使用者 PDF 第 12 頁圖像忠實記錄「燃燒靈氣、能附身其他軀體、與光芒之池有關」。
原始 `TREASURE(0,0,0,500,500,3,2,0x82)` 已接成 3000 GP 等值 pool、3 gems、
2 jewelry 與兩件 seeded random items；後續 COMBAT 正確視為 treasure service，
寶物 UI 返回 ModeDungeon。real-image／playable State regression 均鎖定防重複。

第三百零二輪成果：以 table-driven real-image／State regression 一次接通火刀據點
selector `0x1C–0x20`／terrain `0x9C–0xA0`。五個事件分別使用 `4C11..4C15`
visited byte：走廊奇怪煙味、整齊得異常且由看不見僕人復原的臥室、仍冒煙的焚毀
圖書館、遭同一烈焰完全摧毀的實驗室，以及標示「待復活／待埋葬」的兩排覆屍。
圖書館保留原始兩次 Continue；第二段取走焦屍手中未燒毀紙張後才解鎖手札 29。
中文依使用者 Adventurer's Journal 核對，保留盟友控制火焰、在軀體間移動、
異次元力量及「烈焰之主就是泰蘭索斯」線索。所有房間均驗證返回 ModeDungeon
且重訪不重播；新增 READY spec 302。
第三百一十輪成果：反組 ECL5 block `0x31` terrain `0x8A` 與共用子程序
`+0x0E0A`，完成 PICTURE 59 阿卡巴會面、YES／NO、`ADD NPC 0x3B,0x64`、
MON5CHA／SPC／ITM 入隊資料及解放前旅店。阿卡巴實際為 38 歲五級人類魔法師，
有兩件裝備、11 個 known spells 與 `4/2/1` 容量。子程序從 TeamList slot 0
逐人比對 `AKABAR BEL AKAS`，據此修正共用 `LOAD CHARACTER` 為 zero-based，
並新增 `Character.ScriptName`，使中文顯示名不再破壞 ECL script identity。
哈普解放後現會正確顯示阿卡巴的祕密商路提示。視覺契約仍為 640×480：
原圖／戰鬥小人 nearest-neighbour 整數放大，繁中正文 24×24 級、緊湊欄位
16×15 級，兩條 raster pipeline 分離。
第三百一十一輪成果：哈普地城出口現依 `4C5E` 提供地圖 CAVES 路線，
由真實 `NEWECL 0x32` 載入 Area 5 GEO block 50、pieces `8,FF,FF`，落在
`(15,5,W)` 的古老熔岩洞。入口伏擊使用 MON5 `0x39×4 + 0x31×3`，即四隻
火蜥蜴與三名黑暗精靈戰士；戰勝後保留同一 block 探索。修正 menu transaction
內 block switch 未清除來源 `7ED5/7EC9`，避免勝利後誤返荒野。新增
`-lava-tube` 真實 initial-entry 預覽與 640×480 `hap-lava-tube.png`；一般 ECL
menu 現能同時顯示 24px 中文 narrative，不再只剩選項。
第三百一十二輪成果：盤點 ECL5 block `0x32 +0x05B5` terrain dispatch，
確認 `ON GOTO` selector 1 才是第一個 target；terrain `0x8A` 第十項因此落在
`+0x10C6`，不是零起算會誤判的 `0x89`。真實 GEO5 `(9,10)` 現可觸發
火蜥蜴守門巡邏，使用 MON5 `0x39×3 + 0x31×3 + 0x33×1`。勝利後
`4C48 |= 0x08`，同一 resumable ECL 直接顯示繁中夢境警告並返回熔岩洞探索。
規格 312 與 Gold Box 指令集知識庫已保存 selector base 與戰後 presentation
不一定先停泛用勝利頁的契約。
第三百一十三輪成果：ECL5 block `0x32` per-turn 事件已在真實 GEO5
`(0,5,N)` 接通。terrain `0x89` 必須面北才顯示 PICTURE 57 的間歇泉與熔岩池，
接著保留 `COMBAT/WAIT/FLEE/PARLAY` 原始順序；已驗證 COMBAT 路徑建立 15 隻
MON5 `0x39` 火蜥蜴。勝利後 `4C48 |= 1`，發現六只防火桶；YES 進入 WHO，
一般英雄因熱度過強退回，再選 NO 返回洞內。繁中與 regression 已涵蓋每個
PICTURE／PRINT RETURN／menu／COMBAT／WHO boundary；知識庫新增方向敏感
per-turn terrain 與戰後環境志願者 selection 契約。

第三百一十四輪成果：重新對照 `CMD_EncounterMenu` reference 與 ECL5 block
`0x32 +0x01B7/+0x0281`，推翻上一輪「WAIT 直接進戰鬥」斷言。distance 0 的
WAIT／PARLAY 都把 behavior mode 4 解析為 result 3，進入五態度 PARLAY；
FLEE 的 mode 1 解析為 result 2，只有 COMBAT 才進 15 隻火蜥蜴。VM 新增
可恢復 PARLAY opcode，State 不再提前攔截 ECL FLEE／PARLAY；真實長流程驗證
WAIT→友善警告→無旗標離開，重訪後 COMBAT→戰鬥→防火桶。同步完成 640×480
renderer pass：24px 正文／16px compact 雙 CJK face、系統字型自動尋找、
dungeon 24×24 tile nearest-neighbor 2×、Combat／Dungeon HUD 分行與 Unicode
rune 換行，並移除 ModePlace 選單重複繪製。

第三百一十五輪成果：熔岩洞 GEO5 `(6,15,W)` 現依 block `0x32` per-turn
方向 gate 寫入 `C04B/C04C/C04D=7/15/3` 並 `NEWECL 0x33`，正式載入
GEO5 block `0x33`、pieces `14/15/FF` 與 PICTURE 51 五層法師塔庭院。
同一 resumable session 接續德拉坎德羅斯兩次 APPROACH、普通
`COMBAT/WAIT/FLEE/PARLAY` 選單、塔頂黑龍群、屠龍命令與煙霧幻象。
使用者提供的 Adventurer's Journal 條目 15 已整理成兩頁繁中，只在原 ECL
真正輸出 journal marker 後解鎖；事件再次寫入 `4CFF=1` 並令德拉坎德羅斯的
枷印消退，但不把已被火刀事件設定的同位址誤稱為第二枚計數器。最後停在
`ATTACK DRAGONS/ATTACK WIZARD/FLEE/PARLAY WITH THE DRAGONS` 真實 vertical
menu。新增 `-wizard-tower` 640×480 重現入口、READY spec 315 與共用知識庫。

第三百一十六輪成果：法師塔四項真實選單的 `ATTACK WIZARD`（index 1）已
沿同一 ECL5 block `0x33` session 接通。黑龍宣告不介入人類爭端後飛離，
德拉坎德羅斯召來守軍並逃下樓；戰場由 MON5 原始 records 建立 1 名伊弗利特、
2 名黑暗精靈戰士與 1 名法師。勝利後從原 COMBAT PC 續接「屋頂可安全休息」，
再由 EXIT 回 block `0x33` 地城。新增 `-wizard-tower-battle` 可重現入口、
READY spec 316，並在 Gold Box 知識庫保存一般文字選單不可按 label 提前攔截、
CLEAR MONSTERS 只清 encounter build list 的契約。

第三百一十七輪成果：法師塔 `PARLAY WITH THE DRAGONS` 已接入原始
`+0x05EA PARLAY 1,0,0,0,1,[7F79]`。五態度依序為
HAUGHTY/SLY/MEEK/NICE/ABUSIVE；傲慢與威嚇進入 14 隻 MON5 `0x35` 黑龍戰，
其餘三種會播放「沒有對付龍族的陰謀」繁中對話，再匯入德拉坎德羅斯四名
守軍戰，勝利後仍回安全屋頂。新增獨立 `gold-box-parlay.md` 知識文件、
READY spec 317、`-wizard-tower-parlay` 重現入口與 640×480 黑龍繁中實機圖。

第三百一十八輪成果：法師塔 outer menu 的 `ATTACK DRAGONS` 與 `FLEE`
均由原 ECL 匯入 MON5 `0x35×14` 黑龍戰，證實此處撤退不成功。勝利後保存
`7EC7 > 0x80` raw 重戰 gate；`4C61==1` 時可選擇取龍心，YES 會播放酸液繁中、
自動解析全隊 `DAMAGE 0xC0,3d4+3,save type 1` 並寫 `4C64=1`，NO／不符合
條件則跳過。State whole-party resolver 現可從混合 pending queue 取出 `0xC0`
packet 而保留需 selected target 的舊 packet。新增 READY spec 318、
`-encounter-area` graphics namespace 與 640×480 Area 5 原版 14 黑龍實機圖。

第三百一十九輪成果：法師塔塔頂 GEO5 block `0x33 (7,15,E)` terrain `0x01`
出口已由真實 ECL 接通。第一層保留 `CAVES/WILDERNESS/STAY HERE`；WILDERNESS
不再被泛用 label adapter 提前攔截，會繼續顯示 `VILLAGE/DEPART`，四條結果分別
鎖定 block `0x32/0x31/0x30` 或留在 `0x33`。NEWECL 後同步 destination GEO 與
target initial registers，新增 READY spec 319、`-wizard-tower-exit` 及
640×480／24px 繁中實機圖。共用 Gold Box 知識庫同步確立：原圖 nearest-neighbour
整數放大、CJK 直接在高解析 logical canvas rasterize，ordinary menu label 不可
脫離 block context 當成引擎 action。

第三百二十輪成果：法師塔祕道 DEPART 已沿真實 session 完成 ECL5 block `0x30`
離場程序。`LOAD CHARACTER` 現將 party `ControlMorale` 投影到 selected-player
`0x7CB8`，阿卡巴不再因只有姓名投影而被錯判不存在；完成塔與哈普時，他會以
繁中告別並由原 DUMP 離隊。下一個獨立 Continue 顯示日光使黑暗精靈裝備腐朽，
並銷毀 item `0x5E/0x60/0x61`。最後 `NEWECL 0x50` 顯示 BIGPIC 121，回到
ENTER CITY／JOURNEY ON／CAMP 世界流程。新增 READY spec 320、synthetic
selected-window regression、real block 0x30 regression 與完整塔頂長流程驗證；
Gold Box 共用知識庫同步記錄 control/morale projection 與 DEPART cleanup 契約。

第三百二十一輪成果：Area 5 離場回到哈普後，`JOURNEY ON → ESSEMBRA → TRAIL`
現沿 ECL1 block `0x50 +0x149A` 顯示龍巫妖復仇繁中事件並進入戰鬥。此 script
雖位於 ECL1，卻 `LOAD MONSTER 0x3C,1,0x3C`，實際 record 是 MON5 DRACOLICH；
State 因此改為逐 spawn 依 monster ID range 選擇 MON*CHA／MON*SPC，ECL chapter
只作 fallback。MON5 record 解出 66 HP、raw AC 66→AC -6、3d8，並使用 CPIC5
原版小人。勝利後正式抵達艾森布拉城外。新增 READY spec 321、real-image 與
法師塔至艾森布拉長流程 regression，並更新 640×480 README 實機圖。

第三百二十二輪成果：依 DOS 原版城市／戰鬥截圖重建 640×480 畫面拓撲。
冒險畫面恢復左圖、右隊伍 AC／HP、下方敘事與最底命令列；戰鬥恢復左戰術格、
右 active／target 狀態、下方訊息與命令列。reference `draw_head_and_body`
的 `row+5` 已由錯誤 5px 修正為五個 8px 列（40px），並以 HEAD→BODY layering
修復臉貼在胸口。戰場改為 clipping target，不讓大型怪物越入狀態欄；在
combat terrain selector 尚未解出前使用中性戰術格，不再誤鋪 TILES icon atlas。
新增 READY spec 322、Docker DOSBox reference captures、兩張 Xvfb 實機圖，
並把可沿用規則寫入 Gold Box graphics knowledge。

第三百二十三輪成果：設計審查以 DOS `combat-aim`／`fight-black-dragon`
逐像素量測，推翻第 322 輪仍過高的戰鬥面板。上方改回原生 320×184 的精確
2× geometry：戰場 `(16,16,336,336)`、16px 中央石框、256px active status；
640×480 多出的 80px 僅作中文 log，最後 32px 保留兩列 footer。移除可見
checkerboard、紅藍 team bars 與右欄 target card，改用 EGA 灰底、青／綠／黃
資訊層級及 48×48 active cursor。大型敵人 occupancy 未解出前不顯示錯誤的一格
target box。新增 READY spec 323；terrain、原始 stone-frame tiles 與大型怪物
anchor 明確保留為下一輪 visual RE boundary。

第三百二十四輪成果：確認原版戰鬥 terrain 不在一般 TILES，而是
`DUNGCOM/WILDCOM/RANDCOM`。三個單 block payload 均為 17-byte SSI header 加
24×24 4bpp items，CoAB 分別為 25／34／6 張；新增 bounded codec、原始檔
regression 與三套 gallery。Ebiten 地城戰場現由既有 `GenerateDungeon`
50×25 background buffer 取 7×7 slice，再依 `BackgroundTile.TileIndex`
繪製真正 DUNGCOM 石牆／斜牆／轉角，取代單色 placeholder。完整 encounter
terrain-mode selector、WILDCOM procedural placement、RANDCOM decoration、
原始 stone-frame tiles 與大型怪物 occupancy 仍保留為後續 RE boundary。

第三百二十五輪成果：正式接通 WILDCOM 野外戰場。renderer 以 `MapX/MapY`
為中心查 `SetupWildernessFloor` 已還原的 50×25 buffer，將 7×7 entries 的
`TileIndex 0..33` 對應 WILDCOM 34 張原始 tile；實機圖已顯示樹、倒木、岩石、
草地與水岸。terrain family selector 改為只依 `Area.InDungeon` 選
DUNGCOM／WILDCOM，不再使用 `GameArea>1` heuristic；`-combat-terrain`
只保留作 deterministic visual verification。RANDCOM 六張特殊物件明確保持
decoration overlay boundary。新增 selector／camera tests、READY spec 325 與
640×480 野外戰鬥截圖。

第三百二十六輪成果：由 reference `CombatMap.size = player field_DE & 7` 與
MON1–MON6 原始 records 收斂怪物形狀碼：1／2／3／4 分別是
1×1／1×2／2×1／2×2。`CombatSize` 現由 monster parser 投影到 fighter，
移動、復活、鄰接與 camera 均使用完整矩形 footprint。Ebiten marker 依同一
shape 顯示；2×2 龍巫妖為 96×96，CPIC 鏡像仍保留 `6-x` 左上錨點，避免因
額外扣除寬度而被戰場 clipping。更新 READY spec 326、Gold Box graphics
知識庫與 640×480 DUNGCOM 龍巫妖實機圖。

第三百二十七輪成果：接通已存在但先前未繪出的 RANDCOM 原版裝飾 pass。
reference `sub_370D3` 在 GEO terrain bit `0x40` 的開放區，以 dice stream 寫入
table／chair BackgroundTiles entries `0x1A/0x1B`；其 graphic ID
`0x22/0x23` 屬於全域 namespace，應映射 RANDCOM `0/1`，不是拿去查只有
25 張的 DUNGCOM。renderer 現先畫 DUNGCOM floor `0x16`，再透明疊加
RANDCOM `id-0x22`；WILDCOM `0..33` 保持獨立。原始 catalog 掃描與
`GEO2 block 01, center (13,0), seed 1` 的 640×480 Xvfb 圖均確認桌椅可見。
新增 READY spec 327、atlas bounds tests、`-dungeon-x/-dungeon-y` deterministic
visual flags，並同步 README 與 Gold Box graphics 知識庫。

第三百二十八輪成果：將 BackgroundTiles 的 `MoveCost`／`0xFF` 從畫面資料接入
可玩的 combat MOVE transaction。新增 renderer-neutral `MovementTerrain`
callback；Battle 在 occupancy／attack／座標 mutation 前檢查目的地完整
footprint，任一格不可通行即原子拒絕，多格 cost 取最大值，並在 remaining
points 不足時保持位置與 move mode。State 改依 `MoveResult.MovementCost` 扣點。
作品 adapter 分流 reference `x≈22,y≈10` 絕對 CombatMap 座標與目前 0..6
formation fallback；地城查 `(18+x,7+y)`，野外查 MapX/MapY centered floor，
coordinate namespace 在 StartCombat 固定，不會移動途中重新猜測。新增 READY
spec 328、2×2／2×1 terrain regressions、State budget regression，並更新
README 與 Gold Box state 知識庫。

第三百二十九輪成果：移除 640×480 renderer 的 3px 仿石紋戰鬥框。兩張公開
DOS oracle 的 frame／divider／裂紋位置完全相同，原始 ZIP 94 members 又沒有
獨立 UI frame DAX，因此把 boundary 修正為固定 320×184 panel raster，而非尚待
尋找的 encounter stone tiles。新增 `gfx.CombatFrame()`：透明 battlefield／
status interiors、五個 8px frame regions、原生 1px EGA bevel、alternating
dotted inner edge 與固定 crack pixels；Ebiten 啟動時轉一次並 nearest-neighbour
2×。新增 native geometry／transparency tests、READY spec 329，更新 Gold Box
graphics 知識庫與 README 的 640×480 龍巫妖實機圖。

第三百三十輪成果：地圖引擎拆分。原本位於 CoAB `internal/geo` 的完整
16×16 GEO decoder／牆／門／移動規則，以及 `internal/gfx` 的原版
Draw3dWorld 遠中近視角走訪，已移至獨立 `golden-box-remake-engine`
的 `geometry`／`viewport` package；CoAB 僅保留相容 wrapper。game-pack
schema 新增 `maps`，散提爾堡內城以 JSON 宣告 area 4、GEO block `0x20`、
spawn `(2,0,S)`、wrapped 與 2× nearest-neighbour。下一步是 WALLDEF/8X8D
composition/resource loader 搬移，以及 640×480 第一人稱地圖實機截圖。

第三百四十二輪成果：獨立 engine 新增 `graphics`，接管 SSI indexed picture、
EGA RGBA、WALLDEF、LOAD PIECES global offsets 與 8X8D stamps；中立 block map
消除對 CoAB `internal/dax` 的反向依賴。CoAB map JSON 現指定 GEO/WALLDEF/8X8D
檔名並驗證 base filename。該輪曾把 recovered `-5..15` wall traversal columns
誤當成 176px panel 寬；第 347 輪 DOS oracle 已證實實際 chrome 是左 128px、
右 192px。舊 debug floor 不再混入 production。Docker/Xvfb 實機圖為
`docs/screenshots/tilverton-first-person-remake.png`。door／roof overlays、
斜向逐像素 DOS oracle 與 wilderness world map 仍是後續 map fidelity 邊界。

第三百四十三輪成果：世界地圖改讀原始 `BIGPIC1 block 0x79`，不再把
WILDCOM 50×25 combat floor 誤稱 overland。使用者提供的 Clue Book PDF 第 35
頁與攻略確認 CoAB 只能在興趣點間旅行。獨立 engine schema／`worldmap` 支援
作品中立 image、localized points 與 cardinal selection；CoAB JSON 保存 A–N
14 個 values／座標／翻譯。正式 `ModeWilderness` 顯示 608×240 nearest-neighbour
地圖、目前位置、旅行選單與繁中 HUD；`-world-map` 的 Docker/Xvfb 實機圖為
`docs/screenshots/coab-overland-map-remake.png`。route graph 自 ECL 匯出、
Shadowdale AREA overhead map 與 optional travel encounters 尚待後續。

第三百四十四輪成果：補上與世界地圖、戰鬥地板分離的 AREA 俯視地圖。獨立
engine `areamap.Project` 由 16×16 GEO grid 產生 terrain cells、去重實體牆段、
wall type 與 door detail；engine commit 為 `2ef18ca`。CoAB game-pack JSON
新增 `tilverton.area-map`，指定 Area 2、`GEO2.DAX` block 1；
與 2× scale。正式地城按 A 開啟、A／Esc 返回，舊 `ModeMap` 畫面也不再誤用
WILDCOM combat floor。中文字改讀本機倚天 `STDFONT.15` Big5 分區字模，
以 Monkey Island 2 已驗證的逐列水平 1px embolden 顯示 16×15 粗體；
optional `SPCFONT.15` 處理全形符號，字型檔因著作權不提交。640×480 實機圖為
`docs/screenshots/coab-area-map-remake.png`。本輪的向量 renderer 與
`8X8D2/01` 推測已由第 345 輪原版 symbol renderer 取代。

第三百四十五輪成果：依公開 reference commit `9dc46f1` 的
`ovr031.DrawAreaMap`／`seg001` 還原原版 AREA。全域 symbol set 4 在
`game_area=1` 時載入 `8X8D1.DAX/CA`；11×11 camera offset 為
`clamp(party-5,0,5)`，每格 N/E/S/W wall presence 組成 `1/2/4/8` mask，
選 local item `4+mask`，隊伍方向選 item `direction>>1`。原版不顯示門；
reference door pass 是 cheat。獨立 engine `443281a` 新增
`areamap.BuildOriginal` 與 schema `symbol_block`。CoAB renderer 現直接以
2× nearest-neighbour 畫 8×8 EGA 灰牆與白色方向箭頭，取代向量 16×16 全圖；
JSON 指定 `8X8D1.DAX/CA`。更新後的 640×480 倚天粗體實機圖仍為
`docs/screenshots/coab-area-map-remake.png`。

第三百四十六輪成果：修復正式第一人稱 viewport 的背景與 screen transform。
公開 reference `ovr031.Draw3dWorldBackground`／`seg040.DrawColorBlock` 證實
native sky/horizon/ground 為 `(24,24,88,44)`、`(24,68,88,2)`、
`(24,70,88,42)`；SKY FA／FB 依戶外 palette、hour、方向顯示，FC 固定地面
overlay。獨立 engine `8ea72d9` 新增 reusable projection 與 schema
`sky_file/sky_blocks`；CoAB JSON 指定 `SKY.DAX [250,251,252]`，原始 image
regression 驗證 88×16／24×24／88×48。另修正 wall stamp：只保留 logical
row/column 0..10，native position 是 `(column+3,row+3)×8`；舊 renderer
錯誤右移 48px、上移 32px，才形成 README 舊圖的三片巨牆。`-eten-font`
現在同時接管 regular/compact face，全畫面使用倚天 16×15 embolden。最新
Docker/Xvfb 圖已覆寫 `docs/screenshots/tilverton-first-person-remake.png`。

後續 DOSBox oracle 進度：已完整走到原版角色建立的 stats／命名／戰鬥小人
配色流程，確認選單需以游標與彩色功能鍵混合操作。實機新建 male dwarf
fighter 顯示 `AGE 55`（另一輪為 52），原生 320×200 證據保存為
`docs/reference/original-dos/character-age-create.png`，並補入 spec 251 與
Gold Box save-format 知識庫。角色 pool 的下一個阻礙是原版要求 A: floppy；
`/tmp` DOSBox harness 已嘗試以同一暫存目錄掛載 A:，尚未取得可載入 party，
不應把黑畫面誤列為 adventure oracle。倚天字型則已確認與 Monkey Island 2
`build_eten_font.py` 完全同構：原生 `STDFONT.15` 16×15，每列將 source pixel
向右 OR 1px；不做容易黏筆劃的 24→16 縮圖。README 已將此啟動方式列為建議值。

第三百四十七輪成果：發現原版標題的 `D` 可直接進入內建 demo，無須先完成
A: character-pool 流程；因此取得真正 native 320×200 冒險 chrome oracle
`docs/reference/original-dos/tilverton-first-person-demo.png`。畫面訊息明示
`NOWHERE IN THE REAL...`，只可用於 GUI／SKY／status layout，不可當 GEO2/01
牆配置證據。實測 top row 在 native x=128 分割、y=136 進入 message：
640×480 remake 現為 first-person 256×272、roster 384×272、message
640×176、footer 640×32；多出的 80px 僅擴充繁中訊息。CoAB JSON 新增
`tilverton.first-person`，明列 GEO2/01、WALLDEF2、8X8D2、SKY FA–FC、
spawn `(7,13,N)` 與 outdoor sky selector 3。獨立 engine `908cfb7` 已推送，
新增 `FindMapByKindLocation` 與 indoor/outdoor sky schema，使同一 geometry
block 的 AREA／first-person projections 不再互相誤選。正式 `-opening`
Docker/Xvfb 圖已更新 `docs/screenshots/tilverton-first-person-remake.png`。
