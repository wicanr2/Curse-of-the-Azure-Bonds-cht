# 第五百一十八輪：DOS `CALL 2E10h` 位址空間稽核

狀態：`READY`（僅限位址空間／候選排除；不是地圖規則完成）
日期：2026-08-09

## 目的

第 517 輪把下一個工作定為追蹤火刀戰後 ECL2 block 3 的
`CALL 2E10h`。本輪先在原始 DOS `START.EXE` 的 IDA Pro 資料庫中做有界、非
破壞性的候選稽核，確認是否可以直接把 ECL operand `2E10h` 對成 resident
程式的 `sub_2E10` 或 IDA 線性位址。結果是否定：目前資料庫沒有這個直接 code
對應；唯一的 little-endian raw 命中是字串資料。

這項結果只排除一條錯誤搜尋路徑，不等於證明 `CALL` 的完整 handler、目的地
producer 或 map service projection。間接指標、overlay dispatch、ECL interpreter
與 runtime 外部服務仍需分開追查。

## 輸入與可重現命令

| 輸入／工具 | 雜湊／版本 | 位址基準 |
|---|---|---|
| DOS `START.EXE` | SHA-256 `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` | MZ／IDA `seg000` 起始 `0x10000` |
| baseline `START.EXE.i64` | SHA-256 `9df802ee4ef71fb2eda83257e0ed2d87adf0ee2d10241d3bdbdc6bc369fe47eb` | IDA database linear EA／segment |
| extracted `GAME.OVR` overlay-02 | SHA-256 `ba41e4f437d5a86b09078f65a09197ebd6728e1d0b062edc13309174c48aa201`；完整 `GAME.OVR` SHA-256 `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215` | `GAME.OVR` extracted overlay-local offset |
| IDA | Docker image `ida-pro-9.4-ver2:uidfix-v1`；IDA Pro `9.4.0.260610` | 只操作 `/tmp` 的 database 副本 |
| 稽核腳本 | `scripts/ida/dos_start_2e10_candidate_audit.idc`；SHA-256 `220f1944e4cfcfd353a68c2019b83b0cdf3bb229e2be26ccaf8ab0637daef66b` | raw segment offset 與 IDA EA 並列 |
| overlay dispatch 腳本 | `scripts/ida/dos_overlay02_call_dispatch_audit.idc`；SHA-256 `1e43cb66bc5921a4d8484dec3e42882f03b34f7546ca19afe690f57a6908bd17` | extracted overlay-02 local offset |

腳本在 Docker 中把 baseline `.i64` 複製到 disposable 目錄，逐一掃描所有 IDA
segment 的 `LE16=0x2E10` 候選，再分別查詢 IDA EA `0x00002E10` 與
`0x00012E10` 的直接 code xref。它不改名、不改 comment、不改 type，也不回寫
baseline database。

## 原始結果

唯一候選位於：

```text
segment=seg043
IDA EA=0x00016634
raw bytes=10 2E 20 43 68 65 63 6B 20 69 6E 73 74 61 6C 6C
IDA item=db 16,'. Check install.'
code=0
direct xref count=0
```

`10 2E` 是 DOS 長度前綴字串資料的前兩個 byte，並非 `CALL 2E10h` 指令或
handler 位址。IDA report 同時得到：

```text
xref target IDA EA 0x00002E10 count=0
xref target IDA EA 0x00012E10 count=0
```

其中 `0x00012E10` 只是把 `2E10h` 暫以 `START.EXE` 的 IDA base `0x10000`
換算出的候選 EA；它不能取代 ECL external-call 位址空間。

## ECL dispatch 的下一層證據

在抽出的 `GAME.OVR overlay-02` disposable IDA database 中，連續解碼
`local 0x2EA0..0x30C0` 得到：

```text
local 0x2F23  2D FF 7F                 sub ax, 7FFFh
local 0x2F2C  3D 11 AE                 cmp ax, 0AE11h
local 0x2F39  9A 3E 00 7F 01           call far 017F:003Eh
local 0x2FCA  3D 01 32                 cmp ax, 3201h
local 0x2FFD  3D 1F 40                 cmp ax, 401Fh
```

