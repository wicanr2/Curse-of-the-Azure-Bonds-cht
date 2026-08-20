# 完成 remake 所需的 TODO 盤點

狀態日期：2026-08-16

## 這份清單的口徑

驗收標準沿用既有文件，不另立一套：

- 完成的定義見 [`../../AGENTS.md`](../../AGENTS.md) §1 與 §8——**原版忠實模式是驗收基準**，
  prototype、vertical slice、測試通過或截圖都不算完成。
- 每一項的層級用 [`coab-re-coverage-matrix.md`](coab-re-coverage-matrix.md) 的
  `R1` 原始定位 → `R2` 原版語意 → `R3` READY spec → `R4` engine＋資料 → `R5` 玩家驗證。
  本清單只是把矩陣的「仍缺」欄拆成可執行、可勾選的條目，欄位語意與矩陣一致。
- 函式覆蓋沿用 `docs/audit/re-function-ledger.json`（人工判定唯一來源）。

### 兩平台的分工

| | 角色 | 用途 |
|---|---|---|
| **DOS** | **主線與行為 oracle** | 規則數值、事件、時序、存檔格式、玩家路徑驗收一律以 DOS 為準；兩平台衝突時採 DOS |
| **PC-98** | **引擎語意參考** | 帶完整 Borland 除錯符號（49 個遊戲單元名 + 型別/record 版面），用來替 DOS 的無名函式命名、確認 record 欄位邊界、交叉驗證推論 |

PC-98 的作用是**降低 DOS 側的推論成本**，不是交付目標。已建立的機制：
`internal/borlanddebug`、`cmd/borland-symbols`、`cmd/pc98-symbol-audit`，
比對方法見 [`../spec/783-*`](../spec/) 系列與 spec 1094（助憶碼序列比對）。

⚠ 已知兩平台會分岔的地方（**取 DOS**）：角色記錄 `1A6h` vs `1A7h`、
物品節點 `3Fh` vs `67h`、選單節點 `2Eh` vs `56h`、存檔第 5 塊 `1E00h` vs `1E41h`、
DOS 的多職起始年齡表 207 條在 PC-98 只剩 14 條（spec 1094）。

## 已定案的範圍決策（使用者 2026-08-16）

這五項不是從程式碼推得的，是使用者的取捨。它們直接縮小或改變了下面的條目，
不要再依舊描述執行：

| # | 決策 | 影響 |
|---|---|---|
| 一 | **完成內容產出** | `ENG-01` 是主要工作量，照矩陣的區域順序推進 |
| 二 | **game pack 以分檔方式處理** | `ENG-02` 的方式已定案，只剩「依 DAX 或依區域」與 schema 邊界要決 |
| 三 | **補完戰鬥所缺** | `RE-06`／`RE-07`／`ENG-07`／`ENG-08`／`ENG-09` 全部在範圍內 |
| 四 | **remake 讀舊版 DAT、存成自己的格式，不必互通** | `ENG-10`／`VER-07` **移除雙向 round-trip 與寫回原版的要求**；只保留「讀得進來」與 remake 自己的存檔完整性 |
| 五 | **繁中化地基優先** | 見下方 C 區的重新界定 |
| 六 | **讀完 21 支未讀 ECL handler、清掉台帳孤兒；PC-98 與 remake engine 無關的部份不必解讀** | PC-98 的角色從「語意骨幹」收斂成「只讀 remake 需要的部分」，不再追求全模組語意閉合 |

## 現況量測（2026-08-16 實測）

數字一律現場量，不沿用上一輪的文件行。

| 指標 | 數字 | 來源 |
|---|---|---|
| 函式覆蓋台帳 | 2,874 個函式，**已解讀 2,137／不阻塞 162／邊界碎片 575／待解讀 0／孤兒 0** | `coab-function-index.md`（可重生） |
| ├ DOS | 1,386：已解讀 1,016 ／ 不阻塞 133 ／ 邊界碎片 237 | 同上 |
| ├ PC-98 | 1,488：已解讀 1,121 ／ 不阻塞 29 ／ 邊界碎片 338 | 同上 |
| └ 台帳孤兒 | **0**（原有 48 列位址對不上任何函式起點，內容都是 spec 569 的樣板分類，2026-08-16 依決策六刪除） | 同上 |
| └ 證據等級 | `exact` 1,955 ／ `strong inference` 223 | 同上 |
| 規格文件 | **1,115** 份 `docs/spec/*.md`（2026-08-18 重量） | `ls` |
| remake 程式（不含巢狀 repo） | **309 個 `.go`／94,951 行**（2026-08-18 重量） | `find`／`wc` |
| ├ `internal/game` 機制碼 | **16 檔／13,908 行**（非測試） | `wc` |
| └ 檔名 | `state` `combat_state` `creation` `creation_guided` `training` `shop` `temple` `time` `spells` … **沒有區域／劇情專屬檔** | `ls` |
| 共用 engine | 獨立 repo，69 個 `.go` | `golden-box-remake-engine/` |
| 內容資料 | `gamepack/pack/` **四檔**：core 46 KB／content 86 KB／locale.en 66 KB／locale.zh-TW 72 KB | `ls -la` |
| ├ 事件 | 4 個 `events`、113 條 `option_rules`、**1,009 條 `text_rules`** | `json` |
| └ 語系 | `en` **1,269** 條、`zh-TW` **1,269** 條，**一一對齊、沒有漏譯**；譯名一致性由 `internal/glossary` fail-closed 擋住（93 詞條、0 不一致） | `json` |
| UI 詞條 | `assets/locale/zh-TW.json` **872** 條 | `json` |
| 建角規則表 | `gamepack/rules/character-tables.json`：7 種族、17 職業組合、8 職業槽、擲點與體質加值 | `json` |
| 法術主表 | `gamepack/rules/spell-table.json`：原作 100 筆，16 個位元組全部有出處（spec 1111）；占位 13、只能紮營 8、戰鬥可施放 **79** | `json` |
| 原作事件總量 | 6 DAX／25 block／125 lifecycle entry／**4,222 個靜態可達 instruction** | `cmd/ecl-event-catalog` |
| ECL 靜態可達 instruction | **4,222**（spec 1106 補上 `IF` 的 else 路徑後，由 1,355 增為三倍） | `ecl-event-catalog.md` |
| ECL 副作用候選 | **154** 個中 33 個已審（31 筆依效果序列沿用） | `ecl-ordered-effect-reviews.json` |
| ECL opcode commit phase | **55** 個 corpus opcode 中 25 支 handler 已讀、**30 支 `unknown`** | `ecl-opcode-effect-phases.md` |
| **原作文字段落覆蓋** | **1,022** 頁**控制流可達**，`matched` **999**／**`unmatched` 0**／`variable-insert` **16**（頁裡印的是執行期的值，靜態驗不到）／`subroutine` **7**（共用子程式片段，實機不會單獨出現）。⚠ 分母的算法在 spec 1110 換過：上一版用 offset 順序切頁、又沒走訪 `ON GOTO`，197 頁只佔兩成 | `ecl-text-coverage.md` |
| 正常玩家路徑 | **開場 → 結局**（擊敗提朗瑟克斯的結局選單），拆成 23 個段 subtest | `go test` |
| 全套 gate | `./tools/go.sh test ./...` 全綠 | 本輪實跑 |
| 遊戲入口旗標 | `-segment <id>` 統一直入（`-segment list` 列 25 段）＋ 既有專用旗標 | `internal/segment`／`main.go` |

