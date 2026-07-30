# 第三百七十五輪 PC-98 YM2203 Timer B 時鐘與 PCM 有理數排程

狀態：`READY`（限完整 Timer B count period、3,993,600 Hz／prescale 6
證據與無累積小數誤差的 PCM sample accumulator）

本規格接續第 374 輪。第 374 輪已證明 `TrackPlayback.Tick()` 由 MSCDRV
自己的 Timer B ISR 推進；本輪把 register `26h` 的 data 轉成 chip clocks，
並建立不因逐 tick 四捨五入而漂移的 PCM sample 整數排程。

本輪仍未還原 register `27h` reload 時 free-running divide-by-16 的首週期
phase adjustment，也還沒有 YM2203 合成器。因此只能宣稱「完整 count
period 與長時間有理數累加 READY」，不能宣稱 cycle-perfect IRQ edge 或
遊戲內音樂可播放。

## 1. 動態輸入

selector 9 長時間 Hoot S98：

| 欄位 | 值 |
|---|---:|
| SHA-256 | `cd0d834ed7f8f74b1a3934fbacfeaa87101ae142d64a1a18a13a353875f4c5b5` |
| 長度 | 72,749 bytes |
| capture | 45.01 秒 |
| YM2203 device clock | 3,993,600 Hz |
| prescaler writes `2Dh..2Fh` | 0 |
| register `26h` writes | 7 |
| register `27h` writes | 7,131 |

`26h` value 分布：

```text
00h ×2, BAh ×2, D3h ×1, E3h ×2
```

`27h` value 分布：

```text
00h ×2, 02h ×2, 0Ah ×3561, 20h ×3558, 32h ×2, 33h ×6
```

其中大量 `20h`／`0Ah` 與第 374 輪 IDA exact ISR 的 acknowledge／restart
成對吻合。S98 header 的 clock 是 binary metadata，不是由錄音長度反推。
`internal/pc98music.AuditS98Track` 現會版本化輸出上述 timer register 統計；
商業 S98 本身仍只留本機。

## 2. Timer B 完整週期

Yamaha《YM2203 FM Operator Type-N (OPN)》資料表確認：

- `26h` 是 Timer B data；
- `27h` 控制 load、interrupt enable 與 status reset；
- `2Dh..2Fh` 選擇 prescaler。

來源：<https://www.bitsavers.org/components/yamaha/YM2203_198911.pdf>

BSD 授權的官方 ymfm repository commit
`81aec25ccbb98f4873a255f7551ac4dadac59b4a` 提供可執行交叉參考：

- `src/ymfm_opn.h`：YM2203 `DEFAULT_PRESCALE = 6`；
- `src/ymfm_fm.ipp`：Timer B inner period 是
  `16 × (256 − timer_b_value)`；
- timer callback clocks 再乘 `OPERATORS × clock_prescale`；
- OPN 有 12 個 operators。

因此完整 count period：

```text
chip_clocks = 16 × (256 − B) × 12 × prescale
```

CoAB trace 沒有任何 `2Dh..2Fh` write，所以維持 reset prescale 6：

```text
chip_clocks = 1152 × (256 − B)
seconds = chip_clocks / 3,993,600
```

三個實際非零 Timer B data：

| `26h` | chip clocks | 約略毫秒 | 48 kHz 完整週期 samples |
|---:|---:|---:|---:|
| `BAh` | 80,640 | 20.1923 | 969.2308 |
| `D3h` | 51,840 | 12.9808 | 623.0769 |
| `E3h` | 33,408 | 8.3654 | 401.5385 |

表中 sample 數故意保留小數；逐 tick 取 969／623／402 會長時間漂移。

## 3. Engine contract

獨立 engine `audio/ym2203` 新增：

- `TimerBClockCycles(value, prescale)`：
  只接受 YM2203 的 `2／3／6` prescale；
- `TimerBSampleAccumulator`：
  以整數 numerator／remainder 將完整 period 轉成 PCM samples；
- 每次 `Advance` 把未滿一個 sample 的餘數帶到下一 period；
- 不含 CoAB clock、曲目、selector 或 parameter bank。

CoAB adapter 才宣告：

```text
PC98YM2203ClockHz = 3,993,600
PC98YM2203DefaultPrescale = 6
```

在 48 kHz 連續累加 1,000 個 `BAh` periods，engine regression 會比較：

```text
floor(1000 × 80640 × 48000 / 3993600)
```

與逐 period accumulator 的總 sample 數及最終 remainder，確保不累積
浮點或四捨五入誤差。

## 4. Free-running phase 邊界

ymfm 明確指出 Timer B 的 `×16` divider 是 free-running；寫 `27h` reload
時，第一個 period 會依當下 `total_clocks & 15` 有負向 phase adjustment。
MSCDRV 每次 ISR 又會寫 `27h=20h`、稍後寫 `27h=0Ah`，所以精確 IRQ edge
還取決於：

1. ISR 內兩次 register write 間的 8086／I/O 延遲；
2. reload 當下 divide-by-16 phase；
3. 模擬器如何同步 CPU 與 YM2203 chip clocks。

S98 只有 10 ms timestamp resolution，不能證明這個 sub-period phase。
本輪 accumulator 只處理完整 count period 的長時間有理數，不假裝已解決
首週期 phase。

## 5. 完成與未完成

本輪完成：

1. 3,993,600 Hz 動態 header 證據；
2. 正常 trace 無 prescaler write；
3. `26h`／`27h` value distribution；
4. Timer B 完整 count period 公式；
5. 作品中立、無累積小數誤差的 sample accumulator；
6. CoAB clock／prescale adapter。

仍未完成：

1. `27h` reload 的 free-running divide-by-16 phase；
2. CPU I/O delay 與 YM clock 的共同 timeline；
3. YM2203 FM／PSG 合成器與 PCM mixer；
4. fade、SFX/BGM 共存、完整曲長／loop、pause／save-resume；
5. 遊戲內播放與三平台音訊驗收。

因此本規格只能支持「Timer B full-period clock bridge READY」，不能支持
「cycle-perfect Timer B」或「PC-98 音樂已完成」。
