# 第二百三十二輪：README live combat screenshot（READY）

## Evidence

以 repo 目前的 Ebiten renderer、繁中 locale、原始 `curseoftheazurebonds.zip` 與
ECL1 direct-entry encounter 啟動遊戲，在 headless Xvfb 擷取 640×400 實際畫面：

```sh
go run ./cmd/azure-bonds-game \
  -font /path/to/chinese-font.ttf \
  -sound-dir /path/to/missing-sound-dir \
  -encounter -image curseoftheazurebonds.zip
```

產物是 [`docs/screenshots/combat-game.png`](../screenshots/combat-game.png)，可看到繁中
戰鬥訊息、party／enemy 小人、HP 與目標操作提示；素材來自 repo 的 parser／renderer
路徑，不是手工 mockup。

## Boundary

這張圖只證明 direct-entry playable combat slice 與 README progress evidence。它不宣稱
完整 opening→encounter journey、完整 CombatMap placement、所有 ECL routine 或完整
戰鬥 UI 已完成；這些仍依各自 READY spec 驗證。
