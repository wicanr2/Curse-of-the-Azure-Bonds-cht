# 第三百三十九輪：摩安德祭壇藏寶與手札 20

狀態：`READY`

## 原版資料證據

ECL3 block `0x12` 的 SearchLocation selector `0x10` 對應共用 GEO3 block
`0x11` 的 `(12,0)`、terrain `0x90`：

- 只有玩家主動 SEARCH（`7ECA=1`）且面向東／西時才進入藏寶分支；
- `4C4D` bit 0 是一次性狀態，首次搜索後立即消耗；
- 顯示 `YOU HAVE FOUND A CACHE OF JEWELS AND GEMS`；
- TREASURE packet 給予 20 顆寶石、6 件珠寶及 ITEM3 block `0x10`；
- 該 block 解出 `+2 Fine Clerical Scroll` 與 `Gauntlets/Gloves +2`；
- 緊接的 COMBAT 沒有 monster spawn，是 treasure-service boundary；
- 關閉財寶選單後顯示神殿地圖，並記錄 Journal Entry 20。

## 引擎與中文契約

- `ITEM1.DAX`～`ITEM6.DAX` 必須以 `(area << 8) | block` 命名空間載入；
  不能只用 raw block ID，否則不同章節同號 block 會互相覆蓋。
- `TREASURE + COMBAT` 只有在存在 monster spawn 時才是戰後獎勵；
  零 spawn 的 COMBAT 必須先開啟財寶服務，不得顯示假戰鬥。
- 地城 SEARCH 開啟的財寶 UI 必須保存 ECL continuation ownership；
  選擇物品或略過後續跑同一 session，再返回 ModeDungeon。
- 中文顯示祭壇藏寶、神殿地圖與手札第 20 條；解鎖時機保持在原版地圖文字
  出現之後，不可因 ITEM block 載入而提前解鎖。

## 可沿用的 Gold Box 知識

TREASURE 的幣值與 ITEM block 必須先完整解析成功，再原子寫入 pool；素材缺失時
不可只加入寶石／珠寶而留下物品 request，否則重試會重複結算。服務 boundary
的 return mode 與 resume flag 屬於 engine state，不是 renderer state。

## 驗收

- 同一 real-session 取得摩安德護手後返回地城。
- 在 `(12,0)` 面西主動 SEARCH，顯示繁中祭壇藏寶訊息。
- 寶石由 8 增至 28、珠寶由 4 增至 10。
- ITEM3 block `0x10` 顯示兩件繁中物品並開啟財寶選單。
- 關閉選單後顯示神殿地圖文字並解鎖繁中手札 20。
- 返回地城後再次 SEARCH，不得重複取得財寶。
