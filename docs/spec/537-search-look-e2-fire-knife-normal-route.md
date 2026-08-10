# 第五百三十七輪：SEARCH／LOOK、下水道 E2 與火刀 E1 正常路徑

狀態：`READY`（remake 路徑與資料契約；原版 wall writer 仍非 exact）
日期：2026-08-10

## 結論

本輪完成的是一個可重播的 engine＋JSON＋State vertical slice：

```text
開啟 persistent SEARCH
  → (13,10) 沿 GEO2 block 3 正常移動
  → 搜尋發現 wall=09 的候選邊
  → 抵達 (8,15,S) 並以 external-exit transaction 進入 ECL2 block 4
  → 火刀據點入口 (6,1,S)
  → 處理據點前的 ECL 單鍵事件／刀刃等待
  → 抵達北側 E1 候選 (8,0,N)，再次北行越界
  → ECL 回到下水道 (10,15,N)
```

同一輪也驗證火刀首領勝利後的 ECL 夢境、Tilverton 世界地圖邊界，以及
`remake save v12` 的存檔／載入重訪。這些是**已完成的 remake 行為**；不能把
它們擴大解讀成「全 ECL、全地圖或整作已通關」。

## 證據與分級

| 輸入／結果 | 等級 | 說明 |
|---|---|---|
| `curseoftheazurebonds.zip` | `exact` 輸入 | SHA-256 `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` |
| `GEO2.DAX` block 3／4 的四平面 record | `exact` | 原始 archive 解碼；wall／detail／terrain 的位址空間分開保存 |
| `wall=09/detail=0` 是唯一橋接候選 | `strong inference` | GEO graph、手冊 SEARCH 語意與攻略圖例一致；沒有同版 writer→consumer trace |
| `(8,15,S)` E2 → `NEWECL 4`、`LOAD PIECES 1,2,4` | `exact` raw branch／remake trace | ECL branch 與正常 State 交易均有回歸 |
| 火刀北側 `(8,0,N)`、`(11,0,N)`、`(13,0,N)` | `strong inference` | GEO boundary 候選與攻略「E1 全部可用」交叉支持；仍非原版逐指令座標證明 |
| E1 runtime 回到 GEO2 block 3 `(10,15,N)` | `exact` remake／ECL trace | 由實際 block 4 移動與 boundary lifecycle 產生，不是直接設定座標 |
| 首領勝利 → 世界地圖 → save/load | `exact` remake integration | 同一 ECL session、世界選單與 save v12 狀態回歸 |

公開攻略只作地點／玩家可見語意交叉驗證：[GameFAQs Tilverton／Fire Knife 攻略](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)、[GameBanshee Fire Knife Hideout](https://www.gamebanshee.com/curseoftheazurebonds/walkthrough/firekniveshideout.php)。手冊文字以本機 `Curse-of-the-Azure-Bonds_Manual_DOS_EN.pdf` 為主，公開手冊頁只作交叉參考。

## 引擎與 game-pack 契約

獨立 engine 新增：

- `Pack.Search`／`SearchDefinition`：保存 SEARCH 開／關與 LOOK policy；
- `MapDefinition.SearchEdges`：可發現邊與 `discovery`／`confidence`；
- `MapDefinition.ExternalExits`：座標／方向／作品事件 ID 的邊界交易；
- `FindSearchEdge` 與 `FindExternalExit`：只做 map namespace 的 exact lookup；
- JSON schema、ID／方向／推論等級驗證，以及 engine focused tests。

CoAB `pit-of-moander.json` 宣告：

- `tilverton.sewers.wall-09-west`：`(10,12,W)`、`wall_type=09`、
  `search_or_look`、`strong inference`；
- `tilverton.sewers.e2-south-boundary`：`(8,15,S)`、作品
  `ecl-boundary`、`strong inference`；
- `tilverton.fire-knife-hideout.e1-north-west／centre／east`：北側三個 E1
  候選，均為 `strong inference`。

State 的 `S` 現在只切換 persistent `DungeonSearchEnabled`，不立即跑 ECL；`L`
呼叫一次性 `LookDungeonLocation`。每次正常移動仍在同一 `BlockSession` 執行
per-turn／Search continuation；發現的 edge ID 隨 save v12 保存，載入時重新驗證
game-pack。舊 v1–v11 存檔沒有這些欄位，仍可載入並以關閉 SEARCH／空 edge 起步。

`0x7ECA=1` 只在目前 State 的 Search／Look service boundary 與 per-turn entry 1
期間暫存；它不是跨整張地圖的永久旗標。`7F81h` 仍是每步事件 guard，不能因
搜尋或轉向而錯誤清除／保留。

## 正常玩家驗證

Docker image：`coab-go-test:20260729`；測試使用 `GOCACHE`／`GOPATH` 暫存目錄，
不在主機執行 Go／遊戲負載。

```text
go test -modfile=workplace/coab-test.mod ./internal/game \
  -run '^(TestRealNewGameBeginsAtGlobalBlockOne|TestFireKnifeLeaderStateVictoryReturnsToTilverton)$'
go test -modfile=workplace/coab-test.mod ./internal/game \
  -run '^(TestPartySaveLoadRoundTripRestoresDungeonSearchState|TestDungeonSearchToggleDoesNotConsumeTurn)$'
```

第一條測試由角色建立與開場開始，經公會、騎士、wall=09、E2、火刀 E1 與下水道
回返；不是 direct-entry 座標注入。第二條確認 SEARCH 開關與 discovered edge ID
可存檔重訪。火刀首領測試的戰鬥本身是 deterministic fixture，後續 ECL／夢境／
世界地圖與存檔重訪均仍以同一 session 執行；它不冒稱已從火刀所有房間逐格走到
首領。

## 尚未完成

- 原版 DOS／PC-98 的 wall=09 writer、成功率、detail 持久性與逐幀輸入仍待
  runtime evidence；目前 remake 以 `strong inference` 候選資料前進。
- 火刀據點從入口逐房間到首領的完整正常路徑、所有 E1／戰後封閉條件與重訪旗標
  尚未全數閉合。
- 全 ECL／全地圖、完整 AD&D 戰鬥 AI／法術效果／動畫、音樂音效與完整中文化仍是
  後續 P0／P1 工作；本文件只關閉上述 vertical slice。