**函式層已經全部看過一遍，語意層與內容量才是剩下的工作。**
兩層不可互換：函式已盤點不代表系統語意閉合。

### 剩下的工作是什麼形狀

架構分離已經到位：`internal/game` 的 15 個非測試檔全是機制（狀態機、戰鬥、建角、
訓練、商店、神殿、時間、法術），**沒有任何區域或劇情專屬檔**；劇情、座標、選項、
文字都在 game pack JSON 裡，符合 `AGENTS.md` §2 的界線。翻譯管線同樣到位——
game pack 內 `en` 與 `zh-TW` 各 724 條，一一對齊沒有漏。

所以剩下的**不是重構程式碼，是產出資料**：把原作 25 個 block 的事件逐條寫成
game pack JSON。**文字這一層已經接完**（spec 1110）：控制流可達的 1,022 頁全部有規則命中，
`text_rules` 由 464 條成長到 1,009 條。目前的區域分佈：

| 區域前綴 | `text_rules` 條數 |
|---|---:|
| `myth-drannor` | 179 |
| `tilverton` | 158 |
| `zhentil` | 128 |
| `pit` | 99 |
| `dexam` | 75 |
| `yulash` | 60 |
| `wizard-tower` | 56 |
| `world` | 51 |
| `lava-tube` | 48 |
| `fire-knife` | 32 |
| `dark-elf-caves` | 32 |
| `hap` | 26 |
| `journal-trigger` | 15 |
| `hillsfar` | 10 |
| `teshwave` | 8 |
| `ashabenford` | 7 |
| `essembra` | 6 |
| 其餘 7 個前綴 | 各 1–5，合計 19 |

⚠ **條數不再是缺口的指標**，`unmatched` 才是，而它已經是 0。條數少的區域
（艾森布拉 6、阿沙本福德 7、希爾斯法 10）在報告裡沒有未接上的頁。
⚠ 但**世界地圖那兩個 block 有保留**：`ECL1.DAX/0x51` 的走訪會碰到狀態數上限而提早停
（spec 1110 §五），那個 block 可能有走不到的頁沒進分母，而艾森布拉／希爾斯法就在那裡。

**接下來的缺口在文字以外**：事件副作用（旗標、解鎖、戰鬥編成、座標條件）有沒有接對，
那是 `RE-04` 的逐格盤點與 `VER-04`／`VER-05` 的分段驗收要回答的，`ecl-text-coverage` 驗不到。

---

## A. 逆向工程（R1–R3）

矩陣標成 `待逆向` 的列全部在這裡。**全部以 DOS 為主檔，PC-98 符號表當語意骨幹。**

