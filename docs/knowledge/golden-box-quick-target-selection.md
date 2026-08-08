# Golden Box Quick 目標選擇知識

Quick 施法必須分成兩個階段：法術 suitability 只回答「是否有合法目標」；
法術選定後，target helper 才依候選鏈、優先級與亂數把 opaque target 交給
施法 action。把兩者合併成一次 `first target` 查詢，容易在每次候選試驗時多吃
一段 PRNG，導致後面的法術、命中與遭遇結果漂移。

要另外區分 caller：PC-98 overlay 09 的 `04CCh` 是 entry 4 的公開 handler，
但已找到的唯一直接 caller 是通用 action dispatcher `0164h`。因此它可支持
「某類 action 的候選鏈／priority 邊界」，不能單獨證明所有 Quick 法術都走同一
入口。overlay 13 `PICKTARGET` 的 `3D7Fh..3F5Ch` 則是另一個 target consumer：
它先檢查既有 target，再抽取候選、呼叫 `CHECKTARGET`，失敗候選移除後最多
重試 `14h` 次。完整候選 producer 與 target range 仍由作品 adapter 關閉。

目前 PC-98 CoAB 的可重用邊界是：候選以一基底 `LegacyObjectID` 排序，先從
priority `7` 開始，以 `1..7` 的重試數量限制最多掃幾個 threshold；每個
threshold 掃到第一個合法候選就交接。這是 `04CCh..0624h` 的控制流重建，
不是完整 pointer chain 或 random helper 的逐指令還原。

對不需要 priority retry 的 target consumer，可使用 engine
`combat/quicktarget.SelectOne`：它只保留 legacy order 後做一次抽樣，不得在
caller 內再重抽一遍。CoAB Quick Magic Missile 已使用這個 bounded adapter；
手動 target、候選 producer、visibility／range 與 tie 仍不應搬進 engine。

對 ranged target consumer，CoAB adapter 可在距離排序後使用已完整投影的
`LegacyObjectID` 作同距 tie-break；只要任一候選缺少原始身分，就回到 stable ID，
維持 deterministic 且不把不完整資料偽裝成原版順序。這是第 506 輪的
`strong inference` 邊界，不代表 `PICKTARGET` 的完整 producer 或 comparator 已
還原。

實作時應遵守：

- 候選建立、visibility、距離、地形、陣營與法術效果留在 CoAB adapter；
- engine 只接 stable candidate、legacy identity、priority threshold 與 callback；
- game pack 宣告面數與 priority 起點，測試以 stable ID 取資料，不複製中文顯示名；
- no-target 是有界的正常結果，不能偷偷改成另一個法術，也不能消耗法術格；
- 只有取得新 bytes／runtime trace 才能把 `strong inference` 升成 `exact`。

這套分層可沿用到《Death Knights of Krynn》、《Pools of Darkness》與其他 SSI
作品；每一款仍須重新證明自身 record、候選 producer、RNG 與 handoff。
