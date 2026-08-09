# 第五百一十九輪：DOS overlay vector → 16×16 cell-layer accessor

狀態：`READY`（僅限靜態 dispatch／資料存取邊界；不是地圖 handoff 完成）
日期：2026-08-09

## 結論先行

第 518 輪只能把 ECL `CALL 2E10h` 追到 raw far call
`017F:003Eh`。本輪在 Docker 內以 DOS `START.EXE` 的 MZ image、原始
`GAME.OVR` offset 與 IDA Pro 9.4 disposable database 交叉核對，閉合了下列
**靜態**資料流：

```text
ECL operand 2E10h
  → overlay-02 local 2F23／2F2C 的 selector AE11h
  → START control block 017F 的 vector 6（offset 003Eh）
  → GAME.OVR overlay-30 local 07C6h
  → DS:7206 far pointer + 16×16 bounded index
  → ES:[DI+0200h] 的 byte read
```

這只證明「selector 對應哪一個 overlay control vector，以及該 vector 的 code
如何取值」。它尚未證明 `DS:7206` 是哪一張地圖、`+0200h` 是哪一個 plane，
也尚未找到 `DS:720F／7210` 的真正 writer、`C04B..C04F` 的 projection、
`CALL 2E10h` 的 runtime redraw／位置 consumer 或秘密門狀態。因此 P0-1 仍是
未完成的行為主題，不能新增 `secret_door`、地圖座標或 movement 特判。

## 輸入、版本與位址基準

| 輸入／工具 | SHA-256／版本 | 位址基準與用途 |
|---|---|---|
| DOS `START.EXE` | `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` | MZ raw file；header `0x7B0` bytes |
| baseline `START.EXE.i64` | `9df802ee4ef71fb2eda83257e0ed2d87adf0ee2d10241d3bdbdc6bc369fe47eb` | IDA database；只讀複製到 Docker `/tmp` |
| 完整 DOS `GAME.OVR` | `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215` | raw file offset |
| extracted `overlay-02.bin` | `ba41e4f437d5a86b09078f65a09197ebd6728e1d0b062edc13309174c48aa201` | overlay-local offset；dispatch call-site |
| extracted `overlay-30.bin` | `444893f6d239cc57f555287786e3704bc801dd63f3d1f4d4ac14e8742652d468` | overlay-local offset；cell-layer accessor candidate |
| IDA Pro | Docker image `ida-pro-9.4-ver2:uidfix-v1`；`9.4.0.260610` | 16-bit segment decode；不改原始 binary／baseline database |
| overlay-30 稽核腳本 | `10f49e0799e9925d5407ac3569b1c2ec1b4fa75812be5bd2c08560092cab1f10` | [`dos_overlay30_vector_audit.idc`](../../scripts/ida/dos_overlay30_vector_audit.idc) |
| START dseg 稽核腳本 | `3b736d357544d31a5048cce00cddfe73ad712956e860af318dae16354b2ac6f5` | [`dos_start_dseg_offset_audit.idc`](../../scripts/ida/dos_start_dseg_offset_audit.idc) |
| overlay-30 IDA report | `a5dcac047d3b8174b7e9d7e469501397b8d941f90c1643ae50c3530e37b0d7cf` | `/tmp/dos-overlay30-vector-audit.txt` |
| START dseg IDA report | `f053842090ddc59c21b5193d2877c6ee76cc36c0bd3135d15582f20c6f454a99` | `/tmp/dos-start-dseg-offsets.txt` |

`GAME.OVR` 的 overlay-30 extracted bytes 在 raw file offset `0x3DF87` 找到一個
唯一吻合的 `0x147F` bytes 區段；這不是只依 extracted filename 推測的 module
對應。overlay-02 則是 raw file offset `0x8A6`、長度 `0x3A20` 的先前已保存
區段。所有 overlay-local offset、MZ image offset、IDA linear EA 與 ECL work
address 在本規格中分開記錄。

## `START.EXE` control block 與 vector 對齊

MZ header 為 `0x7B0` bytes。控制區段的 image paragraph 是 `017Fh`，所以其 raw
file offset 是：

```text
0x7B0 + (0x017F << 4) = 0x1FA0
```

raw `0x1FA0` 的控制資料為：

```text
cd 3f 00 00 87 df 03 00 7f 14 72 00 0f 00 79 01
```

在本 binary 的可回查欄位中可精確讀出：

| control offset | raw value | 最小可證明語意 |
|---:|---|---|
| `+00` | `CD 3F` | Turbo Pascal overlay control stub 的 `INT 3F` bytes |
| `+04` | `87 DF 03 00` | `GAME.OVR` raw file offset `0x3DF87`；吻合 overlay-30 起點 |
| `+08` | `7F 14` | code length `0x147F`；吻合 overlay-30 extracted length |
| `+0A` | `72 00` | relocation length `0x0072` |
| `+0C` | `0F 00` | vector count `15` |
| `+0E` | `79 01` | 下一個 control block 的 segment 欄位 `0179h` |

