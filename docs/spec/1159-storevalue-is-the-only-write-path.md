# 1159 — 原作只有一條寫入路徑（`STOREVALUE`），以及那 12 處「明確移動」

- 證據等級：`exact`（125 處 `CALL 2E10h` 由 `cmd/ecl-redraw-sites` 逐處分類；
  12 處的守衛與目的地逐處讀出）＋ 一次已還原的實跑探針
- 前置：[spec 1155](1155-redraw-call-coordinate-divergence.md)、
  [spec 1157](1157-restore-previous-cell-sites-and-shared-handlers.md)、
  [spec 1158](1158-hap-village-extent-and-refused-edges.md)

## 結論

原作寫變數只有一條路：`STOREVALUE`。所有會寫變數的 opcode 都走它，而它同時把
隊伍座標鏡射到 `720Fh`／`7210h`／`7211h` 並立髒旗標。⇒ **「哪些 opcode 寫過座標」
在原作裡不是分類問題，任何一個都算。**

remake 的 VM 先前只在 `09h SAVE` 與 `35h SAVE TABLE` 記 `SaveWrites`，算術、
`08h RANDOM`、`2Ah GETTABLE` 的寫入完全沒有記錄。重畫要靠 `SaveWrites` 重建
「這一格被寫過了嗎」，所以那些路徑寫的座標**在 remake 這一側等於沒寫**。
現在每一條「算完再存」的指令都走同一支 `recordStore`。

## ★★★ 分類數字因此改了：12 處移動，不是 8

`cmd/ecl-redraw-sites` 沿 `TraceGraph` 走完全 corpus，依 `CALL 2E10h` 前六條
指令裡的座標寫入分類：

| 類別 | 舊（只認 `SAVE`）| 新（認全部寫入路徑）|
|---|---:|---:|
| 寫了 `C04D`（朝向也換）| 67 | 67 |
| 明確移動 | 8 | **12** |
| 退回上一格 | 15 | 15 |
| 只有 `PICTURE` | 3 | 3 |
| 前六條沒有座標寫入 | 32 | **28** |

差的四處全是 `GETTABLE` 的迷宮傳送（假門），先前被歸到「沒有座標寫入」——
**這是工具與 VM 同一個成因的假零**，不是資料有誤。

## 那 12 處逐處

| 位置 | 情境 | 寫法 ⇒ 目的地 |
|---|---|---|
| `ECL2/0x01:0C7Eh` | 提爾弗頓「THE CURSE」酒館 `(6,10)`，喝完被請出門 | `SUBTRACT 1 C04B` ⇒ 往西一格 `(5,10)`；接著 `0CE9h`..`0D2Dh` 每寫一次座標就 `GOSUB 0D73h` 重畫，演出朝南走兩格、轉朝東走兩格，停在建築側邊 `(7,12)` |
| `ECL2/0x01:0DB5h` | 「AFTER MUCH ROISTERING, THE PARTY PASSES OUT.」（守衛 `COMPARE 4C0A 0Ah; IF <; RETURN` ＝ 喝滿十杯）| `SAVE 0Eh C04C; SAVE 07 C04B` ⇒ 傳送到旅店床上 `(7,14)` |
| `ECL2/0x01:1073h` | 訓練所 `(5,2)`，「YOU EXIT THE HALL.」 | `ADD 1 C04C` ⇒ 往南一格 `(5,3)`。⚠ YES／NO 兩支都會走到這裡，`NO` 只是跳過 `PROGRAM 0` 的訓練畫面 |
| `ECL2/0x02:007Dh` | 盜賊公會，`COMPARE C04B 00` 不成立 | `SAVE 0Eh C04B; SAVE 0Fh C04C` ⇒ `(14,15)` |
| `ECL2/0x02:008Eh` | 同上，`C04B == 0` | ⇒ `(10,15)` |
| `ECL3/0x15:0207h` | 猶拉什迷宮「YOU HAVE FOUND A FALSE DOOR.」 | `GETTABLE 0AD9h/0B14h` 查 `7F7A` ⇒ 表驅動傳送 |
| `ECL4/0x21:00C2h` | 「YOU ARE DRAGGED THROUGH THE TEMPLE.」之後 | `ADD 1 C04C` ⇒ 往南一格 |
| `ECL5/0x33:07CDh` | 巫師塔，守衛 `COMPARE AND C04D 01 4C09 01`（朝東且 `4C09 == 1`）| `ADD 1 C04B` ⇒ 往東一格 |
| `ECL5/0x33:0984h` | 巫師塔的假門 | `GETTABLE 1C6Dh/1C77h` ⇒ 表驅動傳送 |
| `ECL5/0x35:04E3h` | 假門 | `GETTABLE 1537h/156Fh` ⇒ 表驅動傳送 |
| `ECL6/0x42:0D93h` | 密斯卓諾外城，某段崩塌之後 | `SAVE 03 C04B; SAVE 06 C04C` ⇒ `(3,6)`，緊接 `SETUP MONSTER 44h` |
| `ECL6/0x45:01F6h` | 假門 | `GETTABLE 0923h/0963h` ⇒ 表驅動傳送 |

