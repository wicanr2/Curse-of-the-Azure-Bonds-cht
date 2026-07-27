# 第三百二十輪：Area 5 離場清理與阿卡巴告別

狀態：`READY`

## 反組譯證據

法師塔 block `0x33` 的 `WILDERNESS → DEPART` 先 `NEWECL 0x30`。ECL5 block
`0x30` initial entry `+0x0014` 不是空的轉接點，而是 Area 5 離場程序：

1. `GOSUB +0x009B` 以 `LOAD CHARACTER 0..7` 掃描 TeamList。
2. 每個有效 NPC slot 先比較 selected-player `0x7CB8 >= 0x80`，再比較
   `0x7C00 == "AKABAR BEL AKAS"`。
3. 若阿卡巴存在，`4C60==1 && 4C5E==1` 顯示他感謝隊伍、另有要事的告別；
   否則他表示仍須解放哈普。兩路都以 `DUMP` 正式移除他。
4. `FIND ITEM 0x5E/0x60/0x61` 發現任一黑暗精靈裝備時，顯示日光使武器與
   護甲腐朽；Continue 後依序 `DESTROY ITEMS` 三種 type。
5. 保存 `4BFB=0`、`7EE1=0xFF`、`7F12=1`，最後 `NEWECL 0x50` 回到全域旅程。

`0x7CB8` 是 selected-player window 的 control/morale byte，不是 ECL scratch。
阿卡巴 MON5 record 的 `ControlMorale=0xB2` 因此必須由 party context 在每次
`LOAD CHARACTER` 投影；不存在的 selector則投影為 0，讓掃描繼續。

## 引擎與中文契約

- `PartyMemberContext.ControlMorale` 保存作品 adapter 提供的原始 byte；VM 不依
  名稱猜 NPC 身分，也不依 CoAB 固定寫死 `0xB2`。
- 阿卡巴告別與裝備腐朽是兩個獨立 Continue boundary。State 必須在第一段套用
  DUMP、第二段套用 DESTROY ITEMS，不能把所有副作用延後到世界圖。
- 中文只替換顯示文字；script compare 仍使用 `ScriptName="AKABAR BEL AKAS"`。
- block `0x30 → 0x50` 是 ECL session continuation。State 不可看到 DEPART label
  就直接 `enterMap()`，否則會跳過 NPC、inventory 與 chapter work bytes。

## 驗收

- synthetic VM regression：`LOAD CHARACTER` 將 `ControlMorale=0xB2` 投影到
  `0x7CB8`，使 `COMPARE >= 0x80` 成立。
- real-image regression：block `0x30` 先產生阿卡巴告別與 resolved DUMP；恢復
  後再產生黑暗精靈裝備腐朽文字。
- 端到端 State regression：塔頂 DEPART 依序顯示兩段繁中、移除阿卡巴、
  銷毀 `0x5E/0x60/0x61`、保留普通 item，最後抵達 block `0x50` BIGPIC 121
  與 `ENTER CITY / JOURNEY ON / CAMP` 世界選單。
