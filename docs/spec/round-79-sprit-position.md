# 第七十九輪：SPRIT frame position offset

狀態：`READY`（限目前 SPRIT playback 的 frame offset）

## 本輪成果

- `scripts/render_previews/` 將每個 SPRIT frame 的 `Picture.X`／`Picture.Y` 寫入 `assets/sprites/animation.json`。
- Ebiten `combatAnimation` 載入並保存 `x/y`，播放時以原始 pixel 座標乘目前 2× render scale 套用 offset。
- animation manifest 與 `assets/sprites/README.md` 同時記錄 delay、x、y，方便後續 Gold Box 遊戲共用。

## 已知語意

`x/y` 是 frame 在 combat icon canvas 的 placement metadata，不是 world map 座標；本輪只把它套入 sprite draw origin，尚未完成 reference 的 PIC/FINAL XOR delta、direction-specific placement 或八方向 combat grid。

## 驗證

```sh
go test ./...
go run ./scripts
```