| ID | 項目 | 現況 | 要做什麼 | 產物 |
|---|---|---|---|---|
| `RE-01` | **ECL 有序副作用與 exactly-once**（全域 P0-RE-1） | ✅ **主要成果已取得**（spec 1104）：ordered effect record 已產出（`ecl-opcode-effect-phases.md`，46 列）；32 個候選 31 個 `covered/exact`、1 個 `partial`。三條通則——PC 一律在效果之前推進、畫面提交點只有 `CALL 2E10h`、`20h NEWECL` 是終止指令 | 讀完剩餘 21 支 `unknown` handler；補動態 branch 與原版／remake trace diff。⚠ `resume_only` 這一類是空的：等本輪跑完才發生的效果都在 lifecycle 驅動器裡，不在 opcode handler 內 | `ecl-opcode-effect-phases` 已產出 |
| `RE-14` | **ECL↔引擎共用格子清冊** | ✅ **已完成**（spec 1097、`docs/audit/ecl-shared-cells.md`）：81 個 ECL 變數位址中 **24 個是共用格子**，57 個 ECL 私有 | 逐格對上剩餘語意：`7ED2h`／`7ED3h`／`7ED5h` 的引擎側存取點（`overlay-07:01FC`／`overlay-20:0C9C`／`overlay-14:078E`）尚未逐條讀 | `ecl-shared-cells.md` 已產出 |
| `RE-15` | **ECL 變數讀取端** | ✅ **已完成**（spec 1098）：分區表、`×2`、區 3 的 byte 寬度三項完全對稱 | — | spec 1098 |
| `RE-17` | **角色欄位投影** | ✅ 機制已解：讀取側投影表 spec 624／1040 早有；spec 1098 補上**寫入側也有一張表且與讀取側不同** ⇒ 12 個位址是唯讀投影（寫了讀不到） | 寫入側表未逐條人工確認的部分（含 `7C80h`／`7C81h`） | 補進 spec 1098 |
| `RE-16` | **`7ECAh` 還原方式的時機對應** | spec 1097 §三：原作跑完還原成 `and 1`，remake 一律寫 0 | 先確認 remake 的 `SearchLocation` 對應原作哪一段流程，再決定是否照抄 `and 1` | 補進 spec 1097 |
| `RE-02` | **全遊戲事件清冊**（P0-RE-2） | 靜態層完成（6 DAX／25 block／125 entry／1,355 instruction） | 補動態 branch、座標／terrain、條件旗標、consumer、resume、R1–R5 回填 | `ecl-event-catalog` 動態層 |
| `RE-03` | **External `CALL` 登記表** | ✅ **靜態層已完成**（spec 1104 §七）：`2Dh` 是七路 switch（operand 值減 `7FFFh`），23 個靜態可達 CALL 只用到 `2E10h`（12 次）與 `6803h`（11 次），另五路 corpus 從未使用；未列入 switch 的目標**靜默 no-op** | 兩個實際使用目標的 consumer 逐條驗證與 remake adapter；`6803h` 的 `722Ah` 指標陣列版面未取 | `external-call-registry` |
| `RE-04` | **劇情與全地圖事件** | 大量 fixture，缺逐格覆蓋 | 每區逐格／逐事件的 producer、條件、分支、副作用、重訪 | `area-event-coverage` |
| `RE-05` | **DOS save bundle（只讀）** | **角色記錄逐位元組有台帳**（spec 1115）：422 bytes 中 `decoded` 294／`documented` 100／`unknown` 28，`decoded` 用位元組突變量測驗證，雙向對帳跑在 `go test` 裡 ；`.SWG`（63 bytes）與 `.FX`（9 bytes）也蓋滿，`unknown` 都是 0 | 剩 PC-98 `CHARREC`（`1A7h`，多一 byte）。⚠ 決策四：**不需要寫回原版、不需要 round-trip gate**，只要讀得進來 | `dos-save-bundle-schema` |
| `RE-06` | **戰鬥 scheduler／initiative** | 部分 typed core | round/segment、held/delayed、surprise、flee/guard/quick、死亡與戰後 handoff | `combat-turn-lifecycle` |
| `RE-07` | **敵方 AI／怪物特殊能力** | **COMPTACT（overlay-09）38 個函式裡 16 個已解讀，大的全部在內**：AI 一回合（830）、攻擊／移動主迴圈（838）、試方向（837）、走一步（839）、用道具（835）、選法術（836）、施法目標閘門（802）、友軍誤傷掃描（777）、自動換裝（1004）、士氣（758）。其餘 22 筆是 IDA 的邊界碎片 | 缺的是**實作**（`ENG-08`）與怪物特殊能力逐種（群體、抗性、免疫、毒素、凝視）| `monster-ai-and-specials-matrix` |
| `RE-08` | **AREA map** | 有資料與局部畫面 | player marker、探索狀態、秘密區、Journal 59 圖、縮放／色盤、save state | `area-map-contract` |
| `RE-09` | **近戰／弓箭／投射物** | 命中核心與基本箭矢 | 武器速度／射程／彈藥／多攻、逐幀 projectile、sound、impact、death、continuation | `physical-attack-matrix` |
| `RE-10` | **跨遊戲角色轉移** | 手冊證明功能存在 | source selector、record conversion、裝備／法術／等級限制、round-trip | `move-party-transfer-contract`（不阻塞單作通關） |
| `RE-11` | **未定義區段判定** | **`WORKLIST.md` 第 559 輪第 5 步尚未執行** | DOS 16,044 bytes／PC-98 20,319 bytes 未被任何函式涵蓋的區段，逐段判定是資料表、對齊填充還是漏掉的程式碼 | 台帳補列 ＋ 判定紀錄 |
| `RE-12` | **資料段表格取出** | ✅ **建角相關九張已取出**（四張建角表 ＋ 四組選單字串 ＋ 陣營過濾表 ＋ 生命骰／體質加值／起始金錢；spec 1099／1100／1101／1102／1103）。原文：四張已取出（spec 1099、`docs/audit/resident-data-tables.md`）：種族屬性上下限 `DS:3F86h`、職業最低要求 `DS:4172h`、起始年齡 `DS:404Ch`、可選職業 `DS:3FF8h`，全部以 AD&D 1e 指紋驗證 | 剩隨機寶物表未取。★ 索引一律用原作 `+74h` 種族編號 1..7（不是 remake 的 `Race` 常數） | `resident-data-tables.md` |
| `RE-13` | **未被引用的舊腳本判定** | 7 支 `scripts/ida/*.idc` 沒有任何文件按檔名引用 | 逐支判定補引用／合併／archived | 稽核紀錄 |

### 已閉合、可直接接進實作的成果（第 560 輪新增）

不必再逆向，直接列入 B／C 區的輸入：

- **spec 1083 ECL 助憶碼表**（65 路 case）——把幾十份既有 ECL 規格釘到確切指令碼。
  ★ `29h` ＝ `ENCOUNTER MENU`（不是 PARLAY，PARLAY 是 `2Ch`）、`2Bh` ＝ `HORIZONTAL MENU`、`1Fh` 沒有分支。
