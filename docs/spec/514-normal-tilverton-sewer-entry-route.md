# 第 514 輪：提爾佛頓下水道入口至火刀檢查站

狀態：`READY`（有界正常玩家路徑；不是完整下水道通關）
日期：2026-08-09

## 目的與終點

第 513 輪已把公會內部正常走路接到 block 3 的下水道入口；本輪移除入口後第一
個座標輔助段落。玩家從 `tilverton.sewers-entry` 的正常回返位置 `(0,1,S)`，
沿 `GEO2.DAX` block 3 逐格走到火刀檢查站 `(1,8,S)`。途中第一次到達檢查站
會先遇到「仍聽見公會大廳戰鬥聲」的 per-turn PRESS；按下繼續後，玩家再使用
正式 `SEARCH` 交易，才觸發火刀要求投降的事件。

本輪終點是 `tilverton.sewers-checkpoint` 的雙選項事件，以及拒絕投降後的五名
火刀實戰與 `tilverton.sewers-hide-bodies` 戰後 continuation。第 515 輪已另以
資料驅動 map-position contract 關閉第一個 `(13,10)` handoff；本規格仍只記錄
第 514 輪的入口／檢查站邊界，不把後續騎士、block 4 或完整下水道包裝成完成。

## 證據與位址空間

| 證據 | 內容 | 等級 |
|---|---|---|
| 原始 archive | `curseoftheazurebonds.zip`，SHA-256 `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` | `exact`（輸入識別） |
| GEO | archive 內 `GEO2.DAX` block 3；由既有 DAX parser 與 engine `geometry` 解碼，保留 16×16 cell、wall／detail／terrain | `exact`（資料解碼） |
| 正常入口 | `State.MoveDungeon` 從 `(0,1)` 實際走到 `(1,8)`，不寫入目標座標；每步由雙側 GEO movement contract 驗證 | `exact`（remake regression） |
| PRESS／搜尋順序 | 最後一步的 per-turn 結果有 meaningful text，因此 lifecycle 暫停 search；按鍵續跑後由 `SearchDungeonLocation` 觸發同一 cell 的檢查站 | `exact`（remake lifecycle regression）；對 DOS 互動節奏仍是 `strong inference` |
| 戰鬥 | 拒絕投降後實際建立 5 名 Fire Knife，使用 `CombatAct` 完成勝利 | `exact`（remake regression）；特殊能力與逐幀效果仍未閉合 |

此處 `(x,y,direction)` 是 `GEO2.DAX` block 3 的 geometry／ECL 同位座標；與
file offset、ECL work address、Borland `segment:offset` 及 combat object index
分開保存。原始 archive hash 與工具／版本記錄沿既有專案規則保存，不以截圖取代
bytes 或 runtime trace。

## 正常移動序列

入口回返為 `(0,1,S)`。可行路徑由 decoded GEO 的雙側 wall／detail 合法性得到：

```text
(0,1) → (1,1) → (1,2) → (1,3) → (1,4)
      → (0,4) → (0,5) → (0,6) → (1,6) → (1,7) → (1,8)
```

對應 `MoveDungeon` 交易為：

```text
E、S×3、W、S×2、E、S×2
```

中途 `(1,6)` 的 terrain 是 `0x8D`，終點 `(1,8)` 的 terrain 是 `0x81`。本輪不把
terrain 數字直接命名成劇情旗標；它們只作 GEO cell 識別，事件語意仍由 ECL
runtime／game-pack matcher 證明。

## ECL 邊界與資料分層

正常移動最後一步先呈現原始文字：

```text
YOU STILL HEAR THE OCCASIONAL SOUNDS OF BATTLE ECHOING FROM THE GUILD HALL.
```

這筆訊息已移入 CoAB game-pack：
`tilverton.sewers.guild-battle-echoes`，並提供 `en`／`zh-TW` locale。它是
一頁 PRESS，不是火刀檢查站選單；測試先由 stable option ID
`ecl-option.press-button-or-return-to-continue` 按下繼續。因為本次 per-turn
結果含有 meaningful text，`runDungeonLifecycle` 不會在同一呼叫中假設 search
也已完成；返回 `ModeDungeon` 後，玩家發出正式 `SEARCH`，再由
`SearchDungeonLocation` 執行 terrain `0x81` 的檢查站 boundary。

檢查站拒絕投降分支保留原始兩階段 continuation：

1. `tilverton.sewers-checkpoint` 顯示投降／拒絕選項。
2. 選拒絕後建立五名 `FIRE KNIFE`，以實際 `CombatAct` 回合完成戰鬥。
3. 勝利後顯示 `tilverton.sewers-hide-bodies`，按下繼續才返回同一個地城 session。

產品測試期待值一律由 game-pack stable ID／locale resolver 取得，不複製目前繁中
文字；`gamepack/gamepack_test.go` 同時驗證英文 matcher 與 `en`／`zh-TW` 解析。

## 實作契約

- 正常走路只能經 `State.MoveDungeon(sewerGrid, ...)`；不能在下水道入口後直接
  設定 `(1,8)` 再呼叫 lifecycle。
- `TurnDungeonWithGrid` 不是本段必要的 teleport 或事件觸發器；需要主動搜尋時，
  使用作品中立的 `State.SearchDungeonLocation()`，讓 `7ECA` search contract
  與 ECL session continuation 可追蹤。
- per-turn PRESS 與 search event 是兩個 boundary，不能把它們合併成一個訊息，
  也不能在按鍵後自動重跑 block 起點來掩蓋 pending PC。
- `dungeonLifecycleActive`／return mode 只保存 remake caller context；不要把
  `guild-battle-echoes` 或 `(1,8)` 寫進共用 engine 劇情規則。

## 驗證

本輪 focused regression：

```text
go test -buildvcs=false -count=1 ./gamepack ./internal/game \
  -run 'TestRealNewGameBeginsAtGlobalBlockOne|TestTilvertonGuildAndHideoutTransitionIsGamePackDriven'
```

提交前仍需依 AGENTS.md 的週期規則執行 Docker／Xvfb 正式 package gate、
`coab-audit` 與 `git diff --check`；本輪若只有 locale／正常路徑與既有戰鬥接線變更，
不得把 focused 結果擴大寫成完整 combat gate。

## 後續勘誤與未完成邊界

- 第 514 輪原本直接把 block 3 位置設為 `(13,10)` 的測試輔助，已由第 515 輪
  `set_map_position` event 取代；其 predicate、raw ECL `CALL 2E10h`、PC-98
  `MOVEFORWARD` 對照與 confidence 見
  [`docs/spec/515-fire-knife-map-position-transition.md`](515-fire-knife-map-position-transition.md)。
  這是 `strong inference` 的外部 handoff，不是 DOS writer→consumer 的
  `exact` 完成；仍不得用 BFS 穿牆。
- 騎士後的選項、下水道南界、block 4 火刀據點入口與返回世界地圖仍未改成正常
  `MoveDungeon`／exit transaction。
- 完整 ECL、戰鬥特殊能力、遠程／法術／死亡演出、音效、UI pixel parity、全翻譯
  與整作結局仍未完成。
