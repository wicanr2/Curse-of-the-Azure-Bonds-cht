# Spec 362 — PC-98 Disk B、DAX codec 與 WLDTWN 場景語意

狀態：`READY`

本規格只涵蓋 Disk B 的唯讀檔案配置、PC-98 DAX block codec，以及
`WLDTWN` 對 BGM selector 5／6 的場景語意。它不代表 PC-98 音樂驅動、
曲名、音序列或播放 adapter 已完成。

## 1. 證據輸入與邊界

- 原始 Disk B VFD SHA-256：
  `38f56ab7f17690b72afa6e7b6b6462a9ec5511f53b7567303e1dcc681982bfbe`。
- 77×2×8×1024 CHR 順序研究副本 SHA-256：
  `d8ec10e783577b075b002e8e67a0607541a1d2a401ec24019c449da330dcaaf8`。
- 唯讀抽出的 `ECL.DAX`：
  - file offset `0x37400`；
  - size `132662` (`0x20636`)；
  - SHA-256
    `edbd8fabf2d002a38ed84348b8b2c7a3943eb8fd5b795f313b7b9b67c86bca9f`。

原始 VFD、研究副本、抽出 DAX、GAME.EXE 與 IDA database 都不得提交。
Repository 只保存 parser、稽核 script、雜湊及反組譯結論。

## 2. Disk B 是無 BPB 的 FAT12 資料盤

Disk B sector 0 是遊戲自有 boot／label，不含可供一般 DOS FAT 工具接受的
標準 BPB，因此 `mtools` 報告 non-DOS media 並不表示後續資料沒有 FAT。
原始 bytes 證明：

- FAT #1：`0x0400..0x0BFF`；
- FAT #2：`0x0C00..0x13FF`，與 FAT #1 的 2048 bytes 完全相同；
- root directory：`0x1400..0x2BFF`，32-byte FAT directory entries；
- data region／cluster 2：`0x2C00`；
- bytes per sector、cluster：1024；
- `ECL.DAX` directory entry：start cluster 212、size `0x20636`；
- FAT12 chain 212–341 連續，恰好覆蓋 130 clusters。

所以：

```text
ECL.DAX offset
= data_start + (start_cluster - 2) × cluster_size
= 0x2C00 + (212 - 2) × 0x400
= 0x37400
```

先前以 `0x1C00` 當 data start 所得檔案不是有效 DAX，已被 root-directory
長度、FAT chain 與有效 header 三項證據推翻。

## 3. PC-98 DAX index 與 codec

IDA Pro 9.4 使用本機 `GAME.EXE.i64`，輸入 `GAME.EXE` SHA-256
`8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`。
Borland symbol `GETDATABLOCK 0723:0824` 對應 IDA `0x17A54`：

- `(header_size - 2) / 9` 得 block 數；
- 每筆仍為 DOS 共用的 9-byte
  `id/u32 offset/u16 decoded_size/u16 stored_size`；
- stored block 第一 byte 是 codec flag；
- `0xFF` 代表跳過 flag 後直接複製 raw block；
- 其他 flag 交給 `0x17DD5..0x17E2A` 解壓。

`0x17DD5` 的指令級 codec：

1. 跳過第一個 codec flag。
2. 第二 byte 保存共用 fill value。
3. 每個 control byte：
   - bit 7 = 0：原樣複製 `control` bytes；
   - bits 7–6 = `11`：重複 fill value `(control & 0x3F) + 2` 次；
   - bits 7–6 = `10`：讀下一 byte，重複 `(control & 0x3F) + 3` 次。

`internal/dax.ParsePC98` 是這段指令的 typed 實作；24 個真實 ECL blocks
逐一精確符合 declared decoded size。`scripts/ida/pc98_dax_codec_audit.py`
可重列 `GETDATABLOCK` 與 decoder 指令。DOS `Parse` 維持原 codec，不能把
兩種 block decoder 混用。

## 4. WLDTWN writer

IDA 與 raw overlay 已證明：

