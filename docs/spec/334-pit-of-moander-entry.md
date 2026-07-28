# 第三百三十四輪：摩安德之坑入口

狀態：`READY`

## 原版資料證據

尤拉什 GEO3 block `0x10` 北側 `(11,0,N)` 的 raw terrain 為 `0x26`。
ECL3 SearchLocation 先顯示摩安德上次降臨所留下的巨坑，要求隊伍向前一步進入。
事件帶有原版 PICTURE／Continue pause，不能直接略過。

向北跨越地圖邊界後，ECL 由 block `0x10` 切到 `0x11`，並載入
GEO3 block `0x11`。外層地城狀態同步為 `(0,0,E)`。入口開場依序為：

1. 地板上躺著三名死去的邪教徒，前方另有一名受傷牧師；
2. 牧師露出狂熱神情，高喊「被選中之人」；
3. 他擊打突出的岩石，使隊伍身後洞頂崩塌，宣告眾人已被困在坑內；
4. 牧師咳血死亡；
5. 遠處傳來戰鬥聲，空氣中有烤麵包氣味；
6. 返回 ECL3/GEO3 block `0x11` 的 Dungeon mode。

Clue Book 將尤拉什地圖位置 12 標為摩安德之坑的單向入口，與 ECL 的崩塌敘事
一致。

## 引擎與中文契約

- 保留 terrain `0x26` 與 PICTURE／Continue／press-menu 的原始 pause 順序。
- boundary attempt 必須送入 `7ED5`，由 ECL 決定 `0x10→0x11`，不可由座標硬切。
- `LOAD FILES` 後必須發布 GEO3 block `0x11` request，不能繼續顯示尤拉什街景。
- 同步 script map registers 為 `(0,0,E)`，並保持 Area 3／InDungeon。
- 五段入口敘事完整繁中；不能只顯示第一段就把玩家丟回地城。
- 崩塌是單向入口的劇情事實，不在尚未解碼出口前提供返回尤拉什的假選項。

## 可沿用的 Gold Box 知識

同一 GameArea 內的地城 boundary 也能觸發 `NEWECL + LOAD FILES`。跨 block 時，
script namespace、GEO block、座標與 resumable runtime 必須視為同一筆轉場；
只更新 ECL block 或只更新 renderer map 都會造成劇情與碰撞資料分離。

## 驗收

- 延續同一 real-session，先取得紅羽衛指揮官通行，再抵達 `(11,0,N)`。
- 驗證 terrain `0x26` 的繁中巨坑警告與原始 picture pause。
- 驗證 boundary lifecycle 切到 ECL3/GEO3 block `0x11`、`(0,0,E)`。
- 逐段驗證邪教徒、牧師、崩塌、死亡與坑內環境繁中。
- 最終必須回到 `ModeDungeon`，且 block／GEO／座標一致。
