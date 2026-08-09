# 第五百二十三輪：DOS control vector 26 → overlay-07 `1B3Fh`

狀態：`READY`（僅限 control vector 與 overlay-02 靜態 caller；不是正常鍵盤輸入、
地圖移動或 runtime handoff 完成）
日期：2026-08-09

## 結論先行

第 520–522 輪把 overlay-07 `1B3Fh` 的資料欄位更新、`DS:7206h` 配置與四平面
writer 幾何逐步縮小，但曾把「找不到 overlay-07 內的 direct xref」誤留成
「`1B3Fh` 的 entry 尚未找到」。本輪補查 `START.EXE` 的 control block 與
overlay-02 的 dispatch branch，得到下列靜態事實：

```text
START.EXE raw 0x0F02
  → control runtime selector 006Bh
  → vector index 26（每項 5 bytes）
  → CD 3F 3F 1B 00
  → overlay-07 local target 1B3Fh

overlay-02 local 2FFD..3007
  → cmp ax, 401Fh
  → call far 006B:00A2h
  → vector 26
```

因此目前可精確宣稱：`overlay-07 local 1B3Fh` 有一條經 `006B` control vector
table 的靜態入口，且 overlay-02 在 `401Fh` 分支呼叫該 vector。這一輪只關閉
「vector／外部 dispatcher entry」子邊界；尚未證明 `401Fh` 是普通鍵盤按鍵的
唯一來源，也尚未證明 `1B3Fh` 的 runtime 返回值如何投影成目的地、座標、方向或
重繪。11 個行為資料流與 4 個 fidelity／發行主題的盤點數量不變。

## 1. 證據來源與位址空間

所有原始檔唯讀保存；IDA 只操作 Docker 中的 disposable `.i64` 複本。`START.EXE`
的 control block 與 `GAME.OVR` overlay 必須分開看，不能把相同的 local offset
直接當成同一個 linear address。

| 輸入／報告 | SHA-256／版本 | 位址基準 |
|---|---|---|
| DOS `START.EXE` | `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` | MZ raw file offset；IDA image base `0x10000` |
| `START.EXE.i64` disposable input | `9df802ee4ef71fb2eda83257e0ed2d87adf0ee2d10241d3bdbdc6bc369fe47eb` | IDA selector／EA |
| `GAME.OVR` | `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215` | extracted overlay file |
| `overlay-02.bin` | `ba41e4f437d5a86b09078f65a09197ebd6728e1d0b062edc13309174c48aa201` | overlay-02 local offset |
| `overlay-07.bin` | `5483c71f98c5dc668d7d307c18a6b071dcfc42fcba9d62eccb657600e7265125` | overlay-07 local offset |
| IDA | `IDA Pro 9.4`，`ida-pro-9.4-ver2:uidfix-v1` | Docker disposable copy |
| control-vector IDC | `b4484402ce42431f8d816d5518fe085f68f75502b48f4eba57cfa259515d1654` | `scripts/ida/dos_start_overlay07_vector26_audit.idc` |
| control-vector report | `db2f8242c1867fa38ff8773d36fded538b6dd4a8613fe1f050f0c0222ed3dcc6` | raw file offset＋IDA EA＋control selector |
| overlay-02 IDC | `1e43cb66bc5921a4d8484dec3e42882f03b34f7546ca19afe690f57a6908bd17` | `scripts/ida/dos_overlay02_call_dispatch_audit.idc` |
| overlay-02 report | `b5081bad0aa1b9e6ba0accfce1a8bd562f98895806a7176009605752af93e4fe` | overlay-02 local offset |
| overlay-07 indirect IDC | `f073282fd04ba196abe59715d5c1729f2982242a79c52fa9d67c3ba072b1444a` | `scripts/ida/dos_overlay07_indirect_dispatch_audit.idc` |
| overlay-07 indirect report | `86dfbf6a7545a26cbbed77a2d4d3227705980e7d83bd4148fe29ca128b7634cb` | overlay-07 local offset |

### 1.1 `START.EXE` control table

IDA report 使用 MZ header `0x07B0` 與 image base `0x10000` 對齊：

- control block raw file offset：`0x0E60`。
- control 的 runtime selector：`006Bh`。
- 對應 IDA segment selector：`106Bh`、`seg008`，IDA EA `0x106B0`。
- vector table raw file offset：`0x0E80`，IDA EA `0x106D0`。
- 每個 vector entry 是 5 bytes；本輪只把 `CD 3F` 與後面的 little-endian
  local offset 當作 byte-level 格式，不替最後的 flags／loader 行為命名。

vector index 26 的連續 bytes 與定位如下：

| 欄位 | 值 |
|---|---|
| raw file offset | `0x0F02` |
| IDA EA | `0x10752` |
| runtime control selector | `006Bh` |
| vector index | `26`（zero-based） |
| bytes | `CD 3F 3F 1B 00` |
| decoded target | overlay-07 local `1B3Fh` |

這是 control table 的直接 bytes 證據，不是 overlay-07 內的 direct code xref。
同一份報告也保留所有相鄰 vector entry，供日後核對 `006B:00A2h` 的 table index。

### 1.2 overlay-02 的 caller branch

在 `overlay-02` local `2FFD..3007` 找到連續原始指令：

```text
2FFD  3D 1F 40       cmp ax,401Fh
3000  75 07          jnz short 3009h
3002  9A A2 00 6B 00 call far 006B:00A2h
3007  EB 66          jmp 306Fh
```

