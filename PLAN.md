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
- [ ] 對全部實際 ECL block 驗證五個入口，並與事件文字對齊。

## 第十三輪驗收

- [x] 建立有步數上限的 ECL subset runner。
- [x] 實作 `SAVE/COMPARE/IF/GOTO/GOSUB/RETURN/PRINT` 與 packed text 輸出測試。
- [x] 未支援 opcode 以精確 payload offset 停止，未宣稱完整 VM。
- [ ] 實作並驗證 `ON GOTO/GOSUB`、menu input 與完整 memory model。

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

## 第二十三輪驗收

- [x] game command 載入 ECL1 全部 decoded blocks。
- [x] `NewStateFromECLBlocks` 接入 BlockSession。
- [x] State selection sequence 與 session offset 接通。
- [x] 以真實 NEWECL transition 做回歸，保存 bounded 跨 block memory／call stack。

## 第二十四輪驗收

- [x] `BlockSession.RunInteractive` 接入 game runtime。
- [x] selection offset 與 bounded NEWECL switch 有 synthetic regression。
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
- [x] 為已觀察 monster item/effect IDs 加入繁中顯示與未知 fallback。
- [x] 將真實 ECL spawn sequence 合併 MON*CHA，建立 24 個 enemy fighters 並在 COMBAT 邊界停止。
- [x] 建立可操作 party／enemy Battle state、回合攻擊、勝負轉移與 Ebiten 繁中戰鬥畫面。
- [x] 將 bounded ECL `COMBAT` result 的 encounter descriptors 與 `MON*CHA` records 接到 Battle，並提供 ECL1 direct-entry 驗證入口。
- [x] 對 `PROGRAM 0/3/8/9` 建立外部 routine boundary signal，避免把 CAMP／勝利／死亡流程錯誤當作 ECL no-op。
- [x] 加入繁中遊戲內冒險手札、中文歷史筆記，以及 `PROGRAM 9` 到 CAMP state 的可測試控制邊界。
- [x] 將冒險手札摘要做成八頁可翻閱的繁中遊戲內資料，接通 J／方向鍵／Esc 導航。
- [x] 保存 party roster、同步戰鬥 HP，並建立 CAMP 後 HP 恢復的 state boundary。
- [x] 建立六種玩家種族、六種基本職業、能力值與 1–6 人 roster validation。
- [x] 接通繁中角色建立 starter UI，完成後將角色投影並保存到 `State.SetParty`。
- [x] 接通 Unicode／繁中自訂角色姓名輸入與 State 保存。
- [x] 接通六項能力值的繁中編輯、3–18 bounds 與職業最低值 validation。
- [x] 加入六項能力值的 3d6 擲骰／重擲，並保留 seed regression。
- [x] 建立版本化 remake party JSON，接通 F5／F9 與啟動載入；原版 DOS save/import 仍待反組。
- [x] 將 ECL `COMBAT` 在有 party／MON*CHA records 時接到可操作 Battle；缺少資料時保留安全 boundary。
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
- [ ] 將 DOS player record icon mapping、direction-specific position 與八方向 combat placement 接入 renderer。
- [ ] 將 Area codec 接入完整 DOS save/import slot 與所有 file side effects。
- [ ] 從完整玩家流程抵達該 entry，驗證跨 block context 的完整劇情 continuation。
