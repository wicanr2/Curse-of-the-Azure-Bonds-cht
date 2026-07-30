# 379：PC-98 SOUNDFX 符號、平台語意與 cycle PCM

狀態：`READY`（限 Borland selector 語意、平台分離的 sound intent、
V30 timing-reconstructed PCM 與遊戲內 one-shot 接線；原機 wall-clock
仍未校準）

## 1. 本輪關閉的問題

第 378 輪只證明 `GAME.OVR` 有 42 個 `SOUNDFX` caller，以及 selector、
音序表和 port `37h` 程式；除腳步外，不替 selector 猜名稱。本輪由
`GAME.EXE` 隨附的 Borland `0x52FB` debug symbols 找到 exact 常數：

| DS address | symbol | unsigned value | remake event |
|---|---|---:|---|
| `0C29:4838` | `SOUNDHALT` | 255 | `stop` |
| `0C29:483A` | `SOUNDOFF` | 0 | 控制／no-op |
| `0C29:483C` | `SOUNDON` | 1 | 控制／no-op |
| `0C29:483E` | `CASTFX` | 2 | `cast` |
| `0C29:4840` | `MISSFX` | 3 | `miss` |
| `0C29:4842` | `SPELLHITFX` | 4 | `spell_hit` |
| `0C29:4844` | `DEADFX` | 5 | `dead` |
| `0C29:4846` | `WHISTLEFX` | 6 | `whistle` |
| `0C29:4848` | `HITFX` | 7 | `hit` |
| `0C29:484A` | `LIGHTNINGFX` | 8 | `lightning` |
| `0C29:484C` | `SWISHFX` | 9 | `swish` |
| `0C29:484E` | `PADFX` | 10 | `step` |
| `0C29:4850` | `FIREBALLFX` | 11 | `fireball` |
| `0C29:4852` | `ARROWFX` | 12 | `arrow` |
| `0C29:4854` | `OVERTUREFX` | 13 | `overture` |
| `0C29:4856` | `COMBATFX` | 14 | `combat` |
| `0C29:4858` | `CRASHFX` | 15 | `crash` |

這證明 DOS WAV adapter 與 PC-98 不能共用數字 selector。例如箭矢在 DOS
reference resource 是 2，在 PC-98 是 12；法術命中在 DOS 是 3，在 PC-98
是 4。State 現只發出上述語意事件，由平台 adapter 各自映射。

## 2. IDA Pro 與 raw bytes 證據

輸入仍是第 378 輪的 exact 檔案：

- `GAME.EXE` SHA-256
  `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`
- `GAME.OVR` SHA-256
  `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a`

主要分析使用 Docker 內、由
`/home/anr2/ida_94_official/dist` 建立的 IDA Pro 9.4。MZ database 的
load base 是 `10000h`，所以 `0C29:4838` 對應 IDA `20AC8h`。
`scripts/ida/pc98_soundfx_symbol_audit.idc` 逐 WORD 輸出：

```text
SOUNDHALT=255, SOUNDOFF=0, SOUNDON=1, CASTFX=2,
MISSFX=3, SPELLHITFX=4, DEADFX=5, WHISTLEFX=6,
HITFX=7, LIGHTNINGFX=8, SWISHFX=9, PADFX=10,
FIREBALLFX=11, ARROWFX=12, OVERTUREFX=13,
COMBATFX=14, CRASHFX=15
```

`cmd/pc98-ovr-audit -soundfx` 另直接解析 Borland symbols 與 TPOV code，
42 個 caller 總數及 selector 分布仍與第 378 輪一致。新增的函式邊界包括：

- `COMSTUFF.REALMOVE`：`PADFX`；
- `COMSTUFF.ANYUNDEAD`：`HITFX／SWISHFX`；
- `COMSTUFF.SHOWARROW`：`ARROWFX／SWISHFX／WHISTLEFX`；
- `SPELLS.CASTSPELL`：`FIREBALLFX／LIGHTNINGFX／CASTFX`；
- `GENERIC.TWINKLE`：`SPELLHITFX／MISSFX`；
- `LOS.SCAN`：`DEADFX`。

符號名稱、WORD 值、caller 的 DS address 和 `SOUNDFX` consumer 四者形成
閉合證據鏈，不再使用 DOS WAV 檔名替 PC-98 selector 猜語意。

## 3. V30 cycle 模型

> 第 380 輪更正：本節最初把 NEC 表的 branch／exit 欄讀反成 `13/5`。
> exact Unicorn trace 與重新逐列核對後，正確值如下；舊 WAV hash 作廢。

