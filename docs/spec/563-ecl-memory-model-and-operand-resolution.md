# 第五百六十三輪：ECL 記憶體模型與 operand 解析（兩平台原始 bytes）

狀態：`READY`（限本文件列出的四支 helper、bank 邊界與 STOREVALUE 的具名
分支）。日期：2026-08-14

## 結論先行

ECL 直譯器的 operand 解析與記憶體寫入模型已由原始 bytes 讀出，兩平台對照
完成。這是「ECL work address 到底是什麼」這個問題的直接答案，先前是靠逐個
位址的觀察累積出來的。

### 1. operand 由三個平行陣列描述

`INTERPET` 的 handler 不自己解析 bytecode，而是呼叫 ECL2 的 `ADDRESSVALUE(i)`
解第 `i` 個 operand。operand 事先被拆進三個平行的 64 byte 陣列（DS 相對）：

| 陣列 | PC-98 | DOS | 內容 |
|---|---|---|---|
| operand code | `DS:A917h` | `DS:7685h` | operand 型別碼 |
| 高位元組 | `DS:A957h` | `DS:76C5h` | word 的高 byte |
| 低位元組 | `DS:A997h` | `DS:7705h` | word 的低 byte |

三個陣列間距都是 `40h`，索引用 `cbw` 符號延伸，所以是有號 byte 索引。
dispatcher 讀 opcode 的位址（PC-98 `DS:A891h`、DOS `DS:75FFh`）與 code 陣列
起點固定差 `86h`，兩平台一致。

### 2. `ADDRESSVALUE` 的分支＝已知的 operand code 契約

| operand code | 行為 | 對應既有文件 |
|---|---|---|
| `00h` | 直接回傳低位元組（零延伸） | byte literal |
| `01h`／`03h`／`80h` | 把兩個 byte 併成 word，再呼叫記憶體讀取 routine | 可讀 memory word／packed text |
| `02h`／`81h` | 把兩個 byte 併成 word 直接回傳 | word literal／string-memory word |
| 其他 | 不設定回傳值（未定義） | — |

這與 `docs/knowledge/gold-box-ecl-command-set.md` 記載的 operand 契約**完全
一致**，但先前那份是從 parser 工作推得的；本輪是原始碼路徑的直接證據。
新增的資訊是 `80h` 與 `01h/03h` 走同一條解參照路徑，而 `81h` 與 `02h` 一樣
是直接值。

`ADDFNC(a, b)` 不是算術加法，是 **byte pair 併成 word**：回傳 `(b << 8) + a`。
兩平台的實作逐指令相同（16 條指令，`retf 4`）。

### 3. ECL 位址空間的 bank 邊界

`STOREVALUE` 先呼叫一個分類器（PC-98 `0801h`／DOS `0735h`）把位址分成 5 類：

| bank | PC-98 範圍 | DOS 範圍 | 大小 | 寫入寬度與去處 |
|---:|---|---|---:|---|
| 0 | `4B00h..4EFFh` | 同左 | 1,024 | word，`es:[di + 2×addr + 6A00h]` |
| 1 | `7C00h..7FFFh` | 同左 | 1,024 | word，`es:[di + 2×addr + 0800h]` |
| 2 | `7A00h..7BFFh` | 同左 | 512 | word，`es:[di + 2×addr + 0C00h]` |
| 3 | `8000h..9E40h` | `8000h..9DFFh` | 7,745／7,680 | **byte**，`es:[di + addr − 8000h]` |
| 4 | 其餘 | 其餘 | — | 具名特例，見下節 |

各 bank 的位移常數都讓「該 bank 的起始位址」剛好落在陣列第 0 項
（`2×4B00h + 6A00h` 在 16 bit 下溢位為 0，其餘同理）。這同時證明了 bank 起點。

**bank 3 的上界兩平台不同**：DOS `9DFFh`、PC-98 `9E40h`。這是本輪唯一發現的
邊界差異，remake 不得假設兩版相同。bank 3 是載入中的 ECL script 緩衝區，
起點 `8000h` 與事件清冊的 `code_address = 8000h + block offset` 吻合，
且它是**唯一以 byte 為單位寫入**的 bank。

### 4. `STOREVALUE` 的具名特例（bank 4）

位址 `< BF68h` 的特例：

