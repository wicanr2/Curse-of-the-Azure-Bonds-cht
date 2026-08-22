# 原作在哪裡換曲，換成哪一首（PC-98）

由 `cmd/pc98-music-triggers` 產生，不要手改。

⚠ **`MUSICNO` 是 1 起算的**：12 首曲子、寫入值最大 12，直接當 0 起算索引會在最後一首溢位。game pack 的 `reference_selector` 正是 1..12，`driver_index` 才是 0 起算。抄錯一邊會讓每一首都差一格，而**每一格都還是合法曲名**，所以不會有任何錯誤訊息。

⚠ 這是 **PC-98** 的版面（名字由 Borland 除錯符號直接讀出）。DOS 沒有符號，對應的格子要另外認。

| 總計 | 數量 |
|---|---:|
| 換曲點（`mov byte [MUSICNO], imm`）| 11 |
| 被選到的相異曲目 | 10 |
| game pack 宣告的曲目 | 12 |

| 模組 | 單元 | 位移 | 選擇子 | 曲目 |
|---|---|---:|---:|---|
| `PC98-GAME.EXE` | — | `9486h` | 3 | 城鎮 |
| `PC98-GAME.EXE` | — | `94AEh` | 4 | 地城三 |
| `PC98-GAME.EXE` | — | `94C4h` | 6 | 村莊 |
| `PC98-GAME.EXE` | — | `94CBh` | 5 | 荒野 |
| `PC98-GAME.EXE` | — | `94E2h` | 8 | 散提爾堡城壁 |
| `PC98-GAME.EXE` | — | `94F9h` | 9 | 盜賊公會 |
| `PC98-GAME.EXE` | — | `9514h` | 12 | 地城 |
| `overlay-01` | — | `093Ch` | 1 | 標題 |
| `overlay-05` | POSTCOM | `1955h` | 2 | 角色建立 |
| `overlay-17` | GEN | `0B08h` | 2 | 角色建立 |
| `overlay-18` | — | `168Dh` | 10 | 結局 |

## 播放常式的呼叫點（音效那一半）

`SOUNDX` 那一組常式在程式碼段 `0893h`，段內位移取自 Borland 符號表。

| 常式 | 位移 | 呼叫點 | 來源模組 |
|---|---:|---:|---|
| SOUNDFX（音效） | `0000h` | 36 | overlay-01×1、overlay-02×16、overlay-03×2、overlay-13×9、overlay-14×3、overlay-22×1、overlay-24×2、overlay-32×2 |
| INITSOUND（初始化） | `010Dh` | 10 | overlay-02×10 |
| MSCPLAY（放音樂） | `0114h` | 5 | overlay-01×1、overlay-05×1、overlay-17×2、overlay-18×1 |
| BGMPLAY（背景音樂） | `0177h` | 1 | overlay-26×1 |

合計 52 處。

### `SOUNDFX` 每一處在放什麼音效

引數是**音效描述子變數**（`push word [位址]`），不是編號；名字由 PC-98 的 Borland 除錯符號直接讀出。

⚠ 只找立即數的話只解得出 5 處——會得到「大部分音效查不到」這個假結論。真正的形狀是推變數。

⚠ 這一節**直掃位元組**（`9A 00 00 93 08`），不走 far-call 對照表。表只收得到 IDA 認成程式碼的呼叫點，比實際少 12 處，而且其中一處會改結論：`LIGHTNINGFX` 在表裡是 0 處，看起來像「remake 有、原版沒有」，實際上它在 `CASTSPELL` 裡。**假零的來源是掃描面，不是原作。**

★ **交叉檢查**：far-call 對照表列 36 處、直掃 54 處；表裡那些是直掃的真子集。

