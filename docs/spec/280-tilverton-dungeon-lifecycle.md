# 第二百八十輪：提爾佛頓 GEO1 與 dungeon lifecycle

狀態：READY

## Corrected LOAD FILES semantics

Reference `CMD_LoadFiles` 依序把三個 operand 命名為 `var_3/var_2/var_1`：

- dungeon 時，operand 1（`var_3`）選 `current_3DMap_block_id`；
- outdoor 時，operand 3（`var_1`）決定是否 reload BIGPIC；
- `0xFF/0x7F` sentinel 規則保持不變。

因此 block `0x01` 的 `LOAD FILES 1,2,0xFF` 在新遊戲室內狀態載入 GEO2 block `1`，
不是「第三 operand 為 FF，所以不載 GEO」。舊 adapter 的 operand 次序已修正。

## Five lifecycle entries

`vm_init_ecl` 的五個 word entry 依序是：

1. per-turn；
2. SearchLocation；
3. PreCampCheck；
4. CampInterrupted；
5. initial。

`BlockSession.RunEntry` 會在保留 shared memory 的前提下，清除 call stack 並從指定 entry
開始新的 `RunEclVm` invocation。EXIT 現在也保存 memory writes。這與 menu pause 後
resume PC 的 `RunFrom` 是不同 transaction。

每次成功向前移動後，State 將 `(x,y,direction/2,wallType,wallRoof)` 同步到
`0xC04B..0xC04F`，依序執行 per-turn 與 SearchLocation。文字、PICTURE、menu、combat
及既有 State signals 都可形成 UI boundary；兩者 clean EXIT 時維持 `ModeDungeon`。

## Renderer

正式新遊戲完成後會自動使用既有 GEO／WALLDEF／8X8D 3D renderer，不再要求按 `D`
進入 debug preview。起點為提爾佛頓 GEO1 `(7,13)`、面東；↑ 前進，K/M 轉向。

正式 640×480 layout 將 13×5 TILES composition 與 3D WALLDEF view 分成左右兩區，
位置／方向／牆面／屋頂與操作列分行顯示。字型 loader 支援 TTC collection，繁中以
24px 直接重繪。headless `-opening` 實機證據保存於
[`docs/screenshots/tilverton-opening.png`](../screenshots/tilverton-opening.png)。

`ovr011.SetupWildernessFloor` 的 50×25 buffer 位於 `SetupGroundTiles` combat path，
是野外遭遇的戰鬥地面生成器，不是此處的世界地圖。相關 README／knowledge 敘述已修正。

## Regression

- area tests 鎖定 operand 1 GEO／operand 3 BIGPIC；
- session tests 鎖定 lifecycle restart、EXIT memory persistence；
- real-image test 鎖定 GEO1、ModeDungeon、位置／方向與 initial-cell entries clean EXIT。
