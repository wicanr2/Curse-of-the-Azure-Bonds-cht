# 458：角色／怪物效果名稱資料分離

狀態：`READY`

## 範圍

DOS `AffectRecord.Kind` 保持 raw typed ID；二十個既有已命名效果與未知效果格式
由 Go 搬入正式 `assets/locale/zh-TW.json`。stable key 採原始十六進位身分：
`affect_kind_XX`，未知格式為 `affect_unknown`。

`monster.LocalizedAffectName` 只依 raw kind 查詢注入的 `TextResolver`。找不到 key
時顯示 raw ID 診斷，不從法術名稱、攻略文字或相似效果猜測規則語意。

## 消費端與測試

- `cmd/azure-bonds -monster-affects` 改用 `-locale` 載入的同一份正式 catalog。
- `internal/monster` 使用合成 resolver 驗證 known／unknown 組字，不複製 CoAB
  中文效果表。
- 遊戲層 coverage 測試載入正式 `zh-TW.json`，逐一驗證二十個 named kind
  keys 與未知格式 key 均存在。
- 真實 DOS `MON6SPC.DAX` block 67 smoke 證明 `18h` 由 locale 顯示名稱；
  `82h／81h／3Ch` 尚無語意證據，會保留 raw ID 診斷而不猜譯。

## 可量測結果

Go 漢字 literal exact baseline 由 974 降為 953 occurrences，共移除 21 次：

| 分類 | 第 457 輪 | 本輪 | 差異 |
|---|---:|---:|---:|
| 本地化／劇情 fallback | 372 | 372 | 0 |
| Ebiten 前端 UI | 164 | 164 | 0 |
| runtime／工具 UI | 438 | 417 | -21 |
| 合計 | 974 | 953 | -21 |

## 明確未完成

名稱資料化不等於二十種已命名效果的完整規則、生命週期、動畫、音效、存讀檔
與原版玩家路徑皆已完成。未命名 kind 仍必須由 bytes、writer／consumer 與 runtime
證據確認後，才可新增 stable key 或規則語意。
