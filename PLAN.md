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
- [ ] 建立跨 ECL1–ECL6 DAX block session。

## 第二十二輪驗收

- [x] 建立 decoded ECL `BlockSession`。
- [x] 驗證 ECL1 block `0x50/0x51/0x52` initial entries。
- [x] 驗證 NEWECL target switch contract。
- [ ] 將 session 接入 game runtime，保存跨 block state。

## 第二十三輪驗收

- [x] game command 載入 ECL1 全部 decoded blocks。
- [x] `NewStateFromECLBlocks` 接入 BlockSession。
- [x] State selection sequence 與 session offset 接通。
- [ ] 以真實 NEWECL transition 做回歸，保存跨 block memory／call stack。

## 第二十四輪驗收

- [x] `BlockSession.RunInteractive` 接入 game runtime。
- [x] selection offset 與 bounded NEWECL switch 有 synthetic regression。
- [x] 確認 ECL1 initial-entry graph 尚無 reachable NEWECL edge，保留此未知。
- [ ] 從其他 event entries 找出 real NEWECL transition。

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
- [ ] 從完整玩家流程抵達該 entry，保存跨 block memory／call stack。
