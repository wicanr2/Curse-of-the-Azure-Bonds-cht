# 451：法師塔四分支與屋頂出口 JSON 化

狀態：`READY`

## 範圍與證據

本輪接續 READY spec 316–319，不改寫已由 raw ECL 與正常玩家路徑證明的
分支語意：

- `ATTACK WIZARD`：龍群離去、德拉坎德羅斯召來四名守軍、勝利後安全休息；
- `PARLAY WITH THE DRAGONS`：成功態度說服龍群後仍進守軍戰，敵對態度進
  14 黑龍戰；
- `ATTACK DRAGONS／FLEE`：共同進入 14 黑龍戰，依 `4C61` 決定龍心選單，
  YES 後承受原 ECL 全隊酸液傷害；
- 屋頂 terrain `01h`：洞穴／荒野／留下，以及荒野下層村莊／離開選單。

上述規則、怪物、旗標、傷害與 continuation 仍由原 ECL 驅動。本輪只把
作品文字與三個作品專用選項移出 Go State。

## Game-pack 唯一真相來源

CoAB JSON 新增十個事件 message IDs／text rules：龍群離去、召兵、安全屋頂、
交涉成功、龍群定罪、龍屍、龍心詢問、酸液取心、荒野出口與屋頂出口；另以
三個 `option_rules` 資料化 `ATTACK DRAGONS／ATTACK WIZARD／PARLAY WITH
THE DRAGONS`。

英文與繁中 locale 現各有 265 個完全相同的 stable IDs。State 的十個
句子 match cases、三個 option switch cases，以及舊 locale 的十三筆中文複本
都已刪除。`FLEE`、YES／NO、CAVES／WILDERNESS 等共用詞仍走既有通用選項
契約；本輪沒有把作品資料塞進共用 engine。

## 驗證

- `TestWizardTowerDracandrosStoryAndJournalAreGamePackDriven` 現涵蓋整段
  20 個法師塔 text-rule transactions、手札 15 兩頁、三個專用選項及
  en／zh-TW key parity。
- `TestFireKnifeLeaderStateVictoryReturnsToTilverton` 的正常 GEO5／ECL 長度
  有界 vertical slice 抽樣 ATTACK WIZARD、成功／敵對 PARLAY、14 黑龍、
  龍心與屋頂出口；不是從全遊戲開場 marathon。
- 正式 Docker gate 與 marker 記錄於 `CONTEXT.md`。

## 尚未完成

- 法師塔前一區的熔岩洞、離區後阿卡巴／黑暗精靈／龍巫妖文字仍有歷史
  State fallback，應依各自 READY spec 分批遷移。
- 全專案仍需依章節建立「Go 作品文字零殘留」稽核，不能只靠搜尋本輪 IDs
  便宣稱 engine／data 分離全部完成。
