# 專案成果盤點

更新日期：2026-07-30
本 milestone 的 CoAB 基底：`825de71`
依賴的 Golden Box engine checkpoint：`f9fbcaf`

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
  第一人稱場景以 cover 填滿左上內格，HEAD／BODY 人物使用獨立固定舞台。
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
- 角色建立、繁中姓名、能力值、種族／職業、基礎隊伍、裝備與多項法術。
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
- remake JSON save，以及原版 SAVGAM／CHRDAT／FX／SWG 的已驗證欄位
  import、raw preservation 與部分 writeback。
- 角色年齡 offset `0x76..0x77`、種族／職業 mapping 與 DOS 實機角色頁證據。
- 中文手冊、攻略、Gold Box 技術知識庫、READY 規格與 README 實機截圖。
- PC-98 `VFD1.00` 唯讀稽核工具、兩張磁碟雜湊與 absent CHRN 已建立；
  NP2kai Docker 實機可開機至 MEGDOS／`loader.com`。`MSCDRV.EXE` 已確認為
  `INT D2h`／YM2203 (`0x188/0x18A`) 常駐驅動；MAME FDI 身分雜湊與 loader
  三段 EXEC 順序也已交叉驗證。absent sectors 疑似同時參與早期完整性／
  防拷，driver 中間 1 KiB 尚未取回，未宣稱完整 driver 已復原；十二首
  sequence 已證明不跨該缺口，現可由殘存 exact driver 合成播放。
- PC-98 `GAME.OVR` 36 段 TPOV code／relocation 已可重現解碼；`GAME.EXE`
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
  `27h` reload phase、save/resume 與 analog mixer gain。
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
- 第 383 輪修正 ECL session 第一次明確 `RunFrom` 會遺失預載存檔／區域
  記憶體的生命週期缺陷。原始 ECL1 block `0x50` 回歸已證明
  `4C59=1／4C5A=1／4C5B=FF` 時工作計數為 3，灰袍人會揭露自己是
  提朗瑟克斯，並要求隊伍前往 Myth Drannor。第 384／385 輪已接通
  ECL6／GEO6 正常入口與第一個精靈幽魂事件；其餘 Burial Glen、迷斯卓諾
  遺跡、最終神殿與結局仍未完成。

## 尚未完成

- 全部 ECL opcode、外部 routine、副作用及由開場到結局的完整可通關流程。
- 所有城市、地城、門、屋頂、斜向視角與每張地圖的 DOS 像素級校準。
- 戰鬥畫面的完整 DOS oracle 校準、弓箭／法術 projectile 的逐距離 timing、
  所有方向 placement、Fireball 牆面阻擋／同距排序、Lightning Bolt
  牆角／多次反彈 runtime oracle、毒雲每回合重複判定與魔法保護例外、AI、
  Cloudkill DOS 動態時間碼、其餘法術、物品、
  特殊能力、逃跑／交涉與戰後流程。
- 全角色／怪物／物品／法術 AD&D 規則及完整多職業、alignment、升級規則。
- 全英文文本、59 則 Journal（目前新增完成 50／51）、Tavern Tales、
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
  driver 內部 FM SFX 的真實 producer、save/resume、analog mixer gain 與
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
