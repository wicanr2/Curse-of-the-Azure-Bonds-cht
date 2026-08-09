# 第五百二十輪：DOS movement state → overlay cell-layer bridge

狀態：`READY`（僅限靜態欄位／vector bridge；不是正常玩家路徑或秘密門完成）
日期：2026-08-09

## 結論先行

第 519 輪已把 `017F:003Eh` 對到 overlay-30 local `07C6h`。本輪再沿著所有
直接 `DS:7206h`、`DS:720Fh..7213h` 候選，找到一段可回查但仍未完全命名的
DOS 靜態 bridge：

```text
overlay-11 initialization
  → DS:720Fh=07h, DS:7210h=0Dh, DS:7211h=00h
  → overlay-07 local 1B3F 的 16×16 wrap/update
  → 017F:003Eh / vector 6 → overlay-30 local 07C6h
  → DS:7213h = vector 6 result
  → 017F:0034h / vector 4 → overlay-30 local 06BDh
  → DS:7212h = vector 4 result
```

另一個 overlay-28 consumer 也以 `DS:720F／7210` 呼叫 vector 6，將結果寫入
`DS:7213h` 並比較 `7Fh`；稍後它明確呼叫 `017F:0043h`，也就是 vector 7 →
overlay-30 local `0841h`。因此 `0841h` 是獨立的後續路徑，不能再誤當成
ECL `2E10h` 的 vector 6 target。

這輪仍不能把欄位命名成座標、牆、地形、秘密門或劇情旗標，也沒有修改 engine／
JSON。overlay-07 local `1B3F` 在該 overlay 的 direct IDA code xref 為零，
其正常輸入 caller／indirect entry 仍需追；第 521 輪已將 `DS:7206h` 的
`0A54:0329h` owner 對到 Borland `GetMem(Pointer &,Word)`，但配置後 buffer
的填入／plane layout、`C04B..C04F` projection 與 DOSBox runtime handoff 仍未
閉合。

## 輸入與工具 provenance

| 輸入／工具 | SHA-256／版本 | 位址基準 |
|---|---|---|
| DOS `START.EXE` | `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` | MZ raw file |
| baseline `START.EXE.i64` | `9df802ee4ef71fb2eda83257e0ed2d87adf0ee2d10241d3bdbdc6bc369fe47eb` | IDA linear EA／dseg |
| 完整 `GAME.OVR` | `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215` | raw file offset |
| extracted overlay-07 | `5483c71f98c5dc668d7d307c18a6b071dcfc42fcba9d62eccb657600e7265125` | overlay-local offset |
| extracted overlay-11 | `a20b0e529b68eac4ab4638491a1e4e7eb55fef3ad988b4879502954a1fdfc6b7` | overlay-local offset |
| extracted overlay-14 | `437928ced7453a95613557c23742e3cec446df5186e0513b3feedacd3d87ffb4` | overlay-local offset |
| extracted overlay-28 | `6fe4be1fbfd9976f739fc4a15636eb8f557633a39957f20373480e66ad83b675` | overlay-local offset |
| extracted overlay-30 | `444893f6d239cc57f555287786e3704bc801dd63f3d1f4d4ac14e8742652d468` | overlay-local offset |
| IDA Pro | Docker image `ida-pro-9.4-ver2:uidfix-v1`；`9.4.0.260610` | 每個 overlay 各自 disposable database |
| movement report | `331602a43268a99b772734a1568c7208327d10659a102619c040aeb089158293` | `/tmp/dos-overlay07-movement-audit.txt` |
| all-overlay DS candidate script | `619fb467426c703a27bd22dc1f8905996c6e6f1618f3765f7f456861ea860305` | [`dos_overlay_ds_field_audit.idc`](../../scripts/ida/dos_overlay_ds_field_audit.idc) |
| all-overlay DS candidate report | `eec8d38a6cf128c50398156a94078bf1c18ccafae3e4f09bdc36d11073bbe006` | `/tmp/dos-overlay-ds-fields-all.txt` |
| overlay context script | `187b901bef1e29e6a9feb0fd7d9e0bf8bbecc3d61005df1644276b2c620db4c9` | [`dos_overlay_ds_context_audit.idc`](../../scripts/ida/dos_overlay_ds_context_audit.idc) |
| overlay-07 context report | `df731b33527f8faab0d63f45e6829e62e62708ae5d3ed834081dbad6dac8f60a` | overlay-local continuous ranges |
| overlay-11 context report | `7a6760279c9c4322a0f1744f921bdbd1872914ba1ca080d2fb054a4a7e21e770` | overlay-local continuous ranges |
| overlay-14 context report | `6d35081910041e8a10d6d47cb98a8139a74a4c8d6220744804724551674d646f` | overlay-local continuous ranges |
| overlay-28 context report | `66487b4d189b42d0b1cb1dbbce018eb17eae3e44ffd14669f7667c2e1b0dddfe` | overlay-local continuous ranges |
| overlay-30 context report | `484e8179db6198f969681ca06c7018b9dbd5a05bdfd5e66b1ba15e9ae073c333` | overlay-local continuous ranges |

