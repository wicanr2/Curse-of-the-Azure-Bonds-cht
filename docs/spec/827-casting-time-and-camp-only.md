# 827 — 法術屬性 `+0Bh` ＝ 戰鬥中可不可施放、`+0Ch` ＝ AD&D 的施法時間（節）

- 證據等級：`exact`（DOS 側逐條讀完；PC-98 側的差異已逐條列出；兩欄的值取自
  常駐資料段 dump）
- 作法見 spec 783

## `overlay-13:275Ah`（entry#23，DOS 132 條／PC-98 135 條）

| | DOS | PC-98 |
|---|---|---|
| 只能紮營施放 | `'Camp Only Spell'`（15） | `'キャンプでのみ使う呪文だ'`（24 bytes） |
| 開始詠唱 | `'Begins Casting'`（14） | `'は呪文を唱え始めた。'`（20 bytes） |

`retf 8` ＝ 4 個 word。依宣告順序：`(法術, 參數4, 輸出遠指標)`。

```pascal
p := DS:6506h;
輸出^ := 0;
if 法術 = 0 then
    法術 := <far 102Ch>(@var_6, @var_5, 1, 0);        { 讓玩家選 }

if (法術 > 0) and (byte[37E5h + 法術 * 10h] = 0) then begin
    顯示('Camp Only Spell');
    <far 154Ch+3>();
    法術 := 0;
end;

if 參數4 = 0 then begin
    <far 15B3h>();
    DS:75E6h := 1;  DS:75E5h := 1;
    <far 194Dh+1>(1, 3, p);
    <far 1505h+4>(p);
end;

if 法術 <= 0 then 離開;                               { 無號 }
延遲 := 有號(byte[37E6h + 法術 * 10h]) div 3;

if 延遲 = 0 then begin
    <far 11D9h>(法術, 參數4, 1, 輸出);
    輸出^ := <far 159Ah>(p);
end else begin
    輸出^ := 1;
    <far 1552h+2>(p, 名字 ＋ 'Begins Casting', 0Ah, 1);
    p^[18Dh]^[00h] := 法術;
    if 有號(p^[18Dh]^[3]) > 延遲 then
        p^[18Dh]^[3] := p^[18Dh]^[3] − 延遲
    else
        p^[18Dh]^[3] := 1;
end;
```

## `+0Bh`（`DS:37E5h`）＝ 0 就是「只能紮營施放」

100 筆裡只有 **8 筆是 0**，而且**正好是 AD&D 裡的非戰鬥法術**：

| 編號 | 名稱 |
|---|---|
| 14 | `Friends` |
| 18 | `Read Magic` |
| 22 | `Find Traps` |
| 31 | `Knock` |
| 35 | `Strength` |
| 39 | `Cure Disease` |
| 67 | `Neutralize Poison` |
| 75 | `Raise Dead` |

其餘 92 筆的值是 `1`（44 筆）或 `2`（48 筆）。**本支只判是不是 0**，兩個非零值的
差別要靠別的函式才知道（形狀上像「敵對／友方」：`Curse` / `Cause Light Wounds` /
`Burning Hands` / `Charm Person` 是 1，`Bless` / `Cure Light Wounds` /
`Detect Magic` / `Protection From Evil` 是 2）。

## `+0Ch`（`DS:37E6h`）＝ 施法時間，單位是 AD&D 的「節」

| 法術 | `+0Ch` | AD&D 施法時間 | `div 3` 後的延遲 |
|---|---|---|---|
| `Magic Missile` | 1 | 1 節 | 0（立即） |
| `Sleep` | 1 | 1 節 | 0 |
| `Fireball` | 3 | 3 節 | 1 |
| `Lightning Bolt` | 3 | 3 節 | 1 |
| `Bless` | `0Ah`（10） | 1 回合 ＝ 10 節 | 3 |
| `Cure Disease` | `64h`（100） | 1 大回合 ＝ 100 節 | 33 |

**單位換算完全對得上**（1 回合 ＝ 10 節、1 大回合 ＝ 10 回合 ＝ 100 節）。
遊戲把它 **÷ 3** 換成自己的延遲單位，`div` 是**有號**的。

`+0Ch` 的分佈：`0`（12 筆）、`1`（17）、`2`（5）、`3`（13）、`4`（13）、`5`（13）、
`6`（4）、`7`（6）、`8`（4）、`10`（11）、`100`（2 ＝ `Cure Disease` /
`Cause Disease`）。

## 延遲的代價直接扣先攻值

```
p^[18Dh]^[00h] := 法術;                      { 正在詠唱的法術編號 }
if 先攻 > 延遲 then 先攻 := 先攻 − 延遲 else 先攻 := 1;
```

- **戰鬥狀態 `+00h` ＝ 正在詠唱的法術編號**。spec 804 讀到 `overlay-08:026Bh` 會判
  `+18Dh^[0] = 0`，現在知道那是「有沒有在詠唱」。
- 先攻值（`+3`，spec 806）被扣掉延遲，**扣到 0 以下就變成 1**（不會消失，只是排到
  最後）。
- 延遲 ＝ 0 的法術走另一條路，當場結算，**不動先攻值**。

## 明確不宣稱

- 沒有宣稱 `+0Bh` 的 `1` 與 `2` 差在哪。
- 沒有宣稱 `far 102Ch` 的選單長什麼樣。
- 沒有宣稱 `far 11D9h` / `159Ah` / `15B3h` / `194Dh+1` / `1505h+4` 各做什麼。
- 沒有宣稱 `參數4` ＝ 0 那一段（重畫？）的意義。
- PC-98 在 `11D9h` 那條路之前多兩條 `mov al,[bp+arg_6]` ／ `mov ds:0BE2Dh, al`，
  本規格不宣稱那個全域的用途。
