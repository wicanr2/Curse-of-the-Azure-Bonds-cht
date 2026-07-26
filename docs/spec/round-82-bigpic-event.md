# 第八十二輪：BIGPIC PICTURE 分支

狀態：`READY`（限 BIGPIC extraction 與目前事件畫面）

## 已確認

- reference `CMD_Picture` 對 block `>= 0x78` 呼叫 `load_bigpic`，從 `BIGPIC<game_area>.DAX` 載入 unmasked indexed picture；block `< 0x78` 才走 PIC animation。
- 本作原始映像有 `BIGPIC1.DAX`、`BIGPIC2.DAX`、`BIGPIC6.DAX`，目前共找到 4 個大圖 block。

## 本輪成果

- `RunResult` 增加 `BigPictureRequested`，仍保留原始 `PictureBlock`。
- game state 將 PICTURE request 分流保存；Ebiten 對 BIGPIC 使用靜態大圖 key 並置中顯示，Enter 可沿用既有 event continuation。
- generator 對 BIGPIC 使用 `masked=false` 的標準 indexed picture parser，輸出 `assets/sprites/bigpic*-block-*-item-00.png`。
- ECL regression 覆蓋一般 PIC block `0x1D` 與 BIGPIC block `0x78` 的分流。

## 邊界

尚未完成 BIGPIC 與完整 opening／story entry 的自動抵達、`can_draw_bigpic` redraw side effects、BIGPIC 缺檔 fallback catalog 與所有 area file variants。

## 驗證

```sh
go test ./...
go run ./scripts
```
