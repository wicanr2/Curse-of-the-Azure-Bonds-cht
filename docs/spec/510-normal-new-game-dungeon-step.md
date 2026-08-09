# 第五百一十輪：正常新遊戲地城移動交易

狀態：`READY`（有界正常路徑規格，不代表完整遊戲完成）  
日期：2026-08-09

## 目的

把新遊戲進入提爾佛頓後的第一個玩家移動，從「測試先改座標、再呼叫事件」提升
為前端與平台中立遊戲狀態共用的輸入交易。這是完整開場到結局驗收所需的基礎，
但只關閉第一格路徑，不把一格事件擴大成完整可通關聲明。

## 證據與位址空間

| 證據 | 內容 | 等級 |
|---|---|---|
| 原始 `curseoftheazurebonds.zip`／`GEO2.DAX` block 1 | 解碼後的 16×16 GEO 雙側 wall／door fields；起點 `(7,13)` 朝 west（direction `6`）可走 | `exact`（原始 bytes；GEO 位址空間） |
| ECL global block 1 runtime | 新遊戲完成角色建立與開場後，`C04B／C04C／C04D = 7／13／1`；`C04D` 是 0-based cardinal direction selector，映射為 direction `2` | `exact`（ECL work/register 位址空間） |
| `internal/game.State.MoveDungeon` | 以 `DungeonGeometryView` 取得 combined GEO 座標，檢查 cardinal delta、wrap、更新 ECL-facing view／wall／roof、清 `7F81h`、執行 lifecycle | `strong inference`（依現有 ECL register／前端順序；尚無 DOS 逐幀 capture） |
| `TestRealNewGameBeginsAtGlobalBlockOne` | 角色建立→開場 ECL→正常西行→Windlord’s Inn picture／HEAD-BODY／繁中訊息→Journal 31→回原格 | `exact`（remake 原始 bytes integration；不是 DOS pixel／instruction exact） |

注意：`GEO2.DAX` block ID、ECL register、ECL work address 與檔案 offset 是不同
位址空間。本文只以 register 名稱與 decoded map selector 並列，不把相同數字合併
成一個地址。

## 交易契約

`State.MoveDungeon(grid, dx, dy, direction)` 必須依序完成：

1. 驗證 direction 是 `0／2／4／6`，且 `(dx,dy)` 與 direction 一致。
2. 從 `DungeonGeometryView` 取得目前 geometry cell；使用 decoded GEO 的
   `CanMoveDungeonWrapped` 檢查牆與已解鎖門。
3. 計算下一格；先保留是否越過 16×16 邊界，再以原版 wrap 取得下一個 geometry
   cell，交給 `SetDungeonGeometryView` 投影回 ECL local map（若有 adapter）。
4. 由同一 decoded cell 更新 `DungeonWallType` 與 `DungeonWallRoof`，送出 step
   sound intent。
5. 若為 production `ModeDungeon`，普通格執行
   `RunDungeonLifecycle`；邊界嘗試執行 `RunDungeonExitLifecycle`。兩者都保留
   同一 `BlockSession`、pending ECL continuation 與 search-location 行為。

若門被鎖住，交易失敗；Ebiten 才能依目前 wall detail 開啟 pick／knock／bash
選單。門判定不應由故事座標或繁中顯示文字決定。

## 正常玩家路徑

目前已驗證的路徑是：

```text
ActionStart
  → character creation（加入一名角色）
  → FinishCharacterCreation
  → global ECL1 opening pauses
  → 進入 Tilverton GEO2 block 1 at (7,13,E)
  → MoveDungeon(-1,0,6)
  → (6,13,W), terrain selector 0x86
  → ECL event picture／HEAD2+BODY2、繁中旅店訊息
  → Journal 31 unlock／in-game journal read
  → close journal／return to same dungeon cell
```

此路徑中的唯一座標由原始 GEO／ECL evidence 建立；測試沒有以新值直接寫入
`DungeonX／DungeonY` 來假裝移動。後續測試若直接使用 `SetDungeonGeometryView`
只可標為 coordinate-assisted event test，不可標成 normal player path。

## 實作邊界

- 前端仍可維護 preview cursor，但 production movement 必須透過此 State 交易。
- renderer 只在交易成功後同步 cursor、重建 dungeon floor／wall preview；不得另跑
  lifecycle，也不得重複發出 movement sound。
- `7F81h` 是每一步的 transient event guard；它不是整張地圖永久完成旗標。
- 本輪沒有證明 DOS 原版每一 frame 的輸入 repeat、動畫 timing、聲效精確次序、
  邊界 exit 的全部 ECL 分支，這些保持 `strong inference` 或 `unknown`。
- 本輪沒有完成所有 dungeon／city／wilderness map，也沒有完成開場到結局、完整
  AD&D 規則、完整戰鬥、全翻譯、音樂音效與三平台發行。

## 驗證

Docker focused command：

```text
GOWORK=/tmp/go.work GOMODCACHE=/cache/mod GOCACHE=/cache/build GOPROXY=off \
go test -buildvcs=false -count=1 ./internal/game \
  -run TestRealNewGameBeginsAtGlobalBlockOne -v
```

正式 gate 仍須依 `AGENTS.md` 執行 `cmd/...`、`gamepack`、`internal/...`，並執行
`coab-audit`；本規格不能用 focused test 取代正式 gate。

