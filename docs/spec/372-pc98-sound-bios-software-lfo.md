# 第三百七十二輪 PC-98 Sound BIOS 軟體 LFO 核心

狀態：`READY`（限 `SOUND.ROM` timer ISR、六種 waveform、pitch／total-level
投影及現有十二首 S98 的可觀測性邊界）

後續更新：Timer B cadence、sync delay 與 ROM 動態 harness 已由
[第 373 輪規格](./373-pc98-sound-bios-lfo-timer-scheduler.md)補齊；本檔
保留第 372 輪當時的可觀測性邊界，不再是 scheduler 的最新狀態。

本規格接續第 371 輪，還原 NEC PC-9801 Sound BIOS 在 YM2203 硬體之外
執行的軟體 modulation。它不是 MSCDRV stream opcode `90h`：

- MSCDRV FM `90h` 只在 tempo 小於 `C9h` 時加 4，再寫 register `26h`；
- LFO 由 `SOUND.ROM` 自己的 timer ISR 推進；
- `SETPARABLOCK` 會自動啟用該聲道 modulation；
- `MODUON`／`MODUOFF` 只切換 Sound BIOS 聲道狀態 bit 7。

本輪已完成可重用的 waveform 與輸出數學，但尚未把 LFO scheduler 接進
`TrackPlayback`。原因是現有十二首約五秒 Hoot S98 雖含 18 個非零 LFO
參數聲道，卻沒有記錄到獨立的動態 pitch／total-level 寫入；timer cadence
與 Hoot 執行環境仍缺外部交叉驗證。不得以靜態公式代替這項時間證據。

## 1. 證據與 IDA 校正

輸入仍是第 371 輪的 16 KiB `SOUND.ROM`：

```text
sha256 = f05b508d49f31f2a1a61724f013572592abc0833c09c45a72180e84247dc0d0d
load   = CC000h..CFFFFh
CPU    = 8086 / 16-bit
```

分析必須在 Docker 內優先使用
`/home/anr2/ida_94_official/dist` 對應的 IDA Pro 9.4。timer entry 在
`CF47Ah` 透過 `jmp bx` 跳到三個 register-held near offset，因此一般
auto-analysis 只會把後段顯示成 `db`，不能把「沒有函式」當成沒有 LFO。

[`scripts/ida/pc98_sound_rom_tone_audit.idc`](../../scripts/ida/pc98_sound_rom_tone_audit.idc)
現會主動建立這些 16-bit code path：

| 位址 | 作用 |
|---:|---|
| `CF47Ah` | timer ISR entry 與 YM status dispatch |
| `CF4C3h` | 共用 interrupt 收尾 |
| `CF501h` | note／length 與 callback timer path |
| `CF5F3h` | 六聲道軟體 LFO path |
| `CF7EFh` | 計算有效 pitch depth |
| `CF817h` | FM `A4/A0` pitch 輸出 |
| `CF847h` | 四個 operator `40h` total-level 輸出 |
| `CFAB9h` | signed WORD phase accumulator |

ISR 讀 YM2203 status 低兩位後使用 near offset `06C3h`、`0701h`、
`07F3h`，對應上表的 `CF4C3h`、`CF501h`、`CF5F3h`。這是先前 IDA
遺漏 code path 的直接原因。

## 2. 啟用、停用與參數

每聲道 Sound BIOS state 是 32 bytes。與 LFO 直接相關：

| state offset | 語意 |
|---:|---|
| `+14h` | note／key state 與 bit 7 狀態 |
| `+18h` | bit 7 modulation enable；低兩位為內部 phase state |
| `+19h` | delay／phase counter WORD |
| `+1Bh` | 有效 pitch depth WORD |
| `+1Dh` |目前基準 pitch WORD |

`SETPARABLOCK` 完成 tone render 後會：

```text
state.flags |= 83h
effective_pitch_depth =
    parameter.LFO_PITCH_COARSE * parameter.LFO_PITCH_DEPTH
```

也就是 parameter logical field `35 × 25`。`SETVOLUME` 更新 carrier 後
也會重算有效 pitch depth。

`MODUON (AH=1Bh)`：

```text
state.flags |= 83h
重新計算 effective_pitch_depth
```

`MODUOFF (AH=1Ch)`：

```text
state.flags &= 7Fh
```

MSCDRV binary 只有 library wrapper 內各一個 `AH=1Bh/1Ch, INT D2h`；
正常 timer／stream 路徑沒有 caller。`MSCD_98.COM` 也沒有直接呼叫 pattern。
因此「MSCDRV 每首曲目主動關掉 LFO」不受 bytes 支持。

## 3. 六種 waveform

Sound BIOS 以 signed 16-bit `phase`、`step` 保存每聲道 oscillator。
waveform 4 reset 為 `phase=32767, step=-1`；其餘 reset 為
`phase=0, step=1`。

共用 accumulator `CFAB9h`：

```text
next = int16(phase + int16(speed) * step)
if sign_bit(phase) != sign_bit(next):
    保留舊 phase，回報 boundary crossing
else:
    phase = next
```

