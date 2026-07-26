# 第 133 輪：MEMORIZE preparation time

狀態：`READY`（限一級法術的 bounded preparation timing）

## 證據

RuleBook 明確記載：記憶法術需要每法術等級 15 分鐘，另有最低準備時間；一、二級最低 4 小時，三、四級 6 小時，五級 8 小時。本輪對目前已核對的一級 spell catalog 實作「4 小時 + 每個一級法術 15 分鐘」，再向上取整至目前 REST 的整小時單位。

若 `REST_START` 的時間不足，pending selection 不會被清除，也不會錯誤寫入 `SpellSlots`；畫面會告知所需時間。

## 邊界

目前只處理一級法術。二至五級、每角色平行準備的完整時間模型、被遭遇打斷後的部分成功與遊戲時鐘仍待接入。
