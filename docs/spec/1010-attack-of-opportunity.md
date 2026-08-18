# 1010 — 離開接觸就被打：先算差集，再要求對方「面向」你

- 證據等級：`exact`（DOS 側 334 條逐條讀完；PC-98 對側 345 條，
  多出來的 11 條是一道 DOS 沒有的額外條件，已逐條讀完）
- 作法見 spec 783

## `dos overlay-13:0095Ah` ↔ `pc98 overlay-13:00999h`（`retf 6`）

兩側原本都是待解讀。這是 spec 1008 的移動指令走一步時叫的 `overlay-13 entry#6`。

```pascal
procedure 離開接觸(角色: 遠指標;   { bp+08h..0Bh }
                   方向: byte);    { bp+06h }
```

★ PC-98 的 Borland 除錯符號把這支叫 **`CHECKPARTINGBLOWS`**（parting blow ＝
離場時挨的那一下），與扇形判斷的 `INARC` 一樣是原作者的命名。

## ★★ 用「移動前後的相鄰名單取差集」判斷誰失去接觸

```pascal
FillChar(前名單, 0Ch, 0);                       { 最多 12 個 }
編號 := <overlay-32 entry#18>(角色);            { 這個戰鬥員在座標陣列裡的號碼 }

n前 := <overlay-24 entry#32>(1, 角色);          { 現在有幾個敵人貼著 }
if n前 = 0 then exit;
for i := 1 to n前 do 前名單[i] := byte[758Eh ＋ i];

{ ★ 把自己的座標暫時挪到目的格 }
byte[66A3h ＋ 編號 × 4] += dx[方向];
byte[66A4h ＋ 編號 × 4] += dy[方向];
n後 := <overlay-24 entry#32>(1, 角色);
byte[66A3h ＋ 編號 × 4] -= dx[方向];            { 立刻挪回來 }
byte[66A4h ＋ 編號 × 4] -= dy[方向];

{ 移動後還貼著的，從名單裡劃掉 }
for i := 1 to n前 do begin
    仍在 := false;
    for j := 1 to n後 do
        if byte[758Eh ＋ j] = 前名單[i] then 仍在 := true;
    if 仍在 then 前名單[i] := 0;
end;
```

> **實際位置只在兩次查詢之間變動一瞬間，查完就還原。**
> 剩在名單上的，就是「你這一步會離開他的接觸範圍」的那些人。

★ `DS:66A3h`／`66A4h` 是**戰鬥員座標陣列，每人 4 bytes**（`+0` 欄、`+1` 列），
以 `overlay-32 entry#18` 回的編號索引。`DS:758Eh` 是 `overlay-24 entry#32`
回填相鄰者編號的緩衝，也是 1 起算。

`dx`／`dy` 是 spec 999 的九格方向表。

## 每個攻擊者要過的四道關

```pascal
for i := 1 to n前 do begin
    if 前名單[i] = 0 then continue;
    if 角色^[196h] = 0 then continue;             { ★ 移動者本人不能行動就不觸發 }
    DS:75E5h := 1;  DS:75E6h := 1;
    打手 := 遠指標(DS:6D35h ＋ 前名單[i] × 4);

    if <overlay-24 entry#6>(打手) <> 0     then continue;
    if <sub_1144>(角色, 打手) = 0          then continue;
    if <overlay-24 entry#27>(@緩衝, 4Bh, 打手) <> 0 then continue;
    if <overlay-24 entry#27>(@緩衝, 4Ah, 打手) <> 0 then continue;
    …
```

`overlay-24 entry#27(@緩衝, 效果碼, 角色)` 是「這個人身上有沒有某個效果」的查詢
（`4Ah` 與 `4Bh` 兩個效果碼會讓他打不出這一擊）。

★ **`+196h` ＝ 1 是「站著、能行動」**——`overlay-23 entry#23`（`STANDUP`）
寫的是 `+195h := 0` 與 **`+196h := 1`**。
spec 576 的台帳註記把這兩行混成「`+196h` 設 0(可行動)」，已一併更正。

## ★★ 五個朝向裡有一個看得到就打

```pascal
首 := 打手^[18Dh]^[9] ＋ 6;                       { 朝向 − 2 }
末 := 打手^[18Dh]^[9] ＋ 0Ah;                     { 朝向 ＋ 2 }
for k := 首 to 末 do begin
    if 已經打過 then break;
    if (打手^[18Dh]^[3] <= 0) and (打手^[18Dh]^[0Fh] <> 0) then
        if <overlay-31 entry#4>(打手x, 打手y, 角色x, 角色y, k mod 8) = 0 then
            continue;                             { 這個朝向看不到 }
    …攻擊…
end;
```

★★ **`overlay-31 entry#4` 就是 spec 1002 讀完的那支 90° 扇形判斷**，
簽章 `(起點x, 起點y, 目標x, 目標y, 方向)` 一格不差；
本支是它的呼叫端——spec 1002 當時只能說「形狀上是『這個目標在不在我面前』」，
現在確定用途是**機會攻擊的面向檢查**。

