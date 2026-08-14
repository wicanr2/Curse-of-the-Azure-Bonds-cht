# 850 — 升級擲 HD 與有效等級：`DS:0818h`／`0820h` 是 AD&D 的生命骰表

- 證據等級：`exact`（DOS 側兩支逐條讀完；PC-98 側助憶碼序列完全相同；
  四張表由原始 bytes 解出，兩平台逐 byte 相同）
- 作法見 spec 783

## `overlay-17:04A12h` ↔ `overlay-17:050DAh`（兩平台各 98 條）— 升級擲 HD

`retf 6` ＝ 3 個 word。依宣告順序：`(角色, 職業遮罩)`。

```pascal
結果 := 0;
for c := 0 to 7 do begin
    if 有號(角色^[109h + c]) <= 0 then continue;         { 沒有這個職業 }
    if (byte[3EA8h + c] and 遮罩) = 0 then continue;      { 這次不算這個職業 }

    if 角色^[109h + c] >= byte[3EB9h + c] then            { 無號：過了 HD 上限 }
        case c of
            2, 3    : 結果 := 3;                          { Fighter／Paladin }
            0, 4, 6 : 結果 := 2;                          { Cleric／Ranger／Thief }
            5       : 結果 := 1;                          { Magic-User }
        end
    else begin
        n := byte[818h + c];
        if 有號(角色^[109h + c]) > 1 then n := 1;         { 只有一級才用多顆骰 }
        a := <far 145Bh+2>(n, byte[820h + c]);
        b := <far 145Bh+2>(n, byte[820h + c]);
        if b > a then a := b;                             { 無號：擲兩次取大 }
        結果 := 結果 + a;
    end;
end;
```

### 四張 8 bytes 的平行表（兩平台逐 byte 相同）

| c | 職業 | `DS:0818h` 骰數 | `DS:0820h` 面數 | `DS:3EA8h` 遮罩 | `DS:3EB9h` HD 上限 |
|---|---|---|---|---|---|
| 0 | Cleric | 1 | 8 | `02h` | 10 |
| 1 | Druid | 1 | 8 | `02h` | 15 |
| 2 | Fighter | 1 | 10 | `08h` | 10 |
| 3 | Paladin | 1 | 10 | `10h` | 10 |
| 4 | Ranger | **2** | 8 | `20h` | 11 |
| 5 | Magic-User | 1 | 4 | `01h` | 12 |
| 6 | Thief | 1 | 6 | `04h` | 11 |
| 7 | Monk | **2** | 4 | `04h` | 19 |

PC-98 對應位址：`DS:1354h`／`135Ch`／`6F3Ch`／`6F4Dh`。

**這就是 AD&D 一版的生命骰表**：Cleric d8、Fighter／Paladin d10、Ranger 2d8（一級）、
Magic-User d4、Thief d6、Monk 2d4（一級）——全部對得上。
「過了上限就給固定值」的 3／2／1 也正是 AD&D 名望等級之後的
Fighter +3、Cleric／Ranger／Thief +2、Magic-User +1。

`if 等級 > 1 then 骰數 := 1` 精準實作了「Ranger 一級 2d8、之後每級 1d8」。

⚠ **擲兩次取大**不是 AD&D 的規則，是這個引擎自己加的。remake 要照抄的話，
角色的血量期望值會比桌上規則高一截。

### `DS:3EA8h` 是另一張遮罩，不是 spec 811 那張

spec 811 的 `DS:3EA0h` 是 `02 10 08 40 40 01 04 20`（Fighter 與 Paladin 共用 `40h`）。
`DS:3EA8h` 是 `02 02 08 10 20 01 04 04`——**Cleric 與 Druid 共用 `02h`、
Thief 與 Monk 共用 `04h`**，分組方式不同。兩張表相鄰但用途不同，不要混用。

### ⚠ 過上限那條路是覆寫不是累加

`結果 := 3`（`mov`）對上另一條路的 `結果 := 結果 + a`（`add`）。
多職業角色若其中一個職業過了上限，**前面累加的骰值會被整個蓋掉**。
形狀上像是漏寫了 `+`，但兩平台一致，不是反組譯錯位。

## `overlay-17:046ECh` ↔ `overlay-17:04DB4h`（兩平台各 88 條）— 有效等級

`retf 4` ＝ 一個角色遠指標。**兩平台只差一條呼叫目標。**

```pascal
職業數 := 0;  總和 := 0;
for c := 0 to 7 do begin
    if 有號(角色^[109h + c]) <= 0 then continue;
    總和 := 總和 + 角色^[109h + c];
    inc(職業數);
    if (c = 4) or (c = 7) then inc(總和);          { Ranger／Monk 多算一級 }
end;

調整 := 有號(<本模組 45E9h>(角色));
if 調整 < 0 then begin
    if 有號(總和) <= 職業數 + abs(調整) then 總和 := 1
    else 總和 := (總和 + 調整) div 職業數;
end else
    總和 := (總和 + 調整) div 職業數;
結果 := 總和;
```

**有效等級 ＝ (各職業等級總和 ＋ 調整) div 職業數**，也就是**取平均**。
Ranger 與 Monk 在總和裡各多算一級。

⚠ **`職業數` ＝ 0 時 `idiv` 會除以零。**八個職業槽全部 ≤ 0 的角色（零級生物）
走到這裡就是除法例外。這一支自己沒有防護，呼叫端必須保證。

## 明確不宣稱

- 沒有宣稱 `本模組 45E9h(角色)` 回傳什麼（形狀上是等級調整，可正可負；
  被吸等級 `+0E7h` 是候選來源但沒有證據）。
- 沒有宣稱呼叫端傳給「升級擲 HD」的遮罩是怎麼算出來的。
- 沒有宣稱 `DS:3EB9h` 的值是「開始給固定值的等級」還是「最後一個擲骰的等級」——
  程式寫的是 `等級 >= 表值` 就給固定值，本規格照這個寫法記。
