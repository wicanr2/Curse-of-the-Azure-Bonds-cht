# 第七百一十輪：把物品的效果收進清單，與殘留在堆疊上的目標

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-22` 的 `56C4h`、`08A1h`。

## `56C4h`：對整個目標陣列施加傷害

```text
a := <overlay-32 entry#15>(DS:6506h)
b := <overlay-32 entry#16>(DS:6506h)
<sub_175Bh>(a, b, DS:7559h, DS:755Ah, 1, 3)

for k := 1 to DS:7434h do
    p := far [7431h + k×4]                       ← spec 709 的目標陣列
    if p = NIL then 跳過                          ← 兩個 word 同時為零才算 NIL
    flag := (p^[11Ah] = 12h) ? 0 : 1
    d := ROLLDAMAGEDICE(6, 6)                     ← 6d6
    <overlay-23 entry#20>(p, d, byte[3E02h], flag)
```

`NIL` 檢查是必要的——`1C39h`（spec 709）會把不符合的格子寫成 `NIL` 而不動筆數。
這支正好示範了下游該怎麼配合。

### `entry#20` 的第一個參數是殘留

推入 `ROLLDAMAGEDICE` 之前先推了 `p` 的 far 指標，而 `ROLLDAMAGEDICE` 只吃兩個
word（`6`、`6`），所以 `p` 留在堆疊上，成為 `entry#20` 的第一個參數。
`45B5h`（spec 708）的 `entry#20` 也是同樣的形狀。

這和 spec 690／701 的字串殘留是同一個機制，但殘留的是**呼叫端自己推的參數**而
不是某個函式的回傳值——換句話說，殘留不一定來自 `0A54:0634h` 那條鏈。
**看到「推入的比被呼叫者要的少」時，要往上找的不只是前一個 `call`，還有更早
被推入而沒人消耗的東西。**

## `08A1h`：把物品身上還有效的效果收成一張表

```text
if <overlay-24 entry#27>(DS:6506h, 10h, @node) then
    清除 := true
elif DS:6506h^[109h] > 0 then
    清除 := 看 [5CF6h + DS:729Ch^[2Eh] × 16] 是否為 0Ch
elif DS:6506h^[111h] > DS:6506h^[0E6h] then
    清除 := 看同一個表格值是否為 0Ch
else
    清除 := false

if 清除 then DS:729Ch^[35h] := 0
if DS:729Ch^[35h] <> 0 then 返回                  ← 不為 0 就整支不做事

for k := 1 to 3 do
    slot := DS:729Ch^[3Bh + k]                    ← spec 703 的三個效果槽
    收 := (arg_0 <> 0 且 slot > 80h) 或 (arg_0 = 0 且 slot > 0)
    if 收 then
        <sub_3F5h>(slot)
        far [4BF0h + byte[4CB4h] × 4] := DS:729Ch
        byte[4CB4h] := byte[4CB4h] + 1
retf 2
```

`arg_0` 切換兩種篩選：

| `arg_0` | 收的條件 | 對照 spec 703 |
|---|---|---|
| 非 0 | `slot > 80h`（無號） | 最高位為 1 **且** 低 7 位非 0 |
| 0 | `slot > 0` | 只要有東西就收 |

`> 80h` 用的是 `ja`，所以 `80h` 本身**不算**——也就是「最高位為 1 但編號是 0」
會被排除。改寫成 `(slot and 80h) <> 0` 會多收這一種。

`DS:4BF0h` 是收集結果的 far 指標陣列（一格 4 bytes），`DS:4CB4h` 是**目前筆數**
（byte）。這支只加不清，**沒有上界檢查**——`4CB4h` 相鄰的 `4CB5h` 正是 spec 705
那個擲骰門檻，所以溢出會直接寫壞它。

`DS:729Ch` 指到的是一個物品節點：`+2Eh`（類別）、`+3Ch`..`+3Eh`（三個效果槽）
都和 spec 696／703 對得上，`+35h` 是建構子第 8 個參數的落點。

## 明確不宣稱

- `+35h`、`+109h`、`+111h`、`+0E6h`、`+11Ah` 的語意。
- 表格值 `0Ch`、編號 `10h`、`12h` 代表什麼。
- `sub_175Bh`、`sub_3F5h`、`overlay-23 entry#20`、`overlay-32 entry#15/#16` 的行為。
- `byte[3E02h]` 是什麼。
