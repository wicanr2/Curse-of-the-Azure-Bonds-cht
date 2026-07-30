# 399：外圍遺跡提爾雪雅結盟與臨時怪物盟友

狀態：`READY`

## 原始證據

| 成員 | SHA-256 |
|---|---|
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` |
| `GEO6.DAX` | `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c` |
| `MON6CHA.DAX` | `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b` |

手札 5 由使用者提供的
`Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf`
中 `JOURNAL ENTRY 5` 核對；公開的
[Lemon Amiga 手札逐字稿](https://www.lemonamiga.com/doc/curse-of-the-azure-bonds/405)
提供第二份文字交叉證據。手札把作弊者寫成 `Birsheya`，ECL 對話與選項則
寫成 `Beyrha`；本輪保存來源差異，不擅自宣稱原版拼寫問題已釐清。

## 地圖入口與手札 5

沿 `PATH` 進入 ECL／GEO6 block `42h` 的 `(0,12,E)`，向東一步可抵達
`(1,12)` terrain `01h`。首次進入便寫 `4CD0h=1`；writer 位於 payload
`+05F8h`，所以任何初次選擇都會消耗事件，重訪直接 `EXIT`。

初始選單為 `WAIT／ATTACK／FLEE`。`WAIT` 會讓羅剎妖自稱提爾雪雅
（Tirsheya），解鎖手札 5，再詢問玩家是否同行。手札說明他懷疑負責看守
氏族倉庫的貝爾雪雅作弊；玩家若協助闖入，可任取所需，他只要作弊證據。

## 第一場守衛戰

答應同行並攻擊入口守衛後，ECL 建立：

| MON6CHA | 名稱 | 數量 | icon |
|---|---|---:|---|
| `44h` | HELL HOUND | 5 | `44h` |
| `45h` | MARGOYLE | 5 | `45h` |

若拒絕攻擊，守衛會發現隊伍，玩家另有逃跑或被迫開戰分支；本輪正常玩家
路徑採主動攻擊。

## 貝爾哈的最後通牒

第一戰勝利後，貝爾哈（Beyrha）率隨從現身，要求玩家交出提爾雪雅的首級。
原始選單是 `TIRSHEYA／BEYRHA／FLEE`。

背叛提爾雪雅時，第二戰會建立五隻地獄犬、五隻石像鬼及一隻敵方
RAKSHASA。攻擊貝爾哈時，ECL 則建立 RAKSHASA ×1、HELL HOUND ×6、
MARGOYLE ×6，接著執行：

- `LOAD CHARACTER 8`；
- 寫 combat-team byte `7D0Ch=80h`；
- `ADD NPC 43h, morale=100`。

這代表第一個 RAKSHASA copy 是本場的提爾雪雅盟友，而不是敵人或永久
第九名隊員。勝利後 ECL 寫 `4CD1h=1`（writer 之一在 `+09CCh`），執行
`DUMP` 清理原版臨時角色，再回到地城。

## 固定八個玩家戰鬥槽

DOS combatant array 固定保留 index `0..7` 給最多八名玩家，怪物從 index
`8` 開始，與目前實際隊伍人數無關。舊 adapter 錯用
`TeamListIndex-len(activeParty)`；只有八人測試時才碰巧正確。

修正後：

1. ECL runtime 以固定 `8` 投影 `CombatTeamWrite` 到 `MonsterSpawn.PartyMask`。
2. BlockSession 在跨 pause 聚合 `LOAD MONSTER` 與稍後 team write 後重套投影。
3. State 不再把已標成我方的 MON*CHA 怪物 record 當永久玩家 record 解析。
4. encounter adapter 建立 `QuickFight=true`、`TemporaryAlly=true` 的
   monster fighter，戰後不污染 roster。

這是作品中立的 Gold Box ECL／戰鬥陣列語意，不是 CoAB 劇情特例。

## 資料化與驗證

- 事件文字、`TIRSHEYA／BEYRHA` 顯示名及完整手札 5 都來自 game-pack。
- raw ECL regression 驗證 `4CD0／4CD1`、兩場陣容、team write、
  `ADD NPC`、`DUMP` 與重訪 `EXIT`。
- 一人隊伍的 adapter regression 證明固定八槽。
- Standing Stone 起始正常路徑實際走入 block `42h`、讀取手札、完成兩戰，
  並確認提爾雪雅只在第二戰暫時加入我方。

## 明確邊界

- 地獄犬吐息、羅剎妖法術免疫／法術、石像鬼特殊防禦、AI、動畫與音效仍未
  完整還原。
- 其他拒絕／背叛／逃跑分支可由 VM 執行，但尚未全部加入 State 路徑矩陣。
- 倉庫財寶與作弊證據、block `42h` 其他事件、神殿、密道、最終決戰與結局
  尚未接通。
