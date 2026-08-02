# 專案成果盤點

更新日期：2026-08-02
本 milestone 的 CoAB 基底：`d825164`（第 478 輪）
依賴的 Golden Box engine checkpoint：`2ace47d`

實際最新 CoAB 版本以本文件所在 commit／GitHub `main` 為準，避免在同一個
commit 內保存不可能自我引用的 hash。

本文件記錄目前 GitHub 可驗證的成果與尚未完成邊界。專案已是可執行、
可測試、可展示多個垂直切片的 remake prototype，但仍不是完整可通關版本。

## 已完成並有測試／畫面證據

### 共用 Golden Box engine

- 獨立 repo `golden-box-remake-engine`，CoAB 劇情與翻譯不寫死在共用 engine。
- JSON game-pack schema、locale、text rules、events 與 map definitions。
- DAX、SSI indexed picture、EGA palette、GEO 16×16 geometry、
  WALLDEF／8X8D、SKY、AREA 與 world-map projection。
- ECL bounded VM、跨 block session、許多已確認 opcode、menu／picture／combat
  signals 與 deterministic trace。
- renderer-neutral combat、placement、footprint、movement terrain、camera
  及多項 AD&D 戰鬥規則。
- exact kind／area／block map selection；同一 GEO 可同時宣告 AREA 與
  first-person projection。

### CoAB remake 與繁中

- 640×480 Ebiten executable；原始點陣圖採 2× nearest-neighbour。
- 倚天原生 16×15 Big5 點陣字，使用與 Monkey Island 2 相同的水平 1px
  embolden；全形標點可另外載入 `SPCFONT.15`。
- 原版 BIGPIC 世界地圖、11×11 AREA 8X8D symbol map、第一人稱
  GEO／WALLDEF／8X8D／SKY viewport。
- DOSBox 原生 320×200 冒險 layout oracle；正式 top split 已修正為
  native 128／192，即 remake 256／384。
- DOS runtime cracked stone chrome 已抽成透明 native raster；一般 PIC／
  第一人稱場景使用原版灰色內框與 88×88 可見區，HEAD／BODY 人物使用
  獨立黃色裂紋舞台。640×480 以插入 40 個原生訊息列延伸，不拉伸上半部
  或命令帶。IDA／DOS bytes 已交叉證明 HEAD 後 BODY、BODY `row+5` 的
  原作繪製鏈；第一人稱與旅店正常玩家路徑畫面已重新擷取。
  PC-98 640×400 gallery 已建立字級密度參考索引。
- CHEAD／CBODY 頭身合成、CPIC／SPRIT／COMSPR 戰鬥小人及 animation metadata。
- HEAD／BODY 場景人物已依 DOS 320×200 實機量測，使用 game-pack native
  anchor／clip、2× nearest-neighbour 與原版黃色裂紋內框；不再誤走場景
  `cover` 裁切。旅店與剛德神殿正常玩家路徑已有 640×480 回歸畫面。
- DUNGCOM、WILDCOM、RANDCOM combat terrain 與桌椅 overlay。
- renderer-neutral 戰鬥動作時間軸已接入 attack pose、近戰 impact、弓箭與
  Magic Missile travel、phase-aligned 聲音、死亡 overlay，以及逐 action
  敵方回合；弓箭、generic spell projectile 與 magic impact 已改用原版
  COMSPR blocks，並有 DOS 公開影片逐格 oracle。
- Lightning Bolt `0x33` 已接入正常 memorized slot／32×16 tile cursor；
  共用 line-spell 規則支援正交／對角 2／3 step cost、地城牆反向、
  large footprint 去重、反彈重複命中、共用等級 d6 與逐目標 Spell save
  半傷。時間軸可交錯 primary travel、ordered segment、damage impact、
  commit／death；CoAB JSON 宣告 COMSPR `05/85`、`06/86`、`0A/8A`。
  DOS 影片 `07:40:22.50–25.60` 與三張 remake checkpoint 已保存。
- Stinking Cloud `0x22` 已接入正常 memorized slot／tile cursor。規則保存
  target-anchored 2×2 persistent-area instances、passable-cell filter、
  large footprint 去重、Poison save、成功 cough、失敗 `d4+1` helpless、
  caster-level 到期與重疊雲獨立清除。DOS 影片
  `00:42:25.20–27.00` 的建立／持續關鍵幀及兩張 remake checkpoint 已保存。
- 角色建立的 22 個單職／18 個多職模板已移入 game-pack JSON；繁中姓名輸入、
  能力值、種族／職業顯示、基礎隊伍、裝備與多項法術已有可玩切片。正常 title
  至 ECL1 新遊戲與 save round-trip 已回歸；完整原版建角 UX／規則仍未完成。
- 舊 Go switch 的 84 個既有 ECL menu token 已由 game-pack 105 條
  `option_rules` 完整覆蓋；新遊戲、Journey／商店、刀刃房與定身房真實 ECL
  路徑均由 stable ID 解析。另有 15 個既有手札觸發與 23 頁 en／zh-TW 內容
  已移入 `text_rules`／locale，真實 ECL 路徑與遊戲內重讀均由 stable ID
  解析。開場概要與 ECL1 新遊戲的兩個真實文字 boundary 也已由 game-pack
  驅動。Area 5 的 Hap 17 個、熔岩洞穴 10 個事件也已資料化，正常長路徑覆蓋
  其中 26 個 boundary。Area 5 離場、龍巫妖與 Essembra 四個城市 boundary
  也已資料化；其餘劇情文字 fallback 尚未完成資料分離。Hillsfar 的 trail
  伏擊、edge、places、dockside bar 與 Red Plumes
  挑釁，以及 Yulash 的 trail、城牆、檢查哨、指揮官與巨坑入口十二個事件，
  也已由正常玩家路徑驗證。摩安德之坑開場至愛麗雅絲／龍餌入隊、樓層切換
  與散塔林屍體的十四個 boundary 亦已資料化。摩貢祭壇至離坑的十八個
  boundary、兩場戰鬥、護手、祭壇財寶、手札 20、
  離場阻擊與同伴離隊亦已完成同一正常玩家鏈驗證。
  散提爾堡城門至眼魔洞窟的既有資料規則已改用 stable-ID 驗收；德克薩姆與
  散提爾堡部隊兩戰現在實際跑完回合，不再以直接清除敵方 HP 冒充玩家戰鬥。
  提爾隘口、阿沙本福德、暗影隘口、立石群與 Essembra／Hap 路線十一個文字
  boundary 亦已資料化並由正常長路徑驗證。提爾佛頓公會至火刀據點入口另有
  十四個 boundary 完成 stable-ID 資料化與真實 image continuation 回歸；
  block 4 另有十一個房間 boundary 完成資料化，並保留刀刃傷害、手札 9／26、
  辦公室財寶與五個完成旗標回歸。皇家馬車、監牢救援與下水道騎士另十五個
  boundary 已由同一新遊戲 integration session 驗證。
  高階祭司、火刀戰後四段夢境及返城禁止入城另七個 boundary 已資料化；玩家
  回歸會實際走完阻擋與返回選單。
  艾森布拉 Tale 44／60 已由既有真實酒館鏈改成 game-pack stable-ID 驗收；State
  line localizer 的 27 筆中文 fallback 副本已移除。第 476 輪再將 CAMP／REST／
  FIX／VIEW／MAGIC 與 ALTER 的顯示文字改由正式 locale stable IDs 驅動，移除
  128 次 Go 漢字並讓行為測試載入正式 catalog。第 477 輪再將 PROGRAM 0／3／8
  的主選單、全滅、勝利與保存／不保存結束畫面改由正式 locale 驅動；正常 37 人
  終戰在兩次 save／restore 後仍取得同一繁中 catalog。第 478 輪移除新遊戲
  預載的八頁開發用假手札；Engine 現回傳 `journal_message_ids`，CoAB 以穩定 ID
  解鎖／去重，save v10 只保存 ID 並依讀檔時語系重解全文。空手札不再顯示
  `1 / 0`，繁中保存後以英文讀取的跨語系回歸也已通過。第 479 輪再把 23 個
  MON 戰鬥者名稱移入 Engine `combatant_name_rules` 與 CoAB game-pack；save v11
  保存 `SourceName` 並依讀檔語系重解 active-combat 顯示名。真實鷹馬、黑龍、
  龍巫妖、散塔林與摩安德戰鬥路徑均改由 stable source 驗收。最新 baseline
  為 345，`localization_debt` 已降為 0。第 480 輪再把戰鬥 HUD 的 HP／AC、
  施法／移動 prompts、十二個快捷提示與 target／quick status 移入正式 locale
  contract；Ebiten `drawCombat` 只保留版面與顏色，baseline 降為 320、frontend
  debt 108。第 481 輪再將 Cloudkill、Stinking Cloud 與 Lightning Bolt 的逐
  impact／commit／death 訊息移到 State visual-result locale contract；renderer
  不再推論豁免／死亡文字，baseline 313、frontend debt 101。第 482 輪再將
  F5／F9 存讀檔結果、音訊續跑錯誤、ALTER 改名與 ECL `INPUT STRING` 的 value／
  help 文字移入 State typed locale contract；地城與一般事件共用同一輸入文字
  契約，baseline 300、frontend debt 88。第 483 輪再將冒險選單、事件／圖片
  繼續、暗影谷 AREA、世界地圖、角色欄、戰鬥檢視與倒地標記二十五筆文字改由
  typed `PlayerUILabel` 驅動；世界地圖 game-pack 地名也跟隨 State catalog
  language，不再在 renderer 固定 `zh-TW`。baseline 275、frontend debt 63。
  第 484 輪再把鎖門提示、地城 lifecycle 錯誤、Pick／Knock／Bash 結果與正常
  地城操作列移入 State typed locale contract；原門旗標、法術消耗與雙側解鎖
  流程不變。baseline 262、frontend debt 50。第 485 輪再把素材載入、AREA、
  GEO geometry、地城研究 preview 與世界地圖日期二十四筆文字移入 typed
  diagnostic contract；`LOAD PIECES` selectors 保留 `uint16`，preview 不再拼接
  門選項或切割翻譯後時間全文。baseline 238、frontend debt 26。第 486 輪再把
  demo fighter、世界地圖 preview 與法師塔自動路徑最後二十六筆前端漢字移除；
  選項改依 exact ECL source token，劇情抵達改依 stable game-pack message ID，
  demo 名稱改讀 locale。baseline 212、frontend debt 0。第 487 輪再將建角完成、
  手札框、荒野選單、世界地點、NPC 名稱與城鎮／地城提示的中文 fallback 移回
  正式 catalog；動態所在地提示也改用格式 ID。baseline 173、frontend debt 0；
  runtime 173 仍待清理。第 488 輪再把財寶列表、角色收取、取消／略過及缺素材
  訊息移入正式 catalog；火刀辦公室正常 ECL 路徑改用正式 catalog 驗證，原
  財寶數量與 continuation 不變。baseline 164、frontend debt 0、runtime 164。
  第 489 輪再把 PARLAY 提示、五種策略與 generic 結果移入正式 catalog；法師塔
  與羅剎妖居所長路徑同時驗證原 tactic identity 和動態顯示。baseline 157、
  frontend debt 0、runtime 157。
