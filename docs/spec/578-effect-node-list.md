# 第五百七十八輪：effect 節點鏈的資料結構

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`EFFECTS`（overlay-23）。兩平台已配對。

角色身上的每一個持續效果（法術、狀態、地形影響）是**堆積上的一個 9-byte
節點**，串成單向鏈掛在角色記錄的 `+0F2h`。

## 節點格式（9 bytes）

| 偏移 | 寬度 | 內容 |
|---:|---|---|
| `+0` | byte | effect id |
| `+1` | word | 參數（`ADDEFFECT` 的 `arg_4` 原樣寫入） |
| `+3` | byte | `ADDEFFECT` 的 `arg_2` |
| `+4` | byte | `ADDEFFECT` 的 `arg_0`；**非 0 表示解除時要回呼** |
| `+5` | far ptr | next（4 bytes），鏈尾為 nil |

角色記錄 `+0F2h`（4 bytes）是鏈頭 far pointer。

節點大小是 `New`／`Dispose` 的參數直接給的（`81Fh:0043h` 與 `81Fh:01A2h`
都帶 `9`），不是從欄位反推的。

## `ADDEFFECT`（PC-98 `13D7h`／DOS `13EEh`）

```text
ADDEFFECT(f4, f3, param, id, char)
    New(node, 9)
    if char^[0F2h] = nil then char^[0F2h] := node
    else                      <走到鏈尾>^[5] := node      ← 接在尾端
    node^[5] := nil
    node^[0] := id ; node^[1] := param ; node^[4] := f4 ; node^[3] := f3
```

**新效果掛在鏈尾**，不是鏈頭——所以遍歷順序＝加入順序。

## `SPELLOFF`（PC-98 `010Eh`／DOS `010Eh`）

```text
SPELLOFF(node_or_nil, id, char)
    p := node_or_nil
    if p = nil then <從 char^[0F2h] 起，找第一個 p^[0] = id 的節點>
    if p = nil then goto tail                  ← 找不到就只做 tail
    if p^[4] <> 0 then CALLEFFECT(1, p, char, id)   ← 通知該 effect 要解除
    <把 p 從鏈上摘掉>（鏈頭與中間兩種情形分開處理）
    Dispose(p, 9)
tail:
    if id  = 0Eh                then sub_18F3(5, char)
    if id in {0Ch, 26h, 92h}    then sub_18F3(0, char)
```

三個要點：

1. **傳 nil 就用 id 搜第一個符合的**，傳節點就直接摘該節點。`REMOVEFX`／
   `ROUTINGREMOVEFX` 走的是前者。
2. `p^[4] <> 0` 才回呼 `CALLEFFECT(..., 1, ...)`——第一個參數 `1` 與
   `sub_269` 呼叫時的 `0` 相對，形狀是「開／關」兩種通知。
3. **善後在鏈外**：`0Eh` 與 `{0Ch, 26h, 92h}` 兩組 effect 解除後要另外呼叫
   `sub_18F3`，而且**即使該效果根本不在鏈上也照做**（`goto tail` 跳過摘鏈
   但不跳過這段）。remake 把善後寫進「摘鏈成功」分支就會漏。

## 明確不宣稱

- `sub_18F3` 的本體與 `5`／`0` 兩個參數的意義。
- effect id `0Ch`／`0Eh`／`26h`／`92h` 分別是什麼。
- 節點 `+3`／`+4` 除了「非 0 要回呼」以外的語意。
