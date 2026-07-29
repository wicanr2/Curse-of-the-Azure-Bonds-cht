# 專案成果盤點

更新日期：2026-07-29
本 milestone 的 CoAB 基底：`06fbf96`
依賴的 Golden Box engine checkpoint：`a81b963`

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
- DOS runtime cracked stone chrome 已抽成透明 native raster；事件／人物圖
  cover 填滿左上內格。PC-98 640×400 gallery 已建立字級密度參考索引。
- CHEAD／CBODY 頭身合成、CPIC／SPRIT／COMSPR 戰鬥小人及 animation metadata。
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
- remake JSON save，以及原版 SAVGAM／CHRDAT／FX／SWG 的已驗證欄位
  import、raw preservation 與部分 writeback。
- 角色年齡 offset `0x76..0x77`、種族／職業 mapping 與 DOS 實機角色頁證據。
- 中文手冊、攻略、Gold Box 技術知識庫、READY 規格與 README 實機截圖。
- PC-98 `VFD1.00` 唯讀稽核工具、兩張磁碟雜湊與 absent CHRN 已建立；
  NP2kai Docker 實機可開機至 MEGDOS／`loader.com`。`MSCDRV.EXE` 已確認為
  `INT D2h`／YM2203 (`0x188/0x18A`) 常駐驅動；MAME FDI 身分雜湊與 loader
  三段 EXEC 順序也已交叉驗證。absent sectors 疑似同時參與早期完整性／
  防拷，driver 中間 1 KiB 尚未取回，未宣稱音樂可播放。
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
  `5→6→5`，且同曲不重播。曲名、完整 `INT D2h` 轉送、runtime YM trace
  與播放器尚未完成。
- 真實連續主線已由開場延伸到散塔林堡：內城奧莉芙事件、手札 50／51、
  `ECL4/GEO4 0x20→0x21` 密道、神殿 `(10,6,N)` 操作權、南方牢房導航、
  迪姆斯沃特同行、手札 12 六頁、兜帽女子離場、手札 30／7、弗佐爾死亡、
  第四枷印解除及 `0x21→0x22` 眼魔洞穴入口已有 regression。洞穴
  `(15,1,N)` 的德克薩姆／梅杜莎／眼魔／牛頭人決戰、四件普通戰利品、
  取回洛山達護符，以及 19 名散提爾堡部隊第二戰也已有真實檔案 regression。
  洞窟 boundary terrain `0x93` 的奧莉芙／迪姆斯沃特告別與
  `NEWECL 0x51` 暗影谷返回亦已接通。

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
  曲目尚缺完整 driver sector、runtime YM trace、曲名交叉驗證與匯入播放器。
  `WLDTWN` scene-role、ECL block → selector 與同 block 內 selector 5↔6
  context cue 已完成，不再列為缺口。
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
