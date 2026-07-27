# 第八十四輪：PICTURE HEAD/BODY branch

狀態：`READY`（限 Area2 head sentinel 與 scene character event slice）

## 已確認

- reference `Area2.HeadBlockId == 0xFF` 時，PICTURE 使用 PIC/BIGPIC；非 `0xFF` 時使用 `set_and_draw_head_body(body_id, head_id)`。
- `body_id` 來自 PICTURE block，`head_id` 來自 Area2 state；兩者不是同一個 party slot。

## 本輪成果

- game state 增加 `SceneHeadBlock`（`0xFF` sentinel）、`SceneBodyBlock`、`SceneCharacterRequested` 與 `SetSceneCharacter` contract。
- PICTURE result 在非 BIGPIC 且 head sentinel 存在時分流到 `character-area-N-head-XX-body-XX.png`。
- Ebiten 事件畫面顯示 HEAD/BODY composite，Enter 清除 request 並恢復 parent mode。
- regression 覆蓋 head block、body block 及 continuation。

## 邊界

本輪當時只建立 `SetSceneCharacter` adapter；後續第 85 輪已定位 Area2 raw record
`HeadBlockId @ 0x5C2`，第 282 輪再確認 ECL 執行期 mirror `0x7EE1` 必須在 PICTURE
opcode 當下擷取。完整 NPC／其他地圖事件仍待逐一接入。

## 驗證

```sh
go test ./...
```