- 商店、旅店、酒館、神殿、訓練、紮營及多段真實 ECL 主線／支線 vertical
  slices。
- ECL1 14×4 世界目的地圖已移入 JSON；Standing Stone 揭露後可經正常
  Journey On／Wilderness／AREA 抵達 Myth Drannor，選擇 Enter City 後
  進入 ECL6／GEO6 Burial Glen block `0x40`。這只證明入口 vertical slice，
  尚非完整 AREA travel 或 Myth Drannor 章節。
- Burial Glen 入口會採 ECL exact 出生點 `(2,15,E)`；玩家可依 GEO 牆面
  正常走兩步到 terrain `01h`，觸發 PICTURE 72 精靈幽魂。`GREET` 會把
  JSON 版 Journal 25 加入遊戲內手札，`FLEE／ATTACK` 原始分支也有 ECL
  oracle。玩家現可沿正常 GEO 通道抵達 terrain `82h` 紅網；真正的
  `INPUT STRING` 支援 Unicode、退格、ECL 指定長度與同 PC resume，
  `Krrkik` 不再假造成答案選單。四選單、HACK／RETREAT、四隻蜘蛛第一戰、
  PICTURE 72 揭露 rakshasa、第二戰及 `4CBFh` completion 均有 real-image
  ECL regression。正常 State 玩家路徑也已用真實 MON6CHA 完成四蜘蛛勝利、
  同 ECL continuation、羅剎妖第二戰、`4CBFh` 持久狀態及地城返回；原版
  CPIC 蜘蛛戰鬥 checkpoint 已保存。Journal 25 的「強大力量」是誘餌，
  ENTER script 沒有能力值寫入。戰敗路徑、敵人完整 AD&D 能力及後續區域
  仍未完成。
- remake JSON save v11 已保存地圖、時鐘、ECL session／持續 PRNG、戰鬥、
  音訊 continuation 與手札穩定 message IDs；原版 SAVGAM／CHRDAT／FX／SWG
  則有已驗證欄位 import、raw
  preservation 與部分 writeback。任意 UI／戰鬥 frame 的完整續存仍未完成。
- 角色年齡 offset `0x76..0x77`、種族／職業 mapping 與 DOS 實機角色頁證據。
- 中文手冊、攻略、Gold Box 技術知識庫、READY 規格與 README 實機截圖。
- 物品 type／name-number 維持 DOS typed IDs，繁中 base name、修飾詞、加值、
  詛咒與數量格式已由正式 locale stable IDs 驅動；商店、裝備、戰利品、CLI
  與測試不再依賴 Go 中文物品 catalog。二十個既有已命名角色／怪物效果也已
  依 raw affect kind 改由正式 locale 驅動；完整效果規則與演出仍須逐項驗證。
- 第 450 輪將法師塔正常主線從庭院、德拉坎德羅斯現身、龍群幻象一路到
  枷印消退的九段文字與手札 15 兩頁移入 CoAB game-pack。原文 fragment、
  英文、繁中與 `journal_message_ids` 由 11 個 stable ID 驅動；State 與舊
  locale 的複本已移除。原 ECL 事件發生後兩頁手札直接寫入遊戲內並可重讀，
  不要求玩家查 PDF。正常 GEO5／ECL 玩家路徑與 en／zh-TW 252-key parity
  均有回歸；後續四選項分支與其他章節仍須繼續遷移。
- 第 451 輪接續移除法師塔後半段十個 State story cases 與三個專用 option
  cases；攻擊法師、交涉、14 黑龍、龍心酸液及屋頂雙層出口文字都由 CoAB
  JSON stable IDs 驅動，en／zh-TW 各 265 keys。規則、怪物、旗標與傷害仍由
  raw ECL continuation 決定，未移入 Go 或共用 engine。法師塔入口至離開的
  作品文字已完成資料分離；全遊戲其他章節仍須逐段稽核。
- 第 451 輪另以唯讀 Docker 盤點 `/home/anr2/cht/daemon_winter`：其
  `DEMON.INT` 已由該專案證明是具有 3,807 筆 relocation 的原生 MZ 8086
  executable，而非 ECL 類 VM。可沿用的是零硬編碼中文字串 gate、coverage、
  theme／storage／grid／RNG 與 sampled verification 方法；不可直接共用
  Gold Box VM、格式、戰鬥或 save schema。比較結果與 Wasteland 後續入口已寫入
  [`跨專案知識庫`](knowledge/ssi-rpg-cross-project-lessons.md)。
- 第 452 輪借用冬之魔 `uicheck` 經驗，建立 CoAB AST-based Go 漢字字串
  exact baseline gate。初始為 1,260 signatures／1,315 occurrences：本地化債
  409、Ebiten 前端 164、runtime／工具 742。正式 `./internal/...` 測試會阻擋
  新增、改字、搬動、增加副本及未同步 reduced baseline 的刪除；此數量是待清
  技術債，不是允許額度。READY spec 452 與 `docs/audit/README.md` 是權威。
- 第 453 輪完成第一個由該 gate 驗收的設施切片：訓練場提示、確認、失敗、
  升級結果、職業與 49 筆可學法術名稱均改由正式 locale stable ID 取得；
  Go 規則表不再保存中文 `Name`。測試即時載入正式 JSON 並依 ID 解析期望值。
  exact baseline 從 1,315 降為 1,251，移除 64 次 runtime UI 漢字 literal；
  其餘 1,251 次仍是待清債務，不得宣稱全專案已完成資料分離。
- 第 454 輪接續完成剛德神殿資料分離：十種治療只在 Go 保存 stable key 與
  價格，選單、確認、查看、金幣與結果文字全由正式 locale 取得。原始
  terrain `92h`→PICTURE 6→service boundary→治療→返回 `(0,7)` 的玩家路徑
  仍通過。exact baseline 再由 1,251 降至 1,223，移除 28 次 runtime UI
  漢字 literal；神殿完整原版功能與其餘債務仍未完成。
