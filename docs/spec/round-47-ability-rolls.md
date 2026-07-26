# 第四十七輪：能力值 3d6 擲骰與重擲

狀態：READY（可重現能力值擲骰 slice）

## 本輪證據

- `party.RollAbilities(seed)` 為六項能力各擲一次 3d6，並以 seed 保持測試／重播可重現。
- 角色建立畫面按 `R` 重擲目前模板；重擲後仍須通過 `Character.Validate` 的 3–18 與職業最低值檢查。
- State test 驗證相同 seed 產生相同結果；party test 驗證所有結果落在 3–18。

## 邊界與未完成項目

- 尚未重製原版的點數池／重擲確認 UX、性別／年齡修正、exceptional strength、alignment、多職／轉職與存檔。
- UI 使用時間 seed 產生玩家重擲結果；核心 API 仍接受明確 seed 供 regression。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./...
```
