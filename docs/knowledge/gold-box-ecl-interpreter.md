# Gold Box ECL 直譯器內部構造（原版反組譯）

這份文件記錄 **原版執行檔裡 ECL 直譯器實際怎麼運作**：指令分派、operand 編碼、
記憶體 bank 與 helper API。它與
[`gold-box-ecl-command-set.md`](gold-box-ecl-command-set.md) 的分工是——那份講
「有哪些指令、各吃幾個 operand」（由 parser 與公開 reference 累積），這份講
「直譯器本體的機制」（由 IDA 逐指令讀出）。

所有位址都是 **overlay-local code offset**，來自
[`docs/spec/559`](../spec/559-full-module-re-sweep.md) 的全模組全掃。等級標記
沿用專案的 `exact／strong inference／hypothesis／unknown`。

證據規格：[560](../spec/560-ecl-opcode-dispatch-table.md)、
[561](../spec/561-ecl-external-call-registry.md)、
[562](../spec/562-ecl2-helper-api-and-operand-audit.md)、
[563](../spec/563-ecl-memory-model-and-operand-resolution.md)、
[564](../spec/564-ecl-operand-decoding-and-arity-validation.md)、
[565](../spec/565-ecl-memory-read-path-and-asymmetry.md)。

## 模組分工

| 單元 | overlay | 職責 |
|---|---|---|
| `INTERPET` | 02 | opcode dispatcher 與 52 個 handler |
| `ECL2` | 07 | 共用 helper：operand 解碼／解析、記憶體讀寫、字串、選單 |

`INTERPET` 的 handler 幾乎不直接碰記憶體，都是 far call 進 `ECL2`。跨 unit 的
far call 目標是 **`ECL2` control block 的 stub offset**，不是 code offset；
兩平台的 stub offset 與 entry index 相同，只有 code offset 不同。

## 指令分派（`exact`）

| 平台 | dispatcher | opcode 來源 |
|---|---|---|
| DOS | `overlay-02:3377h` | `DS:75FFh`（`cmp ax`，先 `xor ah, ah`） |
| PC-98 | `overlay-02:373Eh` | `DS:0A891h`（`cmp al`） |

線性 `cmp／jz／call／jmp` 鏈，不是跳表。`00h..40h` 中除 `1Fh` 外的 **64 個
opcode 各有 handler，共 52 個**（部分 opcode 共用）。`41h` 以上走到 epilogue。

完整表：[`../audit/ecl-opcode-dispatch.md`](../audit/ecl-opcode-dispatch.md)。

## operand 編碼與解碼（`exact`）

### 三個平行陣列

operand 不是 handler 自己解，而是由 `READVAR` 事先拆進三個各 64 byte 的陣列：

| 陣列 | PC-98 | DOS |
|---|---|---|
| operand code | `DS:A917h` | `DS:7685h` |
| 高位元組 | `DS:A957h` | `DS:76C5h` |
| 低位元組 | `DS:A997h` | `DS:7705h` |

間距固定 `40h`，索引 **從 1 開始**，用 `cbw` 符號延伸取址。dispatcher 讀
opcode 的位址與 code 陣列起點固定差 `86h`（兩平台一致）。

### bytecode 佈局

```text
[opcode][code₁][low₁]([high₁])[code₂][low₂]([high₂])…
```

`code` 為 `1`／`2`／`3` 時多一個 `high` byte，其餘只有一個 value byte。
`code = 80h`（packed text）走另一條變長路徑（**尚未讀完**）。

### `READVAR(n)`：解碼器

```text
for i = 1..n:
    code[i] = script[PC+1]; low[i] = script[PC+2]; PC += 2
    if code[i] in {1,2,3}: PC += 1; high[i] = script[PC]
```

ECL PC 是 `DS:7F21h`（PC-98，`exact`）。DOS 側對應為 `DS:4FB4h`
（`strong inference`）：`overlay-02` 的 `3Ah` handler 在兩平台同形，PC-98 版
`inc DS:7F21h`、DOS 版 `inc DS:4FB4h`。script 緩衝區是下方的 bank 3。

**`READVAR` 的參數就是該指令的 arity。** 用它驗證 `internal/ecl` 的
`KnownCommands`：64 個 opcode 中 62 個吻合（PC-98），其餘 4 個是變長指令。

### `ADDRESSVALUE(i)`：取用第 i 個 operand

| operand code | 行為 |
|---|---|
| `00h` | 回傳低位元組（零延伸）＝ byte literal |
| `01h`／`03h`／`80h` | 兩 byte 併成 word，再走記憶體讀取 |
| `02h`／`81h` | 兩 byte 併成 word 直接回傳＝ word literal |
| 其他 | 不設定回傳值 |

