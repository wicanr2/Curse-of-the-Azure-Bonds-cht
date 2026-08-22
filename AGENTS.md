# Codex working agreement — Curse of the Azure Bonds 中文化／Remake

本檔是 compact／交接後的第一閱讀入口，也是 agent 工作規則的單一權威來源。
它已融合 `CLAUDE.md` 中仍有效的原始需求；後者只保留人類可讀的目標與資料
索引，不應再複製易過期的 checkpoint。現況與真相來源優先序在 `CONTEXT.md`
（2026-08-13 已瘦身，歷史逐行分冊到 `docs/context/`）；全遊戲 RE 與
重建完整度的單一權威入口是
`docs/knowledge/coab-re-coverage-matrix.md`（系統層）與
`docs/audit/coab-function-index.md`（模組／函式層），逐輪成果在
`docs/project-status.md`。
若文件衝突，以目前 worktree、原始 bytes／實機
證據與較新的 READY spec 為準，並主動修掉被推翻的舊斷言。

## 1. 不可縮減的最終目標

完成 SSI《Curse of the Azure Bonds》（青色枷的詛咒）的乾淨重製與繁體
中文化，而不是只做畫面 demo：

- Go／Ebiten 跨平台 remake，可由開場玩到結局。
- 還原 Gold Box ECL bytecode、DAX/GEO/SAVE/圖像、AD&D 規則、戰鬥、
  第一人稱地圖、AREA map、世界旅行、音樂與音效。
- UI、事件、Journal、Tavern Tales、物品、法術與規則說明完整繁中。
- 原作要求查手冊的 Journal 內容整合進遊戲，仍保留中文手冊與攻略 Markdown。
- 建立可被其他 SSI Gold Box 遊戲沿用的獨立 engine 與知識庫。
- 原版忠實模式是驗收基準；日後可提供基於原素材的獨立美化 theme。

Prototype、單一 vertical slice、測試通過或幾張截圖都不等於完成。只有完整
玩家路徑與逐項完成稽核能支持「完成」聲明。

## 2. Repository 架構

這個 workspace 包含兩個獨立 Git repository：

- `Curse-of-the-Azure-Bonds-cht`（目前根目錄）：CoAB game pack、翻譯、
  原作證據、轉換資產、手冊、攻略、截圖與 integration。
- `golden-box-remake-engine/`：作品中立的 ECL/DAX/GEO codec、JSON schema、
  rules、renderer contracts 與可重用 VM/runtime。

界線：

- 劇情旗標、章節／block、地圖座標、遭遇組成、NPC 離隊、英文文字比對與
  翻譯只能放 CoAB 版本化 JSON/game-pack。
- Go engine 只放可重用機制、typed adapter 與有證據的格式語意。
- 若 game pack 無法描述某行為，先擴充 engine schema/runtime，再用 CoAB
  JSON 宣告；不能為趕進度 hardcode 本作情節。
- 不得把 nested engine 加成 CoAB gitlink，也不得把 engine source 複製進來。

## 2.5 反組譯的驗收口徑（2026-08-13 使用者決定）

先前的做法是「遇到問題才反組譯那一段」，結果是 64 支一次性 audit 腳本、
沒有任何全模組清冊，也答不出「還剩多少沒看」。現在改成**先徹底盤點、
再逐系統語意閉合**：

- **全函式覆蓋台帳**：DOS 與 PC-98 的每一個函式都必須在
  [`docs/audit/coab-function-index.md`](docs/audit/coab-function-index.md)
  裡有狀態，三選一：
  - `已解讀`：附推論等級（`exact`／`strong inference`／`hypothesis`）與引用規格。
  - `不阻塞`：附**具體理由**（如 Turbo Pascal RTL、字串格式化、heap 配置器），
    不是「看起來不重要」。判定要能被下一個人推翻。
  - `待解讀`：預設值。
- **兩平台都全掃**：同一機制在 DOS 與 PC-98 的證據分開記錄。PC-98 的 Borland
  符號名可作候選語意，但不自動成為 DOS 的事實。
  ⚠ **2026-08-16 使用者收斂 PC-98 的角色**：PC-98 是拿來反組譯 game engine 用的
  參考，**與 remake game engine 無關的部分不必解讀**。全掃（R1 盤點）仍然要做，
  但 PC-98 的語意閉合（R2）只做 remake 需要的那些；不再以「PC-98 全模組語意閉合」
  為目標。
- **不再以「這一輪玩家路徑用不到」當作跳過盤點的理由**。盤點（R1）與語意
  閉合（R2）是兩件事：盤點必須完整，語意閉合才依玩家可見性排序。
- 台帳與 IDA 匯出都必須可重生：換一台機器重跑 `tools/re-sweep.sh` 要得到
  同樣的函式數與雜湊；差異即代表流程有未記錄的手動步驟。

## 3. Spec-driven development 與證據標準

固定順序：

1. 反組譯／實機觀察／原始資料解碼。
2. 把 bytes、地址、trace、截圖、來源與 confidence 寫入 `docs/spec/`。
3. 清掉被新證據推翻的舊斷言。
4. 規格明確標為 `READY`。
5. 才實作、測試、走玩家路徑並更新知識庫。

證據優先級：

- 本機原始 executable/runtime capture、原始 bytes 與可重現 trace。
- IDA/decompiler 結果必須與 bytes、runtime 或另一權威來源交叉驗證。
- 使用者提供的 manual、Adventurer's Journal、Clue Book PDF。
- 公開原版畫面可作 layout／視覺參考，但必須記錄平台與來源。

畫面與結論要標示：

- `exact`：同一平台／狀態的逐像素或逐 byte 證據。
- `nearby`：同平台相鄰狀態，只支持共用部分。
- `material-exact/layout-reconstructed`：真實素材依量測幾何重組。
- `layout-only`：只能證明區塊與比例。
- `hypothesis`：待 runtime/bytes 驗證，不可寫成完成事實。

### 容易重犯的反組譯與實作錯誤

- **手打的「位址／編號 → 名字」對照表會安靜地漏格。** 漏掉不報錯：報表只是把
  那一格印成「符號表沒有」，讀起來像**原作真的少了那個名字**。第 700 輪的
  `SOUNDFX` 描述子表漏了 `4840h` 與 `4842h`（MISSFX／SPELLHITFX），下游的音效
  對照報表因此少了兩列——正好是 remake 用得最兇的 `SoundSpellHit`（29 處）與
  `SoundMiss`（2 處），**那兩個事件從來沒被比對過**。凡是能從既有結構推出來的
  對照（基底 ＋ 步長、既有的選擇子表）就**用推的**；真的只能手打時，要配一條
  「每一個宣告過的東西都必須有一列，否則讓它紅」的護欄。

- **far-call 對照表是下界，不是全集。** 它只收得到 IDA 認成程式碼的呼叫點，
  而且**不涵蓋常駐**。同一個問題實測：表列 36 處、位元組直掃 54 處。差的 18 處
  裡有一處會**改結論**——`LIGHTNINGFX` 在表裡是 0 處，於是報表寫成「remake 有、
  原版沒有」，而它其實在 `CASTSPELL` 裡。**要下「原版沒有呼叫 X」的結論，先直掃
  位元組，並確認表是直掃結果的真子集**（不是子集就代表直掃漏了，那才要查）。
  直掃的相對風險是假陽性，用「前面推得出已知引數」與那條子集檢查壓住。

- **`all_contains` 規則很容易要求到「隔壁訊息」的字。**
  `tilverton.sewers.spoils-and-master` 列了三個片段，其中 `A SOUND, SUDDENLY CUT OFF`
  其實是**另一條 ECL 字串**。兩句合併出現時規則對得上，單獨出現時**永遠對不上**
  ——玩家就看到一整句英文。譯文一直都在，壞的是規則。
  ⚠ 這種洞**逐頁的覆蓋率結構上看不到**：一頁只要曾經被某一條命中的 run 經過就算
  接上，而玩家走的是**一條** run，不是所有 run 的聯集。
  ⇒ 現在 `ecl-text-coverage` 另外以 **run** 計一份帳（`runs_orphan`，spec 1190），
  這一類洞在報表上看得見了，不必再靠實跑撞到。**那個 0 撐得住**：拿掉任一條規則
  它立刻變成非 0 並指名是哪一份文字，而逐頁的 `unmatched` 每一種情況下都還是 0。
  ⇒ 寫 `all_contains` 時每個片段都要問「這句話真的一定和其他片段同時出現嗎」。
  ⇒ 修的時候**加一條更寬鬆的規則放在原規則之後**（`MatchText` 是 first-match-wins），
  不要改原規則——原規則還要接合併出現的那一版。
  ⇒ **怎麼分辨「規則寫錯」與「本來就是片段」：看兄弟句。** 巫師塔那三句遭遇通報裡，
  `A SULFEROUS SMELL…` 與 `A PARTY OF DARK ELVES…` 各有一條獨立規則，就
  `AN EFREET LEADS A BAND…` 被綁進別條 ⇒ 那是規則寫錯。反過來，問句片段
  （`DOES ANYONE WANT TO…`）的兄弟句**全都是併在前一頁的規則裡**，那就是片段，
  不要替它寫規則。

- **取樣密度也是覆蓋率的一部分。** 「實跑踏到」本來只採樣**觀測迴圈**，走過去又
  馬上被劇情推走的格子從來沒被記到。`COAB_DECISION_LOG` 錄的是 `MoveDungeon` 的
  **每一次**呼叫，`cmd/route-cells` 把它換算成地形碼併進去 ⇒ 實跑 191 → 198、
  聯集 210 → 214、未達成 48 → 44。**同一條路跑同一次，只是記得比較細**（spec 1193）。

- **重放跟丟時要「用過就記」地重新對齊。** 拿「現在在哪一段的哪一格」去整條路線
  裡找同一個起點——⚠ **一定要帶段號**，不同段都有 (7,13)，只比座標會對齊到不相干
  的地方。⚠ 對齊時**把指標跳過去**會丟掉中間的步驟（76→41 格）；**指標不動**會在
  同一格重複對齊到同一步而原地震盪（照路線 759 次只走 57 格）；**用過就記**才對
  （82 格／123 步）。⚠ 代價是話少了 6 句——**兩邊都不是免費的**（spec 1191）。

- **路線知識要用錄的，不要手抄。** `State.Select` 與 `MoveDungeon` 是所有玩家層級
  動作的唯一入口，主線跑一次就錄得到 1607 步（`COAB_DECISION_LOG`），重放端用
  按鍵重放（`COAB_ROUTE_JSON`）。⚠ 錄**選項文字**與**起點座標＋方向**，不是索引
  與終點——索引會隨選單內容錯位，而樓梯事件是「站對方向踏上去」才觸發的。
  ⚠ **只錄選單不夠**：783 個選擇只用得到 13 步，因為決策點要**走到那裡**才會出現。
  補上移動之後 28 → 76 格、37 → 63 句、第一次走出開場那一段（spec 1191）。

- **調參數之前先確認參數是不是限制。** 按鍵那一場停在 28 格／37 句：幀數
  600→4000 不變、卡住門檻 6→3→2 不變、改成輪流選**變差**。⇒ 限制不在走法也不在
  選單策略上，缺的是**路線知識**。三個方向都試過而且都不動，才有資格說
  「不是這裡的問題」——只試一個就換方向會把時間花在調沒有用的旋鈕上（spec 1191）。

- **「多策略取聯集」不能搬到連續 session 上。** 冷走那邊有效是因為**每一種策略
  各跑一趟再取聯集**；按鍵那一場是**一條連續的 session**，換選項不是多一條路，
  是**換**一條路——實測輪流選選項讓走過的格子 28 → 27、記到的話 23 → 16
  （輪到「離開」就真的離開了）。**同一個技巧在兩種量測上效果相反**（spec 1191）。

- **讀完存檔要通知前端換地圖。** `LoadPartyFile` 設了 `GeoMapSet`／`GeoMapBlock`，
  但前端的 `geoGrid` 只有收到 `ConsumeGeoMapRequest` 才會換 ⇒ 少設 `geoMapPending`
  的話，讀檔之後前端拿的還是**上一段的地圖**，而且**不會有任何錯誤訊息**；
  前端一開始沒有地圖時更直接——**按什麼都沒反應**（114 份主線快照裡 104 份是這樣）。
  ⚠ 判「沒反應」的簽章**要帶朝向**：面對牆按 `M` 只轉身，少了朝向會把最常見的
  那一步判成沒反應。⚠ 測試自己少載 `ITEM*.DAX` 也會像輸入層壞掉（spec 1191）。

- **分母漏段會安靜地發生。** `cmd/cell-sweep` 每一段冷開一支隊伍走進去，而有些段
  的入口是**一整段過場**（巫師塔、`ECL4/0x25` 魔法商店），冷走完會停在世界地圖
  ⇒ 那些段整段被記成「進不去」，**格子從來不在分母裡**——而玩家真的走得到。
  ⇒ 冷進不去時改從**主線的段內快照**進（`-snapshots`）：分母 250 → 259、
  聯集 195 → 204。⚠ 那些段的隊伍帶著主線的旗標，報表要標明，別讀成「冷開就走得到」。
  ⚠ `blockSweep.note` 的語意是「**這一段掃不了**」，渲染時看到它就整段跳過——
  拿它放一般註記會讓掃得好好的段從報表消失，**數字安靜地少掉而且不報錯**（spec 1193）。

- **走訪的邊界要記「(格子, 進入方向)」，不是只記格子。**
  樓梯事件是**站對方向踏上去**才觸發的，朝向不對就 `EXIT`、畫面上什麼都不會發生
  （spec 1161）。只記格子的話，第一次從錯的方向踏上去就把那一格封死 ⇒ 靠樓梯才
  進得去的樓層永遠走不到，而報表只顯示「那幾格走不到」，**看不出是走法的問題**。
  ⇒ 修正後巫師塔從 15 格／2 個地形碼變成 207 格／22 個；⚠ 但**聯集不動**——
  多出來的落在不在分母裡的段。修正是對的和指標會動是兩件事（spec 1193）。

- **「直覺上的成因」要拿工具驗，不要拿它當結論。**
  走訪未達成最大的一類是「站得上去但從段入口走不到」，直覺成因是門要劇情旗標。
  `cmd/campaign-snapshot-walk` 拿主線快照帶著旗標重走 ⇒ **只有 4 個**是非得帶旗標
  才走得到的。剩下的是幾何：房間要靠樓梯／傳送事件才進得去，而事件要在**劇情的
  正確位置**才觸發——不是「有沒有旗標」，是「當下在不在那個位置」。
  ⚠ 第一版用**段界**快照，23 份有 7 份讀回來停在荒野 ⇒ 結論看起來一樣但**證據
  不夠**（走不了的那幾段根本沒被驗到）。改存**段內**快照（12 個 block 全涵蓋、
  每一份都走得動）之後結論才站得住。**先確認證據涵蓋到爭議的那一部分**（spec 1193）。
  ⚠ 讀存檔取段號要跟 **ECL session** 拿（`State.CurrentECLBlockID`）：
  `Area.LastECLBlockID` 遊戲跑的時候從來沒有人寫，讀出來是 0，而 0 是合法的 block
  編號 ⇒ 查下去拿到完全不相干的地圖而且不報錯。

- **列舉不到的東西，用走的走得到。**
  酒館傳聞的編號來自 `7F79h`，那一格被許多不相干的流程寫過 ⇒ 靜態推不出可能的
  編號集合，於是它被歸在「還沒接、要先追出寫入來源」。改用**多種選單策略走實跑
  路線**之後，玩家真的看得到的 13 則直接被走了出來，逐則補規則就結案了。
  ⇒ 卡在「先解出來源才能列舉」時，先問一句**能不能走到**。

- **走訪類的自動走法要跑好幾種策略再取聯集。**
  單一策略的結果**看起來都很合理**，但每一種都會被某一類岔路擋住：選單挑第一項
  被收費關卡擋在門外、挑最後一項在「要離開嗎」直接走人、**踏上去被事件搬走時
  不從落點繼續**等於把「走上樓」記成「走到樓梯口」。
  ⚠ 而且**單獨換一種可能更差**：跟傳送讓 `ECL6/0x43` 多走 44 格，卻因為走訪順序
  改變讓 once-only 事件換地方觸發，索引反而少兩個 ⇒ **取聯集，不是換一種**。
  ⇒ 加策略要加到**沒有新增**為止（第 4 項就是 0），不是加到覺得夠了為止（spec 1193）。

- **未達成要拆成成因，不然只是一個數字。**
  走訪的 72 個未達成拆開來是：站得上去但從入口走不到 71、四面都不通 1、
  分母裡的死索引 0。三種的處置完全不同（補路線／補不了／不用補），而**死索引是 0**
  這件事本來只能猜——它成立才表示未達成全部都是真的待辦，不是分母灌水。
  ⇒ 另外算「**走路的上限**」（站得上去的索引，249）：那才是天花板，
  把分派索引總數 250 當目標會讓永遠差的那一格看起來像還有事情可做。
  ⚠ 分類只看幾何（`CanMoveDungeon`）不跑 ECL——走過去又被劇情推回來那種不算
  「走不到」，那是內容不是幾何（spec 1193）。

- **「發得出來」與「有東西會發」要分兩欄問。**
  規則層有某個動作是**能力**，現在有哪一段劇情會發是**有沒有用到**。混成一格會讓
  「寫了但沒人呼叫」看起來像做完了。音訊生命週期就卡在這裡：`stop` 的程式碼在，
  但 game-pack **表達不出來**——共用 engine 的驗證擋掉 `track_id` 空的 binding。
  ⇒ 「沒人寫」補一行 JSON 就好，「寫不出來」要動共用 repo，**兩種卡點的處置差很多**，
  報表要寫的是哪一種，而且那句話要**跑出來的**不是手打的（spec 1192）。

- **前端讀鍵盤只能走 `a.justPressed`／`a.keyDown`。**
  直接呼叫 `inpututil.IsKeyJustPressed`／`ebiten.IsKeyPressed` 會被
  `TestFrontendReadsKeysOnlyThroughTheSeam` 擋下來。理由不是整潔：繞過接縫的
  那一行在按鍵驅動的測試裡**永遠不會觸發**，而測試會照樣綠——它根本走不到那裡。
  ⇒ 按鍵驅動的一場是 `TestKeysDriveARealSessionFromTheTitle`（spec 1191），
  要出報表就設 `COAB_KEY_SESSION_JSON=docs/audit/key-driven-session.json`。

- **「沒到過」有兩種成因，而它們的處置完全相反。**
  可達性報表說主線實跑一格都沒踏到提爾佛頓那三段。那可能是**主線不經過那裡**
  （路線的選擇，沒事）或**那些格子走不進去**（缺陷，玩家會撞到）——而逐格實測
  分不出來，因為它把隊伍**直接放**到目標格上。加一支只用 `CanMoveDungeon`
  從入口廣度優先走的探針就分得開：冷走走得到 25 個索引 ⇒ 是前者。
  ⚠ 而且兩把尺**互不涵蓋**：主線有劇情旗標、開得了冷走打不開的門；冷走沒有劇情
  擋路、走得到主線繞過的地方。**只報其中一個都會低估**（32% vs 聯集 70%）。
  ⇒ 報「覆蓋率」之前先問：**沒被涵蓋的那些，是同一種原因嗎？**

- **覆蓋率的分母不要自己重算——讓產生它的那一支輸出。**
  算「主線實跑踏到逐格實測的幾格」時，我自己掃全圖地形算分母得到 **299**，
  而逐格實測自己的數字是 **250**。差在四條規則：實測走的是 `dispatch.Indexes`
  而不是全圖地形、**跳過索引 0**、只算地圖上真的有那個地形碼的、而且「進不去」
  的段整段不計。自己重寫一遍就會漏掉其中幾條。
  ⚠ **兩個數都自洽**，放在一起才是假的——分子分母來自不同的尺，比值沒有意義。
  ⇒ 分母由**產生它的工具輸出成機器可讀檔**（`-index-json`），下游只消費。
  這樣「分母漂掉」在結構上不可能發生，而不是靠兩邊的作者記得同一套規則。

- **集合裡混兩種語意的字串，會讓每一格都多算一種——而數字看起來完全合理。**
  數「換種子會演出幾種不同敘述」時，我拿 `playerText()` 的結果當集合的第一個
  元素，後面每一次卻放 `state.Message`。`playerText` 在 Message 為空時退回
  `Prompt`，於是那些格子的集合裡永遠有一個「請按任意鍵繼續」佔位，**每一格都
  多一種**。3 種其實是 2 種。
  ⇒ 集合／字典的鍵**只能有一種來源**。要比較就從頭到尾同一個欄位，
  不要「這裡方便就用這個、那裡方便就用那個」。

- **有 fallback 的取值函式，拿來當量測指標會灌水。**
  `playerText()` 是「`Message` 沒有就退回 `Prompt`」——當顯示函式很合理。
  拿它去量「第二次踏上同一格有沒有新內容」，74 格被判成「演出別的字」；
  拆開來 **53 格是「請按任意鍵繼續」、12 格是撿寶物提示、5 格是地城 HUD**，
  真的有新敘述的只有 **3** 格。**分母灌水了近二十倍，而每一格都「確實不一樣」。**
  ⇒ 量測要挑**語意單一**的欄位（這裡是 `Message`），不要用為了顯示而設計的
  fallback 取值器。挑窄的那一個會低估——**低估是可接受的方向，灌水不是**。

- **假缺口比假零更難察覺——因為它會給你一份待辦清單。**
  做「戰役走的每一步玩家按得出來嗎」那支盤點時，硬缺口從 **29 → 28 → 7 → 2 → 1**，
  每一次都是**我自己的掃描面比實際窄**：只收 `CallExpr` 的 Fun（漏掉
  `a.combatAction(a.state.CombatAct)` 這種把方法當值傳的寫法）、可達閉包停在
  第一個 `state.X`（漏掉 `Select()` → `selectCamp()` → `AdvanceGameTimeHours()`）、
  把讀取器與啟動接線都算成「玩家按不出來」。
  ⚠ **中間那些數字每一個看起來都像真的結論**，而且都附得出理由——差別在於
  假零讓你少做事（會被別的證據撞到），假缺口讓你**多做事**（去補一個本來就
  存在的功能），而且不會有任何東西反駁你。
  ⇒ 盤點出一份「缺這些」的清單時，**先逐項確認前端是不是用別的名字／別的路徑
  做了同一件事**，再開始補。

- **「畫面上的字」這一類欄位，`*State` 層的測試結構上看不到。**
  端到端戰役測試從角色建立打到最終王，24 段全綠——但它讀完檔就直接做下一個動作，
  `Prompt`／`Choices`／`LocationName` 在被觀察到之前就被覆蓋掉了。把同一批狀態
  交給**真的前端**畫一張（`tools/campaign-frames.sh`），第一次跑就抓到兩個
  `LoadPartyFile` 的缺陷：讀回來顯示的是開場選單、地名永遠是建檔當下的那一個。
  ⇒ **有一類欄位只有畫出來才會被觀察到**；驗收「玩家路徑」不能只驗 State。
  ⚠ 同一個函式上面幾行就有一段 `eventReturnMode` 的註解，講的是同一類問題，
  當時只補了「按下一步會不會卡住」——**同一類缺口補了一半會看起來像補完了**。

- **報表裡寫死的「交叉印證」句子，會在資料變寬的那一刻變成假的。**
  `pc98-music-triggers` 印著「`MSCPLAY` 的呼叫點正好落在那五個改寫 `MUSICNO` 的
  overlay 上」——把掃描面從 far-call 表改成位元組直掃之後，`MSCPLAY` 多出
  `overlay-10`，那句話就不成立了，而**它是字串常數，不會有任何東西提醒你**。
  ⇒ 交叉印證的結論要**算出來再印**（重合幾個、哪幾個、有沒有只出現在單邊的），
  不要把當時的觀察寫成句子。同一條也適用「⚠ 這一支只認某某形狀」那種警告：
  形狀補上之後要跟著改，否則會擋住下一個人去看真正的答案。

- **沒有符號的那一版，用「形狀」認得出來——但要兩個獨立訊號。**
  DOS 沒有除錯符號，PC-98 有，兩版的 overlay 編號一致。把 PC-98 某個常式的
  **呼叫點分佈**（哪些 overlay 各叫幾次）當指紋去比 DOS 全部 far-call 目標：
  `SOUNDFX` 的分佈跨 8 個模組，第一名差 1、**第二名差 29**——差距本身就是證據。
  ⚠ 只有一個訊號不夠：資料表可以整張**平移一格**，位址仍然連續、名字仍然一一
  對應，光看表看不出來。所以第二個訊號（逐格分佈）必須獨立於第一個（表的位置），
  而且**每一個候選錨點都要試過並印出第二名的分數**。取「最低的那個」當基底會在
  「最低那格剛好沒有呼叫點」時整排偏掉，而且不會有任何徵兆。

- **從偏移量算出來的「陣列第 n 格」，不等於資料本身的分類欄位是 n。**
  spec 1120 由「角色記錄 `+17Dh`／`+181h` ＝ 裝備區塊第 11、12 格」推成
  「彈藥 ＝ 類別表**槽** 11 或 12」——兩個 11 是不同東西。這一款的槽 11／12 是
  **卷軸**，箭與弩矢在**槽 10**（而槽 10 還混著藥水、飾品）。錯誤結論寫進了
  `AmmunitionCount` 的挑選條件，於是它挑的是卷軸；`capByAmmunition` 把 0 當
  「不設限」，所以**不會噴錯、不會有測試變紅**，只是箭的數量從此不再限制射擊
  次數。⇒ **推出索引之後要拿資料回頭驗一次**：把那個索引的東西列出來看看是不是
  你以為的那類（`docs/audit/missile-sound-classes.md` 就是這樣列出來的）。
  同一份規格 §199 其實早就寫著「彈藥（slot 10）」——**兩句話互相矛盾了幾百輪
  都沒人發現，因為沒有任何東西會去比對同一份文件裡的兩句話。**

- **名字是提示，不是證據；語意要看呼叫端。** `MISSFX` 聽起來是近戰揮空，但它在
  整支執行檔裡**唯一**的呼叫點在 `TWINKLE` 裡，與 `SPELLHITFX` 共用同一個 `if`
  ⇒ 它是**法術**沒中的聲音。照名字接的話，會在原作從不出聲的時機（近戰揮空）
  放法術音效，而且測試全綠、玩起來也不會當——只是聲音一直是錯的。

- **用腳本改腳本時，字串取代一定要斷言命中**（`assert s.count(old)==1`）。
  取代失敗**不會有任何徵兆**：檔案還在、語法還對、跑起來也不報錯，只是它改的
  那件事沒發生。第 678 輪我以為給擷取腳本加上了換圖參數，實際上那次取代沒命中，
  於是接下來**五輪**拍的「別張圖」全都是提爾佛頓——檔名照著變數寫成 `geo4-b21-…`，
  內容是另一張圖。兩批共 95 張作廢。

- **兩個量測互相矛盾時，先確認兩者的輸入真的一樣。** 移動探測說「在下水道」、
  畫面比對說「是提爾佛頓」，我先去懷疑那把新做的尺，花了好幾輪；真正的差異在
  命令列——一支帶了 `-ecl-block`，另一支沒有。**先 diff 輸入，再懷疑方法。**

- **不要拿「畫面有幾種顏色」當「有沒有畫出來」的代理指標。** 一張正常的文字
  畫面只有三種顏色（底色＋白＋青），和一張全空的載入畫面**數字一模一樣**。
  第 673 輪我就是這樣把好的程式判成壞的：`-sewers` 明明正確畫出火刀 checkpoint
  的投降問句，卻因為「只有 3 種顏色」被寫成「不報錯但畫面空白」，還寫進了 spec，
  差一步就去修一支本來就好的程式。**代理指標要能分開你要分的那兩種情況**——
  分不開就不是指標，只是巧合。要判「畫出什麼」就去看**非底色像素落在哪些列**
  （文字畫面是散在幾十列的細線，第一人稱是連續整片），或直接把畫面兩兩相減。

