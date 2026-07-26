# 第一百二十五輪：ALTER ICON player sprite selection

狀態：`READY`（限已抽取 CHEAD／CBODY block 的 player icon selection）

## 證據

原始素材 manifest 已抽出四組 CHEAD／CBODY block family：`0x00`、`0x40`、`0x80`、`0xC0`，每組目前可用 block `0x00–0x0D`。角色資料已有 `IconHeadBlock`／`IconWeaponBlock`，combat fighter 也保存對應 party icon 欄位。

## 實作 contract

`CAMP → ALTER → ICON` 先選角色，再以頭部／身體 status、上一個／下一個與完成選項循環選擇已驗證 block。每次切換都寫回 roster，並以 stable character ID 更新目前 combat fighter 的 `PartyHeadBlock`／`PartyBodyBlock`；退出角色編輯返回 ICON 角色列表，再返回 ALTER Menu。

這一輪不猜測原版 recolor、武器語意、attack-state `+0x80` 或未抽出的 icon block；完整 save persistence 仍由 remake save adapter 後續補強。

## 驗證

`TestCampAlterIconUpdatesRosterAndFighter` 覆蓋角色選擇、head／body 循環、roster／fighter sync 與返回流程；`go test ./...` 已通過。
