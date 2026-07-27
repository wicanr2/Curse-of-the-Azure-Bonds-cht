# 第一百六十三輪：Far／Mid／Near wall traversal

狀態：`READY`（限 GEO map traversal 與 layout call，尚未宣稱完整畫面）

## 已確認規則

reference `Draw3dWorld` 以 party direction 計算 `left=(dir+6)%8`、`behind=(dir+4)%8`、`right=(dir+2)%8`，從 party 前方兩格開始，依序處理 depth 2（Far）、1（Mid）、0（Near），每層沿 opposite side delta 推進一格。

- Far：中心牆 4 格、左右側牆各 3 格，並在連續牆／邊界 seam 時使用 layout 9。
- Mid：左右兩側各 3 格，中心 layout 3，側牆 layout 4／5。
- Near：左右兩側各 2 格，中心 layout 6，側牆 layout 7／8。

## 實作結果

- `gfx.TraverseWallView` 將 GEO wall fields 轉成有順序的 `WallLayoutCall`。
- dungeon preview 已改用整段 traversal calls，再展開成 WALLDEF／8×8D `WallStamp`，不再只取單一 wall sample。
- regression 覆蓋 invalid direction、depth ordering 與 Far／Mid／Near map coordinates。

## 邊界

第 167 輪新增 `TraverseWallViewWrapped`，讓 dungeon preview 明確套用 reference 16×16 wrap；未包裝的 `TraverseWallView` 仍保留嚴格邊界，供不允許 wrap 的 context 使用。sky／roof／door overlay、遮擋排序、camera movement 與完整 Ebiten 3D screen 仍待完成。
