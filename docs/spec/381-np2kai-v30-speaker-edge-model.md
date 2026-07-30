# 381：NP2kai V30 speaker edge 與時鐘模型邊界

狀態：`READY`（限 NP2kai exact routine 控制流與 emulator clock 模型；
原機 wall-clock、prefetch、wait state 與類比輸出仍未完成）

## 1. 問題

第 380 輪已由 NEC 官方表得到 V30 `LOOP` 的 execution timing，但仍缺
NP2kai 真正 CPU core 的 port `37h` edge trace。原 PC-98 媒體有缺 sector，
正常 loader 尚不能穩定啟動遊戲；直接等待 MEGDOS 不能回答 speaker routine
的時序問題。

本輪要分開回答：

1. NP2kai CPU core 執行 exact `GAME.EXE` routine 時，OUT 順序與迴圈邊界
   是否相符？
2. NP2kai 的 `CPU_CLOCK` 能否當成原機 V30 wall-clock oracle？

## 2. 輸入、工具與隔離

- `GAME.EXE` SHA-256：
  `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`
- NP2kai：commit `e2dc904`，`BUILD_I286=ON`、SDL2、i286c/V30 target
- exact routine：GAME file `A6BEh..A6FDh`
- NP2kai source anchors：
  - `io/sysport.c:sysp_o37`
  - `sound/beepc.c:beep_eventset`
  - `i286c/i286c_mn.c:_loop`
- Docker／Xvfb，無網路；GAME 與基底 D88 唯讀

版本化研究入口：

- `scripts/research/pc98_speaker_probe_disk.py`
- `scripts/research/np2kai-port37-trace.patch`
- `scripts/research/np2kai-direct-v30-probe.patch`
- `scripts/research/pc98_np2kai_port37_audit.py`

NP2kai 原始檔是 CRLF；套用兩份 patch 時使用
`git apply --ignore-space-change --ignore-whitespace`，不把整份第三方來源
改成 LF。兩份 patch 已在 commit `e2dc904` 的乾淨副本重新套用驗證。

probe builder 驗證 exact GAME SHA 後才抽取 routine；repository 不保存
商業 routine、probe D88／BIN、NP2kai binary 或 runtime log。

## 3. 為何使用 direct core probe

NP2kai 在 `pccore_reset()` 後才掛載 positional floppy；第一次上電不會自動
以新掛載磁碟重新開機。GUI 的 `Emulate → Reset` 已由重複 BIOS trace 證明
有效，但 V30 POST 加上原媒體 loader 問題使玩家路徑不適合做窄時序實驗。

direct patch 僅在設定 `COAB_V30_PROBE` 時：

1. 將本機 1,024-byte probe IPL 載入 RAM `1FE0:0000`；
2. 設定 real-mode CS／DS／SS 與 segment base；
3. 交回原 NP2kai `np2exec()` 執行。

probe IPL 保留原 PC-98 IPL header 與入口，在 `CS:0110` 以自製 wrapper
建立 FAR argument／stack；`GAME.EXE` routine 本身一個 byte 都不修改。
`04h／05h` 只作 wrapper marker，之後的 `06h／07h` 才是 exact routine。

## 4. 動態結果

### 4.1 邊界：period 1、pulse 1

```text
values = 6,6,7,7
IP     = 0145,0157,0166,0172
clock  = 142,201,242,282
deltas = 59,41,40
```

### 4.2 一般：period 1000、pulse 2

```text
values = 6,6,7,6,7,7
IP     = 0145,0157,0166,0157,0166,0172
clock  = 142,201,8234,16282,24315,32347
deltas = 59,8033,8048,8033,8032
```

兩組皆由版本化 auditor 驗證 marker、CS、OUT count、value sequence 與
edge delta。這獨立支持第 380 輪的 routine 控制流，並關閉「NP2kai core
是否真的走到 exact OUT」。

## 5. NP2kai 不是原機 clock oracle

NP2kai `i286c/i286c_mn.c:_loop` 沒有 V30-specific replacement：

```text
CX--；exit → JMPNOP(4)；taken → JMPSHORT(8)
```

因此其 emulator 模型是：

```text
NP2kai busy = 8 × (N - 1) + 4
```

NEC V30 官方 execution timing 則是：

```text
NEC busy = 5 × (N - 1) + 13
```

| period | NP2kai busy | NEC execution busy | 差 |
|---:|---:|---:|---:|
| 1 | 4 | 13 | -9 |
| 1000 | 7,996 | 5,008 | +2,988 |

period 1000 的 NP2kai `6→7` 是 `8,033` clocks，正好是其 busy `7,996`
加固定路徑 `37`；period 1 則是 `41 = 4 + 37`。這不是 prefetch 或 PC-98
I/O wait 的量測，而是 NP2kai 沿用 80286 `LOOP` handler 的結果。

結論分級：

- `exact`：NP2kai i286c core 的 OUT sequence、IP 與 emulator clocks。
- `exact`：NP2kai source 的 `LOOP taken=8／exit=4`。
- `exact`：NEC 文件的 V30 execution `taken=5／exit=13`。
- `contradicted`：把 NP2kai `CPU_CLOCK` 當成原機 V30 wall-clock。

## 6. Remake 決策

- 不把 NP2kai `8/4` 寫回 `V30PrefetchedProfile`。
- 現行 NEC `5/13` profile 保持 timing-reconstructed。
- NP2kai 仍可作控制流、I/O sequence、磁碟與 GUI runtime oracle；涉及
  V30 instruction wall-clock 時，必須先稽核其 opcode model。
- 下一步需原機錄音／logic trace，或具 V30 prefetch／wait-state 模型且
  經已知 microbenchmark 校準的 emulator，才能升級 wall-clock confidence。

## 7. 尚未完成

- 真實 PC-9801 機型的 instruction prefetch 與 RAM／I/O wait；
- 原機 port `37h` logic edge 或錄音；
- caller 到第一 edge、最後 edge 到 caller 返回的 silence；
- speaker 類比增益、DC offset、濾波與機殼聲學；
- save/resume 時 active one-shot phase。
