# 第三百六十七輪 PC-98 音序列 bytecode

狀態：`READY`（限 `sub_10410` 已證實的 command framing、控制流與有界驗證）

本規格不代表 YM2203 音色、tempo、音高、音量或實際播放已完成。它只把
十二首曲目的 84 組 sequence 從不透明 bytes 提升為可驗證的指令串流。

## 1. 證據

- 輸入：使用者本機 `MSCDRV.EXE`，SHA-256
  `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5`。
- 主分析：指定 IDA Pro 9.4，`sub_10410` 的 FM、PSG、timing 三條分支。
- 交叉驗證：原始 descriptor／sequence bytes，以及
  `cmd/pc98-music-audit` 對十二首、七 channel 的實際 trace。
- 商業 executable、IDA database 與抽出的 sequence 不提交 repository。

## 2. Channel family

| channel | family | 已證實的特殊命令 |
|---:|---|---|
| 0–2 | FM | `85 8A 90 A0–A4 B0` |
| 3–5 | PSG | `85 8A 91 92 A0–A4 B0` |
| 6 | timing | `85 8A`；其他高位 byte 逐 byte 略過 |

三類共同規則：

- `00h..60h`：note code，後接一個 duration byte。
- `80h`：rest，後接一個 duration byte。
- `85h`、`8Ah`：各消耗一個參數。語意仍以 opcode 名保存，不提前猜名。

FM／PSG 控制流：

- `A0 target16`：絕對 DS offset jump。
- `A1 target16`：保存 operand 後 PC，再跳至絕對 DS offset；call stack
  最多 16 筆，滿時原程式忽略 push／jump。
- `A2`：return；空 stack 時原程式 no-op。
- `A3 count8`：保存 operand 後 PC 與 8-bit 次數；loop stack 最多 16 筆，
  滿時 no-op。
- `A4`：遞減最上層 8-bit count；非零跳回，歸零 pop；空 stack 時 no-op。
- `B0 arg0 arg1`：FM 只略過兩 bytes；PSG 以兩參數呼叫直接 OPN register
  write helper。

`90` 只屬 FM；`91/92` 只屬 PSG。未在 family 分支證實的 opcode 不可靜默
當作另一 family 的命令。

## 3. Timing channel 的 read-through

Timing channel 不能用 descriptor half-open range 當 execution boundary。
例如第六首 channel 6 宣告 sequence 尾端是：

```text
... 4C 18 A0 D3 18
```

IDA 顯示 timing 分支不執行 `A0`；它依序略過 `A0`、`D3`，再把 `18`
視為 note code，並從 descriptor 尾端後讀 duration。原驅動也沒有在這條
分支檢查 sequence end。

因此 auditor 分成兩種模式：

- FM／PSG：`declared-range-static-and-control`，嚴格驗證宣告範圍、operand
  width、target 與控制流。
- timing：`bounded-runtime-read-through`，從 sequence start 執行，但只在
  完整 driver data 與命令數上限內讀取，不假造 descriptor range gate。

這是原作行為的安全重建，不表示 remake 應允許任意越界；runtime importer
仍先驗證完整 executable SHA-256。

## 4. 實作與驗證

`internal/pc98music/stream.go` 提供：

- family-aware 結構解碼；
- `SequenceMachine` 有界控制流；
- 16-entry call／loop stack 與原版 overflow／underflow no-op；
- target、operand、資料範圍及每個 timed event 最大命令數檢查。

可重現命令（Docker 內）：

```sh
go test ./internal/pc98music ./cmd/pc98-music-audit
go run ./cmd/pc98-music-audit GAME.EXE MSCDRV.EXE
```

真實 corpus 結果：

- 12 首 × 7 channel＝84 組均通過；
- 每組完成 256 個 timed events 的有界 trace，共 21,504 events；
- 12 組 timing channel 均明確標為 read-through mode；
- 未使用缺失的 file `0x4000..0x43FF` 作為 sequence 證據。

## 5. 後續狀態

第 368 輪已完成正常配樂路徑的 note／參數、timer sustain 與 deterministic
Sound BIOS／YM2203 events；以下剩餘項目以 spec 368 為準。

- fade／sound-effect 共存路徑與外部 register trace。
- Sound BIOS parameter block 到 register 的展開。
- 合法授權、可跨平台的 YM2203 合成後端及實際音訊播放。
- 缺失 driver sector 與工作區的合法第二份 dump。
