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

## 現況量測（2026-08-16 實測）

數字一律現場量，不沿用上一輪的文件行。

| 指標 | 數字 | 來源 |
|---|---|---|
| 函式覆蓋台帳 | 2,922 個函式，**已解讀 2,175／不阻塞 162／邊界碎片 585／待解讀 0** | `re-function-ledger.json` |
| ├ DOS | 1,387：已解讀 1,016 ／ 不阻塞 133 ／ 邊界碎片 238 | 同上 |
| ├ PC-98 | 1,535：已解讀 1,159 ／ 不阻塞 29 ／ 邊界碎片 347 | 同上 |
| └ 證據等級 | `exact` 1,955 ／ `strong inference` 223 | 同上 |
| 規格文件 | **1,084** 份 `docs/spec/*.md` | `ls` |
| remake 程式（不含巢狀 repo） | **230 個 `.go`／77,123 行** | `find`／`wc` |
| ├ `internal/game` 機制碼 | **16 檔／13,908 行**（非測試） | `wc` |
| └ 檔名 | `state` `combat_state` `creation` `creation_guided` `training` `shop` `temple` `time` `spells` … **沒有區域／劇情專屬檔** | `ls` |
| 共用 engine | 獨立 repo，69 個 `.go` | `golden-box-remake-engine/` |
| 內容資料 | `gamepack/events/pit-of-moander.json` **261 KB 單一 game pack** | `wc -c` |
| ├ 事件 | 4 個 `events`、113 條 `option_rules`、**369 條 `text_rules`** | `json` |
| └ 語系 | `en` **607** 條、`zh-TW` **607** 條，**一一對齊、沒有漏譯** | `json` |
| UI 詞條 | `assets/locale/zh-TW.json` **870** 條 | `json` |
| 建角規則表 | `gamepack/rules/character-tables.json`：7 種族、17 職業組合、8 職業槽、擲點與體質加值 | `json` |
| 原作事件總量 | 6 DAX／25 block／125 lifecycle entry／**1,355 個靜態可達 instruction** | `cmd/ecl-event-catalog` |
| ECL 副作用候選 | 33 個中 **4 個已審查**、29 個未審查 | `ecl-ordered-effect-reviews.json` |
| 正常玩家路徑 | 走到**眼魔洞穴東門 → 散提爾堡邊緣** | `go test` |
| 全套 gate | `./tools/go.sh test ./...` 全綠 | 本輪實跑 |
| 遊戲入口旗標 | **58** 個，其中 30 個是分段驗收的直入點 | `main.go` |

**函式層已經全部看過一遍，語意層與內容量才是剩下的工作。**
兩層不可互換：函式已盤點不代表系統語意閉合。

### 剩下的工作是什麼形狀

架構分離已經到位：`internal/game` 的 15 個非測試檔全是機制（狀態機、戰鬥、建角、
訓練、商店、神殿、時間、法術），**沒有任何區域或劇情專屬檔**；劇情、座標、選項、
文字都在 game pack JSON 裡，符合 `AGENTS.md` §2 的界線。翻譯管線同樣到位——
game pack 內 `en` 與 `zh-TW` 各 607 條，一一對齊沒有漏。

所以剩下的**不是重構程式碼，是產出資料**：把原作 25 個 block／1,355 個可達
instruction 的事件，逐條寫成 game pack JSON。目前 369 條 `text_rules` 的區域分佈
顯示缺口集中在哪裡：

| 區域前綴 | `text_rules` 條數 | 對照矩陣判定 |
|---|---|---|
| `myth-drannor` | 132 | 待逆向（條數多是因為做過 vertical slice，不是主線已通） |
| `tilverton` | 39 | 局部 |
| `pit` | 33 | 待逆向 |
| `dexam` | 23 | 局部 |
| `zhentil` | 22 | 局部 |
| `wizard-tower` | 20 | 局部 |
| `hap` | 17 | 局部 |
| `journal-trigger` | 15 | 局部 |
| `yulash` | 12 | **待逆向** |
| `fire-knife` | 12 | 局部 |
| `lava-tube` | 10 | 局部 |
| `essembra`（艾森布拉） | **6** | **待逆向** |
| `hillsfar`（希爾斯法） | **5** | **待逆向** |
| `ashabenford` | 5 | 局部 |
| 其餘 8 個前綴 | 各 1–4 | — |

⚠ 條數少的區域（艾森布拉 6、希爾斯法 5、尤拉什 12）就是矩陣標 `待逆向` 的那幾個
——目前只有 fixture 級的內容，沒有正常進出城的完整事件。

⚠ 條數多也不代表閉合：`myth-drannor` 有 132 條卻仍是 `待逆向`，因為那是
Burial Glen 等 vertical slice 加終戰 fixture，缺正常世界入口與三區逐房間串接。

