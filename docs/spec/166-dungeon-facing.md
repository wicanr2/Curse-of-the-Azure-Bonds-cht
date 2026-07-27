# 第一百六十六輪：dungeon facing rotation

狀態：`READY`（限 remake dungeon preview 的 facing state 與 wall traversal）

## 證據與實作

- reference `Draw3dWorld` 使用 `partyDir` 與固定八方向 delta；目前 shared traversal 的順序為 `N, NE, E, SE, S, SW, W, NW`。
- `game.State.TurnDungeon` 以同一八方向順序保存 facing，支援正／負 rotation 與 wrap。
- dungeon preview 的 `Q/E` 分別左右轉 90 度（direction delta `-2/+2`），每次轉向立即重建 Far／Mid／Near wall stamps；方向鍵仍是 GEO cardinal movement。
- version 3 remake save 已保存這個非零 facing；啟動／F9 載入後，`prepareWallPreview` 會用恢復的方向。

## 明確 boundary

本輪只證明 renderer preview 的 facing transaction 與保存邊界；尚未把 Area1 真實 `mapDirection` 寫回、ECL context wrap、轉向／移動時間、遭遇、sky／roof／door overlay 或完整 3D viewport 宣稱完成。

## 驗證

`TestTurnDungeonUsesEightDirectionOrder` 覆蓋正轉、負轉與 wrap；既有 wall traversal、save codec、game F5/F9 round-trip 與 Docker gate 繼續驗證完整資料鏈。