- **spec 1087 `PROGRAM 8`** ——通關開關的完整序列（結局 → 三個旗標 → 全隊復活補滿 → 主選單 → 存檔 → 結束）。
- **spec 1072／1075／1076 存檔版面** ——16 塊固定區、DOS 13,149 bytes、存檔槽 `'A'`..`'J'`。
- **spec 1084 訓練所** ——AD&D 1e 亞人等級上限整表；比的是**目前值**，所以裝備能突破上限。
- **spec 1093／1094 建角** ——17 種職業組合、起始經驗值 25,000 平分、種族／職業屬性表索引算式、
  收尾把目前值抄成基準值。
- **spec 1079 屬性重算** ——`+10h+i*2` 基準值／`+11h+i*2` 目前值；力量 18/xx 的職業限制。
- **spec 1100／1102 建角選單** ——四組選項字串、陣營九宮格順序、種族選單跳過半獸人、
  陣營依職業組合過濾（`DS:41D8h`）。
- **spec 1101 建角 HP** ——八槽各擲一次生命骰、逐槽體質加值、除以職業數；
  `+12Ch` 基礎最大 HP 與 `+78h` 最大 HP 分開存。
- **spec 1103 `ROLLDICE`** ——參數是（骰數, 面數），`ROLLDICE(n,s)` 就是標準 `n d s`；
  建角擲點 `3d6+1` 六次取最大，七個種族的屬性調整全部取出。
- **spec 1091 半形轉全形** ——繁中化的關鍵，見 C 區。

---

## B. 引擎與資料層（R4）

### B-1 內容產出（**主要工作量**）

架構界線已符合 `AGENTS.md` §2，機制與內容確實分離（見上節）。這一區的工作
是把 `RE-04` 盤點出的事件寫成 game pack 資料，不是搬程式碼。

| ID | 項目 | 要做什麼 |
|---|---|---|
| `ENG-01` | **事件內容補齊** | ✅ **全部接上**：`cmd/ecl-text-coverage` 依**控制流走訪**逐頁列出原作文字與 game pack 的接線（spec 1110）。控制流可達的 **1022 頁**中 `matched` 999、**`unmatched` 0**；剩下 `variable-insert` 16（頁裡印的是執行期的值，靜態驗不到）與 `subroutine` 7（共用子程式片段，實機一定被併進呼叫端那一頁）。text_rules 由 464 條成長到 1009 條。⚠ 上一版的分母 197 頁只佔兩成——`25h ON GOTO` 沒被走訪、`15h`／`2Bh`／`25h`／`26h` 的變長指令長度沒算、頁的歸屬用 offset 順序，三者疊起來讓 `unmatched=0` 什麼都不代表。**寫規則前先看報告的 Limitations**，尤其三種會讓實機變差的寫法（片段跨 run、片段短到攔截別人、把執行期的值換成固定句），見 spec 1110 §四 |
| `ENG-02` | **game pack 分檔** | ✅ **第一步已完成**（spec 1105）：切成 `gamepack/pack/` 四檔——`00-core`（機制資料）、`10-content`（`text_rules`／`option_rules`／`events`）、`20-locale.en`／`20-locale.zh-TW`。引擎新增 `LoadPackParts`，合併後才驗證，重複一律失敗不做後蓋前。⚠ **依區域再切留給命名收斂之後**：實測 ID 命名空間不一致（83 條 `ecl-option.*`、約 100 個沒有命名空間的扁平 locale 鍵），現在依區域切會夾帶一次大改名 |
| `ENG-13` | **補 `7CE4h` 的角色欄位投影** | ✅ **已完成**：`Character.ECLFlag192`／`DOSPlayerRecord.ECLFlag192`／`PartyMemberContext.ECLFlag192` 三處加欄位，DOS 記錄 `0x192` 讀寫 round-trip，`LOAD CHARACTER` 時投影並套 `and 1` 遮罩；回歸測試 `TestRunSubsetLoadCharacterProjectsFlag192MaskedToLowBit` |
| `ENG-14` | **原版資料表一律走 JSON** | 使用者要求：原版取出的資料表全部轉成 game pack JSON，Go 只留機制，不再 hardcode。已完成起始年齡（`gamepack/rules/character-tables.json`）。⚠ 仍 hardcode 在 `internal/party/character.go` 的：`WithAgeEffects` 的年齡分期門檻與效果表、`raceAllowsClass`；`internal/party/dos_spell_record.go` 的 `parseDOSRace`／`parseDOSClass` 對應表也應改由 JSON 提供 |
| `ENG-03` | **stable ID 與 schema 版本** | ✅ **規範已立**（spec 1105 §四）：`<群組>.<地點或子系統>.<東西>`，全小寫、`.` 分隔、群組內用 `-`、不用底線。既有 **155 個**不合規範的 ID 進 `docs/audit/gamepack-id-baseline.json`，新增不合規範會紅（`TestGamePackIDNamingBaselineIsExact`）。剩：逐條收斂那 155 個（126 個 locale 鍵 ＋ 29 條 `text_rules`），收斂完才能依區域分檔 |
| `ENG-12` | **engine／遊戲切分複查** | `internal/` 29 個套件對 `golden-box-remake-engine` 69 個 `.go` 的比例偏低；逐套件判定哪些機制該上移。⚠ 這是既有機制的整理，不阻塞內容產出，排在後面 |

### B-2 規則系統