- **腳本流程「停在事件上」不等於流程壞了。** 故事路徑的終點常常是一個問句或
  選單，那是**正確**行為；要的如果是畫面本身（例如第一人稱保真度比對），就另外
  加一支把狀態轉成該畫面的旗標（`-first-person`），不要去改故事流程。

- **測試要斷言身分，不要斷言顯示名。** `fighter.Name` 是**會被翻譯**的欄位，`SourceName` 才是原版名字。拿顯示名當識別碼的測試，在補上翻譯的那一天會整批變紅（本輪 5 個），而且它們本來就不是在測翻譯。同理適用任何「原始值 ＋ 顯示值」成對的欄位——斷言挑原始值那一個。

- **照 reference 實作填出來的對照表，要拿原始資料量過才算數。** 抄一份社群port 的陣列很快，而**索引基底差一格不會有任何錯誤訊息**——出來的名字仍然是合法的名字，只是換了一個。量法是把原始資料整批組回去比字串：對的基底會有壓倒性的相符率，錯的基底是 0。物品名稱那張表就是這樣抓到的（54 筆裡 `9Fh`之後整段偏 3，玩家看到的加值是錯的數字，spec 1178）。

- 已閉合的 `MON*SPC` 傷害能力也必須遵守 engine＋game-pack 分層：raw effect
  kind、damage flag 與 operation 由版本化 JSON 宣告，不能把 `0Ah／70h／87h`
  或中文名稱硬編進共用 engine、State 或測試 fixture。`half` 與 `immune` 必須
  是不同結果；只有完整清除傷害才可標 `Protected`。若 handler、save restore
  或法術 caller 尚未閉合，應保留 `unknown／hypothesis`，不可因同一 cold flag
  或最終 HP 結果相同就合併。參考 engine 知識庫
  `golden-box-remake-engine/docs/knowledge/golden-box-combat-damage-rules.md`。

- ECL 地城事件的 `7F81h` 是「本步已處理」guard，不是整張地圖或整個
  session 的永久完成旗標。玩家座標真正移到另一格時必須清除；原地重繪、
  開關選單、戰鬥 handoff 與戰後續跑則不能清除。若處理錯誤，常見症狀是
  第一個 terrain 事件完成後，後面所有 terrain 都像是失效。
- 不得讓每次 ECL invocation 都用同一個固定亂數種子重新起跑。這會使同一
  terrain 的隨機遭遇永遠得到同一結果，玩家來回踩格也不會改變。測試需要
  可重現性時，`BlockSession` 必須保存同一個由基底 seed 建立的持續 PRNG
  串流；不能每次改用 `seed+序號`，那會改變既有長路徑的原版遭遇結果。
  remake save v6 已以 seed＋底層 source draw count 保存這條持續串流；任何
  新 save path 都必須沿用同一 snapshot，不得只保存 seed。v1–v5 沒有該欄位，
  載入相容不等於恢復原亂數位置。這只證明 remake continuation，不能冒稱
  已還原 DOS／PC-98 原版 RNG 演算法。
- **量「玩家看到什麼」之前，先把上一個畫面清掉。** 從存檔／快照恢復出來的狀態身上
  還掛著存檔當時畫面上的字；不清 `Message`／`Prompt` 就跑下一步，會把**上一句**當成
  這一步演出來的。症狀是整批「都演得出來」而且演的是同一句——看起來像滿分，實際上
  一格都沒量到。（23 份段界快照的逐格取樣第一版就是 143／143 全綠。）
- **IDA 的 `var_N` 是「離 `bp` 幾個位元組」，不是參數或運算元的編號。** 一個
  `array[1..3] of byte` 存在 `[bp-3]`／`[bp-2]`／`[bp-1]`，IDA 會標成
  `var_3`／`var_2`／`var_1`——照著抄成 `o[3]`／`o[2]`／`o[1]` 會把順序**整個
  顛倒**。同一份規格已經因為這件事錯過兩次（`ROLLDICE` 的參數順序、`21h`／`37h`
  的三個運算元）。**轉錄前先看寫入端**：`mov [bp+di+var_4], dl` 配 `di = k`
  代表元素在 `bp-4+k`，也就是 `o[k]`。順序錯了往往仍然自洽——`21h` 的運算元是
  `<地圖> 02 FF`，反過來讀會得到「載入 3D 地圖那一支永遠不執行」，看起來像個發現。
- **排除條件也是斷言。** 抽取工具裡的 `if … : continue` 跟報表裡的結論一樣需要
  證據——濾掉的那一批不會出現在輸出裡，於是「那邊沒有東西」看起來像是資料的
  性質，其實是查詢自己有洞。`scripts/spell_damage_table.py` 的
  `or spell["placeholder"]` 讓法術主表 13 筆無名列從來沒被讀過，而那 13 筆正是
  充能物品的效果列（spec 1169）；排除條件的來源是另一份規格裡一句沒驗證過的
  「玩家取不到」。**寫下一個過濾器時，把它當成要交代出處的斷言。**
- **一支 handler 在下結論前先 `grep docs/spec/` 那個位址。** 這個專案已經有
  1,100 多份規格，同一支函式常常兩平台各有一份。重讀一次不是錯，但把既有結論
  當成新發現寫進 commit 就是錯的；而且既有那份往往就是判別誰對的第二個證人。
- **「還沒做」的標記要當成會過期的斷言，宣稱剩餘工作之前先去看程式碼。**
  這個 repo 的 `WORKLIST.md`／`CONTEXT.md`／`docs/project-status.md` 與各份規格
  同時記著「做了什麼」與「還缺什麼」，而**後者沒有任何東西會在做完的那一刻更新
  它**——完成的 commit 改的是程式碼與那一輪的紀錄，散在別處的現在式句子照樣留著。
  症狀是重複開工：手札 59 的原圖 renderer 在四個檔案裡寫著「尚未接入」，而
  `cmd/azure-bonds-game/journal_image.go` 連縮放平移都做完了；`+11Ch` 的
  「兩種讀法待對」在 spec 1164 自己的表格裡已標成「已對完」，欄位註解卻沒改。
  **接到「還剩什麼」這種任務時，第一步是抽查兩三條最具體的待辦，去 grep 對應的
  程式碼或資料。** 查到已完成的就當場改寫正文（不要在正文敘述當初怎麼寫錯），
  並在那一輪的紀錄裡說明。

- **量「有沒有發生」之前，先確認儀器裝在事情真的會發生的那一層。**
  查提爾佛頓側牆少畫兩格時，我把三處靜默 `continue` 全部加上輸出，重畫一次
  一次都沒觸發，於是判定「沒有磚被丟掉」。那三處全在 `BuildWallLayout`
  **回來之後**，而真正的丟棄在它**裡面**（`if ok { append }`，`ok=false` 的格子
  根本不會出現在回傳值裡）。量到的零是**沒有涵蓋到那條路徑**——與資料層的假零
  同形，只是這次的「掃描範圍」是程式的哪一層。**先問「如果它真的發生了，我這支
  儀器看得到嗎」，再相信那個零。**

- **「沒有數字」也是一種會寫錯的斷言。** 說某一類「還沒被量」跟說某一項
  「還沒做」一樣需要先去看：`docs/audit/` 現在有 60 幾份報表，而它們的檔名不
  一定看得出量的是什麼。第 665 輪把「全城市／全房間 coverage」列成沒有分母，
  下一輪才發現 `cmd/cell-sweep` 早就在量（14 個 block、250 個分派索引、演出來
  是中文 240、落回原文 0）。**列「缺口」之前先掃一遍 `docs/audit/README.md` 的
  目錄，再決定那一格填數字還是填「沒有數字」。**

- **一個「刻意吞錯」的地方要配一條測試，否則它是靜默失敗的溫床。**
  `loadJournalImage` 解不開圖就記成「這一則沒有圖」再照常顯示文字——那個決定是
  對的（缺圖不該讓遊戲停下來），代價是壞掉的資產**不會讓任何東西紅**。清單、
  locale、檔案存在檢查全部通過，玩家只是永遠看不到那張地圖。把吞錯那一側能發生
  的每種壞法都寫成資產層的閘（`TestJournalImagesDecode` 就是補這個洞）。

- **要問「誰跳到這個位址」，用 `cmd/ecl-window -into`，不要用 `-opcode 24 -operand`。**
  `-operand` 比對的是**指令運算元**，而 `ON GOTO`／`ON GOSUB` 的目的地在指令後面
  那張表裡，所以它看不到任何選單分派；`-into` 走 `ecl.TraceGraph`，兩種邊都跟。
  `ECL2/0x04:160Ch` 用 `-operand` 只算得到 3 個入口，`-into` 是 6 個，而**玩家實際
  會踩到的三個全在漏掉的那一半**。純位元組掃描目的地字（如 `0C 96`）則相反，會有
  假陽性。查到之後再用 `-table <位移>` 把那張表拆開，逐個索引對到目的地。
- **要看某個位移前後的程式碼，用 `cmd/ecl-window -at <位移> -before N -after M`。**
  `-opcode`／`-text` 是「找到哪裡」，`-at` 是「已經知道位址，把那一段印出來」——
  追 `RunResult.PC` 停在哪、或順著 `-into` 查到的來源往下讀時就是它。
  ⚠ `-at` 落在指令中間時會顯示**下一條**可達指令，兩者位移不同就代表那個位址
  不是指令起點（多半是別的 block 的程式碼，別照著讀）。
- **問「這一格是誰在用」用 `cmd/ecl-cell-refs -cell`／`-range`，而且要看寫入端與
  唯讀端的分佈。** 一格如果每個會讀它的段自己也寫，那就是段內暫存；出現只讀不寫
  的段，才是跨段交接。這個判準是 spec 1162 切開 `4C00h`..`4C0Fh` 的依據。
- **`go test` 會吞掉通過測試的 stdout。** 加了 `fmt.Printf` 探針卻「什麼都沒印」
  時，先加 `-v` 再下結論——失敗的測試會顯示輸出，通過的不會，所以同一個探針
  在測試修好之後就「消失」了，很容易誤判成沒有執行到。
- **要下「全 corpus 都沒有 X」這種全稱結論，block 清單就不能是手打的。**
  取 `docs/audit/ecl-effect-coverage.json` 該 opcode 的 `blocks` 欄逐一掃，
  掃完的處數要對得上同一筆的 `occurrences`。對不上就是漏了 block，不是資料有誤。
  （`37h` 的第一版普查漏掉 `ECL6/0x45`，22 vs 23——結論剛好沒變，但那是運氣。）
- **一份覆蓋稽核報「未接上 0」之前，先確認那個 opcode 在不在它的分母裡。**
  `cmd/ecl-text-coverage` 走的是「印字指令連成的一頁」，`29h ENCOUNTER MENU` 的旁白
  不在它收的指令裡——所以它的 0 對那批文字是**假零**，那些字連被數到都沒有。
  同理，比對多段文字時要**一句一句比**：把一條指令的三句合起來比，第一句的規則會
  把另外兩句蓋掉，缺口就消失了。
- **同一句話只能有一個來源。** 譯文走 game pack 的 text rule，就不要另外在程式碼裡
  留一條硬編分支；兩個來源並存時，改 pack 改不動它，而按 pack 判讀的工具會把它
  報成缺口。（`A DARK ELF PATROL ARRIVES` 是這樣被誤報的。）
- **「欄位對得上」與「還能繼續玩」是兩件事。** 存檔往返的測試如果只比欄位，漏掉的**不會有任何欄位比對抓得到**：存檔沒保存的執行期上下文（事件結束要回到哪裡、ECL 的續跑位置、地城的生命週期狀態）在讀回來的那一份是零值，欄位當然全對，玩家按下一步才卡住。**存檔類的閘門一律要在讀回來之後真的走一步。**（`eventReturnMode` 是這樣被抓到的：存在事件畫面上的檔讀回來按下一步噴 `event has no continuation`，23 份段界快照裡只有 1 份踩得到。）
- **走訪類的閘門，斷言寫太嚴跟寫太鬆一樣糟。** 「讀回來走一步不可以換 ECL block」看起來像個合理的不變量，實際上世界地圖那幾段的下一步就是出發旅行，換 block 是對的。下不變量之前先確認它在**每一種**被涵蓋的狀態下都成立，不然真缺口會被兩三筆假紅淹掉。
- **位址相近不代表同屬一支函式。** 取字串、常數或表格時要先確認函式邊界；
  緊接在某支函式結尾之後的字串通常屬於下一支。誤認的後果是替一支函式編出
  一組它從來沒用過的選項（spec 611 的 `~HAUGHTY …` 是這樣被誤掛的）。
- **「表有幾筆」一定要有量測依據。** 「掃到哪裡算哪裡」會得到一個看起來合理、
  實際上短掉的長度；有效的反推方式是找程式碼裡實際出現的索引上界
  （spec 815 用 `s = 1..100` 戳破了一份 56 筆的表）。
- **把陣列原封不動寫回，不等於沒有副作用。** 要看讀出與寫回之間有沒有被改過；
  只看「讀出來→寫回去」的形狀會把推進遊戲時鐘誤讀成純視覺效果（spec 790／793）。
- 靜態分析要把**終止型指令**排除在 fallthrough 之外，判準是 handler 有沒有寫
  停止旗標或改寫 PC，**不是指令名稱看起來像不像結束**。漏掉一個終止指令，
  分析器會把它後面那段接上去，產生「自洽但根本不會執行」的序列；而那段通常
  另有入口（lifecycle entry、GOTO／GOSUB 目標），所以指令清單看起來一切正常，
  錯的只有次序。ECL 的 `20h NEWECL` 是這樣被漏掉的（spec 1104 §三）。
- **「清除」「重設」型 handler 的作用範圍要逐格列出寫入點，不能從名字推**，
  也不能沿用既有規格的一句話。最省事的作法是對全 overlay 掃寫入指令的位元組樣式
  （`C6 06 <位址> <值>`、`FE 06`／`FE 0E`），把某一格的**所有** writer 與 clearer
  一次列完再下判斷。名字叫 CLEARMONSTERS 不代表它清掉了 SETUP MONSTER 寫的東西
  （spec 1104 §九之二）。
- ECL 選單不是顯示完文字就算解碼完成。必須驗證動態長度字串的真正
  `stringsEnd`、selection destination、0-based／1-based 編號、branch table、
  resume PC 與選擇後第一個 opcode。若選 LOOT 卻落入「零怪物 COMBAT」等
  不合理狀態，先查選單續跑位址與分支解碼；不得用 frontend 特判掩蓋。
- 每個 ECL boundary（文字、按鍵、選單、字串輸入、戰鬥、寶物）都要保留
  同一 `BlockSession` 的 continuation。重新從 block 起點執行、錯用新的
  selection offset，或在戰後遺失 pending PC，可能重播前文、走錯選項或
  觸發不存在的戰鬥。
- 攻略只能協助定位與交叉驗證。攻略寫「33% 遭遇」「搜刮得到珠寶」時，
  仍須回到 ECL bytes、IDA 與 runtime 查出亂數比較、work address、寶物
  opcode 和旗標副作用；不得直接把攻略數字寫進 Go 程式冒充反組譯結果。
- 不得因目前 runtime 尚未支援某個 ECL opcode，就把它猜成最接近的既有
  UI。例：`INPUT STRING` 是可輸入、可退格、可確認且會寫入指定 ECL 字串
  記憶體的互動，不是預先列出答案的選單；若暫時無法實作，應以
  unsupported／pending 如實封鎖玩家路徑。
- 指令名稱、operand 數量或攻略敘述只能形成假說，不能單獨決定 runtime
  語意。至少要記錄 opcode bytes、operand decoding、目的位址、分支 trace，
  並以 IDA 與原版 runtime／另一項權威證據交叉驗證。
- ECL 對某個 work address 寫入 `02h`、`FEh` 等值，只能證明「事件寫了
  什麼」，不能單憑文字或攻略把該位址命名成命中、AC、年齡等規則欄位。
  必須再找到讀取端／consumer、角色資料投影或原版 runtime 效果；在此之前
  schema、spec 與測試都要標為 `hypothesis`，不得先把推測做成正式規則。
- 若某個 raw work address 只影響作品情節的轉移／一次性 gate，且沒有證據顯示它
  影響 AD&D 數值、戰鬥、存檔相容或玩家可見的規則結果，不得為了替它取名字而
  深挖完整 consumer。需要維持正常玩家路徑時，使用 engine 的中立 raw memory
  action，由 CoAB JSON 宣告 address／value／觸發條件；文件同時標示「原版語意
  未知」與「remake route contract 已驗證」。`0x4C00` 是本作目前的例子。
- IDA 沒有找到某個位址的 literal xref，不代表該欄位未使用；ECL work
  memory 常由基底指標、索引或通用 interpreter 間接存取。反過來，找到一個
  raw little-endian byte pattern 也不等於找到 consumer。必須區分 code
  operand、data coincidence、間接存取與真正執行 trace，避免 compact 後把
  掃描命中數誤讀成已完成語意。
- 相同的十六進位數字不代表相同物件。Borland symbol 的 `segment:offset`、
  overlay local offset、resident effective address、file offset、ECL work
  address 與角色／隊伍 record offset 是不同位址空間；例如 symbol 恰好帶有
  `:4CBB`，不能因此視為 ECL `PARTYBASE+01BBh` 的 consumer。每項命中都要先
  標明位址空間、載入／重定位方式與所屬 module，再追出實際讀寫資料流。
- ECL 欄位到戰鬥公式之間若經過複製、投影或暫存結構，找到公式端的加數仍
  不等於已證明來源。必須追到「事件寫入 → combat preparation／record copy
  → 公式讀取」的完整橋接，或取得等價 runtime trace；只找到兩端、但沒有
  中間資料流時，仍須標成 `hypothesis`，不得先新增正式 JSON modifier。
- 不得為了讓單一劇情點通過，在 frontend、State 或 VM 寫死本作密語、座標、
  怪物、文字、旗標或分支。互動機制放作品中立 runtime；`Krrkik` 等作品
  資料仍由原始 ECL 或 CoAB game-pack 驅動。
- probe 測試可暫時輸出未知分支，但不能被誤當 regression test 或完成證據。
  milestone 完成前要把 probe 轉成具名斷言，或在保留研究價值時移入明確的
  audit 工具；不得留下只會列印結果、沒有驗證條件的常態測試。尤其禁止
  commit `temporary probe`、刻意 `Fatalf` dump 或只為觀察狀態而永遠失敗的
  測試。
- direct-entry、直接設定 PC／旗標或注入戰鬥可以當**分段驗收的進入點**
  （使用者 2026-08-16 修正 rulebook 65：分階段驗收算數）。直入的入口一律走
  `cmd/azure-bonds-game -segment <id>`，註冊表在 `internal/segment`
  （`-segment list` 列出 25 段）；既有的專用旗標保留，但它們做的是**段內的
  檢查點**，不是段的入口。每一段用 debug
  進入點直入並各自對 reference 驗證無誤，即算該段完成；全部段落通過即算
  跑完，不必為了宣稱完成而跑一次連續全程。
  ⚠ 放寬的是驗收成本，不是證據標準：每段仍要對 reference 實測。
  ⚠ **段與段之間的狀態交接本身就是一段**（存檔、旗標、隊伍、pending ECL／
  combat transaction）。debug 旗標注入的是合成起始狀態，未必等於上一段真的
  跑出來的結束狀態；兩端都綠不等於接縫通過，接縫要自己列一段驗。
- 看見第一場戰鬥或第一段文字不代表事件完成。必須追蹤戰後 ECL continuation、
  後續戰鬥、旗標寫入、地圖持久狀態、Journal／獎勵及離開再進入的結果。
- 不得把解碼器「可以讀」誤寫成遊戲「已可玩」，也不得把靜態數學核心、短秒
  trace 或單張畫面擴大宣稱為 scheduler、完整動畫、音訊播放或完整玩家路徑。
- 若新 bytes、IDA 或 runtime 證據推翻舊文件，必須在同一 milestone 更新或
  supersede 舊 spec、README、狀態表與 `CONTEXT.md`；不能同時留下互相矛盾
  的「完成」敘述供 compact 後誤用。
- **相鄰的符號同時掃出零，先懷疑模型不是結論。** 音訊盤點把
  `SOUNDHALT`／`SOUNDOFF`／`SOUNDON` 當成三格狀態掃，三格都是 0 處寫入，
  再照「原作沒寫到就不算待辦」的規則印成三個 ✅ ——實際上那是 `SOUNDFX` 的
  **選擇子常數**，值在資料段裡就寫死，本來就不會有人寫它。**假零披上規則之後
  會變成假的完成**。判一個零是真的之前，先問「這個符號是我以為的那種東西嗎」。
- **報表把某件事標成「被別人擋住」時，先回去確認那件事在原作裡是不是那個形狀。**
  「音樂停不下來，因為共用 engine 不收空的 `track_id`」是個自洽又具體的卡點，
  唯一的問題是**原作沒有「這一段不放音樂」這種資料**：派曲常式查不到就 `ret`。
  停止來自玩家按鍵。卡在別人家的結論最不容易被複查——它聽起來已經有答案了。
- **讀完整支常式再下「這裡少了一段」的結論。** `SOUNDFX` 的早退條件分**兩處**，
  中間隔著別的分支；只讀第一處會得出「`CRASHFX` 被誤判成無聲」這種自洽的錯，
  然後動手「修好」它。函式的開頭不是函式。
- **先確認兩邊是同一個東西產的，再比數字。** 把快照上限調高之後走出來的組合
  反而變少（329 → 274）——調高沒用是個合理的解釋，但**變少**不合理。真正的原因
  是拿來比的舊目錄裡有 23 份現在的指令產不出來的快照。**不合理的方向本身就是
  訊號**：比較的兩邊先各自能重現，再談差異。
- **吃「目錄裡每一個檔」的工具，用法要寫「先清掉」。** 沿用舊目錄會把上一次
  （不同程式、不同測試組合）留下的產物一起算進來，而那個數字**比現在的程式真的
  產得出來還大**，看起來只是比較好看。
- **一份產物有兩個產生者時，用法要把兩個都寫上。** 少跑一支不會有錯誤訊息，
  只會少幾列——而「少幾列」和「那幾段本來就沒東西」長得一模一樣。
- **手寫的「還沒量」旁邊不要抄數字。** 沒有任何東西會在覆蓋率變動時更新那段文字；
  `cmd/remake-status` 的 `unmeasured` 就把 176／250 留了好幾輪。要寫就指回生成的
  那一列。
- **盤點要搭既有的瓶頸，不要另外假設一個。** 接縫盤點記段落轉移時，掛在
  `requestMusicIfBlockChanged` 上——那正是換段要重新選曲的地方。於是「漏掉某個
  轉移」與「那個轉移的音樂會錯」變成**同一件事**，兩個盤點共用一個訊號會一起紅。
  另外寫一個掃描器就是多一個會各自漂移的假設。
- **分類的順序要當成規格的一部分寫下來。** 接縫報表裡 `fresh-start` 必須排在
  `not-visited` 前面：宣告 `00h` 的段是全新開局，本來就沒有前一段，判成「主線
  沒走到」會讓開場那兩段看起來像沒驗到。**兩種都會印出「—」，肉眼分不出來。**
- **比較類的報表：失敗要有自己的欄位，而且不能算進分母。** 交接盤點第一版忘了
  「直入要先有隊伍」，於是每一段的 `SavePartyFile` 都失敗，而失敗路徑把「漏掉
  幾格」留在 0 ⇒ 報表印出「16 段比得成、漏掉 0 格」，**和「完全對得上」長得
  一模一樣**。救回來的是**逐列的備註欄**——每一列都寫著失敗原因。⇒ 產生比較
  報表時，一律留一欄印「這一列是怎麼算出來的／為什麼算不出來」。
- **報表要自己說清楚它是上界還是下界，以及為什麼。** 交接那份的差異包含「主線
  走得比較前面」與「直入只有一名隊員」，所以它是**上界**；「會讀的格子」來自
  靜態可達，所以是**下界**。同一張表裡兩個方向都有，不寫清楚就會被當成精確值。
- **比兩份資料之前，先確認它們是同一個定義域。** 到達取樣走 `MemorySnapshot()`
  （整份、含程式碼視窗與零），存檔那側走存檔編碼器（零不收、程式碼另存）。直接
  比的話差異從 1,132 變成 103,896——**而那個數字看起來一樣像個結果**，沒有任何
  東西會報錯。⇒ 兩側走**同一個存取器**，或明文寫一個 `normalise` 收斂定義域。
- **為盤點而存的中間產物，不要用「看起來像正式格式」的形狀。** 換段那一刻的記憶體
  取樣被寫成完整的 party file，於是很自然地被拿去餵按鍵測試——而那個取樣點在指令
  執行到一半的地方，載回來就 `operand 0 is truncated`。⇒ 只寫真正需要的欄位，
  給它自己的 `schema`，並在檔案裡寫一句「這不是存檔」。
- **修法和量測放同一份報表的前後兩欄。** 「這個修法有效」不該是敘述，該是同一支
  工具跑兩趟印出來的兩個數字（不鋪交接狀態 251 格 → 鋪上 39 格）。分開兩份產物
  或只留一段文字，下一次改動之後沒有人會發現它失效了。
- **殘差要能歸因，而且分類加起來要等於全部。** 補完之後剩下的 39 格裡，10 格是
  引擎進段時自己就要設的、18 格是由位置推出來的視圖暫存器——**那兩類本來就不該
  被蓋過去**，混在一起數會讓修法看起來比實際差。歸不了因的只有 11 格，那才是
  真正剩下的。測試要同時釘住「三類加總 ＝ 總數」，少一類代表分類漏了一種成因。
- **同一份文件裡的兩句話會互相矛盾，而且沒有東西會發現。** 狀態表的生成列寫著
  「DOS 沒有音樂」，手寫的「還沒量」那條卻寫著「DOS 版有沒有那 12 首曲子還沒查」。
  兩句並存了好幾輪。⇒ **同一個事實只留一處**，其餘指過去；手寫段落尤其不要複述
  生成列已經回答過的問題。
- **下「不存在」的結論要三件事：列舉而非搜尋、判準要寬、配正對照。** 「DOS 版沒有
  BGM」靠的是 image 的 94 個成員**逐一列舉**（不是 grep 沒中）、名字判準放到很寬
  （`MSC`／`MUS`／`SONG`／`BGM`／`FM`／`OPN`／`MID`／`SND` 任一命中就人工看）、
  而且同一份列舉找得到音效住的檔案——所以零命中不是因為看不見。**寬判準配零命中
  才有意義；窄判準零命中什麼都證明不了。**

## 4. 可用原始資料與工具

- DOS game image：`curseoftheazurebonds.zip`
- 傳統中文資料：`珍020-青色枷的詛咒.rar`
- 目前已版控的繁中遊玩摘要：
  `docs/manual/curse-of-the-azure-bonds-zh-TW.md`。注意：這份檔案是 remake
  操作／背景摘要，並非《軟體世界》中文說明書或全部冒險手札的逐頁轉錄；
  不得因檔名是 `manual` 就誤認其中已包含手札 59 地圖。
- 第 553 輪已另行重建冒險手札 1–59 的繁中校訂摘要入口：
  `docs/manual/adventurers-journal-zh-TW.md`。三份分冊保存條目編號、掃描頁、
  遊戲提示與 confidence；它們是摘要／重述，不是逐字翻印。手札 1 在現有掃描與
  英文 OCR 均未找到，維持 `unknown`。