- 第 455 輪完成 BAR 傳聞服務及提爾佛頓原始酒館事件的資料分離；六則預設
  傳聞、動作／飲品、紫帶女子、騷動、匕首與手札 17 提示均由正式 locale
  stable IDs 取得。提爾佛頓、阿沙本福德 Tale 28、艾森布拉 Tale 60 真實
  ECL 路徑均改讀正式 JSON；手札 17 的 prefix／全文也不再硬編碼於 Go。
  baseline 從 1,223 降至 1,169，移除 54 次；
  其他酒館事件與飲酒規則仍未宣稱完整。
- 第 456 輪完成商店 UI 資料分離：購買、販售、鑑定、估價、查看、金幣操作、
  選擇與結果格式共 63 個 stable IDs 均由正式 locale 取得；Weaponers
  terrain `84h` 真實 ECL 路徑仍完成購買與返回 `(2,12)`。baseline 從
  1,169 降至 1,100，移除 69 次。物品名稱仍由 `monster.ChineseName` 的 Go
  catalog 產生，是下一個明確資料債，不能把本輪稱為完整 item JSON 化。
- PC-98 `VFD1.00` 唯讀稽核工具、兩張磁碟雜湊與 absent CHRN 已建立；
  NP2kai Docker 實機可開機至 MEGDOS／`loader.com`。`MSCDRV.EXE` 已確認為
  `INT D2h`／YM2203 (`0x188/0x18A`) 常駐驅動；MAME FDI 身分雜湊與 loader
  三段 EXEC 順序也已交叉驗證。absent sectors 疑似同時參與早期完整性／
  防拷，driver 中間 1 KiB 尚未取回，未宣稱完整 driver 已復原；十二首
  sequence 已證明不跨該缺口，現可由殘存 exact driver 合成播放。
- PC-98 `GAME.OVR` 36 段 TPOV code／relocation 已可重現解碼；第 412 輪
  進一步驗證 resident control 的 `20h` header、五 byte `CD 3F` entry stub
  與遞增 `u16` fixup offsets，typed resolver 可由 far-pointer stub exact
  解析 overlay-local handler；`GAME.EXE`
  的 Borland `0x52FB`／9-byte legacy symbol table 已解析 1,725 symbols。
  53 筆 compiler modules 也已解析，可辨識 `INTERPET`、`MENUS`、
  `COMBAT` 等 unit。`MSCPLAY`／`MSCSTOP`／`BGMPLAY` 的地址、IVT `7Eh`
  wrapper 與七組 selector 已由 bytes 證實；`INTERPET` 的 writer 又證明
  selector 輸入是全域 `CURRENTECL`。exact ECL block mapping 已放進 CoAB
  JSON，獨立 engine 提供 `music_tracks`／`music_bindings`／`music_cues`
  contract，State 會發出一次性音樂 intent。Disk B 的無 BPB FAT12 配置與 PC-98 DAX
  codec 也已由 IDA／24-block corpus 驗證；`WLDTWN` 四個 ECL writer 進一步
  證明 selector 5 是區域／戶外導航、selector 6 是城鎮設施選單。CoAB JSON
  已把 `PICTURE 0x50/0x79` 資料化為服務／導航 cue；阿沙本福德 block
  `0x50` 與希爾斯法 block `0x51` 的正常 ECL 玩家路徑均驗證
  `5→6→5`，且同曲不重播。第 364 輪又證明 MSCDRV 直接安裝
  `IVT 7Eh → CS:0080`，public ABI 是 `AH=0/AL=track` play、`AH=1` stop，
  再透過內部 clients 接到 D2h。`cmd/pc98-music-audit` 會以 executable
  雜湊與 raw bytes 驗證 bridge、17 組 Sound BIOS 命令及 direct YM2203
  helper；它們全在 driver 缺口前。NEC 官方 BIOS 手冊已證明 `CEE0` 是
  Sound BIOS 固定介面表，而非未知 provider。曲名、播放器、正常 800ms
  換曲與無限 loop 已建立；完整 SFX trace 尚未完成。第 366 輪另證明十二首、84 個
  channel sequence 全在
  file `0x1B61..0x3C58`，沒有跨越 `0x4000..0x4400` 缺口；Hoot metadata
  已補齊十二首中英文曲名，runtime importer 會驗證 driver 雜湊與每段範圍。
  第 367 輪另完成 `sub_10410` 的 family-aware bytecode framing、
  `A0–A4` 控制流與原版 16-entry stack；84 組 stream 各驗證 256 個
  timed events。Timing channel 依原驅動採有界 read-through，不假造
  descriptor end gate。
  第 368 輪已把正常配樂路徑推進為 deterministic Sound BIOS／YM2203
  events：FM 12-word F-number、PSG 71-word period、duration、envelope、
  modulation 與 tempo 均由 verified driver data 驅動；十二首各驗證
  4,096 ticks，共 68,291 events。
  第 369 輪另由 IDA 證明 FM 音色表位於
  `seg003:0542`／file `0x45A2`，typed parser 已驗證二十組 NEC 50-WORD
  內嵌音色並逐曲列出所有呼叫索引。第 370 輪以修正游標後的十二首 Hoot
  S98 v3 trace 證明 NEC rate／level 反相、operator 寫入順序及 signed
  DETUNE shift；共用 engine `1a6a252` 已提供 S98／YM2203 parser。
  `20,21,23,24,25,26,27,58` 只在 descriptor 初始化短暫載入，第一個
  stream `85h` 會在首次 key-on 前改回內嵌 `0..19`。十二首的可聽音色
  現均可由二十組 bank 覆蓋。第 371 輪又用指定 IDA Pro 9.4 對
  `SOUND.ROM` 的 16-bit consumer 及十二首共 72 組 S98 啟動序列，證明
  `TL=127-OUTPUT_LEVEL`、algorithm carrier `4→2→3→1` 與
  `OPERATOR_MASK` key-on；事件 runtime 不再固定強制四個 operator。
  共用 engine `77683a3` 提供 `audio/ym2203` 拓樸、帶 mask 的 key-on
  snapshot 與 `audio/pc98soundbios` LFO 核心。
  第 372 輪又恢復 `SOUND.ROM` timer ISR 的 register-held 間接分支，
  完成六種軟體 LFO waveform、pitch／TL 投影及 S98 動態 extractor；
  第 373 輪以 45.01 秒 Hoot trace 確認其 Timer B ISR 可觀測性缺口，再以
  exact ROM 8086 harness 動態證明 sync 8 第 30 tick 首次輸出、80 tick
  共 51 組。engine Timer B scheduler 與 CoAB parameter adapter 已完成；
  第 374 輪又證明 MSCDRV 自己接管 YM2203 IRQ，只在 Timer B 呼叫
  `TrackPlayback` 且不鏈回 Sound BIOS ISR；所以 faithful BGM 不啟用上述
  LFO。第 375 輪由 S98 證明 3,993,600 Hz／prescale 6，engine 已完成
  Timer B 完整 count period 與無累積小數誤差的 PCM sample accumulator。
  第 376 輪已固定 BSD 授權 `ymfm`，完成 Sound BIOS intent →
  register writes → YM2203 → 44.1 kHz PCM → Ebiten player；可由
  `-pc98-music-driver` 播放，亦可用 `pc98-render-track` 產生本機 WAV。
  selector 5 兩次十秒輸出 hash 完全一致並通過非靜音／無 clipping 稽核。
  第 377 輪再證明正常 `MSCPLAY` 是 stop → 800ms delay → play，而不是
  driver fade；播放器已保留 35,280-frame 靜音且不推進音序列。正常
  loop count 0 會轉成 `0xFF` 無限循環。driver 內部 40-tick fade／FM SFX
  沒有正常 GAME caller。第 378 輪另由 IDA 8086 與 raw TPOV auditor
  找到正常 `GAME.OVR → SOUNDFX` 的 42 個直接 caller、selector 分布、
  port `37h` pulse routine 與 `GAME.EXE` file `E66Ch` 的 16×20 WORD
  音序表；作品端 importer 會驗證 exact executable 後輸出 typed
  pulse／delay steps，不提交商業 bytes。第 379 輪再由 Borland exact
  symbols 證明 `CASTFX` 到 `CRASHFX` 的 selector，State 改發平台中立
  sound intent；DOS WAV 與 PC-98 mapping 不再混用。共用 engine
  `audio/cyclepcm` 會積分 pulse duty cycle，作品 adapter 可用 V30 8 MHz
  profile 重建 deterministic one-shot；正式 Ebiten 開場已在
  Docker／Xvfb／ALSA null device 載入。仍缺原機 wall-clock／wait 校準、
  `27h` reload phase、實體裝置 loopback 與 analog mixer gain；remake BGM／
  one-shot save/resume 已由 spec 447／448 完成。
