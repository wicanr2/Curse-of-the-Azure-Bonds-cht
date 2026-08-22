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

★ **交叉印證**：`MSCPLAY` 的呼叫點正好落在上表那五個改寫 `MUSICNO` 的 overlay 上（`GEN`×2、`overlay-01`、`POSTCOM`、`overlay-18`）——兩次獨立的掃描（資料格寫入 vs 函式呼叫）指到同一組地方。

⚠ 這裡只數**跨 overlay 的 far call**。常駐自己呼叫 `SOUNDX` 的次數不在裡面（那是段內近呼叫，far-call 表看不到），所以是**下界**。

## 沒有任何換曲點選到的曲目

11（地城二）、7（戰鬥）

⚠ 不要直接讀成「這首用不到」：這一支只認 `mov byte [MUSICNO], imm` 這一種形狀。從暫存器或變數寫進去的換曲點看不到——**要下「沒有人選它」的結論得先排除那些形狀**。