- 第 554 輪已將同一新遊戲 session 由手札 59 延伸到 `(14,1,E)` LOOK 通路、
  Dexam 兩場決戰、戰利品／護符與戰後回洞穴。ECL4 `22h:+0B8Ah` 證明
  `4C03h>=1` 會略過 Dexam；前一 shrine block 的同位址值以 CoAB JSON 不透明
  `set_memory` 在 cave handoff 歸零，僅為 `strong inference` 的 map-local bank
  重建，不可命名成全域劇情或規則欄位。已揭露秘密 edge 必須雙向可走。
  第 555 輪再依中文手札 59 明示的神殿東門、GEO4/25 wrapped edge 與普通門路網，
  以 `(15,1,E)` LOOK 後十二步正常移動抵達 terrain `93h`／`(6,3)`，完成離場
  敘事並回到散提爾堡世界選單。東門仍是 `strong inference`，不可升格成原版
  SEARCH writer exact；洞內全支線與重訪仍未完成。**後續主線已經接上**：
  同一條 session 由此經立石群走到密斯卓諾，打完最終戰抵達結局選單
  （`TestRealNewGameRunsToTheEnding`）。
- 《軟體世界》中文說明書的原始掃描仍在
  `珍020-青色枷的詛咒.rar`；第 534 輪只把 `MOVEPARTY` 相關頁面與雜湊整理進
  `docs/spec/534-chinese-manual-moveparty-character-transfer.md`。使用者記憶中
  「完整中文手冊已轉成 Markdown」的產物目前不在本 repo／Git 歷史；若後續
  找回，應另以明確檔名納入 `docs/manual/`，不可覆蓋現有摘要。
- Manual：`Curse-of-the-Azure-Bonds_Manual_DOS_EN.pdf`
- Adventurer's Journal：
  `Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf`
- Clue Book：`Curse-of-the-Azure-Bonds_Misc_DOS_EN_Clue-Book.pdf`
- 工作資料：`workplace/`
- 英文 Adventurer's Journal 的既有 OCR：`workplace/journal-ocr-406.txt`；其中
  有手札 59 圖例文字，但 OCR 不保留可靠地圖幾何。中文掃描 `084.jpg`／印刷頁
  52 已確認保存完整手札 59 圖；可作 `layout-only` 拓撲證據，不能單獨換算 GEO
  座標、牆型或 `SEARCH／LOOK` 行為。
- IDA Pro：`/home/anr2/ida_94_official/dist`
- 倚天字型：`/home/anr2/cht/etan_font/stdfont.15`
- 倚天粗體參考：`/home/anr2/scummvm/monkey_island2`
- 其他 remake 參考：`/home/anr2/cht/daemon_winter`

允許使用 Docker 內 DOSBox／Xvfb 取得原版與 remake runtime 截圖。原版
executable 是行為 oracle；網路資料不能取代可取得的本機實機驗證。

### 原版 runtime oracle 的固定入口（第 572 輪起）

| 工具 | 用途 |
|---|---|
| `tools/dos-oracle-session.sh` | 容器常駐的互動式驅動：`start`／`key`／`type`／`shot`／`text`／`stop`（容器名固定 `coab-dos-oracle`）|
| `tools/dos_screen.py` | 把畫面上的 8×8 文字格解成字串，用來**看畫面決定下一鍵** |
| `tools/dos-oracle-capture-cell.sh` | 目前這一格拍四個朝向，順手消掉 `PRESS <ENTER>` 提示 |
| `tools/fp-oracle-compare.py` | 原版擷取與 remake 同格畫面逐格比對（EGA 量化後） |
| `cmd/dos-save-export -base` | 以原版自己寫的存檔為底，只覆寫座標與朝向，把隊伍放到任意格 |
| `cmd/geo-probe` | 印出一張 `GEO` 地圖每格的可走方向數，挑取樣格用 |

**定時送鍵的序列不要再加長延遲**：任何一步的載入時間漂移都會讓後面每一個鍵
錯位，而錯位的鍵可能剛好按到 `EXIT TO DOS`。走完整條選單路徑一律用互動式
session ＋ 讀畫面。詳見 [`docs/spec/1134-original-first-person-oracle.md`](docs/spec/1134-original-first-person-oracle.md)。

凡有反組譯（reverse engineering）需求，優先在 Docker 隔離環境內使用
`/home/anr2/ida_94_official/dist` 的 IDA Pro。其他反組譯器、反編譯器或
自製掃描工具只能作補充與交叉驗證，不得在 IDA Pro 可用時逕自取代主要分析
流程；IDA 結論仍須依本文件第 3 節，以原始 bytes、runtime trace 或另一項
權威證據交叉驗證。

### 反組譯工具鏈（第 559 輪起的固定入口）

| 工具 | 用途 |
|---|---|
| `tools/ida.sh` | IDA headless 唯一入口（固定 `ida-pro-9.4-idapython:py312-v1`） |
| `tools/go.sh` | Docker 內 Go 工具鏈（主機不裝 Go；`.git` 在 workplace 故帶 `-buildvcs=false`） |
| `tools/engine-bootstrap.sh` | 乾淨 clone 的第一步：把 `go.mod` 鎖住的 engine commit 準備好並重建 proxy |
| `tools/engine-proxy.sh` | 把 nested engine 的某個 commit 打包成檔案型 Go module proxy，供 `go get` 鎖版 |
| `tools/re-sweep.sh` | 一個平台的全模組建庫＋匯出：manifest → resident → 36 段 overlay |
| `tools/ida/export_module.py` | 匯出一個 database 的函式／xref／字串／segment／未定義區 |
| `tools/ida/analyze_overlay.py` | raw overlay 建庫：強制 16-bit、種子 entry point、再匯出 |
| `cmd/ovr-manifest` | TPOV 容器結構清冊（每段 code offset／size／entry stub／fixup） |
| `cmd/borland-symbols` | Borland 除錯符號 → JSON，並把 segment 對回 overlay module |

**raw overlay 不能直接丟進 IDA。** DOS `GAME.OVR` 整檔載入只會得到 **1 個
函式**（實測），因為 TPOV code 沒有 MZ header 也沒有 entry point。entry point
全部來自 resident 的 `CD 3F` five-byte stub（見下方 TPOV 段），必須先由
`cmd/ovr-manifest` 匯出再當種子；另外 unit 初始化固定在 code offset 0，
不在 stub 表內，要單獨補。忘記種子的症狀是「這段 overlay 好像沒有程式碼」。

**PC-98 `GAME.EXE` 帶完整 Borland 除錯符號**（1,725 symbols／53 modules／
1,531 types／641 members），符號位址的 segment 等於 overlay control segment、
offset 等於 overlay-local code offset，可直接對上 IDA 函式起點。DOS
`START.EXE` **沒有**這張表。因此：PC-98 是語意骨幹，DOS 是行為 oracle；把
PC-98 的名字搬到 DOS 之前必須另行證明結構對應，不得因偏移相近就套用。

### IDA 非破壞性語意註記

- 本節已參考 `/home/anr2/cht/大時代的故事/CLAUDE.md` 的實務經驗；來源檔只作
  唯讀參考並保留原位，不複製或反向改寫。吸收的是經本專案重新核對後可泛用的
  原則：「原始識別與位址永遠保留、語意只作外部附加、每筆語意必帶推論等級」；
  來源專案仍在驗證中的自動反向索引／語意 dump，不因此視為本專案已驗證工具。
- 開始解讀函式或欄位前，先以原始函式名、位址、offset、symbol 與已知 bytes
  搜尋 `docs/`、`CONTEXT.md`、既有 IDC 報告及 READY spec。搜尋零命中只能表示
  「目前索引沒有答案」，不能自動升格為 `unknown`，更不能據此重新猜一套語意；
  必須再查相鄰位址、基底加 offset、caller／consumer 與不同位址空間的對照。
- 原始 binary、overlay、symbol table 與基準 `.i64` 一律唯讀保存；每次分析先
  記錄輸入檔 SHA-256、檔案版本與載入位址，再複製到本輪 Docker 工作目錄操作。
  IDC、型別匯入、註解、重新分析或 IDA 自動升級 database 都只能作用於分析
  副本，不能原地改寫唯一的原始檔或基準 database。分析完成後，以可重現腳本、
  文字報告與規格保存結論，不以被修改過的 `.i64` 單獨充當證據。
- 若執行檔有壓縮、加殼或自解包 stub，必須先保存並雜湊原檔，再把解包結果視為
  獨立衍生物；記錄解包工具與版本、操作命令、輸入／輸出 SHA-256、載入基址及
  成功判定。不得以解包檔覆蓋原檔，也不得只因 IDA 顯示出較多函式就宣稱解包
  正確；至少要以入口控制流、原始 runtime 或另一項權威證據交叉驗證。
- `.i64`／`.idb`、平面 `.asm`、解包 binary、IDA 暫存檔與本輪可重建的 database
  都屬工作產物，預設只留在被忽略的 `workplace/` 或 `/tmp`，不得取代原始資料或
  規格進入版本控制。應版本化的是 IDC／typed resolver、輸入雜湊、可重現命令、
  外部語意 ledger、必要的連續位址報告與 READY spec；若例外保存 database，仍須
  同時保留產生流程，且不能把其中的 rename／comment 當成獨立證據。
- IDA database 是程式原貌與交叉參考的證據快取，不是把推測改寫成事實的
  地方。不得以推測名稱取代 `sub_XXXXX`、`word_XXXXX`、Borland symbol、
  原始 segment:offset 或 linear address；尤其不得因一次命名方便，就讓原始
  位址從後續輸出消失。
- 新語意只能「附加」，不能「覆蓋」。每筆註記至少同時保存：原始識別字／
  位址、所屬 binary／overlay／module、位址空間、候選語意、證據來源與推論
  等級。推論等級統一使用本專案的 `exact／strong inference／hypothesis／
  unknown`；若後續被推翻，刪除或 supersede 舊斷言並留下訂正依據。
- 上述四級門檻固定如下：`exact` 必須有原始 bytes 加可重現的 consumer／runtime
  trace，或兩項彼此獨立且能閉合資料流的第一級證據；`strong inference` 是多項
  證據一致、但仍缺一段 writer→projection→consumer 橋接；`hypothesis` 只能用來
  指引下一個 probe，不能進正式 schema／規則；`unknown` 表示目前證據不足，不能
  因名稱、常數相同、單一 xref 或反編譯器型別猜測而升級。每次升級或降級都要在
  文件中列出新增／失效的證據，不得只改標籤。
- 即使語意已達 `exact`，文件與工具輸出仍應採「原始位址 → 語意」並列，例如
  `ECL work 4CBBh → side attack-roll modifier`，不能只留下容易漂移的別名。
  結構基底加索引形成的 `[di+offset]`、`es:[di]` 等間接欄位尤其不能靠 IDA
  全域 rename 解決；應由外部 typed ledger／規格在不改原始表示的前提下註記。
- 可重現的函式 dump／xref 報告應在每筆組合語言旁附加外部 ledger 命中的候選
  語意、來源文件與推論等級，但不得把註記回寫成 IDA symbol rename。外部 ledger
  的每列同樣必須保留原始位址空間與等級；不能從 Markdown 自動抽取後，把
  `hypothesis` 與 `exact` 混成同一種可直接採信的名稱。
- `.asm` 是攤平文字，只適合定位候選，不具有 `.i64` 的 xref graph。查 caller、
  reader／writer 或全域 consumer 時，優先以 IDC 查 `.i64`，再回讀完整指令與
  raw bytes；不得以 grep 命中數冒充引用圖或資料流證據。
- 擷取反組譯片段時不得用負向過濾刪掉看似樣板的 `mov`、`add`、`shl`、
  register copy 或跳躍；這些指令常是 record base、索引、迴圈入口與間接欄位
  的唯一證據。需要縮短輸出時只能保留連續位址範圍，並在結論旁引用完整 dump。
- 每個 executable／overlay／module 都要先辨識編譯器、記憶體模型、呼叫慣例、
  參數順序、清理堆疊責任與字串表示，再解讀 caller。不得把 C 的右至左壓棧
  直覺套到 Pascal 的左至右壓棧與被呼叫者清棧，也不得把長度前綴字串誤讀成
  NUL 結尾字串；不同 binary 即使同屬 Borland 工具鏈，也不能未驗證就共用
  prototype。IDA 自動辨識的函式名與 prototype 是重要候選，但仍須以 call-site
  指令、stack delta、raw bytes 或 runtime 行為交叉驗證，並帶推論等級。
- IDC 對直接 data xref 的讀寫分類應使用 `XrefType()`（例如
  `dr_W／dr_R／dr_O`），不得用助憶碼字串、空格排版或固定運算元位置猜測。
  這項分類只能證明 IDA 已建立的直接 xref，不包含指標間接存取。
- IDA 的直接 xref 不涵蓋先取位址、再經暫存器或指標間接讀寫的路徑。若看到
  讀取很多、寫入異常少，必須追蹤 address-taken、register propagation、record
  base 與 runtime trace，不能直接判定欄位只有一個 writer。
- 分析 Borland overlay／TPOV 時，必須分開記錄 code bytes、relocation／fixup、
  overlay-local offset、resident segment 與 runtime far pointer。在重定位尚未解析前，
  表內 raw addend 只能標為「未解析候選」，不得直接命名成 handler 位址；
  相同數字也不得跨位址空間合併。
- 第 412 輪已證明本作 PC-98 resident control 是 `20h` header 加
  `CD 3F + handler-local u16 + flags` 五 byte entry stubs；code 後 relocation
  是嚴格遞增 `u16` fixup offsets。新分析應使用 typed resolver，且仍須同時
  證明 control segment；不得退回以 raw addend 或相同數字猜 handler。
- **IDAPython 可用，但只在 `ida-pro-9.4-idapython:py312-v1` 這顆 image 上**
  （2026-08-13 實測，取代先前「只驗證 IDC 可用」的舊斷言）。基底 image
  （`ver2`／`ver3`）跑 IDAPython 是**零輸出、零訊息**的靜默失敗，看起來與
  「腳本寫錯」「IDA 沒裝 Python」一模一樣；根因是缺 `libpython3.12t64`，以及
  `idapyswitch` 把選定 interpreter 寫進執行身分的 `$HOME/.idapro`。新腳本一律
  優先寫 IDAPython（有 `idautils`／`ida_funcs`／`ida_xref`），IDC 只當退路。
  一律經 `tools/ida.sh` 呼叫，不要手寫 `docker run`。
- headless 稽核腳本必須把結果寫入明確檔案並檢查內容，不能只依賴
  `Message()`／`print`／stdout 或 exit code 判定成功。**exit code 完全不可信**：
  同一種「沒有輸出」的失敗在不同 image 分別回 0 與 1。唯一可信訊號是輸出檔。
  IDAPython 腳本要把 traceback 也寫進 `<輸出路徑>.error.log`，否則失敗是靜默的。
- 被 import 的 IDA 腳本不得在 module 層執行 `main()`／`ida_pro.qexit()`：
  一次實際事故是 `analyze_overlay.py` import `export_module.py` 時觸發它的
  main，把 `sys.argv[1]`（overlay manifest）當輸出路徑覆寫掉。收工動作一律包在
  `if __name__ == "__main__":` 內。
- 來源專案仍在試驗的自動語意 dump／反向索引方法，不因看起來有用就升格為
  本專案硬規則。只有在本專案以實際誤判案例、round-trip 與持續回歸證明有效
  後，才可另行納入工具鏈；未驗證前只能是候選方法。
- 訂正既有語意時，要先引用舊結論原本依賴的證據，再說明新 bytes、xref、
  runtime trace 或 consumer 為何足以推翻它；不能只換一個較順眼的新名稱。
  被推翻的別名不得繼續留在目前索引中冒充並存答案，訂正歷史則移入明確的
  superseded／推翻紀錄，避免 compact 後再次採用。

## 5. 視覺、版面與中文字體

- 固定 logical canvas 為 640×480。
- 原始低解析圖採 nearest-neighbour 整數放大，不做模糊縮放或 AI 補點。
- DOS 320×200 是 frame、geometry、素材與行為 oracle。
- PC-9801 640×400 是 CJK 字級、行距與資訊密度的重要 oracle。
- 一般正文、roster、status、commands 使用倚天粗體 16×15；24px 只用於
  標題或明確強調。
- Adventure chrome 必須使用本機 DOS runtime 抽出的 cracked stone raster，
  不回退為 generic 灰框。
- 左上第一人稱／一般 PIC 事件圖依畫面 contract 使用 cover＋clip 填滿內格；
  HEAD／BODY 人物另依下列固定舞台 contract。
- HEAD／BODY 是分層素材；頭不能塞進胸口。戰鬥 CPIC sprite 依原生 24×24
  tile、footprint 與 anchor，不走一般圖片置中。
- HEAD／BODY 場景人物不能走 PIC／第一人稱場景的 `cover`：依 game-pack
  `presentation.scene_character` 的 native anchor／clip 整數放大，先畫人物、
  再覆蓋原版人物內框。場景與人物是兩種不同 renderer contract。
- 原版忠實 theme 永遠保留；美化只能是額外可切換 theme。

目前最新視覺規格：`docs/spec/391-dos-head-body-character-stage.md`；
全域石框與字級基線仍見
`docs/spec/348-original-dos-frame-pc98-type-density.md`。
舊 spec 329 的手繪 combat frame 已被 supersede；現況是 DOS 石框素材 exact、
combat layout reconstructed，尚未宣稱整張 combat frame pixel-exact。

## 6. 中文、長文與攻略

- Unicode／Big5 斷行不可切 byte；要依中文標點與語意分頁。
- 長文需可立即顯示整頁、調整逐字速度、清楚標示繼續與頁次。
- 選項與翻頁輸入必須分離，避免讀文時誤選。
- Journal、手札與重要劇情可在遊戲內重讀。
- 人名、地名、物品、法術採 stable ID 與一致詞彙表。
- Translation JSON 應保存原文、繁中、來源 block/offset、譯註與校對狀態。

攻略保留「像小說一樣閱讀」的魅力，分三層：

- 冒險紀行：流暢敘事、人物動機與世界背景。
- 逐區攻略：座標、事件條件、戰鬥與寶物。
- 無雷提示：只給方向與規則提醒。

詳細原則：`docs/knowledge/golden-box-remake-for-chinese-readers.md`。

## 7. Git 與提交紀律

### 測試資料與穩定 ID

- 產品層測試不得複製 JSON 內的翻譯、裝備名、法術名、地名或其他可編輯顯示
  文字作為期望值；必須用穩定 ID 從實際 game pack／catalog 取得期望內容。
- 選項翻譯同樣屬於 game pack 資料：使用 stable `option_rule.id` 與原始
  source token 對應 `message_id`，State 只呼叫作品中立的解析介面。不得因
  新增一個劇情選項就擴大 State 裡的中文 `switch`；舊 fallback 只能作遷移
  相容層，碰到相關 vertical slice 時要移入 JSON。
- 原版 ECL 固定英文、原始 bytes、位址與選單 token 只可出現在明確的來源
  oracle／parser 測試，並應同時驗證結構或來源位置；不能拿它們代替產品層
  本地化驗收。
- 測試應驗證「ID 解析、locale fallback、事件綁定與畫面取得同一份資料」，
  使 JSON 改譯文或裝備顯示名時不必同步修改另一份硬編碼測試字串。
- 若測試需要比對畫面文字，期望值也必須在測試執行時由同一份正式 JSON、
  stable ID 與 locale resolver 取得；不能先讀一次 JSON，再把當時結果貼成
  Go 字串常數。修改 JSON 顯示文字後，只應影響內容快照／翻譯審校，不應使
  驗證資料綁定與遊戲規則的測試失效。
- 常見錯誤：在測試直接寫「龍盔」「火球術」「長劍」等目前畫面文字，再用
  `Contains` 判斷。即使測試會通過，仍是在複製 JSON 的真相來源；應改查
  `message_id`／`item_id`／`spell_id`，並另測該 ID 經目前 locale 解析出的
  畫面內容。若 JSON 改名後測試必須手動改同一字串，表示測試分層有誤。
- 既有測試仍有歷史技術債：部分 `internal/game/*_test.go` 直接以繁中字串
  `Contains` 驗證。碰到相關功能時要逐步改成穩定 ID，不得照抄其模式新增
  債務；一次遷移一個真實玩家 vertical slice，並保留原始 ECL oracle 測試。
- 測試 fixture 不得另造一份會與正式 JSON 漂移的裝備、法術、怪物或翻譯
  catalog。除非測試目標就是 schema／parser 的最小合成資料，否則應載入
  版本化 game pack，再以 stable ID 找資料；合成 fixture 也只能斷言結構，
  不可冒充 CoAB 正式內容。

- 角色建立的 D&D saving throw 閾值是 game-pack 資料，不得在 Go 內按職業名稱
  推算或在測試複製顯示文字；使用 template stable ID 的 `saving_throws` 五欄，
  由 engine schema 驗證，State 只投影到角色 record。缺欄位應 fail-closed，不能
  用預設值掩蓋資料缺口。
- 玩家戰鬥法術的 CAST／Quick／pending 入口使用 `combat_player_spells` 的
  stable `id`／原始 `spell_id`／`target_mode`／`message_id`；
  `combat_ai_spells` 只負責 Quick metadata。新增法術時先更新 engine schema
  與 game-pack JSON，再接既有或已證據支持的 adapter behavior；不可把法術名、
  目標或劇情重新塞回 State，也不可把 `behavior` token 當成原版完整規則證明。

- 只有重大、已測試、可展示的 milestone 才集中 commit＋push；不要每個小改
  都提交。
- 兩個 repo 各自 commit／push，歷史保持獨立。
### 乾淨 clone 之後的第一步

`golden-box-remake-engine/` 與 `workplace/` 都在 `.gitignore` 裡，所以剛 clone
完的 CoAB **既沒有 engine 原始碼、也沒有檔案型 proxy**，直接 `tools/go.sh` 會卡在
取不到私有模組。先跑：

```sh
tools/engine-bootstrap.sh      # clone／fetch engine，打包 go.mod 鎖的那個 commit
tools/go.sh test ./...
```

它只認 `go.mod` 鎖住的版本，**不動 engine 的工作區、也不 checkout**，
所以開發者自己在 engine 上的改動不會被踩掉。engine 的 commit 還沒 push 上
GitHub 時它會明講並結束——那時要先去有那份 commit 的機器上推。

### 升級 engine 相依（私有 repo，容器沒有憑證）

`golden-box-remake-engine` 是**私有** repo：`proxy.golang.org` 取不到，容器裡也
沒有 GitHub 憑證（依規則不得放）。所以 `go get <module>@<commit>` 一定會失敗在
`could not read Username for 'https://github.com'`。固定流程是：

```sh
git -C golden-box-remake-engine push origin main   # 先推，zip 內容才對得上 remote
tools/engine-proxy.sh                              # 印出 pseudo-version
tools/go.sh get github.com/wicanr2/golden-box-remake-engine@<印出來的版本>
tools/go.sh test ./...                             # 確認沒有殘留 replace
```

`tools/go.sh` 已把該 proxy 排在 `GOPROXY` 最前面並關掉 sumdb（`GOSUMDB=off`；
go.sum 仍然逐版本鎖雜湊）。
⚠ **不要改用 `GOPRIVATE`**：它會強制走 direct，正好繞過這個 proxy。
⚠ **不要為了讓 CoAB 建得起來而在 `go.mod` 留 `replace`**：nested engine 不在 CoAB
版控裡，留著會讓乾淨 checkout 建不起來。

- CoAB 使用：
  `git --git-dir=workplace/azure-bonds-git --work-tree=.`
  （2026-08-13 由 `/tmp/azure-bonds-git` 搬入 repo 底下並列入 `.gitignore`；
  原位置一次重開機就會連同未推送的 commit 全部消失。根目錄的 `.git` 是
  root 擁有的空目錄，不要當成本 repo 的 git 目錄，也因此 Go 建置要帶
  `-buildvcs=false`。）
- Engine 使用：
  `git -C golden-box-remake-engine`
- 不丟棄使用者或不相關變更；先檢查 dirty worktree。
- compact 後若看到 probe、暫存 regression 或未完成 spec，先讀 diff 與
  `CONTEXT.md` 尾端；它們可能是正在累積的 milestone，不可因尚未提交而刪除。
- 每個 milestone 更新 README、`docs/project-status.md`、READY spec／知識庫
  與本檔或 `CONTEXT.md` 的延續資訊。

## 8. 驗證與完成門檻

開發中跑 focused tests，提交採風險分級，不得把時間浪費在每輪重跑無關套件：

- 純 locale fallback／測試資料分層且沒有改 typed behavior、schema、save、ECL
  continuation、renderer 或資產的低風險 milestone：跑受影響套件、audit、至少
  一條代表性正常玩家路徑；marker 標為 `SAMPLED`，不可寫成全套 gate。
- 規則、PRNG、ECL、save、engine schema、renderer、GUI、字型、動畫、音訊、
  資產 pipeline 或跨平台程式碼有變更時，跑 affected repo 正式套件；CoAB 用
  `go test ./cmd/... ./gamepack ./internal/...`，Ebiten package 在 Docker/Xvfb
  驗證。
- 連續低風險抽樣最多四個 milestone；第五個 milestone、重大整合點、README
  截圖更新、release／完成聲明前必須重跑正式全套 gate。若抽樣曾失敗、碰到
  非預期跨模組副作用或無法清楚界定影響範圍，立即升級全套，不等週期。
- 全套 gate 用 `./tools/go.sh test ./...`。`tools/go.sh` 預設採用本專案映像
  `coab-go-ebiten:1.24`（`tools/build-go-image.sh` 建立），它帶 Ebiten／oto 需要的
  X11 與 ALSA 開發標頭，並用 `with-xvfb` 在 Xvfb `:99` 下執行——沒有這顆映像會
  退回 `golang:1.24`，`internal/sound` 與 `cmd/azure-bonds-game` 會建置失敗。
- `workplace/` 是本地工作目錄（`.gitignore` 已整個排除）。在裡面放一次性的
  `package main` 探查工具會讓 `go test ./...` 出現 `main redeclared`；
  一律加 `//go:build ignore`，單檔 `./tools/go.sh run <檔案>` 不受影響。
- `git diff --check`。
- deterministic screenshot 或 runtime trace。
- 玩家路徑驗證：分段進行，每段可用 direct-entry 旗標進入，但每段都要走到
  該段的正常結束狀態（含接縫交接），不是只驗一個畫面或一次 opcode。
- 原版／remake 對照，明確標示 exact／reconstructed／未完成。
- 戰鬥不能只驗證靜態 layout 與數值；原版 DOS runtime／遊戲影片還要逐項
  對照近戰、弓箭／投射物、法術施放與命中、死亡動畫、音效及回合節奏。
- 公開遊戲影片是動態演出的 oracle：每個遠程／法術能力至少記錄影片 URL、
  平台、絕對時間碼、逐幀順序與對應原始 sprite block；截圖只是關鍵幀，
  不能單靠截圖推論動畫、等待時間或聲音次序。
- 戰鬥驗收表必須分開追蹤 caster windup、travel、impact、damage text、
  saving throw、death、area／persistent effect、sound cue 與 handoff。弓箭、
  Magic Missile、Fireball、Lightning Bolt、Stinking Cloud／Cloudkill 等
  不得因共用 projectile 素材而省略各自的後續效果。
- 若網路影片無法證明規則數值、範圍或目標順序，回到 DOSBox 重現及 executable
  反組譯；影片只能證明實際看見／聽見的演出，不可越界宣稱規則已還原。

不得用窄測試支撐「完整可通關」「完整中文化」「完整戰鬥」等廣泛聲明。

## 9. 目前權威狀態

### 本 milestone 基底

