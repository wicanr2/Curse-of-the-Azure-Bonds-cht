# Spec 366 — PC-98 十二首曲目表與使用者媒體匯入

狀態：`READY`（track descriptor、sequence 完整性與曲名 metadata）

本規格回答殘缺 `MSCDRV.EXE` 中的十二首公開曲目是否被缺失 sector 切斷，
並建立不提交商業音樂資料的 runtime import 邊界。它尚不表示 YM2203
stream interpreter、音色、節拍或遊戲內實際播放已完成。

## 1. 證據

| 輸入 | SHA-256 |
| --- | --- |
| 殘缺 `MSCDRV.EXE` | `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5` |
| 本機 Hoot `ponyca.xml` | `aae112a387d3e163273c191d8b0d826e0cd85b0a02fd4ee615d4ebab81e89b8d` |

反組譯依 `AGENTS.md` 在 Docker 內優先使用 IDA Pro 9.4：

- `scripts/ida/pc98_music_stream_audit.py`
- `internal/pc98music`
- `cmd/pc98-music-audit`

Hoot XML 是次級 metadata oracle；它沒有取代 executable bytes。商業
`MSCDRV.EXE`、`SOUND.ROM`、`MSCD_98.COM`、磁碟映像及抽出的 sequence
均不提交 GitHub。

## 2. Track table 與 descriptor

`sub_1021E`／file `0x041E`：

1. 接受 0-based driver index；
2. 將 index 乘二；
3. 從 `DS:0330` 讀取 track descriptor offset；
4. 保存給 timer dispatch 載入。

IDA 與 raw bytes 對回 `DS:0330`／file `0x1800` 的前十二個 word：

```text
038A 03CA 040A 044A 048A 04CA
050A 054A 058A 05CA 060A 064A
```

每個 descriptor 恰為 64 bytes：

```text
8-byte header
7 × 8-byte channel record
```

`sub_10253` 逐聲道讀取 record。已證明的欄位為：

```text
+0 sequence offset（相對 driver data segment）
+2 sequence length
+4 raw parameter 1
+6 raw parameter 2
```

後兩個 word 尚未完成 producer／consumer 閉環，故 auditor 保留 raw 名稱。

## 3. 缺失 sector 沒有切到十二首 sequence

driver data segment 對應 file base `0x14D0`。auditor 把每個
`sequence offset + length` 映回 file half-open range，所得每首曲目的
完整聯集如下：

| 1-based selector | 0-based index | sequence file range | 結果 |
| ---: | ---: | --- | --- |
| 1 | 0 | `0x1B61..0x208B` | 完整 |
| 2 | 1 | `0x208B..0x2476` | 完整 |
| 3 | 2 | `0x2476..0x2721` | 完整 |
| 4 | 3 | `0x2721..0x2C09` | 完整 |
| 5 | 4 | `0x2C09..0x2D9F` | 完整 |
| 6 | 5 | `0x2D9F..0x2E37` | 完整 |
| 7 | 6 | `0x2E37..0x2FC1` | 完整 |
| 8 | 7 | `0x2FC1..0x30ED` | 完整 |
| 9 | 8 | `0x30ED..0x3616` | 完整 |
| 10 | 9 | `0x3616..0x37D0` | 完整 |
| 11 | 10 | `0x37D0..0x3964` | 完整 |
| 12 | 11 | `0x3964..0x3C58` | 完整 |

缺失 sector 是 file `0x4000..0x4400`，所以 12 個 descriptor 與 84 個
channel stream 都完全位於缺口之前。這項結論是 raw-byte exact；它只證明
音符／命令 sequence 沒有遺失，不證明後段音色、工作區或 playback 所需的
所有資料完整。

IDA 顯示缺口附近落在 Sound BIOS／driver 的可寫工作段，其中多個欄位會在
初始化與 timer 運作時寫入。現有抽出檔在缺口處的零 bytes 是媒體 parser
代填，不能當作原始 sector 內容。要把整份 driver 稱為完整，仍需第二份合法
dump 或 runtime 等價性證明。

## 4. 十二首 selector 名稱

本機 Hoot `ponyca.xml` 的 CoAB entry 指定：

```text
driver=pc98dos
required files=SOUND.ROM, MSCDRV.EXE, MSCD_98.COM
shells=mscdrv, mscd_98
clockmul=8
```

其 Shift-JIS title code 是 0-based driver index。與先前公開
register-log 的十二首清單交叉吻合後，game pack 保存以下 metadata：

| index | 英文 | 繁中 |
| ---: | --- | --- |
| `00` | Title | 標題 |
| `01` | Character Creation | 角色建立 |
| `02` | Town | 城鎮 |
| `03` | Dungeon 3 | 地城三 |
| `04` | Wilderness | 荒野 |
| `05` | Village | 村莊 |
| `06` | Combat | 戰鬥 |
| `07` | Zhentil Keep Walls | 散提爾堡城壁 |
| `08` | Thieves Guild | 盜賊公會 |
| `09` | Ending | 結局 |
| `0A` | Dungeon 2 | 地城二 |
| `0B` | Dungeon | 地城 |

Hoot code 只命名 driver index，不證明每個 ECL block 的場景。既有 spec 355
的 ECL→selector 表仍以 GAME executable 為權威；兩者現在透過
`reference_selector = driver_index + 1` 接合。

## 5. Remake contract

獨立 engine 的 `music_tracks` 新增可選 `title_id`：

- engine 只驗證 `title_id` 在所有 locale 都存在；
- CoAB JSON 保存十二首 selector、driver index 與中英文曲名；
- `internal/pc98music.ExtractTrackSequences` 只接受已辨識 SHA-256，
  只接受 selector 1–12，且拒絕任何跨缺口 stream；
- API 回傳七聲道 bytes 的副本，原商業資料不寫入 repository。

這是 runtime import 的第一層，不是播放器。下一層必須把
`sub_10410` stream interpreter 解成有界、deterministic 的 OPN register
events，再交給作品中立 YM2203 adapter。

## 6. 可重現驗證

```text
go test ./internal/pc98music ./cmd/pc98-music-audit
go run ./cmd/pc98-music-audit GAME.EXE MSCDRV.EXE
```

JSON report 會列出：

- 十二個 descriptor；
- 每首七聲道的 offset、length、file range、SHA-256；
- 每一 channel 是否跨越缺失 sector；
- spec 364／365 的 bridge、Sound BIOS 與 direct YM anchors。

## 7. 尚未完成

- 解出 `sub_10410` 全部 stream opcode、等待／loop／tempo 語意。
- 確認 `0x4400` 後音色／參數表與缺口工作區的完整 consumer。
- 以 Hoot、NP2kai 或另一實作交叉比對每首 YM register event。
- 實作可取消、可調音量、可 save/load resume 的 YM2203 playback adapter。
- 由正常玩家路徑擷取 title、town、combat 三次實際轉場與聲音。