---

## A. 逆向工程（R1–R3）

矩陣標成 `待逆向` 的列全部在這裡。**全部以 DOS 為主檔，PC-98 符號表當語意骨幹。**

| ID | 項目 | 現況 | 要做什麼 | 產物 |
|---|---|---|---|---|
| `RE-01` | **ECL 有序副作用與 exactly-once**（全域 P0-RE-1） | 33 個候選已審 4 個；主迴圈、`24h` handler 與**位址空間映射**已閉合（spec 1095／1096） | 續閉合其餘 29 個候選；建立 opcode 當下的 ordered effect record，標明 immediate／pause-before-commit／deferred／resume-only | `ecl-ordered-effects` READY spec ＋ trace corpus |
| `RE-14` | **ECL↔引擎共用格子清冊** | ✅ **已完成**（spec 1097、`docs/audit/ecl-shared-cells.md`）：81 個 ECL 變數位址中 **24 個是共用格子**，57 個 ECL 私有 | 逐格對上剩餘語意：`7ED2h`／`7ED3h`／`7ED5h` 的引擎側存取點（`overlay-07:01FC`／`overlay-20:0C9C`／`overlay-14:078E`）尚未逐條讀 | `ecl-shared-cells.md` 已產出 |
| `RE-15` | **ECL 變數讀取端** | ✅ **已完成**（spec 1098）：分區表、`×2`、區 3 的 byte 寬度三項完全對稱 | — | spec 1098 |
| `RE-17` | **角色欄位投影** | ✅ 機制已解：讀取側投影表 spec 624／1040 早有；spec 1098 補上**寫入側也有一張表且與讀取側不同** ⇒ 12 個位址是唯讀投影（寫了讀不到） | 寫入側表未逐條人工確認的部分（含 `7C80h`／`7C81h`） | 補進 spec 1098 |
| `RE-16` | **`7ECAh` 還原方式的時機對應** | spec 1097 §三：原作跑完還原成 `and 1`，remake 一律寫 0 | 先確認 remake 的 `SearchLocation` 對應原作哪一段流程，再決定是否照抄 `and 1` | 補進 spec 1097 |
| `RE-02` | **全遊戲事件清冊**（P0-RE-2） | 靜態層完成（6 DAX／25 block／125 entry／1,355 instruction） | 補動態 branch、座標／terrain、條件旗標、consumer、resume、R1–R5 回填 | `ecl-event-catalog` 動態層 |
| `RE-03` | **External `CALL` 登記表** | `2E10／C01E／B200` 只有局部證據 | 23 個靜態可達 CALL 的 caller、operand、state projection、consumer、返回與未知 fallback；只追玩家可見副作用 | `external-call-registry` |
| `RE-04` | **劇情與全地圖事件** | 大量 fixture，缺逐格覆蓋 | 每區逐格／逐事件的 producer、條件、分支、副作用、重訪 | `area-event-coverage` |
| `RE-05` | **DOS save bundle** | raw-preserving parser 與部分 sidecar | 已由 spec 1072／1075／1076 取得 16 塊固定版面與角色檔名表（`148h` ＝ 8×41）；仍缺 `.SAV/.GUY/.FX/.SWG` 全欄位、角色刪除重排、未知 byte consumer、round-trip gate | `dos-save-bundle-schema` |
| `RE-06` | **戰鬥 scheduler／initiative** | 部分 typed core | round/segment、held/delayed、surprise、flee/guard/quick、死亡與戰後 handoff | `combat-turn-lifecycle` |
| `RE-07` | **敵方 AI／怪物特殊能力** | 只有選敵與少量特殊能力 | 移動、目標優先、施法、逃跑、群體、抗性、免疫、毒素、凝視，逐種能力 | `monster-ai-and-specials-matrix` |
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
| `ENG-01` | **事件內容補齊** | 依 `RE-04` 的逐格盤點結果，把每個區域的事件寫成 `text_rules`／`option_rules`／`events`；優先序照矩陣的區域表，`待逆向` 的四個區域（艾森布拉、希爾斯法、尤拉什／摩安德之坑、Myth Drannor 正常入口）先補 |
| `ENG-02` | **game pack 分檔** | 目前是單一 261 KB JSON。區域內容補齊後量會成長數倍，需要先決定分檔方式（依 DAX／依區域）與 schema 邊界，否則後期難以維護與 review |
| `ENG-13` | **補 `7CE4h` 的角色欄位投影** | ✅ **已完成**：`Character.ECLFlag192`／`DOSPlayerRecord.ECLFlag192`／`PartyMemberContext.ECLFlag192` 三處加欄位，DOS 記錄 `0x192` 讀寫 round-trip，`LOAD CHARACTER` 時投影並套 `and 1` 遮罩；回歸測試 `TestRunSubsetLoadCharacterProjectsFlag192MaskedToLowBit` |
| `ENG-14` | **原版資料表一律走 JSON** | 使用者要求：原版取出的資料表全部轉成 game pack JSON，Go 只留機制，不再 hardcode。已完成起始年齡（`gamepack/rules/character-tables.json`）。⚠ 仍 hardcode 在 `internal/party/character.go` 的：`WithAgeEffects` 的年齡分期門檻與效果表、`raceAllowsClass`；`internal/party/dos_spell_record.go` 的 `parseDOSRace`／`parseDOSClass` 對應表也應改由 JSON 提供 |
| `ENG-03` | **stable ID 與 schema 版本** | 事件、物品、法術、怪物、地圖的穩定 ID；schema 變更要有遷移路徑。分檔（`ENG-02`）之前先定案 |
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
| `ENG-08` | 怪物 AI | `RE-07` | 移動、目標選擇、施法、逃跑、抗性與各種特殊能力 |
| `ENG-09` | 全法術表 | 矩陣 `player-spell-matrix` | 目前 12 個 handler；缺 target、range/area、save、duration、stack/dispel |
| `ENG-10` | 存檔完整實作 | `RE-05`、spec 1072／1076 ✅ | 16 塊版面 round-trip；remake save 要能保存所有 pending ECL/combat/audio/UI transaction |
| `ENG-11` | 通關路徑 | spec 1087 ✅ | `PROGRAM 8` 的完整序列接上結局與存檔 |

