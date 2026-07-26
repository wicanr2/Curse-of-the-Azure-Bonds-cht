# 第九十一輪：ITEMS → equipped fighter effect adapter

狀態：`READY`（限已知 base descriptor 的 readied 武器／護甲投影）

`monster.ItemRecord.Effect` 以 item type 查 `BaseItemCatalog`，將 base descriptor 的傷害欄、item magical plus 與 packed AC adjustment 投影成 `EquipmentEffect`。AC 值依 reference 規則解碼：`>=178` 為 body armor/bracers bonus `value-178`，其他 `>=128` 為一般 AC bonus `value-128`。`party.Character.FighterWithEquipment` 只套用 `Readied` items，第一個有傷害骰的物品作為武器，護甲改善 fighter AC。

這個 adapter 不宣稱完成 DOS party inventory：charges、Affects／法術效果、職業限制驗證、雙持、彈藥消耗、商店與 treasure mutation 都仍待接入。`Fighter()` 與沒有 equipment 的舊 JSON 行為維持不變。

驗證：`internal/monster/equipment_test.go` 覆蓋 damage／packed AC；`internal/party/character_test.go` 覆蓋 readied／非 readied equipment projection；`go test ./...` 應通過。
