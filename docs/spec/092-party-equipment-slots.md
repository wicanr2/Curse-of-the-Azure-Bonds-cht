# 第九十二輪：party equipment slot contract

狀態：`READY`（限 base ITEMS class mask／slot transaction）

`party.ItemClassBit` 將本地 class enum 映射到 reference ITEMS usability bits：magic-user `1`、cleric `2`、thief `4`、fighter `8`、paladin `64`、ranger `128`。`Character.CanEquip`／`EquipItem`／`UnequipItem` 以 descriptor slot 驗證：主手、 副手、身體等一般 slot 不可重複；雙手武器與副手衝突；戒指 slot 最多兩個；slot 10 ammunition 不加入一般裝備 slot collision。

這是可重用的 party data contract，不是完整 DOS inventory：尚未處理 stack quantity、彈藥消耗、cursed lock、shop transaction、magic effects、druid/monk player class 與原始存檔 offset。沒有 `equipment` 的舊 JSON 不受影響。

驗證：`internal/party/character_test.go` 覆蓋 class rejection、雙手／副手衝突、一般 slot 與雙戒指上限；`go test ./...` 應通過。