- 工作已於 2026-07-29 恢復，不再遵守舊的「暫停新增功能」文字。
- CoAB 本輪基底：第 515 輪火刀戰後第一個地圖位置轉移資料契約、第 514 輪
  提爾佛頓下水道入口至火刀檢查站正常 GEO 路徑、
  第 513 輪提爾佛頓盜賊公會內部正常 GEO 路徑、第 512 輪
  提爾佛頓城門／皇家馬車正常 GEO 路徑、第 511 輪
  提爾佛頓設施正常移動路徑、第 510 輪正常新遊戲地城
  移動交易、第 509 輪 PC-98 Action
  target／QUICK 清除與第 508 輪 SCAN producer
  milestone；本輪另完成第 517 輪反組譯缺口盤點與 worklist，兩個 repository 的
  實際 HEAD／remote 才是最終版本依據。
- **engine 依賴以 `go.mod` 為準**，不要引用文件裡的 hash：文件寫下的版本會過期，
  而 `go.mod` 鎖的那一個才是建置實際用的。要升級走
  `tools/engine-proxy.sh` ＋ `tools/go.sh get`（見 §7）。
- 本文件所在 commit 會晚於上述 CoAB 基底；compact 後永遠先以兩個 repo 的
  實際 HEAD／remote 為準，不要把文件內 hash 當成可自我引用的 latest hash。
- GUI 原版石框、人物／3D／PIC 分離舞台、16×15 倚天與 PC-98 typography
  study 已完成並 push；角色資訊全頁與所有畫面逐張 fidelity 尚未完成。
- 主線**已經跑得完**：同一條 session 從開場走到擊敗提朗瑟克斯的結局選單，拆成
  23 個段 subtest（`docs/plan/seg-21-ending-report.md`）。這不等於全城市、全地城、
  全事件或全戰鬥規則都完成——逐項狀態看 `WORKLIST.md` 與完整度矩陣。
- 主要缺口見 `WORKLIST.md`（執行順序）與
  `docs/knowledge/coab-re-coverage-matrix.md`（完整度）：全地圖與全事件覆蓋、
  戰鬥規則／AI／戰後、全翻譯校對、音樂音效 cue、完整 save 相容與三平台發行。
  `docs/project-status.md` 只作追溯，不是現況。
- 第 543 輪已把同一新遊戲 session 從第 542 輪的艾森布拉城外接到 Hap 村落、
  熔岩洞、巫師塔、回洞穴與熔岩池第二次戰鬥／防火桶分支；正式測試為
  `internal/game/campaign_normal_test.go` 的
  `TestRealNewGameRunsToTheEnding`。測試只用正常
  `MoveDungeon`、game-pack stable option ID 與同一 ECL continuation，不能因這條
  路徑通過就宣稱全城市、全地城或完整結局。
- 第 543 輪的巫師塔座標修正是 engine＋JSON 通用邊界：
  `original.geo5.block-33` 的 JSON `spawn=(7,15,W)` 表示目的地錨點；有 spawn 的
  map 會保留 live dungeon cursor，避免同區塊 ECL redraw 前的 scratch
  `C04B/C04C/C04D` 覆蓋玩家座標。Hap external exit 的 `roof_type=2` 已加入
  engine schema。engine 本地提交 `9cf5fa5` 若尚未能推送到 remote，不得寫成 GitHub
  已完成；compact 後先查兩個 repo 的實際 status／remote。
- 第 543 輪 READY 規格為
  `docs/spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md`；固定
  `PROGRAM 8` 與 Myth Drannor 終戰 fixture 仍只能算局部／coordinate-assisted
  證據，下一步須沿同一 session 接尤拉什、摩安德之坑、散提爾堡與結局。
- 第 544 輪新增 engine `set_memory`／`Runtime.MemoryWrites`，CoAB JSON 以
  `zhentil-keep.inner-city.route-memory-reset` 保存 `0x4C00` 的 raw route
  dependency；這不是 D&D 規則欄位，也沒有替它建立語意名稱。正常 session 測試
  已改名為 `TestRealNewGameRunsToTheEnding`，只驗證到
  Dexam 洞穴入口 `(4,5,N)`；洞穴內部路徑仍是下一個玩家可見 milestone。
- 第 510 輪已把新遊戲進入提爾佛頓後的第一個正常西行輸入收回
  `State.MoveDungeon`；原始 GEO2 block 1 的起點、雙側牆／門可走性與 ECL
  register 更新是 `exact`，但 DOS movement loop 的逐幀／逐指令對應仍是
  `strong inference`。前端與測試不得再各自實作一套座標交易；必須透過此中立
  方法，事件續跑仍要保留同一 `BlockSession`。這只證明開場後一格正常路徑，
  不得把它擴大成完整開場到結局或完整地圖。
- 第 384 輪已接通 Standing Stone→Myth Drannor→ECL6/GEO6 block `0x40`
  的正常玩家路徑。第 385 輪進一步證明 exact 出生點 `(2,15,E)`，修正帶
  文字 `NEWECL` 的 pending 座標 handoff，並沿兩步可通行 GEO 路徑完成
  terrain `01h` PICTURE 72 精靈幽魂、三分支 oracle 與 JSON Journal 25。
  下一步是 Journal 25 指向的 terrain `82h` 紅網、`Krrkik`、蜘蛛與
  rakshasa。
- 第 386 輪已證明 terrain `82h` 的 `SPEAK` 是
  `INPUT STRING 8,[7F79h]`，且紅網 branch 不比較 `Krrkik`；它是 Journal
  提示而非程式密碼 gate。VM／State／Ebiten 已提供可續跑 Unicode 字串輸入，
  正常 GEO 玩家路徑可抵達紅網；ENTER 的蜘蛛→PICTURE 72→rakshasa 兩戰
  continuation、HACK 與 RETREAT 有 real-image regression。
- 第 387 輪已從上述正常 GEO 路徑選 ENTER，以真實 MON6CHA 完成四蜘蛛
  勝利、同 ECL session 續跑、PICTURE 72 羅剎妖揭露、第二戰勝利、
  `4CBFh=1`、地城返回及重踏不重播。Journal 25 的「強大力量」是陷阱話術；
  ENTER 指令只寫 combat work `7F72h`、`4CBEh` 與 `4CBFh`，不能自行新增
  Strength buff。戰敗、毒素、羅剎妖完整能力與後續區域仍未完成。
- 第 388 輪已沿正常 GEO 路徑抵達 terrain `04h` 的隨機墳墓事件，完成
  2 巨型蜘蛛、3 相位蜘蛛、1 thri-kreen 戰鬥，以及 REBURY／LOOT
  兩分支。`4CBAh` 的 raw 中立值是 `80h`；重新安葬加一，搜刮減一並取得
  一件珠寶。純 coin／gem／jewelry TREASURE 現在會進入 service boundary，
  不再誤成零怪物 COMBAT；三個選項已由 engine `option_rules`＋CoAB JSON
  stable ID 驅動。四批上限長回歸、完整怪物規則、treasure TAKE／SHARE、
  thri-kreen 清楚可見的 placement 及後續 Burial Glen 仍未完成。
- 第 389 輪已由墳墓 `(6,12)` 沿 GEO auditor 證明的九步可行路徑抵達
  `(13,14)` terrain `03h`。黛米爾的 ACCEPT／REJECT／KILL／FLEE、正面
  祝福、負面寬恕、`4CBAh +5／-10`、`4CBBh 02h／FEh` 與一次性
  `4CC0h` 均有 real-image／正常玩家路徑回歸。
- 第 390 輪已關閉 `4CBBh` consumer：ECL6 戰鬥入口會 exact
  `SAVE [4CBBh]→[7F71h]`；Borland type table 證明
  `VARLISTTYPE` 是 `7C00h..7FFFh` 的 1024×2-byte array，故
  `VARLIST+06E2h=7F71h`。IDA `ATTEMPTTOHIT` 對該 byte 執行 `CBW` 並加入
  attack roll，確認 `02h／FEh` 是 `+2／-2`；`DOPOSTCOMBAT` 戰後清除
  `7F70／7F71`。engine `combat_modifiers`、CoAB JSON 與 Battle
  side-scoped modifier 已接通，正常玩家路徑命中邊界證明生效且不改寫
  Fighter 基礎 AttackBonus。離開 Myth Drannor 後清除 `4CBBh` 的 writer、
  所有特殊武器 caller 與完整戰鬥 fidelity 仍未完成。
- 第 391 輪已將使用者提供的 1014×759 DOS 顯示擷取還原至 320×200，
  確認畫中人物為 `HEAD2 02 + BODY2 02`、88×88、native anchor
  `(28,24)`。engine game-pack schema 保存 scene-character anchor／clip；
  CoAB HEAD／BODY 不再走 PIC cover，並恢復 EGA 黃色裂紋人物內框。
  `-inn` 與不同 selector 的 `-temple` 正常玩家路徑均有 640×480 回歸畫面。
  來源經顯示縮放，故整體標示為
  `material-exact/layout-reconstructed`，尚未宣稱 palette-cycle pixel-exact。
- 第 392 輪已從黛米爾沿正常 GEO 路徑繼續五步到 terrain `93h`、再四步
  到 `94h`。raw ECL 與完整玩家路徑共同證明十隻／八隻 PHASE SPIDER、
  `7F82=8／9`、`4C01=10／8`、勝利後 `4CCD／4CCE=1` 與重訪 EXIT。
- 第 393 輪已從 `94h` 沿正常 GEO 路徑回到 `(14,10)` terrain `95h`；
  六隻 PHASE SPIDER、`7F82=10`、`4C01=6`、`4CCF=1` 與骨堆
  `LOOT／REPLACE IN CRYPTS／IGNORE` 三分支均已完成。LOOT 的 exact
  `TREASURE` 是一顆 gem、`ItemBlock=FFh` 無裝備，並使 `4CBA-1`；
  REPLACE 使 `4CBA+1`，IGNORE 不改好感。raw ECL 與 Standing Stone
  起始的正常玩家路徑均有回歸；相位蜘蛛特殊能力與 DOS 動態演出仍未完成。
- 第 394 輪已沿正常 GEO 路徑接通 terrain `8Eh／8Fh／90h`。前兩道
  十二人／六人守軍分別寫 `4CC8／4CC9`；營地先打十二人並寫 `4CCA`，
  再依這兩旗標決定是否追加兩波六人。raw session 的完整 `12→6→6`
  與正常玩家路徑只剩首波均有回歸。原始財寶 exact 為 9500 gold、
  4 gems、6 jewelry、`ItemBlock=81h` 一件 random item；treasure menu
  現保留當次資料包翻譯文字。THRI-KREEN 特殊戰鬥能力與 DOS 動態演出
  仍未完成。
- 第 395 輪已接通 terrain `91h／92h`。`91h` 是八隻 GIANT SPIDER、
  `4CCB=1`；`92h` 先依 `4CBA < 80h` 判定是否略過幽魂警告，高好感
  才顯示 YES／NO。NO 不寫 `4CCC` 且可重訪；YES 在戰前寫
  `4CCC=1`，建立四隻巨蛛並把敵方 attack-roll work `7F70=2`。raw
  高／低好感與正常 Standing Stone 玩家路徑均有回歸。蛛卵只有敘事，
  原 ECL 沒有額外選單／財寶。
- 第 396 輪已沿正常玩家路徑接通西側王庭。terrain `08h` 門口幽魂 YES
  傳送 `(4,2,S)`；`89h` 四個盔甲選項造成好感 `+1／-2／-2／不變`；
  `8Ah` 以 `4CBA >= 80h` 為友善門檻，敵對時建立 `6+4+4` 十四名敵人；
  `8Bh` 友善王后給 12 gems、8 jewelry 與 ITEM6 block `41h` 六筆物品，
  敵對分支先扣五點，再依 YES／NO 給較少財寶或拒絕，最後倒塔傳送
  `(5,2,S)`。友善獎勵測試必須載入真正 ITEM6 blocks，不能把未載素材的
  零怪物 COMBAT fallback 當作通過。READY spec 396 是權威細節。
- 第 397 輪證明王庭不是出口，並由 `(1,3)` 沿 19 步合法 GEO 路徑抵達
  terrain `05h`。紅羽戰士 `WAIT` 會解鎖手札 33；同行陷阱的單條
  `DAMAGE(2,1d6+6,35h)` 是兩次隨機目標箭擊，之後載入
  PHASE SPIDER ×6 與 RAKSHASA ×1。raw 三分支、拒絕繼續、資料化提示／
  選項／手札及 Standing Stone 起始戰後骸骨選單均已回歸。
  `COMBAT／FLEE` encounter action、
  羅剎妖完整能力／戰利品／弓箭演出仍是明確缺口。
- 第 398 輪證明 terrain `07h` 沒有完成旗標，會重複建立六相位蜘蛛與一
  羅剎妖；不可自行改成一次性。正常玩家路徑已接通 terrain `0Ch`，
  `WAIT／PARLAY` 寫 `4CC7=1` 並解鎖完整繁中手札 56。Burial Glen
  東界 `PATH／WOODS／TURN BACK` 也已接通；`PATH` 由原 ECL register
  進入 block `42h` `(0,12,E)`。State 對地城內所有 `NEWECL` 投影
  `C04B／C04C／C04D`，不得為作品座標另寫 Go 特例。下一步探索完整
  block `42h` 遺跡，之後才是神殿與結局。
- 第 399 輪已接通 block `42h` terrain `01h`。`4CD0` 控制一次性提爾雪雅
  事件，`WAIT` 解鎖手札 5；第一戰是 HELL HOUND `44h`×5＋MARGOYLE
  `45h`×5。選擇攻擊貝爾哈時，第二戰是 RAKSHASA `43h`×1＋兩種隨從
  各六，且 `LOAD CHARACTER 8 → team 80h → ADD NPC 43h` 將第一隻
  RAKSHASA 設為 QuickFight 臨時盟友。DOS combatant array 固定保留
  index `0..7`，不得用 active party 長度計算 monster index，也不得把
  這種 monster ally 解析成永久 player record。READY spec 399 是權威細節。
- 第 400 輪已接通 terrain `02h／83h` 倉庫。`4CD1` 讓提爾雪雅結盟與
  直接擊敗六地獄犬＋六石像鬼匯合；普通踏入只顯示物資，真正
  `SEARCH (7ECA=1)` 才發出 9,500 gold、8 gems、8 jewelry 與 ITEM6
  `82h` 兩件裝備，服務返回後 `4CD2=1`，重搜不複製。State 現追蹤
  boundary transaction，只在來源分支 `NEWECL／EXIT` 後清除 `7ED5`，
  避免污染目的 block 又不影響同 block 事件。READY spec 400 是權威。
- 第 401 輪已接通 terrain `04h／05h／06h` 完整事件鏈。救援男子會打
  HELL HOUND ×6，臨終使用 `HEAD6／BODY6 40h` 並設
  `4CD3／4CD4／4CD5=1`；拒絕救援仍可追擊或離開，屍骸由 `4CD4`
  一次性控制。有線索且主動 `SEARCH` 才取得一枚 electrum 與 ITEM6
  `43h` 三件裝備，之後清除 `4CD5`。財寶 GP 投影新增 0–199 copper
  餘數，空白 ECL text 不再覆蓋跨 pause 的繁中藏寶敘事。READY spec 401
  是權威。
- 第 402 輪已接通 terrain `07h／08h／09h`。無名者事件由原始
  `4C06=1` 一次性控制，顯示 HEAD `43h`／BODY `46h`；灌木誘餌由
  `4CD6=1` 控制，拒絕會讓地獄犬帶走受害者，救援則先套用
  `DAMAGE 0Ch,2d8,34h`，再建立 HELL HOUND `44h`×5、MARGOYLE
  `45h`×5、RAKSHASA `43h`×1。ECL 在 `CALL 2E10h` 前寫入
  `(11,10,S)`；VM 保留 block／PC 有序 SAVE 與 CALL trace，State 只在
  session 未跨 block、方向於 CALL 前新寫入且座標屬於同批提交時同步，玩家可
  從新座標向南進 terrain `09h`，且不覆蓋跨 `NEWECL` 出生點。READY
  spec 402 是權威。
- 第 403 輪已接通 terrain `0Bh／0Ch／8Ah／8Dh`。石像鬼陷阱是全隊
  `DAMAGE C0h,3d10,saveFlags=1`，只改寫 X 與方向而保留 Y，之後可突襲
  RAKSHASA ×1；這使 `CALL 2E10h` transaction 支援有新方向提交的部分
  座標，並以 Filani 對話中只有 X／Y scratch writes 的反例防止誤傳送。
  賭局房迎戰 MARGOYLE ×8、RAKSHASA ×6，勝利取得 11,200 gold、
  15 gems、9 jewelry 與 ITEM6 `81h` 一件隨機物品。羅剎妖居所只有
  HAUGHTY 交涉解鎖由使用者 PDF 證實的手札 57；其他態度會迎戰
  HELL HOUND ×5、MARGOYLE ×5、RAKSHASA ×6。下水道柵口經兩次確認
  `NEWECL 43h`，正常玩家抵達昏暗廚房 `(15,15,N)`。READY spec 403
  是權威。
- 第 404 輪已接通 block `43h` terrain `8Ah／8Bh／8Ch`。廚房、班恩
  辦公室與豪華臥房三段繁中均由 game-pack stable ID 提供；臥房搜刮的
  exact TREASURE 是 5,000 GP、5,000 PP、12 gems、15 jewelry，
  ItemBlock `FFh`，合計 30,000 gold。Standing Stone 起始正常玩家路徑
  沿合法 GEO 到 `(10,12)` 並驗證財寶返回。ECL6 block `40h:+0F52h`／
  `42h:+0F70h` 會先寫 block `43h` 重用的全域 `4C05／4C06=1`，因此完整
  支線路徑的辦公室與廚房原作即為靜默；不得按 block 自動清旗標。READY
  spec 404 是權威。
- 第 405 輪已接通 terrain `87h／88h／89h` raw ECL：犬舍是 HELL HOUND
  `44h`×10，活動雕像是 MARGOYLE `45h`×10，私人禮拜堂是 HIGH PRIEST
  `48h`×1 加 PRIEST OF BANE `46h`×4。五段繁中均在 game-pack；
  犬舍原始流程另有一個空文字 PRESS pause，不可自行刪掉。正常 Standing
  Stone 玩家路徑已完成禮拜堂戰並到 `(7,10)`；下一步 `(7,11)` terrain
  `83h` 會觸發 `82h–85h` 提朗瑟克斯／無名者最終儀式，西翼犬舍與雕像
  因此仍只有 raw branch，不可宣稱 player-path 完成。READY spec 405
  是權威；下一步先完成最終儀式 gate。
- 第 449 輪已建立可重現的提朗瑟克斯 37 人終戰畫面：正常初始化 block
  `43h`、經 terrain `97h` 樓梯與十步 GEO 路線抵達 terrain `9Ah`，並以
  runtime 斷言 `45h×28／47h×1／48h×8`。大型 CombatMap 正式鏡頭依
  RuleBook 跟隨主動角色；README 圖片的 `47h` 首領焦點只在
  `-inner-final-battle`＋`-screenshot` 啟用，不可誤搬進正式玩法。日常回歸
  依使用者指示採代表性 vertical slice／高風險狀態抽樣，不必每輪從開場
  marathon 到結局；完整端到端仍是發行前驗收門檻。READY spec 449 是權威。
- 第 450 輪已把法師塔庭院→德拉坎德羅斯→龍群幻象→手札 15→枷印消退
  從 State／舊 locale 複本遷移至 CoAB game-pack stable IDs。手札兩頁必須在
  原 ECL 事件觸發後直接加入遊戲內 `JournalPages`，PDF 只作證據，絕不能叫
  玩家離開遊戲查閱。後續遷移故事時沿用 `text_rules + message_id +
  journal_message_ids`，並同步刪除被取代的 Go／locale fallback；READY
  spec 450 是權威。
- 第 451 輪已把法師塔四分支、龍心與屋頂出口的十段作品文字及三個專用
  選項移入 CoAB JSON，State／舊 locale 複本已刪；法師塔入口至離開不再靠
  Go story switch。後續不能因程式仍有作品中立 UI fallback（例如返回、儲存）
  就與劇情硬編碼混為一談，但所有本作人名、劇情、地名、選項與手札都必須
  逐章遷出 Go。READY spec 451 是權威。

### 發行前視覺與跨作品知識稽核（不可遺忘）

- 完成功能清單後，逐張以 DOS／PC-98 oracle 重查 UI 石框、內框、區塊比例、
  HEAD／BODY 人物組合、人物頭像、第一人稱 viewport、對話框與戰鬥 HUD；
  原版忠實 layout 是基準，繁中字級／分頁可依 640×480 合理調整。
- README 所有圖片都要核對目前 HEAD；錯誤 GUI、過期 renderer、debug-only
  假畫面或落後版本必須重擷取／替換，並保留 exact／reconstructed 標籤。
- 專案完成前，把 SSI Gold Box 的 RE、ECL／DAX／GEO／save、CJK 排版、戰鬥、
  音訊與驗證經驗整理成繁中知識庫與可重用 skill，供後續系列作加速。
- `/home/anr2/cht/daemon_winter` 可作 SSI RPG 比較樣本；Wasteland 是後續
  中文化目標。不可只因同為 SSI 就宣稱共用引擎：必須用第二款遊戲的 bytes、
  runtime、格式與 adapter 驗證後，才把真正共通機制提升到獨立 engine／skill。

### Go 玩家文字資料分離 gate

- 第 452 輪 `cmd/coab-audit`／`internal/sourceaudit` 已建立非測試 Go 漢字
  literal exact baseline；初始 1,260 signatures／1,315 occurrences。正式 gate
  會阻止新增、改字、搬動、重複或刪除後未更新的漂移。這不是豁免額度。
- 每個故事／UI milestone 必須先移入 stable ID＋locale／game-pack、刪除 Go
  與舊資料複本，再於同一 commit 用 `-write-baseline` 更新下降後基線。只改
  baseline 掩蓋新增中文字串屬於 gate 規避，禁止提交。
- 分類只是 heuristic；`runtime_ui_debt` 不代表可永久留在 Go。完整規則見
  `docs/audit/README.md`，READY spec 452 是權威。
- 第 453 輪已用訓練場證明 gate 的下降流程：提示、選項、升級結果、職業與
  可學法術名稱改由正式 locale stable ID 解析，測試也從同一 JSON 取值；
  baseline 現為 1,251 occurrences。後續必須讓此數字持續下降，不能重新加入
  Go 中文 fallback，也不能在測試複製目前譯文。
- 第 454 輪神殿服務再移除 28 次，baseline 現為 1,223 occurrences；十種
  cure 的 Go 結構只保存 stable key／價格，顯示名稱與選單均由正式 locale
  取得。後續設施應沿用此「typed rules＋stable display ID」界線。
- 第 455 輪酒館與手札 17 再移除 54 次，baseline 現為 1,169 occurrences；原始英文
  option／prompt／ECL fragment 仍是來源身分，繁中只能由 locale stable ID
  取得。缺 key 時顯示 ID 可用於 fail-closed 診斷，但不能視為完成中文化。
- 第 456 輪商店 UI 再移除 69 次，baseline 現為 1,100 occurrences；交易規則
  保留 typed item／price／coin，顯示格式由 locale 取得。
- 第 457 輪物品名稱再移除 126 次，baseline 現為 974 occurrences；type 與
  name-number 保留 DOS typed IDs，繁中 base／修飾詞與顯示格式由 locale
  `item_type_XX／item_name_XX` 解析，商店、裝備、戰利品與 CLI 共用同一組字
  函式。
- 第 458 輪效果名稱再移除 21 次，baseline 現為 953 occurrences；二十個既有
  已命名 raw affect kinds 與未知格式由 `affect_kind_XX／affect_unknown` 解析，
  CLI 與測試不再依賴 `ChineseAffectName`。顯示名稱資料化不等於效果規則、
  動畫、音效或生命週期已完成。真實 MON6SPC block 67 另有未命名
  `82h／81h／3Ch`，必須保持 raw 診斷直到 consumer 證據成立。
- 第 459 輪角色建立再移除 86 次，baseline 現為 867 occurrences；22 個單職與
  18 個多職模板由 engine `character_creation.templates`＋CoAB JSON 驅動，
  template display ID、前端提示、種族／職業與能力名稱不再寫死在 Go。錯誤
  template 必須阻止建立開啟，不得靜默過濾；正常 title→建立→ECL1 block 01h
  與 save round-trip 已回歸。完整原版建角 UX／規則仍未完成。
- 第 460 輪 ECL 選項再移除 70 次，baseline 現為 797 occurrences；舊
  `localizeOption` switch 的 84 個 source tokens 已由 CoAB game-pack
  `option_rules` 完整覆蓋。已知 token 不得在 State 恢復作品專用 switch；
  未知 token 要保留原文供診斷，再由真實 ECL 證據補 JSON。若 pack 與 UI
  locale 共用 stable ID，測試必須阻止兩份正式 JSON 漂移。
- 第 461 輪遊戲內手札再移除 75 次，baseline 現為 722 occurrences；15 個
  ECL 來源觸發與 23 頁內容由 CoAB `text_rules`／locale 驅動。runtime 只能
  附加匹配結果，不得依條目編號、人物或中文內容解鎖；產品測試必須從正式
  game-pack stable ID 取得期望頁面。手冊 OCR 只能協助轉錄，事件可達性仍以
  原始 ECL 與正常玩家路徑為準，且不得把整本手冊預先解鎖。
- 第 462 輪開場敘事與 shadowed journal fallback 再移除 34 次，baseline 現為
  688 occurrences。ECL 文字規則必須依實際 service boundary 建模；原始檔的
  換行不是畫面 boundary。新遊戲醒來 PRESS 與手臂印記 PICTURE 是兩個 stable
  messages，不能合併後讓 continuation 重播或漏字。刪除 fallback 前必須證明
  game-pack rule 先命中，並重跑所有受影響正常玩家路徑。
- 第 463 輪 Hap／熔岩洞穴 27 個作品事件再移除 27 次，baseline 現為 661
  occurrences；Area 5 長路徑正常覆蓋其中 26 個 boundary。資料規則 coverage
  不等於玩家分支 coverage：`lava-tube.sly-parlay` 尚未走正常路徑，文件與完成
  表不得把它和已走過的 `nice／combat` 合併宣稱。區域文字遷移必須同時驗證
  戰鬥 continuation、旗標、NPC roster、GEO handoff 與下一區入口。
- 第 464 輪 Area 5 離場與 Essembra 七個事件再移除 6 次，baseline 現為 655
  occurrences；事件數不一定等於漢字下降數，因為舊分支可能只回傳 locale
  key。阿卡巴離隊、黑暗精靈裝備銷毀與龍巫妖文字不可只測畫面；必須同時驗證
  roster、指定 item types、世界 block handoff、戰鬥後位置與城市服務返回。
- 第 465 輪 Hillsfar 五個事件再移除 4 次，baseline 現為 651 occurrences；
  dockside bar 舊分支只有 locale key，故事件數與漢字下降數不同。城市資料化
  除文字與兩場戰鬥外，還要驗證 service return 不重播進城音樂、離城才切回
  荒野 cue；UI 訊息通過不能替代音訊狀態 handoff。
- 第 466 輪 Yulash 十二個事件再移除 13 次，baseline 現為 638 occurrences；
  事件數與下降數仍可能因舊 UI locale 複本而不同。占領城市鏈可能跨 world
  trail、edge、entry menu、GEO terrain、戰鬥、手札與下一區入口；資料化後
  必須保留完成旗標、戰後 continuation、指揮官側門座標及章節 handoff，不能
  只驗證文字 matcher。
- 第 467 輪 Pit 開場／同伴／樓梯／屍體十四個 boundary 再移除 17 次，baseline
  現為 621 occurrences。相鄰原始 ECL 行不一定屬於同一畫面；垂死牧師介紹與
  「被選中之人」之間有 PRESS，必須由 runtime boundary 決定 stable message。
  同伴文字資料化也要驗證 roster stable identity、職業／種族、手札、樓層
  handoff 與一次性事件，不能只看愛麗雅絲／龍餌姓名是否顯示。
