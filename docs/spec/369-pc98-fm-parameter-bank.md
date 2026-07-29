# 第三百六十九輪 PC-98 FM 音色參數庫

狀態：`READY`（限內嵌二十組 NEC WORD 音色區塊與所有參數呼叫索引）

> 第 370 輪 S98 runtime trace 已取代本文件對「額外八組可聽音色／外部
> producer」的推論，並修正 NEC rate／level 到 YM2203 的轉換。高索引只在
> descriptor 初始化時出現，第一個 stream `85h` 會在 key-on 前覆蓋；詳見
> [`370-pc98-s98-ym2203-runtime.md`](370-pc98-s98-ym2203-runtime.md)。

本規格修正第 368 輪尚未驗證的假設：FM 音色區塊不在音序所在的 `dseg`，
而在 `seg003`。使用者提供的 `MSCDRV.EXE` 只內嵌二十組完整音色；十二首
曲目實際還會引用八個不在此內嵌區域的索引。因此第 369 輪當時不能把全部
`SETPARABLOCK` intent 展開成真實 YM2203 暫存器，也不能以零值或鄰近音色
代替。第 370 輪已證明可聽音色均由二十組 bank 覆蓋，但仍保留高索引
初始化副作用。

## 1. 證據來源

| 輸入 | 證據 |
|---|---|
| `MSCDRV.EXE` | SHA-256 `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5` |
| IDA Pro | 9.4，Docker 內使用 `/home/anr2/ida_94_official/dist` |
| NEC Sound BIOS | 《PC-9800 Technical Data Book BIOS》（1992）第 370–378 頁 |
| Yamaha YM2203 | 原廠 1989-11 datasheet，暫存器群組 `30h..B2h` |

公開文件只作格式及晶片暫存器語意的權威來源；商業 executable、IDA database、
批次 log 與音色 bytes 均保留在本機，不提交 repository。

- NEC 文件：
  <https://vtda.org/docs/computing/NEC/PC-9800TechnicalDataBookBIOS%2BOCR_1992.pdf>
- Yamaha 文件：
  <https://www.bitsavers.org/components/yamaha/YM2203_198911.pdf>

可重現 IDA 腳本：
`scripts/ida/pc98_music_event_audit.py`。

## 2. 正確段位址

IDA 的段配置是：

```text
dseg   112D0h..13DE0h
seg003 13E60h..14B80h
```

`sub_112B0`／file `0x14B0` 的 exact 指令：

```text
xchg al, bl
mov  cl, 64h
mul  cl
xchg ax, bx
add  bx, 542h
mov  dx, seg seg003
mov  es, dx
xor  dl, dl
call SoundBIOS_AH16_SETPARABLOCK
```

因此位址公式是：

```text
ES:BX = seg003 : (0542h + parameter_index × 100)
DL = 0  ; NEC WORD format
```

`seg003:0542` 對應 executable file `0x45A2`，不是先前誤用的
`dseg:0542`／file `0x1A12`。錯誤位置其實落在曲目 descriptor／sequence，
所得大數值不符合 NEC 欄位範圍，現已廢棄。

## 3. 二十組內嵌參數

file `0x45A2..0x4D72` 恰有：

```text
20 blocks × 100 bytes = 2,000 bytes
SHA-256 = 7bd538f4b80856aa67195f2ddcfe66226632d2f35e6e0d46d04bccfd2031d113
```

每組是 NEC 官方定義的 50 WORD：

- `FB_ALG`、四組 `AR/DR/SR/RR/SL/OP_L/KEY_SCL/MULT/DETUNE`；
- `OPR_MSK`；
- LFO waveform、sync delay、speed、pitch/amplitude depth 及 coarse depth；
- 兩個 reserved 欄位。

`internal/pc98music.FMParameterBlock` 逐欄解析並驗證官方界線，同時保留
50 個 raw WORD。DETUNE 不被強行正規化：官方表寫成 `-4..3`，但真實內嵌
資料同時出現 sign-extended negative 與 `4..7`。在取得 Sound BIOS
consumer 或 register trace 前，只能把這項差異列為未解，不可任意把
`4..7` 解作正數或二補數。

## 4. 十二曲實際覆蓋

第 368 輪的每首 4,096 ticks deterministic execution 現額外收集所有
初始化及 opcode `85h` 的音色索引：

| selector | 使用索引 | 只靠內嵌二十組是否完整 |
|---:|---|---|
| 1 | 4, 5, 6, 24, 27 | 否 |
| 2 | 3, 4, 6, 10, 27 | 否 |
| 3 | 0, 2, 4, 11, 16, 17 | 是 |
| 4 | 12, 13, 14, 25 | 否 |
| 5 | 3, 5, 6, 7, 15 | 是 |
| 6 | 2, 4, 11, 16, 58 | 否 |
| 7 | 1, 11, 13, 21, 23 | 否 |
| 8 | 1, 14, 15, 21, 26 | 否 |
| 9 | 3, 7, 15, 20 | 否 |
| 10 | 2, 4, 8, 9, 15, 20 | 否 |
| 11 | 1, 2, 8, 14, 15, 16 | 是 |
| 12 | 1, 2, 8, 15, 17, 18, 19 | 是 |

全 corpus 聯集是 `0..21, 23..27, 58`；目前內嵌資料缺
`20, 21, 23, 24, 25, 26, 27, 58`。只有 selector 3、5、11、12 可由這二十
組完整覆蓋。

以下是第 369 輪當時的未決推論，現已由第 370 輪 supersede。Hoot metadata 明列
`SOUND.ROM`、`MSCDRV.EXE`、`MSCD_98.COM` 三個必要檔，強烈暗示額外來源
會補入資料，但目前本機沒有合法 `MSCD_98.COM`、Hoot game archive 或
可執行的 Sound BIOS register oracle。第 370 輪 register trace 已證明它們
不是額外 bank producer；高索引會讀到 table 後方相鄰記憶體，且在首個
key-on 前被內嵌音色覆蓋。

## 5. 實作邊界

- `EmbeddedFMParameterBlocks` 只接受 exact driver SHA，解析二十組真實資料。
- `PlaybackAudit.ParameterIndices` 保存每首所有參數呼叫聯集。
- `EmbeddedParametersComplete` 明示所有呼叫是否只落在內嵌 bank；可聽
  key-on 完整性由第 370 輪新增的 `AudibleParametersComplete` 表示。
- `BridgeReport.FMParameterBank` 只輸出 file offset、大小、雜湊、使用及缺失
  索引，不洩漏音色 bytes。
- 不對缺失索引補零、循環取模、複製鄰近音色或聲稱十二首已可忠實合成。

## 6. 尚未完成

1. 完成 carrier total-level／operator mask／algorithm 音量公式。
2. 補 LFO、fade、SFX 共存與完整 loop trace。
3. 接上作品中立 YM2203 合成器、PCM buffer、遊戲內播放與 save/resume。
