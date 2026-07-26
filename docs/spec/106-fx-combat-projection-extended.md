# 第 106 輪：DOS `.FX` Blind／Bestow Curse／Prayer combat projection

狀態：`READY`

## 目的

把目前能在單一 party fighter 上無歧義套用的 active `.FX` modifiers 接到
remake 的 fighter projection；需要 target、alignment 或 action context 的效果
仍保留在 effect record，不在 importer 階段猜測。

## 已確認規則

依 [Curse of the Azure Bonds rules reference](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)：

- `0x21` Blind：攻擊修正 `-4`，AC `+4`（AC 數值增加代表防禦變差）。
- `0x24` Bestow Curse：攻擊修正 `-4`。
- `0x31` Prayer：對 party 的 friendly effect 為攻擊修正 `+1`。

這些 projection 只處理 `Active=true` 的 effect。hostile Prayer、saving throw、
敵方 target、Haste 的額外攻擊次數、Protection／Mirror Image 的命中或免疫規則，
仍需完整 combat action／target layer 後才能接入。

## 驗證

`party.Character.Fighter` 有回歸測試確認三種 effect 可疊加，inactive effect 不會
改變 fighter。`.FX` 的 duration tick 仍由 `monster.AdvanceAffects` 負責。
