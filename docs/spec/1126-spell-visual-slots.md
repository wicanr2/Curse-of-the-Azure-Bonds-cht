# 1126 — 法術演出：一個入口、三個圖組槽

- 證據等級：`exact`（載入迴圈 `overlay-11@05B3h`、演出入口 `overlay-24 entry#24`、
  逐格 blit `overlay-24 entry#23`、呼叫點由 `scripts/spell_visual_table.py`
  機械抽出）
- 對應工作項：`ENG-09` 覆蓋台帳的 `visual` 欄
- 上游 spec 1033（圖組槽載入）、spec 354（戰鬥動作時間軸）、spec 355（閃電）

## 原作沒有「一支法術一段演出」

戰鬥演出只有一個入口：

```pascal
<overlay-24 entry#24>(槽);
```

它對同一個槽連放四格：

```pascal
<overlay-23 entry#23>(槽, 0, 0, 0);
<overlay-23 entry#23>(槽, 0, 1, 1);
<overlay-23 entry#23>(槽, 1, 2, 1);
<overlay-23 entry#23>(槽, 1, 3, 0);
```

第二個參數在兩個遠指標之間切換（spec 1033：每個槽同時放「編號」與
「編號 ＋ 80h」兩份），第三個是格號。所以「這支法術長什麼樣」這個問題
**只剩一個變數：它把哪個槽號傳進去**。

## ★ 槽 ＝ COMSPR 區塊 ＋ 13

`overlay-11` 開場一次把整組載進來：

```pascal
for i := 0 to 11 do 載入圖組('COMSPR', i, i + 13);
載入圖組('COMSPR', 25, 25);
```

| 槽 | COMSPR 區塊 |
|---:|---:|
| 13..24 | 0..11 |
| 25 | 25 |

這條換算把既有規格裡的數字接了起來：spec 355 記的「icon `17h`」是**槽號**，
`17h = 23`，`23 − 13 = 10` ⇒ COMSPR `0Ah`／`8Ah`，正是那一組魔法命中圖。
spec 354 記的 Magic Missile「`05h`／`85h`」則是槽 18。

## 量到的：整個遊戲只用三個槽

| 槽 | COMSPR 區塊 | 用在哪 |
|---:|---:|---|
| 18 | 5 | **共用施法投射物**、定身／魅惑／妖火再播一次、兩支效果 handler |
| 19 | 6 | 閃電的電弧（`ov22@39F3h`／`3A10h`）|
| 23 | 10 | 魔法傷害命中（效果 handler `ov12@26CFh`）|

★ **共用施法投射物在 `CASTSPELL` 裡**（`ov22@1435h`），而且在分派到各支
handler **之前**就播。所以**每一支法術都有演出**——不是只有火球、閃電那幾支有。
這一點決定了 remake 這一側該怎麼宣告：不是七十幾筆各自的視覺資產，
而是一筆共用的加上少數例外。

逐支見 [`spell-visual-table.md`](../audit/spell-visual-table.md)。

## remake 這一側

- `gamepack/rules/spell-visuals.json`（由腳本產生）。
- game pack 的 `combat_visuals` 多一筆 `coab.spell-cast-projectile`，
  trigger 是 **`spell_cast_shared`**——它不對應任何單一 behavior，
  而是對每一支宣告過的法術都算數。
- `combatFinishSpell` 與保護法術那條路各排一次共用演出；已經有專屬演出
  排隊時就不排（火球、閃電、睡眠有自己的）。
- 覆蓋台帳的 `visual` 欄因此由 6 支 `observed` 變成 **73 支**。

## 明確不宣稱

- 沒有宣稱四格演出的每格延遲；`reference_delay` 沿用 spec 354 從原版影片
  量到的值，本輪沒有重新量。
- 沒有宣稱槽 18／19／23 以外的九個 COMSPR 區塊（0..4、7..9、11）用在哪裡；
  它們載進來了，但戰鬥法術這條路上沒有呼叫點。
- 沒有宣稱 `overlay-22` 尾端那七支演出常式（entry 109..115）屬於誰——
  **整份 overlay 集裡沒有任何呼叫點**，呼叫端在常駐段或 ECL 那一側。
  它們用的槽是 18／19／23，與法術那三個相同。
- 沒有宣稱效果 handler 那三處各自對應哪個效果碼；只知道位址與槽。
- 沒有宣稱 PC-98 側的槽號相同。
