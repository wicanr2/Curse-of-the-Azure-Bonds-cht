# 1061 — 傷害訊息：單複數兩句、五種傷害來源、存活檢定減半／免疫

- 證據等級：`exact`（DOS 側 309 條逐條讀完）
- 作法見 spec 783

## `dos overlay-23:01FD6h`（`retf`）

原本待解讀。

```pascal
DS:6F94h := arg_4;                                   { 傷害量 }
if arg_0 <> 0 then begin                             { 有做存活檢定 }
    case arg_2 of
        1: DS:6F94h := 0;                            { ★ 完全免疫 }
        2: DS:6F94h := DS:6F94h div 2;               { ★ 減半 }
    end;
end;
if DS:6F94h <= 0 then exit;                          { 沒傷害就什麼都不顯示 }
```

★ `arg_2` ＝ 存活檢定的結果碼；`1` 全免、`2` 半傷、其他照原值。

## ★★ 單複數是兩句完整的字串

```pascal
if DS:6F94h = 1 then
    訊息 := 'takes 1 point of damage '               { 24 bytes，整句 }
else
    訊息 := 'takes ' ＋ 數字 ＋ ' points of damage '; { 6 ＋ 18 }
```

> ★★ **`'takes 1 point of damage '` 是獨立的一整句**，不是把 `s` 拿掉。
> 中文沒有單複數，**兩句可以合併成同一句**（與 spec 1018 的
> `' Gem'`／`' Gems'` 同一種處理）。

## ★★★ 五種傷害來源（`DS:6F95h and 0F7h`）

| 值 | 字串 | 長度 |
|---|---|---|
| `1` | `'from Fire'` | 9 |
| `2` | `'from Cold'` | 9 |
| `4` | `'from Electricity'` | 16 |
| `10h`（16） | `'from Acid'` | 9 |
| 其他 | `'from Magic'` | 10 |

★ **`and 0F7h` 把位元 3（`08h`）清掉** ⇒ 傷害類型的 bit 3 是另一件事，
不參與這裡的分類。⚠ 值是 `1`／`2`／`4`／`10h` 的**位元**，但比對是**相等**
——同時帶兩種來源時會落到 `'from Magic'`。

## ★★ 掉法術

```pascal
if DS:4FBAh = 5 then begin                           { 畫面模式 5（戰鬥） }
    角色^[18Dh]^[1] := 0;
    if 角色^[18Dh]^[0] > 0 then begin
        顯示('lost a spell');                        { 12 bytes }
        <overlay-24 entry#16>(角色^[18Dh]^[0], …);
        角色^[18Dh]^[0] := 0;
    end;
end;
```

★ `角色^[18Dh]` 是既有結論裡**載入後要歸零的四個遠指標之一**（spec 1038）
——這裡看到它指的記錄有 `+0`／`+1` 兩個 byte，`+0` 非 0 就掉一個法術。

## ★★★ 倒下／瀕死／死亡

```pascal
if 角色^[196h] = 0 then begin                        { ★ 不能行動 }
    訊息 := 'Goes Down';                             { 9 }
    if 角色^[195h] = 5 then 訊息 := 訊息 ＋ ', and is Dying';   { 14 }
    if 角色^[195h] in [0] then 訊息 := 'is killed';   { 9，★ 整句換掉 }
    …顯示…
end;
```

集合常數 `CS:1FACh` ＝ **`{0}`**。

| `角色^[195h]` | 顯示 |
|---|---|
| `0` | **`'is killed'`**（整句取代） |
| `5` | `'Goes Down, and is Dying'` |
| 其他 | `'Goes Down'` |

> ★★★ **`+196h` ＝ 能不能行動、`+195h` ＝ 倒下的原因。**
>
> ★★ **與 `STANDUP` 的矛盾已解掉**：重讀 `overlay-23:24EDh` 確認它寫的是
> `+195h := 0` **加上** `+196h := 1`。而上面整段判斷**包在 `if +196h = 0` 裡面**
> ——站起來的角色 `+196h ＝ 1`，根本走不到那一段。
> ⇒ **`+195h ＝ 0` 不是「死亡」，是「沒有特殊的倒下原因」**；
> 真正代表死亡的是 **`+196h ＝ 0` 且 `+195h ＝ 0`** 這個組合。
> `STANDUP` 把 `+195h` 清成 0 是「清掉倒下原因」，語意一致。

## 中文化

| DOS | 長度 | 建議中文 |
|---|---|---|
| `'takes '` ＋ 數字 ＋ `' points of damage '` | 6 ＋ 18 | 「受到」＋數字＋「點傷害」 |
| `'takes 1 point of damage '` | 24 | **與上一句合併** |
| `'from Fire'`／`'from Cold'`／`'from Electricity'`／`'from Acid'`／`'from Magic'` | 9／9／16／9／10 | 「（火焰／冰寒／閃電／強酸／魔法）」 |
| `'lost a spell'` | 12 | 「失去一個法術」 |
| `'Goes Down'` | 9 | 「倒下了」 |
| `', and is Dying'` | 14 | 「，瀕死」 |
| `'is killed'` | 9 | 「死亡」 |

⚠ 「takes N points of damage from Fire」是**三段接出來的**，
中文語序要改成「受到 N 點火焰傷害」——**來源要挪到傷害前面**。

## 明確不宣稱

- 沒有宣稱 `sub_3FEh(6)`／`(14h)`／`(0Dh)` 三次呼叫做什麼（形狀上是音效）。
- 沒有宣稱傷害類型 bit 3（被 `and 0F7h` 清掉的那一位）是什麼。
- 沒有宣稱 `角色^[18Dh]` 指的記錄除了 `+0`／`+1` 之外的欄位。
- 沒有宣稱 `DS:65A1h`（顯示位置）與 `overlay-32 entry#20` 做什麼。
- 沒有宣稱 `+195h` 的其他值。
