# 第五百六十二輪：ECL2 helper API 與 handler operand 交叉稽核

狀態：`READY`（限 helper 名稱解析鏈與計數稽核）；helper 的語意與每個
mismatch 的成因仍是 `待解讀`。日期：2026-08-14

## 結論先行

`INTERPET` 的 opcode handler 不自己解 operand，而是 far call 到 **ECL2 單元
（overlay-07）** 的共用 routine。那些 far call 的目標是 ECL2 **control block 的
stub offset**，可以一路解回原始函式名：

```text
INTERPET handler 的 far call 0062:0025（PC-98）／006B:0025（DOS）
  → overlay-07 control stub offset 0025h → entry index 1
  → code offset 0296h（PC-98）／0173h（DOS）
  → PC-98 Borland 符號 ADDRESSVALUE
```

四個最常被呼叫的 helper（PC-98 overlay-02 全部 540 條 call 指令中）：

| stub | entry | PC-98 code | DOS code | 原始名稱 | PC-98 次數 | DOS 次數 |
|---|---:|---|---|---|---:|---:|
| `0025h` | 1 | `0296h` | `0173h` | `ADDRESSVALUE` | 52 | 55 |
| `002Ah` | 2 | `008Eh` | `0034h` | `READVAR` | 41 | 44 |
| `006Bh` | 15 | `0E2Bh` | `0D70h` | `STOREVALUE` | 32 | 33 |
| `004Dh` | 9 | `07DCh` | `06E8h` | `ADDFNC` | 23 | 24 |

**兩平台的 stub offset 與 entry index 完全相同**，只有 code offset 不同。
這使 PC-98 的符號名可以套用到 DOS 的同一個 entry index，等級 `strong inference`
（同一份原始碼、同一個 export 次序；仍缺逐函式的行為對照）。

overlay-07（ECL2）另有這些具名 routine 可直接使用：`INITECL`、`GETECL`、
`GETMONSTERS`、`GETSTR`、`FINDSTR`、`STORESTRING`、`CHECKSTRING`、
`CHECKSTATUS`、`SETUPGOSUBSTACK`、`MOVEFORWARD`、`ECLMENUH`、`ECLMENUV`、
`GODUEL`、`ROBDOUGH`、`ROBSTUFF`、`NONEXT`、`CHARSPEED`、`KILLTHEDUDE`。

## operand arity 交叉稽核

既有的 `internal/ecl/operand.go` `KnownCommands` 是從公開 reference 與 parser
工作得到的，從未與原版 handler 對照過。本輪第一次做這件事：數每個 handler
呼叫 `ADDRESSVALUE` 幾次，與宣稱的 arity 比對。

⚠ **本節使用的訊號已被第 564 輪取代。** 當時數的是 `ADDRESSVALUE` 的呼叫
次數（64 個中 35 個一致），但那個訊號本身是錯的：`ADDRESSVALUE(i)` 是取用
第 i 個**已解好**的 operand，同一個 operand 可被取用零次或多次。正確訊號是
`READVAR(n)` 的參數，換過去之後是 62／64 吻合。結論與現行表格見
[`spec 564`](564-ecl-operand-decoding-and-arity-validation.md) 與
[`docs/audit/ecl-handler-operand-audit.md`](../audit/ecl-handler-operand-audit.md)。
本節保留的仍然有效的部分是下方的 helper 名稱解析鏈。

不一致**不等於 remake 的 arity 錯**。已可看出至少兩種合法的差異來源：

- 算術一族（`04h..07h` ADD／SUBTRACT／DIVIDE／MULTIPLY，共用 handler `01B7h`）
  是 `ADDRESSVALUE`×1 ＋ `ADDFNC`×1 ＋ `STOREVALUE`×1，不是三次
  `ADDRESSVALUE`。取值與存值走不同 helper。
- `01h GOTO`／`02h GOSUB` 完全不呼叫 `ADDRESSVALUE`，改用 `READVAR`。
  跳躍目標是 bytecode 裡的裸 word，不需要 operand descriptor 解析。

因此這張表的用途是**排序待查清單**，不是判定對錯。每一筆都要逐一讀 handler
才能定案；定案後才更新 `KnownCommands` 或在此標記為已驗證。

## 這份規格明確不宣稱

- **helper 的語意**。`ADDRESSVALUE`／`READVAR`／`STOREVALUE`／`ADDFNC` 是原始
  名稱（`exact`），但名稱不是語意證明。上面對「取值 vs 存值 vs 裸 word」的
  描述是 `hypothesis`，要讀完 helper 本體才能升級。
- **任何 mismatch 的判定**。29 筆全部是待查，沒有一筆被判定為 remake 的 bug。
- **DOS 的 helper 名稱**。DOS 沒有符號表；名稱是依 entry index 對應過去的
  `strong inference`。
- **`KnownCommands` 的正確性**。本輪既沒有推翻也沒有確認它；只是第一次讓它
  對上原版證據，並把差異列出來。
