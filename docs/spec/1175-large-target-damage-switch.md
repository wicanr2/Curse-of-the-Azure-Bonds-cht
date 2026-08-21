# 1175 — 大型目標換另一組傷害骰：`+0DEh` 的 **bit 7 單獨成立**

- 證據等級：`exact`（`overlay-13:15EFh`..`1666h` 逐條讀完；`+0DEh` 的分布逐隻
  取自六章 `MON*CHA.DAX`）
- 上游：spec 1174（派生值重算把小型三連寫進現值）、spec 1164（`+0DEh` ＝ `SIZE`）

## 攻擊結算當下的切換

```pascal
if 攻擊者^[151h] <> NIL then                        { ★ 有槽 0 的武器才切 }
    if (目標^[0DEh] > 80h) or ((目標^[0DEh] and 7) > 1) then begin
        類別記錄 := 類別表[攻擊者^[151h]^[2Eh]];     { 整筆 16 bytes 複製出來 }
        攻擊者^[19Eh] := 類別記錄[2];                { 大型骰數 }
        攻擊者^[1A0h] := 類別記錄[3];                { 大型面數 }
        攻擊者^[1A2h] := 攻擊者^[1A2h] − 小加值 ＋ 大加值;
    end;
```

⇒ 三件事：

1. **只有拿著武器才切**——天生攻擊不分大小。
2. 切的是類別表的 `+02h`／`+03h`／`+04h`（大型三連），對照重算時裝的
   `+09h`／`+0Ah`／`+0Bh`（小型三連，spec 1174）。
3. **加值是換掉不是相加**（減小加大，等價於直接用大型那一格）。

## ★★ `and 7` 判不出大型

六章 43 種怪物的 `+0DEh` 逐隻量下來只有七個值：

| `+0DEh` | 大型？ | 隻數 | 例子 |
|---|---|---:|---|
| `01h` | 否 | 39 | 所有人形（也包含全部可玩種族）|
| `02h` | 是 | 1 | `SALAMANDER` |
| `03h` | 是 | 2 | `FIGHTING DOG`、`HELL HOUND` |
| **`81h`** | **是** | **4** | **`BEHOLDER`、`BUGBEAR`、`GIANT SPIDER`、`PHASE SPIDER`** |
| `82h` | 是 | 9 | `ETTIN`、`OGRE`、`TROLL`、`TYRANTHRAXUS`… |
| `83h` | 是 | 9 | `CENTAUR`、`CROCODILE`、`GRIFFON`… |
| `84h` | 是 | 4 | `BLACK DRAGON`、`DRACOLICH`、`WYVERN`、`BIT O' MOANDER` |

`81h and 7 = 1` ⇒ **只看低 3 位會把眼魔、熊地精與兩種巨蛛判成小型**。
bit 7 是獨立的「傷害算大型」旗標，低 3 位是**佔格大小**——同一個位元組裝兩件事。

remake 先前只留 `data[0xDE] & 7`（`CombatSize`，給 `FootprintForSize` 用），
那個位元就消失了。現在 `Record.RawSize` 保留完整位元組。

## remake 這一輪接了什麼

- `Fighter` 帶**兩組**傷害三連加一個 `HasSlotZeroWeapon`；切換在
  `WeaponDamageAgainst(目標)` 一處，攻擊擲骰與加值都走它。
- 隊伍側 `FighterWithEquipment` 順手把大型那一組也投影出來。
- **怪物側接上了**（`monster.ProjectMonsterWeapon`）：開戰時掛完 `MON*ITM`
  就照 spec 1174 投影槽 0 的裝備中武器。這是 spec 1120 留的「規則那一側還沒接」。
- `ReplaceFighterEquipment`（自動換裝的重投影）也一起帶新欄位——少了它，
  換裝之後打大型目標用的還是上一把武器的骰。

## 明確不宣稱

- 沒有宣稱低 3 位的 `1`／`2`／`3`／`4` 對應多大的佔格（`FootprintForSize` 是
  remake 既有的讀法，本規格沒有驗它）。
- 沒有宣稱 `> 80h` 的邊界值 `80h` 本身出現在資料裡（沒有）。
- 沒有宣稱大型三連與小型三連在**遠程武器**上的差別要怎麼算——本支只讀了
  近戰攻擊那條結算路徑。
