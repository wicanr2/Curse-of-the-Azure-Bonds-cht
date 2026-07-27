# 第二百八十四輪：ECL 城市商店服務

狀態：READY

## Reference evidence

CoAB reference `ovr003.CMD_Combat` 證明 `COMBAT (0x24)` 不只代表進入戰鬥。
當 `monstersLoaded == false` 且 `combat_type == normal` 時，它依序派送：

1. `Area2.EnterShop == 1`：先清除此旗標，再呼叫 `ovr007.CityShop()`；
2. `Area2.EnterTemple == 1`：先清除此旗標，再呼叫 temple shop；
3. 兩者皆否：執行一般戰後經驗與寶物處理。

只有已載入怪物或非 normal combat type 才進入 `MainCombatLoop()`。ECL player-memory
mirror 中，`Area2.EnterShop` 對應 `0x7F6C`，其下一欄 `field_6DA` 對應 `0x7F6D`；
`EnterTemple` 對應 `0x7EE2`。因此 VM 必須在 COMBAT boundary 先辨識 service flag，
不能把所有 `0x24` 都誤報為戰鬥。

ECL2 block `0x01` 的 Tilverton 商店分支提供真實資料證據：

- terrain `0x82`（General Store）在 YES 分支以 `TREASURE` 載入 `ITEM2.DAX`
  block `1`，設定 EnterShop，再執行 COMBAT；
- terrain `0x84`（Weaponers of Cormyr）在 YES 分支以 `TREASURE` 載入
  `ITEM2.DAX` block `5`，寫入 `EnterShop=1` 與 `field_6DA=0x10`，再執行 COMBAT；
- Weaponers 離開 service 後，原 ECL 從 COMBAT 下一條 instruction 繼續，顯示
  `MAY YOU ALWAYS STRIKE TRUE.`，再走回地城事件結尾。

`ovr007.CityShop` 以 `items_pointer` 作為庫存；此 list 正是 `TREASURE` item block
建立的商品描述。`ItemsValue` 依 `field_6DA` 計價：`01/02/04/08` 分別右移
4/3/2/1，`20/40/80` 分別左移 1/2/3，其餘值維持原價，所以 Weaponers 的
`0x10` 是原價而非乘以十六。原始 value 為零時先改成 1。

購買成功時 `PlayerAddItem` 將商品 `ShallowClone()` 加到角色 inventory，沒有從
`items_pointer` 移除原商品，因此庫存不耗盡。付款先檢查目前角色五種 coin 的
gold worth，再退回 pooled money；`CityShop` 進入時會清空既有 pooled money。

## Remake transaction

- bounded VM 將 `COMBAT` 建模為可恢復的 engine boundary：EnterShop 優先產生
  `ShopRequested` 與 raw price scale、消耗旗標；EnterTemple 產生獨立 service
  signal；其餘路徑才產生 `CombatRequested`。
- State 在 shop signal 發生時先解析同一 ECL result 的 `TREASURE` requests，以目前
  game-area namespace 取得 ITEM block，轉成 CityShop offers。
- 商品價格套用 reference shift table；`0x10` 保持原價，零價下限為 1。
- 每次購買複製商品到所選角色，不刪除 offer；付款順序為角色 typed coins，再使用
  pooled money fallback。
- 離開 ECL-backed shop 時，以 COMBAT 後已保存的 PC 與同一 `RuntimeState` 恢復原
  ECL，不重播 TREASURE、扣款或進店前對話，最後回到原地城座標。
- Temple 維持獨立 boundary，不誤套 CityShop adapter；其 UI／治療交易已由
  [第 285 輪規格](./285-gond-temple-service.md)依 `ovr005` 證據接通。

## Regression

- synthetic：SAVE `EnterShop=1`、SAVE `field_6DA=0x10`、COMBAT 產生 shop signal，
  清除 EnterShop，且 resume 後執行 COMBAT 下一條 PRINT／EXIT；
- real image：正式 Tilverton dungeon lifecycle 在 terrain `0x84` 顯示 Weaponers
  PICTURE／YES-NO，YES 以 `ITEM2.DAX` block `5` 進入 shop；
- 至少一件商品以原 value 顯示，購買後角色 typed-coin gold worth 正確下降，而相同
  offer 仍留在庫存；
- EXIT 續跑 `MAY YOU ALWAYS STRIKE TRUE.`，再返回觸發商店的地城格；
- 一般有怪物的 COMBAT regression 仍產生 battle boundary，不被 shop dispatch 攔截。

## Reusable implications

Gold Box 共用 VM 應把 `COMBAT` 視為「依 engine state 派送外部服務」的 opcode，
而不是固定 battle call。可共用的是保存下一條 PC、輸出 typed boundary signal、
消耗一次性 service flag，以及 resume 同一 ECL session；Area2 mirror 位址、
TREASURE／ITEM namespace、price-scale table、付款幣制與實際 shop UI 都必須由各作品
reference 與 real-image regression 重新驗證。
