# 第三百七十一輪 PC-98 Sound BIOS 總音量與 Key-on 還原

狀態：`READY`（限 FM `SETPARABLOCK`、`SETVOLUME`、演算法載波及
`OPERATOR_MASK` key-on 路徑）

本規格補完第 370 輪刻意排除的 YM2203 total-level 公式，並修正重製事件層
固定以 `F0h` 開啟全部四個運算子的錯誤。結論同時由三類獨立證據支持：

1. 指定 IDA Pro 9.4 對 `SOUND.ROM` 的正確 16-bit 反組譯；
2. ROM 原始 bytes 與 NEC 50-WORD parameter block；
3. 十二首 Hoot S98 v3 中 72 組啟動音量序列及第一次 key-on。

這仍不是完整 PC-98 音樂播放器。LFO、fade、SFX 共存、完整 loop、合成器與
遊戲內 PCM 輸出仍在完成邊界之外。

## 1. IDA 載入校正與證據

輸入 `SOUND.ROM`：

```text
bytes   = 16384
sha256  = f05b508d49f31f2a1a61724f013572592abc0833c09c45a72180e84247dc0d0d
linear  = CC000h..CFFFFh
CS      = CEE0h
```

必須用 `/home/anr2/ida_94_official/dist` 對應的 IDA Pro 9.4 工具鏈，在
Docker 內以 8086／16-bit raw binary 分析。IDA 預設把 raw ROM 建成 64-bit
所產生的 qword 與反組譯均已作廢，不得引用。

官方 Sound BIOS interface table 位於 ROM file `2E00h`，公開 far entry 為
`CEE0:0008`／linear `CEE08h`。正確 dispatch 與 consumer：

| 功能 | linear 位址 | 行為 |
|---|---:|---|
| 公開 interrupt entry | `CEE08h` | 保存狀態並進入命令 dispatch |
| AH dispatch | `CEF3Dh` | `AH=16h`、`AH=1Fh` 等 |
| `SETPARABLOCK` | `CF309h` | 複製 50 WORD，再呼叫 tone renderer |
| `SETVOLUME` | `CF41Fh` | 依 algorithm 只改 carrier output level |
| parameter accessor | `CFB94h`／`CFB9Dh` | 依欄位常數反相 |
| key-on helper | `CFC69h` | 欄位 5 左移四位後寫 `28h` |
| total-level restore | `CFD2Ch` | key-on 後由目前 parameter 恢復 TL |
| tone renderer | `CFD72h` | 寫 B0、AR、DR、SR、SL/RR、TL、DT/MULT |

`CS` 相對定址是重要校正。指令顯示 `cs:[bx+0DEFh]`，但不可把 operand
displacement 直接加到 ROM 映射起點。依 ROM file bytes 與函式後方資料
邊界，index-0 哨兵位於 file `3BEFh`／raw-load linear `CFBEFh`，
欄位 1 從 file `3BF0h`／linear `CFBF0h` 開始。欄位 1..50 的常數完整符合：

```text
AR  1..4    = 31
DR  6..9    = 31
SR  11..14  = 31
RR  16..19  = 15
SL  21..24  = 15
TL  26..29  = 127
其他欄位    = 0
```

operator register offset table 位於 `CFD6Eh`／file `3D6Eh`：

```text
00 08 04 0C
```

它與 S98 共同證明 logical operator `1,2,3,4` 的 register slot 是
`1,3,2,4`，完整 burst 的迴圈寫入順序則是 `4,2,3,1`。

可重現的 native IDC 腳本：
[`scripts/ida/pc98_sound_rom_tone_audit.idc`](../../scripts/ida/pc98_sound_rom_tone_audit.idc)。
它只輸出位址與短控制表；ROM、IDA database、assembly 與 log 仍只留本機。

## 2. Total level 精確公式

`SETPARABLOCK` 先把 logical operator 的 `OUTPUT_LEVEL` 轉成晶片 TL：

```text
TL(op) = 127 - OUTPUT_LEVEL(op)
```

`SETVOLUME(AL=channel, BL=volume)` 讀取 `FB_ALG & 7`，只把演算法中的
carrier parameter 改成 `volume`；每改一個 operator 就重新呼叫完整 tone
renderer。因此 trace 會看到同一 timbre signature 的一連串完整 burst：

```text
初始 block TL
→ carrier 4 的 TL = 127 - volume
→ 視 algorithm 再改 carrier 2
→ 視 algorithm 再改 carrier 3
→ 視 algorithm 再改 carrier 1
```

