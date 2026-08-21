# 1160 — 拆掉那兩道閘門要一次改六件事，以及盜賊公會那個座標位移是湊出來的

- 證據等級：`exact`（ECL2/0x02、ECL3/0x10 逐條讀出；GEO 與每格分派表交叉比對）
  ＋ 一次**已還原**的整輪實作實驗（六個 remake 改動 ＋ 十個測試斷言重新推導）
- 前置：[spec 1159](1159-storevalue-is-the-only-write-path.md)（12 處明確移動與兩道閘門）

## 為什麼要記這一份

spec 1159 把「拿掉宣告 spawn 的閘門」寫成下一項，判準是三件事一起做。實際做下去
是**六件**，而且其中兩件推翻了既有的斷言。整輪已經還原（工作區乾淨），但推導出來
的東西全部有效——下一輪照這份做就是機械工作。

## 要改的六件事

| # | 改什麼 | 為什麼 |
|---:|---|---|
| 1 | 拿掉 `projectFreshDungeonCoordinatesBeforeCall` 開頭的 `Spawn` early return | 宣告的 spawn 是**進場錨點**，不是「這張圖的腳本不會搬隊伍」 |
| 2 | `case 0x2E10` 的守衛改成逐條看 `CallRequests[i].BlockID == CurrentBlockID()` | 原本還要求整次執行都在同一個 block，跨 block 的執行（`NEWECL`）裡的重畫全被跳過 |
| 3 | `2Dh CALL C01Eh` 改讀腳本當下寫的 `C04D`，走完寫回暫存器 | 原作 `MoveForward` 讀地圖暫存器；腳本會在走之前指定方向 |
| 4 | 重畫要**消費**座標寫入（記「上一次重畫吃到哪個執行序」）| 原作是髒旗標，重畫時清掉；沒有這條游標，迴圈裡的第二次重畫會把進場 spawn 再投影一次，把中間走的步數抹掉 |
| 5 | 宣告的 spawn 改成**沒人指定時**的後備（這次執行投影過就不蓋回去）| 否則會得到「X／Y 來自 pack、朝向來自腳本」的混合值 |
| 6 | 拿掉 `ECL2/0x02` 的座標位移（見下一節）| 那個位移是投影被壓掉時湊出來的 |

⚠ **少做任何一件，主線都會停在某個門口**，而且症狀看起來都像「座標算錯」。

## ★★★ 盜賊公會的座標位移是湊出來的

`DungeonGeometryView`／`SetDungeonGeometryView` 對 `ECL2/0x02` 做過
`x=(x+8) mod 16`、`y=15−y`、朝向鏡射。把上面六件事做完之後可以直接讀出真相：

```text
ECL2/0x02:0047h  SAVE 08 C04B; SAVE 00 C04C; SAVE 01 C04D   { 進場 (8,0) 朝東 }
         0D44h   迴圈五次：GETTABLE 0DD1h[i] → 7F7B         { 方向表 1,2,2,2,2 }
         0D84h   COMPARE 7F7B C04D; IF = → 走一步；否則只轉向
```

⇒ 往東一步到 `(9,0)`，轉朝南，往南三步到 `(9,3)`。
而 [`ecl-cell-events.md`](../audit/ecl-cell-events.md) 的 `ECL2/0x02` 索引 1
（地形碼 `& 0x3F`）正是 `(3,3)`、**`(9,3)`** ⇒
「BEFORE YOU STANDS A BURLY MAN SURROUNDED BY SEVERAL…」＝**公會主人**。

⇒ **腳本的暫存器本來就是 GEO 格子**，公會與提爾弗頓共用 `GEO2/0x01`，
就像哈普村與黑暗精靈那張圖共用 `GEO5/0x32`（[spec 1158](1158-hap-village-extent-and-refused-edges.md)）。
[spec 292](292-tilverton-carriage-guild-transition.md) 記的到站暫存器 `(1,12,0)`
是**套過那個位移之後**的值，不是腳本寫的值。

## ★★ 拿掉閘門之後浮出來的原作行為（逐處已推導）

