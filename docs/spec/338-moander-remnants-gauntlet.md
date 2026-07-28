# 第三百三十八輪：摩安德殘軀與摩安德護手

狀態：`READY`

## 原版資料證據

ECL3 block `0x12` 在摩貢首戰勝利後保存 continuation：

- 異次元裂隙閉合；
- 三塊已穿過裂隙的 Moander pseudopods 長出數百張嘴，尖叫
  `YOU HAVE KILLED ME!`；
- `THE OOZING MOUNDS TURN AND ATTACK YOU!` 後建立第二場戰鬥；
- MON3 block `0x1A` 的 canonical name 是 `BIT O' MOANDER`；
- 原始 encounter 產生三隻，每隻 140 HP、combat size code `4`（2×2）；
- 勝利後顯示 `YOU FIND THE GAUNTLET OF MOANDER IN THE SLIMY REMAINS`；
- 下一個 continuation 將 plot byte `4C5B` 寫為 `1`；
- 一名祭司衝入房間並高喊眾人殺了神。

## 引擎與中文契約

- 摩貢首戰勝利必須 resume 同一 ECL runtime state，不可回到 SearchLocation
  重新觸發祭壇。
- post-combat ECL budget 必須容納建立第二 encounter 前的連續 script；
  原本 180 steps 會在 payload `+0x1BA6` 附近錯誤中止，本輪使用與 dungeon
  lifecycle 相同的 bounded 500-step budget。
- `BIT O' MOANDER` 顯示為「摩安德殘軀」，但保留 MON3 `0x1A`、140 HP、
  2×2 footprint 與原始 affects。
- 護手顯示為「摩安德護手」，並以 `4C5B=1` 驗證劇情神器狀態；
  不虛構一個原版未加入普通 inventory 的背包物件。

## 可沿用的 Gold Box 知識

engine boundary 後到下一個玩家可見 boundary 之間可能包含大量 setup、plot
writes、monster build 與第二次 COMBAT。短 menu resume budget 不足以涵蓋
長 post-combat script；應使用明確 bounded、但與 dungeon lifecycle 同級的
budget，並以真實最長路徑 regression 校準。

劇情神器未必是 inventory item。若 ECL 只寫 plot byte，remake 應保存該 byte
及作品層顯示名稱，不應為了 UI 一致性自行創造可丟棄、可交易的普通物品。

## 驗收

- 同一 real-session 擊敗摩貢首戰，抵達裂隙閉合畫面。
- 驗證三塊殘軀尖叫與第二場戰鬥繁中。
- 驗證三隻 MON3 `0x1A`，各 140 HP、CombatSize `4`。
- 擊敗第二戰後顯示「找到摩安德護手」。
- continuation 後驗證 session memory `4C5B=1`。
- 驗證祭司高喊「他們殺了神」的後續繁中。
