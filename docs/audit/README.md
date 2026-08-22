# 原始碼資料分離稽核

本目錄保存機械產生、可由正式測試驗證的技術債基線。它不是豁免清單，也不是
允許 Go 保留固定數量中文的「額度」。

## ECL 全事件靜態清冊

[`ecl-event-catalog.json`](ecl-event-catalog.json) 與
[`ecl-event-catalog.md`](ecl-event-catalog.md) 由 `cmd/ecl-event-catalog` 從原始
`curseoftheazurebonds.zip` 重生。它保存六個 ECL DAX、25 個 block、125 個 lifecycle
entry、靜態可達 instruction／edge 與跨 effect-kind 候選。這是 parser／控制流 inventory，
不是完整 runtime side effects；限制與驗證見
[`spec 557`](../spec/557-ecl-event-catalog-and-ordered-effects-audit.md)。

JSON 對 packed text operand 只保存長度與 SHA-256，不複製原文 payload。

## ECL opcode 有序副作用記錄

[`ecl-opcode-effect-phases.md`](ecl-opcode-effect-phases.md) 與同名 `.json` 由同一支
`cmd/ecl-event-catalog` 從 `internal/eclcatalog/phases.go` 的表產生，逐 opcode 記錄
DOS handler 位址、PC 推進方式與 commit phase（`immediate`／`pause_before_commit`／
`deferred`／`commit_point`／`terminal`／`control_flow`／`unknown`）。

`unknown` 是預設值：corpus 出現的 opcode 一定要有一列，台帳有的 opcode 也一定要在
corpus 出現，兩邊由 `VerifyPhaseCoverage` fail-closed 擋住。所以「還有幾支 handler
沒讀」可以直接從表尾的計數讀出來，不必重新推導。逐支反組譯依據見
[`spec 1104`](../spec/1104-ecl-opcode-ordered-effect-phases.md)。

## ECL 文字覆蓋

[`ecl-text-coverage.md`](ecl-text-coverage.md) 與同名 `.json` 由
`cmd/ecl-text-coverage` 產生，逐頁列出原作的玩家可見文字與 game pack 的接線
狀態。**一頁 ＝ 一個 `12h PRINTCLEAR`**（那條指令重設文字游標，就是開新頁；
spec 1104）。

這份是內容產出（`ENG-01`）的分母：先前只知道 remake 寫了幾條 `text_rules`，
不知道原作有幾頁，所以看不出還剩多少。每寫一條規則，「未接上」就掉一。

⚠ 分母只會再往上：`ON GOTO`／`ON GOSUB` 的動態目的地與選單分支仍未納入可達性。

⚠ 分頁另有兩種結構性誤差，寫規則前要知道：

- **文字被 `GOSUB` 插入的頁**（報告裡標 ⚠，全 corpus 12 頁）真實文字比報告多一段，
  子程式印的字本工具不追。插入點可能在中間（巫師塔的樓梯），也可能在開頭當前綴
  （`ECL4.DAX/0x25` 八個遭遇頁的「YOU ARE ATTACKED BY」）。
  ⇒ `all_contains` 的片段**不可以跨過插入點**。這些頁的規則寫了也不會讓報告的
  「已接上」加一，**實際覆蓋比報告的數字高**。
- **被 `GOSUB` 呼叫的純文字子程式會自成一頁**（`UP`／`DOWN`／`YOU ARE ATTACKED BY`
  ／`WHAT TO YOU DO?`），實機不會單獨出現。⇒ **不要**替它們寫規則；那種只有
  一兩個字的「頁」寫成規則會攔截到別的文字。

實例：`ECL5.DAX/0x33` 的 `THE STAIRS LEAD ⟨UP｜DOWN⟩ HERE.` 中間那個字由
`02h GOSUB 89B2h` 印出；`ECL4.DAX/0x25` 的八個遭遇頁前面都有 `GOSUB 9534h` 印的
「YOU ARE ATTACKED BY」。兩處的規則都已寫好，由
`gamepack.TestWizardTowerInteriorIsGamePackDriven` 與
`TestWildernessEncountersAreGamePackDriven` 用**實機的字串序列**釘住——
那兩支測試餵的不是規則自己的 `all_contains`，所以片段跨越插入點會紅。

