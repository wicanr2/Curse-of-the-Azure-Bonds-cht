# 第二百七十八輪：正式 new-game global block 0x01

狀態：READY

## Reference dispatch

`ovr003.sub_29758` 在 `Area.LastEclBlockId == 0` 時分成兩條路：

- `gbl.inDemo == true`：`EclBlockId = 0x52`；
- 正常玩家流程：`EclBlockId = 1`，先刷新 PartySummary。

之後共同執行 `load_ecl_dax → vm_init_ecl → RunEclVm(ecl_initial_entryPoint)`。
因此 ECL1 block `0x52` 是 demo，不是角色建立後的正式序幕；正常 new game 是 global
block `0x01`，位於 ECL2.DAX。

## 繁中畫面契約

Remake 的 Ebiten 邏輯畫布與預設視窗使用 `640×480`。88px PIC／人物 sprite 使用
nearest-neighbor 3×，304×120 BIGPIC 使用 2×；全部維持整數像素比例，不做模糊插值。
繁中則在 640×480 畫布上直接以 24px 字型重繪，不跟著低解析圖放大。新增的垂直空間
留給三行敘事與獨立操作提示，避免較大的中文字覆蓋圖片。

## Remake transaction

`BlockSession.Reset(0x01)` 建立全新的 shared RuntimeState、清除 menu／WHO selection
offset，與保留 memory 的 `NEWECL/Switch` 明確分開。`FinishCharacterCreation` 在 production
global ECL session 中呼叫 `BeginAdventure`，不再回到人造的城市／旅程／CAMP menu。
沒有 original image 的 data-neutral tests仍保留舊 fallback menu。
Production title 在沒有 party 時按 Enter 直接開角色建立；建立完成後自動走上述 block
`0x01`，不再要求玩家知道額外的 `C` debug shortcut。

## Real entry evidence

block `0x01:+0x0014`：

- `LOAD FILES 1,2,0xFF`；
- `LOAD PIECES 1,2,3`；
- 第一個 Continue pause：
  `YOU AWAKEN IN A SMALL ROOM...`／裝備與近期記憶消失；
- 第二段先 `PICTURE 0x0A`，再說明持劍手臂出現奇異圖紋、全隊都有相同印記；
- 第二個 Continue pause。

五段原文保留在 ECL result，zh-TW catalog逐行顯示繁中。PICTURE 與 menu 同時出現在
一個 result 時，State 保存通用 deferred result；Ebiten 在圖片下方以三行漸顯繁中文字，
關閉圖片後才恢復原始 Continue menu。

## Regression

real-image test從 `NewStateFromECLBlocks(...,0x50)` 開啟角色建立、加入角色並完成，
確認 session reset 至 `0x01`、兩段繁中、PIC boundary、deferred menu及 pieces 1/2/3。

```text
go test ./internal/ecl ./internal/game
```