| ID | 項目 | 依賴 | 要做什麼 |
|---|---|---|---|
| `ENG-04` | AD&D 規則完整表 | `RE-12` | 全職業／種族限制、能力修正、升級、休息、時間、負重、狀態、item special consumer |
| `ENG-05` | 建角完整流程 | spec 1093／1094／1099 ✅ | ✅ **起始年齡已接上 JSON**（`gamepack/rules/character-tables.json`，由 `cmd/dseg-export` 產生；修正了漏掉德魯伊欄造成的錯位）。✅ 性別欄位、✅ 屬性夾值（兩次夾值）、✅ **力量 18/xx 的獨立階**、✅ **四段選單機制**（`BeginGuidedCreation`／`SelectGuidedOption`／`RollGuidedAbilities`，職業選項由種族查表）、✅ **起始經驗值 25,000 平分**（`class_combinations` 進 JSON）。✅ **UI 已接**（`-guided-creation` 旗標、建角畫面按 `G` 進入，上下移動 ＋ Enter 選定；headless 截圖確認選單與中文正常）、✅ **locale 詞條齊全**（16 個新詞條，另有測試直接驗證正式檔案）、✅ **基準值／目前值成對保存**（`Abilities.Current`／`SyncBaseFromCurrent`）。✅ **陣營九宮格升為 exact**（spec 1100 取出原作字串表，順序確認；種族索引 0 是 `Monster`）。
多職組合：✅ 職業槽等級、✅ 起始經驗值平分、✅ 主職業對應（三個來源交叉驗證）、
✅ **多職 HP**（spec 1101：八個職業槽各擲一次生命骰、逐槽體質加值，再除以職業數；
`+12Ch` 基礎最大 HP 與 `+78h` 最大 HP 分開存；戰士系額外體質加值看的是**職業組合編號**
所以多職拿不到）、✅ **陣營依職業過濾**（spec 1102：聖騎士只有守序善、遊俠只有三個善）、
✅ **種族選單跳過半獸人**（spec 1102：原作顯示迴圈沒有分支收它，編號仍維持 1..7 不連號）。
✅ **名字輸入**（支援中文，退格按字元；空名字被擋下，同原作）、
✅ **`Save <名字>?` 收尾**（只比對 `N`，其餘存檔；答 `N` 角色不留）、
✅ **加進隊伍名冊**、✅ **兩個順序修正**（HP 在重擲迴圈裡、基準值複製在名字之後）。
名字上限維持 remake 自己的 20 字元（原版是 15），remake 有自己的存檔格式，
只需支援讀取原版。
✅ **屬性擲點已定案**（spec 1103：`ROLLDICE` 本體讀完，參數是（骰數, 面數），
每顆 `Random(面數)+1` ⇒ `ROLLDICE(3,6)` 就是標準 3d6；建角是 `3d6+1`、
每個屬性擲六次取最大）、✅ **七個種族的屬性調整全部取出**（逐格對上 AD&D 一版，
含半獸人力量 +1／體質 +1／魅力 −2），資料進 JSON。
**除了「尚未成為預設入口」之外，建角流程已閉合**（使用者已決定兩者並存，
範本流程保留給快速開局）|
| `ENG-06` | 訓練所升級 | spec 1084 ✅ | 亞人等級上限（比目前值）、經驗門檻、HP 保留受傷差額 |
| `ENG-07` | 戰鬥回合生命週期 | `RE-06` | initiative、held/delayed、surprise、flee/guard/quick、死亡與戰後 handoff |
| `ENG-08` | 怪物 AI | `RE-07` ✅（COMPTACT 已解讀）| **移動已接**（spec 1114）：每回合抽行為模式 1..6、模式決定五個候選方向、正向 ×2 斜向 ×3 半格成本、走到射程內才攻擊、20 次上限；方向表逐 byte 對回資料段。剩：目標選擇照原作挑法、AI 用道具（835）、AI 選法術（836）、士氣與恐慌逃走（758／830）、自動換裝（1004）、兩種障礙的豁免效果（837）|
| `ENG-09` | 全法術表 | spec 1111 ✅（資料）／1117 ✅（資料驅動施法）| 原作 100 筆全部在 `gamepack/rules/spell-table.json`，逐欄有出處。**實作不再是一支法術一段程式碼**：效果碼（`+0Ah`）、持續時間係數、豁免類別直接讀表，一次接上 11 支效果碼 remake 解讀得了的法術（定身、加速、緩慢、隱形…），宣告數 12 → **23**／79。**兩張表撐起來**：效果修正表（1123，`CHECKFX` 給清單、handler 給數字，含條件守衛）＋ 傷害骰表（1124，分派表 → handler → 兩支擲骰入口）。分母修準為 **73**（79 減去 6 支職業模型不支援），宣告數 12 → **50**。剩 23 支，每支的 blocker 在 `docs/audit/combat-spell-coverage-ledger.md` 逐支列出；visual／sound 全部未接 |
| `ENG-10` | 存檔完整實作 | `RE-05`、spec 1072／1076／1115／1118 ✅ | **兩道完整性的閘已建**：`Fighter` 每個匯出欄位都要能存進快照再讀回（反射，新增欄位當場變紅，當場抓到 `StatusPartyFled` 沒進 `RestoreBattle` 上限檢查）；整局存檔另有「存 → State → 存」的往返閘，守住 `SavePartyFile` 那 20 個位置參數（漏傳或參數對調編譯器都不會吭聲）。豁免只給快照指標，純量欄位不得豁免。⚠ 決策四：不做原版 round-trip |
| `ENG-11` | 通關路徑 | spec 1087 ✅ | `PROGRAM 8` 的完整序列接上結局與存檔 |

⚠ **原作瑕疵要決定照抄或修正**，並在規格裡記錄決定：
`SURPRISE` 結果碼 3 永遠寫不出去（1087）、隨機寶物表 `56h..5Ch` 與 `5Bh..62h` 重疊（1087）、
職業組合 `10h`（法師/盜賊）只有兩職卻拿三職的 8,333 經驗（1093）、
空物品鏈遺失保留的寶石（1035）、PC-98 `T` 熱鍵留空格（1032）。

---

## C. 繁體中文化

目前 `assets/locale/zh-TW.json` 872 條；game pack 內建雙語 `en` 1,269 ＝ `zh-TW` 1,269，key 完全對齊。

**擋路的不是雙位元組處理（那在 remake 不存在），是 `CHT-04` 的熱鍵欄位；`CHT-09` 的譯名表已於 spec 1107 立完。**

