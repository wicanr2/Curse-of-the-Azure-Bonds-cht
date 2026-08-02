# 第四百七十四輪：青色印記夢境與提爾佛頓返城資料化

狀態：`READY`

## 範圍

本輪將七個作品文字 boundary 移入 CoAB game-pack：提爾佛頓高階祭司問候、
提爾佛頓城外、衛兵禁止返城，以及火刀首領戰後的第一夜、四位主人嘲諷、服侍
預言與冷汗驚醒四段青色印記夢境。每條規則保存 raw ECL token 與 en／zh-TW
訊息；舊 State fallback 與提爾佛頓兩筆 UI locale 副本已移除。

酒館傳聞 44／60 仍暫留舊 UI catalog，因目前沒有對應的實際 ECL 玩家觸發回歸。
它們是明確技術債，不應只靠單元 rule 測試便宣稱已完成資料化。

## 行為證據

- 高階祭司由真實新遊戲 session 在 GEO2 `(1,10)` terrain `8Fh` 觸發，保留
  PICTURE／HEAD／BODY `6`、YES／NO、解除詛咒敘事與手札 19 continuation。
- 火刀首領勝利後保留財寶、手札 54／53、BIGPIC 120 與四段夢境；產品測試會
  收集實際顯示訊息，逐一要求四個 stable ID 均曾出現。
- 夢醒後 `7F12=1 → NEWECL 50h`，BIGPIC 121 顯示城外選單。測試現在實際選
  ENTER CITY，驗證 `GUARDS BAR YOUR WAY`，按鍵返回同一城外選單後才選
  JOURNEY ON；不再跳過禁止返城分支。

raw ECL regression 仍驗證 `FIRST NIGHT OUTSIDE THE CITY`、BIGPIC 120、
`COLD SWEAT`、`7F12=1` 與 block `50h`，不以翻譯測試取代 source oracle。

## 驗證

- `TestTilvertonBondDreamAndReturnBoundaryIsGamePackDriven`：七條 en／zh-TW 規則。
- `TestRealNewGameBeginsAtGlobalBlockOne`：高階祭司真實事件鏈。
- `TestRealFireKnifeLeaderEncounterAndBondProgression`：raw 夢境與 block handoff。
- `TestFireKnifeLeaderStateVictoryReturnsToTilverton`：四段 stable-ID 顯示、禁止返城
  與後續世界旅行。
- Go 漢字基線：`553 → 545`；`localization_debt 58 → 50`，frontend 135、
  runtime 360 不變。下降八次是因高階祭司譯文原本同時存在 batch 與 line fallback。

本輪未改 renderer，不新增 README 截圖；也不能由這段回歸宣稱所有世界旅行分支
或完整主線已完成。
