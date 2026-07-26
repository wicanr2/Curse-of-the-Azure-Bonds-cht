# 第 109 輪：商店 Buy／Sell／ID transaction contract

狀態：`READY`

## 證據

原版繁中化研究使用的 CoAB RuleBook 說明 Shop Menu 包含 `BUY VIEW TAKE POOL
SHARE APPRAISE EXIT`；商店可購買／出售裝備，`ID` 服務收取 200 GP，購買從
party money pool 扣款。

## 本輪實作

- `Character.BuyItem(item, price)` 接受外部解碼的 shop offer price，扣除 gold，
  並以未 ready 狀態加入 inventory。
- `Character.SellItem(index, offer)` 只允許非 readied item，先檢查 gold overflow
  再移除一件並加入 shop offer。
- `Character.PayIdentifyFee()` 實作已確認的 200 GP fee，但不猜測 identification
  後的名稱／魔法效果。

`ITEMS` base descriptor 本身沒有已確認的價格欄位，因此交易 API 刻意要求價格由
後續 shop stock／原始 script adapter 提供；目前尚未宣稱完整 Shop Menu、money
pool、VIEW／TAKE／SHARE／APPRAISE UI 已完成。
