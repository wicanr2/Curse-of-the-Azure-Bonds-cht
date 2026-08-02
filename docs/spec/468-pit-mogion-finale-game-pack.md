# 第四百六十八輪：摩貢祭壇與摩安德之坑結局資料化

狀態：`READY`

## 範圍

本輪把摩安德之坑後半章十八個真實畫面邊界移入 CoAB game-pack：

- 摩貢祭壇、愛麗雅絲辨認大祭司、摩貢招呼。
- 枷印麻痺、同伴遭藤蔓束縛、儀式吟唱、異次元窗口。
- 摩安德返回、枷印褪去／消失、同伴脫困並催促攻擊。
- 裂隙閉合、三塊殘軀吶喊／攻擊、取得摩安德護手。
- 祭司逃跑、祭壇財寶、離坑最後阻擊。

每條規則保存原始 ECL fragments 與 en／zh-TW 訊息。State 不再知道摩貢、
摩安德儀式、護手或最後阻擊的作品中文。

## 正常玩家路徑

第 467 輪長路徑由祭壇 terrain 繼續，逐段驗證：

1. PICTURE 17、八段儀式 boundary 與 `ATTACK／FLEE` 選單。
2. 摩貢 ×1、教徒 ×6、蔓生怪 ×5 的第一戰，愛麗雅絲與龍餌仍在我方。
3. 戰後裂隙閉合，三個 140 HP、四格 footprint 摩安德殘軀進入第二戰。
4. 勝利取得護手並寫 `4C5Bh=1`，祭司逃跑後回到地城。
5. 搜尋祭壇取得 28 gems、10 jewelry、正式 item block 與遊戲內手札 20；重搜
   不複製財寶。
6. 離場阻擊由 sprite blocks `11h×10／1Ch×5／19h×5` 組成。
7. Alias alive 分支正常離隊，護手旗標成為 `FFh`、進度 `7F12h=1`，hero roster
   保留並返回 Yulash edge。

## 驗證與邊界

- 十八條規則在 en／zh-TW 均命中唯一 rule ID、非空訊息且不附帶額外頁面。
- 八段儀式逐一由 stable ID 驗證，不再將整串中文做模糊包含比對。
- Go 漢字 literal exact baseline 由 621 降至 603；`localization_debt`
  126→108，frontend 135、runtime 360 不變。

這證明目前測試路徑可完成摩安德之坑 chapter transaction，不代表 Alias dead、
所有交涉／逃跑分支、完整怪物能力、原版動態演出、音效或全遊戲已完成。
