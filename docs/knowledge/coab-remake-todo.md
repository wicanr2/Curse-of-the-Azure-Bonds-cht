# 完成 remake 所需的 TODO 盤點

狀態日期：2026-08-15（第 560 輪）

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

## 現況量測（2026-08-15）

| 指標 | 數字 | 來源 |
|---|---|---|
| 函式覆蓋台帳 | 2,922 個函式，**已解讀 2,175／不阻塞 162／邊界碎片 585／待解讀 0** | `re-function-ledger.json` |
| ├ DOS | 已解讀 1,016 ／ 不阻塞 133 ／ 邊界碎片 238 | 同上 |
| ├ PC-98 | 已解讀 1,159 ／ 不阻塞 29 ／ 邊界碎片 347 | 同上 |
| └ 證據等級 | `exact` 1,952 ／ `strong inference` 223 | 同上 |
| 規格文件 | 1,075 份 `docs/spec/*.md` | `ls` |
| remake 程式 | 218 個 `.go`／73,539 行（`internal/game` 佔 33,500 行） | `find`／`wc` |
| 共用 engine | 獨立 repo，69 個 `.go` | `golden-box-remake-engine/` |
| 內容資料 | `gamepack/events/` **只有 1 個 JSON**（`pit-of-moander.json`） | `ls` |
| 繁中詞條 | `assets/locale/zh-TW.json` **847 條** | `json` |
| 遊戲入口旗標 | 55 個，其中 **30 個是場景直入／視覺 oracle 捷徑** | `cmd/azure-bonds-game/main.go` |

**函式層已經全部看過一遍，語意層與實作層才是剩下的工作。**
兩層不可互換：函式已盤點不代表系統語意閉合。

---

## A. 逆向工程（R1–R3）

矩陣標成 `待逆向` 的列全部在這裡。**全部以 DOS 為主檔，PC-98 符號表當語意骨幹。**

| ID | 項目 | 現況 | 要做什麼 | 產物 |
|---|---|---|---|---|
| `RE-01` | **ECL 有序副作用與 exactly-once**（全域 P0-RE-1） | 33 個跨類型候選只審 3 個 | 從 ECL2 block `0x02` 的 `COMBAT → text` 開始，逐個閉合其餘 30 個；建立 opcode 當下的 ordered effect record，標明 immediate／pause-before-commit／deferred／resume-only | `ecl-ordered-effects` READY spec ＋ trace corpus |
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
| `RE-12` | **資料段表格取出** | 多份規格明寫「表在資料段，未取出」 | 至少取出：種族屬性上下限 `DS:3F86h`（16 bytes/列）、職業最低要求 `DS:4172h`（6 bytes/筆）、起始年齡 `DS:404Ch`、種族→可選職業 `DS:3FF8h`、隨機寶物表 | 補進 spec 1084／1086／1093 |
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
- **spec 1091 半形轉全形** ——繁中化的關鍵，見 C 區。

---

## B. 引擎與資料層（R4）

### B-1 架構界線

`AGENTS.md` §2 的界線是：**劇情／座標／翻譯只能進 CoAB 的 JSON，engine 只放可重用機制，
不得 hardcode 情節。** 目前 `internal/game` 有 33,500 行，而 `gamepack/events/` 只有一個 JSON。

| ID | 項目 | 要做什麼 |
|---|---|---|
| `ENG-01` | **內容外部化** | 把 `internal/game` 裡的劇情、座標、事件、選項搬進 `gamepack/events/*.json`；每搬一塊要有 schema 與回歸測試。這是 R4 的主要缺口，也是 `RE-04` 的落地端 |
| `ENG-02` | **engine／遊戲切分** | 判定 `internal/` 29 個套件裡哪些是可重用機制（該進 `golden-box-remake-engine`）、哪些是 CoAB 專屬；目前 engine repo 只有 69 個 `.go`，比例明顯偏低 |
| `ENG-03` | **stable ID 與 schema 版本** | 事件、物品、法術、怪物、地圖的穩定 ID；schema 變更要有遷移路徑 |

### B-2 規則系統

| ID | 項目 | 依賴 | 要做什麼 |
|---|---|---|---|
| `ENG-04` | AD&D 規則完整表 | `RE-12` | 全職業／種族限制、能力修正、升級、休息、時間、負重、狀態、item special consumer |
| `ENG-05` | 建角完整流程 | spec 1093／1094 ✅ | 種族／性別／職業／陣營四段選單、17 種組合、起始經驗值、屬性擲點與夾值、起始年齡、目前值→基準值收尾 |
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
| `CHT-06` | **全量內容翻譯** | 全 ECL 文字、UI、物品、法術、怪物、場所、Journal、Tavern Tales、攻略；統一譯名表防漂移 |
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

### E-1 目前量到的失敗（2026-08-15 `./tools/go.sh test ./...`）

25 個套件 `ok`，另有三類失敗，**都不是本輪造成的**：

