# 第 538 輪：火刀據點入口至首領的正常路徑

狀態：`READY`（本規格保留入口至首領證據；戰後世界／城市延伸由第 542 輪接續）

> 勘誤：本檔原先把首領勝利後的 `PATROL FOREST` 誤列為尚未閉合的出口。
> 第 542 輪證明它是前置提爾佛頓事件留下的合法共享工作旗標所觸發的原始分支，
> 並以同一正常 ECL session 完成 Patrol、世界旅行、阿沙本福德、立石群與艾森布拉
> 城外 handoff；本檔只保留本輪入口至首領的局部範圍。

## 結論

`TestRealNewGameBeginsAtGlobalBlockOne` 現在不只驗證 E2 入口。它從真實開場、提爾
佛頓下水道與 E2 handoff 後，留在同一個 ECL／地城 session，從火刀據點
`(8,1,S)` 逐格移動到 `(3,13)` 的首領事件。路徑沒有直接設定座標、直接呼叫
首領 handler 或注入戰鬥；必要的房間事件以正式選單／戰鬥／續跑交易完成。

這關閉的是「入口到首領戰前」的正常重製路徑。首領勝利後的同一 session、共享旗標
分支與世界／城市延伸不在本檔重複宣稱，請見
[`第 542 輪正常主線與城市／地城 handoff`](./542-normal-campaign-spine-and-city-dungeon-handoff.md)。

## 原始資料與位址邊界

- DOS 遊戲 archive：`curseoftheazurebonds.zip`
- archive SHA-256：
  `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d`
- GEO：`GEO2.DAX` block 4；起點 `(8,1,S)`，目標 `(3,13)`、terrain `0x87`
  （起點是下水道 `(10,15)` 踏出南緣後腳本寫的 `X := 10 − 2`，見
  [spec 1199](./1199-tilverton-sewers-map-edge-handoffs.md)）
- 原始 GEO／ECL 是路徑與事件的來源；地圖座標不是把攻略文字硬編碼進 Go 的替代品。
- 本輪測試在 Docker 內執行；測試所需的 `SEARCH` 只在進入據點後以正常 state
  交易關閉，以避免與本路徑無關的隨機遭遇干擾，不改寫地圖座標或事件旗標。

## 已驗證的路徑

逐格座標如下。第一個是起點與它的**朝向**；之後每一個是**走到那一格用的方向**
（`(7,1) W` ＝ 從上一格往西走到 `(7,1)`）：

```text
(8,1) S
→ (7,1) W → (7,2) S → (6,2) W → (5,2) W → (4,2) W → (4,3) S → (3,3) W
→ (3,4) S → (4,4) E → (4,5) S → (5,5) E → (5,6) S
→ (5,7) S → (6,7) E → (7,7) E → (8,7) E → (9,7) E
→ (9,8) S → (10,8) E → (10,9) S → (10,10) S → (10,11) S
→ (9,11) W → (9,12) S → (8,12) W → (7,12) W → (6,12) W
→ (5,12) W → (4,12) W → (4,13) S → (3,13) W
```

此路徑穿過原始資料中的必要特殊區域：

- `(5,2)`：terrain `0x99` 刀刃區，測試以資料驅動選項選擇等待／避開分支。
- `(4,2)`、`(3,3)`：terrain `0x9A` 冰凍房，測試完成正式事件續跑。
- `(5,7)` 與 `(6,7)..(9,7)`：terrain `0x94／0x95` 相位蜘蛛區，遭遇以正式
  combat continuation 完成。
- `(3,13)`：terrain `0x87`，抵達後先顯示 game-pack message ID
  `journal-trigger.fire-knives-leader-11`，再進入首領戰前狀態。

測試以 enemy side／stable combat object 資料確認首領戰包含 20 名火刀戰士加首領，
共 21 名敵人；沒有用翻譯後的顯示字串當測試真相來源。

## 尚未由本規格宣稱的部分

1. 路線上尚未涵蓋的可選房間、所有寶物、重訪旗標與失敗分支仍在 P0。
2. 固定首領戰 fixture 仍只代表固定狀態下的夢境、Tilverton 世界地圖與 save/load；
   不可用它取代第 542 輪的正常 session 證據，也不能由此宣稱所有火刀可選房間與
   重訪旗標完成。
3. 原版 DOS／PC-98 的完整 combat animation、音效次序、palette cycle 與逐像素
   frame 仍須另以 DOSBox／影片／原始資產驗證。

## 驗證

在 Docker image `coab-go-test:20260729` 內執行：

```text
go test -modfile=workplace/coab-test.mod ./internal/game -run '^(TestRealNewGameBeginsAtGlobalBlockOne|TestRealFireKnife(BladeBarrierBranches|FrozenRoomBranches|OfficeStages|AshenRooms|LeaderEncounterAndBondProgression)|TestFireKnifeLeaderStateVictoryReturnsToTilverton)$' -count=1 -timeout=240s
```

本輪結果：`ok`。這是代表性正常路徑與固定 fixture 的抽樣，不是全套通關 gate。