⚠ **原作瑕疵要決定照抄或修正**，並在規格裡記錄決定：
`SURPRISE` 結果碼 3 永遠寫不出去（1087）、隨機寶物表 `56h..5Ch` 與 `5Bh..62h` 重疊（1087）、
職業組合 `10h`（法師/盜賊）只有兩職卻拿三職的 8,333 經驗（1093）、
空物品鏈遺失保留的寶石（1035）、PC-98 `T` 熱鍵留空格（1032）。

---

## C. 繁體中文化

目前 `zh-TW.json` 847 條。矩陣判定 `待實作`。

| ID | 項目 | 要做什麼 |
|---|---|---|
| `CHT-01` | **雙位元組字串處理**（**擋路項**） | spec 1091 證明原作的字串層以 byte 為單位。三個必改點：① 首位元組判定要換成 Big5 `A1h`..`F9h`；② 半形カタカナ段 `A1h`..`DFh` 與 Big5 首位元組**正面衝突，必須整段拿掉**；③ 輸出緩衝 `0FFh`，半形轉全形會讓長度翻倍而溢位——繁中方案是中文原樣通過、半形英數維持半形 |
| `CHT-02` | **名字編輯以字元為單位** | spec 1086 的名字編輯是逐 byte 搬移，按一次刪除會切掉半個中文字；游標左右移同一個問題。**remake 不能照抄這一段** |
| `CHT-03` | **名字長度上限** | DOS 名字欄位長度上限要重新確認（PC-98 是 `0Fh` ＝ 15 bytes ＝ 全形 7 字）；PC-98 另有 `0723:05BDh` 名字驗證，判定規則未取出，很可能擋掉 Big5 |
| `CHT-04` | **熱鍵與文字綁定** | DOS 大量選單靠**文字裡的字母**當熱鍵（`'Keep Exit'`、`'Modify: '`），中文化後熱鍵必須另外保留（spec 1060） |
| `CHT-05` | **固定欄寬排版** | PC-98 是 40 bytes 固定欄位、機能鍵列每格 7 欄（spec 1077／1092）；DOS 是靠熱鍵字母。兩者中文化做法不同，繁中版採 DOS 的做法但要處理全形寬度 |
| `CHT-06` | **翻譯隨內容同步** | 目前 game pack 的 `en`／`zh-TW` 各 607 條**一一對齊、沒有漏譯**。翻譯不是獨立階段，是 `ENG-01` 每寫一條事件就同時寫兩個語系。把「兩語系條數與 key 完全一致」設成 CI gate（`cmd/locale-drift-audit` 已有基礎），漏一條就紅 |
| `CHT-09` | **統一譯名表** | 內容量會成長數倍，人名／地名／物品／法術譯名要先定表再展開，否則後期回頭校對成本高 |
| `CHT-07` | **Journal 整合進遊戲** | 59 則手冊條目已重建、31 則有資料綁定；缺剩餘 producer、解鎖條件、原圖、重讀與不提前劇透 |
| `CHT-08` | **字型** | 倚天點陣字（16×15 粗體）已接通；缺全形標點 fallback 與 24×24 的取捨判定 |

