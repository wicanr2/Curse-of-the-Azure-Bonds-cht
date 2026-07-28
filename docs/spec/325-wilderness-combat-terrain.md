# 第三百二十五輪：野外戰鬥 terrain adapter

狀態：`READY`

## 證據與選擇

- `SetupWildernessFloor` 已還原為 50×25 `WildernessFloor`。
- wilderness segment 的 `BackgroundTile.TileIndex` 範圍是 `0..33`，
  正好對應 WILDCOM 的 34 張 24×24 tile。
- dungeon segment 由既有 `DungeonFloor` 查 DUNGCOM。
- RANDCOM 只有六張桌椅／障礙物，沒有完整 floor，因此是 decoration overlay，
  不得當成 7×7 地板。

## Renderer contract

- 正式自動選擇只看 `Area.InDungeon`：true 選 DUNGCOM，false 選 WILDCOM。
- 不再以 `GameArea > 1` 猜測地城；章節編號不是 terrain mode。
- WILDCOM camera 以 `MapX/MapY` 為中心取
  `(x-3..x+3, y-3..y+3)`，再以 2× nearest-neighbour 繪製。
- `-combat-terrain` 只作 visual verification override；使用 WILDCOM 的
  direct encounter 會建立 seed 1、city flags `0x20` 的可重現 floor。
- RANDCOM placement 尚未反組，不用 heuristic 亂放。

## 驗收

- selector unit test 鎖定 dungeon／wilderness，不接受 area heuristic。
- camera-center test 確認 WILDCOM 中央格就是 `WildernessFloor.Entry(MapX,MapY)`。
- `docs/screenshots/gold-box-layout-combat-wilderness.png` 顯示由真實 floor
  lookup 得到的樹、倒木、岩石、草與水岸。

