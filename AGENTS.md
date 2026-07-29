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

- affected repo 的 `go test ./...`；Ebiten package 在 Docker/Xvfb 或先
  `go test -c` 驗證。
- `git diff --check`。
- deterministic screenshot 或 runtime trace。
- 真正玩家路徑驗證，而不只 direct-entry debug flag。
- 原版／remake 對照，明確標示 exact／reconstructed／未完成。
- 戰鬥不能只驗證靜態 layout 與數值；原版 DOS runtime／遊戲影片還要逐項
  對照近戰、弓箭／投射物、法術施放與命中、死亡動畫、音效及回合節奏。

不得用窄測試支撐「完整可通關」「完整中文化」「完整戰鬥」等廣泛聲明。

## 9. 目前權威狀態

### 本 milestone 基底

- 工作已於 2026-07-29 恢復，不再遵守舊的「暫停新增功能」文字。
- CoAB 基底：`d9718a1`（兜帽女子、手札 30／7、弗佐爾死亡、第四枷印解除）。
- Engine dependency：`051cd71`（ECL `script_block` 與 GEO `geometry_block`
  可獨立宣告）。
- 本文件所在 commit 會晚於上述 CoAB 基底；compact 後永遠先以兩個 repo 的
  實際 HEAD／remote 為準，不要把文件內 hash 當成可自我引用的 latest hash。
- GUI 邊框、左上 cover、16×15 倚天、PC-98 typography study 已完成並 push。
- 專案仍是多 vertical slices prototype，尚未完整可通關。
- 主要缺口見 `docs/project-status.md`：完整 ECL/external routines、開場到結局
  玩家路徑、戰鬥規則/AI/法術/戰後、全地圖、全翻譯、音樂音效、完整 save、
  三平台發行與長時間回歸。

### 目前未提交 milestone（不可遺忘）

- 第 352 輪雙戰已 push。dirty milestone 是第 353 輪洞窟離場：
  GEO4/`0x25 (6,3)` terrain `0x93` 必須走 boundary lifecycle，依序顯示
  Olive／Dimswart、騎士與紫衣女子，寫 `4CE2/7F12`、銷毀 item type
  `0x60/0x61`，再 `NEWECL 0x51` 回暗影谷。
- spec 354 已記錄下一個戰鬥視覺里程碑：建立 renderer-neutral action
  timeline，先完成 melee、bow、Magic Missile、death；目前只有死亡 overlay
  有時間相位，弓箭／法術 projectile 與逐 action 敵方回合尚未完成。
- 完成第 353 輪完整驗證後集中一次 commit＋push；不要為文件或小修拆 commit。

## 10. Compact 後恢復工作清單

1. 讀本檔、`CLAUDE.md`、`docs/project-status.md` 與 `CONTEXT.md` 尾端。
2. 檢查兩個 repo 的 status、diff、HEAD、remote，完整保留 dirty changes。
3. 查看目前 READY spec 與最後一次 player-path test。
4. 不重做已 push 的 milestone，不沿用已 supersede 的推論。
5. 若本節「未提交 milestone」仍存在，先延續它；完成後再把成果移入
   `CONTEXT.md` 並刷新本節。
6. 更新計畫，繼續縮短完整可通關與完整中文化的真實差距。
