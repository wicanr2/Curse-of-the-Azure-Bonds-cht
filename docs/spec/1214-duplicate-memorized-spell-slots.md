# 1214 — 同一法術可重複記憶並合併顯示次數

狀態：`READY`（目前 CAMP 的一環 preparation adapter；高環分組仍依既有邊界）

## 原版契約

- spec 923 已確認角色記錄 `+1Eh..+71h` 是 84 格 ordered memorized spell
  slots，不是每個法術一個布林旗標。
- spec 859 逐指令證實，清單建構器遇到相同 spell ID 會增加次數，而不是丟棄
  duplicate；顯示文字附加 ` (N)`。
- 第五章阿卡巴的實際快照是 `[15 15 15 15 34 34 47]`，提供四格 Magic
  Missile 與兩格 Stinking Cloud 的正常玩家正對照。

## 舊差異

remake 的 MEMORIZE 把每一列當成切換開關：第一次加入，第二次移除。同一法術
最多只能佔一格。因此五級法師即使有四個一環容量，法術書裡只有一支已完成戰鬥
規則時也只能準備一格；這與原版角色記錄及清單顯示都不相容。

## 修正

- 每次選取法術都 append 一個 ordered pending slot，直到該角色的一環容量已滿。
- 相同 spell ID 在 UI 保持一列；兩格以上顯示 `法術名 (N)`，一格仍以星號標示。
- 超過容量只顯示既有容量訊息，不修改 pending slots。
- 「取消此角色」仍清空這名角色整份 pending loadout，提供可逆的重選路徑；REST
  完成前不寫入正式 `SpellSlots`。

## 驗證

- `TestCampMagicMemorizeAllowsDuplicateSlotsAndShowsCount`：同一支燃燒之手連選四次，
  鎖定四個 ordered slots、`(4)` 顯示與第五次容量拒絕。
- `TestCampMagicMemorizeAppliesAtRest`：既有不同法術選擇與 REST writeback 不退化。
- 一般強度按鍵路徑以正式 MEMORIZE 產生「燃燒之手 (4)」，沒有直接改 roster。

## 未完成 gate

四格燃燒之手與五格牧師戰鬥 loadout 仍未讓自動 QUICK 控制組通過八名火刀連戰；
1000 幀報表最終六人 HP 皆為 0。這不影響 duplicate-slot 契約成立，也不能據此
宣稱一般強度通關。
