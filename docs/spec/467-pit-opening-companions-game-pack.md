# 第四百六十七輪：摩安德之坑開場與同伴事件資料化

狀態：`READY`

## 範圍

本輪把摩安德之坑前半章十四個真實畫面邊界移入 CoAB game-pack：

- `pit.opening-dead-cultists`、`pit.opening-chosen`
- `pit.trapped`、`pit.cleric-dies`、`pit.ambience`
- `pit.alias-dragonbait-meet`、`pit.alias-bonded-reaction`
- `pit.alias-dragonbait-introduction`、`pit.alias-dragonbait-join`
- `pit.alias-dragonbait-joined`
- `pit.stairs-down`、`pit.stairs-up`
- `pit.dead-zhentrim`、`pit.zhentrim-scroll`

原始 ECL fragments 與 en／zh-TW 訊息均由 game-pack 保存。開場不是單一段落：
三具屍體／垂死牧師與「被選中之人」之間有 PRESS boundary，因此保留兩個 stable
IDs，不能按原始故事內容自行合併。

## 正常玩家路徑

第 466 輪長路徑由 Yulash terrain `26h` 繼續，依序驗證：

1. `NEWECL 11h`／GEO3 出生點 `(0,0,E)`、垂死牧師與崩塌陷阱。
2. 牧師死亡、地城 ambience 與返回同一層 GEO。
3. terrain `85h` 遇見愛麗雅絲與龍餌，經 PARLAY／NICE、手札 3、YES 入隊。
4. roster 保存 ALIAS Fighter 與 DRAGONBAIT Saurial Paladin 的 stable identity。
5. 南側下樓至 ECL `12h`，再看到北牆上樓提示。
6. 散塔林屍體的 EXAMINE 分支、手札 46 訊息、一次性消耗後不重播。

## 驗證與邊界

- 十四條規則在 en／zh-TW 均命中唯一 rule ID、非空訊息且不附帶額外頁面。
- 玩家路徑訊息由正式 game-pack stable ID 即時取得；測試不複製目前譯文。
- 手札 3 仍由既有 journal rule 解鎖；手札 46 目前只有原作提示訊息，完整頁面
  尚未補入遊戲內，不能宣稱整本手札完成。
- Go 漢字 literal exact baseline 由 638 降至 621；`localization_debt`
  143→126，frontend 135、runtime 360 不變。

本輪未改 renderer，也未產生新截圖。摩貢祭壇、枷印儀式、兩場最終戰、護手、
財寶、離場阻擊與同伴離隊仍由後續 milestone 處理。
