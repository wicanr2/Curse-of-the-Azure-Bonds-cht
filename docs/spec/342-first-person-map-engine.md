# 第三百四十二輪：第一人稱地圖共用引擎與 640×480 viewport

狀態：`READY`（限 GEO／WALLDEF／8X8D 共用邊界、CoAB JSON 資源與正式地城版面）

## 共用引擎

- `golden-box-remake-engine/geometry` 擁有 16×16 GEO decoder、牆／門／移動與 wrap。
- `viewport` 擁有 reference Far／Mid／Near `Draw3dWorld` traversal。
- `graphics` 擁有 indexed picture、EGA RGBA、WALLDEF、LOAD PIECES symbol offset
  與 8X8D wall-stamp expansion。
- `graphics.ParsePieceSet` 只接收 decoded block map，不依賴 CoAB `internal/dax`。

## 資料包

CoAB JSON 的散提爾堡內城宣告 area 4、GEO block `0x20`、spawn `(2,0,S)`、
2× scale，以及 `GEO4.DAX`／`WALLDEF4.DAX`／`8X8D4.DAX`。engine 驗證資源名稱
必須是 base filename；作品 Go adapter 不再 hardcode 這張地圖的資源名稱。

## 版面

WALLDEF logical columns `-5..15` 共 22 格，原生寬 176px，2× 後為 352px。
正式地城畫面因此使用左側 352px 第一人稱 viewport、右側 roster、下方繁中
敘事與最底命令列。debug floor 不再混入 production layout。

實機圖：
[`tilverton-first-person-remake.png`](../screenshots/tilverton-first-person-remake.png)。
它由 `-opening` 的真實 ECL1 block `0x01`、GEO2 block 1 與原始
WALLDEF2／8X8D2 在 Docker／Xvfb 中產生，尺寸 640×480。

## 驗證

- 獨立 engine `go test ./...` 全數通過。
- CoAB Docker／Xvfb `go test ./...` 全數通過。
- game-pack regression 鎖定散提爾堡資源、block、spawn 與方向。
- 原始 ZIP integration 繼續驗證各 GEO block shape 與 WALLDEF2／8X8D2 mapping。
