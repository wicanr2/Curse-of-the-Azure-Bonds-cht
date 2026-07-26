# 第 110 輪：繁中 Shop Menu state

狀態：`READY`

## 本輪成果

城市 `STORE` 現在會進入可操作的繁中 Shop Menu，選項對齊原版手冊：

`BUY／VIEW／TAKE／POOL／SHARE／APPRAISE／EXIT`

每個 action 都有穩定的原始 command mapping。`EXIT` 返回城市場所選單；其他
action 會先顯示明確的「尚未載入／尚待接入」訊息，按 Enter 後返回 Shop Menu。
這使 UI、ECL command 與後續 stock／money-pool implementation 有固定邊界，且
不會把未知價格或庫存偽裝成完成的商店。

## 尚未完成

`BUY` 尚待接入實際 shop stock 與 offer price，`VIEW`／`TAKE`／`POOL`／`SHARE`／
`APPRAISE` 尚待接入角色選擇、party money pool 與物品／寶石 appraisal。底層
price-injected transaction 已在第 109 輪完成。
