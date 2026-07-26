# 第十九輪：Shadowdale map state

狀態：`READY`（限地點入口 state）

ECL1 block 80 的可重現 sequence：

```text
0 ENTER CITY
0 CONTINUE
1 JOURNEY ON
0 SHADOWDALE
→ WILDERNESS / EXIT
```

`game.State` 現在具有 `Location`：預設 `LocationWilderness`，完成上述 Shadowdale selection 後轉為 `LocationShadowdale`，並保留 `OriginalLocation == "SHADOWDALE"`。`WILDERNESS`／`EXIT` 已加入繁中 locale（荒野／離開）。

這是 map state 的第一個入口，不是完整地圖引擎；座標、移動、場所功能、戰鬥與存檔仍未完成。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -run-subset -interactive -select 0,0,1,0
```