| ECL 位址 | 行為 |
|---|---|
| `00FBh`／`00FCh`／`00B1h` | 明確忽略（跳到結尾，不寫入） |
| `03DEh` | 寫 PC-98 `DS:BDDEh`／DOS `DS:8B4Ch`（word） |
| `00B8h` | 寫 PC-98 `DS:BDE0h`（word） |
| `00B9h` | 寫 PC-98 `DS:BDE2h`（word） |
| `5208h` | 寫 PC-98 `DS:7F11h`（byte） |

位址 `>= BF68h` 先減去 `BF68h` 再分派：

| ECL 位址 | 偏移 | PC-98 目的 | DOS 目的 | 附帶動作 |
|---|---|---|---|---|
| `C04Bh` | `0E3h` | `DS:A2A9h`（byte） | `DS:720Fh` | 置 dirty 旗標 |
| `C04Ch` | `0E4h` | `DS:A2AAh`（byte） | `DS:7210h` | 置 dirty 旗標 |
| `C04Dh` | `0E5h` | `DS:A2ABh`（byte） | `DS:7211h` | **值正規化**，見下 |
| `C059h` | `0F1h` | `DS:A87Ah`（byte） | — | 另一組 dirty 旗標 |
| `C05Fh` | `0F7h` | `DS:A87Bh`（byte） | — | 同上 |
| `D022h` | `10BAh` | 比較後即結束（no-op） | — | — |

dirty 旗標：PC-98 `DS:BDFAh`／`BDF9h`，DOS `DS:8B68h`。

`C04Bh`／`C04Ch`／`C04Dh` 的對應**獨立重現了第 548 輪**的結論，而且是從全掃
資料庫、兩平台各自讀出來的。新增的事實是 `C04Dh` 的寫入會做值正規化：

```text
0 → 0    1 → 2    2 → 4    3 → 6
其餘：value −= 4 後重試（等同對 4 取模再乘 2）
```

因此 ECL 寫入的 `C04D=3` 在原版內部是 `6`。**讀取端會除以 2 還原**（第 565 輪
確認），所以對 ECL 而言 round-trip 是一致的，remake 直接保存原值不會產生
ECL 可見的差異；`×2` 只有直接讀 `DS:A2ABh` 的繪圖端才看得到。

## 輸入與可重生

PC-98 `GAME.EXE`／`GAME.OVR`、DOS `START.EXE`／`GAME.OVR`，雜湊見 spec 559。
全部位址是 overlay-07（ECL2 單元）的 code-local offset。

```sh
tools/ida.sh py workplace/re-sweep/pc98/overlays/overlay-07.bin.i64 \
  dump_function.py /work/h-0E2B.json 0E2B          # STOREVALUE
tools/ida.sh py workplace/re-sweep/dos/overlays/overlay-07.bin.i64 \
  dump_function.py /work/d-tail.json 0E92 0F60     # DOS 的 C04B 系列
```

⚠ DOS 的 `STOREVALUE` 在 IDA 裡函式邊界被切短（`0D70h..0E98h`，實際延伸到
`0F34h` 之後），所以要**指定位址範圍**而不是依賴函式邊界。PC-98 那份完整。

## 這份規格明確不宣稱

- **`READVAR` 的語意**。它被呼叫 41 次（PC-98）／44 次（DOS），是四個 helper
  中最大的一支（520 bytes），本輪未讀。`GOTO`／`GOSUB` 用它而不用
  `ADDRESSVALUE`，推測與「直接讀 bytecode 裡的裸 word」有關，但這是
  `hypothesis`。
- **記憶體讀取 routine**（PC-98 `sub_FFC`）的實作。`ADDRESSVALUE` 對
  `01h/03h/80h` 呼叫它，但本輪沒有讀它的 bank 路由是否與 `STOREVALUE` 對稱。
- **bank 0..3 背後緩衝區的擁有者**。`DS:7F05h`／`7F09h`／`7F0Dh`／`7F12h` 是
  存放 far pointer 的位置，不是緩衝區本身；配置者尚未追。
- **具名特例的意義**。`00FBh` 為何被忽略、`03DEh`／`00B8h`／`00B9h` 是什麼、
  `C059h`／`C05Fh` 對應哪個顯示狀態，全部 `待解讀`。
- **`4B00h` 相對特例**（`4BFDh`／`4BFEh` 置旗標、`4BE6h` 觸發 `7F27h/7F28h`
  狀態轉移）本輪只記錄存在，未解讀語意。
