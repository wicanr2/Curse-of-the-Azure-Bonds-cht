# 第四十四輪：繁中角色建立 UI

狀態：READY（starter character creation slice）

## 本輪證據

- Ebiten 按 `C` 開啟「建立冒險隊伍」畫面；方向鍵選擇模板，Enter 加入，`D` 完成，Esc 取消。
- 三個 starter templates（人類戰士、人類牧師、精靈魔法師）均先經 `party.Character.Validate`，完成後由 `Character.Fighter` 投影至 `State.SetParty`。
- `State` 會保存建立後的 party，後續 Battle／CAMP 可使用同一 roster。
- State regression 驗證三個模板可完成建立並返回荒野。

## 邊界與未完成項目

- 這是可玩的 starter slice，不是原版完整隨機建角畫面；尚未提供自訂姓名、擲能力值、性別／年齡、alignment、多職／轉職、裝備選擇、XP／等級與存檔匯入。
- `Character.Fighter` 的 HP／AC／傷害是明確標示的 deterministic defaults，不宣稱已取代原版完整角色表。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./...
```