## 譯名一致性

[`glossary.md`](glossary.md) 與同名 `.json` 由 `cmd/glossary-audit` 產生，把
[`../knowledge/coab-glossary.md`](../knowledge/coab-glossary.md) 的詞條、
`combatant_name_rules` 自動匯入的怪物名，與三份繁中目錄（game pack locale、
UI locale、工具訊息）比對。規則與九組已修正的不一致見
[`spec 1107`](../spec/1107-translation-glossary.md)。

閘是 fail-closed 的：禁用寫法出現、詞條在資料裡完全沒出現、同一原文有兩種譯名、
同一譯名對到兩個原文，都會讓 `internal/glossary` 的測試紅。

⚠ 閘擋不住的那一半：誤用的寫法若同時是**別的詞條的正確譯名**（`Haptooth` 誤寫成
`哈普`，而哈普正是 `Hap`），沒有任何字串比對能分辨，只能逐句對照原文。表的
〈通用寫法〉一節把這類判準寫成規則。

## 誰寫這個 DS 位址

[`dseg-writers-map-registers.md`](dseg-writers-map-registers.md) 由
`cmd/dseg-writers` 產生，把 resident 與每個 overlay 掃過一遍，逐處列出對
`720Fh`／`7210h`／`7211h`（地圖暫存器）的**寫入**、**取位址**與**讀取**。
結論與方法見 [`spec 1183`](../spec/1183-map-register-writer-census.md)。

工具本身是通用的：`-cells` 換一組位址就能問別的格子。要換遊戲版本就換
`-root` 與 `-resident`（⚠ PC-98 的資料段位移與 DOS 差 `0x3292`）。

⚠ 它是**位元組線性掃描**，失敗方向與 `cmd/ecl-cell-refs`（走控制流）相反：
線性掃描有偽陽性、沒有偽陰性；控制流掃描沒有偽陽性、但掃不到沒被認成程式碼
的部分。要下「沒有人寫」的結論之前兩種都要跑。本專案已經被走控制流那一側的
假零咬過兩次（spec 1095 的 `7EE2h`、spec 1153 的 `4BE7h`）。

⚠ 報表本體是英文的。`cmd/coab-audit` 的漢字 gate 只准數量下降，而報表就是
位址與助憶碼的表格；分析文字寫在 spec 1183，工具的推理寫在檔頭註解。

## Go 漢字字串基線

`go-han-literals-baseline.json` 由 `cmd/coab-audit` 使用 Go AST 產生，只掃正式
非測試 `.go` 字串 literal：

- 忽略註解、`*_test.go`、JSON／Markdown、`workplace/` 及 nested engine；
- 以 repository-relative path、函式、完整字串 SHA-256、出現次數及債務分類
  建立 exact multiset；
- 不把中文內容複製進基線；
- 新增、改字、搬動、增加副本或刪除後未更新基線，都會讓
  `TestRepositoryGoHanLiteralBaselineIsExact` 失敗。

目前分類是遷移排序用 heuristic：

- `localization_debt`：位於 localize／Journal bridge 的歷史字串；
- `frontend_ui_debt`：Ebiten command 的玩家可見 UI；
- `runtime_ui_debt`：其他 runtime、工具與尚未細分字串。

分類不代表後兩者可以永久留在 Go。最終目標是：本作劇情、人名、地名、物品、
法術、選項、手札與玩家可見 UI 都由 stable ID＋locale／game-pack 驅動；Go 只
保留 action、format contract、layout 與必要技術診斷。

每次遷移流程：

1. 先把正式文字移入 locale／game-pack，接通 stable ID 並測正常玩家路徑。
2. 刪除 Go fallback 及其他資料複本。
3. 在 Docker 內執行 `go run ./cmd/coab-audit -write-baseline`，只能接受數量下降
   或經審查的分類改善。
4. 執行 `go run ./cmd/coab-audit` 與正式套件 gate。
5. 在 READY spec／狀態表記錄前後數量；不可只更新基線掩蓋新增債務。

第 452 輪初始證據：1,260 個 signatures、1,315 次 occurrences，其中
`localization_debt=409`、`frontend_ui_debt=164`、`runtime_ui_debt=742`。
