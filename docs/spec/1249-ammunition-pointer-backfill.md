# 1249：彈藥指標跨規格回填

狀態：`CONFORMED`（2026-08-30）

## 問題與證據

spec 1000 已以 DOS `overlay-24:00C28h`（391 條逐條讀完，PC-98 對側相似度
0.988）證明角色派生值重算會走完整條物品鏈：

- 物品類別 `49h`（Arrow）寫入角色 `+17Dh`；
- 物品類別 `1Ch`（Quarrel）寫入角色 `+181h`；
- 兩個判斷位於 `q^[34h] <> 0`（已裝備）區塊之外，因此不要求彈藥為 Readied；
- 同類型有多個節點時，每次遇到都覆寫，最後保留鏈上最後一個節點。

spec 1120 後來把「`+151h` 起每四 bytes 一格」誤推成 ITEMS 類別表的槽
11／12，spec 1186 雖證明槽 11／12 實際是卷軸，卻又寫成「正確判斷尚未解出」。
這兩句都漏掉了 spec 1000 已存在的 exact producer。

## READY 行為契約

1. 掃角色／怪物的物品鏈，以類別 `49h`／`1Ch` 建立彈藥 A／B 是否存在；不看
   Readied，也不依 ITEMS 的 `Slot` 欄位。
2. 玩家戰鬥員投影先保存兩類彈藥各自在鏈上最後一疊的原始 `Count`。
3. 已選武器的 ITEMS `AmmunitionType` 以 bit 0 選 `49h`，bit 7 選 `1Ch`；兩者
   同時成立時照原作連續 `if` 的順序由 bit 7 覆蓋。
4. `Count == 0` 保留為 0；`capByAmmunition` 的既有契約會把它視為不設限，對應
   原作只有 `q^[39h] > 0` 才壓射擊次數。
5. 投擲武器、自足武器、實際扣除彈藥與 game-pack 的 raw ammunition-code 映射
   不在本次改動範圍。

## 玩家垂直鏈與驗收

`ITEM*` 物品鏈 → `Character.Equipment`／`MonsterItems` → `+17Dh/+181h` typed
投影 → AI／QUICK 遠程選擇 → `Fighter.AmmunitionCount` → 每回合攻擊次數上限。

回歸必須同時證明：未 Readied 的 `49h` 仍可供弓使用；槽 11／12 的卷軸不能冒充
彈藥；`1Ch` 只供 bit 7 武器；同類多疊採最後一疊；AI／QUICK 有箭才選弓。

## 存檔與權利邊界

不新增存檔欄位；彈藥指標仍由載入後的物品鏈重建。本規格只記錄控制流與欄位語意，
不新增或散布原版資料。

## 驗證收據

- `internal/party` 全套通過，包含 Arrow／Quarrel 分離、鏈上最後一疊與卷軸負對照。
- `internal/game` 全套通過（290.190 秒）；AI／QUICK 聚焦測試另通過。
- `tools/re-resolution-backlinks.sh` 納入 spec 1120／1186 的兩筆勘誤。
