# 1021 — 轉職畫面：AD&D 的種族／職業矩陣，以及「1 環先給 1 個」的另一端

- 證據等級：`exact`（DOS 側 360 條、PC-98 側 437 條，兩邊各自逐條讀完）
- 作法見 spec 783

## `dos overlay-25:00EE1h` ↔ `pc98 overlay-25:00F30h`

原本兩側都待解讀。這是**選新職業（轉職）**的畫面。

## ★★★ 種族可選職業表 `DS:3FF8h`，每種族 14 bytes

```pascal
n := byte[3FF8h ＋ 種族 × 14];               { +0 ＝ 有幾個選項 }
for i := 1 to n do
    職業 := byte[3FF8h ＋ 種族 × 14 ＋ i];
```

配上職業名表 `DS:0CB6h`（筆距 27）與種族名表 `DS:0E9Ch`（筆距 10）：

| 種族 | 數 | 可選職業 |
|---|---|---|
| Dwarf | 3 | Fighter、Thief、Fighter/Thief |
| Elf | 7 | Fighter、Magic-User、Thief、Fighter/Magic-User、Fighter/Thief、Fighter/Magic-User/Thief、Magic-User/Thief |
| Gnome | 3 | Fighter、Thief、Fighter/Thief |
| Half-Elf | **13** | Cleric、Fighter、Magic-User、Thief、Ranger、Cleric/Fighter、Cleric/Ranger、Cleric/Fighter/Magic-User、Cleric/Magic-User、Fighter/Magic-User、Fighter/Thief、Fighter/Magic-User/Thief、Magic-User/Thief |
| Halfling | 3 | Fighter、Thief、Fighter/Thief |
| Half-Orc | 6 | Cleric、Fighter、Thief、Cleric/Fighter、Cleric/Thief、Fighter/Thief |
| Human | 6 | Cleric、Fighter、Magic-User、Thief、Paladin、Ranger |

> ★★★ **這就是 Gold Box／AD&D 1e 的種族—職業矩陣**：
> 半精靈 13 種組合最多、人類只能單職業（但可以轉職，這一支就是轉職）、
> 矮人／地精／半身人一律只有戰士／盜賊／戰士＋盜賊。

⚠ 種族 0（`Monster`）那一列的 `+0` 是 `18`，**超過 14 bytes 的列長**——
那一列不是真的職業清單，不要當資料用。

## 流程

```pascal
建一條選單鏈（每個節點 GetMem 2Eh ＝ 46 bytes，`+2Ah` next、`+29h` 旗標）；
for i := 1 to n do begin
    職業 := 表[i];
    if <sub_D2F>(職業, 父bp) then begin          { 資格檢查 }
        接一個節點;  節點^ := 字串(DS:0CB6h ＋ 職業 × 27);
    end;
end;

if 一個都沒有 then
    顯示 <名字> ＋ " doesn't qualify."
else begin
    鍵 := <選單>('Pick New Class', 'Select', …);
    if 鍵 = 'S' then begin
        角色^[127h] := 0;  角色^[129h] := 0;
        角色^[11Ch] := 2;
        …
        FillChar(角色^[1Eh], 54h, 0);            { ★ 清空記憶法術清單 }
        if 新職業 = 0 then 角色^[12Dh] := 1      { Cleric }
        else if 新職業 = 5 then begin            { Magic-User }
            角色^[137h] := 1;                    { ★ 1 環先給 1 個 }
            角色^[83h] := 1;  角色^[8Ah] := 1;
            角色^[84h] := 1;  角色^[8Dh] := 1;
        end;
        角色^[75h] := 新職業;                    { 職業欄位 }
        顯示 <名字> ＋ ' is now a 1st level ' ＋ 字串(DS:0CB6h ＋ 新職業 × 27);
        …重算…
    end;
end;
```

## ★★ 兩處與 spec 1016 對上

- **`FillChar(角色^[1Eh], 54h, 0)`** —— spec 1016 讀重算施法能力時掃的是
  `for i := 0 to 53h do 角色^[1Eh ＋ i]`，**同一塊 84 bytes 的記憶法術清單**。
  轉職會把它整個清空。
- **`角色^[137h] := 1`** —— spec 1016 的重算也是先寫 `+137h := 1`
  再從等級 2 開始累加。**「1 環先給 1 個」在兩支獨立的程式裡都寫死。**

## 兩平台差異：又一次 `+28h`

| | DOS | PC-98 |
|---|---|---|
| 種族／職業表 | `DS:3FF8h` | `DS:708Dh`（**值完全相同**） |
| 選單節點大小 | `2Eh` ＝ **46** | `56h` ＝ **86** |
| 節點的 next | `+2Ah` | `+52h` |
| 節點的旗標 | `+29h` | `+51h` |
| 名稱複製長度 | `28h` ＝ 40 | `50h` ＝ 80 |
| 迴圈計數器 | `DS:7747h` | `DS:0A9D9h` |

`56h − 2Eh = 28h`——**選單節點也大 40 bytes**，和物品節點同一個原因
（名稱欄位加寬，spec 1013）。

⚠ **兩平台都拿全域當內層迴圈的計數器**（`DS:7747h`／`DS:0A9D9h`），
所以這一支**不可重入**。

## 中文化

| 字串 | 長度 | 備註 |
|---|---|---|
| `'Pick New Class'` | 14 | 選單標題 |
| `" doesn't qualify."` | 17 | 接在名字後面 |
| `'Select'` | 6 | 選項，**大寫 S 是熱鍵** |
| `' is now a 1st level '` | 20 | 前接名字、**後接職業名** |

⚠ `' is now a 1st level '` 中間夾在名字與職業名之間，
中文語序是「◯◯◯成為 1 級的〈職業〉」——**兩段都要留位置**，
而且英文的 `1st` 是寫死在字串裡的。

★ **職業名表 `DS:0CB6h` 筆距 27 ⇒ 中文最多 13 字**（既有結論），
像 `Cleric/Fighter/Magic-User` 這種三職業組合中文要壓到 13 字以內
（「牧師／戰士／法師」剛好 8 字）。

## 明確不宣稱

- 沒有宣稱 `sub_D2F`（資格檢查）用什麼條件。
- 沒有宣稱 `+127h`／`+129h`／`+11Ch`／`+12Dh`／`+83h`／`+84h`／`+8Ah`／`+8Dh` 各是什麼。
- 沒有宣稱種族 0（`Monster`）那一列的意義。