`0x2E10 - 0x7FFF (mod 0x10000) = 0xAE11`，所以這段 code 精確支持
「ECL `CALL` operand 先被轉成 dispatch selector」；它也說明 handler 不會在
overlay-02 內以 `2E10h` literal 出現。第 518 輪報告產出時，`0x2F39` 的 far
pointer `017F:003Eh` 仍只證明 call-site operand，尚未證明其 control block、
重定位後 code offset 或 map side effect。第 519 輪已用 `START.EXE` 的 MZ
control block 與 `GAME.OVR` raw offset 閉合其**靜態** module 對應：它是 vector
6，指向 extracted overlay-30 local `07C6h`。完整 bytes、vector 對齊與該
routine 的存取邊界見
[`spec 519`](./519-dos-overlay-vector-to-cell-layer-accessor.md)。抽取檔名
`overlay-02` 與公開 reference 的 `ovr003` 標籤仍不強行合併；兩者的編號橋接
尚需 loader／overlay header 證明。

## 可證明與不可證明

### `exact`

- 在這份 baseline IDA database 的已建立 segment／xref 圖中，沒有直接可回查的
  resident `sub_2E10` code literal 或對 `0x2E10／0x12E10` 的 direct code xref。
- `0x16634` 的 `LE16=2E10h` 命中是 `seg043` 的非 code 字串資料。
- overlay-02 的 `0x2F23／0x2F2C` 連續指令將 ECL operand 正規化為 `AE11h`；
  `0x2F39` 的原始 far-call operand 是 `017F:003Eh`。

### `unknown`

- `CALL 2E10h` 的真正 overlay／interpreter dispatch、目的地 producer、
  `C04B..C04F` projection 與 redraw／map consumer。
- `017F:003Eh` 對應 routine 的 map plane、DS work-cell writer／consumer 與
  runtime side effect；overlay numbering `overlay-02` ↔ public `ovr003` 的
  橋接仍未證明。
- 是否有經暫存器、far pointer、table index 或 runtime service 完成的間接讀寫。

因此本輪不能把 `2E10h` 改名成地圖座標 writer，也不能把 overlay 22 的
`[di+4BF0h]` candidate 併入這條資料流。後續若要升級 `P0-1`，仍須在同一個
位址空間標示完整的 producer → projection → consumer，並以 DOSBox runtime
位置／方向／register trace 交叉驗證。

## 對工作清單的影響

11 個行為逆向主題與 4 個 fidelity／發行主題的數量不變；本輪把 P0-1 的搜尋
邊界已縮小為「overlay vector 之後的 map service／runtime handoff」，不是完成
一個主題。若後續文件出現「`START.EXE` 有 `sub_2E10`」、把 `0x16634` 當成
handler，或把 `017F:003Eh` 直接命名成秘密門／座標 writer，應以本規格與
[`spec 519`](./519-dos-overlay-vector-to-cell-layer-accessor.md) 訂正。

## 第 519 輪狀態修訂

第 519 輪以 raw control block `017F`、vector table 與 extracted overlay-30
長度／offset 重新核對，確認 `017F:003Eh` 的 vector index 是 **6**，bytes 為
`CD 3F C6 07 00`，目標為 overlay-30 local `07C6h`。相鄰 vector 7 的
`CD 3F 41 08 00` 才是 `0841h`；`0841h` 不是本次 ECL `2E10h` target。

`07C6h` 的目前 exact 邊界是：兩個 word 參數、`0..0Fh` bounded 16×16 index、
`DS:7206h` far pointer、`ES:[DI+0200h]` byte read，以及 `retf 4`。若交接摘要
或臨時筆記出現 `retf 6` 或 `+0100h` 的描述，均以本節與 spec 519 的 raw bytes
勘誤；該描述沒有進入正式 JSON、engine 或正常路徑 regression。這個 accessor
仍只屬 `strong inference` 的 cell-layer candidate，不能升格成 wall／terrain／
secret-door 語意。
