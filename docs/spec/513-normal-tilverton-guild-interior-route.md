# 第 513 輪：提爾佛頓盜賊公會內部正常移動路徑

狀態：`READY`（有界正常玩家路徑；不是完整通關）
日期：2026-08-09

## 目的與範圍

第 512 輪已由提爾佛頓城門正常移動抵達皇家馬車、監牢、盜賊救援與公會
`GEO2` block 2。本輪把公會入口後、直到下水道痕跡事件前的 coordinate-assisted
段落改成同一個 `State.MoveDungeon`／`TurnDungeonWithGrid` 交易：玩家實際逐格
走路，ECL search／per-turn 事件在移動交易內發生，戰鬥與提示事件都保留同一個
session continuation。

本輪的終點是公會下水道門內的 `tilverton.guild-sewer-traces`。選擇繼續後，
block 3 的下水道入口仍可驗證；入口後的 `(1,8)` 檢查點與更深處座標仍是
coordinate-assisted，故本規格不宣稱公會下水道已完成。

## 證據與位址空間

| 證據 | 內容 | 等級 |
|---|---|---|
| 原始 archive | `curseoftheazurebonds.zip`，SHA-256 `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` | `exact`（輸入識別） |
| GEO | archive 內 `GEO2.DAX` block 1；由既有 `internal/dax` 與 engine `geometry` 解碼，16×16 geometry cell、雙側 wall／door detail | `exact`（資料解碼） |
| remake 路徑 | `TestRealNewGameBeginsAtGlobalBlockOne` 從公會 block 2 handoff 以 `MoveDungeon`、`TurnDungeonWithGrid` 逐格走到 sewer traces | `exact`（remake regression） |
| DOS movement 對應 | `MoveDungeon` 的交易順序與 DOS movement loop 的跨平台對應 | `strong inference`；尚無本段逐幀 DOSBox input/timing 對照 |
| 原版行為語意 | 門旗標、ECL boundary、提示文字、戰鬥組成與 continuation 由 raw ECL／GEO、既有 runtime trace 及 game-pack regression 共同支持 | 各項詳列於下；未閉合者不得升級 |

本規格的 geometry 座標是 decoded GEO 空間；ECL `C04B／C04C／C04D`、GEO
`(x,y,direction)`、file offset、ECL work address 與 combat record offset 不可
互相代稱。`GEO2.DAX` 的 block／cell 不是 Borland symbol 的 segment:offset。

## 正常路徑與門檻

公會混合戰勝利後，runtime 回到 GEO `(9,3,E)` 的等價視圖；測試先轉向南方，
再沿下列路徑抵達半身人、犬舍、猴籠與公會內部下水道門：

1. `(9,3)` → `(10,3)` → `(11,3)` → `(11,4)` → `(11,5)` → `(11,6)` → `(11,7)`。
   `(11,3)` 的訪客簿與半身人事件透過正常 search boundary 出現。
2. 在 `(11,7)` 的東側 detail `2` 鎖門前，測試使用可重現的高力量 fixture
   呼叫 `BashDungeonDoor`，再由 `UnlockDoorWrapped` 更新雙側 GEO detail；這只是
   測試用確定性輸入，不是把「力量 25」寫成原作規則。走入 `(12,7)` 觸發犬舍
   介紹與實戰戰鬥。
3. 犬舍勝利後由 `(12,7,S)` 轉向北方，開啟 `(12,7,N)` detail `2` 鎖門，沿
   `(12,6)`、`(12,5)`、`(12,4)`，轉東走至 `(15,4)`，轉南經 `(15,5)`、
   `(15,6)`、`(15,7)`，觸發猴籠事件。
4. 猴籠返回 `(15,7,S)` 後，開啟 `(15,7,N)` detail `2` 鎖門，沿北、東／西
   回廊與南側大廳繞行至 `(10,13)`。原始 GEO 在 `(13,7)` 東側是實心牆，
   不是漏掉的門；不能用 `(13,7)` → `(14,7)` 的直線路徑取代繞行。
5. 在 `(10,13,S)` 開啟第四道 detail `2` 鎖門，正常走過下水道門前與門內的
   search boundary，最後抵達 `tilverton.guild-sewer-traces`。