`ADDFNC(a, b) = (b << 8) + a`，只是 byte pair 併 word，不是算術加法。

⚠ **`ADDRESSVALUE` 的呼叫次數與 arity 無關**：一個已解好的 operand 可以被
取用零次或多次。要驗 arity 一律看 `READVAR` 的參數。

### 變長指令

```text
READVAR(固定前綴) → ADDRESSVALUE(某 operand) 取得 N → dec PC → READVAR(N)
```

| opcode | 指令 | 固定前綴 |
|---|---|---:|
| `15h` | VERTICAL MENU | 3 |
| `25h`／`26h` | ON GOTO／ON GOSUB | 2 |
| `2Bh` | HORIZONTAL MENU | 2 |

## 記憶體模型（`exact`）

讀寫共用同一個位址分類器（PC-98 `overlay-07:0801h`／DOS `0735h`）：

| bank | PC-98 範圍 | DOS 範圍 | 寬度 | 位移 | far pointer |
|---:|---|---|---|---|---|
| 0 | `4B00h..4EFFh` | 同左 | word | `2×addr + 6A00h` | `DS:7F05h` |
| 1 | `7C00h..7FFFh` | 同左 | word | `2×addr + 0800h` | `DS:7F09h` |
| 2 | `7A00h..7BFFh` | 同左 | word | `2×addr + 0C00h` | `DS:7F0Dh` |
| 3 | `8000h..9E40h` | `8000h..9DFFh` | **byte** | `addr − 8000h` | `DS:7F12h` |
| 4 | 其餘 | 其餘 | — | 具名特例 | — |

- 每個 bank 的位移常數都讓「bank 起始位址」剛好落在陣列第 0 項（16 bit 溢位），
  這同時證明了 bank 起點。
- **bank 3 是載入中的 ECL script 緩衝區**，起點 `8000h` 與事件清冊的
  `code_address = 8000h + block offset` 吻合，且是唯一以 byte 存取的 bank。
  `INPUT STRING` 把 packed text 寫回這裡＝ script 自我修改。
- **bank 3 上界兩平台不同**（DOS `9DFFh`／PC-98 `9E40h`）。跨平台工作不得
  假設相同。
- **bank 1 讀取時先算再回退**：呼叫計算 routine（PC-98 `08BDh`）並帶有效性
  旗標，旗標為 0 才讀陣列。所以 `7C00h..7FFFh` 有一部分是虛擬值。
  哪些位址走計算路徑 **尚未解出**。

### bank 4：讀寫不對稱

| ECL 位址 | 讀 | 寫 |
|---|---|---|
| `00B1h`／`00FBh`／`00FCh` | 有值（`DS:A880h`／`A87Ch`／`A87Eh`） | **明確忽略** |
| `033Dh` | `DS:A2ABh` | — |
| `03DEh`／`00B8h`／`00B9h`／`5208h` | — | 有目的位址 |

把它們當成一般可讀寫的 work address，會在其中一個方向產生原版沒有的行為。

### 虛擬地圖暫存器 `C04Bh..C05Fh`

| ECL 位址 | 讀 | 寫 | PC-98 目的 | DOS 目的 |
|---|---|---|---|---|
| `C04Bh` | 符號延伸 | ✓ | `DS:A2A9h` | `DS:720Fh` |
| `C04Ch` | 符號延伸 | ✓ | `DS:A2AAh` | `DS:7210h` |
| `C04Dh` | **÷2** | **×2** | `DS:A2ABh` | `DS:7211h` |
| `C04Eh`／`C04Fh` | 零延伸 | ✗ | `DS:A2ACh`／`A2ADh` | `DS:7212h`／`7213h` |
| `C059h` | no-op | ✓ | `DS:A87Ah` | — |

`C04Dh` 的 `×2`／`÷2` 互為反向，**對 ECL 而言 round-trip 一致**；係數只有直接
讀該 DS 位址的繪圖端看得到。`C04Bh`／`C04Ch` 讀取是 `cbw` 符號延伸。

## external `CALL`（opcode `2Dh`，`exact`）

operand 先減 `7FFFh` 得到 selector：`selector = (operand − 7FFFh) mod 10000h`。
engine 認得 7 個，兩平台相同：

`8000h`、`8001h`、`B200h`、`C018h`、`C01Eh`、`2E10h`、`6803h`

