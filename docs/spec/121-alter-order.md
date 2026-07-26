# 第一百二十一輪：ALTER ORDER party reorder

狀態：`READY`（限 party roster／目前 combat fighter 的順序 contract）

## 證據

RuleBook 的 CAMP 章節定義 `ALTER: ORDER DROP SPEED ICON PICS EXIT`，並說明 ORDER 會改變角色在畫面與戰鬥部署中的順序。本輪只實作這個明確的順序 side effect。

## 實作 contract

`CAMP → ALTER → ORDER` 先選要移動的角色，再選新的位置。成功後依同一個順序重排 `partyRoster`，並以 character ID 對齊目前的 `combat.Fighter` slice；未匹配的 fighter 保留在尾端，不以 slice ordinal 猜測角色身份。Enter 返回 ALTER Menu，離開 ALTER 才回 CAMP Menu。

DROP 已由獨立規格接入永久刪除確認，PICS、SPEED 與 ICON 已分別接入 renderer preferences／message reveal／verified player sprite selection。

## 驗證

`TestCampAlterOrderReordersRosterAndFighters` 覆蓋兩階段選單、roster reorder、fighter reorder 與返回流程；`go test ./...` 已通過。
