# 第四十五輪：繁中角色姓名輸入

狀態：READY（Unicode 姓名輸入 slice）

## 本輪證據

- 角色建立畫面按 `N` 開啟姓名編輯；使用 Ebiten `InputChars` 接收 Unicode／繁中文字元。
- Backspace 刪除最後一個 rune，Enter 確認，Esc 取消；姓名限制 20 個字元且不可空白。
- 確認後模板角色名稱會寫回 `CreationOptions`，再由既有 `FinishCharacterCreation` 投影到 party。
- State regression 驗證繁中姓名「阿勇」可輸入並保存。

## 邊界與未完成項目

- 尚未接能力值逐項分配／重擲、性別／年齡、alignment、裝備、XP／等級與存檔。
- `InputChars` 的實際輸入法行為依作業系統／字型環境；核心 state 以 rune API 可重現測試。

## 驗證

```text
CGO_ENABLED=1 go test -vet=off ./...
```
