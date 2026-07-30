# 398：Burial Glen 手札 56 與東界遺跡出口

狀態：`READY`

## 原始證據

| 成員 | SHA-256 |
|---|---|
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` |
| `GEO6.DAX` | `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c` |
| `MON6CHA.DAX` | `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b` |

手札 56 另由使用者提供的
`Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf`
第 13 個 PDF 頁面、印刷頁 23–24 核對。它說明無名人物曾協助創造最初的
枷印、提朗瑟克斯如今不必與其他人分享控制權，以及其族群會在玩家攻打神殿
時奪回追隨者，藉此削弱提朗瑟克斯。

## terrain `07h`：可重複的蜘蛛墓穴

GEO6 block `40h` 的 `(14,3)` 是 terrain `07h`。每次進入都立即建立：

- `MON6CHA 41h` PHASE SPIDER ×6，icon `41h`；
- `MON6CHA 49h` RAKSHASA ×1，icon `43h`。

勝利後接回既有的 `LOOT GRAVE／REBURY SKELETON／GO` 骸骨選單。原 ECL
沒有持久完成旗標；預先寫入 `4CC3h=1` 或 `7F82h=1` 也不會略過。因此
不能自行改造成一次性事件。

## terrain `0Ch`：無名人物與手札 56

GEO6 `(4,8)` 的 terrain `0Ch` 初次進入會先寫 `4CC7h=1`，再顯示
`COMBAT／WAIT／FLEE／PARLAY`。`WAIT` 與 `PARLAY` 都會進入快速談話、
提示玩家記錄手札 56，最後說 `HURRY ON...` 並離開；之後重踏直接
`EXIT`。正常玩家回歸由螳螂人營地 `(8,9)` 沿合法 GEO 路徑抵達
`(4,8)`，選 `WAIT`，由 game-pack 解鎖完整繁中手札。

## Burial Glen 東界

紅羽戰士墓穴位於 `(13,6)`。向東走到 `(15,6,E)`，再由正常地城出口
lifecycle 執行 ECL6 block `40h` entry 0，會顯示
`PATH／WOODS／TURN BACK`。

| 選項 | ECL 結果 | 目的地 |
|---|---|---|
| `PATH` | `NEWECL 42h` | `(0,12,E)` |
| `WOODS` | `NEWECL 42h` | `(0,6,E)` |
| `TURN BACK` | 留在 `40h`，寫 `7EC9h=FFh` | Burial Glen |

進入 `42h` 後，龍盔會報告提朗瑟克斯在北方。State 現在對所有地城內
`NEWECL` 區塊轉換投影 `C04B／C04C／C04D`；目的地座標來自原 ECL
register，而不是在 CoAB 劇情分支中寫死。

## 資料化與測試

- 手札 56、無名人物、出口提示及三個選項都由 CoAB JSON stable ID 提供。
- 產品測試透過同一份 game-pack 解析顯示文字，不複製繁中常值。
- raw ECL 測試鎖定 `WAIT／PARLAY` 一次性語意與三個出口座標。
- Standing Stone 起始玩家路徑實際走到 terrain `0Ch`，再完成紅羽戰士
  戰鬥、走到東界，先選 `TURN BACK`，再選 `PATH` 進入 block `42h`。

## 明確邊界

- terrain `0Ch` 的 `COMBAT／FLEE` 會發出 encounter action 2；其外部
  routine 完整副作用仍待反組譯。
- terrain `07h` 的怪物完整能力、財寶、AI 與演出仍未完成。
- block `42h` 的完整遺跡、神殿、提朗瑟克斯決戰與結局尚未接通。
- 本輪證明 Burial Glen 到下一區域的正常 handoff，不代表章節或整款遊戲
  已完整可通關。
