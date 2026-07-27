# 第一百九十二輪：localized shop identify

狀態：`READY`

## 證據與邊界

CoAB 手冊／既有 `internal/party.PayIdentifyFee` contract 確認商店鑑定費為 200 GP。`.SWG` 的 `HiddenNameFlags` 已保存，但目前沒有足夠 evidence 把每個 flag 映射成完整魔法名稱、加值或效果，因此本輪不改寫該 raw flag。

## 實作

- Shop Menu 新增繁中 `鑑定`。
- 玩家依序選角色、選物品；`Character.PayIdentifyFee` 成功扣除 200 GP。
- item 不會因 fee transaction 被移除或被猜測改名；success message 明確指出辨識資料仍待載入。
- 金幣不足、角色／物品索引錯誤會安全返回 event error message。

## 驗證

state regression 覆蓋 menu path、250 → 50 GP、item raw `HiddenNameFlags` 保留與繁中 result boundary；全套 Go tests 與 CLI build 在 Docker 通過。

## 後續

仍需反組譯城市 ID routine、各 magic item name／effect table、curse reveal 與完整識別後 UI。
