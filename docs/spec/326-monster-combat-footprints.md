# 第三百二十六輪：怪物戰鬥佔格與大型 CPIC 錨點

狀態：`READY`

## 證據

reference placement routine 將 player record `field_DE & 7` 寫入
`CombatMap.size`。掃描原始 `MON1CHA.DAX` 至 `MON6CHA.DAX` 後，尺寸碼與怪物
外形一致：

| size | 佔格 | 原始資料例 |
|---:|---:|---|
| 1 | 1×1 | 一般人形怪物 |
| 2 | 1×2 | Ettin、Troll、Ogre、Salamander |
| 3 | 2×1 | Crocodile、Displacer Beast、Centaur、Manticore |
| 4 | 2×2 | Black Dragon、Dracolich、Wyvern、Bit O' Moander |

MON5 `0x3C` DRACOLICH 的 `field_DE & 7 == 4`，因此是本輪畫面驗收使用的
2×2 樣本。此欄位是形狀碼，不是像素寬高，也不是單純 large boolean。

## 實作契約

- monster record parser 保留 `CombatSize`，投影 fighter 時不得遺失。
- size 0／未知值保守回退 1×1；1、2、3、4 分別映射為上表。
- destination occupancy 必須比較兩個矩形的所有格，不能只比較原點。
- melee adjacency 是兩矩形邊界相接或斜角相接，但重疊本身不算 adjacency。
- restore／復活位置也必須拒絕落入大型怪物任一佔格。
- camera extent 必須包含 footprint 的右、下邊界。
- 640×480 renderer 的每格是原始 24px 的 nearest-neighbour 2×，所以
  2×2 marker 是 96×96。

## 畫面錨點

邏輯 occupancy 與 CPIC raster placement 必須分層。水平鏡像仍使用原始 combat
column 的 `6-x` 作 CPIC 左上繪圖錨點；大型 footprint 從該畫面點向右、下展開。
若在鏡像時再減 footprint width，DRACOLICH 的 2×2 marker 會有一半落到戰場
左側 clipping boundary，圖像也會被裁掉。

`docs/screenshots/gold-box-layout-combat.png` 是實機驗收：龍巫妖完整位於
96×96 白框中，背景來自 DUNGCOM／GEO lookup，右欄與繁中戰鬥紀錄維持原版
四區拓撲。

## 邊界

本輪沒有宣稱已還原非矩形 mask、方向別 CPIC offset、RANDCOM decoration、
原始 stone-frame tile 或 encounter-specific obstacle placement；這些仍需各自
從 reference code／DOS 畫面取得證據。
