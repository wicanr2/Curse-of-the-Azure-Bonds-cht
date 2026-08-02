# 第四百七十五輪：艾森布拉傳聞與 State 中文 fallback 清理

狀態：`READY`

## 更正第 474 輪判斷

第 474 輪誤稱 Tavern Tale 44／60 缺少實際 ECL 玩家觸發回歸。重新查閱 READY
spec 330 與現行長測試後，證據已存在：同一真實 session 由艾森布拉
`BAR → RELAX` 觸發 60，`HAVE A DRINK → BEER` 觸發 44。spec 474 中該句由本
規格 supersede；不能因 compact 後只看到較新註記而忽略既有深層測試。

## 資料化

- `essembra.tavern-tale-60` 與 `essembra.tavern-tale-44` 保存 raw ECL token、
  en／zh-TW 訊息，並由 CoAB game-pack `text_rules` 匹配。
- State 的兩個 tale fallback、UI locale 的兩份作品傳聞與共用 Tavern catalog
  coverage 要求已移除；共用酒館仍只負責飲品／RELAX／EXIT 機制。
- 產品長回歸以 stable ID 比較完整傳聞，不再 `Contains` 中文片段。

## State line localizer

`localizeECLLine` 尚有訓練、提爾佛頓服務、菲拉妮、武器店、雜貨店與尤拉什的
英文 source line→stable locale ID 映射。這些映射仍是舊 adapter 技術債，但本輪
移除其中 27 個中文 fallback literal，改以 ID 作明確 fallback：

- 正式譯文只存在 `assets/locale/zh-TW.json`；
- key 缺失時顯示 stable ID，而不是靜默使用 Go 內第二份譯文；
- locale coverage、真實新遊戲服務路徑與全套正式測試負責攔截缺 key。

完整 gate 首次移除 fallback 後，揭露四鍵合成 `testCatalog` 與十個訓練／武器店／
雜貨店 ID 的資料缺口。本輪沒有恢復 Go 譯文，而是把十筆譯文補進正式 locale、
讓真實新遊戲長回歸改用正式 catalog、新增 37 個 ECL line stable IDs 的 coverage，
並讓合成 resolver 測試只明確注入它要驗證的兩個 ID。

此變更不是把 adapter 宣稱為作品中立 engine；後續仍應依完整 text batch 逐步移入
game-pack rules，或把真正共用的服務 source mapping 移到資料 schema。

## 玩家與 source 驗證

真實長路徑從火刀主線、法師塔與龍巫妖抵達艾森布拉，驗證：

1. `BAR` 顯示露天酒館及 `HAVE A DRINK／RELAX／EXIT`。
2. `RELAX` 顯示 Tale 60，沒有新增手札 18，Continue 返回同一選單。
3. `HAVE A DRINK` 顯示六項原始飲品；`BEER` 顯示 Tale 44。
4. Continue 返回酒館，EXIT 再返回艾森布拉六場所選單。

raw ECL 選項、原始 tale token、session continuation 與城市返回仍由既有 spec 330
及 integration test 證明；stable-ID 測試不取代它們。

正式驗證另包含 `TestECLLineCatalogCoversEveryDisplayedStableID`、改用正式 catalog
的 `TestRealNewGameBeginsAtGlobalBlockOne`，以及只注入兩個 ID 的
`TestLocalizeECLTextUsesCatalogAndPreservesUnknownLines`。

## 量測

- Go 漢字基線：`545 → 518`。
- `localization_debt 50 → 23`；frontend 135、runtime 360 不變。
- 下降 27 次來自 line localizer 中文 fallback；兩則 Tale 舊分支本來只回傳 locale
  key，本身不含漢字，因此事件數與 occurrence 差值不同。

本輪未改 renderer，不新增 README 截圖，也不代表全遊戲 Tavern Tales 已完成。