vector table 從 control `+20h`、raw `0x1FC0` 開始；每個 vector 佔 5 bytes，
以下編號為 zero-based：

```text
vector 0  @ raw 0x1FC0: CD 3F 78 14 00 → local 0x1478
vector 1  @ raw 0x1FC5: CD 3F FF FF 00
vector 2  @ raw 0x1FCA: CD 3F 6B 01 00 → local 0x016B
vector 3  @ raw 0x1FCF: CD 3F 8C 01 00 → local 0x018C
vector 4  @ raw 0x1FD4: CD 3F BD 06 00 → local 0x06BD
vector 5  @ raw 0x1FD9: CD 3F 87 05 00 → local 0x0587
vector 6  @ raw 0x1FDE: CD 3F C6 07 00 → local 0x07C6
vector 7  @ raw 0x1FE3: CD 3F 41 08 00 → local 0x0841
```

`017F:003Eh` 是 `control + 20h + 6×5`，所以它是 vector 6，目標為
overlay-30 local `07C6h`。`0841h` 是相鄰的 vector 7，**不是本輪 ECL
`2E10h` 的目的地**。這項一格偏移的勘誤已在原始 bytes 對齊後確認；不得再把
`0841h`、vector 7 的 routine 或另一個 plane read 混入本資料流。

Turbo Pascal 7 Programmer's Reference 的 overlay manager 章節可作一般格式
背景參考，但本 binary 的 control bytes、raw file offset 與 extracted length
才是本輪 module 對應的主要證據：
<https://turbopascal.org/wp-content/uploads/Turbo_Pascal_Version_7.0_Programmers_Reference_1992.pdf>

## overlay-02 的 selector 與參數傳遞

在 extracted overlay-02 的 disposable IDA database，連續指令為：

```text
local 0x2F23  2D FF 7F                 sub ax, 7FFFh
local 0x2F2C  3D 11 AE                 cmp ax, 0AE11h
local 0x2F31  ...                      讀 DS:720F、DS:7210 並壓入兩個參數
local 0x2F39  9A 3E 00 7F 01           call far 017F:003Eh
local 0x2F3E  ...                      將 AL 寫入 DS:7213
```

`2E10h − 7FFFh (mod 10000h) = AE11h`，因此這段 code exact 證明 ECL
operand 先經過 selector normalization；它不是在 overlay-02 內直接以
`2E10h` literal 呼叫 resident `sub_2E10`。

同一個 dispatch routine 另在 `local 0x2F9F`／後續路徑呼叫
`017F:0034h`，傳入 `DS:720F／7210／7211` 三個值，結果寫入 `DS:7212`。
這是 vector 4 → overlay-30 local `06BDh` 的另一條資料流，不能與 vector 6
的 `DS:7213` 結果合併；兩組 DS 欄位的 domain 仍是 `unknown`。

## overlay-30 local `07C6h` 的 raw／IDA 證據

連續 raw bytes／IDA 指令（位址皆為 overlay-30 local offset）為：

```text
07C6  55 89 E5 83 EC 02           push bp; mov bp,sp; sub sp,2
07CC  8A 46 08; 50                讀 [bp+8] 並壓入 word
07D0  8A 46 06; 50                讀 [bp+6] 並壓入 word
07D4  0E E8 7E FD                 push cs; call local 0556h
07D8  08 C0                       or al,al
07DA  75 14                       非零時跳到 07F0h
07DC  80 3E 5E 8B 00             比較 DS:8B5Eh 與 0
07E1  74 07                       若為 0，跳到 07EAh
07E3  80 3E 5E 8B 0A             比較 DS:8B5Eh 與 0Ah
07E8  75 06                       不等於 0Ah 時跳到 07F0h
07EA  C6 46 FF 00                local result = 0
07EE  EB 48                       跳到 0838h
07F0  80 7E 08 0F                [bp+8] > 0Fh 時進行邊界處理
07F6  C6 46 08 00                將高位參數限制到 0
07FC  80 7E 08 00                [bp+8] < 0 時進行邊界處理
0802  C6 46 08 0F                將高位參數限制到 0Fh
0806  80 7E 06 0F                [bp+6] > 0Fh 時進行邊界處理
080C  C6 46 06 00                將低位參數限制到 0
0810  80 7E 06 00                [bp+6] < 0 時進行邊界處理
0816  C6 46 06 0F                將低位參數限制到 0Fh
081A  8A 46 08; 98; 8B D0       DX = sign-extended [bp+8]
0820  8A 46 06; 98                AX = sign-extended [bp+6]
0824  B1 04; D3 E0                AX <<= 4
0828  C4 3E 06 72                LES DI, DS:7206h
082C  03 F8; 03 FA                DI += AX; DI += DX
0830  26 8A 85 00 02             AL = ES:[DI+0200h]
0835  88 46 FF                儲存 local result
0838  8A 46 FF                回傳 AL
083B  89 EC; 5D                  還原 stack／BP
083E  CA 04 00                  retf 4
```

