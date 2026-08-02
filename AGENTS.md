# Codex working agreement — Curse of the Azure Bonds 中文化／Remake

本檔是 compact／交接後的第一閱讀入口，也是 agent 工作規則的單一權威來源。
它已融合 `CLAUDE.md` 中仍有效的原始需求；後者只保留人類可讀的目標與資料
索引，不應再複製易過期的 checkpoint。詳細歷史在 `CONTEXT.md`，可驗證完成度
在 `docs/project-status.md`。若文件衝突，以目前 worktree、原始 bytes／實機
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
- direct-entry、直接設定 PC／旗標或注入戰鬥只適合縮小問題；完成驗收必須
  再從正常地圖移動、事件互動與戰後續跑抵達同一狀態。否則只能聲稱局部
  opcode／事件測試通過。
- 看見第一場戰鬥或第一段文字不代表事件完成。必須追蹤戰後 ECL continuation、
  後續戰鬥、旗標寫入、地圖持久狀態、Journal／獎勵及離開再進入的結果。
- 不得把解碼器「可以讀」誤寫成遊戲「已可玩」，也不得把靜態數學核心、短秒
  trace 或單張畫面擴大宣稱為 scheduler、完整動畫、音訊播放或完整玩家路徑。
- 若新 bytes、IDA 或 runtime 證據推翻舊文件，必須在同一 milestone 更新或
  supersede 舊 spec、README、狀態表與 `CONTEXT.md`；不能同時留下互相矛盾
  的「完成」敘述供 compact 後誤用。

## 4. 可用原始資料與工具

- DOS game image：`curseoftheazurebonds.zip`
- 傳統中文資料：`珍020-青色枷的詛咒.rar`
- Manual：`Curse-of-the-Azure-Bonds_Manual_DOS_EN.pdf`
- Adventurer's Journal：
  `Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf`
- Clue Book：`Curse-of-the-Azure-Bonds_Misc_DOS_EN_Clue-Book.pdf`
- 工作資料：`workplace/`
- IDA Pro：`/home/anr2/ida_94_official/dist`
- 倚天字型：`/home/anr2/cht/etan_font/stdfont.15`
- 倚天粗體參考：`/home/anr2/scummvm/monkey_island2`
- 其他 remake 參考：`/home/anr2/cht/daemon_winter`

允許使用 Docker 內 DOSBox／Xvfb 取得原版與 remake runtime 截圖。原版
executable 是行為 oracle；網路資料不能取代可取得的本機實機驗證。

凡有反組譯（reverse engineering）需求，優先在 Docker 隔離環境內使用
`/home/anr2/ida_94_official/dist` 的 IDA Pro。其他反組譯器、反編譯器或
自製掃描工具只能作補充與交叉驗證，不得在 IDA Pro 可用時逕自取代主要分析
流程；IDA 結論仍須依本文件第 3 節，以原始 bytes、runtime trace 或另一項
權威證據交叉驗證。

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
- 目前 `ida-pro-9.4-ver2` image 的 IDAPython 會受主機 Python 路徑影響；本專案
  已驗證 IDC 可用。headless 稽核腳本必須把結果寫入明確檔案並檢查內容，不能
  只依賴 `Message()`／stdout 或 exit code 0 判定成功。
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

- 只有重大、已測試、可展示的 milestone 才集中 commit＋push；不要每個小改
  都提交。
- 兩個 repo 各自 commit／push，歷史保持獨立。
- CoAB 使用：
  `git --git-dir=/tmp/azure-bonds-git --work-tree=.`
- Engine 使用：
  `git -C golden-box-remake-engine`
- 不丟棄使用者或不相關變更；先檢查 dirty worktree。
- compact 後若看到 probe、暫存 regression 或未完成 spec，先讀 diff 與
  `CONTEXT.md` 尾端；它們可能是正在累積的 milestone，不可因尚未提交而刪除。
- 每個 milestone 更新 README、`docs/project-status.md`、READY spec／知識庫
  與本檔或 `CONTEXT.md` 的延續資訊。

## 8. 驗證與完成門檻

開發中跑 focused tests。提交前至少：

- affected repo 的正式套件測試；CoAB 目前用
  `go test ./cmd/... ./gamepack ./internal/...`，Ebiten package 在
  Docker/Xvfb 驗證。`go test ./...` 會因 `scripts/` 兩個獨立 main 同目錄
  的既存結構失敗，修正該 gate 前要如實分開報告。