`k` 跑 `朝向 ＋ 6 .. 朝向 ＋ 0Ah`，`mod 8` 之後是 **`朝向 −2, −1, 0, ＋1, ＋2`**
——正面加左右各 90°，合起來 180°。任何一個朝向讓對方落在扇形內就打得到。

⚠ `打手^[18Dh]^[3] > 0` 或 `打手^[18Dh]^[0Fh] = 0` 時**整個面向檢查被跳過**，
無條件打。

## 攻擊本身

```pascal
槽 := 1;
if 打手^[11Ch] = 0 then 槽 := 2;
for j := 1 to 2 do
    if 打手^[19Bh ＋ j] > 0 then 槽 := j;
if 打手^[19Bh ＋ 槽] = 0 then 打手^[19Bh ＋ 槽] := 1;
打手^[18Dh]^[4] := 槽;

原目標 := 打手^[18Dh]^[0Ah];                      { ★ 先存起來 }
<sub_19D8>(@DS:75DBh, NIL, 1, 角色, 打手);
已經打過 := 1;
打手^[18Dh]^[0Ah] := 原目標;                      { ★ 打完還原 }

if 角色^[196h] <> 0 then begin
    DS:75E5h := 1;
    <overlay-24 entry#5>(角色);
end;
```

★ **打完把「鎖定目標」還原**——機會攻擊不會改變攻擊者原本要打誰。

## ⚠ PC-98 多一道條件

DOS 334 條、PC-98 345 條，差別集中在攻擊前：

```pascal
if <overlay-24 entry#41>(打手) <> 0 then
    if <overlay-24 entry#42>(打手) = 0 then continue;    { ← DOS 沒有 }
```

也就是 **PC-98 多了一組「`entry#41` 成立但 `entry#42` 不成立就不打」的閘**。
兩支在 DOS 的 `overlay-24` 沒有對應呼叫，本規格不宣稱它們判斷什麼。

## 兩平台位址與欄位

| | DOS | PC-98 |
|---|---|---|
| 相鄰者緩衝 | `DS:758Eh` | `DS:0A820h` |
| 戰鬥員座標陣列 | `DS:66A3h`／`66A4h` | `DS:973Dh`／`973Eh` |
| 戰鬥員遠指標表 | `DS:6D35h` | `DS:9DCFh` |
| 兩個旗標 | `DS:75E5h`／`75E6h` | `DS:0A877h`／`0A878h` |
| 戰鬥狀態 | 角色 `+18Dh` | 角色 **`+18Eh`** |
| 能行動 | 角色 `+196h` | 角色 **`+197h`** |
| 武器槽計數 | 角色 `+19Bh` | 角色 **`+19Ch`** |
| 陣營 | 角色 `+11Ch` | 角色 `+11Ch`（**相同**） |

★ **PC-98 的 ＋1 位移從 `+11Ch` 之後、`+18Dh` 之前開始**——
`+11Ch` 兩平台相同，`+18Dh`／`+196h`／`+19Bh` 都要 ＋1。
這是 spec 641 那條「DOS 欄位位移」的具體邊界。

## 中文化

本支沒有字串。

## remake 對照

| remake | 原作 |
|---|---|
| `(*Battle).opportunityAttackFacingAllows` | 五方向迴圈 ＋ 兩個旁路 |
| `combat.InFacingCone` | `overlay-31 entry#4`（spec 1002） |
| `Fighter.CombatFacing`／`CombatActionCount`／`CombatAction.Delay` | `+18Dh` 的 `+09h`／`+0Fh`／`+03` |

`MoveWithTerrainAndFreeAttacks` 的「離開接觸就被打」那一段現在先過這道閘才動手。
⚠ 前面四道（`entry#6`、`sub_1144`、效果碼 `4Ah`／`4Bh`）還沒解讀，remake 沒有對應物，
所以現在的閘比原作**鬆**。

## 明確不宣稱

- 沒有宣稱 `overlay-24 entry#32`、`entry#6`、`entry#5`、`entry#27` 的內部行為。
- 沒有宣稱效果碼 `4Ah`／`4Bh` 是什麼（要查 spec 1005 的分派表才知道處理常式）。
- 沒有宣稱 `sub_1144`（同模組）判斷什麼，也沒有宣稱 `sub_19D8` 怎麼算攻擊。
- 沒有宣稱 `DS:75E5h`／`75E6h` 這兩個旗標給誰看。
- `打手^[18Dh]^[3]` 是先攻、`^[0Fh]` 是這一輪的動作計數（spec 1137 的四道閘），
  但沒有宣稱「先攻還沒歸零就無條件打」在遊戲設計上是什麼意思。
- 沒有宣稱 PC-98 多出來的 `overlay-24 entry#41`／`entry#42` 判斷什麼。
