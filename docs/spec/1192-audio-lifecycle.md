# 1192 — 原機音訊的播放生命週期：從分母到覆蓋率

狀態：`READY`（盤點、對照、限制都已釘住；「停」還接不上，原因在共用 engine）

## 問題

`cmd/dseg-writers` 已經盤出「誰決定什麼時候該響」——六格音訊狀態的 17 處寫入。
但那是**分母不是覆蓋率**：它說原作在幾個地方改音訊狀態，沒說 remake 有沒有
對應的動作。

## 生命週期動作

按寫哪一格分（PC-98 版面，名字取自 Borland 除錯符號）：

| 動作 | 格 | 意思 | 寫入處 |
|---|---|---|---:|
| `select-track` | `MUSICNO` `8BF3h` | 選曲 | 12 |
| `load-track` | `MUSICNUM` `8BE1h` | 曲目編號 | 1 |
| `stop-track` | `MUSICNUM` `8BE1h` | **停止**：寫 `255`（沒有曲子） | 2 |
| `music-switch` | `MUSICSW` `8BE3h` | 音樂開關 | 2 |
| `sfx-halt`／`sfx-off`／`sfx-on` | `4838h`／`483Ah`／`483Ch` | 音效開關 | 0 |

⚠ **`stop-track` 沒有自己的格，它是一個「值」。** 只按位址分類會把「停止」算成
「載入」，於是原作明明會停音樂，報表卻說這一類已經接上了。分類要看值。

三處停止的證據（PC-98 常駐）：

```
2F46h  c6 06 e3 8b 00   mov byte [8BE3h], 0     MUSICSW  := 0     初始化：音樂關著
2F5Ah  c6 06 e1 8b ff   mov byte [8BE1h], 0FFh  MUSICNUM := 255   初始化：沒有曲子
9451h  c6 06 e1 8b ff   mov byte [8BE1h], 0FFh  MUSICNUM := 255   派曲：這裡不放音樂
8A31h  a2 e3 8b         mov [8BE3h], al         MUSICSW  := AL    開關，接著 far call
       9a 77 01 93 08   call far 0893:0177      ＝ BGMPLAY
```

`9451h` 落在 `9400h`..`9520h` 這一段——同一段裡有全部 7 處 `MUSICNO` 的寫入，
所以那是**派曲的常式**：依情境挑一首，挑不到就寫 255。

## 「發得出來」與「有東西會發」是兩件事

規則層有這個動作是**能力**，現在有哪一段劇情會發是**有沒有用到**。混成一格會讓
「寫了但沒人呼叫」看起來像做完了，所以報表分兩欄問。

目前：規則層發得出 7 種（`musicEventForTrack` 的 `stop` 分支），game-pack 真的
會發的只有 5 種。

## 為什麼「停」接不上——卡點在共用 engine

不是「還沒有人寫那條 binding」，是**寫不出來**。engine 的 pack 驗證會把
`track_id` 是空的 binding 擋掉：

```
music_bindings[1] references unknown track ""
```

所以「這裡不放音樂」在 pack 裡沒有任何寫法。`TestEnginePackCannotExpressStopYet`
把這個限制釘住——**它一旦紅，代表 engine 鬆綁了**，那時要做的是把「不放音樂」的
binding 補進 game-pack，不是把測試刪掉。

⚠ 這兩種卡點的處置差很多：「沒人寫」補一行 JSON 就好，「寫不出來」要動共用 repo。
報表寫的是後者，而且那句話是**跑出來的**不是手打的（`cmd/audio-lifecycle-audit`
每次都重驗）。

## 為什麼 `stop` 的程式碼先擺著

少了它，空的 `TrackID` 會發成 `play`，然後在 adapter 那裡查無此曲、**只留一行
log**：音樂繼續放著，畫面上什麼都看不出來，測試也照樣綠。

## 這個數字證明不了什麼

- **PC-98 版面**。DOS 版沒有除錯符號也沒有音樂資料，位址不能套（spec 1187）。
- 位元組直掃有偽陽性（剛好長得像 opcode 的立即數），所以逐處都印所屬常式。
  不走 far-call 對照表是因為表比實際少，而**下界看起來和全集一樣合理**（spec 1186）。
- 「接上了」只表示 remake 有對應的動作，**不表示在正確的時機發**——那要實機比對。

## 相關

- `cmd/audio-lifecycle-audit`、`docs/audit/audio-lifecycle.md`
- `internal/game/music_events.go`：`musicEventForTrack`
- `TestMusicEventForTrack`、`TestEnginePackCannotExpressStopYet`
- spec 1183（dseg 掃描）、1186（音效觸發語意）、1187（DOS 音效描述子）