- 真實連續主線已由開場延伸到散塔林堡：內城奧莉芙事件、手札 50／51、
  `ECL4/GEO4 0x20→0x21` 密道、神殿 `(10,6,N)` 操作權、南方牢房導航、
  迪姆斯沃特同行、手札 12 六頁、兜帽女子離場、手札 30／7、弗佐爾死亡、
  第四枷印解除及 `0x21→0x22` 眼魔洞穴入口已有 regression。洞穴
  `(15,1,N)` 的德克薩姆／梅杜莎／眼魔／牛頭人決戰、四件普通戰利品、
  取回洛山達護符，以及 19 名散提爾堡部隊第二戰也已有真實檔案 regression。
  洞窟 boundary terrain `0x93` 的奧莉芙／迪姆斯沃特告別與
  `NEWECL 0x51` 暗影谷返回亦已接通。
- 第 388 輪已由正常 GEO 路徑接通 Burial Glen terrain `04h` 的隨機
  墳墓掠奪者：2 巨型蜘蛛、3 相位蜘蛛、1 thri-kreen 戰鬥後可重新安葬
  或搜刮。`4CBAh` 已證明是以 `80h` 表示中立的偏移好感；兩分支分別
  raw `+1／-1`。只有一件珠寶、`ItemBlock=FFh` 的 TREASURE 現會進入
  寶物 service，不再誤成零怪物戰鬥；選項翻譯已移入 game-pack
  `option_rules`。四批上限長回歸、完整怪物規則與後續區域仍未完成。
- 第 389 輪新增作品中立 `-geo-path` auditor，證明墳墓 `(6,12)` 到
  黛米爾 `(13,14)` terrain `03h` 有九步正常可行路徑。正常玩家 regression
  已逐格抵達 PICTURE 72；real-image ECL 測試涵蓋 ACCEPT／REJECT／KILL／
  FLEE、祝福／寬恕、一次性 `4CC0h`、`4CBAh +5／-10` 與
  `4CBBh=02h／FEh`。繁中提問與四選項均由 stable ID／JSON 驅動，並有
  原版石框、原始人物圖與 16×15 倚天的 640×480 checkpoint。
  第 390 輪已補上間接 consumer：ECL6 各戰鬥入口 exact
  `SAVE [4CBBh]→[7F71h]`；PC-98 Borland `VARLISTTYPE=0800h` 證明
  `VARLIST+06E2h` 對應 `7F71h`，IDA `ATTEMPTTOHIT` 以 `CBW` 將
  `02h／FEh` 解成 signed `+2／-2` 後加入 attack roll，`DOPOSTCOMBAT`
  則在戰後清零。engine `combat_modifiers`、CoAB JSON、Battle side-scoped
  modifier 與正常玩家路徑命中邊界均已接通；角色基礎 AttackBonus 不會被
  永久改寫。離開 Myth Drannor 後清除此持久效果的 writer、所有特殊武器
  caller 與完整戰鬥 fidelity 仍待完成。
- 第 392 輪把正常玩家路徑從黛米爾繼續到 terrain `93h／94h`：
  原始 ECL 分別建立十隻／八隻 `MON6CHA 41h` PHASE SPIDER，勝利後寫
  `4CCD／4CCE=1` 並可正常重踏。兩段繁中由 game-pack stable ID 驅動；
  terrain `95h` 已由下一輪接續完成。
- 第 393 輪接通 terrain `95h` `(14,10)` 的六隻 PHASE SPIDER、勝利後
  `4CCF=1` 與骨堆 `LOOT／REPLACE IN CRYPTS／IGNORE` 三分支。raw ECL
  證明 LOOT 是 `4CBA-1` 與一顆 gem、`ItemBlock=FFh` 無裝備；
  REPLACE 是 `4CBA+1`，IGNORE 不改好感。正常玩家路徑、繁中選單、
  treasure service、地城 continuation 與重訪 EXIT 均已回歸。
- 第 394 輪把正常玩家路徑延伸至 terrain `8Eh／8Fh／90h` 的螳螂人防線。
  `8Eh` 十二人寫 `4CC8`、`8Fh` 六人寫 `4CC9`；營地 `90h` 首波十二人
  寫 `4CCA`，再依前兩旗標決定是否追加兩波六人。乾淨 raw session
  `12→6→6` 與已清外圍後只剩十二人的正常路徑均已驗證。最終財寶為
  9500 gold、4 gems、6 jewelry 與一件 random item；treasure menu 會保留
  當次 ECL 的資料包繁中文字。terrain `91h／92h` 已由下一輪接續完成。
- 第 395 輪接通 terrain `91h／92h`：八隻 GIANT SPIDER 的陵墓寫
  `4CCB`；蛛網巢穴依 `4CBA < 80h` 決定是否略過幽魂警告。高好感
  YES／NO、NO 不消耗事件、YES 戰前 `4CCC=1`、四隻巨蛛與敵方
  `7F70=2` 均由 raw ECL 與正常玩家路徑驗證。低好感直接開戰分支也有
  獨立回歸；蛛卵沒有原作選單或財寶，未自行添加。
- 第 396 輪把正常玩家路徑延伸至西側精靈王庭：terrain `08h` 門口幽魂
  YES 會傳送到 `(4,2,S)`；terrain `89h` 的 GO UPSTAIRS／TAKE ARMOR／
  ATTACK／RETREAT 會正確影響 `4CBA` 與 `4CC4`。terrain `8Ah` 以
  `80h` 為友善門檻，敵對時建立 `6+4+4` 共十四名敵人；terrain `8Bh`
  的友善王后給 12 gems、8 jewelry 與 ITEM6 block `41h` 六筆物品，
  敵對王后則再扣五點並依 YES／NO 給較少財寶或拒絕，最後倒塔傳送。
  三個完成旗標、繁中 stable IDs、ITEM6 真實解碼與 Standing Stone 起始
  玩家路徑均已回歸。使用者提供的角色資訊實機圖也已納入 spec 391，
  固定人物舞台／右側 roster／下方長文三區 layout oracle。
- 第 397 輪已證明王庭不是出口，並由 `(1,3)` 沿 19 步合法 GEO 路徑抵達
  terrain `05h` 的紅羽戰士。`WAIT` 解鎖資料化手札 33；`AGREE／
  REFUSE PAYMENT／DISAGREE`、幽魂警告與可中止分支均有 raw ECL 回歸。
  同行陷阱會用一條 `DAMAGE flags=2,1d6+6,saveFlags=35h` 解析成兩次
  隨機目標箭擊，再進入六隻 PHASE SPIDER 與一隻 RAKSHASA 的戰鬥；
  Standing Stone 起始玩家路徑已接回戰後精靈骸骨選單。手札內容由使用者
  提供 PDF 第 10 頁核對；提示、敘事及選項均由 JSON stable ID 驅動。
  遭遇 `COMBAT／FLEE` 外部語意、羅剎妖完整能力與戰利品、terrain
  `07h／0Ch`、區域出口及最終神殿仍未完成。
- 第 398 輪已由正常 GEO 玩家路徑接通 terrain `0Ch` 的一次性無名人物；
  `WAIT／PARLAY` 會寫 `4CC7=1` 並解鎖由使用者 PDF 印刷頁 23–24 核對的
  完整繁中手札 56。terrain `07h` 則證明為沒有完成旗標的可重複六相位
  蜘蛛＋一羅剎妖遭遇。紅羽戰士戰後可走到 `(15,6,E)`，經真正地城出口
  lifecycle 選 `PATH／WOODS／TURN BACK`；`PATH` 現依 ECL register
  正確進入 block `42h` `(0,12,E)`。完整 block `42h` 遺跡、神殿與結局
  仍未完成。
- 第 399 輪已把正常玩家路徑延伸到 block `42h` `(1,12)` terrain `01h`。
  提爾雪雅會解鎖完整繁中手札 5；玩家先迎戰五地獄犬＋五石像鬼，再可與
  他共同迎戰六地獄犬＋六石像鬼。ECL 的 `LOAD CHARACTER 8` 證明
  combatant array 固定保留八個玩家槽；第一隻 RAKSHASA 會成為
  QuickFight 臨時盟友，戰後不污染 roster。raw ECL、單人隊伍 adapter
  與 Standing Stone 起始玩家路徑均已回歸。
- 第 400 輪已沿合法 GEO 路線接通倉庫入口 terrain `02h` 與內部 `83h`。
  `4CD1` 讓結盟戰與直接擊敗六地獄犬＋六石像鬼匯合；普通踏入只顯示
  物資，主動 `SEARCH` 才取得 9,500 gold、8 gems、8 jewelry 與
  ITEM6 `82h` 兩件裝備，並在服務返回後寫 `4CD2=1` 防止重複。來源地圖
  boundary `7ED5` 不再污染目的 block 的下一步事件。raw ECL 與 Standing
  Stone 起始玩家路徑均已回歸。
