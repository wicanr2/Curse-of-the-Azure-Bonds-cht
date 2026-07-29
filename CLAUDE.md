# Curse of the Azure Bonds（青色枷的詛咒）中文化／Remake

> 目前所有 agent／compact 後工作規則以 [`AGENTS.md`](AGENTS.md) 為單一入口。
> 本檔保留使用者最初需求與資料索引，不再重複可能過期的 checkpoint。

## 原始目標

1. 透過反組譯還原 SSI Golden Box 引擎，建立類似 ScummVM 的 bytecode VM，
   以 Go／Ebiten 完成可玩的跨平台 remake。
2. 完整繁體中文化 UI、message、規則、物品、法術與劇情。
3. 將原作需要翻閱手冊的 Journal／手札整合進遊戲內。
4. 整理中文說明書、攻略與 Golden Box 歷史／技術知識庫。
5. 還原音樂、音效、戰鬥、地圖、存檔與由開場到結局的完整玩家路徑。
6. 把可重用部分集中在獨立 `golden-box-remake-engine` repo，遊戲內容以 JSON
   game pack 提供，不能 hardcode 在共用引擎。

## 核心工作法

反組譯／實機證據 → `docs/spec/` 規格 → 標記 `READY` → 實作 → 測試與
玩家路徑驗證 → 更新 Markdown／截圖 → 重大進度集中 commit＋push。

必須清掉被新證據推翻的斷言。IDA/decompiler 輸出本身不等於證明；需以原始
bytes、runtime capture 或另一權威來源交叉驗證。

## 原始資料

- 遊戲 image：`curseoftheazurebonds.zip`
- 中文資料：`珍020-青色枷的詛咒.rar`
- DOS Manual：`Curse-of-the-Azure-Bonds_Manual_DOS_EN.pdf`
- Adventurer's Journal：
  `Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf`
- Clue Book：`Curse-of-the-Azure-Bonds_Misc_DOS_EN_Clue-Book.pdf`
- 工作目錄：`workplace/`
- Golden Box 評估：`GOLDEN_BOX_RE.md`

## Repositories 與工具

- CoAB GitHub：<https://github.com/wicanr2/Curse-of-the-Azure-Bonds-cht>
- 共用 engine：`golden-box-remake-engine/`
- IDA Pro：`/home/anr2/ida_94_official/dist`
- 倚天字型：`/home/anr2/cht/etan_font`
- 倚天粗體參考：`/home/anr2/scummvm/monkey_island2`
- 其他 remake 參考：`/home/anr2/cht/daemon_winter`

完成狀態、Git 操作、視覺/字型方向、證據標籤、驗證門檻與 compact 恢復流程
均見 [`AGENTS.md`](AGENTS.md)；歷史紀錄見 [`CONTEXT.md`](CONTEXT.md)。
