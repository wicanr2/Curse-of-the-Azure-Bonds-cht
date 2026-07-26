# 第一百二十六輪：CAMP FIX Cure Light Wounds

狀態：`READY`（限目前已驗證的 cleric spell-slot 與固定治療 service boundary）

## 反組譯／手冊證據

繁中遊玩手冊的 CAMP 指令包含 `FIX`；說明是隊伍中具有一級牧師法術的角色，會把已記憶的 Cure Light Wounds 施放在隊伍成員身上，完成後重新記憶原先的法術。原始法術表的第一級牧師欄位順序可確認為 `Bless`、`Curse`、`Cure Light Wounds`，因此目前 remake 將它映射為 one-based spell ID `3`。

這個 ID 是由目前已讀出的法術表順序推導出的窄證據，不代表完整 DOS spell catalog 已解碼；若後續反組譯取得正式 catalog，應由資料層替換，不應讓 UI 或治療 service 依賴 ordinal 猜測。

## 實作 contract

- `State.Fix` 只接受荒野／事件中的 CAMP boundary，掃描 roster 中牧師角色的 memorized `SpellSlots`，每個 ID `3` 產生一次 Cure Light Wounds cast。
- 每次施法依 roster 順序尋找第一位受傷角色，使用 `1d8` 治療並封頂 `MaxHitPoints`；完成後以 stable character ID 同步目前 combat fighter HP。
- spell slot 不會被消耗，符合「重新記憶原先已記憶法術」的目前手冊語意；無可用法術時顯示明確訊息且不改 party。
- `SetFixSeed` 只為 deterministic test／重現提供 seed；原版施法時間、遊戲時間推進與被打斷分支仍未接入。

## 驗證

`TestCampFixUsesMemorizedCureLightWounds` 覆蓋 CAMP 進入、牧師 slot、固定 seed 的治療、roster／fighter HP sync、slot 保留與返回荒野。Docker 內 `go test ./...` 已通過，`git diff --check` 亦通過。
