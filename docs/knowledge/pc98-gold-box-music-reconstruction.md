# PC-98 Gold Box 音樂重建知識庫

本文件整理《青色枷的詛咒》PC-9801 版音樂與短音效的證據鏈，讓後續
Gold Box 作品沿用方法，但不把本作位址、曲目或商業音序誤寫成共用引擎。

## 1. 不是「抓出 MIDI」就完成

正常配樂跨越：

```text
ECL／場景狀態
  → GAME.EXE BGMPLAY／MSCPLAY
  → IVT 7Eh
  → MSCDRV.EXE stream bytecode
  → YM2203 register writes
  → FM／PSG synthesis
  → PCM mixer
```

正常短音效另走：

```text
GAME.OVR caller
  → GAME.EXE SOUNDFX
  → Borland SOUND／DELAY
  → PC-9801 port 37h gate
  → software-speaker PCM
```

MSCDRV 雖有 dormant FM SFX interpreter，目前沒有正常 GAME producer。
看見程式碼不等於玩家路徑使用，不能混合兩條音效路徑。

## 2. 原始媒體與缺 sector

Disk 1 的 absent sector 落在 `MSCDRV.EXE 0x4000..0x43FF`，所以不能宣稱
驅動完整復原。但十二首 descriptor 與 84 個 channel sequence 位於
`0x1B61..0x3C58`，沒有跨缺口。匯入器必須驗證 exact SHA、每段 bounds，
保留 absent 狀態，且不提交商業 sequence bytes。

## 3. Wrapper、曲目與換曲

Borland symbols 證明 `MSCPLAY 0893:0114`、`MSCSTOP 0893:015E`、
`BGMPLAY 0893:0177`。正常換曲固定是
`stop → DELAY(0x320) → play`，即 800 ms 靜音，不走 driver fade。
loop count 0 會轉成 `0xFF`，形成無限循環。

曲名與場景不能只靠 selector 猜測；本作交叉使用：

1. `CURRENTECL` writer／consumer；
2. decoded ECL 與正常玩家路徑；
3. Hoot `ponyca.xml` 的 0-based Shift-JIS metadata。

作品 JSON 保存 stable track ID、reference selector、中文標題與場景 cue；
共用 engine 只保存 `music_tracks／music_bindings／music_cues` contract。

## 4. Driver stream bytecode

十二首共有七聲道、84 組 stream。`sub_10410` 的 framing 必須依 FM／PSG／
timing family 判讀：

- `A0–A4` jump／call／loop；
- 16-entry 原版 stack；
- overflow／underflow 依原程式 no-op；
- timing channel 採 bounded read-through，不套一般 descriptor end gate。

## 5. YM2203 音色與 register order

FM 音色表位於 `seg003:0542`／file `0x45A2`，採 NEC 50-WORD parameter
block。高索引只在 descriptor 初始化短暫載入；第一個 stream `85h` 會在
首次 key-on 前改用內嵌 `0..19`，但不可刪除初始化副作用。

Hoot S98 與 `SOUND.ROM` consumer 共同證明：

- NEC rate／level 需反相；
- operator 順序是 `1,3,2,4`；
- signed DETUNE 採 8-bit left shift；
- `TL = 127 - OUTPUT_LEVEL`；
- algorithm carrier 數量為 `4→2→3→1`；
- key-on 使用 `OPERATOR_MASK`，不能固定寫 `F0h`。

## 6. Timer B 與 LFO

S98 證明 YM2203 clock 是 3,993,600 Hz、prescale 6。Timer B 完整 period：

```text
16 × (256 - value) × 12 operators × prescale
= 1152 × (256 - value) chip clocks
```

sample accumulator 必須保存 remainder。`27h` reload 的 free-running
divide-by-16 phase 尚未完成，所以目前不是 cycle-perfect IRQ edge。

`SOUND.ROM` 的六種 LFO 與 cadence 已由 exact ROM harness 證明；但本作
MSCDRV 接管 YM2203 IRQ，不鏈回 Sound BIOS ISR，所以忠實 BGM 不啟用該
LFO。功能存在不代表本作正常路徑使用。

## 7. Register 到 PCM

作品 adapter 展開 verified register order；獨立 engine 使用固定
BSD-3-Clause `ymfm` 合成，再以整數有理數 phase 重取樣。驗收至少包含：

- 同 selector 兩次輸出 hash 相同；
- sample rate／channel／bit depth 正確；
- 非靜音且無 clipping；
- 正式 game event 能換曲，而非只通過 renderer 工具。

## 8. Software speaker 與第 380 輪更正

`GAME.EXE 19D1Eh..19D5Bh` 向 port `37h` 送 6／7；NP2kai 證明 6 開 gate、
7 關 gate。exact Unicorn harness 對 `period=1000、pulses=2` 證明：

```text
OUT values: 6,6,7,6,7,7
on LOOP executions:  2000
off LOOP executions: 2000
```

NEC 官方 V30 表的 opcode `E2h` 是 `BCWZ/LOOP`：

- counter 非零、分支：5 clocks；
- counter 歸零、離開：13 clocks。

早期 `13/5` 筆記已作廢。依 exact 指令路徑、無 wait、已 prefetch：

```text
busy loop              = 5 × (N - 1) + 13
首次 gate-on interval  = busy loop + 98
後續 gate-on interval  = busy loop + 30
非末次 gate-off        = busy loop + 56
末次 gate-off          = busy loop + 28
```

此模型仍不含 prefetch、bus wait、caller 與機型 CPU clock，只能標為
timing-reconstructed；原 WORD 不是 Hz。

第 381 輪讓 NP2kai 的 i286c/V30 core 直接執行同一段 exact routine，
`period=1／pulse=1` 與 `period=1000／pulse=2` 都重現相同 OUT 順序。
但 NP2kai `_loop` 實際沿用 80286 `taken=8／exit=4`，period 1000 的
busy loop 是 7,996 clocks，與 NEC V30 的 5,008 clocks 不同。因此
NP2kai 是控制流／I/O sequence oracle，不是本題的原機 wall-clock oracle；
不能拿其輸出覆蓋 NEC profile。

## 9. 可沿用的 repository 邊界

```text
作品 game pack
  ├─ track／scene cue／selector metadata
  ├─ exact executable importer
  └─ 本作位址與證據

共用 engine
  ├─ S98 parser
  ├─ YM2203 topology／Timer B
  ├─ Sound BIOS LFO primitives
  ├─ cycle→PCM duty-cycle integrator
  └─ renderer-neutral music contracts
```

下一款作品仍須重新驗證 wrapper、driver SHA、track table、stream bytecode、
clock 與正常玩家路徑。

## 10. 尚未完成

- driver 缺 sector 的可信恢復；
- Timer B reload phase與 save/resume；
- speaker／YM2203 類比 mixer gain；
- 原機 port 37h edge／錄音，或經 microbenchmark 校準的 V30 emulator；
- dormant FM SFX producer；
- DOS PC Speaker、Tandy、AdLib backend；
- 全場景 music cue 與音效驗收。

深入位址與 hash 見 `docs/spec/355`、`364–381` 及後續 timing 規格。