六種函式：

| parameter waveform | 位址 | 行為 |
|---:|---:|---|
| 0 | `CFA49h` | crossing 時反相 phase |
| 1 | `CFA67h` | 依 shared timer／period 交替 `32767`、`-32768` |
| 2 | `CFA10h` | crossing 時反相 step，形成三角往返 |
| 3 | `CFA8Ah` | shared signed noise state，以 `×899 mod 32767` 更新 |
| 4 | `CFA55h` | 負值時停止並歸零 |
| 5 | `CFA24h` | crossing 反相 step；下降至非正值時停止 |

waveform 3 的 period 是：

```text
period = (65535 / speed) | 1
```

只有 shared timer counter 對 period 的 signed remainder 為零時才更新
noise state。所有 16-bit wrapping、signed division 與 boundary quirk 都是
原 ROM 行為，不應用浮點正弦波取代。

## 4. Pitch 與 total-level 投影

### 4.1 Pitch

`CF817h` 對 FM channel 寫 `A4+channel`、`A0+channel`。`CF5F3h` 的兩次
signed division：

```text
scaled = sample * effective_pitch_depth / 32767
delta  = scaled * base_fnumber / 32767
pitch  = uint16(base_fnumber + delta)
```

### 4.2 Total level

`CF847h` 依 operator offset `0,8,4,12`，也就是 logical
`1,2,3,4` 寫入實體 register slot `1,3,2,4`。每個 operator：

```text
amplitude = int8(operator_coarse * amplitude_depth / 15)
wave      = sample * uint8(amplitude) / 32767
delta     = int8(wave) * int8(base_total_level) / 127
TL        = byte(base_total_level + delta)
```

中間的 `uint8(amplitude)`、`int8(wave)` 與 byte wrapping 是 8086
`IMUL/IDIV` operand width 的結果，看似不自然但必須保留。`base_total_level`
仍是第 371 輪證明的 `127-OUTPUT_LEVEL`，logical operator 要先按
`1,3,2,4` 映射。

## 5. 實作邊界

獨立 engine commit `77683a3` 新增 `audio/pc98soundbios`：

- `Oscillator` 與六種 `Advance`；
- shared timer／noise 輸入；
- exact signed 16-bit phase crossing；
- `Pitch`；
- 8086 byte-stage `TotalLevel`。

CoAB `FMParameterBlock.YM2203Modulation` 只負責把本作 runtime-import 的
NEC 50-WORD parameter 映射到共用核心：

- `LFO_PITCH_DEPTH × LFO_PITCH_COARSE`；
- `LFO_AMPLITUDE_DEPTH`；
- 四個 `LFO_AMPLITUDE_COARSE`；
- logical→physical operator order。

這樣 engine 不含本作音色值、driver offset 或劇情資料，其他使用 NEC
Sound BIOS 的 PC-98 Gold Box 可沿用同一核心。

## 6. S98 可觀測性稽核

engine `audio/s98.YM2203SoftwareLFOUpdates` 現分開抽取：

- 不被 `28h` key-off／key-on 包圍的 `A4/A0` pitch pair；
- 不屬於完整 tone load 的四筆 standalone TL burst。

十二首修正版 S98 重新稽核後：

```text
first-stream 非零 LFO parameter 聲道：18
observed_lfo_pitch_updates：每首皆 0
observed_lfo_level_updates：每首皆 0
dynamic_lfo_observed：每首皆 false
```

這個結果只證明「目前約五秒 Hoot capture 沒記錄到獨立軟體 LFO update」。
它不能證明：

- LFO 在原 PC-98 遊戲中關閉；
- Hoot 不執行 LFO；
- waveform 公式的 wall-clock cadence；
- key-on sync delay 的實際毫秒數。

下一輪外部驗證應優先選用 parameter 3 的聲道，因為它同時有非零 pitch、
amplitude depth 與四個 operator coarse depth；並延長 capture、確認該聲道
確實 key-on。若 Hoot 仍無動態 write，改用 NP2kai／DOS test harness 直接
呼叫 Sound BIOS，而不是杜撰 scheduler。

## 7. 完成與未完成

本輪已完成：

- timer ISR 間接 code path 的 IDA 恢復；
- `SETPARABLOCK`／`MODUON`／`MODUOFF` 狀態語意；
- 六種 waveform 與 phase accumulator；
- pitch／total-level exact integer projection；
- 可重用 engine 核心與 CoAB parameter adapter；
- S98 動態 LFO extractor；
- 十二首 18 個非零參數聲道與零動態觀測的可重現稽核。

仍未完成：

1. 外部 runtime 的 LFO tick cadence；
2. key-on sync delay 與 phase reset 的完整時間驗證；
3. `TrackPlayback` 與 YM2203 player 的 LFO scheduler；
4. fade、SFX／BGM 共存；
5. 完整曲長／loop；
6. 合成器、PCM mixer、遊戲內播放及 save/resume。

因此本規格不能支撐「LFO 播放完成」或「PC-98 音樂完成」。
