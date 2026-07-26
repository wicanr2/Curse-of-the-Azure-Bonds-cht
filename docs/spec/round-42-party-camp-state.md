# 第四十二輪：party 保存與 CAMP 恢復 state

狀態：READY（party HP state boundary；不宣稱完整角色／休息規則）

## 本輪證據

- `State.SetParty` 保存具備穩定 ID 的 party fighter roster，拒絕空 party、錯誤 side 與重複 ID。
- `StartCombat` 會保存 party；戰鬥結束時同步 Battle 中的 party HP 回 State。
- `Camp` 對已保存 party 恢復 HP 至 MaxHP；沒有 party 時仍只顯示紮營事件。
- regression 覆蓋 party 保存與 CAMP 後 HP 恢復。

## 邊界與未完成項目

- 原始 character creation／Pool of Radiance import、能力值、裝備欄、法術欄、經驗與存檔格式尚未完成。
- 原始 CAMP 的被打斷、守夜、法術恢復、毒／疾病與 ECL memory effects 尚未完成；本輪只建立可替換的 HP state boundary。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./...
```
