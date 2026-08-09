# 《青色枷的詛咒》繁體中文化／Remake

這是 SSI《Curse of the Azure Bonds》的資料驅動重製專案。目標是讓玩家以繁體
中文從開場遊玩到結局；目前仍是可執行的多垂直切片 prototype，尚未宣稱完整
通關或完整還原原版。

本 repo 負責 CoAB 的 game pack、翻譯、原始素材轉換、攻略與整合測試；可重用的
ECL／DAX／GEO／戰鬥／存檔引擎位於獨立的
[golden-box-remake-engine](https://github.com/wicanr2/golden-box-remake-engine)。
劇情、座標、物品與翻譯資料不寫死在共用 engine。

## 目前可以看到什麼

已接通的正常玩家路徑包含提爾佛頓開場、城市設施、城門／盜賊公會、下水道前段、
火刀戰後的 Myth Drannor handoff，以及部分 Burial Glen／外圍遺跡／眼魔洞窟戰鬥。
地城移動、ECL continuation、繁中手札、資料驅動選項、戰鬥效果與 640×480 原版
風格畫面已有多個可重播切片；角色建立現在也使用原版裂紋石框與倚天粗體 16×15
密度；本輪並將 12 個既有玩家戰鬥法術的 CAST 入口、目標
模式、施法時間與訊息 ID 接到 engine＋JSON；路徑仍會在後續區域遇到未完成的功能。

最新完成度與驗證邊界請看
[專案成果盤點](docs/project-status.md)。第 531 輪完成角色建立的原版石框／中文字級
版面基線；第 529 輪完成玩家戰鬥法術的資料契約，
不把法術名稱或劇情塞回 Go；同時保留 P0-2 的 PC-98 MOVEPARTY 靜態證據，沒有猜測
秘密門規則。詳細證據見[第 531 輪規格](docs/spec/531-character-creation-original-frame-density.md)、
[第 529 輪規格](docs/spec/529-combat-player-spell-data-contract.md)、
[第 528 輪規格](docs/spec/528-pc98-moveparty-action-transaction-boundary.md)、
[第 527 輪規格](docs/spec/527-pc98-moveparty-action-writer-boundary.md)與
[反組譯工作清單](docs/knowledge/golden-box-reverse-engineering-worklist.md)。

## 畫面預覽

以下是目前 remake 的代表性畫面；每張圖的來源平台與 fidelity 等級以成果盤點和
對應 spec 為準。

![繁中冒險畫面](docs/screenshots/gold-box-layout-adventure.png)

![角色建立畫面：原版裂紋石框與 16×15 倚天字級](docs/screenshots/character-creation-remake-640.png)

![繁中戰鬥畫面](docs/screenshots/gold-box-layout-combat.png)

![提爾佛頓第一人稱畫面](docs/screenshots/tilverton-first-person-remake.png)

![Burial Glen 戰鬥畫面](docs/screenshots/burial-glen-red-web-spiders.png)

更多地圖、人物舞台、戰鬥時間軸與素材圖在[截圖目錄](docs/screenshots/)；原版忠實
theme 與日後美化 theme 分開維護。

## 玩家文件

- [繁體中文攻略入口](docs/guide/README.md)：目前可玩的主線、分區攻略與無雷提示。
- [繁中遊玩手冊](docs/manual/curse-of-the-azure-bonds-zh-TW.md)：按鍵、戰鬥、紮營、
  手札與遊玩說明。
- [給中文玩家的 Gold Box 重製說明](docs/knowledge/golden-box-remake-for-chinese-readers.md)：
  將原本需要查紙本手冊的資訊整合到遊戲內與文件中。

## 開發與研究文件

- [專案狀態](docs/project-status.md)：目前已完成、未完成與驗證方式。
- [正常玩家路徑](docs/knowledge/golden-box-normal-player-path.md)：只記錄真正由輸入
  抵達的路徑；座標輔助會明確標示。
- [Gold Box engine 知識庫](docs/knowledge/)：ECL、圖像、戰鬥、音訊、存檔與 UI
  的可重用知識。
- [反組譯工作清單](docs/knowledge/golden-box-reverse-engineering-worklist.md)：
  只追查會改變遊戲可玩結果的資料流，不逐行翻譯無關函式。
- [格式規格索引](docs/spec/README.md)：每份規格都標示 `READY`／`DRAFT`、證據與
  `exact`／`strong inference`／`hypothesis`／`unknown` 邊界。
- [歷輪 README 報告](README-history.md)：保留原本的長篇里程碑與研究紀錄，僅作
  歷史查閱，不作目前狀態的權威來源。

## 執行原則

1. 先讓正常玩家路徑可走，再補 fidelity 與非必要的深層逆向。
2. 遊戲文字、選項、物品與事件由 CoAB JSON／locale 提供；Go／engine 保留可重用
   機制，不複製劇情常數。
3. 反組譯只做到支撐一項可玩功能所需的最小證據；原始位址、bytes、位址空間與
   推論等級都保留，沒有玩家影響的 function 不列入深挖範圍。
4. 每個重大、可展示的里程碑才集中 commit／push；兩個 repository 分開提交。

## 目前明確未完成

完整 ECL／外部 routine、開場到結局的所有正常路徑、全地圖、完整 AD&D 規則、
戰鬥 AI 與所有法術演出、音樂音效、完整原版存檔、全量繁中校對，以及 Windows／
Linux／macOS 發行包仍未完成。因此目前不製作正式三平台 release 或宣傳片，待可玩
整合達到驗收門檻後再安排。

原始遊戲、手冊、圖片與音樂的權利仍屬各自權利人；本專案保存研究證據與新撰寫的
繁中資料，不重新散布未授權的原始遊戲映像。
