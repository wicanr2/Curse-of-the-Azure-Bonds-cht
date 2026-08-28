# 391：DOS HEAD／BODY 人物舞台與左上內框

狀態：`READY`

## 證據

- 使用者提供的純遊戲擷取為 1014×759、4:3 顯示畫面。DOS 畫面經非方形像素
  顯示後不能直接拿螢幕座標套入 renderer，因此先以 nearest-neighbour
  正規化為
  [`character-head-body-oracle.png`](../reference/original-dos/character-head-body-oracle.png)
  的 320×200 邏輯畫面。
- 畫面人物逐像素比對後，確定是 `HEAD2 0x02 + BODY2 0x02`；現有
  `ComposeHeadBody` 的 HEAD 先畫、BODY 肩頸再覆蓋下緣順序正確。
- 合成素材是 88×88，原版 placement 為 `(28,24)`。黃色／灰色裂紋人物
  內框範圍約為 `x=21..121, y=16..116`，人物可見裁切區為
  `(27,22,90,90)`。
- 這張來源經顯示縮放後只能把 placement、layering 與 frame geometry
  標為 `material-exact/layout-reconstructed`，不能宣稱來源 RGB
  pixel-exact。抽出的內框已重新量化至 EGA 16 色。
- 2026-07-30 使用者另指定一張冒險訊息實機圖
  [`dos-character-info-layout-20260730.png`](../reference/user-provided/dos-character-info-layout-20260730.png)，
  SHA-256 為
  `2afa9ba4b37a205571dff582ee69b2a86b305f3f057343bc40dd9ba0ea9320c5`。
  它再次證明左上 NPC 圖使用同一黃色裂紋人物舞台。2026-08-27 由正常
  `CAMP → VIEW` 實機路徑取得的新 oracle 證明，該圖不是角色資訊頁；不能再用它
  支持 VIEW 的 HEAD／BODY、右側 roster 或下方長文斷言。

## 必要實作

1. HEAD／BODY 人物不能共用 PIC／第一人稱場景的 `cover＋clip`。
2. game pack 的 `presentation.scene_character` 宣告 native canvas、
   整數倍率、sprite anchor 與 clip；frontend 不得散落本作座標常數。
3. 人物以原生像素乘整數倍率繪製，保持完整手臂、武器與肩頸關係。
4. 人物內框在 sprite 後 overlay；外層 DOS cracked-stone chrome 最後再畫。
5. PIC／第一人稱場景仍維持既有 `cover＋clip`，不能因本次人物修正退回
   letterbox。
6. HEAD／BODY 契約只適用於冒險中的 NPC／人物舞台。角色 `VIEW` 是另一套
   全頁數值表，見 spec 1246，不得套用這份舞台配置。

## 驗證

- engine 測試驗證 presentation schema 與越界 clip 拒絕。
- gfx 測試驗證人物內框外緣不透明、中央透明。
- `-inn` 正常角色建立→序幕→移動→ECL 場所 dispatch 的玩家路徑產生
  [`tilverton-inn.png`](../screenshots/tilverton-inn.png)。
- `-temple` 正常玩家路徑驗證不同 selector 的 `HEAD2 0x09 + BODY2 0x06`，
  產生 [`tilverton-gond-temple.png`](../screenshots/tilverton-gond-temple.png)。

本輪只證明 DOS 人物舞台與這兩條玩家路徑；其他平台人物框、動態 palette
與所有事件人物逐張驗收仍需後續完成。
