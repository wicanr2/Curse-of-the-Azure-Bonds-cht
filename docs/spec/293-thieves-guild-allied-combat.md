# 第二百九十三輪：盜賊公會混合陣營戰鬥

狀態：`READY`

## 反組譯證據

ECL2 block 2 `+0x046B..+0x04BC` 先把迴圈次數設為 4、TeamList selector
設為 10。每輪執行：

1. `LOAD CHARACTER [0x7F79]`；
2. 讀 selected-player window `0x7D00`（Player `in_combat @ +0x100`）確認角色存在；
3. `SAVE 0xB2 → 0x7CB8`，設定 control morale；
4. `SAVE 0x80 → 0x7D0C`，設定 `CombatTeam.Ours + QuickFight`；
5. selector 加一，直到四輪完成才執行 `COMBAT`。

公開 CoAB reference 的 `vm_GetMemoryValue／alter_character` 證實 `+0x100` 是
computed `in_combat`，`+0x10C` 的 `0x80` 是我方 quick-fight；不能把這些位址當
一般零值 VM memory。GameFAQs 的逐格資料與真實 ECL 結果一致：玩家隊伍由 4 名
THIEF 協助，敵方是 2 FIRE KNIFE 加 11 THIEF。

## 實作

- bounded VM 在 `LOAD CHARACTER` 後投影 selected TeamList slot 的 `in_combat`；
- `RunResult.CombatTeamWrites` 保存跨 pause 的 selected-player team write；
- `MonsterSpawn.PartyMask` 把單一 `LOAD MONSTER count=15` 中的四個 copy 改為我方；
- `combat.Fighter.QuickFight` 讓友軍保留 Party side，但由 AI 自動攻擊 Enemy side；
- ECL 勝利 continuation 顯示公會首領遺言並解鎖 Adventure Journal Entry 4；
- Entry 4 原文只有「Guild → Hideout」下水道圖，因此繁中手札保存忠實圖像描述，
  不虛構路線文字。

## 驗證

真實 image regression 從新建角色依序走過 Weaponers、Filani、皇家馬車、皇家衛兵、
投降、牢房與盜賊救援，再進 ECL2 block 2 `(1,12,0)`。測試要求戰場恰為
`hero + 4 allied THIEF` 對 `2 FIRE KNIFE + 11 THIEF`，實際勝利後出現繁中遺言、
下水道地圖說明與手札第 4 條。

Ebiten 新增 `-guildmaster`，由相同正式 ECL 路徑建立可重現的 640×480 混合陣營
戰鬥畫面；不直接拼裝假的 encounter。
