# 第四十三輪：角色建立規則核心

狀態：READY（角色建立 validation；尚未接完整 UI／存檔）

## 本輪證據

- 新增 `internal/party`：六個玩家種族（矮人、精靈、侏儒、半精靈、半身人、人類）與六個基本職業（牧師、戰士、遊俠、聖武士、魔法師、盜賊）。
- `Character.Validate` 驗證 ID／姓名、等級、能力值 3–18、種族／職業限制與職業最低能力值。
- `Roster.Validate` 驗證 party 人數 1–6 與 character ID 唯一性。
- 規則交叉核對本地 RuleBook 轉錄與公開 CoAB reference rewrite；本輪只保存規則 contract，不複製參考程式碼。

## 邊界與未完成項目

- 多職／人類轉職、性別／年齡修正、alignment、XP／等級表、法術欄、裝備與角色存檔尚未完成。
- `internal/party` 尚未由 Ebiten 建立角色流程；目前 `State.SetParty` 仍接收已轉換的 combat fighters。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./internal/party
```
