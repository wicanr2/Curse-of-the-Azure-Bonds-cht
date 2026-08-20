# 1139 — AC 與命中的刻度：原作存 `60 − 顯示值`，remake 統一用顯示刻度

- 證據等級：`exact`（`overlay-23:123Fh` 的 65 條逐條讀完；81 筆 `MON*CHA` 的
  欄位值量過；命中門檻用 FIRE KNIFE 對回原作的 18）
- 上游 spec 577（命中式）、spec 767（顯示式）、spec 1000（兩個護甲欄位）、
  spec 1137（攻擊結算挑哪一格）

## 原作怎麼存

| 欄位 | 存的是 | 畫面上顯示 |
|---|---|---|
| `+199h` 命中能力 | `60 − THAC0` | THAC0 ＝ `60 − 它`（spec 1001） |
| `+19Ah` 護甲值 | `60 − AC` | AC ＝ `60 − 它`（spec 767） |
| `+19Bh` 第二個護甲值 | 同上，背後攻擊用 | 不顯示 |

命中判定（`overlay-23:123Fh`，spec 577）：

```
d20 ＋ 攻擊者^[199h] ＋ v >= 目標的 AC 欄位
```

⇒ **儲存值越大越難打**。三段 AC（`+19Ah` ＞ `+19Bh` ＞ `+19Bh − 4`，spec 1137）
由難到易排下來，方向一致。

實測：FIRE KNIFE 的 `+199h ＝ 41`（THAC0 19）、`+19Ah ＝ 59`（AC 1），
自己打自己要 `d20 >= 18`。

## remake 用哪一邊

`Fighter.ArmorClass`／`AttackBonus` 一律是**畫面刻度**：AC 越小越難打、
命中加值越大越好，與 AD&D 1e 的習慣一致。兩邊同時換算之後命中式變成

```
d20 ＋ 命中加值 ＋ 目標 AC >= 20
```

——與原作是同一個函式，只是兩邊各減了一個常數。FIRE KNIFE 那組換過來是
`d20 ＋ 1 ＋ 1 >= 20`，一樣是 18（`TestHitThresholdMatchesTheOriginalNumbers`）。

### 換算點

原作搬過來的算式吃儲存刻度，跨這條邊界要換：

| 位置 | 換法 |
|---|---|
| `monster.Record.Fighter()` | `CombatArmorClass`（`60 − 儲存值`）、`CombatAttackBonus`（`儲存值 − 40`） |
| `party.CanHitECLDamageTarget` 的呼叫端 | `combat.StoredArmorClass` |
| `ecl.PartyContext` 的隊伍戰力（`AC > 60`、`命中 > 39` 才算分） | `StoredArmorClass`／`StoredAttackBonus` |
| `CHECKFX` 的護甲記錄寫入（`add_capped 2 cap 60`） | 換成儲存刻度套完再換回來 |

★ **第二個護甲欄位的換算由第一格決定**：`+19Bh` 系統性地比 `+19Ah` 小 2..8，
可以掉出 `CombatArmorClass` 的視窗（`50..70`），各判各的會變成一格換、一格沒換。

### 方向反過來時會看到什麼

反了不會有例外、不會有錯誤訊息，攻擊照樣結算，只有數字會變：

- 從 `MON*CHA` 匯入的怪物 AC 1..7 配命中加值 41..53 ⇒ `d20 ＋ 41 >= 1` 永遠命中；
- 敏捷越高的隊員 AC 越小，反而**越好打**；
- 護甲改善、防護法術、隱形都變成讓自己更好打；
- 背後攻擊變成比正面難打。

三條回歸測試各釘一段：`TestHitThresholdMatchesTheOriginalNumbers`（門檻對回原作
的 18）、`TestHigherDexterityIsHarderToHit`、`TestRearAttackIsEasierThanFacingTheAttacker`。

## `v` 是 ECL 的兩格命中修正

`bank1^[6E0h]`／`[6E2h]` 就是 ECL 位址 `7F70h`／`7F71h`
（bank 1 的映射是 `(位址 − 7C00h) × 2`，spec 1096）：**`7F70h` 是敵方側、
`7F71h` 是隊伍側的當場命中修正**，由劇情用 `SAVE` 寫進去（spec 390）。
原作依**攻擊者**的隊號（DOS `+197h`）挑哪一格，`0` 取 `+6E2h`（隊伍側）。

remake 這一格接在 `SetSideAttackRollModifier`，來源是 game pack 的
combat modifier 宣告（讀同一個 ECL 記憶格）。全 build 只有
`overlay-05:1736`（開局清 0）與 `overlay-23:123F`（讀）碰這兩格，
其餘都是 ECL 寫入。

## 明確不宣稱
- 沒有宣稱隊員的 AC 與命中加值**數值**與原作相同：`internal/party` 的
  `Fighter()` 仍是自己的近似（`10 − (敏捷 − 10) / 2`、`(力量 − 10) / 2`），
  本規格只處理刻度方向。
- 沒有宣稱 `+199h` 之外還有沒有別的命中修正欄位。