- 第 401 輪已接通 terrain `04h／05h／06h`。救援逃亡男子會迎戰六隻
  HELL HOUND，顯示 `HEAD6／BODY6 40h` 並取得東北藏寶線索；放棄救援的
  追擊／離開及一次性屍骸也有 raw 分支。正常玩家會沿合法 wrap 路線到
  `(14,3)`，主動 `SEARCH` 取得一枚 electrum 與 ITEM6 `43h` 的
  Gauntlets +2、Girdle +1、Long Sword +5。通用財寶投影會保存不足一 GP
  的 copper 餘數。
- 第 402 輪已接通 terrain `07h／08h／09h`。無名者以 HEAD `43h`／BODY
  `46h` 警告神殿在北方；灌木誘餌可放棄，或救援後承受 exact `2d8`
  落石，再迎戰 5 HELL HOUND、5 MARGOYLE 與 1 RAKSHASA。ECL 同 block
  寫入 `(11,10,S)` 後 `CALL 2E10h` 的傳送現會同步 State，玩家可正常向南
  進入可重複的血跡灌木敘事。raw 分支、繁中 stable IDs、合法 GEO 路線與
  Standing Stone 起始玩家路徑均已回歸。
- 第 403 輪已接通 terrain `0Bh／0Ch／8Ah／8Dh`。羅剎妖居所只有
  HAUGHTY 交涉能解鎖手札 57，其餘態度會迎戰 5 HELL HOUND、
  5 MARGOYLE 與 6 RAKSHASA。石像鬼門廊會造成全隊 exact `3d10`
  坍塌傷害，部分寫入座標並推到 `(10,2,N)`，之後可突襲單一 RAKSHASA。
  賭局房勝利可取得 11,200 gold、15 gems、9 jewelry 與 ITEM6 `81h`
  一件隨機物品；下水道柵口則經兩次確認切換至 ECL／GEO `43h`
  `(15,15,N)`。raw 分支、手札 PDF、繁中 stable IDs、部分座標 transaction
  與 Standing Stone 起始玩家路徑均已回歸。
- 第 404 輪已接通 block `43h` terrain `8Ah／8Bh／8Ch`。廚房、班恩辦公室
  與豪華臥房均有繁中 stable ID；臥房 YES 透過無怪物財寶服務取得
  5,000 GP、5,000 PP、12 gems、15 jewelry，合計 30,000 gold。正常
  Standing Stone 玩家路徑沿合法 GEO 抵達 `(10,12)`。同輪證實 block
  `40h／42h` 會先寫與內部房間重用的全域 `4C05／4C06=1`，所以完整支線
  路徑上的辦公室與廚房原本就會靜默；remake 未擅自重設。
- 第 405 輪已接通 block `43h` terrain `87h／88h／89h` raw 分支：
  犬舍為 10 HELL HOUND，活動雕像為 10 MARGOYLE，私人禮拜堂為
  1 HIGH PRIEST 與 4 PRIEST OF BANE；五段繁中與三個 one-shot 旗標均
  已回歸。Standing Stone 起始玩家路徑完成禮拜堂五人戰並走到 `(7,10)`。
  西翼最短合法路線必經 terrain `82h–85h` 最終儀式，因此犬舍與活動雕像
  在第 405 輪尚未以正常玩家路徑完成。
- 第 407 輪已由 `(7,10)` 正常南行進 terrain `83h`，逐段完成提朗瑟克斯
  控制枷印、手札 48、三神器交付、無名者揭露與臨終密語。敵軍 exact 為
  2 HIGH PRIEST、6 HELL HOUND、6 MARGOYLE；勝利後 `4C00=1`，相鄰
  `84h／85h` 靜默。Standing Stone 起始長路徑隨後完成西翼活動雕像與
  犬舍兩戰，`4C02／4C01=1`，第 405 輪的 player-path 缺口已關閉。
  當輪尚未完成真正終戰與二樓區域。
- 第 408 輪已由犬舍沿合法 GEO 路徑走完一樓、terrain `97h` 樓梯與二樓，
  抵達 `(6,1)` terrain `9Ah`。終戰 exact 為 MARGOYLE `45h`×28、
  TYRANTHRAXUS `47h`×1、HIGH PRIEST `48h`×8；正式 scheduler 戰勝後，
  同一 ECL session 立即抵達 `PROGRAM 8` 勝利存檔選單。盟友造成的真實
  零敵人 minion COMBAT、跨 boundary monster setup、二樓房間與繁中 stable
  IDs 均有回歸。這完成 Standing Stone 累積狀態的迷斯卓諾章節路徑，仍不
  等於由新建隊伍自開場至結局的單一完整通關，也不代表終戰特殊能力與
  DOS 動態演出已完成。
- 第 449 輪抽樣重建上述終戰 checkpoint：block `43h` 正常初始化後經
  terrain `97h` 樓梯及十步 GEO 路線抵達 `9Ah`，runtime 再次斷言
  MARGOYLE `45h`×28、TYRANTHRAXUS `47h`×1、HIGH PRIEST `48h`×8。
  大型戰場正式鏡頭已由錯誤的全體 bounds 中心改回 RuleBook 的主動角色
  焦點；README 的 640×480 圖片使用明確隔離的 capture-only 首領觀察鏡頭，
  不移動敵軍、不改 ECL／AI。這是高風險終戰視覺抽樣，不擴大為完整通關或
  完整戰鬥 fidelity。
- 第 409 輪將 remake save 升至 v6，保存 ECL current block、resume PC、
  stack、mutable memory、字串／compare state、輸入 offset、pending monster
  descriptors 與持續 PRNG。獨立 engine `randomstream` 以 seed＋底層 draw
  count 恢復 `math/rand` continuation，並有 replay 上限。真實 ECL6
  Burial Glen terrain `04h` 證明存讀檔後 raw random values、怪物與文字
  和不中斷執行相同；這不證明 SSI 原版 RNG，也尚未保存完整 UI／戰鬥 frame。
- 第 410 輪以 PC-98 Borland symbols 與 IDA 關閉 `MON*SPC` 九 byte 載入
  邊界：`LOADMONSTER` 複製前五 bytes、清除並重建 `+5..+8` linked-list
  pointer，且不改 byte `+4`。因此 raw `+4=0` 不再錯誤停用怪物天生效果。
  正常 Standing Stone 長路徑現載入真實 `MON6SPC`，提朗瑟克斯六筆效果
  均進入終戰；`18h` 偵測隱形已在命中公式抵消隱形 AC bonus。
- 第 411 輪以 PC-98 overlay 12 raw bytes／IDA 證明魔法抗性 common
  routine 的 `base + (11-casterLevel)*5`、`1d100 <= threshold` 與
  `Protected(0)` 傷害清除；50%／15% wrappers 分別在 local
  `23F4h／2404h`。第 412 輪由 `008B:0214 → entry 100 → 2404h` 將
  `6Ah → 15%` 升級為 `exact`，並靜態關閉其餘四筆 handler：`4Fh`
  2d10 fire、`70h` 防火、`84h` Lightning Bolt、`87h` 防電。Magic Missile
  現先擲傷害、再擲抗性，
  成功時傷害歸零，施放格與 continuation 仍消耗／進行；繁中訊息
  來自 locale stable ID。其餘四筆效果的 runtime boundary、所有魔法幾何、
  AI 與演出仍未完成。
- 第 413 輪將 `70h` 防火與 `87h` 防電接入 Fireball／Lightning Bolt 的
  damage flags boundary；inactive effect 反例仍受傷。正常 memorized spell／
  tile cursor／visual timeline 顯示資料化繁中防護摘要，Standing Stone 起始
  長路徑取得的真實提朗瑟克斯也在兩種攻擊下保持 HP，原終戰仍可完成
  `PROGRAM 8`。saving throw 與 protection 的原版 RNG 順序、`84h`
  怪物 Lightning Bolt、完整 AI／演出仍未完成。
- 第 414 輪由 PC-98 overlay 13 的 post-hit caller 與 overlay 23
  `CHECKFX` table 證明：第一、第二物理攻擊槽命中且同一目標仍存活時，
  operational `4Fh` 會追加同目標 `2d10` Fire＋Magic；物理擊殺不觸發，
  沒有 saving throw。Battle 以結構化 `AttackEffectResult` 分開武器傷害、
  擲骰傷害、實際傷害與元素防護，防火時仍消耗兩顆 d10。正式繁中 stable
  IDs、slot／miss／kill／inactive／protected regression 與 Standing Stone
  長路徑真實 `MON6SPC` 邊界均已通過。4F 原版動畫／聲音、自由攻擊與轉移
  目標動態 trace、`6Ah` 對 4F 的時序及 `84h` 怪物閃電仍未完成。
