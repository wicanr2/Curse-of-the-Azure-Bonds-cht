# 第 439 輪：PC-98 睡眠術選定格與戰場 `TACTICALMAP`

狀態：`READY`（限 `SCAN` 呼叫參數、CoAB `TD/TDEF` 投影、手動 Sleep
玩家交易；Quick Sleep、原機 wall/corner 動態、效果解除、存檔及演出仍待續）

## 結論

1. `BackgroundTiles[1:66]` 與 PC-98 resident `TDEF` 65 筆、每筆四 byte
   逐筆一致；`MoveCost／Height／Field／TileIndex` 在 adapter 分別投影為
   Borland 正式欄位 `HT／LOS／SYM` 與未命名 `Raw3`。
2. `DungeonFloor.Tiles`／`WildernessFloor.Tiles` 保存的是一基底
   `TACTICALMAP.TD`，不是 renderer atlas 的 `TileIndex`。
3. Sleep 呼叫端在 `SCAN` 前依 Pascal 左至右壓入選定格 X、Y、
   `AOECOMBAT&07h=1`、`FFh`、常數 `1` 與 `COMBATMAP` far pointer。
   callee 證明 `FFh` 是 `INARC` arc、低三 bits 的 `1` 是 LOS range，
   `(A645h,A646h)` 是 source center；它不是 caster footprint。
4. remake 現以目前 floor bytes 建立 `TacticalMap`；手動 `Z` 睡眠術以選定格、
   range `1`、arc `FFh` 建立原版 object-ID 順序，再進已證實的
   `4d4`／HD／魔抗／effect writer，成功後才消耗法術格。

## 證據與推論等級

- `exact`：overlay 13 `2592h..25B4h` 的壓棧與 far call；overlay 31
  `08D8h..0B9Fh` 對選定座標、arc、range 的 consumer；65 筆 PC-98 TDEF
  raw bytes；floor byte 作為一基底 definition index 的既有 consumer。
- `material-exact/layout-reconstructed`：fallback placement 使用 32×16 local
  namespace，地城 `(0,0)` 對應原 floor `(18,7)`；TD/TDEF bytes exact，但
  local placement geometry 尚未由固定 PC-98 runtime 場面逐格驗證。
- `unknown`：原機 wall/corner/large-footprint 動態、Sleep twinkle、音效順序、
  受傷喚醒／自然解除、effect save round-trip、Quick AI 選格。

## 失敗即關閉邊界

- mode 沒有已恢復 floor、少於 65 筆 definitions、TD 為 `0`／`>65`、越界
  source 或不完整 legacy identity 時拒絕施法。
- `TACTICALMAP` 與 ordered targets 建立成功前不移除 memorized slot；Battle
  writer 失敗也歸還同一 spell ID。
- `BuildLegacyAreaScanTargetIDs` 明確接受 selected center；不得用 caster
  footprint helper 代替。

## 驗證

- frontend regression 證明 raw TD `22` 保持為 `22`，並讀
  `Definitions[21] == BackgroundTiles[22]`；未知 TD `66` 被拒絕。
- Battle regression 只接受靠近選定格的敵人，不接受靠近 caster 的敵人。
- State regression 驗證 selected-cell target 取得 `35h` effect；無效 map
  不消耗 slot。
- 正式 gate 使用 Go 1.24.13、Docker／Xvfb、`--network none`、唯讀公開模組
  快取與暫存 nested-engine replace，`./cmd/... ./gamepack ./internal/...` 共
  31 個套件行通過；`/tmp/coab-round439-formal.log` 結尾為
  `ROUND439_FORMAL_EXIT=0`。

## 未完成

1. PC-98／DOSBox 固定戰場 wall/corner/large-footprint 動態 ledger。
2. effect `35h` 每回合、受傷、dispel、戰鬥結束與存檔生命週期。
3. Sleep twinkle／聲音 oracle 與資料化 visual timeline。
4. Quick Sleep 的原版合法中心選擇與 delayed handoff。
