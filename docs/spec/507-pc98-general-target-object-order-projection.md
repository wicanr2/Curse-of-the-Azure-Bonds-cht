# 第五百零七輪：PC-98 一般敵方選敵物件順序投影

狀態：`READY`（限一般 enemy target consumer 的排序投影；不宣稱完整可達清單、
visibility、牆面、AI priority、原版亂數或完整 monster turn）

## 本輪結論

第 231 輪已建立 `SelectCombatTarget` 的 bounded 邊界：在完整
`Rebuild_SortedCombatantList` 尚未還原前，只從存活且屬於指定對立陣營的 fighter
建立 deterministic 候選，再消耗 Battle PRNG。第 506 輪已證明 ranged consumer
可在同距候選使用完整 `LegacyObjectID`；本輪把相同的 identity projection 套到
一般敵方物理選敵，避免目前 stable ID 字典序把原始 combat-object 順序覆蓋掉。

這只修正候選排序，不擴張候選集合。沒有完整 legacy identity 的候選仍以 stable
fighter ID 作 fallback，因此不會把部分投影誤當作完整原版 `OBJECTLIST`。

## 證據與邊界

- 本輪延伸 audit 的輸入為 `workplace/ida406/pc98-overlays/overlay-24.bin`，
  SHA-256 `0b5d1bebb367414fac93287449a230a36355c075efdc30640a4994bc11d5cd5f`；
  IDA script `scripts/ida/pc98_ranged_target_producer_audit.idc` SHA-256
  `5218dac29a2e8e07d2c75f15374ab726afc93ad0c47bd118dc7b5eb005968d4d`；報告
  `pc98-ranged-target-producer.txt` SHA-256
  `93ea534d30637af7e67abf3ed15748c91ea71dd87ebd7abd143f45aa6cc74cd9`。
  使用 IDA Pro 9.4 image `ida-pro-9.4-ver2:uidfix-v1`，Docker 副本位址空間
  `min=0x0000..max=0x3436`，連續輸出 `overlay-local 0x2820..0x2C80`。
- `find_target → BuildNearTargets → Rebuild_SortedCombatantList` 的原始呼叫形狀與
  「最多 20 次抽樣」仍以第 231 輪的來源／報告為準；其可達、visibility、牆面與
  spell priority 尚未閉合。
- PC-98 `PICKTARGET` 的候選抽樣／失敗移除／20 次上限與
  `LegacyObjectID` identity projection 仍以第 416、505、506 輪的 raw bytes、
  IDA 報告與 adapter contract 為準。
- 本輪新增的非破壞性 IDA 延伸工具
  `scripts/ida/pc98_ranged_target_producer_audit.idc` 只把 overlay 24
  `2820h..2C80h` 連續 bytes 匯出到報告；它沒有替未知欄位命名，也沒有因此把
  producer／tie／RNG 升格為 `exact`。

所有 overlay-local 位址、raw far pointer 與 resident segment 必須分開記錄。ID
或相同十六進位數字不能單獨證明欄位語意；仍須閉合 writer → candidate projection
→ consumer／runtime 的資料流。

## Remake 契約

`Battle.SelectCombatTarget` 現在：

1. 保留既有「存活、指定 side」bounded 候選集合與 no-target error。
2. 候選均有不同非零 `LegacyObjectID` 時依一基底 combat-object 身分排序。
3. identity projection 不完整時依 stable fighter ID 排序。
4. 只執行原有一次 Battle PRNG 抽樣，不因排序修正增加 draw。

原始可達／visibility 清單一旦由新證據閉合，必須在 adapter 層擴充候選 producer，
不能把本輪 stable fallback 或全存活候選宣稱為完整原作規則。

## 回歸驗證

`TestSelectCombatTargetUsesLegacyObjectOrderForEqualCandidates` 使用 stable ID 與
legacy object order 相反的兩個存活目標，以相同 seed 計算一次抽樣預期，驗證一般
enemy target consumer 使用 legacy order 且沒有額外 PRNG 消耗。既有不同 seed、
同 seed replay 與 multi-attack target continuity 測試仍保留。

## 尚未完成

- `BuildNearTargets`／`Rebuild_SortedCombatantList` 的完整 producer、range／LOS、
  visibility、20 次重抽與 action target persistence。
- enemy Quick／Guard／Bandage／flee AI、全部法術與特殊能力、逐幀戰鬥演出、音效、
  全作玩家路徑與完整中文化。
- 因此本輪仍不能宣稱完整戰鬥、完整可通關或整作 remake 完成。