| algorithm | carrier logical operators | 完整 tone-load 數 |
|---:|---|---:|
| 0–3 | `4` | 2 |
| 4 | `4,2` | 3 |
| 5–6 | `4,2,3` | 4 |
| 7 | `4,2,3,1` | 5 |

例如內嵌 parameter 4 是 algorithm 4。其 block base TL 在實體 slot 順序為
`25,25,20,20`；track volume 105 依序形成：

```text
25,25,20,20
25,25,20,22
25,25,22,22
```

這三列可由 ROM 公式獨立計算，也逐 byte 出現在 selector 1 的 S98。

## 3. OPERATOR_MASK 與修正

播放 FM note 時，原 Sound BIOS 不是固定寫 `F0h | channel`：

```text
key_on = (parameter.OPERATOR_MASK & 0Fh) << 4 | channel
write YM2203 register 28h, key_on
```

第 368 輪的 `TrackPlayback` 曾固定使用 `F0h`，會錯誤開啟被音色遮蔽的
operator。本輪把每個 FM channel 的 active parameter mask 納入狀態：

- descriptor parameter 可用時先載入 mask；
- stream `85h` 切換 parameter 時同步換 mask；
- note 的第二次 `28h` write 使用目前 mask；
- 超出二十組 bank 的 descriptor 保留副作用，但因首音前必被內嵌 parameter
  覆蓋，不需假造未知 mask。

共用 engine 的 `audio/s98.YM2203KeyOn` 現保存 trace 的四位
`OperatorMask`；CoAB 稽核器以第一次 key-on 當下有效的內嵌 parameter
逐聲道比較。

## 4. 十二首 S98 corpus 驗收

第 370 輪的十二檔修正版 S98 全部重新稽核。每首三個 FM channel、每聲道
各驗 descriptor 與 first-stream parameter／volume，共：

```text
12 tracks × 3 channels × 2 phases = 72 output-level sequences
```

72 組全部通過：

- embedded parameter 的 base TL；
- 高索引 descriptor 的可觀察 carrier rewrite；
- algorithm 對應 tone-load 數；
- `4→2→3→1` carrier 更新次序；
- `127-volume`；
- 同一序列 timbre signature 不變。

十二首報告的 `operator_masks_match` 也全為 true。約五秒 capture 中只有
已實際 key-on 的聲道才計入 `operator_mask_checks`；稽核不會把未發聲聲道
假算成成功。

`cmd/pc98-s98-audit` 只輸出布林、數量、索引與雜湊，不輸出商業音色值。
新增欄位：

- `output_levels[].base_level_available`
- `output_levels[].base_level_matches`
- `output_levels[].carrier_order_matches`
- `output_levels[].complete_sequence_seen`
- `operator_mask_checks`
- `operator_masks_match`

## 5. 可重用邊界

獨立 engine 的 `audio/ym2203` 保存作品中立的兩項拓樸：

- algorithm 0..7 的 carrier operator；
- logical operator 到 register slot 的映射。

NEC 50-WORD block、`127-parameter`、高索引初始化副作用與 CoAB driver
offset 仍留在作品端 `internal/pc98music`。這避免引擎硬編本作資料，也讓其他
使用 YM2203／NEC Sound BIOS 的 PC-98 Gold Box 可重用同一演算法拓樸與
S98 驗證器。

## 6. 完成與未完成邊界

本輪已完成：

- `SETPARABLOCK` 的 TL 轉換；
- `SETVOLUME` 的 algorithm/carrier 公式與逐 operator 更新順序；
- key-on `OPERATOR_MASK` consumer；
- 十二首 72 組 output-level sequence 與可觀察 first-key mask 稽核；
- 作品中立 YM2203 演算法拓樸及測試；
- 重製正常配樂事件層不再固定 `F0h`。

仍未完成：

1. LFO、`MODUON`／`MODUOFF` 的全部執行期狀態；
2. fade 與 SFX／BGM 共存；
3. 十二首完整曲長與 loop boundary；
4. YM2203 合成器、PCM mixer 與遊戲內播放器；
5. pause／resume、音量設定及 save/load 音樂狀態；
6. DOS PC Speaker／Tandy 音效與 PC-98 theme 的最終政策。

因此不能宣稱「PC-98 音樂完成」或「Sound BIOS 全部還原」。