- 第 468 輪 Pit 摩貢祭壇至離場十八個 boundary 再移除 18 次，baseline 現為
  603 occurrences。多段儀式不能只串接後用 `Contains` 抽樣；每個 pause 都要
  對應 stable ID，並在同一 session 繼續驗證兩戰、護手旗標、一次性財寶、
  手札、最後阻擊、NPC 離隊與下一 world edge。只有其中一段通過不能宣稱章節
  transaction 完成。
- 第 469 輪 baseline 維持 603，但清除散提爾堡／眼魔洞窟測試中的繁中片段
  複本與 `forceEnemyDefeat`。資料早已在 JSON 不代表驗收分層正確；測試也必須
  從 stable ID 取得訊息／手札。直接把敵方 HP 設零只能驗證戰後 continuation，
  不能支撐「戰鬥可完成」；至少要用有界回合推進目前 combat runtime，並如實
  記錄高能力 fixture 與尚未還原的特殊能力。
- 第 470 輪阿沙本福德／立石群／世界路線十一個 boundary 再移除 9 次，baseline
  現為 594 occurrences；事件數與下降數不同，是因河畔酒館與傳聞舊分支只回傳
  locale key。作品 Tavern Tale 不屬於共用酒館 UI catalog；遷移後要同步移除
  舊 catalog coverage 要求，並由 game-pack 兩語系 rule 與正常酒館 continuation
  驗證。世界 route prompt 也屬作品 script 資料，不能留在 State。
- 第 471 輪提爾佛頓公會至火刀據點十四個 boundary 移除後，baseline 為 580
  occurrences。連續 PRESS 劇情不能只盲目呼叫 `Select` 再驗證最後戰鬥；每個
  pause 都要以 game-pack stable ID 比對完整訊息。長測試若直接設定 GEO 座標，
  只能稱 coordinate-assisted integration，不可寫成完整正常移動玩家路徑。
- 第 472 輪火刀據點十一個房間 boundary 移除後，baseline 為 569。raw ECL
  regression 要保留英文 token、menu、傷害／財寶與 work flag；產品 State
  regression 則由 game-pack stable ID 取得完整顯示文字。兩層測試不可互相取代，
  也不得把 direct terrain selector 誇大成逐步行走玩家路徑。
- 第 473 輪皇家馬車／監牢／下水道十五個 boundary 移除十六次後，baseline
  為 553；次數多一是舊假國王 fallback 由兩個 literal 串接。共享
  `DO YOU SURRENDER` 等 token 時，game-pack text rules 必須具體規則在前、寬
  fallback 在後，並以完整 source batch 驗證 RuleID，不能只測最小片段。
- 第 474 輪高階祭司／四段夢境／返城七個 boundary 移除八次後，baseline 為
  545；高階祭司舊譯文原有 batch 與 line 兩份。規格寫明某分支不代表玩家回歸
  已走過；應實際選擇被拒絕的 ENTER CITY、驗證 stable ID／PRESS／返回 menu，
  再繼續世界旅行。第 474 輪曾誤稱 Tavern Tale 44／60 無真實 ECL 回歸；spec
  330 與長測試其實已有艾森布拉 RELAX／BEER 證據，該斷言由第 475 輪 supersede。
- 第 475 輪 Tale 44／60 資料化並移除 State line localizer 27 筆中文 fallback 後，
  baseline 為 518。判定證據缺失前要搜尋 READY spec 與長測試鄰近城市流程，不能
  只搜 raw token。已由 coverage／玩家路徑保護的 locale adapter 使用 stable ID
  作顯式 fallback，不能保留第二份 Go 中文譯文掩蓋 JSON 缺 key。移除 fallback
  後若測試顯示 ID，應補正式 locale／coverage 並讓產品長測試載正式 catalog；
  不得把中文塞回 Go 或擴大合成 fixture。
- 第 476 輪將 CAMP／REST／FIX／VIEW／MAGIC 與 ALTER 的 128 次 Go 漢字移除，
  baseline `518→388`。共用服務的行為測試也必須載入正式 catalog；只做逐鍵
  coverage、但互動測試仍靠合成 catalog＋Go fallback，仍會掩蓋格式參數與語意
  漂移。`go.mod` checkpoint 也屬正式 gate：本輪證明 `f06493f` 缺目前使用的
  `CharacterCreation` schema，已改鎖 exact `f3c652a`；不能任選相容 Engine HEAD
  冒充 pinned dependency。READY spec 476 是權威。
- 第 477 輪將 PROGRAM 0／3／8 的十筆終局文字移入正式 locale，baseline
  `388→378`。locale catalog 不屬於 save payload；凡長路徑建立新的 State 再
  `LoadPartyFile`，初始、一般 restore 與臨時盟友 restore 都必須重新注入同一
  正式 catalog。若後段顯示 stable ID，不能把譯文塞回 handler 或把 catalog
  序列化進 save。PROGRAM side effect 與 `PROGRAM_WIN_SAVE／PROGRAM_END` token
  保持 typed；READY spec 477 是權威。
- 第 478 輪移除新遊戲預載的八頁開發用假手札，baseline `378→368`。手札的
  身分是 game-pack `journal_message_ids`，不是翻譯後文字；Engine 必須把 ID
  與顯示頁面一起交給 adapter，State 以 ID 去重。save v10 只保存有序 ID，
  讀檔依目前 locale 重新解析，禁止把繁中全文或 locale catalog 寫進存檔。
  空手札不得顯示 `1 / 0`，也不得為了填畫面預先建立攻略／開發說明頁。
  READY spec 478 是權威。
- 第 479 輪把 23 個 MON 戰鬥者名稱移入 Engine `combatant_name_rules` 與 CoAB
  game-pack，baseline `368→345`，`localization_debt` 降為 0。名稱身分是原始
  MON `SourceName`，顯示值才是目前 locale 的 `Name`；save v11 必須保存來源名
  並在讀檔時重解。測試用 source 經正式 pack 取得期望名稱，不可再複製「黑龍」
  等當前譯名。Exact overlay 不證明其他 SSI 遊戲共用 MON 格式；READY spec 479
  是權威。
- 第 480 輪把 `drawCombat` 的 HP／AC、施法／移動 prompt、十二個法術快捷提示、
  target／quick status 共 25 筆中文移入 State locale contract，baseline
  `345→320`、frontend `133→108`。Renderer 只負責原版石框上的座標、字型、
  顏色與顯示時機；typed casting／move／target state 決定 stable ID。測試必須
  從正式 catalog 取得期望值，不能把當前提示全文貼入 frontend test。
  READY spec 480 是權威。
- 第 481 輪把 Cloudkill、Stinking Cloud、Lightning Bolt 七筆逐目標動畫訊息移到
  State visual-result locale contract，baseline `320→313`、frontend
  `108→101`。Renderer 只能傳入 typed `VisualEvent／VisualFrame` 並繪製；effect、
  impact、save、protected 與 phase 的文字選擇留在 adapter。錯誤 phase 或
  protected 必須保留 fallback，不能因資料化改變 travel→impact→commit→death
  →handoff 與聲音順序。READY spec 481 是權威。
- 第 482 輪把 F5／F9 存讀檔、音訊續跑錯誤、ALTER 改名與 ECL 字串輸入
  十三筆前端中文移入 State typed locale contract，baseline `313→300`、frontend
  `101→88`。Renderer 不得依錯誤是否為 nil、editor 種類或地城模式自行選翻譯；
  它只繪製 State 提供的 value／help／result。路徑、錯誤與輸入值是 runtime
  argument，不可寫死到 JSON；save／ECL continuation 與輸入限制也不能因
  資料化而改變。READY spec 482 是權威。
- 第 483 輪把冒險選單、事件／圖片繼續、暗影谷 AREA、世界地圖、角色欄、
  戰鬥檢視與倒地標記二十五筆前端中文移入 typed `PlayerUILabel`，baseline
  `300→275`、frontend `88→63`。同一「繼續」identity 可由不同 layout 放在
  不同座標，不能因此複製譯文；game-pack 地名必須跟隨 State catalog language，
  禁止 renderer 固定 `zh-TW` 或另存作品地名表。READY spec 483 是權威。
- 第 484 輪把鎖門、地城 lifecycle、Pick／Knock／Bash 結果與正常探索操作列
  十三筆前端中文移入 State typed locale contract，baseline `275→262`、frontend
  `63→50`。Locale identity 不能取代 `flags=2／3`、一次撬鎖、spell slot 消耗、
  `UnlockDoorWrapped` 或 Bash 規則；先得到 typed result，再解析顯示文字。
  未經第二款作品驗證，不得把 CoAB 門旗標抽成共用 engine 格式。READY spec
  484 是權威。
- 第 485 輪把素材載入、AREA、GEO geometry、地城研究 preview 與世界地圖日期
  二十四筆前端中文移入 typed diagnostic locale contract，baseline `262→238`、
  frontend `50→26`。`LOAD PIECES` selectors／State `LoadPieces` 是 `uint16`，
  禁止因樣本值小就 cast／縮窄成 `uint8`。Preview 門選項選完整 stable ID，
  不拼接局部譯文；日期從 typed clock 格式化，不切割已翻譯時間全文。READY
  spec 485 是權威。
- 第 486 輪清除最後二十六筆 Ebiten 前端漢字，baseline `238→212`、frontend
  `26→0`。Automation 必須用 exact ECL source token／stable option ID 找選項，
  禁止比較 `Choices` 譯文；劇情抵達至少以 stable game-pack message ID 解析目前
  locale，不得在 Go 複製中文 fragment。Demo fighter technical ID 與 locale
  顯示名分離。Frontend 歸零不代表 runtime 212 筆或全翻譯已完成。READY spec
  486 是權威。
- 第 487 輪將建角完成、手札框、荒野選單、世界地點、NPC 名稱與城鎮／地城
  提示的三十九筆執行期中文 fallback 移回正式 catalog，baseline `212→173`。
  `catalog.Text` 的中文 fallback 仍是重複真相；正式玩家流程以 stable ID
  fallback 並由 coverage test 封鎖缺鍵。產品測試載入正式 catalog、動態解析
  期望文字，不能靠膨脹最小 fixture 或複製 JSON 中文通過。READY spec 487
  是權威；runtime 仍有 173 筆待清理。
- 第 488 輪把財寶列表、角色收取、取消／略過與缺 ITEM 素材九筆中文副本移入
  正式 catalog，baseline `173→164`。TREASURE 顯示列使用 locale 物品名／角色
  名，但控制流必須保存 `TREASURE_ITEM_n`、`TREASURE_CHARACTER_n`、cancel／
  exit identity；不可用翻譯文字控制 ECL continuation。正常火刀辦公室路徑
  改載正式 catalog，仍驗證原財寶與一次性搜尋。READY spec 488 是權威。
- 第 489 輪把 PARLAY prompt、五種 tactic 與 generic 結果七筆中文副本移入
  正式 catalog，baseline `164→157`。控制流只使用 `PARLAY_HAUGHTY／SLY／
  MEEK／NICE／ABUSIVE` identity；測試不能複製中文策略陣列。法師塔與
  Myth Drannor 羅剎妖居所正常長路徑同時驗證 identity／locale projection，
  再追到原 ECL 後果。READY spec 489 是權威。
- 第 490 輪清除 `combat_state.go` 戰鬥檢視、移動／防守、近戰、法術、快速
  戰鬥、怪物施法與勝敗等六十三筆中文副本，baseline `157→94`。規則先產生
  typed result，再選 stable message ID 與 runtime arguments；翻譯不得反推
  protected／save／damage。依風險分級跑 `internal/combat + internal/game`
  抽樣 gate，marker `ROUND490_SAMPLED_EXIT=0`。READY spec 490 是權威。
- 第 491 輪資料化 WHO、PICTURE、戰鬥錯誤、紮營 HP 列、未知法術、ECL 字串
  prompt、旅店與遊戲時間，baseline `94→77`；`internal/game` 抽樣 marker
  `ROUND491_SAMPLED_EXIT=0`。角色選擇保存 ID／index，時間保存 typed fields，
  locale 只格式化。READY spec 491 是權威。
- 第 492 輪把反組譯／PC-98／VFD／overlay／Borland symbol／DOS 角色工具的
  最後 77 次 Go 漢字 literal 移到 `internal/tooltext/messages/zh-TW.json`；
  help、error 與 Sound BIOS 報告參數改用 stable ID＋embedded catalog，工具
  控制流、原始 bytes、位址與報告欄位不變。exact AST baseline `77→0`，
  `ROUND492_SOURCE_AUDIT=0`、`ROUND492_FORMAL_EXIT=0`；READY spec 492 是
  權威。這只表示工具鏈資料分離完成，不代表完整遊戲已通關或完整中文化。
- 第 493 輪依 PC-98 overlay 09 的非破壞性 IDA report 接通 Quick AI 第一個
  `MinRange>0` 區域法術 `Sleep (15h)`：State 以 game-pack `min_range`、
  `TACTICALMAP`、敵人格建立 bounded `SCAN` center，重用既有 Sleep effect／
  TWINKLE／slot／active-save pipeline；無合法候選時 fail-closed。Quick target
  linked-list 的 tie／random 細節標為 `strong inference／unknown`，Fireball、
  Lightning Bolt、Stinking Cloud、Cloudkill 的 Quick Area 仍未接通。READY
  spec 493 是權威；本輪 focused regression 已通過，完整 Quick AI／完整遊戲
  仍不可宣稱完成。
- 第 494 輪沿用同一份 PC-98 overlay 09 證據，接通已有完整效果 pipeline 的
  Quick `Fireball (2Fh)`：以 game-pack `min_range=3`、`TACTICALMAP` 與敵人格
  建立 bounded `SCAN` center，將中心保存進 point action；raw
  `CastingTime=03h` 先走 pending scheduler，續跑後才進 Fireball 多目標
  travel／impact／damage／death visual 與 slot transaction。候選 linked-list
  tie／random、完整 area-safety、Lightning Bolt、Stinking Cloud、Cloudkill
  的 Quick Area 仍是 `strong inference／unknown` 或 fail-closed。READY spec
  494 是權威；focused regression、Docker／Xvfb 正式 gate
  `ROUND494_FORMAL_EXIT=0` 與 `coab-audit total=0` 已通過，完整 Quick AI／完整
  遊戲仍不可宣稱完成。
- 第 495 輪沿用同一份 PC-98 overlay 09 證據，接通 Quick `Stinking Cloud (22h)`
  與 `Cloudkill (5Bh)`：前者依 game-pack `min_range=1` 走
  `TACTICALMAP／SCAN`，後者依 `min_range=0` 與 line-terrain 合法敵人格建立
  point；後者 raw `CastingTime=05h` 先走 pending scheduler。兩者都只重用已
  驗證的 persistent-area、豁免／低 HD／中斷與法術格交易 pipeline。候選
  linked-list tie／random、完整 area-safety、Lightning Bolt line target 與
  完整 Quick AI 仍未完成；中心政策是 `strong inference` bounded adapter，
  不得升格為原版 exact。READY spec 495 是權威；focused regression、
  Docker／Xvfb 正式 gate `ROUND495_FORMAL_EXIT=0` 與 `coab-audit total=0`
  已通過，完整遊戲仍不可宣稱完成。
- 第 496 輪沿用同一份 PC-98 overlay 09 證據，接通 Quick `Lightning Bolt (33h)`：
  以 game-pack `min_range=0`、line-terrain 與有位置敵人建立 bounded point，
  raw `CastingTime=03h` 先走 pending scheduler，再重用既有折線／反彈、逐段
  impact、save／電擊保護、音效與法術格 pipeline。candidate pointer projection、
  tie／random、牆角反彈與完整 Quick AI 仍未完成；stable ID 順序是
  `strong inference` 暫時 adapter，不得升格為 exact。READY spec 496 是權威；
  focused regression、Docker／Xvfb 正式 gate `ROUND496_FORMAL_EXIT=0` 與
  `coab-audit total=0` 已通過，完整遊戲仍不可宣稱完成。
- 第 497 輪接通 Quick 牧師 targeted spells：`02h` Curse、`04h` Cause Light
  Wounds、`06h／07h` Protection from Evil／Good。以既有敵方、鄰接敵人或自身
  party target contract 建立 stable target，raw `CastingTime=10／5／4` 先走
  pending targeted action，再重用 effect／音效／slot pipeline。原版 object
  pointer 的候選順序、完整 `cast_on` consumer、敵方 Quick AI 與完整遊戲仍未
  完成；stable 第一筆只是 `strong inference` bounded adapter，不得升格為 exact。
  READY spec 497 是權威；focused regression、Docker／Xvfb 正式 gate
  `ROUND497_FORMAL_EXIT=0` 與 `coab-audit total=0` 已通過。
- 使用者已於 2026-08-09 恢復作業；compact／自動 continuation 應以第 493 輪
  的實際 HEAD、`docs/project-status.md` 與 `CONTEXT.md` 尾端為起點，不要把
  舊的「暫停」文字當成目前狀態。
- 完整玩家事項收斂後必須另做一次原版忠實 UI 終驗：逐畫面核對石框、內框、
  第一人稱視窗、HEAD／BODY 人物組合、頭像 anchor、戰鬥配置與中文排版；文字
  字級與換行可以為繁中調整，但不能藉此改掉原版區塊關係。README 只保留對應
  當時 HEAD、可重現且標明證據等級的畫面，錯誤或過期截圖必須替換。
- 最終經驗要回寫繁中 SSI Gold Box remake 知識庫與可重用 skill。可參考
  `/home/anr2/cht/daemon_winter` 的 SSI 類 RPG 做法，並為後續《Wasteland》
  中文化整理可沿用的解碼、資料化、中文字型與測試方法；不同作品的劇情、格式
  假說與硬編碼不得因此混入 Gold Box 共用 engine。
- 第 439 輪已證明 Sleep 的 `SCAN` source 是玩家／Quick 選定格，不是 caster
  footprint；`AOECOMBAT&7=1` 是 range，`FFh` 是 arc。PC-98 65 筆 TDEF
  exact 對應 `BackgroundTiles[1:66]`，floor byte 是一基底 TD。手動 `Z`
  玩家路徑已接通 terrain-aware order、`4d4`／HD／魔抗／`35h` effect 與
  slot transaction。32×16 fallback placement 仍是 reconstructed；Quick、
  wall/corner 動態、解除／save、twinkle／音效未完成。READY spec 439 是
  權威，compact 後不得再把 Sleep 掃描中心改回施法者。
- 第 440 輪已證明 overlay 23 `PUTDAMAGE` 的正傷害路徑呼叫 `REMOVEFX`，
  resident `DS:159Eh..15B1h` 的 19-byte effect table exact 含 `35h`，因此
  動態 Sleep 受至少一點傷害後解除，零傷害不解除。`[di+159Dh]` 是 DS
  位址，不是 overlay 同 offset 的 CS bytes；compact 後不得混淆兩個位址
  空間。duration、save 與醒來演出仍未完成；READY spec 440 是權威。
- 第 441 輪已證明 `CLOCK_` local `0020h` 是 effect duration consumer：
  `TIMEUNITS` 經 `MAXCOUNT DS:6804h` 換成 tick，遍歷 Player effect linked
  list，duration 零保留、未到期相減、到期呼叫 `SPELLOFF`。戰鬥每個新
  round 現扣一 tick，正常 level 3 Sleep 在總第 15 tick 解除。active battle
  save 與到期演出仍未完成；READY spec 441 是權威。
- 第 442 輪已證明 effect `35h` 的 add／remove callback 都走 `CLEARACTION`；
  `PUTDAMAGE` 必須先移除 Sleep，再做一般 spell interruption，否則會重複
  消耗法術格。成功 Sleep 的 `TWINKLE` 是 runtime 建立的四格 24×6 圖示，
  `16h／17h` 不是 DAX block；成功者逐人播放 `SPELLHITFX`，抵抗與醒來不
  播放。現有 palette pixels 仍是 reconstructed，READY spec 442 是權威。
- 第 443 輪把 remake JSON save 升至 v7；active combat 必須整體保存 Fighter／
  effect／Action、stable TeamList、scheduler selection、turn cursor、areas、
  modifiers、pending interruption 與 battle PRNG。不得只存 `MonsterAffects`
  冒充可恢復戰鬥。第 446 輪已把 visual elapsed 移入 State 並解除中段限制；
  READY spec 443 是 active combat 基底，spec 446 是 visual resume 權威。
- 第 444 輪已把 save v7 接到真實 Standing Stone→Myth Drannor→紅網玩家路徑：
  四蜘蛛第一戰 party-turn 存檔，全新 State 從玩家自備 ECL／MON／ITEM restore，
  再完成蜘蛛、羅剎妖二戰與 `4CBF=1`。這證明 ECL combat handoff，但高數值
  測試英雄不代表原版 encounter balance；mid-animation 已由第 446 輪完成，
  其他 encounter 類型仍未完成。READY spec 444 是權威。
- 第 445 輪已驗證真實 mixed-team save：提爾雪雅事件的 party-side 羅剎妖必須
  只存在 Battle snapshot，保留 `QuickFight／TemporaryAlly`；`Characters`／
  `partyRoster` 只能保存永久英雄。讀檔後戰勝要寫 `4CD1=1`，並從 runtime
  party 移除盟友。不得從 Battle party side 全量回寫 roster，也不得只用 roster
  重建 active Battle；READY spec 445 是權威。
- 第 446 輪已驗證 save v7 的 mid-visual continuation：State 保存納秒 elapsed、
  travel／impact／death 已送 cue markers，frontend 由 saved base 疊加新的 speed
  scaled clock delta。Sleep 700ms 與弓箭 death frame 載入後同幀且不重播離散
  音效；越界 elapsed／marker fail-closed。PCM sample offset 與 BGM driver／
  synth 狀態仍未保存，不得把「cue 不重播」寫成音訊無縫續播。READY spec 446
  是權威。
- 第 447 輪把 remake JSON save 升至 v8；BGM snapshot 必須同時保存 stable track、
  sequence machine、renderer parameter、YM2203 opaque full state、resampler
  remainder、Timer B、silence、pending PCM 與 audible/read-ahead backlog。只存
  selector／tick 或 decoder current position 都會錯。engine `f06493f` 是作品
  中立 synth／resampler 基底；exact driver 十二曲 runtime oracle仍未完成。
  READY spec 447 是權威。
- 第 448 輪把 remake JSON save 升至 v9；DOS WAV／PC-98 speaker active one-shot
  以 backend＋stable selector／event＋44,100 Hz audible frame保存，並連同 enabled
  狀態 round-trip。只保存 `IsPlaying` 在 Position 前後都成立的 voice；自然結束、
  停用及舊版未保存的音效不能復活。backend／asset／seek 不符必須先停止 pre-load
  voice 並失敗即關閉。這是 remake continuation，不是原版 SAVGAM audio；READY
  spec 448 是權威。
- 第 423 輪由 PC-98 `GAME.EXE` spell records 與 overlay 09 Quick consumer
  推翻舊 class-local spell ID 假設：Player memorized／known bytes 使用全域
  ID，Protection From Good=`07h`、Magic Missile=`0Fh`。舊 spec 134／142
  已 supersede，camp／記憶／戰鬥分派均已遷移；完整 ALT+M Quick spell AI
  仍未完成。compact 後不得把手冊職業內排序重新當成 record ID。
- 第 424 輪已接通 `ALT+M` gate、作品中立 `1d7`／每層三次 Quick selector、
  CoAB JSON metadata 與全域 `0Fh` Magic Missile；Standing Stone→紅網正常
  玩家路徑已實際消耗 slot。非零 MinRange helper 會檢查落點範圍、team、save
  與 effect，不得簡化成距離。非即時 action handoff／Bless 已由第 425 輪
  接續；area spell、Cure special 與其餘 Quick 法術仍 fail-closed。
- 第 425 輪已由 overlay 13／08 exact 關閉非即時施法 action handoff：raw
  `CastingTime/3` 非零時保存 spell ID，delay 改為
  `max(1,delay-units)`，同輪再選中才 CASTSPELL。engine `dd99d29` 與 CoAB
  JSON 已保存 raw casting_time；Quick Bless `01h` raw 10→3 已接通。slot
  中斷消耗、其他 Quick 法術與原作 per-action target pointer layout 仍未完成；
  Cure 由 426–427、手動 CAST typed transaction 由 428 接續。spec 425 只代表
  當輪邊界。
- 第 426 輪已由 overlay 09 與 typed
  `00B8:0075h → overlay 13 entry 17 → +1E30h` 關閉 Quick Cure 指定目標
  handoff。注意 `entry 17` 不是 overlay 17；compact 後不可再次混淆位址
  空間。engine action 保存 opaque `TargetID`，CoAB `03h` 由鄰近受傷／倒地
  selector、scheduler delay 與 Standing Stone→Red Plume 真實箭傷正常路徑
  驗證。equal-HP exact tie 與 down-player status predicate 由 427 接續，手動
  CAST delay 由 428 接續；interruption 與其餘 Quick 法術仍未完成。READY
  spec 426 只代表當輪邊界。
- 第 427 輪由 Borland `COMPTARGCURE／CHARSTATUS／DXDIR／DYDIR` 與 raw bytes
  關閉 Quick Cure exact 順序：N→NE→E→SE→S→SW→W→NW→self；equal HP
  保留先掃到者，自身低於半血可覆蓋，active HP `>=8` 才讓合法倒地者優先。
  raw status `DEAD=6／STONED=7／GONE=8` 均被 set predicate 排除；remake
  `HealthStatus` ordinal 不同，compact 後不可直接用 raw 數字比較。spec 427
  是權威；同格多 corpse ordering、完整 raw status round-trip 與非正傷害施法
  中斷仍未完成。正傷害中斷由 429 接續。
- 第 428 輪已讓手動 CAST 與 Quick 共用 PC-98 `CASTCOMBATSPELL` 延遲交易。
  非零 delay 的單體法術保存 stable target ID，區域／直線法術保存 32×16
  格點；resume 後才結算並消耗 slot。這是 typed remake transaction，不得
  誤寫成原作 Action raw layout 已全部證明。正傷害中斷與 slot 時點由 429
  接續；其他中斷仍未完成。
- 第 429 輪已由 overlay 23 `PUTDAMAGE` 與 overlay 24 memorized-byte consumer
  證明：最終 applied damage `>0` 會中斷 pending spell、移除第一個 matching
  slot 並保留 Action delay；零傷害不觸發。remake 必須在所有傷害來源共用
  positive-damage boundary，不能只修近戰。Cloudkill 直接死亡已由第 430 輪
  證明走 effect `44h` 的獨立 consumer；其他非傷害狀態仍不得自行套用。
  READY spec 429／430 分別是兩條觸發路徑的權威。
- 第 430 輪以唯讀 PC-98 overlay 22／12、typed effect-table resolver 與
  overlay 24 consumer 證明毒雲術 raw effect `44h` 會獨立中斷 pending spell。
  `CastCloudkill` 必須在 direct-death handoff 清 Action 前建立 stable interruption
  event；HD 7+／豁免成功不觸發。這不證明沉默、麻痺、睡眠或石化，後者仍須
  分別閉合 writer→effect table→consumer。READY spec 430 是權威。
- 第 431 輪證明 held effects `1Fh／33h／34h／35h` 全指向 overlay 12
  `0075h`，再呼叫 overlay 24 `CLEARACTION 2A5Bh`。它清 pending spell、
  delay、guard 與 unknown `Action+06h`，但不呼叫 memorized-slot consumer；
  remake 不得建立 `SpellInterruption` 或消耗 slot。六章 MON*SPC 沒有這四種
  innate effect，動態 Sleep／Hold writer 仍未完成。READY spec 431 是權威。