因此可標為：

- `exact`：此 routine 接受兩個 word 參數；將兩參數限制到 `0..0Fh`；以
  `([bp+6] << 4) + [bp+8]` 形成 16×16 範圍的 byte index；從 `DS:7206h`
  取得 far pointer，讀取其 `+0200h` plane offset 的 byte；正常返回為
  `retf 4`。前置的 `local 0556h`／`DS:8B5Eh` guard 可能提早返回 0，不能
  將每次呼叫簡化成必然 read。
- `strong inference`：這是與 `WALLDEF.dax`／`Load3DMap` 同一 overlay 的
  16×16 indexed cell-layer accessor candidate；`+0200h` 很可能是某個 map
  layer，但尚不能命名為 wall、terrain、door 或 secret-door。
- `unknown`：`DS:7206h` far pointer 的初始化與 owner、`+0200h` 的 plane
  定義、`DS:720F／7210` 的 producer、`DS:7213` 的 consumer，以及它們與
  `C04B..C04F`／ECL handoff 的關係。

注意：附近 vector 7 的 local `0841h` routine 並非這次 selector 的目標；即使
它讀取另一個 offset，也不能用相鄰 vector 的相似形狀替代 vector 6 的資料流。

## resident dseg 的負面結果

START database 的 `dseg` linear 範圍為 `0x1C1C0..0x250C0`。以
`EA = dseg_start + DS-relative offset` 對照：

| DS-relative | IDA linear EA | direct data xref |
|---:|---:|---:|
| `7206h` | `0x233C6` | `0` |
| `720Fh..7213h` | `0x233CF..0x233D3` | 各 `0` |
| `8B5Eh` | `0x24D1E` | `0` |

這只表示 resident `START.EXE.i64` 的 direct data-xref 圖沒有把 extracted
overlay code 建成同一張 graph；它不是「欄位未使用」或「沒有 writer」的證據。
後續要追 address-taken、overlay loader relocation、間接 pointer 與 DOSBox
runtime state，不能以這個零值刪除候選資料流。

## 對 worklist 與 remake 的影響

- 11 個直接影響正常玩家結果的逆向主題，以及 4 個 fidelity／發行主題，數量
  不變。這輪只關閉 P0-1 的 static dispatch sub-boundary，沒有關閉 P0-1
  的玩家可觀察地圖 handoff。
- 不修改 CoAB JSON、engine movement、`set_map_position` 或 secret-door
  contract；目前既有 `(1,8)`→`(13,10)` adapter 仍是 `strong inference`，
  `(13,10)`→`(8,15)` 仍是未完成的正常路徑。
- 不把 `overlay-30 [DI+0200h]`、`DS:720F／7210` 或任何 `4BF0h` 數字寫成
  座標、牆面、年齡或劇情旗標。所有語意仍要附原始位址空間與推論等級。

下一步按優先級追：

1. `DS:7206h` far pointer 的初始化與 overlay relocation／owner。
2. `DS:720F／7210／7213` 的 writer／consumer，以及 vector 4 的
   `DS:7211／7212` 分支，保持兩條資料流分開。
3. `C04B..C04F` 與 `CALL 2E10h` 的 runtime projection，並以 DOSBox 原版
   可重現狀態對照 map、座標、方向與畫面重繪。
4. 只有 writer、movement consumer、runtime state 與正常玩家輸入閉合後，才
   重新評估秘密門／搜尋 JSON；不能以本輪 accessor 自動新增規則。

## 可重現工具

```text
Docker image: ida-pro-9.4-ver2:uidfix-v1
IDA input: disposable copy of workplace/ida406/overlays/overlay-30.bin
script: scripts/ida/dos_overlay30_vector_audit.idc
report: /tmp/dos-overlay30-vector-audit.txt

IDA input: disposable copy of workplace/ida406/START.EXE.i64
script: scripts/ida/dos_start_dseg_offset_audit.idc
report: /tmp/dos-start-dseg-offsets.txt
```

兩個腳本都只在 disposable database 上解碼／匯出，不改名、不 patch、不回寫
原始 binary 或 baseline `.i64`。
