# 376：PC-98 YM2203 合成器與遊戲播放器

狀態：`READY`

## 1. 範圍與結論

本規格把第 366–375 輪已證實的 PC-98 音序列路徑接成實際 PCM：

```text
game-pack track ID
  → reference selector
  → MSCDRV TrackPlayback
  → Sound BIOS parameter／volume 展開
  → YM2203 register writes
  → ymfm
  → 44.1 kHz stereo PCM
  → Ebiten audio.Player
```

這是第一個能由使用者本機 `MSCDRV.EXE` 在 remake 內合成背景音樂的
end-to-end 路徑，不提交 driver、音序列、S98 或輸出 WAV。

它仍不是完整 PC-98 音訊還原：fade、SFX 共存、完整曲長／loop、
save／resume、硬體類比混音增益與 register `27h` reload 的
free-running divide-by-16 phase 尚未完成。

## 2. 證據與分類

| 項目 | 證據 | 分類 |
|---|---|---|
| 3,993,600 Hz、prescale 6、Timer B 完整週期 | spec 374／375、45.01 秒 S98、`ymfm` OPN core | `exact`（完整 count period） |
| FM 音色寫入順序、rate／level 反相、operator order | spec 369–371、十二首 S98 | `exact` |
| algorithm carrier volume redraw、operator-mask key-on | spec 371、72 組啟動序列 | `exact` |
| 正常 BGM 不執行 Sound BIOS LFO | spec 373／374 的 ROM harness、MSCDRV ISR | `exact` |
| YM2203 數位合成核心 | `ymfm` commit `81aec25ccbb98f4873a255f7551ac4dadac59b4a` | 上游 emulator |
| 166,400 Hz → 44,100 Hz 線性重取樣 | 新增 `audio/pcm.LinearResampler` | `reconstructed` |
| 四輸出等權平均、雙聲道 mono | 避免 int16 clipping 的暫定 mixer | `reconstructed` |
| `27h` reload sub-period phase | 尚未模擬 | `unknown` |

`ymfm` 來源固定於上述 commit，BSD-3-Clause 授權全文與 provenance 保存在
engine `audio/ym2203/ymfm/`。它不含本作商業資料。

## 3. 實作契約

### 共用 engine

- `audio/ym2203/ymfm`
  - 薄 C API／Go 封裝；
  - register address/data write；
  - YM2203 minimum-fidelity native samples；
  - cgo 關閉時以明確錯誤失敗，不靜默播放錯誤替代音色。
- `audio/pcm.LinearResampler`
  - 整數有理數 phase；
  - chunk boundary 不改變輸出；
  - 不使用浮點累積時間。
- `engine.Pack.FindMusicTrack`
  - 由穩定 track ID 取得作品 pack selector，不在 renderer 猜 ID。

### CoAB adapter

- `YM2203EventRenderer`
  - 將 `SETPARABLOCK`／`SETVOLUME` 依 S98 證實順序展開；
  - descriptor 的高索引初始化不捏造音色；本作會在首次 key-on 前以
    `0..19` 可驗證音色覆蓋。
- `TrackPCMStream`
  - 一次 `Tick()` 等於 MSCDRV 一次 Timer B overflow；
  - 先產生目前 period PCM，再套用 tick 事件與下一 period tempo；
  - 輸出 16-bit little-endian dual-mono stereo。
- `internal/sound.Player`
  - `PlayPC98Track` 原子替換目前背景曲；
  - WAV 音效缺檔不再阻止 PC-98 音樂 context 建立。
- `cmd/azure-bonds-game`
  - `-pc98-music-driver MSCDRV.EXE` 啟用；
  - game-pack 的 `reference_selector` 是唯一 track ID → driver bridge。
- `cmd/pc98-render-track`
  - 可在不啟動 GUI 的情況下輸出本機 WAV，供 deterministic 稽核。

## 4. 真實媒體驗證

輸入：

- 殘缺但音序列完整的 `MSCDRV.EXE`
- SHA-256：
  `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5`
- selector：`5`
- 輸出：10 秒、44,100 Hz、stereo、signed 16-bit PCM

Docker 內連續執行兩次：

```sh
go run ./cmd/pc98-render-track \
  -duration 10s MSCDRV.EXE 5 /tmp/coab376-selector5.wav
```

兩次 WAV 均為：

`fded75fe89d5e5af860e92e1541f83f14738c228fe7d792506c282c6bd5847c0`

FFmpeg `astats`：

- 441,000 samples／channel；
- minimum `-418`、maximum `2873`；
- peak `-21.142021 dB`；
- RMS `-29.980606 dB`；
- 10,524 zero crossings；
- 無 int16 clipping；
- 左右聲道相同。

輸出 WAV 保留在本機 `/tmp`，不進 Git。

正常玩家路徑另在 Docker／Xvfb／ALSA null device 啟動正式
`cmd/azure-bonds-game -opening`，同時傳入 exact driver 與
`-pc98-music-driver`。程式成功消耗開場 game-pack `MusicEvent`、建立
Ebiten audio player 並輸出 deterministic frame 後正常離開；本機 frame
SHA-256 為
`8db91ad32cd01eda50a40de600f5b5def6aff3024a51cdf58410999bd798ff6a`。
這證明正式 app path 可建立播放器；音樂內容本身仍由上述 WAV／PCM 稽核
證明，靜態畫面不能替代音訊 oracle。

## 5. 自動驗證

- engine `go test ./...`：通過。
- `CGO_ENABLED=0 go test ./...`：必須通過，ymfm backend 明確不可用。
- CoAB deterministic synthetic PCM：
  `4dba2117508462b1f49cb7f3c4d7b935519629655d454e8224c8fd604d263677`。
- ymfm PSG fixture：
  `3731c84ec67f05b0b731ebd86e5421211e821799089b8277464b8ce93b2ae5e6`。
- 正式 CoAB package tests 必須在 Docker／Xvfb 通過。

## 6. 下一步

1. 取得各 selector 完整曲長與 loop register trace，實作停止／loop 契約。
2. 反組譯 MSCDRV fade 與 SFX arbitration；反組譯仍優先使用指定 IDA Pro。
3. 量測 Hoot／NP2kai 的 mixer level，取代四輸出等權平均的 reconstructed
   gain。
4. 補 `27h` reload phase 與 CPU I/O timing，再評估 cycle-exact 差異。
5. 完成實際遊戲長時間播放、切換場景、save／resume 與三平台發行驗證。
