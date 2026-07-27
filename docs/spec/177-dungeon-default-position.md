# 第一百七十七輪：reference dungeon default position

狀態：`READY`

## 證據

公開 CoAB reference `engine/seg001.cs` 的 `Init()` 將 `mapPosX=7`、`mapPosY=0x0D`、`mapDirection=0`；`InitAgain()` 保留同一座標但將 direction 設為 2。這是初始化 context 的差異，不應把 `(8,8)` 當作作品真實起點。

## 本輪成果

- remake State、save helper、invalid/legacy save fallback、GEO preview fallback 與 startup dungeon floor 均改用 `(7,13,0)`。
- synthetic renderer tests 仍可使用任意座標；這輪只修正預設／恢復 contract，不改變 strict GEO bounds 或 16×16 wrap。
- `InitAgain` 的 direction 2 尚未映射成獨立 restart context；目前 shared default 依 `Init()` 的 direction 0，避免猜測呼叫時序。
