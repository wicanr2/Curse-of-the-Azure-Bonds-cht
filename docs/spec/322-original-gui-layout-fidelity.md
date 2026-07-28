# 第三百二十二輪：原版 GUI 與戰鬥版面忠實度

狀態：`READY`

## 證據

- Simeon Pilgrim 的原版 DOS 截圖同時顯示城市事件與戰鬥畫面：
  <https://simeonpilgrim.com/blog/curse-of-the-azure-bonds-screenshots>
- MobyGames 的各平台截圖索引用來排除 Amiga／C64 等不同版面：
  <https://www.mobygames.com/game/503/curse-of-the-azure-bonds/screenshots/>
- 本地原始 DOS 發行檔已在 Docker 內以 DOSBox 啟動；環境與防拷畫面保存在
  `docs/reference/original-dos/`。完整遊戲畫面的 layout oracle 以上述 DOS
  截圖為準，不以 remake 既有畫面反推原版。
- reference `draw_head_and_body` 將 BODY 畫在 `row + 5`。此處 row 是 8px
  文字列，因此偏移量是 40px，不是 5px。

## 640×480 冒險版面

本節早期依公開截圖估出的 264／360 分割已由第 347 輪本機 DOSBox 原生
320×200 oracle 取代。正式比例是左 128／右 192 native pixels；2× 後為：

- 左上 256×272 圖像框；
- 右上 384×272 隊伍表，固定保留姓名、AC、HP 與底部狀態列；
- 下方 640×176 是擴充繁中敘事區；
- 最底 640×32 是倚天 16×15 compact command line。
- HEAD 先畫、BODY 在 y+40 後畫；BODY 的肩膀／領口可遮住頭部下緣，
  palette index 0 與透明 sentinel 不得塗黑 HEAD。

## 640×480 戰鬥版面

本節的初步矩形已由第 323 輪 DOS 截圖逐像素量測取代；保留本輪作為第一次
修正「自由置中 prototype」的歷史，不再把 352×376 戰場視為正確原版比例。

- 左側 352×376 是 7×7 戰術格與原版 CPIC／CHEAD／CBODY 小人。
- 右側 272×376 是目前角色與目標的姓名、HP、AC。
- 下方 624×64 是戰鬥訊息；最底 624×28 是
  `移動／查看／瞄準／使用／施法／快速／結束` 命令列。
- 戰場是獨立 clipping region。大型怪物、死亡效果及 selection marker
  不得越過右側狀態欄。
- 尚未解出的原版 terrain selector 不得以整套 TILES icon atlas 依序鋪滿；
  本輪使用中性的離散戰術格，等待 combat-map selector 證據。

## 驗收

- `assets/sprites/character-area-2-head-09-body-06.png` 是 88×88 正確頭身合成。
- `docs/screenshots/gold-box-layout-adventure.png` 顯示四區冒險 layout。
- `docs/screenshots/gold-box-layout-combat.png` 顯示裁切後的四區戰鬥 layout。
- gfx、ECL、game regression 與完整 `go test ./...` 通過。
