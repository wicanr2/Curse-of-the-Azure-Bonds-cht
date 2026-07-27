# 第二百八十一輪：Dungeon CAMP lifecycle

狀態：READY

## Reference

`ovr003.TryEncamp` 的順序是：

1. `RunEclVm(PreCampCheckAddr)`；
2. `MakeCamp()`；
3. 只有休息真的 interrupted 時，`LoadPic()` 並
   `RunEclVm(CampInterruptedAddr)`；
4. redraw dungeon view。

Block `0x01` entry 2（`+0x01EF`）依 map registers 設定：

- `(x>=5 && y>=13)`：`0x7ED2/0x7ED3 = 0/0`，不擲休息遭遇；
- `x<5` 或 `y<13`：`1/100`，每小時檢查且必定中斷。

Entry 3（`+0x0225`）是 CampInterrupted；unsafe branch 會輸出
`A PATROL ARRIVES.` 或 `ROYAL GUARDS TELL YOU TO MOVE ALONG.`，經 Continue 後 EXIT。

## Remake transaction

正式 `ModeDungeon` 按 `E`：

1. 同步 `0xC04B..0xC04F`；
2. 執行 lifecycle entry 2；
3. 保存 period／percentage；
4. clean EXIT 才開啟既有繁中 CAMP menu。

安全起點 `(7,13)` 可正常休息。unsafe cell 的 24 小時休息會在第 1 小時中斷，只推進
已完成時間、不套用完整 24 小時自然治療或法術記憶，關閉 CAMP 後執行 entry 3。皇家巡邏
文字已翻成繁中；原始 Continue boundary 消費完後回到同一個提爾佛頓 3D 座標。

CAMP 的 `EXIT` 使用 `campReturnMode` 回到 `ModeDungeon`，不再落入早期 synthetic
wilderness menu。非 dungeon 的既有 CAMP 行為保持相容。

一般 640×480 event message 現以 22-rune、最多五行的 24px layout 顯示，明確保留 ECL
文字中的 newline，避免 interruption 與後續中文事件溢出畫面。

## Regression

真實 image test 覆蓋：

- `(7,13)` PreCamp `0/0`、開 CAMP、EXIT 返回相同 dungeon；
- `x=4,y=13` PreCamp `1/100`；
- REST_START 一小時中斷；
- entry 3 皇家衛兵繁中、Continue、返回 `ModeDungeon`。