- 第 432 輪已證明全域法術 `15h` 經 `INITSPELLS` dispatch 到 overlay 22
  entry 41／local `2547h`，並 exact 還原 `4d4`、HD
  `<=1／2／3／4／5／>=6 → 1／2／4／6／10或20／20` 的 ordered
  capacity filter。engine 只加入此 primitive；目標幾何、豁免、duration、
  raw effect writer 參數與動畫／音效仍未閉合，不得因已有 package 就把
  Sleep 列為完整可施放法術。READY spec 432 是權威。
- 第 433 輪訂正第 432 輪的 targeting 位址空間：預設／非戰鬥
  `0117:0034h` 是 overlay 22 `GETSPELLTARGETS 112Ch`；戰鬥初始化會改寫成
  `00B8:007Ah`，typed resolver 落到 overlay 13 `225Fh`。Sleep 的
  `AOECOMBAT=09h` 經 overlay 31 `SCAN 08D8h` 建立並排序三 byte候選表，
  handler 保留該順序交給 HD 容量篩選。`SAVERESULT=0` 證明無豁免，duration
  exact 為 `5×caster level`。`SCAN` 幾何欄位、large footprint／tie order、
  magic resistance、完整 `PUTEFFECT`、解除與演出仍未閉合；不得提前開放
  Sleep。READY spec 433 是權威，spec 432 的舊 targeting 敘述已 supersede。
- 第 434 輪以 `EFFECTS 013E:0089h → overlay 23 entry 21 → PUTEFFECT
  2325h` 關閉上一輪 module projection。Sleep 的順序是先完成 `4d4`／HD
  容量，再按 selected order 進 `CHECKFX(target,9)`；`6Ah` 魔抗成功會清
  current effect 並跳過 writer，但不退容量。成功 exact 寫 `35h`、
  `5×caster level`、`CASTON=1` 與 raw `+4=caster level`。remake 只新增接受
  上游 ordered IDs 的 bounded core，不自行猜 `SCAN` 幾何；手動／Quick UI、
  解除、save round-trip 與演出仍未完成。READY spec 434 是權威。
- 第 435 輪已由 overlay 31 連續 IDA 指令與 raw bytes 關閉 `SCAN` 三欄及
  local `0035h` 排序：`+0=object ID`、`+1=最短成功 LOS 加權距離低 byte`、
  `+2=方向 payload`。第三欄完全不參與排序；等距時只比較 object ID 與
  奇偶例外，且必須逐迴圈重現，不能換成一般 comparator。engine
  `combat/scanorder` 與 CoAB stable-ID adapter 已接通；terrain／wall 實機、
  手動／Quick Sleep 與演出仍待續。READY spec 435 是權威，spec 433 的舊
  第三欄 tie 敘述已 supersede。
- 第 436 輪已由 Borland symbol／type／member table、全新 IDA raw resident
  report 與 overlay 31 `LOSEXISTS` 閉合 `TDEFTYPE HT／LOS／SYM`、
  `TACTICALMAP XRAY／TD`、一基底 tile、2／3 加權距離與
  `2*range+1` gate。engine `combat/scan` 與 CoAB
  `BuildScanTargetIDs → CastSleepOrdered` bounded transaction 已接通；第四
  byte 仍命名 `Raw3`，`INARC` sector、COMPOBJ builder、wall／corner 動態
  trace、手動／Quick Sleep 與演出仍未完成。READY spec 436 是權威。
- 第 437 輪已由 overlay 31 `INARC 054Ah..08D5h` 關閉 `FFh→8`、八方向
  inclusive 半平面與第一命中方向；engine 以 14,062,500 組全地圖 corpus
  驗證。Borland symbols exact 命名 `LASTOBJECT／OBJECTLIST`，型別表證明
  72×4-byte，overlay 10 builder 證明一基底 linked-list 建表與 X／Y／自身
  index／footprint-active 欄位。combatant linked-list 到 stable fighter ID、
  正常 Sleep UI 與 wall／corner 動態仍未完成。READY spec 437 是權威。
- 第 438 輪已由 Borland `CHARACTERLIST／IDLIST／CHARRECPTR／CHARREC.NEXT`
  與 overlay 10 builder 閉合 object ID 的 pointer identity。State 重建
  `LegacyObjectID=1..N`，Battle 依 stable fighter identity 建立 footprint 並
  串接完整 SCAN producer；72 筆上限與身份缺損 fail-closed。Ebiten 尚缺原
  戰場 `TD/TDEF` projection，故正常 Sleep UI 不得提前開放。READY spec 438
  是權威。
- 第 421 輪以 PC-98 overlays `08／13／18／24` 的非破壞性 IDA 副本完成
  QUICK／GUARD／BANDAGE／SPEED 命令核心。新增的符號與型別只存在分析
  database／報告，不回寫原始檔；每項結論在 READY spec 421 保留地址、bytes
  與 exact／strong inference。engine `combat/action` 與 CoAB runtime 已接通
  命令狀態；`ALT+Q／ALT+M`、敵方命令 AI、原版逐幀時間仍未完成。
- 第 422 輪接通 overlay 08 證實的 `ALT+Q` TeamList transaction、目前角色
  delay `20→19` handoff，以及視覺播放中 Space 收回 PC。overlay 08／09／10
  已證明 `DS:A86Ch` gate 可控制 PC 的 Quick AI 法術 selector，但 selector
  優先序尚未關閉，所以 `ALT+M` 仍不得做成只有 UI 的假功能。READY spec 422
  是權威細節。
- 第 415 輪已由 PC-98 overlay 9／22 raw bytes 與 IDA 關閉提朗瑟克斯
  effect `84h`：type-14 action phase、`ROUND < 4`、spell `33h`、初始格
  `16d6` 與後續 range 10 路徑另一份獨立 `16d6`。作品中立 line profile、
  第 1–3 回合怪物排程、正式 terrain callback、繁中 stable IDs 與動態
  Lightning timeline 已接通；Standing Stone 起始真實 MON6 終戰仍完成
  `PROGRAM 8`。目標候選已由第 416 輪接續；終戰牆面逐幀與 timing
  仍未完成；READY spec 415 是雙傷害池權威。
- 第 416 輪已由 PC-98 `PICKTARGET 00B8:3D7F`、overlay 24／32 與 spell
  `33h` record 接通 range 10、cardinal／diagonal `2／3` 加權距離、雙方
  footprint、牆面候選及二十次不可見移除重抽。effect `84h` 不再先走全體
  living-opponent selector；無候選仍消耗 action，不回退物理攻擊。原始
  combatant-array 同距 tie order 與 `(0,0)` fallback 動畫仍未完成；
  READY spec 416 是權威。
- 第 417 輪已由 PC-98 `CHECKTARGET 00B8:11AF` 與 overlay 12 handlers
  `06F9h／1713h` 關閉 status visibility：`19h` 只在觀察者沒有 operational
  `18h` 時設 hidden 並使 attack roll `-4`；`47h` 無條件 hidden／`-4`，
  所以 Detect Invisibility 不能抵消 `47h`。法術 ranged selector 與物理 AC
  共用作品中立 visibility 規則，並以真實 MON6 提朗瑟克斯 effect `18h`
  驗證。spec 235 已 superseded，spec 410 過度合併 `19h／47h` 的斷言亦由
  READY spec 417 修正；blink／動物視覺的 combat consumer 已由第 418 輪接續，
  完整 effect 生命週期仍未完成。
- 第 418 輪已由 overlay 12 handlers `0BDBh／16C2h` 接通 effect `25h／45h`。
  Blink 只在 target action delay 0 時 hidden 並覆寫 attack roll `FFh`；
  `45h` 只對 `RACETYPE` `13h` observer 生效，`18h` 可取消 hidden 但不取消
  `-4`。`+11Ah` 又由 dragon-slayer `03h` consumer 與四筆真實 Animal records
  關閉為 `RACETYPE`；`ALIGNMENT` 是 `+11Bh`，真正 `MONSTERTYPE` 是 `+14Ch`。
  READY spec 499 supersede spec 418 的欄位命名，visibility bytes 本身仍可回查。
- 第 419 輪已由 PC-98 overlay 24 `DEXRABONUS 1416h`、overlay 13 local
  `0000h` 與 overlay 8 local `01FBh` 關閉 initiative writer／TeamList selector。
  原版使用 shared Player `+17h` DEX reaction table、`1d6` action delay 與每次
  全表逐節點 `1d100`；delay 優先、roll 解 tie、完全 tie 後掃者勝。engine
  `combat/initiative` 保存這套作品中立 primitive，CoAB Battle 保存建構順序，
  MON parser 也修正 combat team 是 `+198h` 而非 `+197h`。完整 effect
  duration／save、`area.field_596` surprise writer、DOS 等價性及原版底層
  PRNG 仍未完成。第 420 輪已推翻「一般 DELAY 是 20→19」：頂層 D 先進
  DONE 子選單，第二層 D exact 寫 delay 1 並由動態 scheduler 同輪重新入列；
  20→19 是 Quick handoff（第 422 輪已接 UI）。READY spec 419／420 是權威。
- 第 406 輪先完成 GUI fidelity 稽核：IDA／DOS bytes 證明 HEAD 後 BODY，
  BODY 執行 `row+5`；DOS runtime 則證明第一人稱／一般 PIC 使用獨立灰色
  88×88 舞台。640×480 frame 現保留原版上半部與命令帶，只在訊息區插入
  40 個原生列；第一人稱與旅店正常玩家路徑畫面已更新。READY spec 406
  只涵蓋這些繪製契約，角色資訊全頁、所有事件圖與戰鬥畫面仍未完成。
- 第 407 輪已從正常 Standing Stone 長路徑進入 terrain `83h`，完成十三段
  提朗瑟克斯／無名者儀式、六個 PICTURE 與完整繁中手札 48。敵軍是
  HIGH PRIEST `48h`×2、HELL HOUND `44h`×6、MARGOYLE `45h`×6；戰後
  `4C00=1`。同一路徑再穿越靜默 `84h／85h`，完成活動雕像與犬舍兩戰，
  `4C02／4C01=1`。`4CBD=0／1` 的 raw 儀式輸出相同，但這不能擴大成整章
  無效果。READY spec 407 是權威。
- 第 408 輪證明 terrain `86h` 不是終戰；正常長路徑由犬舍沿合法 GEO 經
  terrain `97h` 上二樓，抵達 `(6,1)` terrain `9Ah`。終戰 exact 為
  MARGOYLE ×28、TYRANTHRAXUS ×1、HIGH PRIEST ×8；正式 scheduler 勝利後
  同一 ECL continuation 呼叫 `PROGRAM 8`。真實零敵人 minion COMBAT 與跨
  boundary monster setup 已修正並回歸。READY spec 408 是權威；下一步應
  建立不改寫結果的終戰動態 capture，補齊提朗瑟克斯／祭司／石像鬼能力，
  並繼續關閉全新隊伍由開場至結局的單一通關缺口。
- 第 409 輪完成 remake save v6 的 ECL session／持續 PRNG snapshot。
  engine `randomstream` 保存 seed 與底層 draw count；CoAB 保存 current block、
  PC、stack、mutable memory、輸入 offset 與 pending monster descriptors，
  並以真實 ECL6 terrain `04h` 驗證讀檔後分支相同。READY spec 409 是權威；
  完整 UI／戰鬥 frame、音訊位置與 SSI 原版 RNG 仍未完成。
- 第 410 輪以 PC-98 Borland symbols 與 IDA 證明 `EFFECTREC` 是 9 bytes；
  `LOADMONSTER` 複製每筆後清除並重建 `+5..+8` linked-list pointer，卻不改
  byte `+4`。因此 MON*SPC raw `+4=0` 不能作為天生效果停用 gate。Battle
  以作品中立 `Innate` 標記保留 template 能力；Standing Stone 起始正常長
  路徑載入提朗瑟克斯六筆真實效果，`18h` 偵測隱形已抵消隱形 AC bonus，
  並完成 37 人終戰與 PROGRAM 8。READY spec 410 是權威。
- 第 411 輪以 PC-98 overlay 12 raw bytes／IDA 關閉魔法抗性 common
  routine：`base+(11-casterLevel)*5`，`1d100 <= threshold` 時
  `Protected(0)` 清除傷害。local `23F4h／2404h` 是 50%／15% wrappers；
  第 412 輪再以 typed TPOV resolver 證明 `6Ah → entry 100 → 2404h`，升級
  為 `exact`。Magic Missile 現先擲傷害再擲
  抗性，成功時傷害歸零、施放格與 continuation 仍進行；繁中訊息來自
  locale stable ID。READY spec 411／412 是權威；`70h／87h` runtime 已由
  spec 413 接通，`4Fh` 命中後 runtime 已由 spec 414 接通。`84h`
  Lightning Bolt、HIGH PRIEST／MARGOYLE 特殊能力、其餘魔法路徑與動態
  演出仍未完成。
- 第 413 輪把 `70h` Fire protection 與 `87h` Electricity protection 接入
  Fireball／Lightning Bolt。共用 core 只讀 raw damage flags；line spell
  呼叫端必須顯式傳入 flags，不得由 spell ID 或怪物名特判。保護成功仍保留
  visual／slot／turn transaction，只將 damage 清零並標 `Protected`。現行
  saving throw 先於 protection 的 PRNG 順序仍待 DOS runtime 驗證，不得把
  HP 結果正確擴大宣稱成完整亂數 fidelity。READY spec 413 是權威。
- 第 414 輪由 PC-98 overlay 13 post-hit caller、overlay 23 `CHECKFX`
  type 2／3 與 overlay 12 handler 共同證明：前兩物理攻擊槽命中且目標仍
  存活時，operational `4Fh` 對同一目標追加 `2d10` Fire＋Magic。Battle
  以 `AttackEffectResult` 分離武器傷害、擲骰與實際傷害；防火抵消時仍消耗
  兩顆 d10。真實 MON6 長路徑與繁中 stable IDs 已回歸。原版 4F 動畫／
  聲音、轉移目標動態 trace、`6Ah` 對 4F 的時序與 `84h` 仍未完成。
  READY spec 414 是權威。
- 使用者指定的角色資訊 DOS 實機圖已保存於
  `docs/reference/user-provided/dos-character-info-layout-20260730.png`。
  它是原版 layout oracle，不是 remake 截圖；左上 HEAD／BODY 人物舞台、
  右側 roster 與下方長文是三個獨立石框區域。

### 目前戰鬥 milestone（不可遺忘）

- 第 498 輪已由 PC-98 overlay 12 local `029Bh` 閉合 effect `0Ah`：
  `DAMAGEFLAG & COLDFLG(02h)` 後 `DAMAGE/2`，這是半傷，不是 immune；另將
  已閉合 `70h` fire immune、`87h` electricity immune 移入 CoAB JSON 的
  `combat_affect_rules`，由 StartCombat／active save restore 重新注入。這些
  能力必須保持 engine＋game-pack 分層，`Protected` 只代表完全清除傷害。
  effect `08h／09h`、寒冷法術 caller、完整演出與完整遊戲仍未完成；詳見
  `docs/spec/498-pc98-resist-cold-data-driven-affect-rule.md`。

- 第 499 輪已由 overlay 12 local `022Dh／0264h` 與 Borland type/member table
  關閉 effect `08h／09h`：active character `+11Bh` 的 evil／good 集合命中時，
  `SAVEROLL A02Ch += 2`、`ROLLTOHIT A039h -= 2`。新增 engine
  `combat/modifier` 與 CoAB `combat_conditional_modifiers` JSON；`Battle` 將
  防禦方 raw effect 與攻擊者／施法者 alignment 分開，套入物理攻擊、Fireball、
  反射線、Stinking Cloud、Cloudkill 的 save boundary。`+11Ah=RACETYPE`、
  `+14Ch=MONSTERTYPE` 的舊命名已 supersede；未知 alignment fail-closed。
  focused Docker gate 已通過，正式 gate、完整 effect lifecycle、完整戰鬥與全作
  通關仍未完成。權威規格為 `docs/spec/499-pc98-alignment-conditional-effects.md`。

- 第 500 輪已由 PC-98 overlay 12 local `2396h..23F3h／2404h` 與 overlay 23
  `PUTEFFECT`／type-9 list 關閉 effect `6Ah` 的 base `15` 與
  `base + (11-casterLevel)*5` 公式；完整多目標 caller draw order 仍標為
  `strong inference／unknown`。新增 engine `combat/resistance` 與 CoAB
  `combat_magic_resistance_rules` JSON；Fireball／Lightning bounded core 由
  rule slice 判定並傳遞 typed `Resisted`，不把零傷害猜成魔抗，active-combat
  restore 重新掛入配置。`4Fh`、`84h`、Cloudkill／Stinking Cloud 其他 caller
  不得因本輪擴大解讀。Docker／Xvfb 正式全套 gate 已通過，marker 為
  `ROUND500_FORMAL_EXIT=0`、`coab-audit total=0`。權威規格為
  `docs/spec/500-pc98-spell-magic-resistance-boundary.md`；本輪以兩 repo 各一個
  集中 commit／push 收尾，完整遊戲仍未完成。

- 第 501 輪沿用第 414 輪 PC-98 `CHECKFX` 證據，將物理命中後 effect `4Fh` 的
  第一、第二攻擊槽、`2d10` 與 Fire＋Magic `damage_mask=09h` 移入 engine
  `combat/posthit` 與 CoAB `combat_post_hit_rules`。成功物理命中且目標存活、
  物理擊殺／未啟用效果／第三槽不 dispatch，以及先耗骰再判火焰保護均有 Battle
  與 save-restore 測試；生產邏輯不再硬編這組 kind／slot／dice／mask。`6Ah` 是否
  介入 `4Fh`、`84h`、完整動畫／音效／wall-clock timing 仍未知，不得升格。權威
規格為 `docs/spec/501-pc98-post-hit-effect-data-contract.md`，本輪完整遊戲仍未
完成。

- 第 502 輪沿用第 415／416 輪 PC-98 effect `84h` 證據，新增 engine
  `combat/monsterspell` 與 CoAB `combat_monster_spell_rules`。State／Battle 不再
  硬編 `33h`、round 1..3、range／line budget 10、兩份 `16d6`、`0Ch` mask 與
  reflection parameters；active save restore 重新掛入資料。原版 caster level、
  `6Ah` 魔抗順序、同距 tie、逐幀動畫／音效仍不可升格；權威規格為
  `docs/spec/502-pc98-monster-spell-data-contract.md`，完整遊戲仍未完成。

- 第 503 輪沿用 PC-98 overlay 09 local `04CCh..0624h` 的 Quick far-pointer
  candidate chain、`03D3h` suitability 與 `00FA:0048h` handoff，新增 engine
  `combat/quicktarget` 與 CoAB `combat_ai_target_rules`。Quick area、line 與四種
  targeted cleric adapter 現依保留的 one-based `LegacyObjectID` 排序；pointer-chain
  的完整 retry／tie／random、Magic Missile 目標與 Cure 專用規則仍不得升格，權威
  規格為 `docs/spec/503-pc98-quick-target-object-chain-boundary.md`，完整遊戲仍未
  完成。

- 第 504 輪沿用 PC-98 overlay 09 `04CCh..0624h` 的連續控制流，將 `1..7`
  retry range、priority `7` 起點與逐級降低移入 engine `combat/quicktarget`
  及 CoAB JSON。法術 suitability 只做無亂數合法性 probe，最終 area／line／四種
  targeted cleric handoff 才使用同一 Battle PRNG 一次；無目標時不消耗 spell
  slot 並回到普通 action。`3E01:142Dh` helper 演算法、完整 pointer-chain
  producer／tie、Magic Missile／Cure 專用規則與敵方 Quick AI 仍未知；權威規格為
  `docs/spec/504-pc98-quick-target-priority-retry.md`，完整遊戲仍未完成。

- 第 354 輪時間軸、原版 COMSPR projectile 與 engine JSON 資料化已完成。
  `combat.VisualEvent` 使用 windup→handoff；箭、Magic Missile travel／impact
  的 source block、flip、scale、原始 delay 已移入 CoAB `combat_visuals`，
  frontend 不再寫死這三組作品素材表。
- DOS 影片 `wwYsij1wDC4` `00:42:25.20–25.40` 已證實 generic spell travel
  先於文字／area effect。`00:36:15.40–17.00` 另證實 Fireball 是一次 cyan
  travel，之後依序對每名受影響目標播放 red-white impact、damage、選擇性
  death；不是一張大型圓形爆炸圖。相關關鍵幀與時間碼已存入
  `docs/reference/original-dos/` 與 `combat-video-oracle.md`。
- Fireball milestone 已完成 travel／target-impact JSON、
  通用 multi-impact timeline、正常玩家 slot／tile cursor 施法、敵我範圍、
  共用傷害骰、Spell save、逐目標聲音及三張 remake 對照畫面。仍缺有牆
  terrain path reachability、原 combatant-array／direction tie order 與
  原版 wall-clock timing，不可把這個 vertical slice 寫成全部 Fireball
  fidelity 或完整戰鬥。
- 第 355 輪 Lightning Bolt milestone 已找到 DOS 影片
  `07:40:22.50–25.60`；正常 memorized `0x33`／tile cursor、weighted line、
  地城牆反向、large footprint 去重、反彈重複命中、共用等級 d6、逐目標
  Spell save 與 interleaved segment timeline 已實作。CoAB JSON 宣告
  COMSPR `05/85` travel、`06/86` 電弧、`0A/8A` damage impact；三張 remake
  與五張原版關鍵幀已保存。牆角／多次反彈的 DOS runtime 動畫與敵方
  `throws lightning` 仍未完成。
- 第 356 輪 Stinking Cloud milestone 已完成 `0x22` 正常 slot／
  tile cursor、target-anchored 2×2 cells、terrain filter、Poison save、
  cough／`d4+1` helpless、重疊 instance、caster-level expiry、RANDCOM
  item 4 與兩張 remake 畫面；原版影片 `00:42:25.20–27.00` 三張後續關鍵幀
  已保存。coughing AC 仍未完成。
- 第 357 輪 Cloudkill milestone 已完成 `0x5B` 正常 slot／tile cursor、
  target-centred 3×3、HD 0–4 自動死亡、HD 5 的 `-4` Poison save、HD 6
  無修正、HD 7+ 無效果，以及低 HD 入格限制。party／MON*CHA offset
  `0xE5` 已投影到 runtime；RANDCOM item 2 由 CoAB JSON 驅動，持續區
  renderer 不再硬編碼作品 block。640×480 remake checkpoint 已保存。
  DOS 動態時間碼、protect magic／未命名 affect 免疫與每回合重複毒雲判定
  仍未完成。
- spec 354 仍 IN PROGRESS；下一步補 terrain-aware Fireball range／tie
  order、弓箭／Magic Missile／melee／kill 完整時間碼與逐距離 duration，
  並尋找 Cloudkill 的 DOS 動態影片 oracle。
- screenshot demo 必須凍結 `combatVisualElapsed`，不能讓 Ebiten／Xvfb
  啟動耗時推進 timeline；這是戰鬥影片時間碼比對的固定基線。
- 每個新增戰鬥效果完成影片核對、測試與畫面後才集中 commit＋push。

### 目前 PC-98 音樂 milestone（不可遺忘）

- 第 371 輪已用指定 IDA Pro 9.4 正確以 8086／16-bit 分析 `SOUND.ROM`，
  並由十二首 S98 的 72 組啟動序列證明
  `TL=127-OUTPUT_LEVEL`、algorithm carrier `4→2→3→1` 及
  `OPERATOR_MASK` key-on。舊的 64-bit raw ROM IDA 結果作廢。
- engine `audio/ym2203` 保存作品中立 carrier／operator 拓樸；
  CoAB `TrackPlayback` 已由 active parameter mask 產生 key-on，不再固定
  `F0h`。
- 第 372 輪已手動恢復 IDA 漏掉的 timer ISR 間接路徑，完成
  `audio/pc98soundbios` 六種 waveform、pitch／TL 投影及 S98 動態 extractor。
  第 373 輪又用 45.01 秒 selector 9 S98 證明 Hoot 沒執行 ROM Timer B
  LFO，再以 Unicorn 8086 直接執行 exact `SOUND.ROM`：sync 8 第 30 tick
  首次輸出、80 tick 共 51 組 pitch／TL。engine Timer B scheduler 與
  CoAB waveform／sync／speed adapter 已完成，spec 373 保存 scheduler
  本身的 READY 規格。
  第 374 輪再以指定 IDA、raw bytes 與同一份 S98 證明 MSCDRV 自己接管
  YM2203 IRQ：只在 Timer B 呼叫 `TrackPlayback`，不鏈回 Sound BIOS ISR。
  因此 CoAB faithful BGM 不執行上述 LFO，spec 374 保存 IRQ ownership
  的 READY 規格。
  第 375 輪又由 S98 證明 3,993,600 Hz／prescale 6；engine 已完成 Timer B
  完整 count period 與無 rounding drift 的 PCM sample accumulator，
  spec 375 是 READY 規格。
- 第 376 輪已固定 BSD 授權 `ymfm`，完成作品中立 YM2203 native PCM 與
  有理數 phase 線性重取樣；CoAB 已把 Sound BIOS intent 依 S98 exact
  register order 展開，接成 Timer B → PCM → Ebiten player。
  `-pc98-music-driver` 可由使用者本機 driver 播放，`pc98-render-track`
  可輸出 deterministic WAV。selector 5 兩次十秒輸出 hash 一致、非靜音且
  無 clipping；spec 376 保存合成器／播放器 READY 規格。
- 第 377 輪已由指定 IDA Pro 9.4、raw bytes 與 exact driver PCM 證明正常
  `MSCPLAY` 是 stop→800ms silence→play；播放器精確輸出 35,280 靜音
  frames 且不推進音序列。正常 loop count 0 轉成 `0xFF` 無限循環。
  driver 內部 40-tick fade／FM0 SFX 沒有正常 GAME caller，不可擅自接入。
  spec 377 是最新 READY 規格。
- 第 382 輪已依 ymfm 補上作品中立 `27h` reload 公式、phase `0..15`
  驗證及 phase-adjusted PCM accumulator；CoAB `20h→0Ah` 間的七聲道
  interpreter 是資料相依路徑，現行 instant-ISR renderer 維持 phase 0。
  下一步以 CPU／OPN 共時 trace 還原 CoAB 真實 reload phase，並以原機
  edge／錄音或經 microbenchmark 校準的 V30 emulator 校準 SFX machine
  profile，再處理 save/resume 與 analog mixer gain。第 381 輪已證明
  NP2kai i286c `_loop` 沿用 80286 `8/4` clocks，不得拿其 `CPU_CLOCK`
  取代 NEC／原機證據，也不得把目前可播放路徑宣稱成完整音樂／音效還原。

### 目前 PC-98 音訊研究 milestone（不可遺忘）

- 使用者提供的 `pc98/` 兩片原始 VFD 已完成唯讀稽核；來源檔必須保持
  untracked，不得修改或提交。兩片皆為 VFD1.00、77×2×8×1024 幾何。
- Disk 1 有兩個缺失 sector，其中一個正落在 `MSCDRV.EXE` 內，造成 1024
  bytes 缺口；另一個落在 `CED3.DAX`。Disk 2 尾端另缺 24 sectors。因此目前
  不可把抽出的 `MSCDRV.EXE` 宣稱為完整驅動，也不可用零填內容推論其 API。
- 已建立 read-only `internal/pc98vfd` parser 與 `cmd/pc98-vfd-audit`，保留
  缺 sector 狀態而不靜默補零。NP2kai 可從省略缺 sector 的可寫暫存 D88
  進入 MEGDOS 0.25 loader，但遊戲尚未成功啟動；該畫面只屬啟動鏈 exact
  證據，不是遊戲 GUI。
- 已從殘缺驅動交叉確認 YM2203 I/O ports 0x188/0x18A、timer handler 與
  INT D2 hook；公開 soundtrack 資料只作曲名／作曲者及平台交叉參考，不提交
  受著作權保護的錄音。
