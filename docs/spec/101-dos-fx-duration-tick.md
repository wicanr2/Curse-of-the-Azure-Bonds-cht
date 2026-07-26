# 第一百零一輪：DOS `.FX` duration／strength fidelity

狀態：`READY`（限欄位修正與 duration tick，不含 effect gameplay）

公開 effects list 將 9-byte record 的欄位定義為：第一 byte effect kind、第二／三 byte little-endian remaining minutes、第四 byte strength（`255` 表示 permanent），第五 byte 以後是 effect-specific data。原先 parser 把第四 byte 當 duration，本輪已修正為 `Duration uint16`／`Strength uint8`，並保留 `Value` 作為 raw duration 相容欄位。

`monster.AdvanceAffects` 以分鐘消耗 finite effects；duration 小於等於消耗量的 effect 移除，permanent strength `255` 保留，其餘 raw data 不變。`party.Character.AdvanceEffects` 與 `game.State.AdvancePartyEffects` 提供 party model/state adapter。

本輪不把 kind／strength 自動映射成 AC、攻擊、HP 或狀態，因為 effect-specific rules、combat tick 邊界與 `.FX` slot lifecycle 尚未完整反組。CAMP／戰鬥上層可在確認時間語意後呼叫 adapter。

格式參考：[Curse of the Azure Bonds creature effects list](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)。
