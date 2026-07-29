# 第三百七十輪 PC-98 S98／YM2203 執行期音色驗證

狀態：`READY`（限 Hoot 2023 runtime、S98 v3、十二首啟動音色與首次
key-on）

本規格以外部 YM2203 register trace 修正第 369 輪的兩項未決推論：

1. NEC 50-WORD rate／level 欄位不是可直接寫入 YM2203 的值；Sound BIOS
   會反相並重排 operator。
2. `20,21,23..27,58` 並非十二首實際可聽部分所缺的額外音色。它們只在
   descriptor 初始化時被呼叫，隨即由第一個 stream `85h` 音色覆蓋，
   然後才發生該聲道第一次 key-on。

因此二十組內嵌 bank 足以覆蓋目前十二首 4,096-tick 正常播放路徑中所有
實際 key-on 音色。這不代表整個音樂系統已完成。carrier total-level、
algorithm 與 operator-mask 已由後續
[第 371 輪規格](371-pc98-sound-bios-total-level-and-key-on.md)補完；
fade、SFX 共存、LFO、完整曲長、合成器與遊戲內播放器仍待完成。

## 1. 證據與保存邊界

Hoot PC-98 archive 下載來源：
<https://hoot.joshw.info/pc98/adndcoab_98.zip>。

| 輸入 | SHA-256 |
|---|---|
| `adndcoab_98.zip` | `ec1d6d8fe0b390c0e5f7d1596ac41ea5b71aede90ac57c666567739ec95b6d84` |
| `MSCDRV.EXE` | `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5` |
| `SOUND.ROM` | `f05b508d49f31f2a1a61724f013572592abc0833c09c45a72180e84247dc0d0d` |
| `MSCD_98.COM` | `9ccfd48ddf4bf8411ccca15af0d7aa9673bf0c86a394ed2b36f1d6de7c7f70a0` |

archive、runtime binary、S98、Wine prefix、IDA database 與 log 都只保存在
本機 `/tmp` 證據工作區，不提交 Git。Repository 只保存 parser、稽核器、
測試、雜湊與結論。

執行環境：

- `coab-hoot-oracle:20260730`，Wine 8、Xvfb、Mesa software GL；
- Hoot `2023`，`clockmul=8`，S98 logging 開啟；
- PulseAudio null sink 位於獨立有界容器；
- 每首曲目各用一次 `docker run --rm`，無網路、1 GiB、2 CPU、
  `pids-limit=256`；
- 每次搜尋遊戲後進入曲目表，先送 `Home`，再移到絕對 row，避免 Hoot
  保存上次游標而重複擷取。

最初使用相對游標所得的十二檔有重複曲目，已判定無效並移到本機暫存；下表
只列修正後 corpus。

## 2. Hoot row、code 與 game selector

S98 檔名尾碼是擷取流水號，不是 track code。`ponyca.xml` 的顯示順序與
`code` 必須分開保存；game selector 是 `code + 1`。

| row／trace | Hoot code | selector | 日文 metadata |
|---:|---:|---:|---|
| `0000` | `00h` | 1 | タイトル |
| `0001` | `01h` | 2 | キャラクター作成 |
| `0002` | `02h` | 3 | 街 |
| `0003` | `08h` | 9 | 盗賊ギルド |
| `0004` | `06h` | 7 | 戦闘 |
| `0005` | `0Bh` | 12 | ダンジョン |
| `0006` | `04h` | 5 | 荒野 |
| `0007` | `05h` | 6 | 村 |
| `0008` | `0Ah` | 11 | ダンジョン2 |
| `0009` | `07h` | 8 | ズヘンティル城壁 |
| `0010` | `03h` | 4 | ダンジョン3 |
| `0011` | `09h` | 10 | エンディング |

十二檔皆是 S98 v3：

- timer numerator／denominator：`1/100`，每 tick 10 ms；
- 一個 device，type `2`（YM2203）；
- clock：`3,993,600 Hz`；
- compression：0；
- 有合法 end marker。

每檔約 4.88–5.02 秒；register write 數量不同。完整 SHA-256：

| trace | SHA-256 |
|---|---|
| `0000` | `e86987aad9a7cca9b1fa02278a0e3e75184584b27b0443777d87ca47a145cc1b` |
| `0001` | `46fb36a0e7228c44f606520356f5f838e31a9ba586cb503d5664ad39a357beaf` |
| `0002` | `9108395e7a2e86927d48eec83d252238c8f9e8f1a729aeec87804aa16a322099` |
| `0003` | `596989509b4c886cb61072ff6924705096ec13f6e9e705a2f543bea021c7e92b` |
| `0004` | `b0aa5fff2acad5b1d9db2026c746781a864142e9d53b07dbd62adbc0732fdb1c` |
| `0005` | `2044cece6ef775843e7148f03a20a960131aa191bcbb35a199a3eb4c924b083b` |
| `0006` | `c73c71d935718a3ec13370e1ec57acaaff6df95d538372473b301a7ff2f928eb` |
| `0007` | `18969a53ba5ebf0369e955a5c8af1e9ff52b3dcd2be24f907b6c8b0605ce859a` |
| `0008` | `1aa1e99f4a76a7c84e8b0ca651d9890d9589e79b887f40da42188d808a88a154` |
| `0009` | `8f673a40c92c030ace75bbe7ff6d81931827753d3fa1aa0412ef067fd92d44e2` |
| `0010` | `7484073ff7ed8def66ecafa0ecda7b923920d6aff29994ed8fdb09f8ff6acf9d` |
| `0011` | `0a3055de7e4dfa0cbdf2a3bb8961c876e74d59611bf3667007648977d1ea9214` |