- `WLDTWN = 0C29:7F11`；
- `INITECL 0062:0317` 在 overlay 7 local `0x0478` 寫 0；
- `STOREVALUE 0062:0E2B` 在 local `0x0F3D` 比較特殊目的位址 `0x5208`，
  local `0x0F45` 將 SAVE value 的低 byte 寫入 `WLDTWN`；
- `BGMPLAY 0893:0177` 對 block `0x50/0x51` 依此 byte 選 5 或 6。

PC-98 `ECL.DAX` 全 24 blocks 解壓後逐 byte 掃描，只得到四個
`SAVE → 0x5208`：

| ECL block | payload offset | value |
| --- | ---: | ---: |
| `0x50` | `0x063B` | 1 |
| `0x50` | `0x0740` | 0 |
| `0x51` | `0x038B` | 1 |
| `0x51` | `0x046D` | 0 |

這些不是 packed stream 的偶然 bytes：`ParsePC98` 解壓後 operand cursor
可完整解碼；writer、特殊地址 consumer 與 BGM consumer 三方一致。

## 5. 跨平台場景語意

兩個 value=1 writer 後的 PC-98 指令序列都是：

```text
SAVE 1 → 0x5208
SAVE 1 → display flag
CLEAR BOX
SAVE 0 → display flag
PICTURE 0x50
... PRINT／HORIZONTAL MENU
```

DOS ECL1 同 block、同相鄰控制流程沒有平台專用 `0x5208` SAVE，但其可讀文字
逐項對齊：

```text
YOU ARE IN [location]. WHAT PLACE WILL YOU VISIT?
INN / STORE / BAR / ...
```

value=0 writer 則返回 block 的主要區域／戶外導航入口；`0x51` 還明確接
`PICTURE 0x79` 世界圖。故場景角色已由 `hypothesis` 提升為 `proven`：

- `WLDTWN == 0`：區域／戶外導航；
- `WLDTWN != 0`：城鎮設施選單。

CoAB JSON 因而讓 `0x50/0x51` 的 context-free 初始 binding 指向 selector 5；
這與 `INITECL` 必定先清零一致。selector 6 的 context 改為
`pc98-town-services-menu`。

第 363 輪進一步在作品中立 engine 加入 `music_cues`：它只接受
`ECL block + signal + raw value → opaque context`，不認識城市、選單文字或
作品旗標。CoAB JSON 宣告：

| ECL blocks | signal | raw value | context | binding 結果 |
| --- | --- | ---: | --- | --- |
| `0x50/0x51` | `picture` | `0x50` | `pc98-town-services-menu` | selector 6 |
| `0x50/0x51` | `picture` | `0x79` | `pc98-world-navigation` | context-free fallback → selector 5 |

State 會在處理真實 `PICTURE` signal 時解析 cue，並依 `MSCPLAY` 已證實的
行為抑制同一 track 重播。正常 ECL 玩家路徑已驗證：

- block `0x50`：阿沙本福德入城 `5→6`，離城 `6→5`；
- block `0x51`：希爾斯法入城 `5→6`，酒館返回服務選單不重播 6，離城
  `6→5`。

這只證明選曲 intent 時序；缺失 driver、音序列、YM trace 與播放 adapter
仍未完成。

## 6. 驗證

```text
go test ./internal/dax ./internal/ecl ./cmd/azure-bonds ./cmd/pc98-ovr-audit
azure-bonds -input-file ECL.DAX -pc98-dax \
  -find-save-destination 5208
```

真實 corpus 結果必須是 24 blocks，且只有上表四個 decoded candidates。

同 block 音樂 cue 的正常玩家路徑 regression：

```text
go test ./internal/game \
  -run '^TestFireKnifeLeaderStateVictoryReturnsToTilverton$' -count=1
```

## 7. 後續

- 繼續 spec 358 的 IVT `7Eh`／INT D2h、缺失 driver sector 與 YM runtime
  trace；在音序列證據完成前仍不填曲名。
