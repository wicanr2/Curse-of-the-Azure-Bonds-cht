# 第三百四十五輪：原版 AREA 8X8D symbol map

狀態：`READY`

## 反組譯證據

公開 CoAB reference commit `9dc46f1`：

- `engine/seg001.cs` 在 `game_area=1` 時呼叫 `Load8x8D(4, 0xCA)`；
- `engine/ovr031.cs::DrawAreaMap` 畫 11×11，offset 為
  `clamp(partyCoordinate-5, 0, 5)`；
- N／E／S／W wall presence 分別是 mask bit `1/2/4/8`；
- wall global symbol 是 `0x104+mask`，party 是
  `0x100+(partyDir>>1)`；
- door overlay 只存在於 `Cheats.improved_area_map`，不屬原版 AREA。

原始 `8X8D1.DAX/CA` regression 證實 picture 為 8×8、40 items，足以覆蓋
party local items `0..3` 與 wall local items `4..19`。

## 實作

- 共用 engine `areamap.BuildOriginal` 產生 clamped 11×11 tile selection，
  engine commit `443281a`。
- schema 新增 `symbol_block`；CoAB JSON 指定 `8X8D1.DAX/CA`。
- renderer 直接畫原始 EGA indexed symbols，2× nearest-neighbour；不再畫
  16×16 向量格、門線或自製隊伍方塊。
- 中文 HUD 沿用本機倚天 16×15 embolden face。

## 驗證

- engine `go test ./...`。
- CoAB game-pack、original-image graphics 與 ETen focused tests。
- Docker build 與 Xvfb 640×480 實機圖
  `docs/screenshots/coab-area-map-remake.png`。
