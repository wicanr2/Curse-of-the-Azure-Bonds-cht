# 第五百九十一輪：`SKIP` 的 arity 表，與它跟 handler 的兩處不一致

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-07:1FB0h`／DOS `overlay-07:1FF5h`（`ECL2` entry#29，
stub `00B1h`）。

## `16h`～`1Bh` 的語意確定了

`0062:00B1` 是**跳過下一條 ECL 指令**：

```text
SKIP:
    op := script[ECL_PC]                  ← 從 bank 3 讀下一條的 opcode
    DS:A891h := op
    n := arity(op)                        ← 寫死的對照表
    if n > 0 then READVAR(n)              ← 讀掉它的 operand，PC 隨之前進
    else          ECL_PC := ECL_PC + 1
```

所以 `16h`～`1Bh`（[spec 590](590-ecl-control-flow-if-and-branch.md)）是
**「旗標為 0 就跳過下一條指令」**。配上 `14h`（相等設 `A88Ah`、不等設
`A88Bh`），ECL 的條件式就是：

| 組合 | 效果 |
|---|---|
| `14h` ＋ `16h` | 相等時執行下一條，不等時跳過 ⇒ `IF EQUAL` |
| `14h` ＋ `17h` | 不等時執行下一條 ⇒ `IF NOT EQUAL` |

## 這張表是 arity 的第二個獨立來源

先前的 arity 表來自「每個 handler 自己呼叫 `READVAR(n)` 的 n」；`SKIP` 這張
來自「別人要跳過它時認為它有多長」。兩者**互相獨立**，對得上才可信。

`scripts/skip_arity_crosscheck.py` 的比對結果：

| | 數量 |
|---|---:|
| 兩邊都有且一致 | **42** |
| 不一致 | **2** |
| 只在 `SKIP` 表 | 1 |
| 只在 arity 表 | 20（全部是 arity 0，走 `SKIP` 的 `inc PC` 分支） |

那 20 個「只在 arity 表」的 arity 全是 0，與 `SKIP` 把它們丟給 `inc PC` 完全
自洽——這本身就是一道通過的檢查。

## `1Fh`：有 arity、沒有 handler

`SKIP` 認為 `1Fh` 有 **2 個 operand**，但 dispatcher **沒有 `1Fh` 的 handler**
（[spec 560](560-ecl-opcode-dispatch-table.md)：`00h..40h` 中只有 `1Fh` 沒有）。

所以 `1Fh` 是一條「跳過時要正確計算長度、執行時什麼都不做」的指令。remake
若把它當成非法 opcode 而報錯或停止，會與原版不同。

## `34h` 與 `36h`：原版自己不一致

| opcode | 指令名 | handler 的 `READVAR` | `SKIP` 認為 |
|---|---|---:|---:|
| `34h` | ECL CLOCK | **2** | **1** |
| `36h` | ADD NPC | **2** | **1** |

`34h` 的 handler（`2E2Ch`）第一條就是 `READVAR(2)`
（[spec 586](586-ecl-handlers-31-33-34.md)），`36h` 的 audit 解析也是 2。
而 `SKIP` 表把這兩個歸進 arity 1 的那一組（`1FEFh`／`1FF3h` 的 `jz` 都指向
`200Bh`，也就是 `mov al, 1`）。

**後果**：如果 script 裡出現「`16h`～`1Bh` 緊接著 `34h` 或 `36h`」而且條件不
成立，`SKIP` 會少讀一個 operand，ECL PC 就此錯位，後面整段指令都會被讀成別的
東西。

這是**原版的行為**，不是本專案的判讀錯誤——兩個獨立來源各自明確。remake 要
重現原版，就必須照抄 `SKIP` 的表（1），而不是照抄 handler 的 `READVAR`（2）。
「修正」它反而會產生原版沒有的行為。

**尚未驗證**：CoAB 的 ECL1–ECL6 裡到底有沒有出現這個組合。沒觸發過的話這只是
一顆沒爆的地雷，但 remake 仍應照抄——不能假設所有 script 都不會踩到。

## 明確不宣稱

- `1Fh` 這條指令原本打算做什麼。
- `34h`／`36h` 的不一致是筆誤還是刻意。
