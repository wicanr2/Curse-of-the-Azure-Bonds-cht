# 第八十三輪：HEAD／BODY 場景人物圖層

狀態：`READY`（限一般 scene character 素材與合成）

## 已確認

- reference 一般角色場景使用 `HEAD<game_area>.DAX` 與 `BODY<game_area>.DAX`，不走 CHEAD／CBODY combat icon loader。
- `draw_head_and_body` 先在同一 row/column 畫 HEAD，再以 `row + 5` 畫 BODY。
  此處 row 是 8-pixel text row，因此實際 BODY offset 是 40 pixels；早期 generator
  曾誤作 5 pixels，造成頭像落進胸口，已由第 322 輪視覺 oracle 修正。
- reference 這條路徑呼叫 `LoadDax(0, 0, ...)`，因此 generator 以 unmasked indexed picture 解析 HEAD／BODY。

## 本輪成果

- 抽取 `HEAD2–6.DAX`／`BODY2–6.DAX`：40 張 HEAD、31 張 BODY PNG。
- 依 area 與共同 block ID 產生 30 張 `character-area-N-head-XX-body-XX.png` 合成圖。
- `gfx.MergePicturesAt` 成為可重用的 indexed layer API，保留 transparent／bitwise-OR semantics 並支援 pixel offset。
- Ebiten asset loader 會載入 scene character composites，後續城鎮／事件畫面可直接使用相同 key。

## 邊界

尚未將 ECL `PICTURE` 的 HEAD/BODY overlay branch、城鎮 NPC record 與完整場景 script 接入 game state；本輪完成的是原始素材、offset composition 與 loader boundary。

## 驗證

```sh
go test ./...
go run ./scripts
```
