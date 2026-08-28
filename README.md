# 《青色枷的詛咒》繁體中文化／Remake

SSI《Curse of the Azure Bonds》的資料驅動重製專案。目標是讓玩家以繁體中文從
開場玩到結局。**主線已經從開場走到結局**：除了規則層的同一條
session，連續鍵盤輸入也已在第 **14,380 幀**觸發 `GameWon()`，全程只走
`(*app).Update()`（899 格、16 個 ECL 段、393 句話、落回原文 0）。這證明輸入層
可達通關，但不等於一般強度隊伍的人工完整試玩驗收。

本 repo 負責 CoAB 的 game pack、翻譯、原始素材轉換、攻略與整合測試；可重用的
ECL／DAX／GEO／戰鬥／存檔引擎在獨立的
[golden-box-remake-engine](https://github.com/wicanr2/golden-box-remake-engine)。
劇情、座標、物品與翻譯資料不寫死在共用 engine。

## 目前狀態（2026-08-28 實測）

目前整體自評 **97／100**，交付目標是玩家可見體驗近似原版 **99%**。後續採
風險導向代表性抽樣，不再以所有狀態全量逐像素／逐分支閉合作為完成條件；
阻塞通關、存檔毀損、落回原文與主要規則偏差仍須歸零。詳見
[完成度自評](docs/knowledge/remake-completeness-assessment.md)。

| | |
|---|---|
| 反組譯覆蓋 | 2,874 個函式全部進台帳，**待解讀 0**（已解讀 2,137／不阻塞 162／邊界碎片 575）；未定義位元組 36,386 已逐段形狀分類＋對回函式台帳，**每一個位元組都有分類或一手判定（無判定 0 段）**（spec 1203 四刀）|
| 規格目錄 | 1,229 份 `docs/spec/*.md`（2026-08-28 現場計數，含索引）；狀態與證據邊界以各文件及 [`docs/spec/README.md`](docs/spec/README.md) 為準，不能由文件數推論完成度 |
| 程式 | 498 個 `.go`／142,653 行（2026-08-28 現場計數，不含巢狀 engine repo 與 `workplace/`）；`internal/game` 沒有區域或劇情專屬檔 |
| ECL 文字 | 控制流可達 1,022 頁，**未接上 0** |
| ECL 副作用 | 可達 14,177 條指令，`done` **14,177**／`partial` **0**；opcode `done` 61／`partial` 0 |
| 戰鬥法術 | 可宣告 **73／73** 全部有 handler、視覺與音效 |
| 命中判定 | 與原作同一式：FIRE KNIFE 自己打自己，原作與 remake 都要 **d20 ≥ 18** |
| 第一人稱畫面 | 585 種牆面配置**全覆蓋**、共 **1,498 張**與原版逐格比對：**1,498 張全部完全相同、差 0 格**（第三次全量重測）|
| 全滅 | 隊伍全倒會**結束遊戲**、落到全滅畫面再回標題，與原作的 `DS:4FC7h` 同一個結局 |
| 存檔 | 角色記錄 422 bytes：decoded 299／documented 123／**unknown 0** |
| UI 詞條 | game pack `zh-TW` 1,598 條、`assets/locale/zh-TW.json` 1,183 條 |
| **主線** | **開場到結局同一條 session 跑得完**，拆成 25 個段 subtest（23 段主線＋2 段接進主線的段內支線），每段結束的快照都存得下去讀得回來 |
| **鍵盤通關** | 從標題開始、全程只走真實前端按鍵，第 **14,380 幀**觸發 `GameWon()` |
| **一般強度基線** | pending Bless 無限重試已修；關閉隊伍強化後跑 2,455 幀：251 格、8 個 ECL 段、199 句、落回原文 0，最後真實全滅重開 1 次。手動戰鬥對照也在同一戰役區段全滅；這是合法停止結果，不再把戰術／loadout 或自動通關列為產品缺口 |
| **發行工程** | `0.1.0-dev.20260828` 已重生 Linux x86_64 AppImage、Windows x86_64 ZIP、macOS x86_64／arm64 ZIP；Linux AppImage 與 Windows EXE（Wine）已實際啟動與繁中截圖驗收，兩張開場 PNG 逐位元組相同；Windows／macOS 尚待真機啟動；內部推廣片已通過技術與畫面抽樣但音樂權利仍 pending（[稽核](docs/audit/release-0.1.0-dev.20260828.md)） |
| 主線分段 | 25 段全部可直入（`-segment ECL{成員}/0x{block}`）；47 條 `NEWECL` 邊逐條交接；段入口有中文劇情 22／不出文字 3 |
| 戰利品 | `1Ch CLEARMONSTERS` 連同**還沒領走的戰利品**一起丟掉（spec 1145）；corpus 唯一走得到那條路的是提爾佛頓火刀首領的重打迴圈，少了它重打一次就多領一次 |
| 遭遇距離 | `0Ch` 擺距離、`0Dh APPROACH` 走近、`29h` 重設（spec 1146）。⚠ `APPROACH` 的減一是**演出**——`29h` 進門會把它蓋回去 |
| 遭遇選單 | `29h` 的三句旁白依**距離**挑（spec 1144）：20 處逐句對過譯文，玩家演得到的 20 句**全部接上**（`cmd/ecl-encounter-text`）。⚠ 距離本身還是拿上限近似——原作由地圖座標算 |
| 段內支線 | 全遊戲對照表 **328 個場景對到格子**（`cmd/ecl-cell-events`，17 個有地形分派的 block）；逐格實測 **288 個分派索引**（`cmd/cell-sweep`），**276 格中文、落回原文 0 格**，12 格沒演出來全部歸得了因（四種盤點限制，還沒歸類 0）|
| 世界地圖 | 路段編號 `4C9D ＝ 出發地 × 4 ＋ 方向`（spec 1143）；從 13 個地點各掃一次、**698 個分支**，敘述與選項**落回原文 0 條**（`cmd/route-sweep`）|
| 跨段不變量 | 經驗值不倒退、隊伍變動位置宣告好；**裝備／記憶法術／效果**在 23 個段界存下去讀得回來且沒有變動（`SEG-31`）；選曲逐格對回 PC-98 原作的 selector 表（`SEG-33`，spec 355）|
| 怪物資料 | `MON*ITM` 物品鏈與怪物自動換裝已接；特殊效果的已知實作缺口 0，但仍有 28 個效果碼／69 筆記錄是 `unread`，不能宣稱全部對齊 |
| 音訊格式 | 正式播放以 OGG 為主：12 首 PC-98 音樂與 9 個音效；`MSCDRV.EXE` 即時合成只保留作研究 oracle／缺檔 fallback |
| 圖像獨立化 | `assets/` 共 2,528 張 PNG：780 張人物／怪物／場景、1,741 張 runtime 圖像、6 張 Journal 插圖與 1 張參考圖；TILES、戰場地形、AREA／共用符號、SKY 與牆片皆已切換 PNG／JSON，無圖像 DAX fixture 的 runtime 代表抽樣已通過，還差各平台封包 smoke |
| 現代介面 | 預設 A6 細石框＋左上明亮金雕內框；F1 Help、F2 即時 theme、F3 雙層攻略地圖、F4 三種解析度、F10 存檔後離開；視窗可任意拉伸並填滿 |
| 手札 | locale 宣告 **64 則**，全部由 48 條內容規則的 `journal_message_ids` 解鎖；六張手冊地圖／插圖（含手札 59 的眼魔洞穴圖）在手札畫面按 `I` 彈窗顯示（spec 1109）|
| `24h COMBAT` | 199 處逐處分類：**153 處真的要打、46 處走服務分派**（`cmd/ecl-combat-sites`）|
| 跨段不變量 | 整條主線 260 句話**落回原文 0 句**；同一角色的經驗不倒退、隊伍變動位置宣告好；22 個段界都有曲子 |
| 走訪可達性 | 288 個分派索引裡**走得到 287**（走路的上限也是 287）；沒達成的 1 格是幾何上斷開，而逐格實測站上去演得出來 |
| 地圖圖集 | 18 張第一人稱地圖逐張畫成「事件在哪一格、從哪幾格離開」（`cmd/map-atlas`，`docs/reference/maps/`）；宣告的出口逐個對回 GEO 的移動遮罩 |
| 全套測試 | `./tools/go.sh test ./...` 全綠 |

### 已接通的正常玩家路徑

同一個新遊戲 session 可從開場一路打到結局：提爾佛頓開場 → 城市設施 →
城門／盜賊公會 → 下水道 → 火刀據點至首領 → 戰後世界地圖 → 阿沙本福德 →
立石群 → 艾森布拉城外 → Hap／熔岩洞／巫師塔 → 散提爾堡 →
眼魔洞穴（手札 59、Dexam 雙戰）→ 猶拉什 → 希爾斯法 → 立石群的灰袍男子 →
密斯卓諾墓園（紅網、黛米爾公主）→ 外城遺跡 → 內城遺跡 → 儀式與爪牙戰 →
二樓東北角 → **擊敗提朗瑟克斯的結局選單**。

可與主線相容的段內支線也在同一條 session 裡：內城一樓九間房＋二樓五間房、
墓園挖墓的盜墓者（骸骨選 REBURY）、外城下水道口（放走石像鬼、拒絕進柵口）。
**刻意不走的兩類**：螳螂人營地三場（入口即十二隻螳螂人，攻略明寫可避開的
高風險選配，逐格取樣已涵蓋敘述）與下水道整條（與正門互斥的另一條進內城
路線，一條 session 只能走一邊，`myth_drannor_test` 的直入測試整條走過）。
其餘支線由段界快照逐格取樣涵蓋（9 段、123 格、落回原文 0 格）。

連續鍵盤路線已通關；後續只依風險抽測代表性玩家路徑、非第一人稱
畫面、中文與音訊，不再追求全量閉合。對外交付的硬関門是無圖像 DAX
fixture 封包抽測、Windows／macOS 真機啟動、macOS 簽署／公證與授權清單。

### 角色建立

正常入口已依原版順序重現「種族 → 性別 → 職業 → 陣營 → 能力值／HP →
姓名 → 存檔」，資料全部取自原作資料段並轉成 JSON；remake 只增加目前步驟、
有效按鍵與隊伍人數提示：

- 六個可選種族——原作的顯示迴圈沒有收半獸人的分支，所以建角選不到它，
  但編號仍維持原作的 1..7 不重新連號。
- 職業選項由種族查表；陣營選項再由職業組合查表（聖騎士只有守序善、
  遊俠只有三個善）。
- 屬性擲點是 `3d6 + 1`、每個屬性擲六次取最大，加上七個種族各自的屬性調整。
- 多職角色的 HP 是八個職業槽各擲一次生命骰、逐槽體質加值後除以職業數。
- 名字支援中文（退格按字元不按位元組），`Save <名字>?` 答 `N` 角色不留。

完成一名角色後會回到隊伍組裝頁，可繼續建立角色或完成隊伍。範本 adapter 只留給
測試與內部相容用途，不是正常玩家入口。

## 畫面預覽

### 原版 DOS 畫面總覽

下圖只使用 `docs/reference/original-dos/` 內已保存的原版 DOS 擷取，依序涵蓋
標題／設定、建角能力值、角色狀態、冒險第一人稱與戰鬥。各格維持原始長寬比並以
最近鄰縮放；沒有混入 remake 畫面。第一人稱的 88×88 場景原本就位於 128×136
左上面板內，場景外圍的黑色區域不是圖片漏貼或未填滿。

![原版 DOS 畫面總覽：標題、設定、建角、角色狀態、第一人稱冒險與戰鬥](docs/reference/original-dos/original-screen-overview.png)

總覽中的第一人稱畫面可證明 DOS GUI 外框與場景尺寸，但其訊息為
`NOWHERE IN THE REALMS`，不能當成提爾佛頓特定座標的牆面 oracle；同狀態牆景
比對另見 [spec 1134](docs/spec/1134-original-first-person-oracle.md)。

### Remake 代表畫面

以下是目前 remake 的代表畫面，皆在 Docker／Xvfb 產生並逐張檢視。石框與 88×88
場景是原始素材證據；640×480 的中文延伸與完整戰鬥畫面仍是 `layout-reconstructed`，
不能據此宣稱整作完成。

![原作開場序幕：營火伏擊、三名 NPC 入隊與五枚青色符印的緣起](docs/screenshots/opening-prologue-remake.png)

![旅店人物事件：DOS 石框、HEAD／BODY 舞台與繁中敘事](docs/screenshots/gold-box-layout-adventure.png)

![A6 現代 theme：細石外框、左上雙層結構與明亮金雕內框](docs/screenshots/a6-modern-theme-adventure.png)

![A6 現代 theme 的完整攻略疊圖：目前座標、16×16 地圖與事件點說明](docs/screenshots/a6-guide-overlay.png)

![原版四段建角的名字輸入段](docs/screenshots/guided-creation-name.png)

![繁中戰鬥：原作佈署演算法放的開場隊形，遠距離遭遇的敵隊要逐回合逼近才入鏡](docs/screenshots/gold-box-layout-combat.png)

![提爾佛頓第一人稱：木板牆、石牆與逐格收斂的側牆，原版 88×88 場景內框](docs/screenshots/tilverton-first-person-remake.png)

![Burial Glen 紅網戰鬥：四隻巨蛛的 CPIC 圖示、原版地城素材與戰鬥 footer](docs/screenshots/burial-glen-red-web-spiders.png)

### Remake PNG sprite／tileset 總攬

下圖直接抽樣 `assets/sprites/` 與 `assets/runtime-images/` 的獨立 PNG，涵蓋人物／場景、
戰鬥 sprite、世界／區域圖塊與戰場 tileset。這些是 release 會載入的 remake 資產，
來源仍是合法持有的原版資料經轉換後獨立存放；**不是現代重繪素材**。A6 目前只替換
介面框線與疊圖，日後的現代 sprite／tileset 必須維持相同索引、錨點、方向與碰撞語意。

![Remake 獨立 PNG sprite／tileset 總攬：場景人物、戰鬥小人、世界區域圖塊與戰場地形](docs/reference/remake-png-asset-overview.png)

總攬圖不需原版 ZIP，可由 `go run ./cmd/asset-overview` 在 Docker 內重生；抽樣採排序後
等距選取，所以同一份資產集合會產生相同畫面。

### 法術演出的四種通道

原作不是「一支法術一套圖」。下面每一格都是從遊戲檔案直接取出來的原始素材，
由左到右是四種通道：

![法術演出的四種通道：共用投射物、閃電電弧、魔法命中、兩種雲、睡眠、弓箭](docs/reference/spell-visual-channels.png)

| 通道 | 圖 | 說明 |
|---|---|---|
| 共用施法投射物 | `COMSPR 05`／`85` | 在分派到各支 handler **之前**就播，所以每一支法術都有 |
| 閃電電弧 | `COMSPR 06`／`86` | 沿線逐格播，全表唯一 |
| 魔法傷害命中 | `COMSPR 0A`／`8A` | 對每個受傷的目標各播一次 |
| 持續區域格 | `RANDCOM 04`（綠＝惡臭之雲）／`02`（藍白＝致命毒雲）| 寫進地圖格，效果活著就一直畫 |
| 產生器／投射物 | `COMSPR 09`（睡眠）／`00`（弓箭）| 睡眠由參數合成；弓箭有八個方向 |

判讀見 [spec 1126](docs/spec/1126-spell-visual-slots.md) 與
[spec 1128](docs/spec/1128-cloud-areas-are-the-obstacle-terrain.md)。
怪物的特殊攻擊（吐酸、龍息、凝視、丟電光）也走同三組圖——那七支 handler
由 `INITSPELLS` 尾段接進效果分派表，見
[spec 1202](docs/spec/1202-initspells-effect-table-tail.md)。
兩種雲同時是戰術地圖的**障礙格**（地形碼 `1Eh`／`1Ch`）——低階角色繞開毒雲、
七級以上的老手硬闖。

目前 14 張代表畫面的產生指令與雜湊在
[`docs/screenshots/manifest.json`](docs/screenshots/manifest.json)，由
`cmd/screenshot-audit` 驗；近期修掉的圖層對齊錯誤見
[spec 1130](docs/spec/1130-screenshot-layer-alignment.md)，第一人稱牆面的
第 0 段符號與洋紅透明鍵見 [spec 1131](docs/spec/1131-wall-symbol-group-zero.md)。

第一人稱那一張是**與原版逐格比對過的**：同一格、同一個朝向，把原版 DOS 版
在 DOSBox 裡跑出來的畫面與 remake 的畫面各取 88×88 可見區，EGA 量化之後比
16 色索引。全作 585 種牆面配置全部有原版畫面可比，共 1,498 張比完、
**1,498 張逐格完全相同**。原版擷取與
索引收在 [`docs/reference/original-dos/first-person/`](docs/reference/original-dos/first-person/)，
重跑比對是一行 `python3 tools/fp-oracle-compare.py …/index.tsv`；
方法與剩下那一格見 [spec 1134](docs/spec/1134-original-first-person-oracle.md)。

更多地圖、人物舞台、戰鬥時間軸與素材圖在[截圖目錄](docs/screenshots/)；原版忠實
theme 與日後美化 theme 分開維護。

## 從乾淨 clone 開始建置

可重用的 engine 是**獨立的私有 repo**，而 `golden-box-remake-engine/` 與
`workplace/` 都不在本 repo 的版控裡，所以剛 clone 完要先把相依準備好：

```sh
git clone https://github.com/wicanr2/Curse-of-the-Azure-Bonds-cht.git
cd Curse-of-the-Azure-Bonds-cht
tools/engine-bootstrap.sh      # clone／fetch engine，打包 go.mod 鎖的那個 commit
tools/go.sh test ./...         # 全套測試，Go 工具鏈跑在 docker 裡
```

`tools/engine-bootstrap.sh` 只認 `go.mod` 鎖住的版本，不動 engine 的工作區。
編譯一律走 docker（`tools/go.sh`），主機不需要裝 Go。

## 玩家文件

- [繁體中文攻略入口](docs/guide/README.md)：目前可玩的主線、分區攻略與無雷提示。
- [繁中遊玩手冊](docs/manual/curse-of-the-azure-bonds-zh-TW.md)：按鍵、戰鬥、紮營、
  手札與遊玩說明。
- [給中文玩家的 Gold Box 重製說明](docs/knowledge/golden-box-remake-for-chinese-readers.md)：
  把原本需要查紙本手冊的資訊整合進遊戲與文件。

## 授權

本專案中由權利人擁有的程式碼、繁體中文翻譯、文件與原創內容採
[PolyForm Noncommercial License 1.0.0](LICENSE)：允許非商業使用、修改與散布；
商業使用請透過 GitHub Issues 洽談另外授權。SSI 原版素材、Noto Sans TC、獨立 engine
及第三方依賴不會因此被重新授權，完整邊界見 [`NOTICE.md`](NOTICE.md)。

## 開發與研究文件

- [`HANDOFF.md`](HANDOFF.md)：compact／接手後先讀的最短現況、下一步與不可重開 gate。
- [`AGENTS.md`](AGENTS.md)：**工作規則的單一入口**（完成定義、證據標準、Git 紀律、
  驗證門檻、compact 恢復流程）。
- [歷史缺口盤點](docs/knowledge/coab-remake-todo.md)：2026-08-16～24 的工作拆解與
  證據索引，已封存，**不是目前 TODO**。
- [完成度自評](docs/knowledge/remake-completeness-assessment.md)：逐層自評與
  「到可交付還差的四步」，數字取自 `docs/audit/remake-status.md`。
- [共用 engine 抽取盤點](docs/knowledge/engine-extraction-review.md)：`internal/combat`
  逐檔對「作品中立」界線比一次，標出該搬、要參數化、以及不該搬的部分。
- [完整度矩陣](docs/knowledge/coab-re-coverage-matrix.md)：全遊戲 RE／重建完整度的
  單一權威矩陣（`R1` 原始定位 → `R5` 玩家驗證）。
- [Gold Box 系列重用邊界](docs/knowledge/gold-box-series-reuse.md)：下一款作品可直接
  沿用、只能當候選與永遠留在作品層的最短契約。
- [`WORKLIST.md`](WORKLIST.md)：逐輪工作與決策證據的歷史封存，**不是目前執行順序**。
- [`CONTEXT.md`](CONTEXT.md)：現況與已被推翻的斷言；歷史分冊在 [`docs/context/`](docs/context/)。
- [格式規格索引](docs/spec/README.md)：每份規格標示 `READY`／`DRAFT`、證據與
  `exact`／`strong inference`／`hypothesis`／`unknown` 邊界。
- [歷輪 README 報告](README-history.md)：長篇里程碑與研究紀錄，僅供歷史查閱，
  **不是目前狀態的權威來源**。

## 執行原則

1. 先讓正常玩家路徑可走，再補 fidelity。
2. 遊戲文字、選項、物品、事件與原版資料表由 CoAB JSON／locale 提供；
   Go／engine 只保留可重用機制，不複製劇情常數，也不 hardcode 原版資料表。
3. **先盤點後語意**（2026-08-13 起）：兩平台全模組先建庫盤點，每個函式都要進
   覆蓋台帳，再依玩家可見性排序做語意閉合。不再「遇到問題才反組譯那一段」。
   口徑見 [`AGENTS.md`](AGENTS.md) §2.5。
4. 每份結論都要能回指到哪個模組、哪個位址、哪幾個 byte；IDA／decompiler 的輸出
   本身不是證明。規格要寫明「明確不宣稱」的邊界。
5. 驗收採分段：每段用 `-segment <id>` 直接進入（`-segment list` 列出全部），
   但每段都要走到該段的正常結束狀態，而且**段與段之間的狀態交接本身也要
   列成一段驗**。
6. 每個重大、可展示的里程碑才集中 commit／push；兩個 repository 分開提交。

## 歷史收斂細節（非目前待辦）

以下保留部分規則閉合的證據索引，不可依「剩下」「尚未」等舊措辭重開工作。
目前只有 [HANDOFF](HANDOFF.md) 與
[完成度自評](docs/knowledge/remake-completeness-assessment.md) 列出的現行 frontier。

- **戰鬥回合生命週期**：回合開始那一段已收完（spec 1135／1136）——先攻算式、
  選誰動、DELAY／GUARD／QUICK、定身與機會攻擊都與原作相符，突襲遮罩判定為死碼。
  面向、動作計數與累計轉向三個欄位也接上了（spec 1138），背後攻擊的三道條件
  是字面移植，第二個 AC 的算式已解出並拿 81 筆 `MON*CHA` 驗過（spec 1000 §七）。
  離開接觸的機會攻擊接上了 180° 面向閘（spec 1002／1010），
  AC 與命中的刻度也與原作對齊（spec 1139）。
  離開接觸的機會攻擊四道閘全部解讀並接上（spec 1010），隊員的命中與 AC
  也改吃原作的職業／等級表與力量／敏捷調整表（spec 1140／694／697）——
  整條派生鏈用原版存檔逐項驗過，五個欄位全中。
  **剩下的**：原作 `+09h` 的幾何意義刻意留白（兩個寫入端互相矛盾），
  以及 ECL 的 `24h COMBAT` 那 199 處仍只做了分派。
  （法術那一半已經收斂：可宣告 73 支全部有 handler、視覺與音效。）
- **ECL 可達副作用已全部接線**：14,177 條 `done`、`partial` 0。
  下面這段是逐支收斂的過程紀錄，保留是因為每一支的證據位址還用得到：
  `PRINT RETURN`、`TREASURE`、`DAMAGE` 等。`PROGRAM` 已經收掉（spec 1154）：
  打通關現在看得到原作那段五頁的結局過場，先前直接跳到存檔詢問。其中 `PRINT RETURN`（120 條）已逐條讀完
  （spec 1147）：它是硬換行，連續兩條會空一行——缺口在 UI 的行模型，不是 ECL VM。
  `PICTURE`（199 條）已收掉（spec 1148）：`0FFh` 是關閉，關閉訊號接進畫面、`4FBAh`／`4FBBh` 的不重繪旁路也建了模型。
  `COMBAT`（199 條）的三選一順序已照抄原作（spec 1149）：場上有怪就直接打，
  商店旗標排在後面——remake 原本反過來。`CALL`（168 條）七支分派逐條讀完
  （spec 1150）：corpus 用到四個運算元，`6803h` 是圖片序列的下一格、`B200h`
  的第二個音效走不到。`TREASURE`（63 條）也讀完了（spec 1151）：戰利品池是覆寫、
  物品鏈是前插所以清單反序，隨機表的區間 bug 已修，隨機那一路也接上了原作的
  造物品常式（spec 1036），開出來的東西帶得出加值與名稱修飾。`DAMAGE`（24 處）也讀完了（spec 1152）：旗標最高位元清空時
  整個 byte 是「連打幾下」，每下隨機挑人擲命中——正式路徑先前只結算全隊那一種，
  現在三種目標形式都算得出來。`CLEARMONSTERS`（206 條）已逐條讀完
  並跟上（spec 1145），只剩 `7603h := 8` 的語意未解讀。每一支 handler 的位址與
  條數在 [`docs/audit/ecl-opcode-handlers-dos.md`](docs/audit/ecl-opcode-handlers-dos.md)。
- **存檔**：角色記錄 422 bytes 已全部 decoded 或 documented、unknown 0；跨遊戲角色轉移未做。
  存讀檔本身已有「讀回來還走得動」的閘（23 份段界快照逐份讀回來真的走一步），
  ⚠ 只比欄位抓不到這一類缺口——存檔沒保存的執行期上下文在讀回來的那一份是零值，
  欄位全對，玩家按下一步才卡住。
- **表現層與音訊**：畫面逐張對照、每個場景與戰鬥 phase 的音效綁定。
  第一人稱已有全作逐格對照（585 種牆面配置全覆蓋、1,498 張，spec 1185），天空色也包含在
  比對範圍內：它由該段 ECL 的載入時常數寫入決定（`eclBlockLoadTimeWrites`），
  pack 的宣告值只是沒有存檔時的底值、隨後會被同一組寫入蓋掉。SPRIT 畫布相對於戰場格的錨點仍未量（現在只在沒有
  CPIC 時才會走到那條路）。
- **戰鬥佈陣**：原作演算法已解讀並接進 `StartCombat`（spec 1200）——佈署
  模板是「每列欄範圍」表（5 種形狀）、兩隊起點＝(0,0) 與距離×方向 delta、
  找空位是理想格為中心的狀態機掃描，`internal/combat.Deployment` 忠實轉錄並有
  測試；正式遭遇會帶入地圖朝向、遭遇距離、地面與地城牆面。仍須以更多原版
  同狀態戰鬥畫面驗證所有方向、距離與大型 footprint。
- **按鍵重放**：單一連續 session 已通關；見 `docs/audit/key-driven-session.json`。
- **三平台發行包與推廣片**：執行 `tools/package-release.sh <版本>` 與
  `tools/build-promo.sh <版本>`，現行交付物一律集中在 `dist-all/<版本>/`。
  `patch/` 不含原版 ZIP 與未授權 PC-98 音樂；`full-local/` 只供本機驗收。
  Linux AppImage 與 Windows EXE（Wine）已啟動與繁中截圖驗收；兩張開場 PNG
  逐位元組相同。Windows／macOS 仍不宣稱已真機驗收。

目前交付狀態與建議順序見[完成度自評](docs/knowledge/remake-completeness-assessment.md)；
[歷史缺口盤點](docs/knowledge/coab-remake-todo.md)只供追溯。

---

原始遊戲、手冊、圖片與音樂的權利仍屬各自權利人；本專案保存研究證據與新撰寫的
繁中資料，不重新散布未授權的原始遊戲映像。
