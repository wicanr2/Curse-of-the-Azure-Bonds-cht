# 380：PC-98 speaker exact routine 與 V30 timing 更正

狀態：`READY`（限 exact routine 控制流、OUT／LOOP 動態 trace、NEC
execution-unit no-wait timing 與修正後 PCM；原機 wall-clock 仍未完成）

## 1. 更正結論

第 379 輪把 NEC V30 `BCWZ/LOOP` 表的兩個 condition 讀反，使用了
`taken=13／exit=5`。重新以指定 IDA Pro 9.4 匯出 exact routine、Unicorn
2.1.4 執行原 bytes，並逐列核對 NEC 官方表後，正確值是：

- counter 非零、branch taken：5 clocks；
- counter 歸零、exit：13 clocks。

第 379 輪舊 PCM hash 與時長作廢；selector 語意、caller 與 platform intent
結論不受影響。

## 2. 輸入與工具

- `GAME.EXE` SHA-256：
  `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`
- IDA Pro 9.4：Docker 內使用
  `/home/anr2/ida_94_official/dist` 對應工具鏈
- Unicorn：2.1.4、x86 16-bit mode
- NEC：
  [16-BIT V SERIES Instruction User's Manual](https://datasheets.chipdb.org/NEC/V20-V30/U11301EJ5V0UMJ1.PDF)，
  Table 2-8

商業 executable、IDA database、log 與 runtime JSON trace不提交。

## 3. Exact routine

MZ header 是 `0x9A0` bytes；IDA load base `0x10000`。speaker routine：

- IDA `19D1Eh..19D5Bh`
- raw file `A6BEh..A6FBh`
- sound period word：IDA `280E6h`

控制流：

```text
OUT 37h,6                 ; 初始開 gate
for each pulse:
    OUT 37h,6             ; 第一輪是 redundant
    LOOP period times
    OUT 37h,7
    LOOP period times
OUT 37h,7                 ; 最後一次 redundant
```

`scripts/research/pc98_game_speaker_harness.py` 從 exact MZ load image執行
原始 bytes，不內嵌商業程式。

## 4. 動態 trace

`period=1000、pulses=2`：

```text
instruction_count = 4045
on LOOP count      = 2000
off LOOP count     = 2000
OUT values         = 6,6,7,6,7,7
OUT addresses      = 19D2A,19D3C,19D4B,19D3C,19D4B,19D57
```

這證明第一個 6 與最後一個 7 是 routine 邊界 gate write；中間每個 audible
pulse 各有 N 次 on LOOP 與 N 次 off LOOP。

## 5. 逐指令 cycle

NEC V30 no-wait execution clocks：

| instruction | path | clocks |
|---|---|---:|
| `BCWZ/LOOP` | branch | 5 |
| `BCWZ/LOOP` | exit | 13 |
| `MOV reg,reg` | — | 2 |
| `MOV reg,imm` | — | 4 |
| `MOV acc,dmem` | even | 10 |
| `OR reg,reg` | — | 2 |
| `INC reg16` | — | 2 |
| `CMP reg,mem` | even | 10 |
| conditional branch | taken | 14 |
| conditional branch | not taken | 4 |
| `BR short` | — | 12 |
| `OUT imm8,acc` | no wait | 8 |

因此：

```text
busy loop              = 5 × (N - 1) + 13
首次 gate-on interval  = busy loop + 98
後續 gate-on interval  = busy loop + 30
非末次 gate-off        = busy loop + 56
末次 gate-off          = busy loop + 28
```

這些 interval 依 OUT edge 間 exact instruction path 計算。官方表明確不含
prefetch、pre-decode 與 bus wait，因此仍是 execution-unit profile，不是
原機 wall-clock exact。

## 6. Remake 修正

`pc98sfx.V30PrefetchedProfile` 現保存：

```text
LoopTakenCycles             5
LoopFinalCycles            13
InitialGateOnOverhead      98
GateOnOverhead             30
GateOffOverhead            56
FinalGateOffOverhead       28
```

`RenderPCM` 分開第一個 on、後續 on、非末次 off 與末次 off，不再把所有
half-wave 視為同長。

8 MHz／44.1 kHz deterministic output：

| effect | frames | duration | SHA-256 |
|---|---:|---:|---|
| `ARROWFX` | 1,865 | 0.042290 s | `b9fc898253a380679e84c2026c84c9725a30302dbb15f7814a411664f2f50a5a` |
| `FIREBALLFX` | 736 | 0.016689 s | `b8922db10390746d5bb5b06f28385c0ca779a17f35bc01fef29a3bbeea7c5be8` |

ARROWFX 兩次 hash 一致；FFmpeg 證明 44.1 kHz、stereo、s16、peak
`-14.7 dB`。

## 7. 後續狀態

> 第 381 輪已完成 NP2kai 真正 i286c/V30 core 的 OUT edge trace，但證明
> 該 core 的 `LOOP` 使用 80286 `taken=8／exit=4`，不能作原機 V30
> wall-clock oracle。詳見
> `381-np2kai-v30-speaker-edge-model.md`。

- prefetch／pre-decode；
- memory／I/O wait 與不同 PC-98 機型；
- SOUNDFX caller／table interpreter 間的 silence；
- 原機類比 gain、濾波與錄音對照。

所以第 380 輪關閉「routine 到底執行幾次 LOOP、NEC taken／exit clock
是多少、PCM profile 是否讀反」；仍不宣稱 8 MHz profile 是原機
cycle-perfect。
