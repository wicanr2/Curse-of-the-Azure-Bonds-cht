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

## 5. 視覺、版面與中文字體

- 固定 logical canvas 為 640×480。
- 原始低解析圖採 nearest-neighbour 整數放大，不做模糊縮放或 AI 補點。
- DOS 320×200 是 frame、geometry、素材與行為 oracle。
- PC-9801 640×400 是 CJK 字級、行距與資訊密度的重要 oracle。
- 一般正文、roster、status、commands 使用倚天粗體 16×15；24px 只用於
  標題或明確強調。
- Adventure chrome 必須使用本機 DOS runtime 抽出的 cracked stone raster，
  不回退為 generic 灰框。
- 左上第一人稱／人物／事件圖依畫面 contract 使用 cover＋clip 填滿內格；
  必須完整可見的素材才用 contain。
- HEAD／BODY 是分層素材；頭不能塞進胸口。戰鬥 CPIC sprite 依原生 24×24
  tile、footprint 與 anchor，不走一般圖片置中。
- 原版忠實 theme 永遠保留；美化只能是額外可切換 theme。

目前最新視覺規格：`docs/spec/348-original-dos-frame-pc98-type-density.md`。
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
- 原版 ECL 固定英文、原始 bytes、位址與選單 token 只可出現在明確的來源
  oracle／parser 測試，並應同時驗證結構或來源位置；不能拿它們代替產品層
  本地化驗收。
- 測試應驗證「ID 解析、locale fallback、事件綁定與畫面取得同一份資料」，
  使 JSON 改譯文或裝備顯示名時不必同步修改另一份硬編碼測試字串。
- 常見錯誤：在測試直接寫「龍盔」「火球術」「長劍」等目前畫面文字，再用
  `Contains` 判斷。即使測試會通過，仍是在複製 JSON 的真相來源；應改查
  `message_id`／`item_id`／`spell_id`，並另測該 ID 經目前 locale 解析出的
  畫面內容。若 JSON 改名後測試必須手動改同一字串，表示測試分層有誤。
- 既有測試仍有歷史技術債：部分 `internal/game/*_test.go` 直接以繁中字串
  `Contains` 驗證。碰到相關功能時要逐步改成穩定 ID，不得照抄其模式新增
  債務；一次遷移一個真實玩家 vertical slice，並保留原始 ECL oracle 測試。

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
- CoAB 本輪基底：`ff944bf`；第 383 輪 ECL 首輪記憶體生命週期與
  Myth Drannor 解鎖 milestone 會由本文件所在 commit 完成。
- Engine dependency：`031d4ea`（含繁體中文音訊架構知識庫及中立
  `audio/cyclepcm`、`audio/s98`、`audio/ym2203`、
  `audio/pc98soundbios`、
  `combat_visuals`、
  `music_tracks`／`music_bindings`／`music_cues` 與跨 locale
  `title_id` schema）。
- 本文件所在 commit 會晚於上述 CoAB 基底；compact 後永遠先以兩個 repo 的
  實際 HEAD／remote 為準，不要把文件內 hash 當成可自我引用的 latest hash。
- GUI 邊框、左上 cover、16×15 倚天、PC-98 typography study 已完成並 push。
- 專案仍是多 vertical slices prototype，尚未完整可通關。
- 主要缺口見 `docs/project-status.md`：完整 ECL/external routines、開場到結局
  玩家路徑、戰鬥規則/AI/法術/戰後、全地圖、全翻譯、音樂音效、完整 save、
  三平台發行與長時間回歸。
- 第 383 輪已證明 Standing Stone 以 `4C59=1／4C5A=1／4C5B!=0`
  計數三個已解除主人，接著揭露 Tyranthraxus 並要求前往 Myth Drannor。
  `BlockSession.RunFrom` 首輪遺失預載 memory 的缺陷已修正並有合成與真實
  ECL regression。下一步接通 `JOURNEY ON → MYTH DRANNOR → ECL6/GEO6
  block 0x40` 的正常玩家路徑；不得宣稱最終章已完成。

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