- 第 415 輪由 PC-98 overlay 9／22 關閉 effect `84h` 的 type-14 action
  phase、`ROUND < 4`、spell `33h`、初始格 `16d6` 與後續 range 10 路徑
  另一份獨立 `16d6`。Battle 的中立 line profile 保留玩家法術既有共用骰，
  State 則只在第 1–3 回合先於一般攻擊排程怪物閃電，並使用正式 terrain
  callback、元素防護、資料化 timeline 與繁中 stable IDs。Standing Stone
  起始長路徑的真實 MON6 提朗瑟克斯已產生動態閃電事件且仍完成
  `PROGRAM 8`。第 416 輪已接續完成 target range／牆面候選；終戰牆面反射
  與逐幀 timing 仍未完成。
- 第 416 輪由 PC-98 `PICKTARGET 00B8:3D7F`、overlay 24 候選建立器、
  overlay 32 footprint／地形 consumer 與 spell `33h` record，接入 range 10、
  cardinal／diagonal 2／3 加權距離、牆面阻擋、雙方 footprint 與二十次
  不可見候選移除重抽。effect `84h` 不再先從所有存活者任選；無候選時仍
  依原 handler 消耗 action，不回退物理攻擊。core／State tests 覆蓋超距、
  牆後、大型 footprint、不可見與無目標 continuation。第 417 輪已接通
  PC-98 status visibility；原始 combatant-array 同距順序與 `(0,0)` fallback
  動畫仍未完成。
- 第 417 輪由 PC-98 `CHECKTARGET 00B8:11AF` 與 overlay 12 effect handlers
  證明：`19h` 只在觀察者缺少 operational `18h` 時隱藏並給 attack roll
  `-4`，`47h` 則無條件隱藏並給 `-4`。State 的怪物 ranged selector 現使用
  同一作品中立 `VisibleTo` 契約；物理 AC 也修正為 `18h` 只抵消 `19h`，
  不抵消 `47h`。core／State 正反例與 Standing Stone 起始真實 MON6
  提朗瑟克斯長路徑均為驗收門檻。blink、動物視覺與完整 effect 生命週期
  由第 418 輪接續；完整 effect 生命週期仍未完成。
- 第 418 輪由 typed TPOV resolver 與 overlay 12 handlers 關閉 `25h／45h`：
  blink 在 action delay 0 時 hidden 並把 attack roll 寫 `FFh`；`45h` 只對
  MonsterType `13h` observer 生效，`18h` 只取消 hidden、不取消 `-4`。
  `+11Ah` 另由 dragon-slayer `03h` consumer 與 WORG／FIGHTING DOG／MONKEY／
  OWL BEAR 真實 records 證明為 MonsterType。Battle／State 已接通 pending／
  completed delay lifecycle，Tilverton 犬舍與 Standing Stone 長路徑均通過。
- 第 419 輪由 PC-98 overlay 24 `DEXRABONUS`、overlay 13 initiative writer
  與 overlay 8 TeamList selector 關閉先攻排程：每人先寫
  `1d6 + DEX reaction adjustment`，下一人由最大 action delay 與逐節點
  `1d100` 選出，完全 tie 時後掃到者勝出；全零掃描仍會消耗每節點一骰。
  `Battle` 現保存建構順序並投影 party／MON shared Player `+17h` Dexterity，
  不再使用 d20、fighter ID tie-break 或 MON `+1A5h` 假 bonus。可重用核心在
  engine `combat/initiative`，完整 engine 與 CoAB 31 套件 gate 均通過。
  第 420 輪已證明頂層 D 進入第二層、一般 Delay 寫 1 並同輪重新入列，且
  動態 scheduler 不預抽未來 d100；`20→19` 其實是 Quick handoff。後者 UI／
  runtime trace、`area.field_596` surprise writers、DOS 等價性與原版底層
  PRNG 仍未完成。
- 第 421 輪由 PC-98 overlays `08／13／18／24` 關閉 QUICK、GUARD、BANDAGE
  與 SPEED 的核心語意。作品中立 `combat/action` 保存清除、守備與 `0..9`
  動畫速度；CoAB 接通目前角色 Quick／手動收回、進入鄰接格 Guard 反應攻擊、
  第一名 Dying 隊員包紮，以及 D 子選單速度調整。原始位置保持唯讀，IDA
  名稱只存在報告／規格；spec 421 保存 exact／strong inference 分界。
  `ALT+Q` 全隊 Quick 已由第 422 輪接通，`ALT+M` 與 selector 已由第 424 輪
  接續；每動作 target pointer、敵方選擇 Guard／Bandage 的 AI 與原版
  wall-clock timing 仍未完成。
- 第 422 輪由 overlay 08 exact 關閉 `ALT+Q`：目前 Action delay 先寫 20，
  TeamList 全員經同一 setter 切成 Quick，下一次 action entry 轉為 19 後交給
  AI。Ebiten 視覺播放期間現在仍接受 Space，當前動作後可收回
  `ControlMorale < 80h` 的玩家角色。overlay 08／09／10 也證明 `DS:A86Ch`
  是玩家 Quick AI 法術選擇 gate；當輪未用空 toggle 冒充功能，selector 與
  `ALT+M` 已由第 424 輪接通。Standing Stone→紅網正常玩家路徑另抓出
  Space 只清 transient Battle、未同步持久 party 的缺陷；現已修正，第二場
  羅剎妖戰不會錯誤沿用 stale Quick。
- 第 423 輪由 PC-98 `GAME.EXE` 原始 16-byte spell records 與 overlay 09
  Quick selector consumer 證明 Player spell bytes 使用全域 ID。牧師
  Protection From Good 是 `07h`，Magic Missile 是 `0Fh`；舊 spec 134／142
  的 class-local 重疊 ID 推論已 supersede。camp label、記憶法術與戰鬥
  targeting／availability／cast 分派現使用同一全域 identity，怪物與玩家
  Magic Missile 也不再有 `0Fh／07h` 分裂。ALT+M selector 已由第 424 輪
  接通；全部 spell 效果與未支援法術 handoff 仍未完成。
- 第 424 輪接通 PC-98 `ALT+M` gate 與 Quick Magic Missile。overlay 09
  exact 證明 `1d7` priority tiers、每層三次隨機 memorized slot；typed TPOV
  resolver 把成功 handoff 落到 overlay 13 `CASTCOMBATSPELL 27A1h`。engine
  `combat/quickspell` 保存中立 selector，CoAB JSON 保存 priority／CastOn／
  MinRange；全域 `0Fh` Magic Missile 已從 Standing Stone→紅網正常玩家路徑
  由 ALT+M＋ALT+Q 實際施放並消耗 slot。非零 MinRange 的 area safety helper、
  當輪 casting delay、Cure special 與其餘法術仍 fail-closed；casting delay／
  Bless 已由第 425 輪接續，仍不能宣稱完整 Quick AI。
- 第 425 輪由 overlay 13 `CASTCOMBATSPELL` 與 overlay 08 pending consumer
  關閉非即時施法排程：raw `CastingTime/3` 非零時保存 Action spell ID，delay
  改為 `max(1,delay-units)` 並同輪重新入列。engine action／quickspell 與
  CoAB 十一筆 JSON metadata 已接通；Quick Bless raw `10→3` 有 scheduler／
  State regression。手動 CAST 與 typed 目標交易由第 428 輪接續；其他 Quick
  法術、原作目標指標 layout 與 interruption fidelity 仍未完成。
- 第 426 輪由 overlay 09 `03D3h..04C9h`、typed
  `00B8:0075h → overlay 13 entry 17 → +1E30h` 與 IDA 證明 Quick Cure
  先選同隊鄰近受傷／倒地目標，再把 target 帶過 pending action。engine
  action 保存 opaque stable `TargetID`；CoAB `03h` 已由 focused regression
  與 Standing Stone→Red Plume 真實箭傷正常路徑接通。九格 equal-HP exact
  tie order、down-player status predicate、手動 CAST delay 與 interruption
  已由第 427 輪接續；手動 CAST delay 再由第 428 輪接通，interruption 仍未
  完成，不能宣稱完整 Quick AI。
