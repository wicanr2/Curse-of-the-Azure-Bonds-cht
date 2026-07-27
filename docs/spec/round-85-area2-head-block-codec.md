# 第八十五輪：Area2 HeadBlockId codec

狀態：`READY`（限 Area2 `HeadBlockId` binary field 與 PICTURE sync）

## 已確認

- reference `Area2.HeadBlockId` 的 byte offset 是 `0x5C2`；`0xFF` 是沒有 HEAD layer 的 sentinel。
- reference `CMD_Picture` 會讀此 Area2 state：sentinel 時走 PIC/BIGPIC，其他值走 HEAD/BODY，body block 仍來自 PICTURE operand。

## 本輪成果

- `internal/area.State.HeadBlockID` 與 Area2 codec 已讀寫 `0x5C2`，並保留未知 bytes。
- `game.State.SetAreaState` 會同步 `HeadBlockID` 到 scene request boundary。
- 新增 raw Area2 codec regression 與 Area2 → ECL PICTURE → HEAD/BODY game regression。

## 邊界

Area2 仍只解讀已定位欄位；第 285 輪已接通 ECL mirror 的 `EnterTemple` service，
完整 Area2 loader 與其他未知 area side effects 尚未完成。

## 驗證

```sh
go test ./...
```
