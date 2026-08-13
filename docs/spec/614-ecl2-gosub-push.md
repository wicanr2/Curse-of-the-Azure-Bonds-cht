# 第六百一十四輪：`SETUPGOSUBSTACK`（GOSUB 的推入端）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-07:1A66h`（`ECL2` stub `0098h`，134 bytes）。

```text
SETUPGOSUBSTACK(n):
    if DS:A882h = nil then
        New(DS:A882h, 6)
        節點^[2] := nil
    else
        old := DS:A882h
        New(DS:A882h, 6)                      ← 新節點成為鏈頭
        節點^[2] := old                        ← 插在鏈頭
    節點^[0] := ECL_PC                         ← 存返回位址
    ECL_PC := ADDFNC(DS:[A957h + n], DS:[A997h + n])   ← 跳到 operand n 指定的位址
```

## 與 `RETURN` 完全配對

| | 推入（本輪） | 彈出（`13h`，[spec 588](588-ecl-return-and-gosub-stack.md)） |
|---|---|---|
| 位置 | **鏈頭** | **鏈頭** |
| 節點 | `New(6)` | `Dispose(6)` |
| `+0` | 寫入當前 `ECL_PC` | 讀出並寫回 `ECL_PC` |
| `+2` | 指向舊鏈頭 | 成為新鏈頭 |

**插在鏈頭 ＝ LIFO**，是正確的堆疊行為。這與 `ADDEFFECT` 接在鏈尾
（[spec 578](578-effect-node-list.md)）、`0Bh` 複製時插鏈頭
（[spec 604](604-ecl-spawn-monsters.md)）並列——同一個引擎裡三種鏈的插入策略
各自不同，各有各的道理。

## 兩件事

1. **存的是呼叫當下的 `ECL_PC`**，不是 `PC + 指令長度`。`02h`（GOSUB）的
   handler 先 `READVAR(1)` 把 PC 推過 operand，才呼叫這一支——所以存進去的
   已經是「下一條指令」的位址。
2. `operand n` 用的是 `n` 而不是固定 1：`02h` 傳 `1`，但這支接受任意索引，
   代表別的指令也能借它做「呼叫」。

## 明確不宣稱

- 除了 `02h` 之外還有誰呼叫它。
- 堆疊沒有深度上限——`New` 失敗時的行為未讀。