control header raw `0x0E60` 對應 runtime `006B:0000h`，因此 vector table 的
runtime 起點是 `006B:0020h`（raw `0x0E80`）；vector 26 的 table offset 是
`0x20 + 26 × 5 = 0xA2`。所以 `call far 006B:00A2h` 正好指向 vector 26
entry，而不是另一個未命名的 control 位址。這個關係同時由 raw file offset、
runtime selector／offset、IDA EA 與 5-byte stride 閉合：

```text
006B:0020h + 26 × 5 = 006B:00A2h
raw 0x0E80 + 26 × 5 = raw 0x0F02
IDA 0x106D0 + 26 × 5 = IDA 0x10752
```

因此經由同一份 overlay-02 report 直接看到的靜態事實是：`401Fh` 比較成功後，
CPU 執行 `call far 006B:00A2h`，而該 vector entry 的 bytes 是
`CD 3F 3F 1B 00`，指向 overlay-07 local `1B3Fh`。這仍不等於已完成
runtime／loader trace；返回值、按鍵 producer 與 routine 的地圖副作用仍待驗證。

> **重要勘誤：** `006B:00A2h` 是 runtime control-vector table 的 offset；
> `0x0F02` 是 raw file offset；`0x10752` 是 IDA EA。三者不能在文件中互換，
> 但本輪已以 MZ header、control 起點與 stride 將它們對齊。

### 1.3 為什麼 overlay-07 內沒有 direct xref 仍不矛盾

overlay-07 的 IDA／raw audit 目前輸出：

```text
target_local=0x1B3F direct_cref_count=0
raw_LE16_target_1B3F_count=0
```

這只能表示在 overlay-07 自身的 IDA direct cref 與有界 raw little-endian literal
掃描中沒有命中；它不能表示 routine 沒有 caller。caller 位於外部的 `START.EXE`
control／overlay-02 dispatcher，正是本輪補上的證據。indirect audit 只搜尋真正
以 `FF` opcode 開頭、且含記憶體運算元的間接 call／jump，沒有把一般 far call
誤列成 indirect candidate；它仍不涵蓋 loader 在 runtime 以暫存器／指標形成的
呼叫。

## 2. 推論等級與尚未解讀的部分

| 結論 | 等級 | 證據邊界 |
|---|---|---|
| control raw `0x0F02` 的 5 bytes 是一個 control entry | `exact` | 連續 raw bytes、5-byte stride、IDA EA 與可重現報告 |
| 該 entry 的 local target word 是 overlay-07 `1B3Fh` | `exact` | `CD 3F 3F 1B 00` 的 little-endian 解碼；overlay-07 同一 local offset |
| overlay-02 的 `401Fh` branch 執行 far call `006B:00A2h` | `exact` | local `2FFD..3007` 的 raw bytes與 IDA 指令邊界 |
| `401Fh` 是與 ECL `C01Eh` 相關的移動 selector | `strong inference` | 既有 ECL／dispatch 對照；本輪未以正常輸入 runtime trace 閉合 |
| `401Fh` 代表普通鍵盤的「向前移動」且唯一抵達 `1B3Fh` | `unknown` | 仍缺輸入 producer、control loader trace、返回值與相鄰按鍵比較 |
| `DS:7212／7213` 的正式欄位名稱、`C04B..C04F` projection、目的地／座標／方向 | `unknown` | 本輪沒有新增 consumer 或 runtime 證據 |

## 3. 對 remake 與 worklist 的影響

本輪只更新證據與工作清單，不修改 engine、CoAB JSON、movement graph 或
`set_map_position` adapter。原因是已證實「entry 存在」仍不足以證明正常玩家輸入
和地圖 state 的完整資料流。

目前 P0-1 應寫成：

1. `.dax` record／解碼輸出 → 四平面正式 map 語意；
2. `006B` control／overlay-02 dispatcher → overlay-07 `1B3Fh` 的正常輸入來源
   與 loader runtime；
3. `DS:7212／7213`、`7Fh` sentinel、vector 7 `0841h` 的 consumer；
4. `C04B..C04F`／map service projection；
5. DOSBox 原版目的地、座標、方向、重繪與戰後續跑 trace。

只有第 2 項的 **static vector／caller 子邊界** 已達 `READY`；「正常鍵盤」和
「完整 movement handoff」仍是未完成工作。正常玩家路徑、引擎接線與 regression
必須等資料流／runtime 閉合後再做，不能用 direct-entry 或 `401Fh` 常數特判掩蓋。

## 4. 下一個最小可重現工作

在 Docker 中以同一份 `START.EXE`／`GAME.OVR` 複本：

1. 追 control manager 對 `006B:00A2h` 的 loader／`CD 3F` stub 執行，確認它如何
   取出 vector 26 target；保留 runtime segment、overlay-local offset 與 return
   register。
2. 從原版 DOSBox 的實際鍵盤輸入分別觸發前進、左右轉、後退與無效方向，捕捉
   `401Fh`／相鄰 selector、`DS:720F..7213`、`DS:7206h` 與位置／重繪變化。
3. 把動態結果與已知 `1B3Fh` 靜態 routine 的 writer／consumer 串成一份正常
   玩家路徑 trace；在此之前不新增地圖座標規則。

本 spec 的 IDA scripts：

- [`dos_start_overlay07_vector26_audit.idc`](../../scripts/ida/dos_start_overlay07_vector26_audit.idc)
- [`dos_overlay02_call_dispatch_audit.idc`](../../scripts/ida/dos_overlay02_call_dispatch_audit.idc)
- [`dos_overlay07_indirect_dispatch_audit.idc`](../../scripts/ida/dos_overlay07_indirect_dispatch_audit.idc)