- MAME 官方 `azurebnd` FDI 雜湊與原 `LOADER.COM` 已交叉驗證；loader exact
  順序是 SETUP→MSCDRV→GAME。保留 absent sectors 的副本可進 MEGDOS，
  修改 sector topology、FAT root 或 loader 則在 banner 前停止，故 absent
  sectors 可能參與早期完整性／防拷；目前只能標 `hypothesis`。
- `GAME.OVR` 已確認是 TPOV；`cmd/pc98-ovr-audit` 由 resident control
  records 重建 36 段完整 chain，分離 code／relocation 後仍無 literal
  `CD D2`。`GAME.EXE` 的 Borland `0x52FB`／version `0x0208` 表則已由
  `cmd/pc98-symbol-audit` 解析為 1,725 筆 9-byte symbols、2,305 names 與
  53 筆 16-byte compiler modules。
- 音訊 symbols exact：`SOUNDFX 0893:0000`、`INITSOUND 0893:010D`、
  `MSCPLAY 0893:0114`、`MSCSTOP 0893:015E`、`BGMPLAY 0893:0177`。
  `MSCPLAY` 透過 IVT vector `7Eh` wrapper 送 0-based track；`BGMPLAY`
  已證明 selector `3/4/5/6/8/9/12`。IDA data-segment 校準與
  `INTERPET` writer／BGMPLAY consumer 已證明輸入是 `CURRENTECL
  0C29:BDF0`；完整 ECL block → selector／driver index 已寫入 CoAB JSON，
  engine `a81b963` 提供作品中立
  `music_tracks`／`music_bindings`／`music_cues`。
- State 會在初始 ECL 與 block transition 發出一次性 `MusicEvent`。`0x30`
  維持目前曲目、`0x52` 無分支。第 362 輪已由 IDA、PC-98 decoded ECL 與
  DOS 英文事件證明 `WLDTWN==0` 是區域／戶外導航、非零是城鎮設施選單；
  JSON 初始 binding 使用 selector 5，`pc98-town-services-menu` 使用
  selector 6。第 363 輪已將 `PICTURE 0x50/0x79` 做成作品資料化 cue；
  阿沙本福德 block `0x50` 與希爾斯法 block `0x51` 的正常 ECL 玩家路徑
  均證明 `5→6→5`，相同 track 不重播。仍不可先猜曲名。
- 第 364 輪已用指定 IDA Pro 9.4 與 raw bytes 證明 `MSCDRV.EXE` 直接把
  IVT `7Eh` (`0000:01F8`) 設成自身 `CS:0080`；public ABI 是
  `AH=0/AL=0-based track` play、`AH=1` stop，再由內部 D2h clients 接到
  `CEE0` Sound BIOS 服務。六個 bridge anchors 都位於 file
  `0x0280..0x1376` 或 GAME
  `0x9410..0x95D9`，不與 MSCDRV 缺口 `0x4000..0x43FF` 重疊。
  `cmd/pc98-music-audit`、兩支 `scripts/ida/pc98_*music*` 與 READY spec 364
  是可重現入口。
- 第 365 輪以 NEC 官方《PC-9800 Technical Databook BIOS 1992》證明
  `CEE0` 是 Sound BIOS 固定介面表、N88-BASIC 預設用 D2h；IDA＋raw bytes
  又確認本作 17 組命令 client 與兩個 direct YM2203 helper。新增
  `scripts/ida/pc98_sound_bios_audit.py`、READY spec 365；`cmd/pc98-music-audit`
  現會逐 byte 驗證命令集合。`AH=01 PLAY`、`AH=15 SETTEMPO` 在本作 wrapper
  區未出現，不可因官方 API 有定義便假設使用。
- 第 366 輪由 IDA＋raw bytes 解出 `DS:0330` 十二筆 64-byte descriptor
  與 84 個 channel stream；所有 sequence 位於 file
  `0x1B61..0x3C58`，沒有跨 `0x4000..0x4400` 缺口。Hoot `ponyca.xml`
  的 Shift-JIS code 又提供 exact 0-based 曲名 metadata；CoAB JSON 現有
  十二首中英文 `title_id`。`internal/pc98music.ExtractTrackSequences`
  只接受已辨識 driver SHA、selector 1–12，並複製完整七聲道 bytes，
  不把商業 sequence 放進 repo。
- 第 367 輪已解出 `sub_10410` 的 FM／PSG family-aware 指令寬度、
  `A0–A4` 控制流、16-entry call／loop stack 與 overflow／underflow
  no-op。84 組 sequence 各通過 256 timed events；channel 6 timing 分支
  會忽略控制 opcode 並 read-through descriptor end，auditor 必須維持
  `bounded-runtime-read-through`，不可誤套 FM／PSG range gate。
- 第 368 輪已完成正常配樂 `TrackPlayback`：tick 0 初始化、FM note、
  PSG 71-word period／envelope／modulation、tempo 與 Sound BIOS intent
  皆由 verified driver bytes 驅動。十二首各通過 4,096 ticks，共 68,291
  deterministic events；這尚未包含 fade／SFX 共存、外部 YM trace、
  parameter block 展開或合成器，不可宣稱音樂已可播放。
- 第 369 輪以指定 IDA Pro 9.4 修正音色表位址：
  `sub_112B0` 使用 `seg003:0542 + index×100`，對應 file `0x45A2`，不是
  舊假設 `dseg:0542`。NEC 50-WORD typed parser 已驗證二十組內嵌音色；
  十二曲所有呼叫另含 `20,21,23,24,25,26,27,58`。
- 第 370 輪使用 Hoot 2023、Wine／Xvfb 與修正後 `Home` 絕對游標流程，
  各自擷取十二首約五秒 S98 v3。最初相對游標 corpus 有重複曲目，已作廢。
  engine `1a6a252` 提供作品中立 S98 parser、YM2203 tone-load 與 key-on
  snapshot；CoAB `cmd/pc98-s98-audit` 以 exact driver、stream intent、
  內嵌 signature 與 trace 四方驗證。NEC rate／level 需反相，operator
  canonical 順序是 1,3,2,4，signed DETUNE 採 8-bit left shift。
- 高索引全部只在 descriptor 初始化短暫載入，第一個 stream `85h` 隨即
  改用內嵌 `0..19`，之後才首次 key-on；十二首
  `audible_parameters_complete` 均為 true。不得刪除高索引副作用，也不得
  再追假設中的外部音色 bank。第 371 輪已補完 total-level／carrier 與
  operator-mask key-on；第 372 輪已補 LFO 靜態核心，第 373 輪已補
  Sound BIOS Timer B cadence、sync state 與 ROM 動態 harness。第 374 輪
  證明本作 MSCDRV 接管 IRQ 並繞過該 LFO。第 375 輪已補 Timer B 完整
  count period 與 PCM 有理數 accumulator；第 376 輪已接入 `ymfm`、
  PCM stream 與遊戲播放器。第 377 輪已證明正常換曲 800ms silence 與
  `0xFF` 無限 loop，並把未使用 driver fade／FM SFX 分離。第 378 輪已
  找到正常 `GAME.OVR → SOUNDFX` 的 42 個直接 caller、八個 Borland
  module 的 selector 分布、port `37h` pulse routine 與
  `GAME.EXE` file `E66Ch` 的 20-WORD 音序表。`MOVEMENT` 三處皆為
  selector 10，與 DOS `step.wav` 共同證明腳步語意。作品端
  `internal/pc98sfx` 只從 exact 本機 executable 匯入 typed
  pulse／delay，不提交商業 bytes。第 379 輪由 Borland symbols 證明
  `CASTFX` 到 `CRASHFX` 全 selector，並命名 `REALMOVE／ANYUNDEAD／
  SHOWARROW／CASTSPELL／TWINKLE／SCAN` caller。State 現發平台中立
  intent；engine `audio/cyclepcm` 與 CoAB V30 8 MHz profile 已接入
  deterministic one-shot mixer。第 380 輪再以 IDA、raw bytes、Unicorn
  harness 與 NEC 官方表校正 `LOOP taken=5／exit=13`，並分離第一次／
  後續 gate-on 與一般／最後 gate-off overhead；第 379 輪舊 PCM 時長與
  hash 已作廢。第 381 輪已用版本化 probe 讓 NP2kai i286c/V30 core
  執行 exact routine，重現 `6,6,7,6,7,7`；但 source 與動態 delta
  都證明它使用 80286 `LOOP taken=8／exit=4`，不能作原機 V30
  wall-clock oracle。該 profile 仍是 timing-reconstructed；下一個真實
  缺口是原機 edge／wait／prefetch 校準、reload phase、save/resume 與
  analog mixer gain。原 WORD 不是已證明的 Hz。
- NP2kai backend 已證實 `C3/H0/R8/N3` baseline 被要求四次；首讀
  not-found、第二讀起補零仍停在 MEGDOS banner，CPU 尚未進
  `INT 21h/AH=4Bh`。不可把 absent sector 簡化成永久零填或一次性錯誤。
  Trace 見 `docs/reference/original-pc98/vfd-runtime-trace.md`。
- selector 規格 `docs/spec/355-pc98-ecl-bgm-selector.md` 已 READY；來源媒體
  規格 `docs/spec/358-pc98-vfd-and-fm-audio-source.md` 仍為 IN PROGRESS。
  `docs/spec/362-pc98-disk-b-dax-and-wldtwn.md` 已 READY；作品中立 cue 與
  正常玩家路徑已完成；`docs/spec/364-pc98-music-vector-bridge.md` 也已
  READY，Sound BIOS 規格 `docs/spec/365-pc98-sound-bios-d2-api.md` 也已
  READY，track/import 規格
  `docs/spec/366-pc98-track-table-and-runtime-import.md` 也已 READY，
  stream bytecode 規格 `docs/spec/367-pc98-stream-bytecode.md` 與 OPN
  event 規格 `docs/spec/368-pc98-opn-event-runtime.md`、音色 bank 規格
  `docs/spec/369-pc98-fm-parameter-bank.md` 與 S98 runtime 規格
  `docs/spec/370-pc98-s98-ym2203-runtime.md` 也已 READY。
  第 371–382 輪規格也已 READY。下一步用原機 audio edge／錄音或經
  microbenchmark 校準的 V30 emulator 校準 port `37h` profile，再補
  MSCDRV 的 CPU／OPN 共時 reload phase、save/resume 與 analog mixer gain。

### 第 505 輪 Quick target caller 補充

- overlay 09 `04CCh` 雖是 TPOV entry 4，但目前 IDA audit 找到的唯一直接
  caller 是通用 action dispatcher `0164h`；`0164h` 不是 TPOV handler。後續
  不得把 `04CCh` 的控制流自動套成所有 Quick 法術的共同入口。
- overlay 13 `PICKTARGET` `3D7Fh..3F5Ch` 可證明「已投影候選的一次抽樣、
  `CHECKTARGET`、失敗移除、最多 `14h` 重試」形狀；候選 producer、range、
  visibility、tie 與 RNG 仍須分開標示。engine `combat/quicktarget.SelectOne`
  只負責 legacy order 後的一次抽樣，CoAB adapter 才能提供候選。
- 這輪的完整輸入 hash、IDA image、report hash 與 resolver 輸出以
  `docs/spec/505-pc98-quick-target-caller-and-single-draw.md` 為準；不得以
  `0164h`、`04CCh` 或 IDA label 的相同數字跨位址空間推導正式欄位名稱；
  raw far call 必須保留完整 bytes 與解碼後的 `014A:00CAh`／`014A:00C0h`
  位址基準。

### 第 506 輪 ranged target tie 補充

- `PICKTARGET` 的同距候選排序目前只能作 bounded adapter：兩個候選均有完整、
  非零且不同的 `LegacyObjectID` 時使用原始一基底 combat-object 身分；任一方缺
  少投影時回到 stable fighter ID。這是 `strong inference`，不是完整原版
  comparator 的 `exact` 證明。
- `SelectRangedCombatTarget` 不得因這項 tie 修正額外抽亂數，也不得把候選
  producer、pointer chain 或 `3E01:142Dh` 的亂數演算法硬編進 core；完整證據仍要
  以原始 bytes、consumer 與 runtime trace 閉合。
- 第 506 輪規格與回歸測試是
  `docs/spec/506-pc98-ranged-target-object-order-tie.md` 與
  `TestSelectRangedCombatTargetUsesLegacyObjectOrderForEqualDistance`。

### 第 508 輪一般敵方 SCAN producer 補充

- overlay-09 caller audit 必須保留 `9A C0 00 4A 01` 的 raw bytes 與
  `014A:00C0h` resident far-call 位址；不可把 IDA flat label 或相同數字跨
  overlay／resident／file offset 位址空間合併成正式語意。
- overlay-24 local `285Bh` 的候選 producer 仍以 `DS:9F2Eh`、`DS:9F30h`、
  `DS:A820` raw work address 記錄；terrain／footprint／SCAN 的 exact 證據沿用
  spec 433／435／436，未知欄位不可因 consumer 叫作 physical target 就改名。
- 正式 CoAB JSON 的 `combat_target_rules` 宣告 `legacy_scan`、`max_range`、
  `arc`、`retry_attempts` 與 `retry_with_xray`；engine `combat/targetselect`
  只處理穩定 ID 的有限 random/remove loop，CoAB adapter 才建立 legacy object、
  TacticalMap 與 visibility。不得把 `max_range=255`、20 次或 `XRay` 寫死在
  renderer／State 的劇情分支。
- 一般敵方 State 只有在正式 game pack 與 TACTICALMAP provider 都存在時才走
  producer；舊 synthetic 測試可保留 bounded fallback，但正式玩家路徑應讓
  producer／map 錯誤 fail-closed，不可靜默回到 party[0]。
- 本輪只關閉候選 producer、visibility remove 與第二輪 wall-bypass；方向 tie、
  raw persistent Action target layout、movement／flee／guard、完整 monster AI、法術／
  近戰逐幀動畫與音效仍不可宣稱完成。權威規格為
  `docs/spec/508-pc98-general-target-scan-producer.md`。

### 第 509 輪 Action target／QUICK 補充

- PC-98 overlay 8 local `1375h..140Fh` 的 Quick setter 在同隊 target 時清除
  target far pointer；這是已證實的控制流，但 `Action+06h` 與完整 pointer layout
  仍未知。不要用同一數值或 IDA rename 替 raw offset 命名。
- engine `combat/action` 將目前 action 的 `ActionTargetID` 與延遲法術的
  `TargetID`、格點 target 分開；`TakeTargetedSpell`／`InterruptSpell` 不得誤清
  action target。stable ID 只是 `strong inference` typed projection。
- `combat_action_rules.clear_same_team_on_quick` 是 game-pack policy；State
  正式 JSON 路徑依它呼叫 Battle Quick，不能把同隊清除寫死在 renderer 或劇情
  分支。敵隊 target 保留、政策關閉時同隊 target 也保留的 regression 需持續存在。
- 權威規格為 `docs/spec/509-pc98-action-target-quick-clear.md`；movement／flee／guard
  producer、同格排序與完整 AI 仍是後續工作。

### 第 510 輪正常新遊戲地城移動補充

- `GEO2.DAX` block 1 的起點 `(7,13)` 與 `C04B／C04C／C04D = 7／13／1` 是原始
  bytes／ECL runtime 的 `exact` 證據；西行 direction `6` 的可走性必須由
  decoded GEO 雙側牆／門資料判定，不能在測試或 State 寫死 Windlord’s Inn
  座標結果。
- `State.MoveDungeon` 是正常地城輸入的唯一交易入口：檢查 cardinal delta、依
  `DungeonGeometryView` 處理 ECL local／combined GEO adapter、wrap 座標、更新
  wall／roof registers、清除 `7F81h` 單步 guard、送出 step sound，最後才執行
  per-turn 與 search-location continuation。前端只負責輸入、門選單與 renderer
  refresh，不得重複呼叫 lifecycle 或直接改寫 `DungeonX／DungeonY`。
- 這個交易與原 DOS movement loop 的對應目前是 `strong inference`；尚未有
  DOSBox 逐幀 input／timing capture，因此不能標為 instruction exact。若後續
  bytes 或實機觀察推翻順序，保留本輪 spec 並追加 supersede 勘誤。
- 權威規格為 `docs/spec/510-normal-new-game-dungeon-step.md`；正常路徑回歸為
  `TestRealNewGameBeginsAtGlobalBlockOne`。Windlord’s Inn 的文字期待值仍由
  game-pack stable message ID 取得，不得複製繁中顯示字串到測試。

### 第 511 輪提爾佛頓設施正常移動與 one-shot 群組補充

- `TestRealNewGameBeginsAtGlobalBlockOne` 已透過同一個 `State.MoveDungeon` 交易，
  從開場中斷休息後的位置逐格走到 Filani、Weaponers、Gond altar、Training Hall、
  Tavern 與高階祭司所在格；招牌／下水道傳聞 pause 必須由 locale JSON stable ID
  驗證，未知 pause 不得靜默吞掉。
- 原始 ECL2 block 1 SearchLocation 先以 terrain 得到 `7F7Ah`；高階祭司 handler
  `+1104h..+110Fh` 的 `AND 4C03h,7F7Ah,7F79h` 非零即 `EXIT`。正常路徑先經
  terrain `0x0D` 招牌後，`4C03h=0x80` 使 terrain `0x8F` 的高階祭司重訪保持
  靜默。這是 raw ECL `exact`、跨 handler 的「共用 `0x80` one-shot 群組」則是
  `strong inference`；不可為了讓後續 assertion 通過而清除 `4C03h`。
- 高階祭司本身以未先消耗該群組的 fresh ECL session 驗證 PICTURE 6、HEAD6／BODY6、
  YES／NO、Remove Curse、Journal 19 與離場 continuation；這不等同於把獨立
  direct-entry branch 冒稱成同一條正常玩家路徑。
- `TurnDungeonWithGrid` 只更新面向與 wall／roof 投影，不自行執行 ECL；後續
  movement／SEARCH 才消費新面向。權威規格為
  `docs/spec/511-normal-tilverton-facility-route.md`，知識庫為
  `docs/knowledge/golden-box-normal-player-path.md`。

### 第 512 輪提爾佛頓城門／皇家馬車正常路徑補充

- 高階祭司後到城門必須沿 decoded `GEO2.DAX` block 1 的 16 步可行路徑，逐次呼叫
  `State.MoveDungeon`：`(1,10)` → `(1,0)`，不能直接寫入城門座標或在測試中
  注入 ECL boundary。最後西行步驟在 `(1,0,W)` 觸發
  `tilverton.carriage-gate-closed`，選項續跑後仍在同格。
- 城門封路是在正常移動途中發生，不是路徑結束後可任意重播的提示。續跑後只以
  `TurnDungeonWithGrid` 轉向北方，再執行 lifecycle，才進入皇家馬車 `PICTURE 11`。
  此順序要保留同一 `BlockSession`，不可用兩次 lifecycle 製造同一事件。
- 皇家馬車段的穩定驗收依序包含強制攻擊、皇家衛兵戰、紅袍人綁走假國王、投降／
  入獄、盜賊 `PICTURE 2` 救援與公會 block 2 handoff；這只證明本段 continuation，
  不代表公會／下水道內部已完成正常移動。
- 綠袍女人傳聞已移入 CoAB game-pack `text_rules`，ID 為
  `tilverton.green-robes-rumor`，英文片段與繁中 locale 必須同時保存。State 不應
  再增加該事件的中文 fallback；測試用 game-pack resolver 取得期待值，不能複製繁中
  顯示字串。
- 權威規格為 `docs/spec/512-normal-tilverton-city-gate-route.md`，知識庫為
  `docs/knowledge/golden-box-normal-player-path.md`。本輪的 GEO／ECL／remake
  boundary 是 `exact`；State movement 對 DOS 逐幀／逐指令的對應仍是
  `strong inference`，不可擴大成整作通關或 DOS pixel exact。

### 第 513 輪提爾佛頓盜賊公會內部正常移動補充

- 公會 block 2 handoff 後到 `tilverton.guild-sewer-traces` 必須沿 decoded
  `GEO2.DAX` 的雙側 wall／detail 資料逐格走，使用 `State.MoveDungeon`；不能把
  `(11,7)`、`(15,7)`、`(10,13)` 或事件終點直接寫入 `DungeonX／DungeonY` 再跑
  lifecycle。`TurnDungeonWithGrid` 只轉向，不能代替移動交易。
- 本輪已證明 `(13,7)` 東側是 GEO 實心牆，不是遺漏的門；正常路徑要繞過回廊。
  四道 detail `2` 鎖門分別由測試用 deterministic bash fixture 開啟後，以
  `UnlockDoorWrapped` 更新兩側 detail。力量 25 只是一個可重播的測試輸入，不能
  被寫成原版開門規則或作品劇情常數。
- 新增公會內部文字／選項必須走 game-pack stable IDs：
  `tilverton.running-thieves`、`tilverton.option.remain-calm`、
  `tilverton.running-thieves-warning`、`tilverton.fire-knives-spot-you`、
  `tilverton.guild-assassins-attack`、`tilverton.guild-metal-and-animals`、
  `tilverton.guild-bodies-after-battle`；測試由 `option_rules` 解析原始選項，
  不得複製目前繁中顯示字串。
- 犬舍與路途中隨機遭遇要以真正 `CombatAct` 回合完成。生命週期可能先經過
  engine-only `CALL` 才建立 encounter；`dungeonLifecycleActive` 與
  `combatReturnMode=ModeDungeon` 用來保存 remake caller context，戰鬥勝利後若
  暫停在 `ModeEvent`，必須依同一 session `Continue()` 回地城，不能重跑 block
  起點或直接設座標掩蓋 continuation 遺失。
- 本輪終點只到 block 3 的 sewer traces；入口後 `(1,8)` 檢查點、騎士／火刀事件、
  深層下水道與返回世界地圖仍有 coordinate-assisted 段落。權威規格為
  `docs/spec/513-normal-tilverton-guild-interior-route.md`，知識庫為
  `docs/knowledge/golden-box-normal-player-path.md`。本輪是 `READY` 有界路徑，
  不得擴大成完整公會、完整戰鬥、完整中文化或整作通關。

### 第 515 輪火刀戰後地圖位置轉移補充

- 火刀檢查站勝利後，玩家按下 `tilverton.sewers-hide-bodies` 的 PRESS，ECL
  同一 session 續跑到 block 3 `+1B5Bh` 的 `CALL 2E10h`。目前 remake 的
  `RunResult` 沒有 `C04Bh／C04Ch` `SaveWrites`，若只續跑 bytecode 會停在
  `(1,8,S)`；不能把這種外部 runtime 缺口誤判成 ECL 沒有後續。
- Engine 的 `set_map_position`／`Runtime.MapPositions` 是作品中立 contract。
  CoAB JSON predicate（`ECL block 3`、`7F72h=FDh`、`4C2Bh=1`）才宣告目前
  `(1,8)`→`(13,10,E)` handoff；State adapter 投影 GEO set/block、wall／roof
  與 `C04B..C04F`。這筆轉移是 `strong inference`；DOS ECL work address 與
  overlay 22 `[di+4BF0h]` indexed table 的 writer→projection→consumer 尚未
  閉合，不能把兩者當成同一位址空間。
- `Select` 不得在 `CombatRequested` 時先套用 data-pack event；拒絕投降必須先
  建立五名 Fire Knife 戰鬥，只有戰勝後的 PRESS continuation 才能觸發位置事件。
- 第 515 輪已把測試中第一個直接座標注入移除。第二個 `(8,15)`→block 4
  外部 handoff、騎士後續、完整下水道與整作正常玩家路徑仍未完成；不可因本輪
  regression 通過宣稱完整 remake。權威規格為
  `docs/spec/515-fire-knife-map-position-transition.md`，engine 知識庫為
  `golden-box-remake-engine/docs/knowledge/golden-box-map-position-transition.md`。

### 第 514 輪提爾佛頓下水道入口正常路徑補充

- block 3 入口回返 `(0,1,S)` 到火刀檢查站 `(1,8,S)` 必須逐格呼叫
  `State.MoveDungeon`：`E、S×3、W、S×2、E、S×2`。途中 `(1,6)` terrain
  `0x8D` 與終點 `(1,8)` terrain `0x81` 是 GEO cell 識別，不得直接把 terrain
  數字命名成劇情旗標。
- 最後一步的 per-turn 結果會先產生
  `tilverton.sewers.guild-battle-echoes` PRESS。它含 meaningful text 時，
  lifecycle 不會在同一次呼叫中假設 SearchLocation 已完成；按下
  `ecl-option.press-button-or-return-to-continue` 返回地城後，玩家再以正式
  `SearchDungeonLocation` 觸發檢查站。不能把 PRESS 與檢查站 menu 合併，或重跑
  block 起點掩蓋 pending continuation。
- 檢查站拒絕分支保留五名 `FIRE KNIFE` 的真正 `CombatAct` 回合與
  `tilverton.sewers-hide-bodies` 戰後 pause；文字與期待值走 game-pack stable ID，
  不在測試複製繁中顯示字串。
- 第 514 輪原本的火刀戰後 `(13,10)` 騎士座標輔助已由第 515 輪
  `set_map_position` event 取代；該外部 handoff 仍標為 `strong inference`。
  block 3 在關閉狀態 movement predicate 下沒有合法路徑，不能用 BFS 穿牆假造
  相連；這不判定地圖永久分成兩個 component，騎士後到 block 4 的第二個 handoff
  仍須追 ECL／NEWECL／邊界 evidence，不得再新增直接座標注入。
- 權威規格為 `docs/spec/514-normal-tilverton-sewer-entry-route.md`，知識庫為
  `docs/knowledge/golden-box-normal-player-path.md`。本輪是 `READY` 有界路徑，
  不得擴大成完整下水道、完整戰鬥、完整中文化或整作通關。

### 第 516 輪反組譯盤點與斷言清理補充

- 剩餘反組譯工作以 `docs/knowledge/golden-box-reverse-engineering-worklist.md`
  為權威路由。P0 是火刀戰後第一個外部地圖 handoff 的 DOS 目的地 producer→
  `CALL 2E10h` redraw／位置 consumer、騎士後 `(13,10)` 到 E2 `(8,15)` 的
  secret-door／SEARCH bridge，以及 block 4 後續外部地圖／返回世界 handoff；不要
  以「READY spec 數量」估算完成率。
- PC-98 `PC98-GAME.EXE` SHA-256
  `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` 的
  overlay-local 證據：`MOVEMENT` `PREMOVEPARTY` 的 `S` 只切換目前角色 record
  `+594h` bit 0，再呼叫 `014A:00DE` TPOV stub；typed resolver 對應 overlay 24
  entry 38、handler `2E8Ch` `SHOWLOCATION`。目前沒有 S→第三平面 writer 的閉合證據，
  不能把攻略 `~` 或 wall detail 0 當作可走秘密門。
- `BLOCKCODE` `017C:04DE` 已 exact 證明 wall type 09/detail 0 不是普通可走格；
  P／K／B action 與 detail=1 的雙側門狀態另有明確 writer。這些 movement bytes
  只能描述 PC-98；不能把同數值映射成 DOS `4C28h` 或 `4BF0h` consumer。
- DOS extracted overlays 的 raw little-endian 掃描未找到 literal `4C28h`，只表示
  沒有直接 literal 命中，不表示 `4C28h` 沒有經指標／通用 interpreter 使用。這項
  負面結果不可寫成「未使用」。
- 已移除錯誤位址基準的 `pc98_generic_local_audit.idc` 與重複 candidate／xref
  誤導風險高的 `dos_map_workcell_audit.idc`；保留的 `pc98_overlay14_pre_move_audit.idc`、
  `pc98_overlay24_generic_audit.idc`、`pc98_overlay2_search_state_audit.idc`、
  `pc98_overlay7_specials_audit.idc` 與 `dos_map_workcell_raw_audit.idc` 只輸出
  可回查的 raw bytes／候選，不能自動升格語意。
