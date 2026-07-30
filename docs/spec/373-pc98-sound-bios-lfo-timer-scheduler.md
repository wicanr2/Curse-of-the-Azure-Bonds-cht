# 第三百七十三輪 PC-98 Sound BIOS LFO Timer B 排程

狀態：`READY`（限 Timer B cadence、sync delay 狀態機與 ROM 動態 harness）

後續修正：本規格正確描述 `SOUND.ROM` scheduler 本身；但
[第 374 輪規格](./374-pc98-mscdrv-timer-b-ownership.md)已證明 CoAB 的
`MSCDRV` 接管硬體中斷且不鏈回 Sound BIOS ISR。故 CoAB faithful BGM
不得把此 scheduler 接進 `TrackPlayback`。本檔第 2、7 節所稱 Hoot
可觀測性限制與整合下一步，均由第 374 輪的新證據取代。

本規格接續第 372 輪。第 372 輪已還原六種 waveform 與 pitch／total-level
投影，但尚未證明何時推進 oscillator。本輪以三層證據補齊：

1. Hoot 長時間 S98，確認一般 `pc98dos` 曲目 trace 的可觀測邊界；
2. 指定 IDA Pro 9.4 對 `SOUND.ROM` timer dispatch 與 sync state machine
   的 8086 反組譯；
3. 在 Docker 內以 Unicorn 直接執行原始 ROM bytes，動態觸發 80 次
   Timer B LFO path。

本輪仍不代表完整 PC-98 音樂播放器；fade、SFX/BGM 共存、完整 loop、
YM2203 合成器、PCM mixer 與遊戲內播放仍未完成。

## 1. 輸入與工具

| 輸入 | SHA-256 |
|---|---|
| `SOUND.ROM` | `f05b508d49f31f2a1a61724f013572592abc0833c09c45a72180e84247dc0d0d` |
| `MSCDRV.EXE` | `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5` |

反組譯依 `AGENTS.md`，在 Docker 內優先使用
`/home/anr2/ida_94_official/dist` 建立的 IDA Pro 9.4 ver2 映像。ROM、
driver、S98、IDA database 與執行 trace 只留在本機。

動態 harness 使用 Unicorn 2.1.4 的 8086 mode，映射本機
`SOUND.ROM`，並攔截 `OUT 188h/18Ah`。可重現入口是
[`scripts/research/pc98_sound_rom_lfo_harness.py`](../../scripts/research/pc98_sound_rom_lfo_harness.py)；
harness 不重寫 waveform：

- 直接執行 ROM `SETPARABLOCK CF309h`；
- 直接執行 ROM `SETVOLUME CF41Fh`；
- 直接執行 ROM `NOTE CF239h`；
- 直接執行 ROM `MODUON CF3E7h`；
- 為每個 Timer B tick 建立原 ISR 的 register／IRET stack frame；
- 執行 ROM `CF5F3h` 到 `CF4C3h` 共用收尾。

## 2. 長時間 Hoot S98 的負面邊界

選用 selector 9（盜賊公會）是因為三個 FM 聲道 first-stream 都載入
parameter 3；該參數同時具有非零 pitch 與 amplitude modulation。

45.01 秒 S98：

```text
sha256 = cd0d834ed7f8f74b1a3934fbacfeaa87101ae142d64a1a18a13a353875f4c5b5
bytes = 72749
duration_ticks = 4501（S98 1/100 秒）
register_writes = 22743
first_key_ons = 3
lfo_parameter_channels = 3
observed_lfo_pitch_updates = 0
observed_lfo_level_updates = 0
```

register 分布另證明三聲道的 `A4/A0` 數量只等於正常 note burst：

| 聲道 | key-on | `A4/A0` pair |
|---:|---:|---:|
| 0 | 265 | 267 |
| 1 | 345 | 347 |
| 2 | 155 | 157 |

多出的兩組是初始化，不是持續 modulation。這證明目前 Hoot `pc98dos`
曲目 path 沒有記錄 Sound BIOS Timer B 軟體 LFO。第 373 輪當時不能從
零結果判斷原因；第 374 輪已由 MSCDRV ISR ownership 證明這是 CoAB 正常
BGM 不鏈回 Sound BIOS 的結果。

另建立的 Hoot COM shell probe 可載入自訂 archive，但同樣只看到 driver
初始化，沒有觸發 ROM Timer B ISR。這項 probe 本身仍不能代表所有 PC-98
程式；CoAB 的結論改由第 374 輪 MSCDRV exact control flow 支持。

## 3. IDA Timer B dispatch

IDA 恢復的 `CF47Ah` timer ISR：