| ID | 項目 | 要做什麼 |
|---|---|---|
| `CHT-01` | ~~雙位元組字串處理~~ | ✅ **remake 不受此限**：`internal/etenfont` 以 rune 為單位，只在查字模時才轉 Big5，原作那條 byte 管線沒有被沿用。spec 1091 記的三個必改點（Big5 首位元組判定、半形カタカナ段衝突、`0FFh` 緩衝溢位）都只適用原版程式碼 |
| `CHT-02` | ~~名字編輯以字元為單位~~ | ✅ **已滿足**：`BackspaceGuidedName`／`BackspaceCreationName` 用 `[]rune` 退格，`BackspaceECLString` 用 `utf8.DecodeLastRuneInString`，長度上限用 `utf8.RuneCountInString`。原作逐 byte 搬移的作法沒有被照抄 |
| `CHT-10` | **原版位元組的解碼邊界** | ✅ **機制已建立**：`internal/origtext` 以 Big5 解原版位元組（ASCII 相容，英文原版逐位元組不變）；`party.ParseOriginalDOSPlayerRecord` 是角色記錄的匯入入口，`monster.ParseRecord` 直接套用（MON*CHA 唯讀）。<br>⚠ **界線是「位元組從哪來」，不是「版面長什麼樣」**：角色記錄與物品記錄的版面同時被原版與 remake 自己的存檔使用，而 remake 寫入端寫的是 UTF-8——在共用解析函式裡一律當 Big5 會把 remake 自己的存檔讀壞（兩次都被測試擋下）。<br>✅ **`ENG-10` 的匯入入口已接**（spec 1121）：`LoadOriginalSAVGAMSlot` → `ParseOriginalDOSPlayerFiles` → `ParseOriginalItems`，CLI 旗標 `-savgam-import`。剩：原版 ITEM DAX 等其他資料檔的載入層要各自做同樣的分流 |
| `CHT-03` | **名字長度上限** | DOS 名字欄位長度上限要重新確認（PC-98 是 `0Fh` ＝ 15 bytes ＝ 全形 7 字）；PC-98 另有 `0723:05BDh` 名字驗證，判定規則未取出，很可能擋掉 Big5 |
| `CHT-04` | **指令列的按鍵與標籤沒有對應** | 缺口不在 `option_rule`——ECL 選單是**按索引**選的，沒有字母熱鍵（spec 1105 §六）。真正的問題在指令列：畫面用 locale 字串顯示中文（`combat_menu_main` ＝「移動　查看　瞄準　施法…」），按鍵卻是散在前端的英文首字母常數（`ebiten.KeyM`／`KeyL`／`KeyC`／`KeyQ`），**翻譯後玩家看不出要按哪個鍵**，而且標籤與繫結沒有任何地方對得起來。處置方向：把「指令 ID → 按鍵 → `message_id`」做成一張表，繪製端與輸入端讀同一張。⚠ 在有消費端之前不要先往引擎 schema 加欄位 |
| `CHT-05` | **固定欄寬排版** | PC-98 是 40 bytes 固定欄位、機能鍵列每格 7 欄（spec 1077／1092）；DOS 是靠熱鍵字母。兩者中文化做法不同，繁中版採 DOS 的做法但要處理全形寬度 |
| `CHT-06` | **翻譯隨內容同步** | 目前 game pack 的 `en`／`zh-TW` **條數一致、一一對齊、沒有漏譯**（`TestDefaultPackMergesAllCommittedParts` 檢查對齊，不釘死條數——釘死的快照擋不住錯誤卻每批必紅）。翻譯不是獨立階段，是 `ENG-01` 每寫一條事件就同時寫兩個語系。把「兩語系條數與 key 完全一致」設成 CI gate（`cmd/locale-drift-audit` 已有基礎），漏一條就紅 |
| `CHT-09` | **統一譯名表** | ✅ **已完成**（spec 1107）：表在 `docs/knowledge/coab-glossary.md`（68 詞條，含 `combatant_name_rules` 自動匯入的怪物名），閘在 `internal/glossary` ＋ `cmd/glossary-audit`，fail-closed 掃三份繁中目錄。建表時量到**九組同名異譯**並全數修正（Bane 貝恩／班恩、Zhentil Keep 散提爾堡／散塔林堡、Flamed One 三種寫法…），其中五組在手札長文裡。⚠ 閘擋不住「誤用的寫法同時是別的詞條的正確譯名」那一類（`Haptooth` 誤寫成哈普），只能逐句對照原文 |
| `CHT-07` | **Journal 整合進遊戲** | ✅ **已完成**（spec 1108／1109）：遊戲內閱讀器可用（`ModeJournal`，依字型 advance 分頁，spec 556）；全 corpus 窮舉普查出 **59 則裡有 44 則會被引用**，這 44 則全部收錄且全部有 producer，另加第 1 則（開場導言，掛在甦醒那一幕）＝ **45 則玩家讀得到**，只在劇情命中時解鎖，`TestEveryJournalEntryHasAProducer` 擋住死條目。手冊本文明寫「some tales are false」，所以逐則解鎖是必要條件而非保守作法。**六張地圖與插圖也進遊戲了**：手札畫面按 `I` 彈窗，`Z` 切原尺寸、方向鍵平移（spec 1109）；條目本文那句「原始圖像保存於 Adventurer's Journal 第 N 頁」已刪。剩 14 則 corpus 沒有任何引用（第 8／39 則只有插圖、九則講別部作品的地點、三則查無引用），依手冊那句話判斷應為誘餌，不補 |
| `CHT-08` | **字型** | 倚天點陣字（16×15 粗體）已接通；缺全形標點 fallback 與 24×24 的取捨判定 |

---

## D. 表現層與音訊

