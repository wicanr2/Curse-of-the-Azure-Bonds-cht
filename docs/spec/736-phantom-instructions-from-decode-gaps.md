# 第七百三十六輪：反組譯掉位元組會生出「合理的假指令」

狀態：`READY`。等級：`exact`。日期：2026-08-14
工具：`scripts/decode_gap_scan.py`

## 起因

`overlay-12:356Ch` 的匯出只有五條指令：

```text
356C 55     push bp
356D 89e5   mov bp, sp
3570 ec     in al, dx        ← 位址從 356Fh 跳到 3570h
3571 5d     pop bp
3572 cb     retf
```

去 `.bin` 讀原始位元組：`55 89 E5 89 EC 5D CB`。`89 EC` 是 `mov sp, bp`，
被切掉了前面的 `89`，剩下的 `EC` 正好是 `in al, dx`。

**實際上這是一支空函式**，而匯出把它變成「讀 I/O 連接埠」。

## 為什麼危險

假指令**看起來完全合理**。`in al, dx` 是硬體存取，照著寫進規格就會有人去追一個
根本不存在的連接埠。`hlt`、`retf 19Bh`、`lds ax, [bp+si−7110h]` 這些同樣是
「單獨看不奇怪」的指令。

判準很單純：同一支函式裡，前一條的 `ea + 位元組數` 必須等於下一條的 `ea`。

## 掃描結果

全工作區共 **236 處**位元組數對不上。依缺口大小：

| 缺口 | 處數 |
|---|---|
| 1 byte | 50 |
| 2 | 28 |
| 3 | 19 |
| 4 | 10 |
| 5..9 | 12 |
| >= 10 | 117 |

`>= 10` 的多半是 IDA 正確地把資料（跳表、字串）標成非程式碼，不是問題。
**1..4 bytes 的那 107 處才是要逐一確認的**。

## 已經標成「已解讀」的函式裡有 23 處

全部在 PC-98 側。其中 **11 處是 `INT 3Fh` 後面的 3 bytes**——那是 VROOMM stub
的描述子（spec 562），IDA 正確地跳過，**不是問題**。

剩下 12 處需要回去對原始位元組：

| 位置 | 缺口後那條 |
|---|---|
| `pc98 overlay-02:3CE4h` 內 `3CE8h` | `in al, dx`（1 byte） |
| `pc98 overlay-23:15ECh` 內 `1627h` | `hlt`（4 bytes） |
| `pc98 overlay-12:1437h` 內 `1464h` | `lds ax, [bp+si−7110h]`（4 bytes） |
| `pc98 overlay-12:147Ah` 內 `1515h` | `xchg ax, si`（3 bytes） |
| `pc98 overlay-24:1739h` 內 `17B7h` | `test ax, 0C282h`（1 byte） |
| `pc98 PC98-GAME.EXE:10C90h` 內 `10CB8h` | `or ax, [bx+si]`（1 byte） |
| `pc98 PC98-GAME.EXE:10D00h` 內 `10D04h` | `retf 19Bh`（2 bytes） |
| `pc98 PC98-GAME.EXE:17DD5h` 內 `17DFFh` | `test cl, 40h`（1 byte） |
| `pc98 PC98-GAME.EXE:18EE0h` 內三處 | `push bx`／`mov byte_280E5, 13h`／`xor al, al`（各 1 byte） |
| `pc98 PC98-GAME.EXE:1B756h` 內兩處 | `mov ax, …`／`mov ah, 2Ch`（各 2 bytes） |

`overlay-12:147Ah` 本來就有已知的邊界問題（IDA 把 155 bytes 的函式報成 86），
兩個症狀應該同源。

這 12 筆的台帳註記已加上警告。**它們的狀態維持「已解讀」**——判讀的主體不受影響，
但引用缺口附近那幾行之前必須回去對 `.bin`。

## 往後的作法

讀完一支函式、要寫規格之前，先跑

```sh
python3 scripts/decode_gap_scan.py | grep '<模組> .* <函式位址>'
```

有輸出就把那幾行的原始位元組撈出來自己解一次。

## 明確不宣稱

- 這 12 處各自的正確解讀（本輪只標出位置）。
- 缺口的成因是 IDA 的分析、匯出腳本，還是原始碼裡真的有資料夾在中間。
