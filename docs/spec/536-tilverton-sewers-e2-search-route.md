# 第五百三十六輪：提爾佛頓下水道 E2 路徑與 Search 模式勘誤

狀態：`SUPERSEDED`（實作與目前 remake 狀態改由第 537 輪規格管理；本檔證據仍有效）
日期：2026-08-10

> 本檔保留第 536 輪「尚未實作」時的 GEO／手冊盤點。第 537 輪已把 Search／Look
> 與 E2 transaction 接入 engine＋JSON；這不會把本檔的原版 wall writer 缺口
> 升格成 `exact`。

## 結論先行

第 535 輪把 `(13,10) → (8,15)` 正確改稱為「E2 外部地圖出口／邊界轉場」。本輪
再用原始 `GEO2.DAX` block 3 的四平面資料做有界 BFS，確認它不是單純把騎士事件
座標 teleport 到據點，而是有一條**只差一個 `wall=09/detail=0` 橋接邊**的正常
地城路徑：

```text
(13,10) W→ (12,10) W→ (11,10) W→ (10,10)
         S→ (10,11) S→ (10,12)
         W→ (9,12)   ← 雙側 wall=09/detail=0；目前普通 movement 會擋住
         S→ (9,13) S→ (9,14) S→ (9,15) W→ (8,15)
         S→ E2 外部地圖邊界       ← 目前 ECL block 3→4 branch
```

關閉 `wall=09` 候選橋接時，`(13,10)` 到 `(8,15)` 不可達；只把該邊暫時視為
可通行時才可達。這證明了目前 remake 的「普通移動先判牆、未知牆面 fail-closed」
會在兩個不同位置停下：中間的 `wall=09`，以及 `(8,15,S)` 的 `wall=0C/detail=0`
E2 邊界。它仍**不證明** `wall=09` 一定由 `SEARCH` 改成 detail `1`，也不證明
任何特定 ECL work byte 是其 writer。

## Search／Look 的手冊勘誤

本輪重新核對本機 DOS 手冊與公開的同版手冊文字：

- `SEARCH` 是持續的開／關模式。開啟後每次前進耗時較長，並提高發現秘密門與
  隨機遭遇的機率。
- `LOOK` 才是只檢查目前格子的單次操作；手冊將它描述成「像以 Search 開啟狀態
  進入目前格子一次」。
- 因此目前 remake 將生產環境 `S` 直接接到一次性
  `State.SearchDungeonLocation()`，是暫時的 `LOOK` 形狀，不是原版 Search
  toggle。既有 `7ECA=1` 的 ECL 事件測試仍可保留為「主動搜尋一次」邊界，但不能
  再把它當成完整 Search mode 已完成。

本輪採用的原始手冊輸入：

