# 第四十六輪：角色建立能力值編輯

狀態：READY（能力值編輯 slice）

## 本輪證據

- 角色建立按 `A` 進入能力值編輯；左右切換力量、智力、智慧、敏捷、體質、魅力，上下調整。
- `party.Abilities.Adjust` 將每項能力限制在 3–18；狀態更新後仍由 `Character.Validate` 檢查職業最低值。
- Enter 可從能力值編輯直接加入目前角色，完成流程仍由 `FinishCharacterCreation` 統一投影到 party。
- State／party tests 覆蓋能力值調整、上下界與能力游標。

## 邊界與未完成項目

- 尚未實作原版擲骰／重擲點數池、性別／年齡修正、exceptional strength、alignment、多職／轉職與裝備選擇。
- 目前模板的能力值是 starter defaults；尚未保存角色能力值到檔案格式。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./...
```
