# 第三百三十七輪：摩貢祭壇與摩安德裂隙儀式

狀態：`READY`

## 原版資料證據

ECL3 block `0x12` SearchLocation selector `0x0C` dispatch 到 `+0x10B4`。
共用 GEO3 block `0x11` 的下層 `(11,3)` terrain `0x8C` 可重現完整事件：

- 女祭司站在祭壇前，摩安德教徒低聲吟唱；
- Alias 已在隊伍時會認出她是摩安德大祭司；
- PICTURE 17 顯示 MOGION；
- 原版 encounter menu 為 `ATTACK / FLEE / WAIT / PARLAY`；
- 無論玩家先選何種行動，枷印藍光都會使隊伍無法移動；
- 藤蔓纏住 Alias 與 Dragonbait；
- 摩貢從枷印抽取能量，在祭壇上方形成 dimensional window；
- 摩安德的穢物開始穿過裂隙，Moander bond 同時褪去；
- 枷印與麻痺消失，Alias、Dragonbait 脫困；
- 最終選單為 `ATTACK / FLEE`。

選 ATTACK 後，原始 MON3 records 產生：

- `MOGION` ×1：MON3 block `0x18`，60 HP；
- `CULTIST` ×6：MON3 block `0x11`；
- `SHAMBLING MOUND` ×5：MON3 block `0x19`；
- Alias 與 Dragonbait 以 party side 參戰。

連同測試英雄，戰鬥 roster 共 15 名。

## 引擎與中文契約

- 保留 PICTURE 17、兩層 encounter 選單與所有 continuation 邊界。
- 儀式文字完整繁中，明確表達力量來自 Moander bond，且該枷印在裂隙擴大時解除。
- MOGION、CULTIST、SHAMBLING MOUND 分別顯示「摩貢」、「摩安德教徒」、
  「蔓生怪」，但 record identity 與 sprite block 不變。
- ECL 選單 index `0` 不得被通用 State 誤顯示為 world-menu 的「進入城市」；
  ECL continuation 應保留該選項本身的中文 label，本事件為「攻擊」。
- 本輪驗收邊界是摩貢戰正式開始；戰勝、三塊摩安德殘軀與護手屬後續規格。

## 可沿用的 Gold Box 知識

通用 State 不可用選單 index 猜語意。index `0/1/2` 只有在無 ECL 的 world menu
才代表 ENTER CITY／JOURNEY ON／CAMP；ECL menu 的同一 index 可能是 ATTACK、
FLEE 或任意 script option。語意 identity 必須來自 original option text 與
目前 ECL continuation。

## 驗收

- 延續同一 real-session，在 `(11,3)` terrain `0x8C` 觸發祭壇。
- 驗證 PICTURE 17、四項 encounter menu 與最終 ATTACK/FLEE menu。
- 驗證儀式繁中包含癱瘓、藤蔓、異次元窗口、Moander 回歸與枷印解除。
- 選 ATTACK 後 Mode 為 Combat，底部訊息為「攻擊」而非 world-menu 文案。
- 驗證 1 Mogion／6 cultists／5 shambling mounds，Alias 與 Dragonbait 位於
  party side。
