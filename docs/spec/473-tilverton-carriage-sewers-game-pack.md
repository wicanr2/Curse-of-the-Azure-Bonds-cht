# 第四百七十三輪：皇家馬車與提爾佛頓下水道資料化

狀態：`READY`

## 範圍

本輪將同一條真實新遊戲 integration chain 上的十五個作品文字 boundary 移入
CoAB game-pack：

- 皇家馬車封路、讓路、青色印記強迫、假國王、警鐘與紅袍綁架；
- 投降、入獄、盜賊救援及抵達盜賊公會；
- 下水道火刀檢查哨、藏起屍體、迷斯卓諾騎士現身、效忠問題與公主分支。

每條規則保存 raw ECL token 與 en／zh-TW 訊息。較具體的下水道投降規則排列在
只含 `DO YOU SURRENDER` 的通用馬車投降規則之前，避免同一 batch 被寬規則攔截。

## 真實行為證據

`TestRealNewGameBeginsAtGlobalBlockOne` 使用原始 DAX／GEO／MON 資料與同一 ECL
session，驗證：

- PICTURE 11 皇家馬車、四段 pause 與五名皇家衛兵實戰；
- 戰後紅袍綁架、投降、牢房、PICTURE／HEAD／BODY `2` 的盜賊救援；
- `NEWECL 2` 抵達盜賊公會，後續混合戰與手札 4；
- block 3 下水道檢查哨五名 FIRE KNIFE 實戰與藏屍 continuation；
- 迷斯卓諾騎士現身、三個效忠選項、公主分支與重訪消耗。

產品訊息全部以 `requireGamePackText` 從 stable ID 取得，不再比對中文片段。測試
仍為新遊戲衍生的長 integration，但城市內使用已知 GEO 座標定位部分事件；因此
不能單獨宣稱逐步移動的開場至公會玩家路徑已完成。

## 驗證

- `TestTilvertonCarriageAndSewersStoryIsGamePackDriven`：十五條 en／zh-TW 規則。
- `TestRealNewGameBeginsAtGlobalBlockOne`：上述完整 continuation 與兩場實戰。
- Go 漢字基線：`569 → 553`；`localization_debt 74 → 58`，frontend 135、
  runtime 360 不變。下降十六次而非十五次，是因舊假國王 fallback 由兩個中文
  literal 串接。

本輪未改 renderer，不新增 README 截圖，也不擴大宣稱皇家馬車或下水道的所有
選擇分支均已完成。
