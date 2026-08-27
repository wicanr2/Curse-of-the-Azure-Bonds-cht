# Curse of the Azure Bonds（青色枷的詛咒）中文化／Remake

> compact／接手先讀 [`HANDOFF.md`](HANDOFF.md)；所有工作規則仍以
> [`AGENTS.md`](AGENTS.md) 為單一權威。本檔只保留使用者最初需求與資料索引，
> 不重複可能過期的 checkpoint。

## 原始目標

1. 透過反組譯還原 SSI Golden Box 引擎，建立類似 ScummVM 的 bytecode VM，
   以 Go／Ebiten 完成可玩的跨平台 remake。
2. 完整繁體中文化 UI、message、規則、物品、法術與劇情。
3. 將原作需要翻閱手冊的 Journal／手札整合進遊戲內。
4. 整理中文說明書、攻略與 Golden Box 歷史／技術知識庫。
5. 還原音樂、音效、戰鬥、地圖、存檔與由開場到結局的完整玩家路徑。
6. 把可重用部分集中在獨立 `golden-box-remake-engine` repo，遊戲內容以 JSON
   game pack 提供，不能 hardcode 在共用引擎。

完成口徑（2026-08-27）：以玩家可見體驗近似原版 **99%** 為目標；驗證採
風險導向代表性抽樣。已知近似與未驗範圍須明列，但不再要求所有狀態全量逐像素、
逐分支或逐週期一致。詳細閘門以 `AGENTS.md` §1 為準。

## 核心工作法

反組譯／實機證據 → `docs/spec/` 規格 → 標記 `READY` → 實作 → 測試與
玩家路徑驗證 → 更新 Markdown／截圖 → 重大進度集中 commit＋push。

必須清掉被新證據推翻的斷言。IDA/decompiler 輸出本身不等於證明；需以原始
bytes、runtime capture 或另一權威來源交叉驗證。

**2026-08-13 起先盤點後語意**：不再「遇到問題才反組譯那一段」。兩平台全模組
先建庫盤點，每個函式都要進覆蓋台帳，再依玩家可見性排序做語意閉合。口徑與
理由見 `AGENTS.md` §2.5，方法見
[`docs/spec/559-full-module-re-sweep.md`](docs/spec/559-full-module-re-sweep.md)。

### 通則：用公開攻略協助金盒子 remake（2026-08-24）

金盒子這一族的遊戲**玩家可見層在網路上早就被寫完了**——每一張圖的地點清單、
房間順序、進出口、關鍵物品。那些問題**不要用反組譯去問**：查攻略便宜太多，
而反組譯的成本要留給攻略答不了的東西（旗標語意、公式、時序、儀器的洞）。

四條，順序就是使用順序：

1. **先查攻略，再決定要不要反組譯。** 問題若是「這一格有什麼」「這張圖怎麼進出」
   「該去哪」，先查。反組譯留給「為什麼」「這個值從哪來」「原作在哪一條分支決定」。

2. **攻略是語意的交叉核對，不是座標的來源。** 座標、旗標、地形碼**一律以原始
   GEO／ECL 為準**；攻略只用來替已經解出來的東西**命名**（把索引 7 叫「首領」）
   並確認語意沒讀反。攻略會錯也會漏——提爾佛頓下水道的攻略寫「兩個往據點的
   出口」，原始資料是三個。

3. **攻略最大的用處是當檢查清單，用來抓「盤點少了一整塊」。** 它告訴你**該存在
   什麼**；查詢回空的時候，那份清單讓你分得出「原作真的沒有」與「自己的掃描面
   有洞」。火刀據點那 33 個索引整批不在每格事件表裡，攻略列得出九個地點——
   對一次就知道是儀器壞了，不是遊戲沒內容。⇒ **攻略說有、資料裡找不到，先懷疑
   自己的儀器**（配 `~/diagnosis-notes/docs/02-query-returned-empty/`）。

4. **絕對不要把攻略的座標直接抄進 game pack。** JSON 的宣告會**蓋過原始資料**
   （`external_exit` 讓 `CanMoveDungeon` 不看 GEO 直接回 true），所以抄錯一格的
   症狀是「測試綠、玩家走不到」——兩者在報表上分不出來。每一格宣告都要先用
   `cmd/geo-move-mask` 對回原始資料（spec 1199）。

對照結果要**寫進 repo 並附來源網址**，不要留在對話裡；本專案放在
[`docs/reference/maps/README.md`](docs/reference/maps/README.md)，地圖本身由
`cmd/map-atlas` 產生。

## 原始資料

- 遊戲 image：`curseoftheazurebonds.zip`
- 中文資料：`珍020-青色枷的詛咒.rar`
- DOS Manual：`Curse-of-the-Azure-Bonds_Manual_DOS_EN.pdf`
- Adventurer's Journal：
  `Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf`
- Clue Book：`Curse-of-the-Azure-Bonds_Misc_DOS_EN_Clue-Book.pdf`
- 工作目錄：`workplace/`（全部 gitignore，含 git 目錄與 IDA 產物）

## 反組譯輸入（兩平台）

| 平台 | resident | overlay 容器 | 特徵 |
|---|---|---|---|
| DOS | `START.EXE` | `GAME.OVR`（TPOV，36 段） | 行為 oracle；**沒有**除錯符號 |
| PC-98 | `GAME.EXE` | `GAME.OVR`（TPOV，36 段） | 帶完整 Borland 除錯符號，語意骨幹 |

原始 Turbo Pascal 單元名（PC-98 符號表）：`GAME`、`INTERPET`、`COMBAT`、
`SPELLS`、`EFFECTS`、`MOVEMENT`、`TACMAP`、`THREED`、`LOS`、`OVERLAND`、
`LOADSAVE`、`SHOP`、`TEMPLE`、`TRAINING`、`CAMP`、`MENUS`、`PORTRAIT`、
`SOUNDX`、`GLOBALS` 等 49 個遊戲模組（另有 4 個 Turbo Pascal RTL）。

## Repositories 與工具

- CoAB GitHub：<https://github.com/wicanr2/Curse-of-the-Azure-Bonds-cht>
- 共用 engine：`golden-box-remake-engine/`
- IDA Pro：`/home/anr2/ida_94_official/dist`，image `ida-pro-9.4-idapython:py312-v1`
  （只有這顆能跑 IDAPython），入口一律 `tools/ida.sh`
- 全模組全掃：`tools/re-sweep.sh <dos|pc98> <exe> <ovr>`
- 倚天字型：`/home/anr2/cht/etan_font`
- 倚天粗體參考：`/home/anr2/scummvm/monkey_island2`
- 其他 remake 參考：`/home/anr2/cht/daemon_winter`

完成狀態、Git 操作、視覺/字型方向、證據標籤、驗證門檻與 compact 恢復流程
均見 [`AGENTS.md`](AGENTS.md)；現況見 [`CONTEXT.md`](CONTEXT.md)，
歷史分冊在 [`docs/context/`](docs/context/)。
