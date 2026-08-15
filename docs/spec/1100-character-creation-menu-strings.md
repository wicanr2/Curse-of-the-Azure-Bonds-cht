# 1100 — 建角四段選單的原作字串：種族 0 是 Monster，陣營順序確認為九宮格

- 狀態：`READY`
- 證據等級：`exact`（位元組取自 `workplace/re-sweep/dos/dseg/dos-dseg-dseg.bin`，
  四組都是等間距的 Turbo Pascal 短字串陣列，每一筆都完整解出且無例外）
- 產物：[`../audit/resident-data-tables.md`](../audit/resident-data-tables.md)
  （由 `scripts/dseg_tables.py --write` 重生）
- 對應 spec 1093 §一 的四個選單；把 spec 1099 §四 與建角實作的兩個推論升成 exact

## 四組字串

| 用途 | 基底 | 每筆 | 筆數 | 對應欄位 |
|---|---|---|---|---|
| 職業組合名稱 | `DS:0CB6h` | `1Bh`（`string[26]`） | 17 | `角色^[75h]` |
| 種族名稱 | `DS:0E9Ch` | `0Ah`（`string[9]`） | 8 | `角色^[74h]` |
| 陣營名稱 | `DS:0EECh` | `11h`（`string[16]`） | 9 | `角色^[11Bh]` |
| 性別名稱 | `DS:0F85h` | `7`（`string[6]`） | 2 | `角色^[119h]` |

## ★★★ 一、種族索引 0 是 `Monster`

| 索引 | 名稱 |
|---|---|
| **0** | **`Monster`** |
| 1..7 | `Dwarf`／`Elf`／`Gnome`／`Half-Elf`／`Halfling`／`Half-Orc`／`Human` |

> ★★★ **這解釋了 spec 1099 §四 的現象**：種族屬性上下限、可選職業、起始年齡
> 三張表都沒有第 0 列，因為索引 0 是 **`Monster`——非玩家種族**，
> 不需要建角用的屬性限制、可選職業或起始年齡。
> ⇒ 先前寫「`0` 不是合法值」只說對了一半：它是合法的種族編號，
> 只是不出現在建角選單裡。
> ★ 1..7 的順序與 spec 1084／884 由 AD&D 等級上限反推的編號**完全一致**。

## ★★★ 二、陣營是九宮格，順序確認

| 索引 | 名稱 | | 索引 | 名稱 | | 索引 | 名稱 |
|---|---|---|---|---|---|---|---|
| 0 | `Lawful Good` | | 3 | `Neutral Good` | | 6 | `Chaotic Good` |
| 1 | `Lawful Neutral` | | 4 | `True Neutral` | | 7 | `Chaotic Neutral` |
| **2** | **`Lawful Evil`** | | **5** | **`Neutral Evil`** | | **8** | **`Chaotic Evil`** |

> ★★★ **`陣營 ＝ 守序軸 × 3 ＋ 善惡軸`**（善惡軸 0 善／1 中立／2 惡）。
> `{2, 5, 8}` 正好是三個 `Evil`——與 spec 1066 的
> 「`(陣營 ＋ 1) mod 3 = 0` ⇒ 邪惡」**完全吻合**。
> ⇒ spec 1066 的「沒有宣稱 `0`..`8` 各自對應哪一格」可以解除；
> 建角實作裡標為 `strong inference` 的九宮格推論升成 `exact`。
> ★ 索引 4 的原作用字是 **`True Neutral`**（不是 `Neutral`）。

## ★★ 三、職業組合 17 筆與 spec 1093 §二 一字不差

| 編號 | 名稱 | | 編號 | 名稱 |
|---|---|---|---|---|
| 0 | `Cleric` | | 8 | `Cleric/Fighter` |
| 1 | `Druid` | | 9 | `Cleric/Fighter/Magic-User` |
| 2 | `Fighter` | | 10 | `Cleric/Ranger` |
| 3 | `Paladin` | | 11 | `Cleric/Magic-User` |
| 4 | `Ranger` | | 12 | `Cleric/Thief` |
| 5 | `Magic-User` | | 13 | `Fighter/Magic-User` |
| 6 | `Thief` | | 14 | `Fighter/Thief` |
| 7 | `Monk` | | 15 | `Fighter/Magic-User/Thief` |
| | | | 16 | `Magic-User/Thief` |

> ★★ 這是**第三次**確認同一組編號：spec 1093 §二 從建角流程的 switch 讀出
> 職業槽與起始經驗值、spec 1099 §二 從 `DS:4172h` 讀出六屬性最低要求、
> 本規格從字串表讀出名稱——三個來源互相對上，沒有一筆例外。
> ★ 編號 1（`Druid`）與 7（`Monk`）有名稱、有屬性要求、有職業槽，
> 但**不在任何種族的可選職業清單裡**（spec 1099 §五）⇒ 建角選不到。

## 四、性別

| 索引 | 名稱 |
|---|---|
| 0 | `Male` |
| 1 | `Female` |

★ 與 spec 1086 的「種族屬性表每組排成（男, 女）兩格、索引是 `角色^[119h]`」一致。

## 對 remake 的意義

- 四段選單的選項清單與順序全部有 exact 佐證，不必再推論。
- `internal/game/creation_guided.go` 的陣營表註解可以從 `strong inference`
  改成 `exact`，並指向本規格。
- 種族選單要略過索引 0（`Monster`）——remake 已經只列 1..7，行為正確。
- 職業組合的顯示名對照（remake 的 `classComboKeys`）與原作 17 筆一一對應；
  其中 `Druid` 與 `Monk` 沒有對應的中文 key，因為建角選不到它們。

## 明確不宣稱

- 沒有宣稱這四組字串各自由哪一支函式讀取（本規格只解出資料本身）。
- 沒有宣稱 `Monster` 這個種族在遊戲中的實際用途。
- 沒有宣稱 `Druid`／`Monk` 為何有完整資料卻不可選。
- 沒有宣稱 PC-98 側對應字串的位置（本規格只取 DOS）。
