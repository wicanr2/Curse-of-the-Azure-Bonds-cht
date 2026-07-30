# 378：PC-98 GAME SOUNDFX selector 與喇叭程式

狀態：`READY`（限 `GAME.EXE SOUNDFX` 控制流、`GAME.OVR` 直接呼叫、
selector 分布與作品端音序匯入；實際聲波時鐘仍未完成）

## 1. 結論

PC-98 版的正常遊戲短音效不是第 377 輪找到的 MSCDRV 內部 FM SFX。
`GAME.OVR` 共有 42 個直接呼叫，固定採：

```text
PUSH word ptr DS:[selector 常數]
CALL FAR 0893:0000             ; GAME.EXE SOUNDFX
```

`SOUNDFX` 分成三類：

| selector | 行為 |
|---|---|
| `0,1,13,14,15,255` | 立即返回，不發聲 |
| `2,4,6,9` | `SOUND(selector×10)`，再執行 `250/selector` 次脈衝 |
| 其餘 `3,5,7,8,10,11,12` | 讀 20-WORD 音序表；零值終止；大於 2500 時 `DELAY(5)`；其餘值先 `SOUND(value)`，再執行 `2000/value` 次脈衝 |

Borland `SOUND` 在本作只把參數存進全域 WORD。脈衝 routine 反覆向 PC-98
port `37h` 寫入 `06h／07h`，每個半波以該 WORD 次 `LOOP` 忙等。這不是
IBM PC PIT，也不能直接把參數當成 Hz。

作品端新增 `internal/pc98sfx`，只接受 exact `GAME.EXE`，將上述程式解成
renderer-neutral pulse／delay steps。商業音序 bytes 不放進 Git。
`cmd/pc98-sfx-audit` 可對使用者本機 executable 重建 JSON 報告。

## 2. 輸入與位址證據

| 輸入 | SHA-256 |
|---|---|
| `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` |
| `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` |
| `GAME.EXE` 音序表 640 bytes | `65e9cf2cd93ae31edb497666415b54f936b98b8d89f87d475ebf3c4815c59ac4` |

IDA linear address：

- `SOUNDFX`：`18930h..18A3Ch`；
- Borland `SOUND`：`19286h`；
- port `37h` pulse routine：`19D1Eh..19D5Dh`；
- 音序表：DS `1A3Ch`，IDA linear `1DCCCh`。

音序表的 exact raw file base 是 `E66Ch`。這不是由 MZ linear address
硬減一個固定差值猜出來；方法是先由 IDA 匯出記憶體中的 640 bytes，再以
多個 64-byte 視窗回搜 raw `GAME.EXE`。除全零重複區外，所有有效視窗都
一致映到 `E66Ch`，整表雜湊也相符。

## 3. IDA Pro 與 raw bytes 交叉驗證

依 `AGENTS.md`，主要分析使用 Docker 內、對應
`/home/anr2/ida_94_official/dist` 的 IDA Pro 9.4：

- `scripts/ida/pc98_game_music_transition_audit.idc`
- `scripts/ida/pc98_overlay_soundfx_audit.idc`

raw overlay 必須在 IDC 內明確切到 `8086` 並把 segment bitness 設成 16-bit。
IDA 預設 64-bit 時產生的 qword／`push qword` 結果作廢。

第二條證據是 `cmd/pc98-ovr-audit -soundfx`：

- 先以 TPOV control chain 分離 36 段 code 與 relocation；
- 只接受緊鄰的 `FF 36 xx xx 9A 00 00 93 08`；
- DS `4838h..4858h` 必須落在 exact resident selector 常數表；
- 常數表第一筆是 unsigned `00FFh`，不是 signed `-1`。

IDA 8086 結果與 raw auditor 的 42 組 call offset／selector 全部一致。

## 4. Overlay／模組分布

overlay index 對應既有 Borland compiler module index 減一：

| overlay／module | selector 分布 |
|---|---|
| 1 `INTRO` | `13×1` |
| 2 `INTERPET` | `255×13, 10×3, 11×1, 15×1` |
| 3 `PROTECT` | `1×1, 5×1` |
| 13 `COMSTUFF` | `10×1, 7×2, 9×3, 12×2, 6×2` |
| 14 `MOVEMENT` | `10×3` |
| 22 `SPELLS` | `11×1, 8×1, 2×2` |
| 24 `GENERIC` | `4×1, 3×1` |
| 32 `LOS` | `5×2` |

總計：

```text
1:1, 2:2, 3:1, 4:1, 5:3, 6:2, 7:2, 8:1,
9:3, 10:7, 11:2, 12:2, 13:1, 15:1, 255:13
```

`MOVEMENT` 三處只送 selector 10，與既有 DOS 資源名稱 `step.wav` 交叉後，
selector 10「腳步」可標為模組級實證。`SPELLS`、`COMSTUFF` 等只證明
功能群，尚未逐一追到每個 spell／attack caller；因此本規格不擅自替
selector 3–12 全部命名。

## 5. Remake contract

- `pc98sfx.Import` 必須先驗證 exact executable SHA-256；
- selector 0–15 全部產生 typed `Effect`；
- no-op、公式與音序表三條路徑分開保存；
- `FrequencyOrPeriod` 保留原 `SOUND` 參數，不命名為 Hz；
- pulse count 保存原整數除法結果；
- table 中 `>2500` 只變成 5ms delay，不發 pulse；
- 原始 executable、音序表、IDA database 與 log 都只留本機。

目前還不能把 pulse steps 直接送進 44.1kHz mixer。實際半波時長取決於
8086／V30 `LOOP`、`OUT`、記憶體存取及 PC-9801 CPU clock；在建立
machine-profile cycle model 或 NP2kai 音訊 trace 前，任意指定 tone Hz
都只能算重製音，不是 faithful PC-98 聲音。

## 6. 驗證與尚未完成

已完成：

- focused unit tests 驗證 unknown executable fail-closed；
- no-op、公式、table、delay、零終止與整數 pulse count；
- exact `GAME.EXE` 可輸出 16 selectors；
- IDA 8086 與 raw overlay auditor 共同確認 42 個 callers。

尚未完成：

- 逐 caller 的函式名稱、遊戲事件與 selector 3–12 完整語意；
- PC-9801 CPU／port 37h cycle timing 與可聽 PCM；
- Ebiten 正常玩家路徑的 PC-98 one-shot mixer；
- 音樂與短音效同時播放的 gain／priority；
- 原版實機或影片的聲音時間碼與聽感對照；
- MSCDRV dormant FM SFX 是否有目前尚未取得的外部 producer。

因此本輪關閉「正常 GAME 短音效究竟呼叫哪條程式、有哪些 selector、
音序資料在哪裡」的缺口，不代表 PC-98 音效已可播放或全遊戲音效完成。
