# 842 — 眼魔的眼柄光線：一回合最多四道，每種只用一次

- 證據等級：`exact`（DOS 側逐條讀完；PC-98 側**助憶碼完全相同**，差異全是位址與字串）
- 作法見 spec 783

## `overlay-13:042FDh` ↔ `overlay-13:0426Bh`（entry#40，兩平台各 312 條）

`retf 0Ah` ＝ 5 個 word，與 spec 819／826／831 的怪物特殊攻擊同一種簽章：
`(角色, 參數4, 參數2)`——**本支只用到角色**，另外三個 byte 沒被讀。

| | DOS | PC-98 |
|---|---|---|
| 分解 | `'fires a disintegrate ray'` | `'は原子分解光線を発射した。'` |
| 分解成功 | `'is disintergrated'`（**原版錯字**） | `'は粉々になった。'` |
| 石化 | `'fires a stone to flesh ray'`（**寫反了**） | `'は石化光線を発射した。'` |
| 石化成功 | `'is Stoned'` | `'は石になった。'` |
| 死亡 | `'fires a death ray'` | `'は殺人光線を発射した。'` |
| 死亡成功 | `'is killed'` | `'は死んだ。'` |
| 傷害 | `'wounds you'` | `'はきみを傷つけた。'` |

```pascal
用過 := 0;  剩 := 4;
角色^[18Dh]^[0Ah] := NIL;
<本模組 41CCh>(1);                                  { 巢狀程序，挑第一個目標 }

repeat
    目標 := 角色^[18Dh]^[0Ah];
    r    := <far 1593h+2>(角色, 目標);
    dec(剩);
    if 目標 = NIL then break;

    if      (r <= 2) and (用過 and 01h = 0) then begin
        用過 := 用過 or 01h;
        <far 1552h+2>(角色, 名字 ＋ 'fires a disintegrate ray', 0Ah, 1);
        <本模組 4162h>(5);
        if not <far 1457h+1>(目標, 3, 0) then
            <far 1434h+1>(目標, 8, 名字 ＋ 'is disintergrated');
        <本模組 41CCh>(0);
    end
    else if (r <= 3) and (用過 and 02h = 0) then begin
        用過 := 用過 or 02h;
        訊息('fires a stone to flesh ray');
        <本模組 4162h>(0Ah);
        if not <far 1457h+1>(目標, 1, 0) then
            <far 1434h+1>(目標, 7, 名字 ＋ 'is Stoned');
        <本模組 41CCh>(0);
    end
    else if (r <= 4) and (用過 and 04h = 0) then begin
        用過 := 用過 or 04h;
        訊息('fires a death ray');
        <本模組 4162h>(5);
        if not <far 1457h+1>(目標, 0, 0) then
            <far 1434h+1>(目標, 6, 名字 ＋ 'is killed');
        <本模組 41CCh>(0);
    end
    else if (r <= 5) and (用過 and 08h = 0) then begin
        用過 := 用過 or 08h;
        訊息('wounds you');
        <本模組 4162h>(5);
        <far 1492h+2>(目標, <far 1461h+1>(2, 8) + 1, 0, 0);
        <本模組 41CCh>(0);
    end
    else if 用過 and 10h = 0 then begin
        <far 11D9h>(54h, 1, 1, @暫存);  用過 := 用過 or 10h;   { Fear }
    end
    else if 用過 and 20h = 0 then begin
        <far 11D9h>(37h, 1, 1, @暫存);  用過 := 用過 or 20h;   { Slow }
    end
    else if 用過 and 40h = 0 then begin
        <far 11D9h>(15h, 1, 1, @暫存);  用過 := 用過 or 40h;   { Sleep }
    end;
until (剩 <= 0) or (角色^[18Dh]^[0Ah] = NIL);       { 無號 }
```

## 三個常數就是法術編號

`54h` ＝ 84、`37h` ＝ 55、`15h` ＝ 21。查 `DS:27BDh` 的法術名稱表
（`docs/audit/resident-data-tables.md`）：

| 編號 | 法術 |
|---|---|
| 84 | **Fear** |
| 55 | **Slow** |
| 21 | **Sleep** |

配上四道具名光線（分解、石化、死亡、傷害），這正是 AD&D 眼魔（Beholder）的眼柄
能力表——**這一支就是眼魔的特殊攻擊**。

## `far 1434h+1(目標, 狀態碼, 訊息)` ＝ 設狀態並報訊息

三個呼叫點的狀態碼分別是 **8**（分解）、**7**（石化）、**6**（死亡），
與 spec 840 那個「算死亡」的 Pascal 集合 `[6, 7, 8]` 完全對上。
`+195h` 的值表因此再確認一次：

完整的狀態名稱表（遊戲自己印的字）見 spec 825：`6` ＝ `Dead`、`7` ＝ `Stoned`、
`8` ＝ `Gone`。本支寫的三個值分別是 8（分解）、7（石化）、6（死亡）。

Turn Undead 的「摧毀」與眼魔的「分解」共用同一個狀態碼 `8`，也共用
spec 840 那道「`+195h` ＝ 8 就不印攻擊結果」的規則。

## `far 1457h+1(目標, 類別, 0)` ＝ 豁免檢定

三道光線各自帶不同的類別碼：分解 `3`、石化 `1`、死亡 `0`。
**回傳真代表擋下來了**（成立就不掛狀態），與 spec 832 的用法一致。

## 一回合最多四道

`剩` 從 4 開始，每繞一圈先減 1，`剩 ≤ 0` 就結束。所以**四道具名光線之後就停**，
後面那三個施法分支（Fear／Slow／Sleep）在 `剩` 用完之前**幾乎打不到**——
只有前面四種都因為 `r` 落不進區間而跳過時，才輪得到。

`用過` 是七個位元的旗標，**每種能力一回合只用一次**。

## 中文化：兩處原文有問題

- **`'is disintergrated'` 少了一個字母**（應為 `disintegrated`）。
  PC-98 翻成 `'は粉々になった。'` 時沒有沿用錯字。
- **`'fires a stone to flesh ray'` 寫反了**：這道光線的效果是**把目標變成石頭**
  （後續掛 `+195h := 7` 並印 `'is Stoned'`），AD&D 的原名是 *flesh to stone*。
  PC-98 翻成 `'は石化光線を発射した。'`——**照效果翻，不照原文翻**。
  中文化跟 PC-98 一致即可。

`'is Stoned'` 與 `'is killed'` 在別的 overlay 也各有一份（spec 832／840），
**各自獨立，不能用單一取代**。

## 明確不宣稱

- 沒有宣稱 `far 1593h+2(角色, 目標)` 回傳什麼（形狀上是 0..5 的距離或關係分級）。
- 沒有宣稱 `本模組 41CCh(旗標)` 與 `4162h(n)` 各做什麼（前者形狀上是「挑下一個目標」，
  後者是動畫）。
- 沒有宣稱 `far 1461h+1(2, 8)` 的參數順序（形狀上是擲骰）。
- 沒有宣稱 `far 11D9h(法術, 1, 1, @暫存)` 的後三個參數。
- 沒有宣稱豁免類別碼 `0`／`1`／`3` 各對應 AD&D 的哪一欄。