| ID | 項目 | 現況 | 要做什麼 |
|---|---|---|---|
| `UI-01` | 畫面狀態逐張 contract | 640×480、石框、部分舞台完成；**第一人稱已有逐格數字**——提爾佛頓五格 × 四朝向與原版 19／20 完全相同（spec 1134） | adventure／combat／map／dialog／roster／shop／spell／save／ending 每個狀態一份對照。第一人稱剩其餘 17 張 `first_person` 地圖（天空色仍是宣告值，多半 0 ＝ 黑天花板）|
| `UI-02` | 戰鬥演出 | SPRIT／CPIC 解碼與部分 timeline | 近戰、弓箭、12 個法術、死亡、area persistent effect 的逐幀 oracle。`AGENTS.md` §8 要求分開追蹤 caster windup／travel／impact／damage text／saving throw／death／area effect／sound cue／handoff，**不得因共用 projectile 素材而省略後續效果** |
| `UI-03` | AREA map 畫面 | 資料有、畫面局部 | 依 `RE-08` 實作 |
| `AUD-01` | DOS 音效 | 9 個 WAV ＋ 部分 selector | 全 caller、缺的 selector、場景、優先權、重疊、停止、存檔與原版時序 |
| `AUD-02` | PC-98 音樂／音效 | YM2203/S98/driver 研究深入 | VFD 缺 sector 的合法來源、真實 SFX caller、save/resume、reload phase、gain。**這一項服務於 PC-98 參考價值，不阻塞 DOS 主線交付** |

---

## E. 工程衛生與驗收

### E-1 全套 gate（2026-08-15 更新）

三類失敗已全部處理，`./tools/go.sh test ./...` 現在可以當全套 gate。
前提是先跑一次 `tools/build-go-image.sh` 建好 `coab-go-ebiten:1.24`。

| ID | 失敗 | 原因 | 處置 |
|---|---|---|---|
| `VER-01` | `internal/sound` 與 `cmd/azure-bonds-game` **建置失敗** | ✅ **已解決**：新增 `tools/Dockerfile.go-ebiten`（`golang:1.24` ＋ ALSA／X11 開發標頭 ＋ Xvfb）與 `tools/build-go-image.sh`；`tools/go.sh` 自動採用 `coab-go-ebiten:1.24` 並用 `with-xvfb` 在 Xvfb `:99` 下執行。⚠ 不用 `xvfb-run`——它在 `-u <uid>:<gid>` 且該 uid 不在 `/etc/passwd` 時會**無限等待** | — |
| `VER-02` | `internal/sourceaudit` Han literal baseline drift | ✅ **已解決**：29 筆新增與既有 9 筆性質相同（全是開發工具的中文訊息，不是玩家可見 UI），用既有的 `cmd/coab-audit -write-baseline` 重生成 38 筆。⚠ `category()` 目前把所有非 `cmd/azure-bonds-game/` 的都歸成 `runtime_ui_debt`，開發工具訊息與遊戲執行時訊息混在一起，這個分類值得後續調整 | — |
| `VER-03` | `scripts/`／`workplace/` 建置失敗 | ✅ **已解決**：`scripts/*.go` 五支各自移進同名子目錄成為獨立套件；`workplace/` 的兩支一次性 probe 加 `//go:build ignore`（保留檔案，單檔 `go run` 不受影響） | — |

✅ `go test ./...` 現在可以當全套 gate（前提是先跑 `tools/build-go-image.sh`
建好 `coab-go-ebiten:1.24`）。`AGENTS.md` §8 已同步更新。

### E-2 分段驗收清單（rulebook 65，使用者 2026-08-16 修正）

`cmd/azure-bonds-game/main.go` 有 55 個旗標，其中 **30 個是直接進入某一段的入口**。

修正後的口徑：**分階段驗收算數**——每一段用 debug 進入點直入並各自對 reference
驗證無誤，即算該段完成；全部段落通過即算跑完，不必為了宣稱完成而跑一次連續全程。
所以下表不再是「待移除清單」，而是**覆蓋率清單**：每個旗標對應一段，
要記錄那一段驗過沒有、用什麼 oracle 驗的。

| 類別 | 旗標 | 分段驗收時的角色 |
|---|---|---|
| 場景直入（`-segment <id>` ＋ 25 個既有旗標） | `-encounter` `-character-creation` `-tilverton-dungeon` `-inn` `-filani` `-weapon-shop` `-temple` `-training` `-tavern` `-high-priest` `-carriage` `-guildmaster` `-sewers` `-lava-tube` `-wizard-tower` `-wizard-tower-battle` `-wizard-tower-parlay` `-wizard-tower-exit` `-burial-red-web` `-burial-red-web-battle` `-burial-grave-battle` `-burial-daemir` `-inner-ritual` `-inner-final-battle` `-world-map` | **各自是一段的進入點**；該段要走到正常結束狀態，不是只驗一個畫面。統一入口是 `-segment <id>`（`-segment list` 列 25 段，註冊表在 `internal/segment`），既有旗標保留但它們做的是**段內**檢查點 |
| 視覺 oracle（5） | `-dungeon-x` `-dungeon-y` `-area-map` `-combat-terrain` `-combat-visual-demo` | deterministic 截圖比對；不單獨構成一段 |
| 正常入口 | `-opening` | ⚠ 它跳過建角 ⇒ 建角是**另外一段**，用 `-guided-creation` 驗 |
| 資產／設定 | `-font` `-eten-font` `-locale` `-image` `-sound-dir` `-savgam-dir` 等 | 允許 |

⚠ **接縫本身也是一段。** debug 旗標注入的是合成起始狀態，未必等於上一段真的跑出來的
結束狀態；兩端都綠不等於接縫通過。狀態交接（存檔、旗標、隊伍、pending ECL／combat
transaction）要自己列一段驗——這是 rulebook 65 來源案例的失敗點，放寬驗收成本時
不能一併放掉。

