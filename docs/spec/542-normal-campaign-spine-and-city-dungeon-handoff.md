# 第 542 輪：正常主線與城市／地城 handoff

狀態：`READY`（正常 session 主線骨架；不是完整結局 gate）

## 結論

`TestRealNewGameBeginsAtGlobalBlockOne` 現在以同一個新遊戲 ECL session 完成下列
連續玩家路徑：

```text
開場／角色建立
  → 提爾佛頓下水道／E2
  → 火刀據點逐格房間路徑／首領戰
  → 首領夢境與 Tilverton 世界選單
  → 正常 PATROL FOREST 戰鬥續跑
  → JOURNEY ON → 阿沙本福德城外
  → 進城 → 河畔酒館 → Tavern Tale 28 → 離城
  → 立石群 → 提爾隘口戰鬥 → 灰袍男子／尋紅線索
  → JOURNEY ON → 艾森布拉城外
```

這是本作目前最長的「新隊伍不注入座標」主線骨架。它確認先前被誤判為 bug 的
`PATROL FOREST` 是正常狀態：前置提爾佛頓高階祭司事件已留下 `4C03=0x80` 共享
事件群組旗標；固定只注入火刀首領的夾具沒有該前置狀態，所以會得到不同的
`ENTER CITY` 分支。不能為了讓兩個夾具畫面相同而清除 `4C03`。

## 證據與邊界

- 原始輸入：`curseoftheazurebonds.zip`，SHA-256
  `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d`。
- ECL：同一 `BlockSession` 從開場、ECL2 block 4、ECL1 block 0x50／0x51 續跑；
  戰鬥勝利、圖片、選單與世界抵達都沒有重建 block 起點。
- 資料：新加的選項輸入使用 game-pack `option_rule.id`，例如
  `ecl-option.enter-city`、`ecl-option.journey-on`、`ecl-option.bar`、
  `ecl-option.relax`、`ecl-option.the-standing-stone` 與
  `ecl-option.essembra`；沒有把繁中顯示文字複製到測試。
- 原版／重製證據等級：ECL branch、memory handoff、stable combat object 與
  玩家輸入續跑為 `exact`（以本機原始 ECL 與 remake runtime 閉合）；畫面配置仍是
  `layout-reconstructed`，不是 DOS／PC-98 逐像素相同。
- 第 541 輪的 14 個世界點位 arrival 與 JSON directed graph 仍是獨立 baseline，
  本輪只增加其中一條正常主線的連續使用證據。

## 城市／地城覆蓋的正確解讀

目前的事件主體由原始 ECL 執行，CoAB JSON 提供 locale、stable option、地圖／
人物／戰鬥資料與 typed adapter；不能因測試檔能讀到一段事件，就把同一段劇情
硬編碼回 Go。現有固定狀態整合測試另外覆蓋：

- 阿沙本福德、艾森布拉、哈普、熔岩洞與法師塔的城市／地城事件與續跑。
- 希爾斯法、尤拉什、摩安德之坑與散提爾堡的主要入口、戰鬥、同伴／手札與離場。
- `TestRealPlayerPathStandingStoneToBurialGlen` 的立石群→Myth Drannor→Burial Glen
  正常地圖路徑，以及紅網、蜘蛛／羅剎妖與後續持久旗標。

這些測試合在一起是「已有大量可重播垂直切片」，不是單一新隊伍從開場走到結局。
仍不能宣稱完成的項目：

1. 每一座城市所有設施的所有分支、每張 GEO 的所有可選房間、隨機遭遇與重訪旗標
   尚未形成全量 coverage matrix。
2. 哈普→熔岩洞→法師塔→散提爾堡→摩安德之坑→Myth Drannor 的完整新遊戲同一
   session 串接尚未成為正式 P0 gate；現有長測試部分採固定狀態或局部
   coordinate-assisted setup，必須保留此證據限制。
3. 最終提朗瑟克斯終戰、`PROGRAM 8` 結局、完整存檔／重載與失敗分支仍須同一
   正常主線驗收；完整戰鬥 AI、法術／弓箭演出、音效與 UI fidelity 也不因本輪縮減。

## 驗證

Docker image `coab-go-test:20260729` 內執行：

```text
go test -count=1 -modfile=workplace/coab-test.mod ./internal/game \
  -run '^TestRealNewGameBeginsAtGlobalBlockOne$' -timeout=240s
```

結果：`PASS`。另已在同一容器執行火刀房間／首領固定回歸：

```text
TestRealFireKnifeBladeBarrierBranches
TestRealFireKnifeFrozenRoomBranches
TestRealFireKnifeOfficeStages
TestRealFireKnifeAshenRooms
TestRealFireKnifeLeaderEncounterAndBondProgression
TestFireKnifeLeaderStateVictoryReturnsToTilverton
```

這輪可把 P0 的「首領後正常世界出口、阿沙本福德與立石群主線 handoff」從待辦
移除；不能把剩餘全城市、全房間、完整結局或所有戰鬥功能移除。
