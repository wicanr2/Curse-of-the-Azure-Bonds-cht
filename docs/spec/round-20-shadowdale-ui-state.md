# 第二十輪：Shadowdale UI map state

狀態：`READY`（限 location state／UI visibility）

`game.State` 在完成 ECL1 Shadowdale sequence 後會保存：

- `LocationShadowdale`，原始名稱 `OriginalLocation == "SHADOWDALE"`。
- locale 驅動的 `LocationName`，Ebiten 會在 opening／menu 畫面顯示地點。
- 下一個 ECL menu 仍由 interactive sequence 控制；`WILDERNESS` 分支回到原始野外 menu，`EXIT` 分支保留為待實作事件。

這輪沒有假設座標或完整地圖格式；地圖移動、場所功能、戰鬥與音效仍需反組譯／回歸後逐步加入。

## 驗收

```sh
go test ./...
```