| 音效 | 呼叫點 | 來源模組 |
|---|---:|---|
| ARROWFX（箭） | 2 | overlay-13×2 |
| CASTFX（施法） | 2 | overlay-22×2 |
| CRASHFX（撞擊） | 1 | overlay-02×1 |
| DEADFX（死亡） | 3 | overlay-03×1、overlay-32×2 |
| FIREBALLFX（火球） | 2 | overlay-02×1、overlay-22×1 |
| HITFX（命中） | 2 | overlay-13×2 |
| LIGHTNINGFX（閃電） | 1 | overlay-22×1 |
| MISSFX（揮空） | 1 | overlay-24×1 |
| OVERTUREFX（序曲） | 1 | overlay-01×1 |
| PADFX（腳步） | 7 | overlay-02×3、overlay-13×1、overlay-14×3 |
| SOUNDHALT（停止） | 18 | overlay-02×13、常駐×5 |
| SOUNDOFF（關） | 3 | 常駐×3 |
| SOUNDON（開） | 5 | overlay-03×1、常駐×4 |
| SPELLHITFX（法術命中） | 1 | overlay-24×1 |
| SWISHFX（揮擊） | 3 | overlay-13×3 |
| WHISTLEFX（哨音） | 2 | overlay-13×2 |

解出 54 處，還有 0 處的引數靜態看不出來。

#### 逐處：哪一支常式在放

所在常式取**同段裡位移不大於呼叫點的最後一個符號**。符號表只收得到公開程序，所以模組內部的靜態常式會掛在前一個公開名字底下——標成 `A＋n`，那個 `n` 就是它離公開入口多遠，不要當成「就是 A」。

