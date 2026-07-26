# 第 111 輪：shop stock／party money pool

狀態：`READY`

## 本輪成果

- `game.ShopOffer` 以 `ItemRecord + explicit Price` 表示城市商店庫存；價格由
  ECL／city data 注入，因 `ITEMS` descriptor 沒有價格欄位而不在 parser 猜測。
- `State.PoolPartyGold` 將角色金幣集中到 party pool。
- `State.TakeGold` 從 pool 提取到指定角色，並檢查單一角色 `uint16` 上限。
- `State.ShareGold` 依 marching order 平均分配，餘數由前方角色取得。
- `State.BuyShopOffer` 從 pool 扣除外部 offer price、加入未 ready item，並刷新已載入
  fighter projection。

Shop Menu 的 BUY 目前只確認 stock 已載入，實際 item selection UI 仍待下一輪；
VIEW、TAKE 數量輸入與 APPRAISE 仍保留明確 boundary。
