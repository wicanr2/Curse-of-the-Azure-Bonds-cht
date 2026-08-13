# 第七百零四輪：解除效果，與 `4289h` 少了哪一行

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-22` 的 `2FD1h`、`4311h`。

## `2FD1h`：查一個編號，解掉另外幾個

```text
result := 0
DS:6F9Ch := 1                                   ← 整段期間立起的旗標
if <overlay-23 entry#16>(目標, 22h) then result := 1
if <overlay-23 entry#16>(目標, 2Bh) then
    result := 1
    <overlay-23 entry#3>(目標, 2Ch, 0)
    <overlay-23 entry#3>(目標, 1Fh, 0)
if <overlay-23 entry#16>(目標, 32h) then
    result := 1
    <overlay-23 entry#3>(目標, 39h, 0)
DS:6F9Ch := 0
回傳 result
```

**被檢查的編號和被解掉的編號不一樣**：

| 檢查 | 解除 |
|---|---|
| `22h` | （不解，只回報） |
| `2Bh` | `2Ch`、`1Fh` |
| `32h` | `39h` |

所以這不是「有什麼就解什麼」，是一張寫死的對應表。remake 若寫成「解掉查到的
那個」，`2Bh`／`32h` 會留在身上，而 `2Ch`／`1Fh`／`39h` 不會被清掉。

`DS:6F9Ch` 在整段期間為 1，離開時歸 0——`entry#3` 很可能會看它。

## `4311h`：兩條路

```text
if 目標^[0E5h] >= 6 then                        ← 有號 jge
    <overlay-24 entry#20>(目標, 'smashes them flat'（CS:42FFh）, 0Ah, 1)
else
    n := <overlay-24 entry#36>(DS:6F97h)
    <sub_F06h>(DS:6F97h, n, 0, 0, 8, 空字串（CS:42FEh）)
    if <overlay-24 entry#27>(目標, 3, @node) then       ← spec 692 的「用 key 找節點」
        <overlay-23 entry#2>(0, node, 目標, 3)
```

`CS:42FEh`（長度 0）與 `CS:42FFh`（長度 17）又是**緊鄰的兩條字串**，
和 spec 703 的 `4425h`／`4426h` 同一種寫法。

## `4289h` 少的是哪一行

spec 703 記錄了 `overlay-22:4289h` 把 `[bp−4]`／`[bp−2]` 這兩個從未寫入的
區域推給 `overlay-23 entry#2`。`4311h` 給出了那兩個位置**應該**是什麼：

```text
4311h：lea di, [bp+var_4] ; push ss ; push di ; call <entry#27>   ← 由它填入
       ...
       push [bp+var_2] ; push [bp+var_4] ; call <entry#2>

4289h：（沒有對應的 entry#27 呼叫）
       push [bp+var_2] ; push [bp+var_4] ; call <entry#2>
```

兩支的最後那兩條 `push` **逐位元組相同**（`ff76fe` ＋ `ff76fc`）。差別只在
`4311h` 先用 `entry#27` 把那個 far 指標填進去，`4289h` 沒有。

所以 `4289h` 傳給 `entry#2` 的第二個參數位置，在 `4311h` 裡是**找到的節點指標**。
這使得「`4289h` 是漏抄了填值那一步」成為最合理的解釋——但**仍需實機驗證**：
也可能 `entry#2` 在 `4289h` 走的那條路徑（第五個參數 `40h` 對 `3`）根本不讀它。

在驗證之前，remake 不應該自行補上 `entry#27` 那一步，也不應該改傳 0——兩者都
是在猜原作的行為。

## 明確不宣稱

- `overlay-23` entry#3／#16 的內部行為、`DS:6F9Ch` 的語意。
- 編號 `1Fh`／`22h`／`2Bh`／`2Ch`／`32h`／`39h` 各代表什麼效果。
- `目標^[0E5h]` 是什麼、為什麼門檻是 6。
- `4311h` 對應哪一個法術。