- `docs/spec/297-fire-knife-hideout-transition.md` 的舊「formal regression crosses
  (8,15)」斷言已移入勘誤並標成 `SUPERSEDED`；`PLAN.md` 的 block 4 完成項也已
  改為未完成。compact 後不得恢復這兩個錯誤結論。

### 第 517 輪逆向缺口盤點與 GEO 斷言勘誤

- 目前以資料流計算，仍有 11 個直接影響正常玩家結果的逆向主題：P0 外部地圖／
  正常路徑 3 個、P1 ECL／外部 routine 4 個、戰鬥規則／AI 2 個、存檔／AD&D
  角色規則 2 個；另有 4 個 fidelity／音訊／發行主題。這不是 binary 行數、
  函式數或 READY spec 數量的完成百分比。
- 第 517 輪把「GEO component 永久不相連」降級為正確的「GEO2 block 3 在目前
  第三平面／門狀態未變更時，movement graph 沒有合法路徑」。診斷器把
  `wall=09/detail=0` 邊暫時視為開啟後找到的路徑，只是秘密門候選，不能直接
  寫入 movement 或 JSON。
- PC-98 `S` 仍只精確到 record `+594h` bit 0／`SHOWLOCATION`；`BLOCKCODE`
  只證明 detail 0 不是普通可走格。CoAB `SearchDungeonLocation` 的 `PRESS`
  是 remake entry observation，不是原版秘密門 writer 證據；`workplace/`
  probe 不得移入正式 regression。
- 下一步先追 DOS ECL2 block 3 `CALL 2E10h` 的 consumer 與前置目的地 producer，
  並分開追 ECL work `4BF0h／4BF1h` 與 overlay 22 `[di+4BF0h]` indexed table
  的 projection；只有資料流閉合後才能建立中立
  `secret_door`／`search` JSON contract。完整盤點見
  `docs/spec/517-reverse-engineering-gap-inventory.md`。

### 第 518 輪 DOS `CALL 2E10h` 位址空間稽核

- IDA Pro 9.4 在 `START.EXE.i64` 的 disposable copy 中逐一掃描所有 resident
  segments；`LE16=2E10h` 唯一命中為 `seg043:0x6634`／IDA EA `0x16634` 的
  `db 16,'. Check install.'` 非 code 字串，對 `0x2E10` 與 MZ base 換算的
  `0x12E10` 都沒有 direct code xref。這是位址空間排除，不是「CALL 沒有
  handler」的結論。
- 不得把這個 raw 字串命中當成 `sub_2E10`，也不得因此把 overlay／ECL／far
  pointer 間接路徑刪掉；`CALL 2E10h` 仍要追 ECL／overlay dispatch 或 map
  service 的 producer→projection→consumer。完整 hash、raw bytes、工具版本與
  可重現腳本見 `docs/spec/518-dos-start-ecl-call-address-space-audit.md`。
- 抽出的 `GAME.OVR overlay-02` 在 local `0x2F23` 執行 `sub ax,7FFFh`，
  local `0x2F2C` 比較 `AE11h`，local `0x2F39` 以 raw far pointer
  `017F:003Eh` 呼叫外部 target。這只證明 ECL operand 的 dispatch normalization
  與 call-site；`017F:003Eh` 的 module／重定位／consumer 尚未閉合。不要把
  `overlay-02` 與 public `ovr003` 編號直接合併。
- 第 518 輪沒有減少 11 個行為逆向主題或新增 secret-door 規則；下一步仍是
  P0-1 的間接 dispatch／runtime trace，所有 `unknown／hypothesis` 仍須維持原
  推論等級。

### 第 519 輪 DOS overlay vector → cell-layer 靜態邊界

- 第 519 輪以 Docker 內 `ida-pro-9.4-ver2:uidfix-v1` 的 IDA Pro `9.4.0.260610`
  重新核對 `START.EXE` MZ control block 與 `GAME.OVR` raw offset。`START.EXE`
  SHA-256 為 `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5`，
  baseline `START.EXE.i64` 為
  `9df802ee4ef71fb2eda83257e0ed2d87adf0ee2d10241d3bdbdc6bc369fe47eb`，完整
  `GAME.OVR` 為
  `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215`。
- MZ header `0x7B0`、image paragraph `017Fh` 對到 raw control block `0x1FA0`；
  `+04=0x3DF87`、`+08=0x147F` 與 extracted overlay-30 的 raw offset／長度
  吻合。vector table 從 control `+20h` 開始，每筆 5 bytes；`017F:003Eh`
  是 zero-based vector 6，raw `CD 3F C6 07 00`，目標為 overlay-30 local
  `07C6h`。
- **位移勘誤固定規則：** 相鄰 vector 7 的 raw `CD 3F 41 08 00` 才指向
  `0841h`；`0841h` 不是 ECL `2E10h` 的 target。若 compact 後看到交接摘要把
  `0841h` 當成 vector 6，必須回到 raw vector table，不得沿用。
- overlay-30 local `07C6h` 的 exact static boundary 是：兩個 word 參數、
  `0..0Fh` bounded 16×16 index、`DS:7206h` far pointer、
  `ES:[DI+0200h]` byte read 與 `retf 4`。若舊臨時筆記寫成 `retf 6`／`+0100h`，
  以 raw bytes 與 `docs/spec/519-dos-overlay-vector-to-cell-layer-accessor.md`
  為準；該錯誤沒有進正式 engine／JSON／regression。
- 這只把「ECL selector → control vector → overlay routine」靜態邊界閉合；
  `DS:7206h` owner／初始化、`DS:720F／7210／7213` writer／consumer、
  `C04B..C04F` projection、map plane、runtime redraw 與 secret-door 語意仍是
  `unknown` 或 `strong inference`，不得新增 movement 特判。11 個行為主題與
  4 個 fidelity／發行主題數量不變。
- 版本化工具為 `scripts/ida/dos_overlay30_vector_audit.idc` 與
  `scripts/ida/dos_start_dseg_offset_audit.idc`；只操作 disposable database，
  不改原始 binary／baseline `.i64`。完整 hash／report 見 spec 519。

### 第 520 輪 DOS movement／overlay cell-layer bridge

- `overlay-07` local `1B3Fh` 的 raw／IDA code exact：讀 `DS:7211h` 的
  `0／2／4／6`，把 `DS:720Fh／7210h` 各自以 `0..0Fh` 循環更新；接著
  `call far 017F:003Eh`，AL 寫入 `DS:7213h`，再以三參數
  `call far 017F:0034h`，AL 寫入 `DS:7212h`，最後 `DS:8B68h=1`。此 overlay
  對 local `1B3Fh` 的 direct IDA code xref 為 0，不能因此宣稱正常輸入已接通。
- `overlay-11` local `00E9h` 把 `DS:7206h` 與 `0400h` 傳給
  `0A54:0329h`；`03F2..03FCh`／`078Eh..0798h` 寫入 `DS:720F=07h`、
  `DS:7210=0Dh`、`DS:7211=00h／02h`。這是 writer／call-site exact；第 521 輪
  已閉合 callee 為 Borland `GetMem(Pointer &,Word)`，配置後 buffer 語意仍未知。
- `overlay-30` vector 4 `017F:0034h → 06BDh` 讀 `+000h／+100h` 的 high／low
  nibble；vector 6 `017F:003Eh → 07C6h` 讀 `+200h` byte。`overlay-28:00CCh`
  呼叫 vector 6 並比較 `DS:7213h` 與 `7Fh`，稍後另呼叫
  `017F:0043h → overlay-30:0841h`。因此 `0841h` 是 vector 7 的獨立 routine，
  永遠不能當成 vector 6／ECL `2E10h` target。
- `overlay-14` control segment `00D2h` 的 vector 4 `003Eh` 另有
  `ES:[DI+300h]` mask read/write（`FC／F3／CF／3F`）；不能與 overlay-30
  `+000／+100` 或 ECL `4BF0h` 跨位址空間合併。
- 本輪只縮小 P0-1 的 DS／vector bridge；11 個行為主題與 4 個 fidelity／發行
  主題數量不變。不得新增 secret-door／search JSON、movement 特判或把
  `DS:7212／7213` 命名成正式規則。完整 report／hash 見
  `docs/spec/520-dos-movement-to-overlay-cell-layer-bridge.md`，腳本為
  `scripts/ida/dos_overlay07_movement_audit.idc`、
  `scripts/ida/dos_overlay_ds_field_audit.idc`、
  `scripts/ida/dos_overlay_ds_context_audit.idc`。

### 第 521 輪 DOS `0A54:0329h` GetMem buffer owner

- Docker 內以 `ida-pro-9.4-ver2:uidfix-v1`／IDA Pro `9.4.0.260610` 分析原始
  `START.EXE` 與唯讀 baseline database。`START.EXE` SHA-256 為
  `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5`，baseline
  `START.EXE.i64` 為
  `9df802ee4ef71fb2eda83257e0ed2d87adf0ee2d10241d3bdbdc6bc369fe47eb`。
- segment inventory 顯示 runtime selector `0A54h` 在此 IDA baseline 的
  `+1000h` paragraph mapping 下對到 IDA selector `1A54h`；runtime
  `0A54:0329h` 是 `seg050`／IDA EA `0x1A869`。入口 raw／IDA bytes、原始
  Borland symbol `@GetMem$qm7Pointer4Word; GetMem(Pointer &,Word)` 與 direct
  callers 見 `docs/spec/521-dos-getmem-buffer-owner.md`。
- 因此 `overlay-11:00E9h` 的 `DS:7206h`＋`0400h` call-site 已可標為由
  `GetMem` 接收 1 KiB size 的 pointer-buffer owner `strong inference`；但這
  不得把 buffer 命名成 GEO、wall、terrain 或 secret-door plane。
- 在第 521 輪當時仍未知的是配置後 buffer writer／清除時機／
  `+000/+100/+200/+300` plane layout、overlay-07 `1B3Fh` 的正常或間接 entry、
  `DS:7212／7213` 與 vector 7 `0841h` consumer、`C04B..C04F` projection，以及
  DOSBox runtime handoff；第 522 輪已補上其中的 writer 幾何與暫存 pointer 回收，
  第 523 輪再補上 control vector／靜態 dispatcher entry。
- 版本化腳本為 `scripts/ida/dos_start_segment_inventory.idc` 與
  `scripts/ida/dos_start_0a54_call_audit.idc`；只操作 disposable database，
  不改原始 binary／baseline `.i64`。本輪沒有修改 engine、JSON、movement
  predicate 或 secret-door 規則；11 個行為主題與 4 個 fidelity／發行主題不變。
- 不得再把 `0A54:0329h` 當成「callee 完全未知」；也不得因 GetMem 名稱把
  `DS:7206h` 直接升格成地圖 buffer、GEO plane 或已完成 map handoff。

### 第 522 輪 DOS `DS:7206h` 四平面 writer／暫存 loader

- Docker 內以 `ida-pro-9.4-ver2:uidfix-v1`／IDA Pro `9.4.0.260610`，在
  `START.EXE`／`overlay-30.bin` 的 disposable database 上重新輸出連續 raw bytes、
  原始 IDA symbol、function end 與 direct caller；原始 bytes、baseline `.i64`、
  overlay-local offset、resident EA 與 runtime selector 必須保持分開。完整來源、
  hash、報告與工具見 `docs/spec/522-dos-buffer-four-plane-fill.md`。
- `overlay-30:133Ah..1475h` 的四次 resident `@Move$qm3Anyt14Word` call 是
  `exact`：來源分別為暫存 pointer `+002h/+102h/+202h/+302h`，目的為
  `DS:7206h +000h/+100h/+200h/+300h`，每次 `0100h` bytes。最後同一暫存
  pointer 與返回 word 傳給 `@FreeMem$qm7Pointer4Word`，所以暫存配置的回收也是
  `exact`；不得再把「writer／四平面幾何完全未知」當成目前斷言。
- runtime `0636:08DEh` 在此 baseline 的 IDA EA `0x16C3E` 原始名稱仍是
  `sub_16C3E`；function dump 可證明 `BlockRead`、暫存配置、資料轉換／展開候選
  與暫存釋放；第 524 輪已由 overlay-30 `GEO`／`.dax` source fragments、
  `0402h` gate 與 GEO corpus 補上 decoded record／四平面 payload 對照，但
  `DS:5BEEh` 正式欄位、selector producer 與 runtime map projection 仍是
  `strong inference／unknown`。前置 `Store string`／`Concat` 的
  stack cleanup 與 target `retf 12h` 閉合；不能因表面三組 push 誤判成另一個
  routine，也不能因 `.dax` literal 直接命名 GEO block。
- 本輪後仍未閉合 overlay-07 `1B3Fh` 的普通鍵盤 producer／control-loader runtime、
  `DS:7212／7213`／vector 7 `0841h` consumer、`C04B..C04F` projection 或 DOSBox
  runtime handoff；第 523 輪已補上其 control vector／靜態 dispatcher entry。11 個
  行為主題與 4 個 fidelity／發行主題數量不變。沒有修改 engine、JSON、movement
  predicate 或 secret-door 規則。
- 版本化腳本為 `scripts/ida/dos_overlay30_buffer_copy_audit.idc` 與
  `scripts/ida/dos_start_buffer_routines_audit.idc`；只操作 disposable database，
  不改原始 binary／baseline `.i64`。第 522 輪後優先追 `.dax` record／解碼輸出與
  正式 plane consumer，再追正常 input／欄位 projection／DOSBox trace；第 524 輪
  已補上 GEO source／decoded payload，後續不重做同一份 DAX header／plane scan。

### 第 523 輪 DOS control vector 26／overlay-07 靜態 dispatcher entry

- Docker 內以 `ida-pro-9.4-ver2:uidfix-v1`／IDA Pro `9.4.0.260610`，在
  `START.EXE.i64`、`overlay-02.bin`、`overlay-07.bin` 的 disposable copy 上保留
  raw file offset、runtime selector、IDA EA 與 overlay-local offset 分離；完整
  provenance、hash、bytes 與信心等級見 `docs/spec/523-dos-overlay07-vector26-entry.md`。
- `START.EXE` MZ header 是 `0x07B0`；control block raw `0x0E60` 對應 runtime
  selector `006Bh`，vector table 從 `006B:0020h` 開始、每項 5 bytes。vector 26
  位於 runtime `006B:00A2h`／raw `0x0F02`／IDA EA `0x10752`，bytes
  `CD 3F 3F 1B 00`，target 是 overlay-07 local `1B3Fh`。這是 `exact` 的
  control-vector／local-target 對照，不是推測性 rename。
- overlay-02 local `2FFD..3007` 的連續 bytes 是 `cmp ax,401Fh`、成功後
  `call far 006B:00A2h`；因此 `401Fh` branch 到 vector 26 的靜態 caller 也是
  `exact`。既有 ECL `C01Eh` 對照仍是 `strong inference`，不能把 `401Fh` 單獨
  寫成普通鍵盤「向前」事件。
- overlay-07 自身的 `direct_cref_count=0`、`raw_LE16_target_1B3F_count=0` 只
  適用 local code／literal 有界掃描；它不能再被寫成「整個程式沒有 caller」。
  間接 audit 只列真正 `FF` memory-form call／jump；loader 以 register／pointer
  形成的 runtime call 仍未知。`DS:7212／7213` 正式 consumer、正常輸入 producer、
  `C04B..C04F` projection、目的地／座標／方向／重繪仍未閉合。
- 本輪沒有修改 engine、CoAB JSON、movement graph、`set_map_position` adapter 或
  secret-door 規則；11 個行為主題與 4 個 fidelity／發行主題不變。下一個最小工作
  是追 `006B` loader／DOSBox 實際鍵盤 trace，再接 `.dax` record、欄位 consumer
  與 map projection，不可用 direct-entry 取代正常路徑。
- 版本化工具為 `scripts/ida/dos_start_overlay07_vector26_audit.idc`、
  `scripts/ida/dos_overlay02_call_dispatch_audit.idc` 與
  `scripts/ida/dos_overlay07_indirect_dispatch_audit.idc`；原始 binary／baseline
  `.i64` 未被修改。

### 第 524 輪 DOS overlay-30 `GEO<區域>.dax` loader／四平面來源

- Docker 內以 `ida-pro-9.4-ver2:uidfix-v1`／IDA Pro `9.4.0.260610` 分析 extracted
  `overlay-30.bin` 的 disposable database；原始 overlay、START／GAME.OVR 與
  baseline `.i64` 均未修改。完整 provenance、hash、raw bytes、推論分層見
  `docs/spec/524-dos-overlay30-geo-loader-source.md`。
- overlay-30 local `1310h`／`1314h` 的 raw Pascal fragments 是 `03 'GEO'` 與
  `04 '.dax'`；local `1341h` 讀 `DS:5BEEh`，以 width `1` 呼叫 resident
  `0A54:12ABh` numeric-to-string helper，`1361h`／`137Bh` 再以 resident
  `Store string`／`Concat` 組合。完整 `GEO<區域值>.dax` 路徑仍因缺 runtime
  file-open trace 標為 `strong inference`；不得把 `DS:5BEEh` 直接改名成已證實
  的 global area 欄位。
- local `1385..1393h` 將 `[bp+6]` selector、output word／pointer 傳給
  `0636:08DEh`；`1398..13A3h` 只讓 decoded size `0402h` 進入成功分支。既有
  DOS archive 的 `GEO2.DAX` block 1／3／4 都是 raw `1026=0402h`，GEO3–GEO6
  也相同；output `+002h` 後四段各 `0100h` 正好對上既有 round 56 parser 的
  four-plane layout。此來源／payload 格式對照達 `exact`，不是新猜的地圖規則。
- 第 524 輪清除「此 loader 的 DAX record／0402h／四平面格式完全未知」的現行
  斷言；仍未閉合 `[bp+6]` selector producer、`DS:7206h`／`DS:7212／7213` 正式
  consumer、`C04B..C04F` projection、普通鍵盤／control-loader runtime 與
  DOSBox 地圖 handoff。沒有修改 engine、JSON、movement graph 或 secret-door。
- 版本化工具為 `scripts/ida/dos_overlay30_geo_loader_audit.idc`；本輪不重做已由
  round 56／spec 524 閉合的 GEO DAX header／plane parser，下一步只追 selector→
  consumer→runtime。

### 第 525 輪 PC-98 `TEMPSEARCH/BDF1` 邊界（2026-08-09）

- PC-98 `PC98-GAME.EXE` 的 Borland symbol table 以 `TEMPSEARCH=0C29:BDF1h`
  正式命名該欄位；overlay-02 local `3BB8h..3BFDh` 的連續 bytes 會從目前角色
  `+594h` 讀值、`AND FFFDh` 後寫入 `BDF1`，暫時把 `+594h` 設成 `1`，再從
  `BDF1` 還原並呼叫 far stub `014A:00DEh`。TPOV resolver 將 stub 對到
  overlay-24 `SHOWLOCATION` local `2E8Ch`。這是 `+594h` 暫存／顯示狀態 bridge，
  不是已證實的秘密門第三平面 writer。
- overlay-11 local `06BEh..06CDh` 的 `C606F1BD00`、`F2BD`、`F3BD`、`F4BD`
  只是 `BDF1..BDF4` 初始化清零；resident `.i64` 對 `BDF1` 沒有 direct dref
  只表示 overlay／間接讀寫不在 resident graph 中，不能寫成「BDF1 沒有 consumer」。
- 不得再把 `S`→`SHOWLOCATION`、暫時 `+594h=1`、Borland `SEARCHREC` type、
  `SECRET` 或 `HIDDEN` 名稱當成秘密門已開、第三平面 writer 或 `wall=09/detail=0`
  可走。第 525 輪已閉合 `LOAD3DMAP (017C:1253h)`／`BLOCKCODE
  (017C:04DEh)` 的 loader／普通 reader；第 526 輪又確認 `SEARCHREC` 是 DOS
  檔案搜尋記錄，不是地圖 writer。P0-2 若仍阻塞，只追 `MOVEPARTY
  (00C9:0BCCh)` 的 action trigger、狀態持久化與 movement consumer；不能重做已
  閉合的 loader／普通 reader，也不能把未驗證的欄位直接寫成秘密門規則。在
  bridge 與 runtime trace 閉合前，`MoveDungeon` 必須維持 fail-closed。
- 完整來源／hash／命令見 `docs/spec/525-pc98-tempsearch-display-state.md`；
  新增非破壞性工具 `scripts/ida/pc98_search_state_xref_audit.idc` 與
  `scripts/ida/pc98_load3dmap_dataflow_audit.idc`。本輪沒有修改
  engine、CoAB JSON、movement graph 或 secret-door 規則；11 個行為主題與 4 個
  fidelity／發行主題數量不變。

### 第 526 輪：以可玩性界定反組譯範圍（2026-08-09）

- 反組譯不是逐行翻譯整個 executable。每次只從一個實際玩家阻塞點開始，確認
  它對應的輸入、資料流、狀態變化與可重播結果；一旦證據足以實作並測試該遊戲
  功能，就停止對同一區域的無關 function 深挖。
- 與遊戲可玩結果無關的檔案服務、工具程式、初始化樣板或未被路徑使用的 helper
  只保留最小位址／用途分類，不列入 P0 worklist。若未來新功能真的需要它，才
  重新開一個有明確玩家結果的窄任務。
- `SEARCHREC` 的檔案搜尋結構與 `MOVEPARTY` 的 map-buffer 靜態 writer 只作
  錯誤路由修正；在沒有玩家路徑、runtime trace 或資料流 consumer 需求前，不再
  追查其餘 PC-98 檔案服務。完整邊界見
  `docs/spec/526-pc98-moveparty-map-writer-searchrec-correction.md`。
- 任何尚未完成的遊戲功能都必須回到 engine＋CoAB JSON 的正常玩家交易；不能以
  direct-entry、座標注入或未驗證的反組譯名稱冒充可玩完成。反組譯深度降低不會
  降低原始位址、bytes、位址空間與推論等級的保存要求。

## 10. Compact 後恢復工作清單

1. 讀本檔、`CLAUDE.md`、`docs/project-status.md` 與 `CONTEXT.md` 尾端。
2. 檢查兩個 repo 的 status、diff、HEAD、remote，完整保留 dirty changes。
3. 查看目前 READY spec 與最後一次 player-path test。
4. 不重做已 push 的 milestone，不沿用已 supersede 的推論。
5. 若本節「未提交 milestone」仍存在，先延續它；完成後再把成果移入
   `CONTEXT.md` 並刷新本節。
6. 更新計畫，繼續縮短完整可通關與完整中文化的真實差距。

## 11. 第 541 輪交接：外部 routine 與世界旅行

- 共用 engine 只保存 ECL 的有序 raw／typed boundary、續跑、資源 selector、map
  graph 與事件契約；`0x2E10`、`0xC01E`、`0xB200`、`PROGRAM` caller context、
  CoAB work address 與劇情旗標不得因本作證據直接變成跨作品 engine API。
- `arriveAtWorldLocation` 的荒野抵達交易必須在 ECL1 arrival entry 前後提交
  native destination `4C9B`；`4C9C` 是本作 dispatch selector。兩個位址都是
  CoAB 位址空間，不能在 engine 文件改名成通用欄位。
- `TestRealOverlandArrivalAndRouteGraphCoverage` 只證明 14 個世界點位到達與
  JSON directed adjacency 可達；它不證明所有城市／地城事件、房間、隨機遭遇、
  出口、重訪或整作通關。事件仍要由正常 ECL session 驗證。
- 交接入口：`docs/spec/541-ecl-external-routine-engine-boundary.md`、
  `WORKLIST.md`、`CONTEXT.md` 尾端。若沒有第二款 SSI 遊戲的 producer→consumer
  證據，不要新增 external-address mapping 到獨立 engine。

## 12. 第 558 輪交接：ECL 候選審查與位址勘誤

- 第 557 輪列出的三組 `TREASURE → COMBAT` 不是現行 blocker。第 255／257／258
  輪已完成 raw treasure、戰鬥前 pending、勝利後 loot 與下一條 PC continuation；
  第 558 輪再以 PC-98 IDA 與三段 DOS 真實 DAX regression 補強。不要 compact 後
  重寫同一 transaction；下一個 P0 是 ECL2 block `02h +04BC..+053A` 的
  `COMBAT → text`。
- `docs/audit/ecl-event-catalog.json` format v2 的 candidate ID 由
  `member/block/start-end` 組成；人工語意只能放
  `docs/audit/ecl-ordered-effect-reviews.json`。ledger 中不存在於目前 corpus 的 ID
  必須失敗即關閉，不可模糊比對或依 opcode 相似度套用舊結論。
- Borland symbol 的 `segment:handler offset`、resident `CD 3F` stub offset 與
  overlay-local code offset 是三種位址空間。PC-98 `INTERPET.GOECL=0037:3A21`、
  `POSTCOM.DOPOSTCOMBAT=0057:1775` 的 offset 是 handler-local；typed resolver 的
  `overlay 5 stub 0025h → code 1775h` 才是 stub→handler bridge。不得把 handler
  offset 再拿去當 stub 查詢。
- ECL4 block `25h` 戰後 `PICTURE FFh` 是清除圖片，不是新的 picture request；
  continuation 是 `COMBAT +12A2 → +12A3 → GOTO +1529 → CALL 2E10h → EXIT +1534`。
- 權威規格：`docs/spec/558-pc98-ecl-treasure-combat-boundary.md`；第 557 輪原始
  優先排序已在原文件追加勘誤，不能只讀舊表格第一列。

## 13. 第 559 輪交接：全模組全掃與函式覆蓋台帳

- **先盤點、再語意**。使用者 2026-08-13 指示先徹底完成反組譯分析，避免邊做
  邊分析。驗收口徑見 §2.5；方法與基線見
  `docs/spec/559-full-module-re-sweep.md`。
- 第 559 輪的起點是「DOS 1,344 ＋ PC-98 1,481 共 2,825 個函式**全部 `待解讀`**」
  ——刻意的誠實起點，不用關鍵字比對把它調低。**現在 `待解讀` 已經是 0**
  （2,874 個函式：已解讀 2,137／不阻塞 162／邊界碎片 575），數字一律以
  `cmd/re-ledger` 現跑的輸出為準，不要引用這裡或任何文件裡的快照。進度只透過
  `docs/audit/re-function-ledger.json` 的明確記錄變動，改完重跑 `cmd/re-ledger`。
- **不要重跑一次性 IDC 去回答「這個 overlay 有什麼」**。全掃結果已在
  `workplace/re-sweep/<平台>/out/<模組>.json`：函式、chunk、xref、字串、
  具名資料、未定義區都在裡面。要逐指令證據用 `tools/ida/dump_function.py`。
- 重建全掃：`tools/re-sweep.sh dos|pc98 <exe> <ovr>`（約各數分鐘，JOBS 預設 4，
  這台機器同時有別的工作，超過一半核心不要用）。
- overlay 對應：PC-98 的 36 個原始 Turbo Pascal 單元名見 spec 559 表；29／36
  段的 entry_count 與 DOS 相同。這只是 **module 級對應假說**，個別函式位址
  必須各自證明，不得把 PC-98 符號名直接寫成 DOS 的事實。
- ECL opcode → handler 全表**已完成**（spec 560；DOS 那份在
  [`docs/audit/ecl-opcode-handlers-dos.md`](docs/audit/ecl-opcode-handlers-dos.md)，
  分派器 `overlay-02:3377h` 逐條讀出 52 個 opcode，含每支的位址範圍與指令數）。
  它現在的用途是**排順序**：剩下的 `partial` opcode 一支一支收，成本差一個數量級
  的先做。⚠ 巢狀條件裡的 opcode 當初是讀完控制流才定案的，不是文字比對——後續
  要擴充這張表也一樣不得硬湊。