假門那四處是兩層查表：`AND C04F 3Fh → 7F79`，`GETTABLE` 把 `7F79` 換成 `7F7A`，
再由 `7F7A` 查 X／Y 兩張表。

## ★★★ 還有兩道閘門擋著這 12 處的一半以上

修好寫入路徑之後，**能不能真的搬**還取決於 `applyECLCallSignals` 的兩個條件：

**一、宣告了 spawn 的地圖整張跳過。**
`projectFreshDungeonCoordinatesBeforeCall` 開頭：地圖定義有 `Spawn` 就直接
`return`。命中的有 `tilverton.first-person`（提爾弗頓，area 2 的 block 1 由
geometry fallback 命中）、`zhentil-keep.dark-shrine`、`original.geo5.block-33`
⇒ 上表 12 處裡有 6 處在這三張圖上，一步都搬不動。

**二、跨 block 的執行整批跳過。**
同一個 `case 0x2E10` 還要求
`SessionStartBlockID == SessionEndBlockID == CallRequests[i].BlockID`，
所以任何跨 block 的執行（`NEWECL`）裡的 `CALL 2E10h` 一律不投影。盜賊公會那兩處
（`ECL2/0x02`）就在這種執行裡。

⇒ 目前真的會生效的只有 `ECL3/0x15`、`ECL5/0x35`、`ECL6/0x42`、`ECL6/0x45`
四處（其中三處是假門，本輪才隨著 `GETTABLE` 記錄一起接上）。

## ★★ 拿掉第一道閘門會看到什麼（本輪已量、已還原）

把 `Spawn` 那個 early return 拿掉之後跑主線，提爾弗頓一次浮出六個原作行為：

| 場景 | 原作做的事 |
|---|---|
| 賢者菲拉妮 `(6,5)` | 收尾 `1441h` 退回進門前的招牌格 `(5,5)` |
| 科米爾武器店 `(2,12)` | 退回 `(3,12)` |
| 剛德神殿祭壇 `(0,7)` | 退回 `(1,7)` |
| 高階祭司 `(1,10)` | 「YOU MOVE AWAY.」退回 `(2,10)` |
| 「THE CURSE」酒館 `(6,10)` | 走出門 ＋ 走位動畫，停在 `(7,12)` |
| 城門衛兵 `(1,0)` | 第一次「…AND THEY SEND YOU BACK.」送回 `(2,0)` |

★ 但**還不能就這樣拿掉**：`2Dh CALL C01Eh`（`MOVEFORWARD`）目前用的是
`s.DungeonDirection`，而不是腳本當下寫的 `C04D`。盜賊公會的抓捕動畫
（`ECL2/0x02:0DA6h` 的走四步迴圈）在 `0D90h` 先 `SAVE 7F7B C04D` 指定方向；
朝向沒同步過去，四步就會走錯方向——[spec 292](292-tilverton-carriage-guild-transition.md)
量到的到站暫存器是 `(1,12,0)`（朝北），朝西走會變成 `(13,0)`。
⇒ 拿掉閘門要**連同 `C01E` 改成讀 `C04D`** 一起做，否則主線會停在公會門口。

## remake

- `internal/ecl/runtime.go`：新增 `recordStore`，`04h`–`07h`、`08h`、`09h`、
  `2Ah`、`2Fh`／`30h`、`35h` 全部走它。
- `cmd/ecl-redraw-sites`：新的可重生清冊，數字對得上 spec 1155 的 125 處。
- 兩個 fixture 跟著改：`GETTABLE` 那條多一筆、死精靈毒氣陷阱那條多八筆
  （`08B0h` 的 `ADD 1 7F79` 走整隊）。

## 明確不宣稱

- 沒有宣稱假門那四處的兩層查表**每個索引**都對得上原作——只讀了表本身與
  取值方式，沒有逐格實跑比對。
- 沒有宣稱除了 `Spawn` 與跨 block 這兩道閘門以外沒有第三道。
- 沒有宣稱 `menu.Location`、`WHO`／`FIND ITEM` 那些引擎側寫入也要記進
  `SaveWrites`；它們在原作同樣走 `STOREVALUE`，但本輪沒有量到會影響座標。
