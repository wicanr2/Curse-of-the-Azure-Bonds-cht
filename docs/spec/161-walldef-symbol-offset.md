# 第一百六十一輪：WALLDEF global symbol offset

狀態：`READY`（限 raw WALLDEF → global 8×8D symbol lookup）

## 已確認規則

reference `WallDefBlock.Offset` 只調整大於等於 `0x2D` 的 graphic ID；`LoadWalldef` 對載入的 symbol set 使用 `symbol_set_fix[symbolSet] - symbol_set_fix[1]`。因此 LOAD PIECES 的三個 WALLDEF slot 使用 global base：

| WALLDEF slot | global 8×8D base |
|---|---:|
| 1 | `0x2E` |
| 2 | `0x74` |
| 3 | `0xBA` |

這與 `Put8x8Symbol` 的 set ranges 一致；`0x01` 是 area-map 的另一個 8×8D set，不應混入 dungeon WALLDEF offset。

## 實作結果

- `gfx.WallDef.OffsetSymbols` 保存 structural IDs `<0x2D`，只平移 graphic IDs。
- `gfx.ParsePieceSet` 依 WALLDEF record index 計算 global set，並保存 `SymbolSetIDs`。
- `gfx.PieceSet.WallSymbol` 以 WALLDEF cell → global symbol ID → 8×8D item 做 bounded lookup。
- regression 覆蓋多 record 的 set 2／3 offset 與 item range。

## 邊界

本輪尚未把 `draw_3D_8x8_titles` 的九種 viewport layout、GEO 方向／深度遍歷與 sky／roof layer 接成完整 3D 畫面；目前 API 只提供可驗證的素材 lookup，避免把 raw WALLDEF column 當作畫面座標。
