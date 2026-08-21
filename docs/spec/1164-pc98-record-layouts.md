# 1164 — PC-98 除錯符號展得開 record 版面，角色記錄的 `unknown` 歸零

- 證據等級：`exact`（型別／成員表逐筆讀出）＋ `measured`（原版存檔逐格對過）
- 工具：`cmd/borland-symbols -record <名稱>`／`-records`
- 產物：[`docs/audit/pc98-record-layouts.md`](../audit/pc98-record-layouts.md)（32 個具名 record）

## 連法

Borland legacy 除錯表的型別記錄有八個位元組：`id`、名字索引、大小，尾巴三個
先前只當成 `Detail` 收著。那三個裡的**後兩個是小端序的成員索引**：

- 指標型別（`id=22`）指到目標型別。`CHARRECPTR` 的 `Detail` 是 `00 E0 03`
  ⇒ `0x03E0` ＝ 992 ＝ `CHARREC`。
- record 型別（`id=30`）指到**第一個成員的 1-based 索引**。`CHARREC` 的
  `Detail` 是 `00 67 00` ⇒ 103 ⇒ 成員表第 103 筆（陣列索引 102）就是 `NAME`。

欄位依序排列、沒有對齊填塞，所以位移是前面各欄大小的累加。

⚠ **成員數不在型別記錄裡**，只能加到記錄大小為止。代價是記錄大小若讀錯而且
剛好比實際大，會悄悄多吃一個成員；只有欄位跨過記錄邊界才擋得下來。

## `CHARREC`：24 個 `unknown` 全部有名字了

`cmd/save-field-coverage` 從 `unknown=24` 變成 **`unknown=0`**
（`decoded` 299／`documented` 123／合計 422）。新命名的十一段：

| 位移 | 長度 | PC-98 符號 | 意思 |
|---|---:|---|---|
| `+072h` | 1 | `MINREST` | 最短休息時數 |
| `+0DEh` | 1 | `SIZE` | 體型 |
| `+0F6h` | 1 | `RAISED` | 被復活過 |
| `+0F8h` | 1 | `MODIFIED` | 派生值已重算 |
| `+0F9h` | 1 | `OLDCLASS` | 換職前的職業 |
| `+0FAh` | 1 | `OLDLEVEL` | 換職前的等級 |
| `+11Ah` | 1 | `RACETYPE` | 種族大類 |
| `+11Dh` | 1 | `BASEATTBLOWS[1]` | 攻擊次數基準的第二個武器槽 |
| `+126h` | 1 | `RANDOMID` | 角色亂數種子 |
| `+13Ch` | 2 | `BASEEXP` | 基準經驗值（word）|
| `+13Eh` | 1 | `EXPPERHP` | 每點 HP 的經驗 |
| `+13Fh` | 1 | `HEAD` | 頭部造型 |
| `+140h` | 1 | `BODY` | 身體造型 |
| `+145h` | 7 | `COLORLIST` | 人像配色表 |
| `+191h` | 1 | `PDLNREMOVECURSE` | 解除詛咒的來源標記 |
| `+193h` | 1 | `DUM2` | 保留欄 |
| `+194h` | 1 | `DUM3` | 保留欄 |

## DOS 比 PC-98 少一個位元組

`CHARREC` 在 PC-98 是 **423** bytes，原版 DOS 的 `CHRDAT?.SAV` 是 **422**。
差的那一格落在 `+14Ch`..`+186h` 之間，之後所有欄位往前挪一格。

拿原版存檔（`docs/reference/original-dos/save-samples/CHRDATA1.sav`）對，
挪一格之後每一個能驗的錨點都成立：

| 位移 | 值 | 對照 |
|---|---|---|
| `+078h` `MAXHP` | `33h` ＝ 51 | 與 `+1A4h` `CURRENTHP` ＝ 51 相同（滿血）|
| `+0E4h` `BASEMOVE` | `0Ch` ＝ 12 | 與 `+1A5h` `MOVE` ＝ 12 相同 |
| `+124h` `BASEAC` | `32h` ＝ 50 | 建角寫 `32h`（AC 10），spec 1000／1140 |
| `+145h` `COLORLIST` | `91 A2 B3 C4 E6 F7 00` | 七格 EGA 色號對 |
| `+187h` `ENCUMBERANCE` | `2C 01` ＝ 300 | 不挪就會讀成 1 |
| `+196h` `STATUSOK` | 1 | 活著且能行動 |

不挪的話 `CURRENTHP` 會落在 `+1A5h` 讀成 12、`MOVE` 落在記錄外。⇒ **挪一格是
確定的**；少掉的那一格**最可能是 `MONSTERTYPE`**（PC-98 版面裡它在 `+14Ch`，
而 remake 台帳一直把 `NUMITEMS` 放在 `+14Ch`），但 `+14Ch`..`+186h` 之間都是
0 或指標，這份樣本分不出來。

## 三處與既有規格的讀法不一致

新名字沒有覆蓋既有的 `decoded` 欄位，只並列。三處要另外對：

| 位移 | 既有讀法 | PC-98 符號 |
|---|---|---|
| `+0E6h` | 多職角色的現行等級（spec 185）| `HIGHESTPREVLEVEL`（前一個最高等級）|
| `+11Ch` | 武器槽選擇，0 時回落到槽 2（spec 1010）| `BASEATTBLOWS[0]` |
| `+192h` | ECL 旗標，投影位址 `7CE4h`（spec 1098）| `DUM1`（保留欄）|

`+192h` 兩者相容——ECL 投影到一個 Pascal 端的保留欄沒有矛盾。另外兩處是真的
要再讀一次 DOS 側的用法。

## 還能展開的

同一張表裡另有 31 個具名 record，逐欄版面都在 `docs/audit/pc98-record-layouts.md`：
`WALLSET`（2,340）、`TACTICALMAP`（1,257）、`MAP3D`（1,024）、`CHARITEMREC`（103）、
`CHARITEMFILREC`（63，就是 `.SWG`）、`TREASUREREC`（32）、`COMBATVARREC`（22，
spec 806 那份 22 bytes）、`EFFECTREC`（9，就是 `.FX`）等。

## 明確不宣稱

- 沒有宣稱每個欄位的**語意**——符號給的是名字與大小，怎麼用要另外讀程式碼。
- 沒有宣稱 DOS 與 PC-98 的欄位語意一定相同；只證明版面在挪一格之後對得上。