| ID | 項目 |
|---|---|
| `VER-04` | 建立**分段驗收矩陣**：每一段列出進入點旗標、正常結束狀態、用什麼 oracle 驗、驗過沒有；另外把段與段的接縫列成獨立條目（建角→開場、提爾佛頓→世界地圖、各戰後 continuation、存檔→重開） |
| `VER-05` | 對 reference 實測：DOSBox 原版與 remake 逐段對照，標明 `exact`／`reconstructed`／未完成 |
| `VER-06` | 每個遠程／法術能力記錄影片 URL、平台、絕對時間碼、逐幀順序與對應 sprite block（`AGENTS.md` §8） |
| `VER-07` | 存檔**匯入**測試：原版 `SAVGAM?.DAT` 讀進 remake 後隊伍、角色、旗標正確（spec 1072／1076 的 16 塊版面）。⚠ 決策四：不驗「寫回原版可讀」。DOS 13,149 與 PC-98 13,214 bytes 仍不能互換 |

---

## 每一輪的執行流程

沿用 `CLAUDE.md` 的核心工作法，**每一輪實作前一定先寫 spec**：

```
反組譯／實機證據 → docs/spec/NNNN-*.md → 標 READY → 實作 → 測試與玩家路徑驗證
                                                    → 更新 Markdown／截圖 → commit
```

- spec 先寫，而且要標明證據等級（`exact`／`strong inference`）與**明確不宣稱**的邊界。
- IDA／decompiler 輸出本身不是證明，要有原始 bytes、runtime capture 或另一權威來源交叉驗證。
- 實作只能落在 spec 已宣稱的範圍內；spec 沒寫的地方不准在程式裡猜一個值填進去。
- 被新證據推翻的斷言直接改寫 spec 正文；容易再犯的部分寫成規則進 `AGENTS.md` §3，
  能加回歸測試就加。只有在來源改不乾淨時才需要另留紀錄。
- 每輪結束更新本清單對應條目與覆蓋矩陣。

## 建議順序

擋路的在前面，能平行的標出來。

1. ✅ **`VER-01`／`VER-02`／`VER-03` 已完成**——`./tools/go.sh test ./...`
   現在 31 個套件全綠，可以當全套 gate（先跑 `tools/build-go-image.sh`）。
2. **`RE-14` 共用格子清冊 ✅ 已完成**——81 個 ECL 變數位址中 24 個是共用格子。
   風險面積從「69 個位址未校準」收斂成「24 個要逐格對上、57 個自洽即可」。
   剩下 `RE-15`（讀取端）與 `RE-16`（`7ECAh` 時機）兩個小尾巴，不擋路。
3. ✅ **`RE-01` ECL 有序副作用的擋路部分已解除**（spec 1104）——`ENG-01` 的事件資料形狀現在有依據了：
   效果一律在 PC 推進之後發生（停下再續跑不會重播），畫面提交點只有 `CALL 2E10h`，
   `NEWECL` 終止本次執行。剩下 21 支未讀 handler 不擋 `ENG-01`，可與內容產出平行。
4. **`RE-12` 資料段表格取出**——多份規格卡在同一個原因，取出後 `ENG-04`／`ENG-05`／`ENG-06` 一起解鎖。
   可與第 2 項平行。
5. **`CHT-04`（熱鍵欄位）**——剩下的地基：熱鍵是 schema 缺口，必須在分檔定案前補。`CHT-09`（譯名表）已於 spec 1107 完成，大量產出的前提解除。`CHT-01`／`CHT-02` 已確認 remake 不受原作 byte 管線影響。
6. **`ENG-02`／`ENG-03`**——game pack 分檔與 stable ID 定案。必須在大量產出內容之前，否則後期分檔會動到每一條資料。
7. ✅ **`ENG-01` 的文字層已接完**（spec 1110）——控制流可達的 1,022 頁 `unmatched` 為 0，
   `CHT-06` 隨之同步（兩語系各 1,269 條、key 對齊）。**剩下的是 `RE-04`**：
   事件的**副作用**（旗標、解鎖、戰鬥編成、座標與重訪條件）逐格盤點並接上，
   以及 `ECL1.DAX/0x51` 走訪截斷的那一塊要換方法補齊。文字已經不是這一項的瓶頸。
8. **`RE-06`／`RE-07` ＋ `ENG-07`／`ENG-08`**——戰鬥系統。可與第 7 項平行（不同人／不同輪）。
9. **`RE-05` ＋ `ENG-10` ＋ `VER-07`**——存檔。
10. **`UI-01`／`UI-02`／`AUD-01`**——表現層，需要原版 runtime 當 oracle。
    ⚠ oracle 的取法已經不是瓶頸了：`tools/dos-oracle-session.sh` 容器常駐、
    逐步送鍵、每一步讀回畫面文字再決定下一鍵（spec 1134），一步約一秒。
    **不要再寫「一整串定時按鍵」的擷取腳本**——載入時間會漂，漂掉的那一鍵
    會讓後面每一步錯位，而錯位的鍵可能剛好按到 `EXIT TO DOS`。
11. **`RE-11`（未定義區段判定）／`RE-13`／`ENG-12`**——盤點與整理的收尾，不阻塞玩家路徑。
12. **`VER-04`／`VER-05`／`VER-06`**——最終驗收：分段驗收矩陣（含接縫）＋ 對 DOSBox 原版實測。
13. **`RE-10`／`AUD-02`**——不阻塞 DOS 單作通關，最後補。

## 維護

- 每個項目完成時更新 [`coab-re-coverage-matrix.md`](coab-re-coverage-matrix.md) 對應列的層級與判定，
  並用 [`re-closure-record-template.md`](re-closure-record-template.md) 留紀錄。
- 被新證據推翻的斷言直接改寫正文；判準與處置見 `CONTEXT.md` 的「已被推翻的斷言」一節。
- 本清單不使用百分比；狀態只由 R1–R5 的逐層證據推進。
