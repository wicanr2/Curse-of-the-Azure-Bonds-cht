# 主線分段驗證計畫

從頭到尾跑一次不是驗收方式，是煙霧測試。這份計畫把主線切成可獨立進入、
獨立驗收的段，並把**段與段之間的交接**也列成要驗的東西。

## 一、現況（量出來的）

| 項目 | 數字 | 來源 |
|---|---:|---|
| ECL block | 25 | `docs/audit/ecl-event-catalog.md` |
| lifecycle entry | 125（每個 block 五個） | 同上 |
| `NEWECL` 出邊 | **47** | `docs/audit/ecl-block-graph.md`（`cmd/ecl-block-graph`）|
| 沒有出邊的 block | **2**（開場 `0x52`、結局 `0x43`）| 同上 |
| game pack 宣告的地圖 | 20（其中 **9 張還是 `original.geoN.block-NN` 佔位名**） | `gamepack/pack/00-core.json` |
| 直入旗標 | 62 個 flag，其中約 25 個是劇情 checkpoint | `cmd/azure-bonds-game` |
| 主線整條跑的測試 | `TestRealNewGameRunsToTheEnding`，**單一函式、檔案 1,625 行** | `internal/game/campaign_normal_test.go` |

連續的新遊戲 session 目前走到眼魔洞穴與散提爾堡世界選單。

## 二、為什麼要分段

- **定位成本**：現在那條主線測試是一個函式。中間任何一步壞掉，得到的是
  「整條紅」，要靠讀 1,600 行去找是哪一步。
- **迴圈長度**：每加一章，那條就更長；到結局會變成一次跑十幾分鐘的東西，
  而它每次只回答一個 yes/no。
- **段與段的交接才是真正會壞的地方**：`NEWECL` 換 block、換 GEO／WALLDEF／
  8X8D 資產、隊伍狀態帶過去、pending continuation。整條跑會**掩蓋**這些——
  中間狀態對不對，從最後一步看不出來。

判準沿用 `AGENTS.md`：每段可用直入旗標進入，但每段都要走到**該段的正常結束
狀態**，而且段與段之間的狀態交接本身也要列成一段驗。

## 三、段從哪來：不是憑感覺切的

**一個 ECL block ＝ 一段**，**一條 `NEWECL` 邊 ＝ 一段交接**。
完整的圖在 [`docs/audit/ecl-block-graph.md`](../audit/ecl-block-graph.md)：
**25 個 block、47 條出邊**，只有 `ECL1/0x52`（開場）與 `ECL6/0x43`（結局）沒有出邊。

兩個世界地圖 hub 是樞紐，其餘 block 進出都經過它們：

| hub | 出邊 |
|---|---|
| `ECL1/0x50` | `0x03`、`0x25`、`0x31`、`0x35`、`0x40`、`0x51` |
| `ECL1/0x51` | `0x10`、`0x11`、`0x15`、`0x20`、`0x45`、`0x50` |

換 block 的機制是：主迴圈（`overlay-02:3772`）載入 `bank0 +1E4h`（LastECL）
指的 block、載完寫回去；`20h NEWECL` 就是改那一格（spec 1141）。

⚠ **不要用事件目錄的圖判斷「這個 block 沒有出口」**：它的可達性不跟 `ON GOTO`
的目的地，算出來只有 21 條邊，而且兩個 hub 會看起來一條出邊都沒有。

## 四、工作項目

### 階段 0：把段切出來（不改遊戲行為）

| ID | 項目 | 產出 | 驗收 | 粗估 |
|---|---|---|---|---|
| `SEG-01` | ✅ **完成**：轉移機制查清楚了（spec 1141）——主迴圈讀 `LastECL`、`NEWECL` 改它；圖由 `cmd/ecl-block-graph` 產生 | `docs/audit/ecl-block-graph.md` ＋ spec 1141 ＋ 形狀回歸測試 | 25 個 block 每一個都說得出怎麼進去、怎麼出來 ✓ | 已完成 |
| `SEG-02` | ✅ **完成**：`LOAD FILES` 第一個運算元 ＝ 那一段載的地圖區塊；`area_id`／`script_block` 與 block 編號 join 得起來 | `docs/audit/ecl-block-graph.md` 的段落清單 ＋ 三條回歸測試 | 25 個 block 的地圖歸屬都算得出來 ✓；剩三張地圖缺 `script_block`、一組宣告不一致（見報告）| 已完成 |
| `SEG-03` | ✅ **前半完成**：段的 id 一律 `ECL{成員}/0x{block}`（機械、與 game pack 命名無關），標籤放 `docs/plan/segment-labels.json` | 段落清單的標籤欄 ＋ 報告 | 25 段的標籤逐條有原作敘述為證 ✓；進入方式／結束條件／快照三欄要等 `SEG-04`／`SEG-11` | 已完成前半 |
| `SEG-04` | ✅ **完成**：`-segment <id>` ＋ 註冊表 `internal/segment`；收完整 id、只給 block 編號、或既有旗標名 | `internal/segment` ＋ `State.StartStorySegment`／`EnterSegment` ＋ 五條測試 | 25 段逐段進得去 ✓；順手修掉兩個讓段入口畫面落空的素材鍵錯誤，剩一個真的缺口（見報告）| 已完成 |

### 階段 1：把既有的整條跑拆成段（行為不變）