| 音效 | 模組 | 位移 | 所在常式 |
|---|---|---:|---|
| ARROWFX（箭） | `overlay-13` | `2A7Ch` | **SHOWARROW**＋Ah |
| ARROWFX（箭） | `overlay-13` | `2B4Ah` | **SHOWARROW**＋D8h |
| CASTFX（施法） | `overlay-22` | `1701h` | **CASTSPELL**＋2C4h |
| CASTFX（施法） | `overlay-22` | `1D72h` | **CASTSPELL**＋935h |
| CRASHFX（撞擊） | `overlay-02` | `35A1h` | **LOADINTERPET**＋35A1h |
| DEADFX（死亡） | `overlay-03` | `03C5h` | **DOPROTECT**＋294h |
| DEADFX（死亡） | `overlay-32` | `1537h` | **SUBTRACTDUDE**＋11h |
| DEADFX（死亡） | `overlay-32` | `15D4h` | **SUBTRACTDUDE**＋AEh |
| FIREBALLFX（火球） | `overlay-02` | `31AFh` | **LOADINTERPET**＋31AFh |
| FIREBALLFX（火球） | `overlay-22` | `16E7h` | **CASTSPELL**＋2AAh |
| HITFX（命中） | `overlay-13` | `15C0h` | **ANYUNDEAD**＋10Ch |
| HITFX（命中） | `overlay-13` | `1863h` | **ANYUNDEAD**＋3AFh |
| LIGHTNINGFX（閃電） | `overlay-22` | `16F6h` | **CASTSPELL**＋2B9h |
| MISSFX（揮空） | `overlay-24` | `2494h` | **TWINKLE**＋8Ah |
| OVERTUREFX（序曲） | `overlay-01` | `0A5Fh` | **DOINTRO**＋129h |
| PADFX（腳步） | `overlay-02` | `319Fh` | **LOADINTERPET**＋319Fh |
| PADFX（腳步） | `overlay-02` | `31BAh` | **LOADINTERPET**＋31BAh |
| PADFX（腳步） | `overlay-02` | `3CAFh` | **GOECL**＋28Eh |
| PADFX（腳步） | `overlay-13` | `095Dh` | **REALMOVE**＋1CFh |
| PADFX（腳步） | `overlay-14` | `080Eh` | **LOADMOVEMENT**＋80Eh |
| PADFX（腳步） | `overlay-14` | `0ACFh` | **PREMOVEPARTY**＋1EBh |
| PADFX（腳步） | `overlay-14` | `0B1Ah` | **PREMOVEPARTY**＋236h |
| SOUNDHALT（停止） | `overlay-02` | `1863h` | **LOADINTERPET**＋1863h |
| SOUNDHALT（停止） | `overlay-02` | `1895h` | **LOADINTERPET**＋1895h |
| SOUNDHALT（停止） | `overlay-02` | `18B8h` | **LOADINTERPET**＋18B8h |
| SOUNDHALT（停止） | `overlay-02` | `18F6h` | **LOADINTERPET**＋18F6h |
| SOUNDHALT（停止） | `overlay-02` | `1928h` | **LOADINTERPET**＋1928h |
| SOUNDHALT（停止） | `overlay-02` | `194Bh` | **LOADINTERPET**＋194Bh |
| SOUNDHALT（停止） | `overlay-02` | `196Eh` | **LOADINTERPET**＋196Eh |
| SOUNDHALT（停止） | `overlay-02` | `19A0h` | **LOADINTERPET**＋19A0h |
| SOUNDHALT（停止） | `overlay-02` | `19C3h` | **LOADINTERPET**＋19C3h |
| SOUNDHALT（停止） | `overlay-02` | `19E7h` | **LOADINTERPET**＋19E7h |
| SOUNDHALT（停止） | `overlay-02` | `1A56h` | **LOADINTERPET**＋1A56h |
| SOUNDHALT（停止） | `overlay-02` | `1A88h` | **LOADINTERPET**＋1A88h |
| SOUNDHALT（停止） | `overlay-02` | `1ABAh` | **LOADINTERPET**＋1ABAh |
| SOUNDHALT（停止） | `常駐` | `0AC1h` | （這一段前面沒有符號） |
| SOUNDHALT（停止） | `常駐` | `0BA5h` | （這一段前面沒有符號） |
| SOUNDHALT（停止） | `常駐` | `0C01h` | （這一段前面沒有符號） |
| SOUNDHALT（停止） | `常駐` | `0C37h` | （這一段前面沒有符號） |
| SOUNDHALT（停止） | `常駐` | `0C7Bh` | （這一段前面沒有符號） |
| SOUNDOFF（關） | `常駐` | `0ADBh` | （這一段前面沒有符號） |
| SOUNDOFF（關） | `常駐` | `842Ah` | （這一段前面沒有符號） |
| SOUNDOFF（關） | `常駐` | `89FBh` | （這一段前面沒有符號） |
| SOUNDON（開） | `overlay-03` | `03BCh` | **DOPROTECT**＋28Bh |
| SOUNDON（開） | `常駐` | `846Ch` | （這一段前面沒有符號） |
| SOUNDON（開） | `常駐` | `8567h` | （這一段前面沒有符號） |
| SOUNDON（開） | `常駐` | `864Ah` | （這一段前面沒有符號） |
| SOUNDON（開） | `常駐` | `8A21h` | （這一段前面沒有符號） |
| SPELLHITFX（法術命中） | `overlay-24` | `2489h` | **TWINKLE**＋7Fh |
| SWISHFX（揮擊） | `overlay-13` | `193Bh` | **ANYUNDEAD**＋487h |
| SWISHFX（揮擊） | `overlay-13` | `2B81h` | **SHOWARROW**＋10Fh |
| SWISHFX（揮擊） | `overlay-13` | `2C48h` | **SHOWARROW**＋1D6h |
| WHISTLEFX（哨音） | `overlay-13` | `2BB4h` | **SHOWARROW**＋142h |
| WHISTLEFX（哨音） | `overlay-13` | `2C01h` | **SHOWARROW**＋18Fh |

★ **交叉印證**：`MSCPLAY` 的呼叫點正好落在上表那五個改寫 `MUSICNO` 的 overlay 上（`GEN`×2、`overlay-01`、`POSTCOM`、`overlay-18`）——兩次獨立的掃描（資料格寫入 vs 函式呼叫）指到同一組地方。

⚠ 這裡只數**跨 overlay 的 far call**。常駐自己呼叫 `SOUNDX` 的次數不在裡面（那是段內近呼叫，far-call 表看不到），所以是**下界**。

## 沒有任何換曲點選到的曲目

11（地城二）、7（戰鬥）

⚠ 不要直接讀成「這首用不到」：這一支只認 `mov byte [MUSICNO], imm` 這一種形狀。從暫存器或變數寫進去的換曲點看不到——**要下「沒有人選它」的結論得先排除那些形狀**。