```text
CF494  IN      AL,188h
CF498  AND     AL,3
CF4A3  TEST    AL,1
CF4A5  JNZ     CF501h
CF4A7  TEST    AL,2
CF4A9  JNZ     CF5F3h
```

YM2203 status bit 0 是 Timer A，進入 note／length path `CF501h`；
status bit 1 是 Timer B，進入軟體 LFO path `CF5F3h`。因此 LFO cadence
不是固定 10 ms，也不是 MSCDRV opcode `90h` 自身；它是每次 YM2203
Timer B overflow 推進一次，實際 wall-clock period 隨 register `26h`
tempo 值改變。

## 4. Sync delay exact 狀態機

每聲道 `state.flags +18h` 低兩位是 phase state，bit 7 是 enable；
`state.counter +19h` 是 WORD counter。

### 4.1 Restart

`SETPARABLOCK`、`MODUON` 與新 note 會使低兩位進入 state 3。
Timer B path 在 state 3 讀 parameter word `0Fh`（`LFO_SYNC_DELAY`）：

```text
delay = uint8(sync - 1)
if delay == 0:
    立即 reset oscillator，進入 state 2
else if delay == FFh:
    使用 shared phase
else:
    counter = uint16(delay) * 4
    走共用 counter++，進入 state 1
```

### 4.2 Waiting

state 1 每個 Timer B tick 只遞減 counter 的低 byte；低 byte 成為零時才
reset oscillator 並在同一 tick 產生第一組輸出。等待分支不執行共用
counter++。

### 4.3 Active

state 2 每個 Timer B tick：

1. 以目前 counter 計算 waveform；
2. 輸出 FM pitch `A4/A0`；
3. FM 聲道輸出四個 TL registers；
4. 執行共用 `counter++`。

disabled、note state 為零或 speed 為零時不輸出；ROM 仍會依分支重設
phase state 或推進 counter，不能簡化成單一 `enabled` 布林。

## 5. ROM 動態結果

動態 harness 使用本作 parameter 3：

```text
waveform = 1
sync = 8
speed = 8870
pitch depth = -2
pitch coarse = 15
amplitude depth = 2
```

直接執行 ROM command 後：

```text
note_state = FFh
modulation_flags = 83h
phase_counter = 0
effective_pitch_depth = FFE2h
```

連續觸發 80 次 `CF5F3h`：

```text
first_pitch_tick = 30
first_level_tick = 30
pitch writes = 102（51 組 A4/A0）
level writes = 204（51 組 × 4 operators）
final phase_state = 2
final phase_counter = 51
```

sync 8 的第一個輸出落在第 30 tick，與 ROM exact state machine 一致：

- tick 1：`(8-1)*4 = 28`，再走共用 increment 成 29；
- tick 2..29：低 byte 由 29 減到 1；
- tick 30：低 byte成 0、reset oscillator、立即輸出；
- tick 30..80：共 51 組輸出，counter 最後為 51。

這是直接執行原 ROM bytes 的動態證據，不是 Go 公式的自我驗證。

## 6. Engine contract

獨立 engine 的 `audio/pc98soundbios` 排程器應：

- 由呼叫端每次 Timer B overflow 呼叫一次；
- 保存 enable、低兩位 phase state、WORD counter 與 oscillator；
- exact 保留 sync `0／1／2+` 三分支；
- exact 保留 state 1 只遞減 counter 低 byte；
- active tick 回傳一個 signed sample，由既有 `Pitch`／`TotalLevel` 投影；
- 不含 CoAB parameter bank、selector、曲名或 driver offset。

CoAB adapter 才負責把 `FMParameterBlock` 映成 scheduler config，並將
sample 套用到目前 F-number／TL。

## 7. 完成與未完成

本輪完成：

- 45 秒、三聲道 active parameter 3 的 Hoot 可觀測性邊界；
- Timer A／Timer B dispatch 的 IDA 校正；
- sync delay 三態 state machine；
- 原始 ROM 的 80-tick 動態執行；
- sync 8 第一輸出 tick 30 與 51 組輸出的外部驗證。

仍未完成：

1. 不將 scheduler 接進 CoAB faithful `TrackPlayback`；其他使用 Sound
   BIOS ISR 的軟體仍需獨立 adapter；
2. Timer B register `26h` 到 PCM wall-clock 的 sample-accurate bridge；
3. fade、SFX/BGM 共存與 track transition；
4. 完整曲長、loop、pause/save-resume；
5. 遊戲內播放與三平台音訊驗收。

因此本規格只能支持「Timer B LFO scheduler contract READY」，不能支持
「PC-98 音樂完成」或「遊戲音樂已可播放」。
