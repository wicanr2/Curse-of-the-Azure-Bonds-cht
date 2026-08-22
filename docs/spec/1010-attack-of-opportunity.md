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
    if <overlay-24 entry#27>(打手, 4Bh, @緩衝) <> 0 then continue;
    if <overlay-24 entry#27>(打手, 4Ah, @緩衝) <> 0 then continue;
    …
```

### ★★ 四道閘各自是什麼

`overlay-24 entry#27(角色, 效果碼, @緩衝)` 是「這個人身上有沒有某個效果」的查詢
——走一遍 `+0F2h` 起的效果串列比對每一節的第一個位元組，**不看持續時間**。

| 閘 | 內容 | 證據 |
|---|---|---|
| `entry#6(打手)` | 對 `entry#27` 連問四次，效果碼取自 `DS:27CAh`..`27CDh` ＝ **`33h`、`34h`、`35h`、`1Fh`** | DOS `overlay-24:0BE3h` 逐條讀完；那張表在已初始化資料區（BSS 界線是 `47E0h`，spec 804） |
| `sub_1144(角色, 打手)` | `CHECKFX(01h, 角色)` 與 `CHECKFX(00h, 打手)` 的否決權，任一支效果設了中止旗標 `DS:6F9Bh` 就是「不行」 | 兩個 far call 都解到 `overlay-23 entry#4` ＝ `CHECKFX`（spec 766 的結構 ＋ far call 對照表） |
| `4Bh`／`4Ah` | 士氣崩潰而且跑得掉時掛上的兩個效果碼（spec 831） | 同一組碼 |

★ **那四個碼就是「這一回合動不了」那一組**：`CHECKFX(07h)` 的成員，
四個共用同一支 handler `overlay-12:0075h`，其中 `35h` 印的是 `falls asleep`。
⚠ 它們是 `MonsterIsHeld` 那五個的**子集**——原作這張表**沒有 `1Bh`**。

★ **時機 `00h` 在 CoAB 是空的**（分派表裡沒有任何效果碼），
所以第二道實際上只剩 `CHECKFX(01h, 角色)` 那一問；而 `01h` 的成員是
`25h`（閃現）、`19h`／`47h`（隱形）、`45h`——**看不見離場的人就打不到**。

⚠ 參數是**由左往右推**，宣告順序是 `(角色, 效果碼, @緩衝)`；
`entry#27` 自己的框架（`arg_6` 讀 `+0F2h` 效果鏈頭、`arg_0` 回填）證實這個順序。

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
if 打手^[11Ch] = 0 then 槽 := 2;                  { BASEATTBLOWS[0] 是 0 就用槽 2 }
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
| `ATTBLOWS` 的 1-based 基底 | 角色 `+19Bh` | 角色 **`+19Ch`** |
| `BASEATTBLOWS[0]` | 角色 `+11Ch` | 角色 `+11Ch`（**相同**） |

★ **PC-98 的 ＋1 位移從 `+11Ch` 之後、`+18Dh` 之前開始**——
`+11Ch` 兩平台相同，`+18Dh`／`+196h`／`+19Bh` 都要 ＋1。
這是 spec 641 那條「DOS 欄位位移」的具體邊界。

## 中文化

本支沒有字串。

## remake 對照

| remake | 原作 |
|---|---|
| `(*Battle).opportunityAttackAllowed` | 動手前的四道閘，順序照原作 |
| `opportunityAttackBlockingAffects` ＝ `{33h,34h,35h,1Fh}` | `entry#6` 查的那張表（`DS:27CAh`）|
| `Fighter.VisibleTo` | `sub_1144` ⇒ `CHECKFX(01h)` 的成員 `25h`／`19h`／`47h`／`45h` |
| `opportunityAttackWithdrawnAffects` ＝ `{4Bh,4Ah}` | 連問兩次的 `entry#27` |
| `fighterHasAnyAffect` | `entry#27` 本身（只比對碼，不看持續時間）|
| `(*Battle).opportunityAttackFacingAllows` | 五方向迴圈 ＋ 兩個旁路 |
| `combat.InFacingCone` | `overlay-31 entry#4`（spec 1002） |
| `Fighter.CombatFacing`／`CombatActionCount`／`CombatAction.Delay` | `+18Dh` 的 `+09h`／`+0Fh`／`+03` |

`MoveWithTerrainAndFreeAttacks` 的「離開接觸就被打」那一段現在四道閘都會過。
⚠ 仍沒有對應物的是差集那一段本身（原作把座標暫時挪到目的格再查一次相鄰者），
remake 用的是「移動前後各算一次相鄰」的等價寫法，沒有借用座標陣列。

## 明確不宣稱

- 沒有宣稱 `overlay-24 entry#32`、`entry#5` 的內部行為。
- 沒有宣稱效果碼 `4Ah`／`4Bh` 各自代表什麼狀態，只宣稱它們由士氣崩潰的撤退掛上。
- 沒有宣稱 `sub_19D8` 怎麼算攻擊。
- 沒有宣稱 `CHECKFX(01h)` 的四個成員裡誰真的會設中止旗標——remake 用
  `VisibleTo` 覆蓋同一組碼，但那是從別處讀來的行為。
- 沒有宣稱 `DS:75E5h`／`75E6h` 這兩個旗標給誰看。
- `打手^[18Dh]^[3]` 是先攻、`^[0Fh]` 是這一輪的動作計數（spec 1137 的四道閘），
  但沒有宣稱「先攻還沒歸零就無條件打」在遊戲設計上是什麼意思。
- 沒有宣稱 PC-98 多出來的 `overlay-24 entry#41`／`entry#42` 判斷什麼。