## 3. S98 與 YM2203 parser

獨立 engine commit `1a6a252` 新增 `audio/s98`：

- 驗證 `S983` header、timer、compression、device table、dump offset；
- 解碼 device／port register write、`FFh` 1-tick wait、
  `FEh` variable-length wait 與 `FDh` end；
- 拒絕未知版本、壓縮、越界 device、截斷 write／varint 及缺 end marker；
- 辨識 Sound BIOS 完整 tone burst；
- 依 YM2203 實體 register slot 建立 volume-independent signature；
- 在每次 `28h` key-on 保存當時 signature。

tone signature 包含 `B0..B2`、`30..3E`、`50..8E`，刻意排除
`40..4E` total level。原因是 `SETVOLUME` 會改寫 carrier level，但不應被
誤判為另一個音色。

CoAB `cmd/pc98-s98-audit` 接受：

```text
pc98-s98-audit MSCDRV.EXE SELECTOR:TRACE.s98 [...]
```

它以 exact driver SHA、deterministic stream intent、內嵌 bank signature
及 S98 key-on 四方交叉驗證；只輸出 metadata、數量與 SHA-256。

## 4. NEC WORD 到 YM2203 的實測轉換

第 369 輪的 raw 欄位位置正確，但「欄位值可直接寫入晶片」不正確。
S98 對全部二十組內嵌音色逐項支持以下轉換。

NEC logical operator `1,2,3,4` 到 YM2203 register slot 的 canonical 順序是：

```text
1, 3, 2, 4
```

Sound BIOS 實際寫入 burst 則由高位 slot 往回：

```text
4, 2, 3, 1
```

每個 operator：

```text
DT/MULT = byte(DETUNE) << 4 | MULT
KS/AR   = KEY_SCALE << 6 | (31 - ATTACK_PARAMETER)
DR      = 31 - DECAY_PARAMETER
SR      = 31 - SUSTAIN_RATE_PARAMETER
SL/RR   = (15 - SUSTAIN_LEVEL_PARAMETER) << 4
          | (15 - RELEASE_PARAMETER)
```

DETUNE 的 sign-extended negative 不應先截成 3-bit。S98 證明 Sound BIOS
採 8-bit left shift：例如 `-1` 會形成 `F0h | MULT`，因此先前等待
consumer 才能判定的 bit 7 行為現已解開。

total level 另受 track volume、operator mask 與 algorithm carrier 選擇影響；
本輪只將它排除於 timbre signature。其 consumer 與完整公式已在第 371 輪
以 IDA、raw bytes 及同一批 S98 補證。

## 5. 高索引的真實作用

所有超出二十組 bank 的索引只出現在三個 FM descriptor 的
`RawParameter1`。同一聲道的第一個 stream timed event 會先執行 `85h`，
改用 `0..19` 內的音色，之後才 note／key-on。

| 索引 | selector／聲道 | 完整 register burst | 第一次 key-on 時仍有效 |
|---:|---|---|---|
| 20 | `9/ch2`、`10/ch2` | 是 | 否 |
| 21 | `7/ch2`、`8/ch2` | 是 | 否 |
| 23 | `7/ch1` | 是 | 否 |
| 24 | `1/ch2` | 是 | 否 |
| 25 | `4/ch2` | 是 | 否 |
| 26 | `8/ch0` | 是 | 否 |
| 27 | `1/ch0`、`2/ch2` | 是 | 否 |
| 58 | `6/ch2` | 是 | 否 |

這些 burst 的 signature 不等於任何內嵌 `0..19` 音色；它們符合
`seg003:0542 + index×100` 讀到 table 後方相鄰記憶體的結果。trace 不支持
「`MSCD_98.COM` 或 `SOUND.ROM` 補上一個外部音色 bank」的舊假設。

更重要的是，十二首每個 FM 聲道都符合：

```text
descriptor SETPARABLOCK
→ first stream 85h SETPARABLOCK（內嵌 0..19）
→ first key-on 使用後者
```

所以 `PlaybackAudit` 現同時輸出：

- `parameter_indices`：所有呼叫，保留高索引以忠實描述初始化副作用；
- `audible_parameter_indices`：實際 key-on 時有效的音色；
- `embedded_parameters_complete`：所有呼叫是否都在 bank；
- `audible_parameters_complete`：可聽音色是否都在 bank。

十二首的 `audible_parameters_complete` 均為 true。不得刪除 descriptor
呼叫、把高索引取模或假裝它們是額外可聽音色；忠實 runtime 可先執行該
副作用，再於首音前覆蓋。

## 6. 完成與未完成邊界

本輪已完成：

- S98 v3 作品中立 parser；
- YM2203 tone-load／key-on extraction；
- 二十組 NEC parameter 到 volume-independent register signature；
- 十二首 Hoot trace 與 selector／stream intent 的端到端稽核；
- 高索引的初始化副作用與不可聽性判定。

仍未完成：

1. LFO、MODUON／OFF、fade 與 SFX／BGM 共存；
2. 十二首完整曲長與 loop boundary trace；
3. 作品中立 YM2203 合成器與 PCM mixer；
4. CoAB 遊戲內音樂播放、pause／resume、save／load；
5. DOS 音效與 PC-98 配樂的 theme／平台選擇政策。

因此本規格不能用來宣稱「音樂完成」或「PC-98 driver 完整復原」。
