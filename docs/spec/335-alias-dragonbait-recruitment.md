# 第三百三十五輪：愛麗雅絲與龍餌入隊

狀態：`READY`

## 原版資料證據

摩安德之坑第一層 GEO3 block `0x11` 的 `(1,4)` 為 terrain `0x85`。
ECL3 block `0x11` SearchLocation dispatch 到 `+0x09DC`：

- `4C2E != 0` 時直接 EXIT，故事件只完成一次；
- `4C5B` 是後續摩貢狀態 gate，本輪不替它命名其他語意；
- PICTURE 18 顯示一名女戰士與外貌奇特的蜥蜴人；
- 女戰士看見枷印後驚呼隊伍也受控制，接著進入原版 encounter menu；
- PARLAY → NICE 可和平交談；
- 她自稱 ALIAS，介紹 DRAGONBAIT，並要求隊伍說明經歷；
- `TELL HER YOUR STORY` 解鎖 Journal Entry 3；
- YES 分支發出 `ADD NPC 0x16` 與 `ADD NPC 0x17`，兩人正式入隊。

MON3 player records 證實：

- `0x16 ALIAS`：human fighter level 6，morale `0xB2`；
- `0x17 DRAGONBAIT`：raw race `0x00`、paladin level 7，morale `0xB2`。

Dragonbait 是 saurial；raw race `0x00` 在一般 monster records 也表示非玩家種族，
因此只在 NPC parser 且名稱為 DRAGONBAIT 時投影為 `RaceSaurial`，不可全域把
所有 raw `0x00` 怪物改成 saurial。

## 引擎與中文契約

- 保留 PICTURE 18、encounter menu、五態度 parlay 與三項回答的原始順序。
- 本輪主路線固定 PARLAY → NICE → TELL HER YOUR STORY → YES。
- Alias 顯示名為「愛麗雅絲」，Dragonbait 為「龍餌」；ScriptName 保留原文，
  供 ECL name lookup 使用。
- Journal Entry 3 只在坦白故事後解鎖，以三頁繁中摘要呈現。
- `ADD NPC` 必須載入 MON3 CHA／SPC／ITM，保留原始能力、裝備、effects、
  icon 與 morale。
- 原版 NPC records 是 campaign data，不套用玩家創角最低能力門檻；
  仍須通過基本欄位、race/class 結構與能力值範圍驗證。
- 完成後 `4C2E` 防止重複招募，並返回同一 Pit level 1 dungeon。

## 可沿用的 Gold Box 知識

MON*CHA 同時承載玩家種族與特殊 NPC／怪物的共享 record layout。raw race `0`
不能在全域 race table 中猜成單一種族；應由作品 adapter 結合 `ADD NPC` context
與 canonical record identity 投影。另一方面，創角資格和 campaign NPC 合法性
是不同層次：玩家最低能力門檻不應拒絕原版隨附 NPC。

## 驗收

- 延續同一 real-session 進入 Pit level 1，先完成相鄰 monster-remains pause，
  再踏入 `(1,4)` terrain `0x85`。
- 驗證初見、枷印反應、encounter/parlay、三項回答與邀請完整繁中。
- 驗證手札 3 三頁在 TELL HER YOUR STORY 後解鎖。
- 驗證 YES 後 roster 增加為三人：英雄、愛麗雅絲、龍餌。
- 驗證 Alias fighter 6、Dragonbait saurial paladin 7、ScriptName 與 sidecars。
- 重跑 lifecycle 不得再次觸發或重複加入 NPC。
