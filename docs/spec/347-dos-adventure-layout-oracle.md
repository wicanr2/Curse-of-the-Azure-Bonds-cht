# 第三百四十七輪：DOS 冒險畫面 layout oracle

狀態：`READY`

## 原版實機證據

原始 DOS 發行檔以 Docker／DOSBox 啟動後，在標題頁選 `D` 進入內建 demo，
取得原生 320×200 冒險畫面：

[`tilverton-first-person-demo.png`](../reference/original-dos/tilverton-first-person-demo.png)

畫面狀態列為 `(7,13) N 00:00`，但訊息明確寫著 `NOWHERE IN THE REAL...`。
因此這張圖是 GUI chrome、SKY 背景及座標欄的可靠 oracle，不能拿來斷言
GEO2/01 的實際牆配置。

## 量測與 640×480 投影

原版 top row 以 native `x=128` 分割，並在 `y=136` 進入訊息區：

- first-person panel：native 128×136 → remake 256×272；
- roster panel：native 192×136 → remake 384×272；
- status line 位於 roster 底部，而不是訊息框；
- 原版訊息區從 native y=136 開始；
- 原版 footer 從 native y=192 開始。

重製版維持所有原始圖像 2× nearest-neighbour。640×480 比原生 2× 的
640×400 多出 80px，只擴充繁中訊息區；top panels 與 footer 的水平比例不變：

- first-person `(0,0,256,272)`；
- roster `(256,0,384,272)`；
- message `(0,272,640,176)`；
- footer `(0,448,640,32)`。

倚天 16×15 粗體直接在 640×480 畫布繪製，不把原版 8px 英文字型放大成中文。

## 資料邊界

`tilverton.first-person` 在 CoAB JSON 明確宣告 area 2、GEO2/01、
WALLDEF2、8X8D2、SKY FA–FC、spawn `(7,13,N)` 及 outdoor sky selector 3。
共用 engine 的 `FindMapByKindLocation` 讓同一 GEO2/01 同時具有 AREA 與
first-person projection，不再因 `FindMapByKind` 取到其他城市的第一筆資料。

## 驗證

- engine commit `908cfb7` 的 schema／selector tests 通過；
- CoAB game-pack regression 驗證 Tilverton 與 Zhentil Keep 不互相覆蓋；
- Docker／Xvfb 正式 `-opening` 流程產生
  [`tilverton-first-person-remake.png`](../screenshots/tilverton-first-person-remake.png)；
- 完整 `go test ./...` 在 Xvfb 通過。
