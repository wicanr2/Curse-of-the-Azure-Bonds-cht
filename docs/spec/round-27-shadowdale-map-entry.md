# 第二十七輪：Shadowdale 荒野入口

狀態：`READY`（限已觀察的入口狀態與輸入 contract）

原始 ECL1 的導航 sequence `0,0,1,0` 會得到：

```text
location: SHADOWDALE
menu: WILDERNESS / EXIT
```

本輪將這個 menu 接成第一個可操作的 map slice：

- 選擇 `WILDERNESS` 進入 `ModeMap`，初始座標為 `(0, 0)`。
- 方向鍵呼叫 `State.Move(dx, dy)`，Ebiten UI 顯示目前座標。
- `Esc` 呼叫 `State.LeaveMap()`，返回已知的 Shadowdale location menu。
- `EXIT` 直接返回已知的野外 menu。

座標目前是資料中立的輸入 contract；尚未把原始地圖 tile、碰撞、場所或遭遇表誤當成已解碼資料。

驗證：`go test -vet=off ./...`，以及 `TestShadowdaleWildernessMapMovementAndExit`。

- [x] 保存 `ModeMap`、座標與 Shadowdale map prompt。
- [x] 接入 Ebiten 方向鍵與 Esc。
- [x] 加入繁中 UI 與 state regression。
- [ ] 解碼原始地圖 tile／座標規則。
- [ ] 接入 ECL1 block 0x51 的 INN／STORE／BAR／LEAVE 場所事件。