實際測試序列使用的移動交易如下，方向數字是 GEO cardinal direction：

```text
(9,3) E×2 → S×4 → halfling
(11,7) 開 E 門 → kennel
(12,7) 轉 N、開 N 門 → N×3、轉 E、E×3、轉 S、S×3 → monkey cages
(15,7) 轉 N、開 N 門 → N×3、轉 W、W×3、轉 S、S×3、轉 W、W×1
  → 轉 S、S×2、轉 W、W×1、轉 N、N×3、W×2、轉 S、S×6
  → 轉 E、E×2、轉 S、開 (10,13) 南門、S×3 → sewer traces
```

程式內保留了每個實際 cell 的明確步驟與 route context；上面的 grouped notation
只作規格閱讀，不能成為另一份繞過 GEO 的 teleport API。

## 事件、資料與 continuation

本輪新增或正式驗收的 CoAB game-pack stable IDs：

- `tilverton.guild-guest-book`
- `tilverton.running-thieves`
- `tilverton.option.remain-calm`
- `tilverton.running-thieves-warning`
- `tilverton.fire-knives-spot-you`
- `tilverton.guild-assassins-attack`
- `tilverton.guild-metal-and-animals`
- `tilverton.guild-bodies-after-battle`

公會路徑會驗證訪客簿、奔逃盜賊的 `REMAIN CALM` 分支及其 PRESS、火刀發現、
刺客突襲、金屬與野獸聲響、戰鬥後屍體等 boundary。選項索引由正式 game-pack
`option_rules` 的 stable ID 解析，不在測試複製當前繁中或英文顯示文字。所有
新增事件都同時保存英文 source matcher 與 `en`／`zh-TW` locale。

犬舍與路途中隨機遭遇使用真正 `CombatAct` 回合推進；沒有把敵人 HP 設為零來
假造勝利。公會戰後的戰鬥 return mode 由 dungeon caller 保存；若生命週期中間
經過 engine-only `CALL` 才建立戰鬥，仍須回到 `ModeDungeon`，不能停在只顯示
「戰鬥勝利」的 `ModeEvent`。`dungeonLifecycleActive` 是 remake 生命週期邊界的
typed context，不是作品劇情旗標。

## 實作契約

- 正常探索輸入只能經 `State.MoveDungeon`；它負責 cardinal delta、GEO 雙側
  wall／door legality、座標投影、`7F81h` 單步 guard、per-turn 與 search。
- `TurnDungeonWithGrid` 只轉向並更新 wall／roof projection，不自行觸發 ECL。
- 開門仍由 typed door action 結果與 GEO adapter 更新雙側 detail；不能在 State
  中為公會座標加劇情特例。
- 文字、選項與提示只由 CoAB game-pack 提供；測試期待值使用 stable ID／resolver。
- 戰鬥後若 `eventReturnMode == ModeDungeon`，必須明確呼叫既有 continuation，
  不可用重新設定座標或重新從 block 起點執行來掩蓋 pending PC 遺失。

## 驗證

本輪 focused regression：

```text
go test -buildvcs=false -count=1 ./gamepack ./internal/game \
  -run 'TestRealNewGameBeginsAtGlobalBlockOne|TestTilvertonGuildAndHideoutTransitionIsGamePackDriven'
```

提交前還要在 `coab-go-test:20260729`、手動 Xvfb 與 pinned engine workspace 中執行
正式 package gate、`coab-audit` 與 `git diff --check`；結果寫入本輪交接記錄。

## 未完成邊界

- `RunDungeonExitLifecycle` 進入 block 3 後，測試仍直接設定 `(1,8)` 檢查點，
  後續下水道騎士、忠誠選項、火刀據點與返回世界地圖也仍有 coordinate-assisted
  段落；這是下一個正常路徑 milestone。
- 完整 ECL／external routine、所有 GEO／世界旅行、完整 AD&D 戰鬥 AI、法術／
  遠程／死亡動畫、音樂音效、UI 逐幀 parity、全翻譯與全作結局仍未完成。
- 本輪沒有新增 README 截圖；既有圖片不能用來宣稱本段 layout 或 combat
  pixel-exact。原版石框、人物 HEAD／BODY、640×480 中文排版仍依既有視覺 spec
  驗收。
