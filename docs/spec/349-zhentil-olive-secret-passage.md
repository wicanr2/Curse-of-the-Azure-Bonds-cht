# 第三百四十九輪：散塔林堡奧莉芙密道

Status: `READY`

## 目標

延伸已驗證的尤拉什／摩安德之坑／散塔林堡入口主線：隊伍在散塔林堡內城遇見
奧莉芙・魯斯克特爾，閱讀手札 50、選擇跟隨她，穿牆進入貝恩的幽暗神殿，再
閱讀手札 51，最後回到可操作的神殿地圖。這一段必須使用原版 ECL 與 GEO
資料推動，不另寫一套替代劇情。

## 原版證據

- `ECL4.DAX` block `0x20`、`GEO4.DAX` block `0x20`
  - `(10,11,N)` 的 terrain `0x04` 觸發 PICTURE `42`。
  - 同一事件也由 `(7,14)` 的 terrain `0x84` 入口觸發。
  - 對話依序為奧莉芙現身／手札 50、`DO YOU FOLLOW HER?`。
- 選 `YES` 後，原版 `NEWECL` 轉入 ECL4 block `0x21`：
  - PICTURE `42`，魔法裝置穿牆並消失。
  - 奧莉芙沿途說明迪姆斯沃特／解鎖手札 51。
  - 「迪姆斯沃特就在門後」後，奧莉芙離開，回到 dungeon mode。
- 手札文字以附帶的 *Adventure Journal* TXT／PDF 為原始資料；TXT 的第 51
  條在換頁處截斷，後半以收藏版 Adventure Journal 掃描交叉核對。
- 英文 walkthrough 僅作玩家路徑佐證：Olive 位於內城 location 11，接受後
  經密道進入 Shrine E2；它不取代 ECL/GEO 的執行證據。

## 資料與引擎邊界

- CoAB 專屬地名、人物、翻譯、手札與 text matching rule 寫入
  `gamepack/events/pit-of-moander.json`。
- 通用 `golden-box-remake-engine` 不得認識 Olive、Dimswart、Zhentil Keep
  或任何 CoAB 劇情。
- 仍由既有 ECL runtime 執行 `PICTURE`、choice、`NEWECL`、地圖載入及
  continuation；JSON 只提供 title data 與繁中呈現。

## 繁中呈現

- 正文沿用 640×480、16×15 倚天粗體字級，不另放大事件字。
- 人名統一為「奧莉芙・魯斯克特爾」、「迪姆斯沃特」、
  「弗佐爾・錢布瑞爾」；神器為「洛山達護符」。
- 手札 50、51 各拆成兩頁，保留奧莉芙活潑、像小說般可閱讀的口吻，避免為
  塞入一頁而刪去人物性格。

## 驗收

1. 在真實 `ECL4/GEO4 0x20` 的 `(10,11,N)` 觸發 PICTURE 42。
2. 畫面顯示繁中奧莉芙事件，手札 50 兩頁可讀。
3. 問句與 `YES/NO` 顯示為「要跟她走嗎？」與「是／否」。
4. 選「是」後確實轉到 block `0x21`，而非 title-local 假轉場。
5. 穿牆、沿途說明、門後提示、奧莉芙離場皆為繁中。
6. 手札 51 兩頁可讀，事件最後回到 `ModeDungeon` 與 `GEO4/0x21`。
7. `go test ./internal/game ./gamepack` 與完整 regression 通過。

## 本輪邊界

完成範圍止於神殿內重新取得操作權。迪姆斯沃特牢房、祭壇、占卜鏡、兜帽女子、
弗佐爾與德克薩姆屬後續玩家路徑，不在本規格中假裝完成。
