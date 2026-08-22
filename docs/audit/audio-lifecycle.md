# 原機音訊的播放生命週期：原作有幾種動作，remake 對上幾種

由 `cmd/audio-lifecycle-audit` 產生，不要手改。

`cmd/dseg-writers` 盤的是**分母**（誰決定什麼時候該響）；這一份把那些寫入點按**生命週期動作**分類，再拿去對 remake 真的會發出來的動作。差額就是待辦。

⚠ **這是 PC-98 的版面**，名字取自 Borland 除錯符號。DOS 版沒有符號，位址不能直接套（spec 1187）。

⚠ 位元組直掃，不走 far-call 對照表——表比實際少，而**下界看起來和全集一樣合理**。代價是偽陽性，所以每一處都印出所屬常式讓人看得出合不合理。

remake 規則層目前發得出來的動作：`play`、`stop`

⚠ **「發得出來」與「有東西會發」是兩件事**，分兩欄問：規則層有這個動作是**能力**，現在有哪一段劇情會發是**有沒有用到**。混成一格會讓「寫了但沒人呼叫」看起來像做完了。

| 動作 | 格 | 意思 | 原作寫入處 | 規則層發得出來 | pack 有用到 | 由誰負責 |
|---|---|---|---:|---|---|---|
| `select-track` | `MUSICNO` | 選曲：換成哪一首 | 12 | ✅ | ✅ | `State.requestMusicForCurrentBlock`（game-pack 的 music binding） |
| `load-track` | `MUSICNUM` | 曲目編號：驅動程式要載哪一份資料 | 1 | ✅ | ✅ | 同上：remake 的 `TrackID` 就是曲目，載入由 adapter 負責 |
| `stop-track` | `MUSICNUM` | 停止：曲目編號寫 `255`（沒有曲子） | 2 | ✅ | — engine 的 pack 驗證不收 `track_id` 空的 binding，「這裡不放音樂」目前**表達不出來**（`TestEnginePackCannotExpressStopYet`） | `State.requestMusicForCurrentBlock`：binding 的 `TrackID` 是空的就發 `stop` |
| `music-switch` | `MUSICSW` | 音樂開關：整個音樂要不要響 | 2 | ✅ | — engine 的 pack 驗證不收 `track_id` 空的 binding，「這裡不放音樂」目前**表達不出來**（`TestEnginePackCannotExpressStopYet`） | 同上；開關的「開」由一般的選曲表達 |
| `sfx-halt` | `SOUNDHALT` | 音效停止 | 0 | ✅ | ✅ | 原作一處都沒寫到這一格 |
| `sfx-off` | `SOUNDOFF` | 音效關 | 0 | ✅ | ✅ | 原作一處都沒寫到這一格 |
| `sfx-on` | `SOUNDON` | 音效開 | 0 | ✅ | ✅ | 原作一處都沒寫到這一格 |

合計 17 處寫入、7 種動作：規則層發得出 **7** 種，game-pack 真的會發的有 **5** 種。

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