| 輸入／工具 | 雜湊／版本 | 用途 |
|---|---|---|
| `Curse-of-the-Azure-Bonds_Manual_DOS_EN.pdf` | SHA-256 `d4a3fc873a983cd7c1b84414caf3f8aad77bce1e3518ebccac7d77f80f73ff8c` | Search、Look、前進時間與秘密門說明 |
| [公開手冊文字／PDF](https://www.freegameempire.com/games/Curse-of-The-Azure-Bonds/manual-pdf) | 網頁快取；僅作文字交叉核對 | 同一操作語意；不取代本機檔 |
| `coab-go-test:20260729`／Go `1.24.13` | Docker 內執行 | GEO2 route audit |
| `scripts/research/tilverton_e2_route_audit.py` | repo 版控工具 | 唯讀解碼／BFS，不修改 archive／runtime |

## 地圖與攻略交叉證據

[GameFAQs 的提爾佛頓下水道地圖／攻略](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)
將該區標為 `ECL Script 3`，列出 `(13,10)` 的 Knights of Myth Drannor 事件，並把
`E2` 定義為通往 Fire Knife Hideout 的出口；同一地圖圖例也把 `~` 作為秘密門。
這支持「Search 與隱藏通路相關」及「終點是 E2」，但攻略 ASCII 地圖與本機 GEO
plane 不可直接當成同一個 byte-for-byte 座標圖，所以這些仍是 `nearby`，不是
本機 `wall=09` writer 的 `exact` 證據。

[RPG Gamers 的遊玩紀錄](https://rpggamers.com/walkthrough/curse-of-the-azure-bonds)
也把訓練廳秘密區與下水道末端標示為 Hideout 出口；它能區分「秘密訓練廳」與
「E2 據點出口」，不能單獨證明中間 wall code 的語意。

## 本機 GEO 證據

輸入 archive `curseoftheazurebonds.zip` 的 SHA-256 為
`c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d`。工具解碼
`GEO2.DAX` block 3 的 16×16 payload：wall plane 使用 `+000h/+100h`，terrain
使用 `+200h`，detail plane 使用 `+300h`；這些是原始 record／engine geometry
位址空間，不與 ECL work address 混用。

可重生命令（必須在 Docker 內）：

```text
python3 scripts/research/tilverton_e2_route_audit.py curseoftheazurebonds.zip
```

關鍵輸出：

```text
allow_wall_09=false reachable=False
allow_wall_09=true reachable=True
(10,12) --W wall=09 detail=0 other_wall=09 other_detail=0--> (9,12)
(9,13) --S wall=0F detail=1 other_wall=0F other_detail=1--> (9,14)
e2_candidate_boundary=(8,15,S) wall=0C detail=0 other=(8,0,N) wall=0C detail=0
```

其中 `wall=0F/detail=1` 只表示目前 engine movement contract 判定該雙側邊可走；
本文件不替 raw wall nibble 命名成某種門。`wall=09/detail=0` 的「秘密門／Search
發現候選」是 `strong inference`：由唯一橋接、攻略圖例與手冊 Search 語意共同
支持；尚缺同版 runtime 的 before／after detail、成功機率、輸入時機與持久性。

本輪也在 `ida-pro-9.4-ver3:latest` 內以原始 DOS overlay-07 的暫存副本重跑既有
非破壞性稽核。輸入 `overlay-07.bin` SHA-256 為
`5483c71f98c5dc668d7d307c18a6b071dcfcba9d62eccb657600e7265125`；overlay-local
`1B3Fh` 精確保留四個方向的 `DS:720Fh/7210h` wrap/update，並呼叫
`017F:003Eh`、將 `AL` 寫到 `DS:7213h`，再呼叫 `017F:0034h`、將 `AL` 寫到
`DS:7212h`。這補強「movement result 由 cell-layer accessor 回傳」的靜態邊界，
但沒有找到 `wall=09` 的 writer 或 Search consumer；仍不能以 helper 名稱替代
runtime trace。工具版本／位址基準與第 520 輪一致，原始 overlay 與 baseline database
均未修改。

E2 端點則分成兩件事：

1. `(8,15,S)` 的 wall／detail 與普通 movement 阻擋是 `exact` GEO record。
2. ECL2 block 3 的 `CALL 0xC01E`、`Y=0`、`X-=2`、`NEWECL 4`，以及 block 4 的
   `LOAD FILES 4,2,FFh`／`LOAD PIECES 1,2,4` 是 `exact` raw branch／remake
   continuation；「玩家如何讓這個 boundary 嘗試發生」仍是 `unknown`。

## 對 remake 的影響

本輪只更新證據與記憶，沒有把未證實的猜測寫進遊戲：

- 暫不新增 `secret_door` JSON、`wall=09` 特判或把所有 `wall=09` 放行。
- 暫不把 `MOVEPARTY` 當成 E2；中文手冊支持的 `MOVEPARTY` 仍是跨遊戲角色／
  隊伍資料轉移候選。
- 保留目前 `SearchDungeonLocation()` 作為 ECL 主動搜尋一次的 service boundary；
  下一個最小實作應新增作品中立的 Search toggle／Look action 分離，再以同版
  runtime trace 決定 `wall=09` 的 discover／mutation contract。
- E2 成為正常玩家路徑前，仍需把「`(8,15)` boundary attempt → block 4」接到
  engine＋game-pack 的來源 map／座標／方向資料契約，並測試返回、重訪與 save-load。

因此 P0-2 從「是否為傳送門」縮小成兩個可驗證工作：`wall=09` 的 Search／第三平面
橋接，以及 `(8,15,S)` 的 E2 boundary input／external-exit transaction；本規格仍是
`DRAFT`，不宣稱下水道或整作完成。