| 場景 | 原作做的事 | 依據 |
|---|---|---|
| 賢者菲拉妮 `(6,5)` | 退回招牌格 `(5,5)` | `ECL2/0x01:1441h` 共用收尾 |
| 科米爾武器店 `(2,12)` | 退回 `(3,12)` | 同上 |
| 剛德神殿祭壇 `(0,7)` | 退回 `(1,7)` | 同上 |
| 高階祭司 `(1,10)` | 「YOU MOVE AWAY.」退回 `(2,10)` | `142Eh` |
| 訓練所 `(5,2)` | 「YOU EXIT THE HALL.」往南一格 `(5,3)`（YES／NO 兩支都走這裡）| `1073h` |
| 「THE CURSE」酒館 `(6,10)` | 走出門到 `(5,10)`，再演走位動畫到建築側邊 `(7,12)` 朝東 | `0C75h` ＋ `0CE9h`..`0D2Dh` |
| 城門衛兵 `(1,0)` | 第一次「…AND THEY SEND YOU BACK.」送回 `(2,0)`；第二次才演皇家馬車 | `139Ah`／`13C4h` |
| 盜賊公會 | 進場 `(8,0)` 朝東 → 走位到 `(9,3)` 朝南（公會主人那一格）| `0047h`／`0D44h` |
| 猶拉什 `ECL3/0x10` | 進場 `(0,8)` 朝西 → `0127h` 到 `(1,0)` → 一串 `C01E` 走到 `(0,3)`，**收尾朝西**；離開指揮官那段的朝向由 `0185h` 的 `SAVE 4C00 C04D` 決定 | `006Ch`／`0127h`／`0185h` |

★ 這些原本都被壓成「留在原地」，所以主線路線測試的每一段起點都要往前挪一格
——不是把失敗值抄進期望值。

## ★★ 巫師塔塔頂怎麼下來（實驗裡卡住的那一處）

拿掉閘門之後，`ECL5/0x33:022Bh`（`SAVE 03 C04B; SAVE 01 C04C; SAVE 02 C04D`
＋「YOU ARE SUDDENLY ON THE ROOF OF THE TOWER AMIDST A HUGE HOST OF BLACK
DRAGONS.」）真的把隊伍搬到塔頂 `(3,1)`。整段巫師塔劇情都演完了
（`wizard-tower.*` 全部命中），但**隊伍下不來**：

- 站在 `(3,1)` 跑生命週期沒有任何輸出（`mode=dungeon`、無選項、無訊息）。
- `(3,1)` 在 `GEO5/0x33` 上是孤立的，走不到別的格子。
- 走到那裡之前 `wizard-tower.dragons-depart` 已經命中，代表 CAVES／WILDERNESS
  那個選單已經被消費掉了。

★ **出口找到了，而且不是「下不來」——是測試走訪器把路切斷了。**

原作的塔內出口是 `ECL5/0x33:0811h` 那段：

```text
0811h  PRINT  "ALSO NOTE A SECRET PASSAGE THAT WILL TAKE YOU DIRECTLY"
083Eh  PRINT  "TO THE WILDERNESS. WHICH DO YOU TAKE?"
085Dh  HORIZONTAL MENU          { CAVES / WILDERNESS / 第三項 }
087Ch  ON GOTO → 088Bh / 08A6h / 1BD6h
088Bh  SAVE 06 C04B; SAVE 0F C04C; SAVE 00 4BE7; SAVE 00 4BE8; NEWECL 32h
```

⇒ 選 `CAVES` 就是**把隊伍設到 `(6,15)` 再 `NEWECL 0x32`**，
⚠ 而且那兩句座標寫入**後面沒有 `CALL 2E10h`**——交接靠 `NEWECL`，
所以 remake 這一側要由 `syncDungeonStateFromECLRegisters` 接。
第三項走 `1BD6h`，就是 spec 1157 那 15 處退格之一。

而塔頂 `(3,1)` 在 `GEO5/0x33` 上**不是孤立的**（`geo-probe` 顯示它有兩個可走
方向）。實驗裡走不出去是因為 `walkNormalDungeonTo` 有一份寫死的排除清單
（`(10,2)`／`(8,11)`／`(8,15)`／`(12,10)`），其中 `(8,15)` 與 `(12,10)` 正好落在
這張圖上，把第 15 列的走廊切斷了。⇒ 下一輪要處理的是**走訪器的排除清單**，
不是遊戲機制。

## 明確不宣稱

- 沒有宣稱那六件事做完之後**只剩**巫師塔一處——實驗跑到那裡就停了，
  後面的段落（希爾斯法、贊提爾、密斯卓諾）沒有走完。
- 沒有宣稱 `ECL2/0x02` 以外沒有別的地方也被加過座標位移。
- 沒有宣稱 spec 292 的其他斷言有問題；只指出它記的 `(1,12,0)` 是套過位移之後的值。