- 第 427 輪以 Borland `COMPTARGCURE／CHARSTATUS／DXDIR／DYDIR`、typed
  TACMAP stubs 與 raw bytes 關閉 Quick Cure 選人順序：N→NE→E→SE→S→SW→
  W→NW→self，equal HP 保留先掃者，自身低於半血覆蓋先前候選；active HP
  8 以上才由合法 down-player 取代，raw `DEAD／STONED／GONE` 被排除。
  CoAB selector 與 Stoned gate 已修正，四組 exact boundary regression 通過。
  同格多 corpse ordering、完整 raw status round-trip 與非正傷害中斷仍未
  完成；手動 CAST delay 由第 428、正傷害中斷由第 429 輪接續。
- 第 428 輪把手動 CAST 接回第 425 輪已證明的共同 `CASTCOMBATSPELL`
  handoff。非零 `CastingTime/3` 現保存 stable combatant ID 或 32×16 格點，
  同輪重新入列後才結算；Bless 可觀察 pending transaction，Fireball 即使
  delay=1 立即重新入列仍保持玩家選定格。Quick Bless／Cure 共用 resolver
  回歸亦通過。正傷害中斷與 slot 消耗由第 429 輪接續；其他中斷與尚未實作
  法術仍未完成。
- 第 429 輪由 PC-98 overlay 23 `PUTDAMAGE`、CP932 原文與 overlay 24
  memorized-slot consumer 關閉正傷害中斷：最終傷害 `>0` 才清 pending spell
  並消耗第一個 matching slot，零傷害／防護不觸發，Action delay 保留。
  engine action、Battle 全傷害邊界、正式 roster 與繁中 stable ID 已接通。
  Cloudkill 直接死亡由第 430 輪接續，held 麻痺／睡眠的不同 Action-clear
  政策再由第 431 輪接續；沉默／石化仍未完成。
- 第 430 輪由 PC-98 overlay 22 毒雲術 writer、overlay 12 effect table 與
  effect `44h` handler、overlay 24 memorized-slot consumer 閉合獨立中斷鏈。
  HD 0–4 自動死亡及 HD 5–6 豁免失敗現在會先中斷 pending spell，再進死亡
  handoff；HD 7+／豁免成功不消耗 slot。Battle／State regression 使用 stable
  fighter／spell ID 與正式 locale；held 麻痺／睡眠的 Action consumer 由第
  431 輪接續，沉默／石化仍未解析。
- 第 431 輪關閉 held effects 的不同取消政策：overlay 12 table slots
  `1Fh／33h／34h／35h` 全指向同一 handler，再呼叫 overlay 24
  `CLEARACTION 2A5Bh`；它清 pending spell／delay／guard，但沒有呼叫
  memorized-slot consumer。正常 scheduler 四例會清 Action、跳過回合且保留
  roster slot。六章正式 MON*SPC 沒有這四種 innate effect，Sleep／Hold 動態
  writer 尚未實作，因此不宣稱完整法術玩家路徑。
- 第 432 輪閉合全域法術 `15h` 到 overlay 22 entry 41 的 dispatch，以及
  Sleep 專屬 `4d4` 容量與 ordered HD 成本篩選。engine `combat/sleep`
  已保存 exact primitive；`+74h` 欄位仍採中立 predicate 名稱。上游目標
  幾何／排序、save、magic resistance、duration、效果寫入參數與演出尚未
  全部閉合，因此尚未開放手動或 Quick Sleep。
- 第 433 輪以 combat init writer、typed TPOV resolver 與唯讀 IDA ledger
  訂正 Sleep 上游：戰鬥 `DOSPELLTARGETING` 是 overlay 13 `225Fh`，不是
  預設／非戰鬥的 overlay 22 `GETSPELLTARGETS`。`AOECOMBAT=09h` 會呼叫
  overlay 31 `SCAN`，並依其三 byte候選表順序建立 targets；共用 writer
  對 Sleep 不做 saving throw，持續時間為 `5×caster level`。這仍只是
  targeting dispatch／排序骨架與 writer 參數的 bounded READY milestone；
  `SCAN` 幾何欄位、large footprint／tie order runtime、magic resistance、
  完整 `PUTEFFECT`、解除條件、動畫／音效及正常施法路徑仍未完成。
- 第 434 輪 exact 解析 `013E:0089h` 為 overlay 23 `PUTEFFECT 2325h`，並
  閉合 Sleep 在容量篩選後的 `CHECKFX type 9 → 6Ah` 魔抗與
  `ADDEFFECT 35h` writer。bounded `CastSleepOrdered` 保留四次 d4、target
  order、魔抗 d100、容量不退還及 raw `+0..+4`；未知／重複 target
  失敗即關閉，隊伍／死亡仍留在上游。這不是 UI 完成：`SCAN`
  terrain／footprint／tie 實機、
  手動／Quick delayed cast、效果解除、存檔、twinkle／聲音與玩家路徑仍待續。
- 第 435 輪以 overlay 31 完整 `SCAN／LOSEXISTS／INARC／sort` 指令訂正
  三 byte候選表：第一欄是 object ID、第二欄是最小成功 LOS 加權距離低 byte、
  第三欄是方向 payload；排序器不讀第三欄。等距時只依 object ID 與奇偶例外
  執行原版巢狀交換。engine `combat/scanorder` 與 CoAB stable-ID adapter 已
  接通並 fail-closed；terrain property、wall／large-footprint 實機、手動／
  Quick Sleep、效果生命週期與演出仍未完成。
- 第 436 輪以 Borland 0x52FB symbols／types／members、全新非破壞性 IDA
  resident raw table 與 overlay 31 `LOSEXISTS` 關閉 `TDEFTYPE HT／LOS／SYM`、
  `TACTICALMAP XRAY／TD`、一基底 tile index、cardinal 2／diagonal 3 metric
  及 inclusive `2*range+1` gate。engine `combat/scan` 已建立 terrain-aware
  footprint producer；CoAB adapter 以 explicit object ID 映射 stable fighter
  ID，並有 `producer → order → CastSleepOrdered` bounded transaction。
  `Raw3`、`INARC` direction sector、COMPOBJ builder、PC-98 wall／corner 動態
  trace、正常手動／Quick Sleep 與演出仍明確未完成。

- 第 437 輪已由 overlay 31 `054Ah..08D5h` 關閉 `INARC` 八方向整數扇區、
  `FFh→8` 與第一命中方向；engine 以 14,062,500 組全座標語料驗證。Borland
  symbols 另證明 `LASTOBJECT 9740h`、`OBJECTLIST 9741h` 與 72×4-byte 容量，
  overlay 10 builder 證明 X／Y／index／footprint-active 欄位及 linked-list
  建表。linked-list 到 stable fighter ID、正常手動／Quick Sleep、PC-98
  wall／corner 動態仍未完成；READY spec 437 是權威。
- 第 438 輪由 Borland symbols/types 閉合 `CHARACTERLIST 9598h`、
  `IDLIST 9DD3h` 的 72×`CHARRECPTR` 身份表及 `CHARREC.NEXT` linked-list。
  `StartCombat` 現重建一基底 `LegacyObjectID`，`Battle` 可由 stable fighter
  identity 展開真實 footprint，再完整串接 `INARC／LOS／sort`。Ebiten 尚未
  提供原戰場 `TACTICALMAP TD/TDEF`，故正常 Sleep UI、wall／corner runtime
  仍未完成；READY spec 438 是權威。
- 第 439 輪證明 Sleep `SCAN` 以玩家選定格而非 caster footprint 為中心，
  range `1`、arc `FFh`；`BackgroundTiles[1:66]` 與 PC-98 65 筆 TDEF
  逐 byte 對應，Dungeon／Wilderness floor bytes 是原始一基底 TD。
  frontend→State→Battle 已接通手動 `Z` 選格、terrain-aware ordered targets、
  `4d4`／HD／魔抗／`35h` effect 與成功後 slot 消耗；無效 map fail-closed。
  32×16 fallback placement 仍是 reconstructed，Quick Sleep、wall/corner
  動態、解除／save、twinkle／音效未完成；READY spec 439 是權威。
- 第 440 輪以非破壞性 IDA 副本、Borland symbols 與 resident raw bytes
  證明 `PUTDAMAGE` 的正傷害路徑會經 `REMOVEFX` 移除表中的 `35h` Sleep。
  `Battle.applyPositiveDamage` 已統一解除動態睡眠；零傷害與 innate MON*SPC
  record 不受影響。duration consumer、combat-end／save、醒來文字、twinkle
  與聲音仍未完成；READY spec 440 是權威。
- 第 441 輪以 `CLOCK_` overlay 20、Borland `TIMEUNITS` 與 resident
  `MAXCOUNT` 閉合 `EFFECTREC+1` duration consumer。獨立 engine 新增
  `combat/effecttime`；Battle 每個新 round 扣一 tick，duration 零保留。
  正常 level 3 手動 Sleep 路徑已驗證 15 tick handoff／到期；active battle
  save、到期文字、twinkle 與音效仍未完成。READY spec 441 是權威。
