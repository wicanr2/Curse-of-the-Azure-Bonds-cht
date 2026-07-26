# 第一百一十九輪：CAMP MAGIC spell-slot view

狀態：`READY`（限已驗證 spell-slot 的只讀查看 boundary）

## 實作 contract

`CAMP → MAGIC` 先列出 party roster 的角色與已記憶法術數量；選取角色後顯示其 `Character.SpellSlots` 中保存的 byte-sized spell IDs（十六進位），不修改 spell slots。Enter 返回角色列表，返回項回到 CAMP Menu。

這一輪只重用 DOS player spell record 已確認的 memorized slot 資料與 `Roster.FindSpell` 的 ordered slot contract；法術名稱、準備／遺忘 mutation、施法消耗、恢復時機與 saving throw 仍由後續 rules／spell catalog 提供。

## 驗證

`TestCampMenuMagicListsMemorizedSlots` 覆蓋角色列表、兩個已記憶 spell ID 的摘要、返回列表與離開 MAGIC。`go test ./...` 已通過。