`GAME.OVR` raw offset 與 START control block 的 mapping 是獨立以第一個 32-byte
prefix 搜尋確認：

| extracted overlay | `GAME.OVR` raw offset | length | START control raw offset | image paragraph |
|---|---:|---:|---:|---:|
| overlay-07 | `0x802F` | `0x23A8` | `0x0E60` | `006Bh` |
| overlay-11 | `0xF4FA` | `0x08E3` | `0x10B0` | `0090h` |
| overlay-14 | `0x18375` | `0x0EC2` | `0x14D0` | `00D2h` |
| overlay-28 | `0x3D494` | `0x01BE` | `0x1F00` | `0175h` |
| overlay-30 | `0x3DF87` | `0x147F` | `0x1FA0` | `017Fh` |

START control raw offset 由 `MZ header 0x7B0 + (image paragraph << 4)` 計算；
overlay-local offset、control segment 與 ECL work address 不互換。

## overlay-07 local `1B3Fh`：欄位 update 與兩個 vector call

IDA 解出的連續 bytes 為：

```text
1B3F  55 89 E5                    push bp; mov bp,sp
1B42  A0 11 72                    AL = DS:7211h
1B47  3D 00 00                    若 direction == 0
1B4C  80 3E 10 72 00              檢查 DS:7210h 是否 > 0
1B53  FE 0E 10 72                 DS:7210h--
1B59  C6 06 10 72 0F              否則 DS:7210h = 0Fh
1B60  3D 02 00                    若 direction == 2
1B65  80 3E 0F 72 0F              檢查 DS:720Fh 是否 < 0Fh
1B6C  FE 06 0F 72                 DS:720Fh++
1B72  C6 06 0F 72 00              否則 DS:720Fh = 00h
1B79  3D 04 00                    若 direction == 4
1B7E  80 3E 10 72 0F              檢查 DS:7210h 是否 < 0Fh
1B85  FE 06 10 72                 DS:7210h++
1B8B  C6 06 10 72 00              否則 DS:7210h = 00h
1B92  3D 06 00                    若 direction == 6
1B97  80 3E 0F 72 00              檢查 DS:720Fh 是否 > 0
1B9E  FE 0E 0F 72                 DS:720Fh--
1BA4  C6 06 0F 72 0F              否則 DS:720Fh = 0Fh
1BA9  A0 0F 72; 50                push DS:720Fh
1BAD  A0 10 72; 50                push DS:7210h
1BB1  9A 3E 00 7F 01              call far 017F:003Eh
1BB6  A2 13 72                    DS:7213h = AL
1BB9  A0 0F 72; 50                push DS:720Fh
1BBD  A0 10 72; 50                push DS:7210h
1BC1  A0 11 72; 50                push DS:7211h
1BC5  9A 34 00 7F 01              call far 017F:0034h
1BCA  A2 12 72                    DS:7212h = AL
1BCD  C6 06 68 8B 01              DS:8B68h = 1
1BD5  CB                         retf
```

可證明：

- `exact`：該 routine 對 `DS:7211h` 的四個值 `0／2／4／6` 做兩個 byte 欄位的
  `0..0Fh` 循環更新；兩次 far-call 的 raw operands 與結果 writer 如上。
- `strong inference`：因為 `DS:720F／7210` 具有 16×16 邊界、初始值為 `7／0Dh`，
  且直接作為 vector 6 的兩參數，兩欄很可能是 map index／座標；目前仍不把它們
  命名成 X／Y，也不把 direction code 直接命名成北東南西。
- `unknown`（第 520 輪當時）：`1B3Fh` 的正常輸入 caller／indirect entry。當時此
  overlay 的 direct IDA code xref report 對 target `1B3Fh` 為 `0`，所以不能只靠
  這個 routine 宣稱已完成 DOSBox 正常移動路徑。第 523 輪已在
  `docs/spec/523-dos-overlay07-vector26-entry.md` 補上 control vector 26 與
  overlay-02 靜態 caller；本 spec 不再把「無 direct xref」解讀成「無 caller」。

## overlay-11 初始化：buffer candidate 與起始欄位

overlay-11 local `00E9..00F5`：

```text
00E9  BF 06 72                    DI = 7206h
00EC  1E 57                       push DS; push DI
00EE  B8 00 04; 50                push 0400h
00F2  9A 29 03 54 0A              call far 0A54:0329h
```

這只能證明當時 call-site 把 `DS:7206h` 位址與 `0400h` 參數傳給該外部 routine；
第 521 輪已由 resident IDA symbol／入口 bytes 確認 target 是 Borland
`GetMem(Pointer &,Word)`，因此「呼叫慣例／owner 完全未知」不再是當前斷言。
配置後 buffer 的填入、清除時機與內容語意仍未知。相鄰初始化 block 在
`03F2..03FC` 及 `078E..0798` 寫入：

```text
DS:720Fh = 07h
DS:7210h = 0Dh
DS:7211h = 00h 或 02h
```

這支持「16×16 index 的初始 state」推論，但不替代原版 runtime 的位置／方向
trace，也不能把 `7／0Dh` 當成所有地圖的固定出生點。