| ID | 失敗 | 原因 | 處置 |
|---|---|---|---|
| `VER-01` | `internal/sound` **建置失敗** | docker `golang:1.24` image 缺 ALSA 開發標頭：`Package alsa was not found in the pkg-config search path`（`ebitengine/oto/v3` 需要） | 建一份帶 `libasound2-dev` 的 build image，或把音訊套件的測試改成 build tag 隔離 |
| `VER-02` | `internal/sourceaudit` `TestRepositoryGoHanLiteralBaselineIsExact` **FAIL** | Han literal baseline drift，29 筆新增全在 `cmd/borland-symbols`、`cmd/ovr-manifest`、`cmd/re-ledger`，分類 `runtime_ui_debt` | 判定這 29 筆是該外部化還是該進 baseline，然後更新 baseline |
| `VER-03` | `scripts/`／`workplace/` 建置失敗 | `main redeclared in this block`（`p0_search_probe.go` vs `p0_geo_route_probe.go`；`extract_dos_*.go` 四支互撞） | `AGENTS.md` §8 已記錄這是既存結構問題。這兩個是暫存目錄，處置方式是各自獨立成子目錄，讓 `go test ./...` 能成為真正的全套 gate |

⚠ 在 `VER-03` 修好之前，`go test ./...` **不能當成全套 gate**，
`AGENTS.md` §8 要求「如實分開報告」。

### E-2 debug 捷徑清單（rulebook 65 [HARD]）

`cmd/azure-bonds-game/main.go` 有 55 個旗標，其中 **30 個是繞過正常玩家路徑的捷徑**。
rulebook 65 的規定是：**debug／後門串起來的「能跑完」不算能跑完**。
最終驗收必須全部關掉，從正常開場一路玩到結局。

| 類別 | 旗標 | 驗收時 |
|---|---|---|
| 場景直入（25） | `-encounter` `-character-creation` `-tilverton-dungeon` `-inn` `-filani` `-weapon-shop` `-temple` `-training` `-tavern` `-high-priest` `-carriage` `-guildmaster` `-sewers` `-lava-tube` `-wizard-tower` `-wizard-tower-battle` `-wizard-tower-parlay` `-wizard-tower-exit` `-burial-red-web` `-burial-red-web-battle` `-burial-grave-battle` `-burial-daemir` `-inner-ritual` `-inner-final-battle` `-world-map` | **全部不得使用** |
| 視覺 oracle（5） | `-dungeon-x` `-dungeon-y` `-area-map` `-combat-terrain` `-combat-visual-demo` | 只用於 deterministic 截圖比對，不得出現在通關驗收 |
| 正常入口 | `-opening` | ⚠ 它「用一個產生好的角色開場」，**跳過建角**；正式驗收要從建角開始 |
| 資產／設定 | `-font` `-eten-font` `-locale` `-image` `-sound-dir` `-savgam-dir` 等 | 允許 |

| ID | 項目 |
|---|---|
| `VER-04` | 建立「零捷徑通關」腳本：`-opening` 也不用，從建角 → 提爾佛頓 → … → Myth Drannor 終戰 → 結局 → 存檔重開 |
| `VER-05` | 對 reference 實測：DOSBox 原版與 remake 逐段對照，標明 `exact`／`reconstructed`／未完成 |
| `VER-06` | 每個遠程／法術能力記錄影片 URL、平台、絕對時間碼、逐幀順序與對應 sprite block（`AGENTS.md` §8） |
| `VER-07` | 存檔互通測試：原版 `SAVGAM?.DAT` 讀進 remake、remake 寫回原版可讀（spec 1072／1076 的 16 塊版面）。⚠ DOS 13,149 與 PC-98 13,214 bytes **不能互換** |

---

## 建議順序

擋路的在前面，能平行的標出來。

1. **`VER-03`**（讓 `go test ./...` 能當 gate）與 **`VER-01`**（音訊套件能建置）——
   沒有可信的全套訊號，後面每一項的「完成」都無法驗證。成本低，先做。
2. **`RE-01` ECL 有序副作用**——矩陣列為全域 P0，也是「單點測試通過、正常流程仍卡住」的首要嫌疑。
   在它閉合前，`RE-04`／`ENG-01` 做出來的內容都可能要重做。
3. **`RE-11` 未定義區段判定**——`WORKLIST.md` 第 559 輪第 5 步，函式層盤點的最後一塊。
   它會決定 585 個「邊界碎片」裡有沒有漏掉的真實程式碼。可與第 2 項平行。
4. **`RE-12` 資料段表格取出**——多份規格卡在同一個原因，取出後 `ENG-04`／`ENG-05`／`ENG-06` 一起解鎖。
5. **`CHT-01`／`CHT-02`**——雙位元組處理是繁中化的地基，愈晚改代價愈高。可與 2–4 平行。
6. **`RE-04` ＋ `ENG-01`**——事件清冊與內容外部化是同一件事的兩端，一起做。
   這是工作量最大的一塊，也是「開場到結局」的實體。
7. **`RE-06`／`RE-07` ＋ `ENG-07`／`ENG-08`**——戰鬥系統。
8. **`RE-05` ＋ `ENG-10` ＋ `VER-07`**——存檔。
9. **`UI-01`／`UI-02`／`AUD-01`**——表現層，需要原版 runtime 當 oracle。
10. **`CHT-06`／`CHT-07`**——全量翻譯，在內容外部化之後才有穩定的翻譯單元。
11. **`VER-04`／`VER-05`／`VER-06`**——最終驗收。
12. **`RE-10`／`AUD-02`**——不阻塞 DOS 單作通關，最後補。

## 維護

- 每個項目完成時更新 [`coab-re-coverage-matrix.md`](coab-re-coverage-matrix.md) 對應列的層級與判定，
  並用 [`re-closure-record-template.md`](re-closure-record-template.md) 留紀錄。
- 被新證據推翻的斷言直接改寫正文，推翻紀錄集中到 `CONTEXT.md` 的「已被推翻的斷言」。
- 本清單不使用百分比；狀態只由 R1–R5 的逐層證據推進。
