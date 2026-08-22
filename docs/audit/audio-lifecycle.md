# 原機音訊的播放生命週期：原作有幾種動作，remake 對上幾種

由 `cmd/audio-lifecycle-audit` 產生，不要手改。

`cmd/dseg-writers` 盤的是**分母**（誰決定什麼時候該響）；這一份把那些寫入點按**生命週期動作**分類，再拿去對 remake 真的會發出來的動作。差額就是待辦。

⚠ **這是 PC-98 的版面**，名字取自 Borland 除錯符號。DOS 版沒有符號，位址不能直接套（spec 1187）。

⚠ 位元組直掃，不走 far-call 對照表——表比實際少，而**下界看起來和全集一樣合理**。代價是偽陽性，所以每一處都印出所屬常式讓人看得出合不合理。

remake 規則層目前發得出來的動作：`play`、`stop`

⚠ **「發得出來」與「玩家碰得到」是兩件事**，分兩欄問：規則層有這個動作是**能力**，實際有沒有東西會觸發它是**有沒有用到**。混成一格會讓「寫了但沒人呼叫」看起來像做完了。

⚠ 「停止」不是劇情資料。原作沒有「這一段不放音樂」這種宣告——派曲常式（`sub_18AA7`）查不到就 `ret`，音樂繼續放。**唯一**會停的是玩家把音樂關掉（`MUSICSW`，Ctrl+O），所以那一欄問的是「按鍵綁上去了沒」，不是「pack 寫了沒」（spec 1192）。

| 動作 | 格 | 意思 | 原作寫入處 | 規則層發得出來 | 玩家碰得到 | 由誰負責 |
|---|---|---|---:|---|---|---|
| `select-track` | `MUSICNO` | 選曲：換成哪一首 | 12 | ✅ | ✅ | `State.requestMusicForCurrentBlock`（game-pack 的 music binding） |
| `load-track` | `MUSICNUM` | 曲目編號：驅動程式要載哪一份資料 | 1 | ✅ | ✅ | 同上：remake 的 `TrackID` 就是曲目，載入由 adapter 負責 |
| `stop-track` | `MUSICNUM` | 停止：曲目編號寫 `255`（沒有曲子） | 2 | ✅ | ✅ | `State.ToggleMusicSwitch`（Ctrl+O）：關掉時 `stopMusic` 發 `stop` |
| `music-switch` | `MUSICSW` | 音樂開關：整個音樂要不要響 | 2 | ✅ | ✅ | 同上；開關的「開」由一般的選曲表達 |

合計 17 處寫入、4 種動作：規則層發得出 **4** 種，玩家真的碰得到 **4** 種。

## 不在這張表裡的三個名字

`SOUNDHALT`（`4838h`）、`SOUNDOFF`（`483Ah`）、`SOUNDON`（`483Ch`）看起來像三格音訊狀態，**但它們不是狀態**：那是 `SOUNDFX` 的選擇子常數，和 `CASTFX`…`CRASHFX` 排在資料段同一張表裡，值分別寫死成 `255`／`0`／`1`，全程式一處都沒有寫入。

⚠ 這一支本來把它們當狀態格掃，掃出 0 處寫入，然後照「原作沒寫到就不算待辦」的規則印成 ✅ ——**三個假零被當成三項做完的工作**。三個相鄰符號同時掃出零，本身就該當成模型錯了的訊號，不是結論。選擇子的身分與解碼歸 `internal/pc98sfx`；玩家關音效走的是 `SOUNDTYPE := 2`（Ctrl+S），`SOUNDFX` 開頭就擋掉，和音樂開關互不影響。

## 逐處

| 檔案 | 常式 | 位移 | 格 | 形式 | 值 |
|---|---|---|---|---|---|
| `PC98-GAME.EXE` | — | `9486h` | `MUSICNO` | `mov [addr],imm` | `3` |
| `PC98-GAME.EXE` | — | `94AEh` | `MUSICNO` | `mov [addr],imm` | `4` |
| `PC98-GAME.EXE` | — | `94C4h` | `MUSICNO` | `mov [addr],imm` | `6` |
| `PC98-GAME.EXE` | — | `94CBh` | `MUSICNO` | `mov [addr],imm` | `5` |
| `PC98-GAME.EXE` | — | `94E2h` | `MUSICNO` | `mov [addr],imm` | `8` |
| `PC98-GAME.EXE` | — | `94F9h` | `MUSICNO` | `mov [addr],imm` | `9` |
| `PC98-GAME.EXE` | — | `9514h` | `MUSICNO` | `mov [addr],imm` | `12` |
| `overlay-01` | DOINTRO＋6h | `093Ch` | `MUSICNO` | `mov [addr],imm` | `1` |
| `overlay-05` | DOPOSTCOMBAT＋1E0h | `1955h` | `MUSICNO` | `mov [addr],imm` | `2` |
| `overlay-17` | DOGEN＋83Dh | `0B08h` | `MUSICNO` | `mov [addr],imm` | `2` |
| `overlay-17` | DOGEN＋24B1h | `277Ch` | `MUSICNO` | `mov [addr],acc` | — |
| `overlay-18` | FINAL＋47Ah | `168Dh` | `MUSICNO` | `mov [addr],imm` | `10` |
| `PC98-GAME.EXE` | — | `9400h` | `MUSICNUM` | `mov [addr],acc` | — |
| `PC98-GAME.EXE` | — | `2F5Ah` | `MUSICNUM` | `mov [addr],imm` | `255` |
| `PC98-GAME.EXE` | — | `9451h` | `MUSICNUM` | `mov [addr],imm` | `255` |
| `PC98-GAME.EXE` | — | `2F46h` | `MUSICSW` | `mov [addr],imm` | `0` |
| `PC98-GAME.EXE` | — | `8A31h` | `MUSICSW` | `mov [addr],acc` | — |
