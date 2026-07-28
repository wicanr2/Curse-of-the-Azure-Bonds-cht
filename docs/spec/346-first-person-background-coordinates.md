# 第三百四十六輪：第一人稱背景與 wall screen coordinates

狀態：`READY`

## 反組譯證據

公開 CoAB reference commit `9dc46f1`：

- `seg040.DrawColorBlock` 對輸入座標加 8px screen origin；
- `ovr031.Draw3dWorldBackground` 產生 native rectangles：
  sky `(24,24,88,44)`、黑色 horizon `(24,68,88,2)`、
  EGA dark-gray ground `(24,70,88,42)`；
- `SKY/FA` 只在戶外、sky palette 11、面北時顯示；
  `SKY/FB` 依 hour 1–5／13–18 與 E／S／W 方向移動；
  `SKY/FC` 固定 overlay 在 row 7、column 2；
- `draw_3D_8x8_titles` 僅接受 wall row／column `0..10`，呼叫
  `Put8x8Symbol(row+2,col+2)`；後者再加一格，因此實際 native screen
  position 是 `(column+3,row+3)×8`。

## 實作

- engine `8ea72d9` 的 `viewport.BuildBackground` 輸出原生 rectangles、
  indoor/outdoor palette 與 SKY overlay roles；schema 新增
  `sky_file`／`sky_blocks`。
- CoAB JSON 宣告 `SKY.DAX` blocks `[250,251,252]`。
- 原始 image regression 驗證 FA=`88×16`、FB=`24×24`、FC=`88×48`，
  均為單 item masked picture。
- renderer 在展開 WALLDEF 後先排除 0..10 外 stamps，再以正確 +3 cell
  transform 作 2× nearest-neighbour。
- 指定 `-eten-font` 時，16×15 embolden face 接管 regular 與 compact
  中文 UI，避免缺少系統 TTF 時出現方框。

## 驗證

- engine 與 CoAB focused tests、Ebiten build。
- Docker/Xvfb 由正式 `-opening` 流程擷取
  `docs/screenshots/tilverton-first-person-remake.png`。
