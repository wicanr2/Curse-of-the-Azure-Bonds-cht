# 第二百八十六輪：提爾佛頓訓練場

狀態：`READY`

## 證據

- ECL2 block 1 的 GEO terrain `0x8C` 位於 `(5,2)`；事件顯示 PICTURE 4、詢問
  `DO YOU WANT TO TRAIN?`，YES 分支呼叫 `PROGRAM 0`。
- reference `ovr018.train_player` 從選定角色收取 1000 GP，依 class experience
  table 選出可升級職業。DOS `Player.exp` 是 `0x127` 的 little-endian dword；
  Tilverton 的 Area2 training mask 是 `0xFF`。
- HP 使用職業 hit die 擲兩次取高；ranger 第一次升級擲 `2d8`，多職業除以職業數，
  並套用 Constitution bonus。

## 本輪交易

`PROGRAM 0` 只有在 dungeon return context 且 terrain 為 `0x8C` 時映射成訓練場；
其他 context 維持既有返回主選單語意。訓練選單檢查健康、角色自身金錢、經驗門檻，
確認後扣除 1000 GP、提升對應 class level、增加 current/max HP，並保存經驗值。

本輪涵蓋 cleric、fighter、paladin、ranger、magic-user、thief 的 reference 門檻。
高等級固定 HP、多職業完整 Constitution 算法、種族等級上限，以及 magic-user／ranger
升級選法術仍是明確未完成邊界。
