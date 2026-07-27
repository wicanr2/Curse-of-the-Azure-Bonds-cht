# 第 257 輪：TREASURE random 與 party pickup

狀態：`READY`（限 reference random table、seeded loot 與目前事件選單）

## 已完成

- `ITEM{area}.DAX` deterministic block 與 `0xFF` no-item branch 均可解析。
- `0x80+n` 依 reference `CMD_Treasure` 的 d100 table 產生 n 件 raw item type；State 使用 `eclSeed` 保持可重播。
- Ebiten／State 事件 menu 會列出 pending loot，玩家先選物品，再選由哪位角色收下；未收下的物品不會被偷偷丟給第一位角色。
- 新增 regression：random count、金錢／寶石／珠寶、兩 area block namespace 與明確 party pickup。

## 保留邊界

完整 `ItemDisplayNameBuild` name-number／魔法 identify semantics、原版 item capacity／重量限制，以及所有真實劇情入口仍待接入。若 headless adapter 尚未載入 ITEM DAX，raw TREASURE request 會保留並讓後續 ECL control flow（例如 COMBAT）繼續。
