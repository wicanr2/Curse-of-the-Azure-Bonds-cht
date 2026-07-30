# 377：PC-98 換曲延遲、循環與內部音效邊界

狀態：`READY`（限 `GAME.EXE MSCPLAY` 正常路徑、MSCDRV loop counter，
以及未被正常遊戲路徑使用的 driver fade／FM SFX 邊界）

## 1. 結論

PC-98 版正常換曲不是淡出。`GAME.EXE MSCPLAY` 會：

1. 把 1-based selector 減一並略過相同曲目；
2. 呼叫 `MSCSTOP`，經 IVT `7Eh` 立即停止舊曲；
3. 呼叫 Borland `DELAY(0x0320)`，等待 800 毫秒；
4. 才經 IVT `7Eh` 播放新 selector。

MSCDRV 另有 40 個 Timer B tick 的淡出與單聲道 FM 音效 interpreter，
但正常 `GAME.EXE MSCPLAY` 先送 stop，令 active track 歸零，因此不會進入
driver fade。`GAME.EXE SOUNDFX` 呼叫的是 Borland `SOUND`／`DELAY` 路徑，
不是 MSCDRV 的 FM SFX request。現有 driver xref 只看到 request 欄位清零，
沒有找到正常遊戲可送出非零 request 的 public ABI；故該 FM SFX 路徑保留為
`dormant／unknown caller`，不得接進 faithful runtime。

正常 public play 以 loop count 0 呼叫 driver；初始化會把 0 轉為 `0xFF`。
stream end consumer 對 `0xFF` 不遞減並重設七個 channel cursor，所以十二首
正常 BGM 是無限循環。有限非零值才在四個 FM 或四個 PSG channel 都抵達
結尾時遞減，歸零後停止。

## 2. 輸入與可重現證據

| 輸入 | SHA-256 |
|---|---|
| `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` |
| `MSCDRV.EXE` | `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5` |

主要分析依 `AGENTS.md`，在 Docker 內使用指定
`/home/anr2/ida_94_official/dist` 對應的 IDA Pro 9.4：

- `scripts/ida/pc98_game_music_transition_audit.idc`
- `scripts/ida/pc98_mscdrv_transition_sfx_audit.idc`

兩支原生 IDC 輸出 IDA address、instruction bytes、caller 與 data xref。
GAME 的 MZ segment/file mapping 不是單一線性公式，所以 IDC 不猜 file
offset；`cmd/pc98-music-audit` 另以 raw bytes 驗證：

- `GAME file 0x93E4`：`MSCPLAY_STOP_DELAY_800MS`，44 bytes；
- `GAME file 0x9BF9`：`BORLAND_DELAY_BUSY_LOOP`，45 bytes。

## 3. 關鍵控制流

`GAME 18A44h MSCPLAY`：

```text
18A64  call 18A8Eh       ; MSCSTOP
18A67  mov  ax,0320h
18A6A  push ax
18A6B  call 19259h       ; Borland DELAY
18A70  command=play
18A75  selector=0-based
18A83  call IVT trampoline
```

MSCDRV queue path 若 active track 非零，才把 fade counter 設成 `0x28`。
每個 Timer B tick 遞減；只在奇數結果調低 FM／PSG 音量。正常 GAME wrapper
先 stop，因此 queue 時 active track 已是零。

driver loop counter 初始化：

- 呼叫參數非零：照值保存；
- 呼叫參數為零：保存 `0xFF`；
- FM end 在 `105CCh`、PSG end 在 `10952h` 檢查 `0xFF`；
- 非 `0xFF` 才遞減，歸零呼叫 stop；否則重設七聲道 cursor。

## 4. Remake contract

- `NewGameTrackPCMStream` 在每次新曲前輸出 800ms 靜音；
- 44.1 kHz stereo s16 對應 35,280 frames／141,120 bytes；
- 靜音期間不推進 `TrackPlayback.Tick()`；
- `internal/sound.Player` 使用此 game wrapper；
- `NewTrackPCMStream` 與預設 `pc98-render-track` 仍表示直接 driver path；
- `pc98-render-track -game-transition` 可重現 GAME wrapper 音訊時序。

原版 800ms 是 busy-loop，會阻塞遊戲 CPU；remake 保留聲音起始時序，但不凍結
UI。故「800ms 音訊 pre-roll」為 `exact timing/value`，「UI 持續回應」為
刻意現代化差異，不能宣稱整個 wall-clock execution pixel/cycle exact。

## 5. 驗證

exact driver selector 5、三秒輸出：

```text
pc98-render-track -duration 3s -game-transition MSCDRV.EXE 5 output.wav
```

FFmpeg：

- `0.0–0.8s`：35,280 samples／channel，min/max 皆 0，peak `-inf`；
- `0.8–3.0s`：97,020 samples／channel，min `-418`、max `2873`，
  peak `-21.142021 dB`，已開始發聲。

unit regression 另驗證 141,120 bytes 全零、driver tick 仍為 0，下一段 PCM
非零。WAV、driver、IDA database 與 log 都只留本機。

正常玩家路徑另在 Docker／Xvfb／ALSA null device 啟動正式
`cmd/azure-bonds-game -opening`，傳入 exact driver 與
`-pc98-music-driver`。程式成功消耗開場 `MusicEvent`、建立採用 800ms
pre-roll 的 Ebiten player、輸出第三個 deterministic frame 後正常離開；
本機 PNG SHA-256：
`8ab3e88ed74668788dfb3d37e5d6fdafbccf672de365fe827a933a5213c30fdd`。
畫面只證明正式 app path；音訊內容與時序仍由 WAV／PCM 稽核證明。

## 6. 尚未完成

- MSCDRV 內部 FM SFX request 的真實非零 producer／GAME caller；
- PC speaker `SOUNDFX` 每個 selector 的頻率、節奏與遊戲事件完整 mapping；
- 有限 loop count 是否被其他工具／未取得 caller 使用；
- YM2203 `27h` reload phase、save/resume、類比 mixer gain；
- 全遊戲音效與跨平台長時間播放驗收。

因此本規格只關閉正常換曲 800ms 與正常 BGM 無限 loop 的缺口，不代表音樂或
音效系統完整完成。