- `git diff --check`。
- deterministic screenshot 或 runtime trace。
- 真正玩家路徑驗證，而不只 direct-entry debug flag。
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
- CoAB 本輪基底：`ea01be8`（第 437 輪 `INARC／OBJECTLIST`）；第 438 輪
  `CHARACTERLIST／IDLIST` stable identity milestone 會由本文件所在 commit
  完成。
- Engine dependency：`f3c652a`（含作品中立 game-pack
  `character_creation.templates` schema／validation、繁中角色建立知識庫，
  以及 YM2203 opaque full-state／PCM
  resampler snapshot，以及 `combat/effecttime`、`combat/scan` 的
  `TDEFTYPE／TACTICALMAP` 地形視線、footprint producer 與 `INARC` 八方向
  inclusive 扇區、`combat/scanorder`
  的三 byte record 原版排序、`combat/sleep` 的 `4d4`／
  ordered HD capacity filter、`combat/action` 的 delayed
  spell `TargetID`／point-target transaction 與 interruption clear、
  `combat/initiative`、
  `combat/quickspell`、`randomstream`，以及 game-pack
  `presentation.scene_character` native geometry、繁中人物版面知識庫、
  `combat_modifiers`、
  signed low-byte decoder、繁中 ECL 戰鬥修正知識庫、`option_rules`、世界目的地
  有向圖 schema／validation、
  繁體中文音訊架構知識庫及中立
  `audio/cyclepcm`、`audio/s98`、`audio/ym2203`、
  `audio/pc98soundbios`、
  `combat_visuals`、
  `music_tracks`／`music_bindings`／`music_cues` 與跨 locale
  `title_id` schema）。
- 本文件所在 commit 會晚於上述 CoAB 基底；compact 後永遠先以兩個 repo 的
  實際 HEAD／remote 為準，不要把文件內 hash 當成可自我引用的 latest hash。
- GUI 原版石框、人物／3D／PIC 分離舞台、16×15 倚天與 PC-98 typography
  study 已完成並 push；角色資訊全頁與所有畫面逐張 fidelity 尚未完成。
- 專案仍是多 vertical slices prototype，尚未完整可通關。
- 主要缺口見 `docs/project-status.md`：完整 ECL/external routines、開場到結局
  玩家路徑、戰鬥規則/AI/法術/戰後、全地圖、全翻譯、音樂音效、完整 save、
  三平台發行與長時間回歸。
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
  `45h` 只對 MonsterType `13h` observer 生效，`18h` 可取消 hidden 但不取消
  `-4`。`+11Ah` 又由 dragon-slayer `03h` consumer 與四筆真實 Animal records
  關閉為 MonsterType。READY spec 418 是 visibility／effect consumer 的權威。
- 第 419 輪已由 PC-98 overlay 24 `DEXRABONUS 1416h`、overlay 13 local
  `0000h` 與 overlay 8 local `01FBh` 關閉 initiative writer／TeamList selector。
  原版使用 shared Player `+17h` DEX reaction table、`1d6` action delay 與每次
  全表逐節點 `1d100`；delay 優先、roll 解 tie、完全 tie 後掃者勝。engine
  `combat/initiative` 保存這套作品中立 primitive，CoAB Battle 保存建構順序，
  MON parser 也修正 combat team 是 `+198h` 而非 `+197h`。完整 effect
  duration／save、`area.field_596` surprise writer、DOS 等價性及原版底層
  PRNG 仍未完成。第 420 輪已推翻「一般 DELAY 是 20→19」：頂層 D 先進
  DONE 子選單，第二層 D exact 寫 delay 1 並由動態 scheduler 同輪重新入列；
  20→19 是尚未接 UI 的 Quick handoff。READY spec 419／420 是權威。
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

## 10. Compact 後恢復工作清單

1. 讀本檔、`CLAUDE.md`、`docs/project-status.md` 與 `CONTEXT.md` 尾端。
2. 檢查兩個 repo 的 status、diff、HEAD、remote，完整保留 dirty changes。
3. 查看目前 READY spec 與最後一次 player-path test。
4. 不重做已 push 的 milestone，不沿用已 supersede 的推論。
5. 若本節「未提交 milestone」仍存在，先延續它；完成後再把成果移入
   `CONTEXT.md` 並刷新本節。
6. 更新計畫，繼續縮短完整可通關與完整中文化的真實差距。
