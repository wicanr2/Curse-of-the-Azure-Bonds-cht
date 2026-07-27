# 第三百零四輪：提爾佛頓放逐與阿沙本福德山徑

狀態：`READY`

## 章節返回

火刀首領線完成 `7F12=1 → NEWECL 0x50` 後，world dispatcher 顯示
BIGPIC 121 與 Tilverton edge menu。`ENTER CITY` 於 `+0x056C` 顯示
`GUARDS BAR YOUR WAY` 並返回城外；玩家已遭阿祖恩國王逐出科米爾，不能直接
重返提爾佛頓。

`JOURNEY ON` 於 `+0x0CEA` 提供 Shadowdale、Ashabenford、Dagger Falls。
選 Ashabenford 後，`TRAIL／WILDERNESS／EXIT` 路線選單使用 ECL world
destination bytes，不可再依新遊戲開場的固定 selection index 推斷地點。

## Tilver's Gap

Ashabenford 的 TRAIL 第一次通過提爾隘口時：

- `+0x12E2` 先寫 `4C83=1`，使事件只發生一次。
- 顯示群山、隘口與雪峰飛行身影的文字，沒有 PICTURE。
- `SETUP MONSTER 4,2,4`。
- `LOAD MONSTER 0x51,8,0x51`：八隻 MON1 `HIPPOGRIFF`，中文顯示「鷹馬」。
- 沒有 TREASURE；勝利後直接回 world dispatcher。
- dispatcher 寫 `4C9B=2`、`4CA1=2`，正式抵達 Ashabenford edge menu。

## State transaction

可玩 regression 從火刀首領勝利開始，實際經過 loot、手札 54／53、四段夢境、
Tilverton edge、JOURNEY ON、Ashabenford、TRAIL、八隻鷹馬戰，最後驗證：

- `LocationAshabenford` 與 `ENTER CITY／JOURNEY ON／CAMP`。
- `4C83=1`、`4C9B=2`、`4CA1=2`。
- 略過的火刀財寶不會在鷹馬戰後再次出現。

事件沿用 640×480 邏輯畫布與 24px 繁中；戰鬥 icon block `0x51` 仍按原始
像素以 nearest-neighbour 整數放大。
