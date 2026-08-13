# 第六百零三輪：`1Eh`（全隊屬性統計）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:149Ch`（544 bytes，199 條指令；IDA 標 50）。

```text
READVAR(6)
sel := <operand 1 的值> − 7FFFh               ← 與 external CALL 同一種選擇器編碼
min := FFh ; max := 0 ; sum := 0 ; count := 0
case sel of
    8001h:
        走角色鏈，對每人呼叫 <far 014A:00A7>（效果查找）
    0A5h..0ACh:
        idx := sel − 0A4h                     ← 1..8
        走角色鏈：v := node^[0E9h + idx]
                  min := min(min, v) ; max := max(max, v)
                  sum := sum + v ; count := count + 1
        avg := sum div count
    9Fh:
        同上，但取的是 node^[1A6h]
    其他：什麼都不做
<sub_143C>(bp)                                ← 把結果寫回四個 ECL 位址
```

## 選擇器用 `−7FFFh`

與 external `CALL`（opcode `2Dh`，[spec 561](561-ecl-external-call-registry.md)）
**同一種編碼**：operand 減 `7FFFh` 之後才是選擇器。`8001h` 同時出現在兩邊的
認得集合裡。

認得的選擇器只有三組：`8001h`、`0A5h..0ACh`、`9Fh`。**其餘一律什麼都不做**，
而且不設任何旗標——與 `2Dh` 對未知 selector 的處置一致。

## `+0E9h` 起是 8 個屬性

`0A5h..0ACh` 減 `0A4h` 得到 `1..8`，索引 `node^[0E9h + idx]`，所以實際欄位是
**`+0EAh`..`+0F1h` 共 8 個 byte**。

## `sub_143C`：共用的四值輸出

它拿 caller 的 `bp` 當參數，直接讀 caller frame 上的變數再寫回 ECL 位址：

```text
STOREVALUE(ss:[bp-13h], ss:[bp-11h])      ← min
STOREVALUE(ss:[bp-15h], ss:[bp-10h])      ← max
STOREVALUE(ss:[bp-17h], ss:[bp-0Fh])      ← avg
STOREVALUE(ss:[bp-19h], ss:[bp-0Ah])      ← 第四個值
```

**跨函式直接讀對方的 frame**——不是傳參數而是傳 `bp`。remake 若把它拆成一般
的函式呼叫要小心：這四個偏移是硬編碼的，改動 `1Eh` 的區域變數配置就會壞掉。

## 兩個數值細節

- `count = 0`（隊伍空）時 `sum div count` 會**除以零**。原版沒有防護。
- 平均是**有號**除法，且結果存成 byte。

## 明確不宣稱

- `+0EAh`..`+0F1h` 那 8 個屬性各是什麼。
- `+1A6h` 是什麼。
- `8001h` 那一路的完整效果。
