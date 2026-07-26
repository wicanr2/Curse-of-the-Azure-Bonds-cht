# Combat team start layout（READY）

## 證據

反組譯 `ovr011.PlaceCombatants` 與 `try_place_combatant` 可確認：party team start 是 `(0, 0)`；enemy team start 的 x/y 分別是 `encounter_distance` 乘以 `MapDirectionXDelta/YDelta[map_direction]`。party direction group 為 `map_direction / 2`，enemy direction group 為 `((map_direction + 4) % 8) / 2`。

`MapDirectionDelta` 的八方向順序是 north、northeast、east、southeast、south、southwest、west、northwest。`SETUP MONSTER` 的第二個 operand 已由 `ecl.MonsterSetup.MaxDistance` 保存，但目前 `map_direction` 尚未由 Area／Player record 解碼。

## 實作邊界

`combat.EncounterTeamStart` 只接受 encounter distance、map direction 與 team side，回傳 team origin 和四方向 facing group。它不選 candidate row/column，也不處理 `unk_1AB1C` occupancy、ground collision 或 camera transform；這些仍需獨立解碼後才能接入 `StartEncounter`。

## 驗證

- `internal/combat/placement_test.go` 覆蓋 party origin、enemy offset、opposite facing 與非法輸入。
- `go test ./...` 應通過。