| ID | 項目 | 驗收 |
|---|---|---|
| `SEG-10` | ✅ **完成**：那一條 790 行的函式拆成 12 個 subtest（同一條 session）| 正規化逐行比對證明**沒有任何遊戲動作被加進來或拿掉** ✓；段紅的時候直接看得出是哪一段 ✓（報告見 `docs/plan/seg-10-verification-report.md`）|
| `SEG-11` | ✅ **完成**：25 段的**入口**往返閘 ＋ 12 段的**結束**快照往返 ＋ `-segment-snapshot`／`-party-load` 交接 | 存得下去、讀得回來 ✓（報告見 `docs/plan/seg-11-verification-report.md` 與 `seg-10-verification-report.md`）|
| `SEG-12` | ✅ **完成**：`TestEveryNewECLEdgeHandsOff`——來源段存快照 → 讀回 → 帶著來源 block 當 `LastECL` 進目的段 | 47 條邊每條一個子測試 ✓（報告見 `docs/plan/seg-12-verification-report.md`）；來源用的是段的入口狀態，段結束狀態要等 `SEG-10` |
| `SEG-13` | ✅ **完成（作法有調整）**：主要 gate 已經是段測試（12 個 subtest ＋ 段界快照往返），整條跑是它們共用的同一條 session | ⚠ 快照**不進 repo**：存檔格式一改整批失效，而分段診斷與快照往返兩項好處已經拿到（理由見報告 §四）|

⚠ `SEG-11` 是關鍵：沒有快照交接，後面的段還是得從頭跑，分段就只是把同一條
長跑切成看起來比較短的樣子。**所以它排在 `SEG-10` 前面做**——先把交接機制驗
起來，拆段時才有東西可以接。

### 階段 2：把未接的段接上（依 `NEWECL` 圖排序）

每個 block 一個工作項目，內容固定四件事：**進入邊 → `initial` →
`per_turn`／`search_location` → 離開邊**。

現況由 `cmd/segment-coverage` 量出來，逐段的結果在
[`docs/audit/segment-coverage.md`](../audit/segment-coverage.md)：

- **入口有劇情：22 段**。入口文字與選項**沒有一段落回原文**，從入口再走一步也
  都走得動（`ECL1/0x52` 與 `ECL5/0x32` 走到戰鬥，其餘走到世界地圖或地城）。
- **入口不出文字：3 段**（`ECL2/0x02`、`ECL3/0x12`、`ECL4/0x21`）。這三段是從
  別段被帶進來的——城區共用地圖、地下第二層、神殿與牢房——劇情在前一段的轉移
  與每回合／搜尋生命週期裡，不在 `initial`。**不是資料缺漏**。

⇒ 階段 2 剩下的工作不在「段的入口」，而在**段內**：每回合／搜尋生命週期、
離開邊的觸發條件，以及段內的戰鬥。已經有段內逐步覆蓋的是 `SEG-10` 那 12 段
（哈普村到眼魔洞穴）；其餘 13 段只驗到入口與交接。

⚠ 只有 `ECL6/0x43`（內城遺跡，結局）沒有出邊；`0x40` 與 `0x42` 都有
（`0x40 → 0x42`／`0x50`、`0x42 → 0x40`／`0x43`）。

### 階段 3：跨段的不變量（每段都要過）

| ID | 不變量 |
|---|---|
| `SEG-30` | **存檔往返**：每段的結束快照都存得下去、讀得回來、狀態一致 |
| `SEG-31` | **隊伍連續性**：HP／記憶法術／經驗／物品／效果跨段不掉 |
| `SEG-32` | **語系**：該段所有玩家看得到的字都有 `zh-TW`，沒有 fallback |
| `SEG-33` | **音樂／音效綁定**：該段的 cue 有宣告且觸發得到 |

## 五、每一段共用的驗收門檻

一段算「完成」要同時滿足：

1. 由 `-segment <id>` 直入進得去；
2. 走到**該段的正常結束狀態**（不是走到一半截斷）；
3. 該段的 lifecycle entry（`initial`／`per_turn`／`search_location`／
   `pre_camp`／`camp_interrupted`）全部跑過或明講為什麼跑不到；
4. 產生的快照能被**下一段**載入；
5. 階段 3 的四個不變量都過。

## 六、明確不做

- **不**把「從頭到尾跑一次」當主要驗收。它保留為 smoke，回答的是「串得起來嗎」，
  不是「每一段對不對」。
- **不**為了讓段測試變綠而放寬該段的正常結束條件。
- **不**在段與段之間手動塞狀態：交接一律走存檔快照，手塞會讓交接的缺口驗不出來。

## 七、開工前要先有答案的

1. ~~`SEG-01`~~ ✅ 已解（spec 1141）：轉移是 `NEWECL` 改 `LastECL`，主迴圈載它。
2. ~~9 個佔位地圖的名字~~ ✅ 已解（`SEG-03`）：段的 id 不綁地圖名，標籤逐條
   有原作敘述為證。`zhentil-keep.beholder-cave` 的 `script_block 0x22` 實機量過
   是對的（`SEG-10` 的 `ECL4/0x22` 段）；⚠ 未解的是**幾何怎麼換成 `GEO4/0x25`**
   ——`ECL4/0x22` 的 `LOAD FILES` 是 `21`，remake 目前用 game pack 事件重建。
3. ~~段的快照要存哪一層~~ ✅ 已解：整份 `SavePartyFile`，順便驗了存檔往返
   （`SEG-11`）。
