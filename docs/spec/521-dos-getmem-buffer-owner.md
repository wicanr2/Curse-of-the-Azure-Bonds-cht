# 第五百二十一輪：DOS `0A54:0329h` 的 `GetMem` buffer owner

狀態：`READY`（僅限 Borland runtime／buffer ownership 邊界；不是 map plane 或 handoff 完成）
日期：2026-08-09

## 結論先行

第 520 輪當時仍把 overlay-11 的 `0A54:0329h` 記成未知外部 service；本輪在
resident `START.EXE.i64` 的 segment selector、IDA symbol、入口 bytes 與 direct
caller 交叉核對後，得到：

```text
overlay-11 local 00E9h
  → push DS:7206h
  → push 0400h
  → runtime far call 0A54:0329h
  → IDA +1000h paragraph bias
  → seg050 selector 1A54h, IDA EA 1A869h
  → @GetMem$qm7Pointer4Word; GetMem(Pointer &,Word)
```

因此可以把 `DS:7206h` 從「完全未知的 far pointer」降為：**一個由 Borland
`GetMem` 接收、要求 `0400h` bytes 的 pointer variable／buffer owner candidate**。
這仍不證明該 buffer 是哪張 GEO、哪個 plane、是否保存座標或如何投影到
`C04B..C04F`；不修改 engine／JSON，也不新增秘密門規則。

## Provenance 與位址空間

| 輸入／工具 | SHA-256／版本 | 位址基準 |
|---|---|---|
| DOS `START.EXE` | `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` | MZ raw file |
| baseline `START.EXE.i64` | `9df802ee4ef71fb2eda83257e0ed2d87adf0ee2d10241d3bdbdc6bc369fe47eb` | IDA database；disposable copy |
| extracted overlay-11 | `a20b0e529b68eac4ab4638491a1e4e7eb55fef3ad988b4879502954a1fdfc6b7` | overlay-local offset |
| IDA Pro | Docker image `ida-pro-9.4-ver2:uidfix-v1`；`9.4.0.260610` | 16-bit segment decode |
| segment inventory script | `64b245ea719704c5e1136a55de924bb1fac857a07b961955f00308f1c04aa69f` | [`dos_start_segment_inventory.idc`](../../scripts/ida/dos_start_segment_inventory.idc) |
| segment inventory report | `ff97710ab70edf4b304f2ab8334a48b7af9d0090310220c7d2a2254ed41b1fab` | `/tmp/dos-start-segments.txt` |
| `0A54:0329h` audit script | `adc934c2bdbd101d6c031730ba6fddef7951fa739ea27f0bb6e54de0c86e16dd` | [`dos_start_0a54_call_audit.idc`](../../scripts/ida/dos_start_0a54_call_audit.idc) |
| `0A54:0329h` audit report | `a9f7829d8b108b5a36b2eb79970e92bff5ec420c1ad2273dc7714fffc79cfb17` | `/tmp/dos-start-0a54-call-audit.txt` |

`START.EXE.i64` 的 segment inventory 顯示：

```text
seg000 start=0x10000 selector=0x1000
seg050 start=0x1A540 selector=0x1A54
dseg   start=0x1C1C0 selector=0x1C1C
```

本規格採用的位址對照是：resident runtime selector `0A54h` 在此 IDA database
以 `+1000h` paragraph bias 表示為 selector `1A54h`；因此：

```text
runtime 0A54:0329h
→ IDA selector 1A54h : offset 0329h
→ IDA linear EA 0x1A540 + 0x0329 = 0x1A869
```

這是**本 baseline IDA database 的位址映射**，不是把 overlay-local offset、
ECL work address 或 file offset 改寫成同一種數字。若日後換 executable／載入基址，
必須重新保存 bias 與 segment inventory，不能直接複製 `1A869h`。

## resident target 的入口與符號

IDA 在 `0x1A869` 的連續入口 bytes 為：

```text
1A869  55 8B EC                  push bp; mov bp,sp
1A86C  8B 46 06                  mov ax,[bp+6]
1A86F  E8 14 01                  call local 1A986h
1A872  73 1B                     jnb short 1A88Fh
...
1A88F  C4 7E 08                  les di,[bp+8]
1A892  26 89 0D                  mov es:[di],cx
1A895  26 89 5D 02               mov es:[di+2],bx
1A899  5D                       pop bp
1A89A  CA 06 00                  retf 6
```

