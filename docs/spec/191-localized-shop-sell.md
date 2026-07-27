# 第一百九十一輪：localized shop sell

狀態：`READY`

## 證據與決策

DOS `ItemRecord` 已解碼 `Value`（`.SWG` record `0x3A:0x3C`）。在尚未取得各城市 stock／鑑定 routine 的前提下，販售 transaction 使用該已保存的 item value；不臆測額外城市倍率或隨機價格。

## 實作

- Shop Menu 新增繁中 `販售`。
- 玩家依序選角色、選物品；成功時移除 equipment item，將 `Value` 加入 party money pool，並刷新該角色的 fighter projection。
- readied／cursed／非正值 item 被拒絕，避免破壞 inventory transaction contract。
- success／failure 都回到可繼續的繁中 event screen。

## 驗證

state regression 覆蓋 menu path、75 GP sale、roster mutation、money pool，以及 readied／cursed rejection；全套 Go tests 與 CLI build 在 Docker 通過。

## 邊界

原版城市 stock、ID fee routine、鑑定 flag／文字與出售價格修正仍需各自 ECL／反組譯證據。
