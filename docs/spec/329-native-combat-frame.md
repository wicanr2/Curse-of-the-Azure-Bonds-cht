# 第三百二十九輪：原生 320×184 戰鬥框 raster

狀態：`READY`

## Oracle

以兩張 DOS 版 2× 戰鬥截圖逐像素比較：

- <https://simeonpilgrim.com/blog/static/977b71ee/combat-aim%402x.png>
- <https://simeonpilgrim.com/blog/static/8d8a2230/fight-black-dragon%402x.png>

兩場戰鬥的外框、中央 divider、內緣點狀線及裂紋位置完全相同。原始 ZIP 的
94 個 members 也沒有獨立 UI frame DAX；DUNGCOM／WILDCOM／RANDCOM 只包含
24×24 戰場 terrain。現有證據因此支持固定 panel raster／drawing routine，
不支持「每場 encounter 載入 stone-frame tile」。

## Native geometry

先在原生 320×184 透明 canvas 產生 frame，再以 nearest-neighbour 2× 畫入
640×480 remake：

- top：`(0,0,320,8)`
- left：`(0,0,8,184)`
- center divider：`(176,0,8,184)`
- right：`(312,0,8,184)`
- bottom：`(0,176,320,8)`

兩個 panel interiors 保持透明，不能由 frame codec 塗掉 battlefield／status。
內緣使用原生 1px white／light-gray／dark-gray／black bevel，加上 alternating
pixel dotted edge。固定裂紋也只落在 native integer pixels；2× 後每筆必為
2px 倍數，不能再使用舊 renderer 的 3px 任意線寬。

## 實作

- `gfx.CombatFrame()` 回傳標準庫 `image.RGBA`，不依賴 Ebiten。
- Ebiten app 啟動時只轉換一次，繪製時指定 `FilterNearest` 並 scale 2。
- frame 在 terrain 前後各 overlay 一次，確保 terrain 與大型 sprite 都被 panel
  clipping topology 包住，而透明 interiors 不影響內容。

## 驗收

- tests 固定 320×184 bounds、兩個透明 interiors、五個 opaque frame regions、
  alternating inner edge 與固定 crack pixels。
- `docs/screenshots/gold-box-layout-combat.png` 已重新由 Docker/Xvfb 擷取；
  可見 16px 外框／divider、2px 原生裂紋與點狀內緣。

## 邊界

這是依 DOS oracle 重建的固定 raster contract，不宣稱已找到遺失的 SSI 原始
函式名稱或 source code。若後續 GAME.OVR 反組譯提供不同 crack table，應以該
證據更新座標並清除本規格被推翻的部分。
