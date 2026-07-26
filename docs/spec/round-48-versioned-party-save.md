# 第四十八輪：版本化 party JSON 存檔

狀態：READY（remake party descriptor save slice；不是原版 DOS save）

## 已完成

- 新增 `internal/save`，以 version `1` 的 JSON 保存 `party.Roster`。
- 編碼前與解碼後都執行角色、能力值、種族／職業限制及 1–6 人 roster validation。
- `game.State.SavePartyFile`／`LoadPartyFile` 將角色資料投影回目前可玩的 combat party。
- Ebiten 接通 `F5` 儲存、`F9` 載入；預設路徑為 `party.json`，亦可用 `-party-save` 指定路徑，或用 `-party-load` 啟動時載入。
- regression 覆蓋 JSON round trip、中文姓名與未知版本拒絕；game test 覆蓋儲存後載入回 State。

## 明確邊界

- 這是 remake 自有的角色描述檔，不宣稱已解碼原版 DOS save/import 格式。
- 目前不保存戰鬥中的當前 HP、地圖座標、事件 cursor、物品、法術、XP 或完整 CAMP 中斷狀態；載入後從隊伍建立完成的荒野入口開始。
- `-encounter` 的 debug party 仍是明確的測試入口，不能直接以 F5 保存成角色 roster。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./internal/save ./internal/game ./internal/party ./cmd/azure-bonds-game
```
