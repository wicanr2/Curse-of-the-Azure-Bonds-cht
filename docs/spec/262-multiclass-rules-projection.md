# 第二百六十二輪：multi-class rules projection

狀態：`READY`（限目前已驗證的裝備／一級法術職業判定）

## 證據與結論

DOS player record 與角色建立資料已保存 `ClassLevels[8]`；因此不能再只用
`Character.Class` 判斷一個角色是否具有某職業。`party.Character.HasClass` 現在會先查
非零的 raw class-level slots，並在舊 JSON 或沒有 class-level metadata 時退回 primary
class。這讓舊存檔保持相容，也讓新匯入的多職業角色保留能力。

目前已接入的規則投影：

- `CanEquip` 會逐一檢查角色擁有的職業與 ITEMS class mask；例如戰士／法師可使用兩邊
  都允許的裝備。
- CAMP MAGIC 與 combat spell gate 會依 `HasClass` 判定 cleric／magic-user；牧師／戰士
  不會因 primary class projection 而失去牧師法術。
- Protection from Good 的 combat spell identity 也依實際 caster 的 cleric slot 判定。

## 明確邊界

THAC0、生命骰、等級顯示、部分 UI spell label 與尚未解出的 DOS multi-class serializer
仍使用 primary-class adapter。這些欄位需要逐一對照 reference 的 class-level lookup，
本輪不以推測補齊。`ClassLevels` 的 raw slot mapping 與 `RawClassID` 仍由
[`259-dos-multiclass-levels.md`](./259-dos-multiclass-levels.md) 定義。

## 驗證

```text
go test ./internal/party ./internal/game ./internal/ecl
```

測試覆蓋既有單職業行為與 DOS multi-class metadata；後續應加入含實際 ITEMS mask 的
fighter／magic-user equipment fixture，以及各職業等級對應的完整法術容量表。