NEC《16-BIT V SERIES Instruction User's Manual》的 V30 timing 表證明
opcode `E2h`（`BCWZ／LOOP`）在 counter 非零而分支時是 5 clocks，歸零
離開時是 13 clocks。
`GAME.EXE 19D1Eh..19D5Dh` 每個 gate interval 都使用：

```text
LOOP count times
OUT 37h,06h／07h
```

本輪 profile 在「instruction 已 prefetch、無額外 wait」假設下，依 exact
指令路徑保存：

```text
busy-loop              = 5 × (N - 1) + 13
首次 gate-on interval  = busy-loop + 98 clocks
後續 gate-on interval  = busy-loop + 30 clocks
非末次 gate-off        = busy-loop + 56 clocks
末次 gate-off          = busy-loop + 28 clocks
```

NP2kai 的 `io/sysport.c` 與 `sound/beepc.c` 交叉證明：

- port `37h` value 6 清除 system-port bit 3，開啟 buzzer gate；
- value 7 設定 bit 3，關閉 buzzer gate；
- edge 以目前 CPU clock timestamp 送進 one-shot PCM duty-cycle integrator。

共用 engine `audio/cyclepcm` 因此只接受「clock cycles＋level」區段，以整數
duty-cycle 積分成 PCM；它不知道 CoAB、selector 或 port `37h`。
`internal/pc98sfx` 才負責把 exact `SOUNDFX` step 套入 V30 profile。

### Fidelity 邊界

原作是 CPU busy-loop，音高會隨機型 CPU clock、prefetch、memory／I/O wait
改變。目前 CLI 預設提供 8 MHz profile，但仍是
`timing-reconstructed`，不是原機 wall-clock exact。若未來 NP2kai trace
或原機錄音證明不同時鐘／wait，應更換 profile，不得改寫原 WORD 為假 Hz。

## 4. Runtime 接線

- `game.SoundEvent` 改為平台中立字串事件；
- `sound.DOSID` 保留既有 DOS WAV selector mapping；
- `pc98sfx.SelectorForEvent` 保存 Borland exact PC-98 mapping；
- `sound.Player.LoadPC98Effects` 從使用者本機 exact `GAME.EXE` 建立 one-shot
  players，與 YM2203 music player 共用 Ebiten audio context；
- `sound.Player.PlayEvent` 在 PC-98 backend 啟用時不再回退 DOS 數字；
- `cmd/pc98-render-sfx` 可獨立輸出 WAV；
- 正式遊戲以
  `-pc98-sfx-game GAME.EXE -pc98-sfx-clock 8000000` 啟用。

商業 executable、音序 bytes、IDA database、log 與產生的 WAV 仍只留本機。

## 5. 可重現驗證

8 MHz profile、44.1 kHz：

| effect | frames | duration | SHA-256 |
|---|---:|---:|---|
| `ARROWFX` | 1,865 | 0.042290 s | `b9fc898253a380679e84c2026c84c9725a30302dbb15f7814a411664f2f50a5a` |
| `FIREBALLFX` | 736 | 0.016689 s | `b8922db10390746d5bb5b06f28385c0ca779a17f35bc01fef29a3bbeea7c5be8` |

箭矢連續兩次輸出 hash 相同；FFmpeg 證明它是
44.1 kHz／stereo／signed 16-bit PCM，peak `-14.7 dB`，不是靜音。

正式 `azure-bonds-game -opening -pc98-sfx-game ...` 已在
Docker／Xvfb／ALSA null device 載入 backend、跑進一般開場路徑並輸出
640×480 screenshot；hash
`8ab3e88ed74668788dfb3d37e5d6fdafbccf672de365fe827a933a5213c30fdd`。

## 6. 仍未完成

- 原機／NP2kai port `37h` edge trace 對 8 MHz profile 的 wall-clock 校準；
- I/O／memory wait、prefetch queue miss 與不同 PC-98 機型 profile；
- 原版音量、類比濾波與 YM2203＋speaker mixer gain；
- 每種 effect 的原版錄音／影片時間碼與聽感對照；
- `SOUNDON／SOUNDOFF` 的完整設定頁玩家路徑；
- save/resume 時 one-shot 與 music phase；
- MSCDRV dormant FM SFX 的未知外部 producer；
- DOS PC Speaker／Tandy／AdLib 的完整平台 backend。

所以本規格證明「PC-98 selector 已正確命名、可由 exact 程式重建並在遊戲內
播放」，不證明全平台音效或原機 cycle-perfect audio 已完成。
