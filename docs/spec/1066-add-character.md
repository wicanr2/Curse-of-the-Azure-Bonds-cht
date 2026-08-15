# 1066 — 加入角色：`DS:7584h` 是「從哪一款遊戲來的」，以及入隊的四條限制

- 證據等級：`exact`（DOS 側 378 條逐條讀完）
- 作法見 spec 783

## `dos overlay-17:03680h`（`retf`）

原本待解讀。

## ★★★★ 關掉 spec 1038／1043 的核心疑問

```pascal
鍵 := <overlay-26:003D4h>('Add from where? ', 'Curse Pool Hillsfar Exit', …);
case 鍵 of
    'C': DS:7584h := 0;      { ★ Curse of the Azure Bonds }
    'P': DS:7584h := 1;      { ★ Pool of Radiance }
    'H': DS:7584h := 2;      { ★ Hillsfar }
    'E', #0: exit;
end;
```

> ★★★★ **`DS:7584h` 就是「這個角色來自哪一款遊戲」。**
> 對照 spec 1038 的三種記錄長度與 spec 1043 的三個副檔名：
>
> | 鍵 | `DS:7584h` | 遊戲 | 記錄長度 | 副檔名 |
> |---|---|---|---|---|
> | `C` | 0 | **Curse of the Azure Bonds**（本作） | 422／423 | `.guy` |
> | `P` | 1 | **Pool of Radiance** | 285 | `.cha` |
> | `H` | 2 | **Hillsfar** | 188 | `.sav` |
>
> ⇒ spec 1038 猜的「很可能是《Pool of Radiance》與更早的角色檔」**確認**，
> 而第三個是 **Hillsfar**（同系列的動作 RPG），不是更早的版本。
> ⇒ spec 1043 的「版本就是副檔名」也在這裡對上。

★ 選項行 `'Curse Pool Hillsfar Exit'`（24）的熱鍵就是 `C`／`P`／`H`／`E`
——spec 1060 的「掃字串比對」機制。

## 選檔與建立

```pascal
<overlay-16 entry#1>(@清單, …);                     { 掃出可用的角色檔 }
if 清單 = NIL then exit;
鍵 := <overlay-26 entry#5>('Add a character: ', 'Add ', …, @清單, …);
if 鍵 not in [0Dh, 'A'] then …重來…;
if 選中項^[1] = '*' then …不能選…;                   { ★ 已在隊伍裡的標記 }
新角色 := GetMem(1A6h);                              { ★ 422 bytes }
<overlay-26 entry#1>(…);
<overlay-16 entry#7>(新角色, …, 0, 1);               { ＝ spec 1038 的讀檔 }
顯示('* ' ＋ 檔名);                                  { ★ 加上已選標記 }
```

★ `'* '`（2 bytes）是**已加入的標記**，直接寫回清單項的前兩格
——與 spec 1057 的「☆」是同一種就地覆蓋。
★ `GetMem(1A6h)` ＝ 422 ⇒ **不管來源是哪一款，記憶體裡一律是本作的版面**
（舊格式在 spec 1038 的讀檔函式裡轉換）。

## ★★★ 掃全隊算四個量，再判四條限制

```pascal
for 每個隊員 p do begin
    if (p.名字 = 新角色.名字) and (p^[126h] = 新角色^[126h]) then 拒絕;  { 重複 }
    if p^[0F7h] < 80h            then inc(NPC數);
    if p^[10Dh] > 0              then inc(遊俠數);
    if ((p^[11Bh] ＋ 1) mod 3) = 0 then 有邪惡 := 1;
    if p^[10Ch] > 0              then 有聖騎士 := 1;
end;
```

| 欄位 | 意義 |
|---|---|
| `+0F7h` | 劇情 NPC 旗標（`< 80h` ＝ NPC，spec 623／1028／1044） |
| `+10Ch` | **聖騎士等級**（＝ spec 1034／1055 的 `+109h ＋ 3`） |
| `+10Dh` | **遊俠等級**（＝ `+109h ＋ 4`） |
| `+11Bh` | **陣營** |
| `+126h` | 用來配合名字判重複的另一個欄位 |

> ★★★ **`+10Ch`／`+10Dh` 落在 `+109h ＋ j` 的職業槽上**，
> `j ＝ 3`（聖騎士）與 `j ＝ 4`（遊俠）——與 spec 1034／1055 的職業順序
> **第三次對上**。
>
> ★★★ **陣營編碼**：`(陣營 ＋ 1) mod 3 = 0` ⇒ 陣營 ∈ `{2, 5, 8}` 是邪惡。
> 這只有在順序是
> **LG(0) LN(1) LE(2) NG(3) N(4) NE(5) CG(6) CN(7) CE(8)**
> 時成立——**「守序／中立／混亂」為外層、「善／中立／惡」為內層**。

### 四條限制

```pascal
if 新角色^[0F7h] < 80h then     { 新的是 NPC }
    需要 NPC數 < 6
else
    需要 bank1^[67Ch] < 8;      { ★ 隊伍上限 8 }
if (新角色^[10Ch] <> 0) and 有邪惡  then 拒絕;    { ★ 聖騎士不與邪惡同隊 }
if (新角色^[10Dh] <> 0) and (遊俠數 >= 3) then 拒絕;  { ★ 遊俠最多 3 個 }
```

對應三句拒絕訊息：

| 訊息 | 長度 |
|---|---|
| `'paladins do not join with evil scum'` | 35 |
| `'too many rangers in party'` | 25 |
| `名字 ＋ ' will tolerate no evil!'` | 23 |

> ★★★ **這三條正是 AD&D 1e 的組隊規則**：聖騎士不與邪惡同行、
> 遊俠不超過三人同隊。**上限 8 人**存在 `bank1^[67Ch]`。

## 中文化

| DOS | 長度 | 建議中文 |
|---|---|---|
| `'Add from where? '` | 16 | 「從哪裡加入？」 |
| `'Curse Pool Hillsfar Exit'` | 24 | ⚠ **熱鍵 C／P／H／E 靠掃字串**，中文要改用 PC-98 熱鍵表 |
| `'Add a character: '` | 17 | 「加入角色：」 |
| `'Add '` | 4 | 「加入」 |
| `'* '` | 2 | ⚠ **就地覆蓋前兩格，中文要留同寬度** |
| `'paladins do not join with evil scum'` | 35 | 「聖騎士不與邪惡之徒同行」 |
| `'too many rangers in party'` | 25 | 「隊伍裡的遊俠太多了」 |
| 名字 ＋ `' will tolerate no evil!'` | 23 | 名字＋「不容許邪惡同行！」 |

⚠ `Curse`／`Pool`／`Hillsfar` 是**遊戲名**，中文可作
「詛咒」「輻光之池」「希爾斯法」或保留原文。

## 明確不宣稱

- 沒有宣稱 `+126h`（配合名字判重複的那一格）是什麼。
- 沒有宣稱 `bank1^[67Ch]` 除了「目前人數」之外的用途。
- 沒有宣稱 NPC 上限 6 與隊伍上限 8 是不是同一個計數。
- 沒有宣稱 `overlay-26 entry#1`／`entry#4` 做什麼。
★ **陣營編碼已由原作字串表確認**（spec 1100）：`0` 守序善良、`1` 守序中立、
`2` 守序邪惡、`3` 中立善良、`4` True Neutral、`5` 中立邪惡、`6` 混亂善良、
`7` 混亂中立、`8` 混亂邪惡——`陣營 = 守序軸×3 + 善惡軸`，`{2,5,8}` 正是三個 Evil。
