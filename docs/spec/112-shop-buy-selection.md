# 第 112 輪：Shop BUY 商品選擇 UI

狀態：`READY`

## 本輪成果

- Shop Menu 的 `BUY` 會列出注入的 `ShopOffer`，顯示繁中物品名與 GP 價格。
- 選取商品後，從 party money pool 扣款並加入目前 shop character 的未 ready inventory。
- 成功購買後移除該 stock entry、顯示繁中確認訊息，並返回 Shop Menu。
- 沒有 stock 時保留「商店庫存尚未載入」訊息；沒有 party／金幣不足時保留可診斷錯誤。
- `SetShopCharacter` 提供目前 active character 邊界，預設為第一位；完整原版 VIEW
  角色選擇仍待接入。

價格仍由 city／ECL data 注入，未從 `ITEMS` descriptor 猜測。