- 第 442 輪證明 effect `35h` add／remove callback 都是 `CLEARACTION`，並修正
  受傷順序，避免睡眠喚醒再走通用 interruption 消耗 slot。成功 Sleep 現逐
  target 播放四格 `VisualTwinkle`、每人 1440ms、施法聲與 `SPELLHITFX`；
  抵抗／醒來不播放。24×6 geometry exact，palette pixels 尚為 reconstructed；
  active battle save 與 PC-98 runtime capture 未完成。READY spec 442 是權威。
- 第 443 輪完成 remake save v7 active-combat snapshot：Fighter／effect／Action、
  scheduler selection、turn cursor、persistent areas、battle modifiers、pending
  interruption、visual 起點與 battle PRNG 均可 round-trip。正常 Sleep 讀檔後
  自然到期／傷害喚醒與不中斷分支一致。第 443 輪當時的 mid-animation
  fail-closed 已由第 446 輪解除；原版 SAVGAM combat records 仍未反組譯。
  READY spec 443 是 active-combat 基底，spec 446 是 visual resume 權威。
- 第 444 輪已由 Standing Stone 正常世界旅行、GEO6 合法移動、精靈幽魂、
  紅網 `SPEAK／Krrkik／ENTER` 抵達真實四蜘蛛戰；第一戰 party-turn save v7
  後以全新 State restore，Battle／ECL session 逐欄相同，loaded state 完成
  蜘蛛→Picture 72→羅剎妖二戰、`4CBF=1` 與 dungeon return。高數值英雄只
  加速回歸，不證明 encounter balance；READY spec 444 是權威。
- 第 445 輪已在真實 outer ruins 提爾雪雅→攻擊貝爾哈第二戰插入 save v7；
  party-side 羅剎妖的 `QuickFight／TemporaryAlly`、Battle 與 ECL session 在
  全新 State restore 後逐欄相同，而永久 roster 始終只有英雄。loaded state
  戰勝、寫 `4CD1=1`、返回 dungeon 後，runtime party／roster 都無盟友污染。
  READY spec 445 是權威。
- 第 446 輪已將 visual elapsed 移入 State／save v7，戰鬥動畫可在 Sleep
  `TWINKLE` 中段及弓箭 death frame 存檔並由同一幀續跑；travel／impact／death
  cue markers 會保存，已送出的離散音效不重播。elapsed／marker 損壞資料
  fail-closed，frontend speed 0／4／9 都保留 saved base。播放器 PCM sample
  offset、BGM driver／synth 狀態及原版 SAVGAM combat layout 仍未完成；
  READY spec 446 是權威。
- 第 447 輪將 remake JSON save 升至 v8：stable track ID、七聲道 machine、
  YM2203 full state、resampler remainder、Timer B、transition silence、pending
  PCM 與 Ebiten audible/read-ahead backlog 均可 round-trip。engine `f06493f`
  提供作品中立 ymfm／resampler snapshot。合成七聲道 fixture 從第一個 audible
  sample續跑一致；本機缺 exact 完整 MSCDRV，故十二首真實曲目 runtime oracle
  與原版 SAVGAM audio仍未完成。READY spec 447
  是權威。
- 第 448 輪將 remake JSON save 升至 v9：DOS WAV 與 PC-98 software-speaker
  會保存仍在播放的 stable selector／event、enabled 狀態與 44,100 Hz audible
  sample frame。多音效可同時續跑；自然結束、停用或舊版未保存的音效不復活；
  backend／asset／seek 錯誤會先停止 pre-load 聲音並失敗即關閉。這是 remake
  player continuation，不代表原版 SAVGAM audio 已解；READY spec 448 是權威。
- 第 383 輪修正 ECL session 第一次明確 `RunFrom` 會遺失預載存檔／區域
  記憶體的生命週期缺陷。原始 ECL1 block `0x50` 回歸已證明
  `4C59=1／4C5A=1／4C5B=FF` 時工作計數為 3，灰袍人會揭露自己是
  提朗瑟克斯，並要求隊伍前往 Myth Drannor。第 384／385 輪已接通
  ECL6／GEO6 正常入口與第一個精靈幽魂事件；其後 Burial Glen 至終戰已由
  第 385–408 輪逐段接通，但全遊戲開場至結局的單一通關仍未完成。

## 尚未完成

- 全部 ECL opcode、外部 routine、副作用及由開場到結局的完整可通關流程。
- 所有城市、地城、門、屋頂、斜向視角與每張地圖的 DOS 像素級校準。
- 戰鬥畫面的完整 DOS oracle 校準、弓箭／法術 projectile 的逐距離 timing、
  所有方向 placement、Fireball 牆面阻擋／同距排序、Lightning Bolt
  牆角／多次反彈 runtime oracle、毒雲每回合重複判定與魔法保護例外、AI、
  Cloudkill DOS 動態時間碼、其餘法術、物品、
  特殊能力、逃跑／交涉與戰後流程。
- 全角色／怪物／物品／法術 AD&D 規則及完整多職業、alignment、升級規則。
- 全英文文本、59 則 Journal、Tavern Tales、
  Clue Book／攻略的完整繁中化。
- 原版音樂與 PC Speaker／Tandy 音效的完整還原；PC-98 12 首 YM2203
  曲目已可由本機 driver 合成播放，正常 stop→800ms→play 與無限 loop
  已證明；PC-98 正常短音效的 42 個 GAME.OVR caller、selector 分布、
  音序表與 port `37h` pulse 程式也已證明並可由本機 GAME.EXE 匯入，
  第 379 輪已補具名 caller／事件、cycle→PCM 與遊戲內 one-shot mixer；
  第 380 輪再由 exact routine harness 與 NEC 官方表更正 V30
  `LOOP taken=5／exit=13`、四種 gate overhead 與 deterministic WAV。
  第 381 輪以 NP2kai direct V30 core probe 重現 exact OUT sequence，
  同時證明其 `_loop` 沿用 80286 `8/4` clocks，不能當原機 wall-clock
  oracle。第 382 輪已依 ymfm 補上 Timer B free-running divide-by-16
  重載公式、phase `0..15` 驗證與無漂移 PCM accumulator API；但 CoAB
  `27h=20h→0Ah` 間的七聲道 interpreter 是資料相依路徑，仍缺 CPU／OPN
  共時 trace，現行 instant-ISR renderer 明確維持 phase 0。目前仍缺原機
  edge trace／機型 wait／prefetch 校準、
  driver 內部 FM SFX 的真實 producer、十二曲 exact-driver runtime
  save/resume oracle、analog mixer gain 與
  CoAB 真實 Timer B reload phase。MSCDRV Timer B-only
  IRQ ownership 及 faithful BGM 不執行 Sound BIOS LFO 已證明；六種
  LFO waveform、完整 Timer B count period 與 PCM 有理數 accumulator
  已完成；
  pitch／TL 數學核心、sync delay 與 Timer B scheduler 已由 ROM harness
  驗證；曲名、十二首 sequence、控制流、total-level、
  algorithm carrier、operator-mask key-on、正常配樂 deterministic
  events 及啟動 S98 register trace 已交叉驗證；
  driver sector 仍需
  恢復，但已證明不與 84 個 channel stream 重疊。
  `WLDTWN` scene-role、ECL block → selector 與同 block 內 selector 5↔6
  context cue、7Eh play／stop → D2h bridge，以及 Sound BIOS command ABI
  已完成，不再列為缺口。
- 完整 DOS save serialization、未知欄位、所有 sidecar 副作用與跨 Gold Box
  作品角色轉移。
- Windows／Linux／macOS 發行包、長時間遊玩、全路線通關及回歸驗證。

## 可重現驗證

- 共用 engine：`go test ./...`
- CoAB 正式程式：在 Docker／Xvfb 執行
  `go test ./cmd/... ./gamepack ./internal/...`。目前 `go test ./...` 另會因
  `scripts/` 保存兩個可各自 `go run` 的獨立 `main()` 而 build failed；這是
  script 目錄結構 gate，不可誤報成正式套件測試失敗或全綠。
- 原版畫面：Docker 內以 DOSBox 啟動本地原始發行檔，oracle 保存在
  `docs/reference/original-dos/`。
- 最新公開畫面保存在 `docs/screenshots/`，README 只引用實際產生的 PNG。

## 後續工具

若恢復反組譯，可使用使用者提供的 IDA Pro：

`/home/anr2/ida_94_official/dist`

優先把 IDA 發現整理成 `docs/spec/` 的 READY 規格，再修改 engine／game-pack。
目前已恢復作業；重大畫面進度集中 commit／push。