## overlay-30 vector 4：與 vector 6 分開的 layer read

START control `017F` 的 vector 4（far pointer `017F:0034h`）指向 overlay-30
local `06BDh`。連續 code 的最小可證明語意是：

- 接收三個 word 參數；`[bp+6]` 只分支 `0／2／4／6`，`[bp+8]` 與 `[bp+0Ah]`
  形成 16×16 index，並由 guard／clamp 限制到 `0..0Fh`。
- selector `0／2` 從 `ES:[DI+000h]` 取高／低 nibble；selector `4／6` 從
  `ES:[DI+100h]` 取高／低 nibble；最後以 `retf 6` 返回。
- `overlay-07:1BC5` 的三個參數正好是兩個 DS 欄位加 `DS:7211h`，但這只證明
  呼叫形狀，不能單獨命名 nibble 為 wall／door／方向阻擋。

vector 6 的 `07C6h` 則是第 519 輪記錄的兩參數、`ES:[DI+200h]` byte read；
兩者的 plane／用途保持分開。

## overlay-28 consumer 與 `0841h` 勘誤

overlay-28 control segment 是 `0175h`；local `00CCh` 的連續 code：

```text
00F7  A0 0F 72; 50                push DS:720Fh
00FB  A0 10 72; 50                push DS:7210h
00FF  9A 3E 00 7F 01              call far 017F:003Eh
0104  A2 13 72                    DS:7213h = AL
0107  80 3E 13 72 7F              cmp DS:7213h, 7Fh
...
018F  A0 0F 72; 50                push DS:720Fh
0193  A0 10 72; 50                push DS:7210h
0197  A0 11 72; 50                push DS:7211h
019B  9A 43 00 7F 01              call far 017F:0043h
```

最後一個 far pointer 是 vector 7；START control bytes 為
`CD 3F 41 08 00`，目標是 overlay-30 local `0841h`。`0841h` 會進入另一個
較長的 routine，並非 vector 6 的 `07C6h` cell read。這是本輪最重要的錯誤斷言
清理：不能因為同一個 control segment 或相鄰 offset，就把 vector 7 的 renderer／
map routine 合併成 ECL `2E10h` handler。

## overlay-14 的獨立 `+300h` 候選

overlay-14 的 control segment 是 `00D2h`，其 vector 4 目標為 local `003Eh`。
該 routine 也使用 `DS:7206h` 與 16-stride index，但讀／寫的是
`ES:[DI+300h]`，依 selector `0／2／4／6` 做 mask `FCh／F3h／CFh／3Fh` 的讀寫。
這是獨立 module／control segment 的 `+300h` layer candidate；不能把它直接等同
overlay-30 vector 4 的 `+000h／+100h`，更不能直接等同 ECL work `4BF0h` 或
secret-door plane。

## 推論等級與剩餘工作

### `exact`

- overlay-07 `1B3Fh` 的 DS 欄位更新、範圍與兩個 far-call operand。
- overlay-11 對 DS:7206／0400h 的 call-site，以及 `720F／7210／7211` 的
  初始化 bytes。
- overlay-30 vector 4／6 的 control-vector mapping、參數數量、index 及 layer
  offsets；overlay-28 的 vector 6 consumer 與 vector 7 call-site。

### `strong inference`

- `DS:720F／7210` 是 16×16 map index／座標 candidate；`DS:7211` 是 direction／
  side selector candidate。
- `DS:7206` 指向一個至少被以 `0400h` 大小參數處理的 map-like buffer candidate。

### `unknown`

- `DS:7206` 配置後 buffer 的實際 writer、清除時機與 `+000/+100/+200/+300`
  plane layout。
- overlay-07 `1B3Fh` 的 normal input producer／control-loader runtime；第 523 輪
  已關閉 static vector／dispatcher entry，但本輪仍未取得正常鍵盤 trace。
- `DS:7212／7213` 對應的正式規則、`7Fh` sentinel 意義、`C04B..C04F` projection、
  ECL external handoff 與 DOSBox runtime 結果。

因此 11 個行為逆向主題與 4 個 fidelity／發行主題數量不變；本輪只縮小 P0-1
的欄位資料流。下一步是追 `GetMem` 後的 buffer writer／plane、overlay-07 的
control-loader／正常輸入 trace、overlay-14 `+300h` writer 與原版 DOSBox map／座標／方向 trace。在
這些證據閉合前，不新增
secret-door／search JSON、不改 movement predicate，也不把已有 coordinate-assisted
handoff 升格成 exact。

## 第 521 輪狀態修訂

第 521 輪已把本規格原本的「`0A54:0329h` callee 呼叫慣例未知」縮小：在
`START.EXE.i64` 的 `+1000h` selector mapping 下，該 target 是 IDA
`seg050:1A54:0329h`／EA `0x1A869`，原始符號為 Borland
`GetMem(Pointer &,Word)`。overlay-11 的 `DS:7206h`＋`0400h` 因而可標成
`GetMem` 配置 1 KiB pointer 的 `strong inference`；buffer writer／plane layout／
map projection 仍未知。完整 evidence 見
[`spec 521`](./521-dos-getmem-buffer-owner.md)。
