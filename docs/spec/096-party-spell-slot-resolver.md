# 第九十六輪：party spell-slot resolver

狀態：`READY`（限 remake roster spell slot model 與 ECL signal bridge）

`party.Character.SpellSlots` 保存 data-neutral 的 ordered spell IDs；`party.Roster.FindSpell` 依 party marching order、再依 character slot 順序回傳第一個 `SpellMatch`。`game.State.ResolveSpellSearch` 已接收 ECL `SpellSearch` 並查目前已載入的 remake roster。

這不是原始 DOS player record parser：SpellSlots 是可替換的 intermediate model，尚未映射原始 spell-known／memorized offset，也尚未把 slot／character index 寫回 ECL resumable memory。沒有 spell slots 的舊 JSON 仍合法。

同一輪也將 `ITEMS` catalog 載入 Ebiten game state；character creation／party load 若有 equipment，會使用 `FighterWithEquipment` 投影基本 readied 武器／護甲效果。

驗證：`internal/party/character_test.go` 覆蓋 first-match resolver；`internal/game/state_test.go` 覆蓋 ECL resolver bridge 與 ITEMS catalog → creation fighter；`go test ./...` 應通過。
