# 第九十四輪：consumable item use signal

狀態：`READY`（限卷軸／藥水／魔杖 property decode 與 inventory mutation）

依 item format 的 `Affects` properties：base type 60–62 解為 scroll，收集三個 spell IDs；71／84 解為一次性 potion；78／79 解為 charged wand，`Affects[0]` 是 charge、`Affects[1]` 是 effect ID。`ItemRecord.DecodeConsumable` 只產生 `ConsumableUse` data signal；`Character.UseConsumable` 會移除一個 scroll／potion unit，或將魔杖 charge 減一，不直接施放 spell/effect。

沒有 charges 的 wand 會安全失敗；readied wand 可以使用，scroll／potion 仍遵守 readied removal protection。這輪尚未處理 spell targeting、法術效果、魔杖空後是否保留、彈藥、商店與完整 DOS inventory semantics。

驗證：`internal/monster/equipment_test.go` 覆蓋三類 property decode；`internal/party/character_test.go` 覆蓋 scroll stack 與 wand charge mutation；`go test ./...` 應通過。