IDA 原始函式符號／prototype 顯示為：

```text
@GetMem$qm7Pointer4Word; GetMem(Pointer &,Word)
```

在同一 IDA database，target `0x1A869` 有 4 個 direct far-call xref：

```text
from=0x14600  call far ptr @GetMem$qm7Pointer4Word; GetMem(Pointer &,Word)
from=0x14667  call far ptr @GetMem$qm7Pointer4Word; GetMem(Pointer &,Word)
from=0x1638F  call far ptr @GetMem$qm7Pointer4Word; GetMem(Pointer &,Word)
from=0x16DED  call far ptr @GetMem$qm7Pointer4Word; GetMem(Pointer &,Word)
```

這些是 resident IDA linear EA，不能直接當成 overlay-local address。入口的
`[bp+8]` far pointer write、`[bp+6]` word size 與 `retf 6`，和 Borland
`GetMem(Pointer &,Word)` 的呼叫形狀一致；因此「它是 GetMem」有 raw bytes、IDA
原始符號與 direct caller 三項交叉證據。

## overlay-11 call-site

overlay-11 local `00E9..00F5` 的 bytes／IDA 指令為：

```text
00E9  BF 06 72                  mov di,7206h
00EC  1E                        push ds
00ED  57                        push di
00EE  B8 00 04; 50              mov ax,0400h; push ax
00F2  9A 29 03 54 0A            call far 0A54:0329h
```

這個 Pascal stack shape 與 resident target 的 `Pointer &,Word` 一致，因此：

- `exact`：overlay-11 以 DS:7206 的 far variable 與 `0400h` word 呼叫
  runtime `0A54:0329h`；該 runtime target 在 baseline IDA address mapping 中是
  `GetMem`。
- `strong inference`：DS:7206 所指的資料會被要求配置為 `0400h` bytes，亦即
  是一個 1 KiB 的 buffer owner／pointer variable。這不等於已知 buffer 內容或
  layer layout；「map buffer」仍只可作候選描述。
- `unknown`：配置後 bytes 的填入者、`+000/+100/+200/+300` planes 的排列、
  overlay relocation／lifetime、與 ECL `C04B..C04F` 的 projection。

同一 overlay 的初始化 bytes 仍如 spec 520：

```text
03F2..03FC：DS:720F=07h、DS:7210=0Dh、DS:7211=00h
078E..0798：DS:720F=07h、DS:7210=0Dh、DS:7211=02h
```

這些初始化值不因此成為所有地圖固定起點；它們只證明本 module 在兩條初始化
路徑寫入該 state。

## 推論等級與下一步

### `exact`

- 在此 IDA baseline 的 `+1000h` selector mapping 下，runtime `0A54:0329h`
  對到 `seg050:1A54:0329h`／EA `0x1A869`。
- target 的 IDA 原始符號為 Borland `GetMem(Pointer &,Word)`，入口 bytes、
  far-pointer output、size operand 與 4 個 direct caller 可回查。
- overlay-11 的 `DS:7206h`／`0400h` call-site 與 stack 順序 exact。

### `strong inference`

- `DS:7206h` 是由 `GetMem` 配置／接收的 1 KiB buffer pointer；它很可能是
  overlay map-like data 的 owner，但尚不能命名 plane。

### `unknown`

- 配置後 buffer 的 writer、map file／GEO block 來源、plane offset 定義、
  `C04B..C04F` projection 與 runtime redraw。
- `0A54:0329h` 的所有失敗／記憶體不足分支在原版玩家路徑的實際呈現；本輪只
  解析 call／target，不將 runtime error handler 擴大成遊戲行為。

第 521 輪沒有減少 11 個行為逆向主題或 4 個 fidelity／發行主題，也沒有修改
engine、JSON、movement predicate 或 secret-door。下一個窄工作是追 `DS:7206`
配置後的 writer／plane fill，以及把 `DS:7212／7213` 與同一 buffer 的讀取端
對上；仍需以 DOSBox 原版狀態驗證，不能只靠 GetMem 名稱完成 map handoff。