---

## D. 表現層與音訊

| ID | 項目 | 現況 | 要做什麼 |
|---|---|---|---|
| `UI-01` | 畫面狀態逐張 contract | 640×480、石框、部分舞台完成 | adventure／combat／map／dialog／roster／shop／spell／save／ending 每個狀態一份對照 |
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
| 場景直入（25） | `-encounter` `-character-creation` `-tilverton-dungeon` `-inn` `-filani` `-weapon-shop` `-temple` `-training` `-tavern` `-high-priest` `-carriage` `-guildmaster` `-sewers` `-lava-tube` `-wizard-tower` `-wizard-tower-battle` `-wizard-tower-parlay` `-wizard-tower-exit` `-burial-red-web` `-burial-red-web-battle` `-burial-grave-battle` `-burial-daemir` `-inner-ritual` `-inner-final-battle` `-world-map` | **各自是一段的進入點**；該段要走到正常結束狀態，不是只驗一個畫面 |
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
| `VER-07` | 存檔互通測試：原版 `SAVGAM?.DAT` 讀進 remake、remake 寫回原版可讀（spec 1072／1076 的 16 塊版面）。⚠ DOS 13,149 與 PC-98 13,214 bytes **不能互換** |

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
- 被新證據推翻的斷言直接改寫 spec 正文，推翻紀錄集中到 `CONTEXT.md`。
- 每輪結束更新本清單對應條目與覆蓋矩陣。

## 建議順序

擋路的在前面，能平行的標出來。

1. ✅ **`VER-01`／`VER-02`／`VER-03` 已完成**——`./tools/go.sh test ./...`
   現在 31 個套件全綠，可以當全套 gate（先跑 `tools/build-go-image.sh`）。
2. **`RE-14` 共用格子清冊 ✅ 已完成**——81 個 ECL 變數位址中 24 個是共用格子。
   風險面積從「69 個位址未校準」收斂成「24 個要逐格對上、57 個自洽即可」。
   剩下 `RE-15`（讀取端）與 `RE-16`（`7ECAh` 時機）兩個小尾巴，不擋路。
3. **`RE-01` ECL 有序副作用**——矩陣列為全域 P0，也是「單點測試通過、正常流程仍卡住」的首要嫌疑。
   **它決定 `ENG-01` 寫出來的事件資料是什麼形狀**；在它閉合前大量產出內容，等於押注在可能要重做的 schema 上。
4. **`RE-12` 資料段表格取出**——多份規格卡在同一個原因，取出後 `ENG-04`／`ENG-05`／`ENG-06` 一起解鎖。
   可與第 2 項平行。
5. **`CHT-01`／`CHT-02`／`CHT-09`**——雙位元組處理與譯名表是地基，愈晚改代價愈高。可與 2–4 平行。
6. **`ENG-02`／`ENG-03`**——game pack 分檔與 stable ID 定案。必須在大量產出內容之前，否則後期分檔會動到每一條資料。
7. **`RE-04` ＋ `ENG-01` ＋ `CHT-06`**——**主要工作量**。事件盤點、寫成 game pack、同步兩語系是同一輪的三個動作，不分階段。
   依矩陣區域表的順序推進：艾森布拉 → 希爾斯法 → 尤拉什／摩安德之坑 → Myth Drannor 正常入口 → 各區補完。
8. **`RE-06`／`RE-07` ＋ `ENG-07`／`ENG-08`**——戰鬥系統。可與第 7 項平行（不同人／不同輪）。
9. **`RE-05` ＋ `ENG-10` ＋ `VER-07`**——存檔。
10. **`UI-01`／`UI-02`／`AUD-01`**——表現層，需要原版 runtime 當 oracle。
11. **`RE-11`（未定義區段判定）／`RE-13`／`ENG-12`**——盤點與整理的收尾，不阻塞玩家路徑。
12. **`VER-04`／`VER-05`／`VER-06`**——最終驗收：分段驗收矩陣（含接縫）＋ 對 DOSBox 原版實測。
13. **`RE-10`／`AUD-02`**——不阻塞 DOS 單作通關，最後補。

## 維護

- 每個項目完成時更新 [`coab-re-coverage-matrix.md`](coab-re-coverage-matrix.md) 對應列的層級與判定，
  並用 [`re-closure-record-template.md`](re-closure-record-template.md) 留紀錄。
- 被新證據推翻的斷言直接改寫正文，推翻紀錄集中到 `CONTEXT.md` 的「已被推翻的斷言」。
- 本清單不使用百分比；狀態只由 R1–R5 的逐層證據推進。
