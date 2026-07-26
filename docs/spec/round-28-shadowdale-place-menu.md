# 第二十八輪：Shadowdale 場所 menu contract

狀態：`READY`（限已觀察選項與 state-level event contract）

原始 ECL1 block `0x51` 的 bounded trace 可觀察到：

```text
YOU ARE IN SHADOWDALE. WHAT PLACE WILL YOU VISIT?
menu: INN / STORE / BAR / LEAVE
```

本輪將這組已確認的選項接入 Shadowdale map slice：

- 地圖模式按 Enter 開啟場所 menu。
- `INN`、`STORE`、`BAR`、`LEAVE` 顯示對應繁中事件文字。
- 客棧／商店／酒館事件按 Enter 返回場所 menu。
- `LEAVE` 按 Enter 返回地圖模式。

這是可測試的 UI／state contract，不宣稱已完成原始場所內部的角色、商店交易、酒館情報、客棧休息或戰鬥效果；那些仍需解碼 block `0x51` 的完整 command path。

驗證：`TestShadowdalePlaceMenuAndEvents`，以及 `go test -vet=off ./...`。

- [x] 新增 `ModePlace` 與場所選項。
- [x] 接入繁中場所名稱與事件回復。
- [x] Ebiten 以 Enter 開啟場所 menu。
- [ ] 將 state contract 對齊完整 ECL1 block `0x51` command path。
- [ ] 實作場所內部功能與 AD&D 規則。
