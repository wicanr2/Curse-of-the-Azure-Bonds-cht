# 投射武器的第二聲：`SHOWARROW` 依物品類別分歧

由 `cmd/missile-sound-classes` 產生，不要手改。分歧鏈的位元組證據見 spec 1186。

`SHOWARROW` 進場**無條件**放 `ARROWFX`，之後依 `CHARITEMREC.ITEMPTR`（＝物品類別，remake 的 `ItemRecord.Type`）在飛行動畫尾端再放一聲。

⚠ **類別表有一列不代表遊戲裡有那件東西**。`實例` 是六章 `ITEM*.DAX` 裡真的存在的件數；0 就表示那條分支玩家走不到，是「不必接」而不是「還沒接」。

⚠ `射程` 與 `MISSLETYPE` 取自類別表（`ITEMREC` 的 `+0Ch`／`+0Eh`）。原作的 `USINGMISSLEWEAPON` 用的是 `射程 > 1`，`USINGHURLEDWEAPON` 用的是`MISSLETYPE and 14h ＝ 14h`——**兩個判斷式不同**，不要拿其中一個去解釋另一個。

| 第二聲 | 位移 | 類別 | 名稱 | 實例 | 射程 | `MISSLETYPE` |
|---|---:|---:|---|---:|---:|---|
| ARROWFX（箭） | `2B4Ah` | `09h`（9）| 4 Dart | 3 | 6 | `1Ah`（發射／投擲） |
| ARROWFX（箭） | `2B4Ah` | `15h`（21）| 	2 Javelin、1 Javelin of Brazier | 4 | 7 | `1Ah`（發射／投擲） |
| ARROWFX（箭） | `2B4Ah` | `1Ch`（28）| 
20 Quarrel、12  Quarrel | 4 | 0 | `8Ah`（發射） |
| ARROWFX（箭） | `2B4Ah` | `1Fh`（31）| Spear | 3 | 4 | `14h`（投擲） |
| ARROWFX（箭） | `2B4Ah` | `49h`（73）| 10 Arrow、20 Arrow、10 Arrow +1 | 5 | 0 | `00h` |
| ARROWFX（箭） | `2B4Ah` | `64h`（100）| Dart of Bowl | 1 | 6 | `1Ah`（發射／投擲） |
| SWISHFX（揮擊） | `2B81h` | `02h`（2）| Hand Axe | 3 | 4 | `14h`（投擲） |
| SWISHFX（揮擊） | `2B81h` | `07h`（7）| Club | 3 | 4 | `14h`（投擲） |
| SWISHFX（揮擊） | `2B81h` | `14h`（20）| Hammer | 3 | 4 | `14h`（投擲） |
| WHISTLEFX（哨音） | `2BB4h` | `55h`（85）| **（遊戲裡沒有這一類的物品）** | 0 | 4 | `1Ah`（發射／投擲） |
| WHISTLEFX（哨音） | `2BB4h` | `56h`（86）| 1 Oil | 3 | 4 | `1Ah`（發射／投擲） |
| WHISTLEFX（哨音） | `2C01h` | `2Fh`（47）| Sling | 3 | 21 | `0Ah`（發射） |
| WHISTLEFX（哨音） | `2C01h` | `62h`（98）| **（遊戲裡沒有這一類的物品）** | 0 | 18 | `1Ah`（發射／投擲） |
| WHISTLEFX（哨音） | `2C01h` | `65h`（101）| Small Raft Sling | 5 | 24 | `0Ah`（發射） |
| SWISHFX（揮擊） | `2C48h` | 其餘全部 | — | — | — | — |

| 指標 | 數字 |
|---|---:|
| 分歧鏈點名的類別 | 14 |
| 其中遊戲裡真的有物品的 | 12 |
| 其中一件都沒有的（走不到）| 2 |

## 落到預設分支的發射武器

有發射位元（`MISSLETYPE` bit 3）、遊戲裡也真的有，但**分歧鏈沒點名**⇒ 走 `2C48h` 的 `SWISHFX`。這一份是「預設分支不是空的」的證據：少了它會以為沒被點名的類別不會走到 `SHOWARROW`。

| 類別 | 名稱 | 實例 | 射程 |
|---:|---|---:|---:|
| `29h`（41）| Composite Long Bow | 3 | 22 |
| `2Ah`（42）| Composite Short Bow | 3 | 19 |
| `2Bh`（43）| Long Bow | 3 | 22 |
| `2Ch`（44）| 	Short Bow | 3 | 16 |
| `2Eh`（46）| Light Crossbow | 4 | 19 |
