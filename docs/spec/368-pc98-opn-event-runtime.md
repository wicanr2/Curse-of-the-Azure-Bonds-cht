# 第三百六十八輪 PC-98 OPN 事件 runtime

狀態：`READY`（限正常配樂路徑的初始化、timer tick、Sound BIOS intent 與
YM2203 register event）

本規格把第 367 輪已解出的 sequence bytecode 推進為 deterministic
event stream。它尚不是音訊合成器，也不宣稱已與 NP2kai／Hoot 的實機
register log 逐事件一致。

## 1. 證據

| 項目 | 證據 |
|---|---|
| `MSCDRV.EXE` | SHA-256 `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5` |
| interpreter | IDA `0x10410..0x10CB9`，`sub_10410` |
| direct register adapter | IDA `0x11189..0x111A1` |
| tempo adapter | IDA `0x111E2..0x111ED`，固定寫 register `26h` |
| `SETVOLUME` | IDA `0x1128F..0x1129D`，Sound BIOS `AH=1Fh` |
| `SETPARABLOCK` | IDA `0x112A2..0x112C6`，Sound BIOS `AH=16h` |
| FM F-number table | DS `0210h`，file `0x16E0`，12 words |
| PSG period table | DS `0228h`，file `0x16F8`，71 words |

可重現 IDA script：
`scripts/ida/pc98_music_event_audit.py`。它只保存位址與擷取規則；原
executable、IDA database、log 與表內容留在本機。

PSG table 的 71-word 邊界不是視覺猜測。opcode note path 接受
`00h..60h`；大於 `18h` 時先減 `18h`，再減 2 並乘 2。因此最大合法
index 是 `(60h - 18h - 2) = 70`，consumer 會讀 `0228h + 70×2`。

## 2. Track 初始化

`sub_10253` 的順序已映射成 tick 0 events：

1. descriptor header word 1 寫入 YM2203 register `26h`。
2. FM channel 0–2：
   - `SETPARABLOCK(channel, raw_parameter_1)`；
   - `SETVOLUME(channel, raw_parameter_2)`；
   - register `28h = channel` key-off。
3. PSG channel 3–5：register `channel+5 = raw_parameter_2`。
4. register `07h = B8h`。

Event 分成 `register_write`、`set_volume` 與 `set_parameter_block`；
作品場景、曲名與 ECL block 不進這層 runtime。

## 3. Timer tick

每次 `Tick()` 對應一次正常 `sub_10410` 呼叫：

- duration 非零時先減 1；FM 等待，PSG 執行 mode 與 envelope step。
- duration 為零時執行 bytecode，直到 note／rest 建立下一個 duration。
- duration byte 先減 1，使用 16-bit wrap，保留原版零值行為。
- 每個 timed event 最多執行 4,096 commands，防止損壞資料造成無限迴圈。

### FM note

非 rest：

1. register `28h = channel` key-off；
2. `(opcode-3) / 12` 取得 octave 與 12-entry F-number index；
3. register `A4h+channel` 寫 octave／F-number high；
4. register `A0h+channel` 寫 F-number low；
5. register `28h = F0h+channel` key-on。

Rest 只執行 key-off。`85` 發出 `SETPARABLOCK`，`8A` 發出
`SETVOLUME`；`90` 在 tempo 小於 `C9h` 時加 4 並寫 register `26h`。

### PSG note／sustain

Note 先把 amplitude register `channel+5` 清零，再由 71-entry table 寫
tone low／high registers；base volume 加上 envelope 起點 word 後寫回
amplitude。Rest 只清 amplitude。

- `85`：envelope pointer = `operand×32 + 10h`。
- `8A`：更新 base／current volume 並立即寫 amplitude。
- `91/92`：依全域 tick counter 位元執行低 tone register ±1 modulation。
- `B0 register value`：直接產生 register write。
- 每個 sustain tick 將 envelope pointer 加 2；word `0080h` 表示停在前一
  step，否則依原 16-bit arithmetic 更新 amplitude。

Timing channel 沿用第 367 輪的有界 read-through，不產生音高 register。

## 4. 實作與 corpus 驗證

`internal/pc98music/playback.go` 提供：

- `NewTrackPlayback`：只接受 exact driver SHA 與 selector 1–12；
- `Tick`：七 channel deterministic scheduler；
- `MusicEvent`：renderer／synthesizer-neutral event；
- `PlaybackAudit`：event count 與 SHA-256 regression digest。

真實十二首各執行 4,096 ticks，全部完成，共產生 68,291 events：

| selector | events | SHA-256 |
|---:|---:|---|
| 1 | 5,689 | `8af2b32f469edb222cd29dbf0344405069412ed7d848f4fa075db8f96de10110` |
| 2 | 5,252 | `ded127b3fc0d7ac5bd2888672ba1d4ad5c21e8ddf11e4afade564e2024ff6a28` |
| 3 | 3,676 | `d0df95594aed7894284fdf6e8d5ad32026b172bf45625f04a8e2c10f49dfc906` |
| 4 | 4,496 | `c7ee0e4092070db63ead0c20287c66f2cf798dddaf8cbe4d9b52805f581531d5` |
| 5 | 9,213 | `fd4c455e318ff27542e7a3be9bf03ba06828ba5cd92e9f4869f6b079f2dffbb5` |
| 6 | 5,587 | `181c282528a8019329b1f135e24af73bfadf173342455c474953795927f96067` |
| 7 | 4,653 | `978a5af7d3549afc16f0cf2bf3a4ae41f8cf7c353f267031f81cd34151ba8b03` |
| 8 | 6,242 | `059c63fd003cb0faf97299780db57a51554588ed0c6ddc3a92fb1bfcc92f5d9a` |
| 9 | 7,444 | `cd8cd69858d08bf8b152decea3f95ac2d5bcc85215fa79b77ace4a2854486ebd` |
| 10 | 4,360 | `40c72b70954eb98090b3f17eb51dba094e010a3bc40d419a0e792e9c7846da26` |
| 11 | 7,172 | `ce5efa278862225335fb879a2e9e040a4a9e7604b74b5c35a1a4dd112cd744dc` |
| 12 | 4,507 | `c5b7c58a86377cc612ab6a881655359b6f83b0756d995286bdc92609d2b14bf7` |

## 5. 尚未完成

- fade／track transition 與插入式 sound-effect channel 的完整共存路徑。
- NP2kai／Hoot 實機 YM2203 register trace 逐事件交叉驗證。
- Sound BIOS parameter block 對實際 YM registers 的展開；第 369 輪已證明
  內嵌 bank 只有二十組，十二曲另引用八個未取得索引，不得補零展開。
- YM2203 合成器、PCM buffer、音量／暫停／save-resume 與遊戲內播放。
- 缺失 `MSCDRV.EXE 0x4000..0x43FF` 的合法第二份 dump。
