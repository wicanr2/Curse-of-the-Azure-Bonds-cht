# 《青色枷的詛咒》繁體中文攻略

這裡收錄的是本專案自行整理、重新撰寫的繁體中文攻略。內容以 repo 內隨附的
Adventure Journal、Clue Book、反組譯結果及可重現測試為依據，目標是讓玩家能
實際走完目前 remake 已接通的主線，而不是逐字翻譯任何既有英文攻略。

## 從哪裡開始

- [目前可玩主線攻略](./main-story-zh-TW.md)：從提爾佛頓開場一路走到尤拉什的
  指揮官等候室。
- [繁中遊玩手冊](../manual/curse-of-the-azure-bonds-zh-TW.md)：按鍵、戰鬥、
  紮營與背景資料。
- [格式與事件規格索引](../spec/README.md)：攻略背後的反組譯及驗證證據；這不是
  遊玩時必讀文件。

## 進度標記

攻略使用兩種醒目標記：

- **remake 已實作**：已有 READY 規格及自動化流程驗證，目前版本可操作。
- **原版後續／尚未實作**：只供理解完整原版旅程，不保證目前 remake 能繼續。

本文所稱「已實作」是指事件主線已接通，不表示所有 AD&D 規則、所有支線、每個
談判結果或所有原版介面細節都已完成。若實際版本與攻略不同，請先查看攻略開頭
的「適用進度」及 repo 最新提交。

## 遊戲內攻略地圖

遊戲中按 `F3` 會以目前 GEO 座標開啟 16×16 攻略地圖。預設只顯示已探索格與
已到達的事件點；按 `V` 可切到完整攻略，第一次會先顯示劇透警告。現行事件點
存在 `assets/guide/maps.zh-TW.json`，每點保留 GEO／ECL 或本攻略的證據路徑。

使用者指定以《軟體世界》攻略為說明基礎；但目前 repo 並沒有該完整攻略掃描，
只有《軟體世界》中文手冊／手札掃描的部分引用。因此未取得掃描頁的點一律標為
本專案攻略、Clue Book 或 GEO／ECL 整理，不冒稱為《軟體世界》原文。

## 資料來源與著作權

主要本地來源如下：

- [`Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf`](../../Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf)
- [`Curse-of-the-Azure-Bonds_Misc_DOS_EN_Clue-Book.pdf`](../../Curse-of-the-Azure-Bonds_Misc_DOS_EN_Clue-Book.pdf)
- `Curse-of-the-Azure-Bonds_Misc_DOS_EN_Docs-as-TXT.zip` 內的 Adventure
  Journal 純文字轉錄
- [`docs/spec/`](../spec/) 中標為 `READY` 的事件規格
- [`docs/manual/curse-of-the-azure-bonds-zh-TW.md`](../manual/curse-of-the-azure-bonds-zh-TW.md)

遊戲名稱、原始劇情、圖片、地圖、手冊及其他原作內容的權利仍屬其各自權利人。
本攻略只摘錄完成路線所需的事實，中文敘述均重新摘要、編排；不重製完整英文
攻略，也不取代合法取得的原版遊戲與手冊。
