# 第七十四輪：SPRIT animation codec

狀態：`READY`（限 SPRIT frame stream decode）

## 本輪成果

- 依 reference `load_pic_final` 還原 `SPRIT*.DAX` block layout：frame count、4-byte delay、height／width、位置、保留 byte、8-byte metadata 與 packed EGA pixels。
- `internal/gfx.ParseAnimation` 產生帶有 delay、座標與透明 indexed pixels 的 `AnimationFrame`。
- 從 `SPRIT1.DAX`–`SPRIT6.DAX` 實際抽出 138 個 frame PNG，加入 [`assets/sprites/README.md`](../../assets/sprites/README.md)。
- parser 不把標準 picture header 套到 SPRIT，並對零 frame、異常 dimensions、truncated pixels 與 trailing bytes 報錯。

## 邊界

目前只完成 SPRIT 原始 frame stream 與獨立 PNG extraction；`PIC/FINAL` 的跨 frame XOR delta、動畫播放時序、方向 flip 與 combat placement 尚未接入 renderer。

## 驗證

```sh
go test ./internal/gfx ./...
go run ./scripts
```
