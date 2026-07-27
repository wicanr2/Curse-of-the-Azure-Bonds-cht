# 第三百零六輪：艾森布拉、哈普與黑龍

狀態：`READY`

## 立石群的複合選項

ECL1 block `0x50` 在 Standing Stone 提示結束後顯示 `PATROL FOREST`、
`JOURNEY ON`、`CAMP`。原始 horizontal-menu token stream 會把第一項的兩個
英文單字解成兩個顯示片段，但它們只佔一個選項索引。

State adapter 必須把完全相符的 `PATROL / FOREST / JOURNEY ON / CAMP`
正規化回三項；否則畫面上的 `FOREST` 會錯誤對應到實際的 `JOURNEY ON`
branch。這是顯示文字 framing，不能改動 VM 的 selection index。

## Essembra route

Standing Stone 的 `JOURNEY ON` 提供 Ashabenford、Essembra、Hillsfar。
選 Essembra 再選 TRAIL，本輪固定流程沒有額外遭遇，抵達城市邊緣後：

- world current-location `4C9B=8`；
- State 投影為 `LocationEssembra`；
- edge menu 為 `ENTER CITY / JOURNEY ON / CAMP`。

地點編號不是連續 enum；可重用 engine 必須保存作品原始值。

## Hap 與三隻黑龍

Essembra 的 `JOURNEY ON` 提供 Hap 與 The Standing Stone。選 Hap 的 TRAIL
會顯示天空中的巨大黑影俯衝而下，接著由 MON1 record `0x35` 建立三名敵人：

- raw name `BLACK DRAGON`，繁中顯示「黑龍」；
- count `3`；
- combat icon block `0x35`；
- 本次原始資料投影 HP 48、AC 3。

事件在開戰前先寫入 `4CA2=1`；這代表「Hap 黑龍接近事件已觸發」，不能誤解成
已戰勝黑龍。遭遇沒有 PICTURE、TREASURE 或 NEWECL。

勝利 continuation 必須恢復同一份 ECL session，抵達 Hap village edge，
寫入 `4C9B=9`、`4CA1=9` 並投影為 `LocationHap`；不可直接返回通用荒野選單。

## UI contract

所有場景使用 640×480 邏輯畫布。原始圖片與戰鬥小人只做 nearest-neighbour
整數倍放大；繁中直接以高解析字型重繪，一般敘事／選單採 24×24 級距，
空間受限的數值欄位可採約 16×15。
