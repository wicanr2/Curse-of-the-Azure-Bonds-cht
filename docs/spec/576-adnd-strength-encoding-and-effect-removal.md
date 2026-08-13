# 第五百七十六輪：18/xx 力量的 byte 編碼與 effect 移除

狀態：`READY`。等級：`exact`（本文列出的函式逐指令讀過）。日期：2026-08-14
模組：`EFFECTS`（overlay-23），兩平台已配對，判讀同時適用 DOS。

## 18/xx 力量怎麼塞進一個 byte

AD&D 的力量是 3..18，而 18 還額外帶百分位 01..00（＝ 100）。原版用**一個
byte 同時表示這兩者**：

```text
CONVERTSTRTOSPEC(percentile, str) -> byte     ← PC-98 1693h／DOS 16A6h
    if str = 18 then return percentile + 1     ← 1..101
    else            return str + 100           ← 103..118

CONVERTSPECTOSTR(rec, out_pct, out_str)       ← PC-98 16BCh／DOS 16CFh
    v := rec^[3] and 7Fh
    if v <= 101 then  out_pct := v - 1;  out_str := 18
    else              out_pct := 0;      out_str := v - 100
```

兩者互逆。幾個直接的結論：

- **編碼欄位在角色記錄的偏移 `+3`，最高位是另作他用的旗標**（讀取時
  `and 7Fh` 遮掉）。
- `18/00`（最強）存的是 `101`，不是 `100`——百分位 `00` 在遊戲裡代表 100，
  加一之後是 101。remake 若把 `18/00` 存成 100，會與 `18/99` 撞號。
- 非 18 的力量存 `str + 100`，所以 3..17 佔 `103..117`，**`102` 沒有人用**。

`0x12`（18）這個數字的意義是由 `DONEWSTRENGTH` 釘死的，不是猜的——見下。

## `DONEWSTRENGTH`：只有變強才寫回

```text
DONEWSTRENGTH(out, ?, new_pct, new_str, char) -> boolean   ← PC-98 170Dh／DOS 1720h
    if new_str > char^[10h]
       or (new_str = 18 and new_pct > char^[1Dh]) then
        out^ := CONVERTSTRTOSPEC(new_pct, new_str)
        return true
    return false
```

於是角色記錄的兩個欄位也確定了：**`+10h` 是力量、`+1Dh` 是百分位**。
`new_str = 18` 才比較百分位——這就是 `0x12 = 18` 的證據。

`1756h`（PC-98／DOS 同址，4 個呼叫者）是同一條比較規則的另一種用法：參數是
一個指向 `SS` 上區域結構的指標 `p`，比較 `p[-2] : p[-1]`（力量）與
`p[-4] : p[-3]`（百分位），較大者寫回 `p[-1]`／`p[-3]`。形狀是「把當前值
同步到最大值欄位」。

## effect 的分派與移除

```text
CALLEFFECT(a, b, c, d, e, id)                 ← PC-98 00C9h／DOS 00C9h
    if DS:A66Eh <> 0 then
        far call DS:A28Ch (a, b, c, d, e)      ← 全域攔截，忽略 id
    else
        far call (DS:A040h + id*4)^ (a, b, c, d, e)
```

`DS:A040h` 起是 **每項 4 bytes 的 far pointer 表**，用 effect 編號索引。
`DS:A66Eh` 非零時整批改走 `DS:A28Ch` 這一支——攔截後 `id` 不再參與，所以那
是「不分種類一律做同一件事」的模式（**是什麼模式尚未確認**）。

```text
REMOVEFX(char)                                ← PC-98 158Ah／DOS 15A1h
    for i := 1 to 19 do SPELLOFF(char, DS:[159Dh + i], 0, 0)
    if <查找>(char, 4Dh, @tmp) then
        if char^[0F7h] = 0B3h then char^[198h] := 0

ROUTINGREMOVEFX(char)                         ← PC-98 15ECh／DOS 1603h
    for i := 1 to 4 do SPELLOFF(char, DS:[15B1h + i], 0, 0)

CUREEFFECT(id, char) -> boolean               ← PC-98 1630h／DOS 1643h
    if not <查找>(char, id, @tmp) then return false
    <格式化輸出>; <停頓(1, 10)>
    SPELLOFF(char, id, tmp, tmp2)
    return true
```

`REMOVEFX` 移除 19 種、`ROUTINGREMOVEFX` 移除 4 種，清單各自存在 `DS` 的
位元組表（`159Dh`／`15B1h`，1-based）。

## 明確不宣稱

- 兩張移除清單的**內容**。表在 `DS`，不在 overlay code 裡，要另外定位。
- `DS:A040h` 那張 effect 分派表的**項數與內容**。
- `DS:A66Eh` 這個攔截旗標代表什麼模式。
- `REMOVEFX` 尾段 `4Dh`／`0B3h`／`char^[0F7h]`／`char^[198h]` 的語意。
- `SPELLOFF`（`010Eh`）本體尚未讀。