其餘 operand 走 epilogue，不做事直接返回。CoAB 的 ECL1–ECL6 靜態只用到
`2E10h`（12 次）與 `6803h`（11 次）。
表：[`../audit/ecl-external-call-registry.md`](../audit/ecl-external-call-registry.md)。

## `ECL2` 具名 routine（PC-98 Borland 符號，`exact`）

`INITECL`、`GETECL`、`GETMONSTERS`、`ADDRESSVALUE`、`READVAR`、`STOREVALUE`、
`ADDFNC`、`GETSTR`、`FINDSTR`、`STORESTRING`、`CHECKSTRING`、`CHECKSTATUS`、
`SETUPGOSUBSTACK`、`MOVEFORWARD`、`ECLMENUH`、`ECLMENUV`、`GODUEL`、
`ROBDOUGH`、`ROBSTUFF`、`NONEXT`、`CHARSPEED`、`KILLTHEDUDE`。

DOS 沒有符號表，名稱依 **entry index** 對應過去（`strong inference`）——
不能用 code offset 對，兩平台的 code offset 不同。

## 相鄰的規則核心：`EFFECTS`（overlay-23）

不屬 ECL 直譯器，但 ECL 的傷害／法術指令最終都會落到這裡。Borland 符號給出
單元骨架：`TRYTOHIT`、`ATTEMPTTOHIT`、`MAKESAVE`、`ROLLDICE`、
`ROLLDAMAGEDICE`、`ADDEFFECT`、`REMOVEFX`、`CUREEFFECT`、`PUTDAMAGE`、
`RECALCULATESTATS`、`CONVERTSTRTOSPEC`／`CONVERTSPECTOSTR` 等 21 個。

### 已讀完的部分

| 主題 | 規格 |
|---|---|
| 骰子 | [568](../spec/568-rolldice-and-original-rng-entry.md) |
| 18/xx 力量的 byte 編碼、effect 移除 | [576](../spec/576-adnd-strength-encoding-and-effect-removal.md) |
| 命中判定完整式、effect 鏈遍歷 | [577](../spec/577-attempttohit-and-effect-chain-walk.md) |
| effect 節點鏈（9 bytes／掛在角色 `+0F2h`） | [578](../spec/578-effect-node-list.md) |
| 角色狀態欄位 | [579](../spec/579-character-status-fields.md) |
| `PUTDAMAGE` 傷害管線與傷害屬性位元 | [581](../spec/581-putdamage-pipeline.md) |

角色記錄目前確定的欄位：`+78h` 最大 HP、`+0F2h` effect 鏈頭、`+196h` 狀態碼、
`+197h`／`+198h` 旗標、`+19Ah` 命中修正、`+1A5h` 目前 HP、`+3` 力量編碼。

傷害屬性 `DS:A02Fh`：bit 0 Fire、bit 1 Cold、bit 2 Electricity、bit 3 Magic、
bit 4 Acid。

### 骰子

```text
ROLLDICE(count, sides) = Σ_{i=1..count} (Random(sides) + 1)   ← 回傳 byte
```

`Random` 是 Turbo Pascal RTL 的 `Random(n)`，已由位址換算證明
（DOS far call `0A54:1105` ＝ `@Random$q4Word`）。**沒有重骰、沒有上下限修正、
沒有爆擊特例**。

亂數本體已完整解出（[spec 575](../spec/575-random-core-and-pc98-vram.md)，`exact`）：

```text
RandSeed := (RandSeed * 134775813 + 1) mod 2^32
Random(n) := if n = 0 then 0 else (RandSeed shr 16) mod n
```

**只取高 16 位**，且對非 2 冪的 `n` 有取模偏差——這是原版行為。remake 照抄
這兩行就能重現原版的整條骰子序列。

## 跨作品沿用

指令分派、operand 三陣列、`READVAR`／`ADDRESSVALUE`／`STOREVALUE` 的分工與
bank 分類的**形狀**很可能是 Gold Box 共用的；但**每一個常數都必須各作品重解**：
bank 邊界、helper 的 stub offset、selector 集合、DS 目的位址在 CoAB 的兩個
平台之間就已經不同。沿用結構、重解數值。

## 尚未解出

- `code = 80h`（packed text）的長度規則。
- bank 1 的計算 routine 算什麼、哪些位址走它。
- bank 0–3 背後緩衝區的配置者（`DS:7F05h/7F09h/7F0Dh/7F12h` 只是存 far pointer）。
- 52 個 handler 的個別語意（目前全部 `待解讀`，見
  [`../audit/coab-function-index.md`](../audit/coab-function-index.md)）。
- 7 個 external routine 分支主體的效果。
