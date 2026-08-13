# 第六百二十一輪：`ROBSTUFF` —— 偷竊機率與物品價值

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-07:1F00h`（176 bytes，62 條指令）。

```text
ROBSTUFF(char, chance):
    item := char^[14Eh]
    while item <> nil do
        p := chance
        if item^[5Fh] > 0FFh then      p := max(0, p − 5Ah)   ← 價值 > 255：−90
        else if item^[5Fh] > 18h then  p := max(0, p − 32h)   ← 價值 > 24：−50
        roll := <far 013E:004D>(1, 64h)                        ← d100
        next := item^[52h]
        if roll <= p then <移除>(item, char)
        item := next
```

## 物品 `+5Fh`（word）是價值

[spec 610](610-ecl-load-encounter-items.md) 只讀出「`+5Fh` 是 word」。這一輪
從偷竊機率的分級確認它是**價值**——兩個門檻 `18h`（24）與 `0FFh`（255）都拿它
比較。

## 機率依價值遞減

| 物品價值 | 機率調整 |
|---|---|
| `> 255` | **−90** |
| `> 24` | **−50** |
| 其他 | 不調整 |

**減完不會變負**：低於門檻一律歸 0（`if p <= 5Ah then p := 0`）。所以貴重物品
在 `chance <= 90` 時**完全偷不走**，不是「機率很低」而是 0。

判定是 `roll <= p`，`roll` 來自 `d100`（`1, 64h`）。`p = 0` 時 `roll` 最小為 1，
**永遠不成立**。

## 逐件判定，先取 next 再移除

每件物品各擲一次骰，不是全體共用一個結果。而且**先取 `next` 才移除**——與
`40h`（[spec 596](596-ecl-party-item-sweep.md)）同一個寫法，移除會釋放節點。

## 明確不宣稱

- `013E:004D` 是哪個 unit 的 `ROLLDICE`（`EFFECTS` 的在 overlay-23:1368h，
  這裡是 far call 到另一個 control segment）。
- `chance` 由呼叫端怎麼算。
