# 第八十一輪：ECL PICTURE → 繁中事件畫面

狀態：`READY`（限 PIC event request 與目前 Ebiten playback slice）

## 本輪成果

- ECL opcode `0x0E PICTURE` 會解碼一個 block operand，`0xFF` 保持清除／不請求語意，其餘值產生 `RunResult.PictureRequested` 與 `PictureBlock`。
- `game.State` 保存 picture request，進入 `ModeEvent`，Enter／`Continue` 後清除 request 並返回原本的 wilderness／place event parent mode。
- Ebiten 由 `Area.GameArea` 與 block 組合 `picN-block-XX` animation key，播放 PIC manifest 中的 delay／x／y frames；缺素材時顯示繁中 fallback。
- ECL runtime 與 game state regression 覆蓋 request、block ID 與可恢復返回。

## 邊界

第 82 輪已補上 `BIGPIC` block 分支；目前仍尚未完整接入 BIGPIC 的 redraw side effects、HEAD/BODY overlay、PICTURE `0xFF` 行為與所有 ECL story path。本輪證明的是 PICTURE request → localized event screen 的 vertical slice。

## 驗證

```sh
go test ./...
```
