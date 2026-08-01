# 第四百一十三輪：提朗瑟克斯防火／防電 runtime（READY）

## 範圍與結論

本輪把第 412 輪已達 `exact` 的 `70h` 防火與 `87h` 防電 handler 接入
Fireball／Lightning Bolt 的實際玩家施法路徑。它不包含 `4Fh` 天生火焰
攻擊、`84h` 怪物閃電術 AI，也不表示所有 spell／item damage boundary 完成。

## 證據鏈

| 證據 | 結論 | 等級 |
|---|---|---|
| overlay 12 local `249Dh` | damage flags 含 Fire `01h` 時呼叫 `Protected(0)` | `exact` |
| overlay 12 local `2B87h` | damage flags 含 Electricity `04h` 時呼叫 `Protected(0)` | `exact` |
| `Protected(0)` local `001Bh` | 清除 pending damage／affect bytes | `exact` |
| typed TPOV resolver | `70h → 249Dh`、`87h → 2B87h` | `exact` |
| Standing Stone→終戰真實 `MON6SPC` | 提朗瑟克斯同時帶有 operational `70h／87h` | `exact` |

## 作品中立實作契約

- `DamageFlagFire=01h`、`DamageFlagElectricity=04h`、`DamageFlagMagic=08h`
  保存原始 bit，不以法術名稱判斷免疫。
- `Fighter.MonsterProtectedFromDamage(flags)` 只接受 operational
  `Active || Innate` effects；inactive 合成反例不得生效。
- `70h` 只消費 Fire，`87h` 只消費 Electricity；`6Ah` 隨機魔法抗性仍是
  獨立 handler，不能混成同一個布林免疫。
- Fireball 送入 Fire＋Magic；Lightning Bolt 經
  `ReflectingLineOptions.DamageFlags` 送入 Electricity＋Magic。後者使 line
  renderer 維持作品中立，不在共用 core 以 spell ID 猜 damage type。
- 保護成功時 impact 仍存在於 travel／impact timeline，但 `Damage=0`、HP
  不變並標記 `Protected=true`；施法格、路徑、音效 handoff 與回合仍進行。
- 現行 Fireball／Lightning Bolt 先完成既有 saving throw，再由 pre-damage
  protection 清除 applied damage。功能結果有 exact handler 支持；save 與
  protection 的原版 RNG draw 順序仍待 DOS runtime trace，不能宣稱完整亂數
  fidelity。

## 繁中呈現

正式 `zh-TW.json` 新增 stable IDs：

- `combat_fireball_protected`
- `combat_lightning_bolt_protected`

State 由同一份正式 JSON 產生防護次數摘要；visual impact 保存
`Protected`，避免 line renderer 把受保護目標誤報成「受到 0 點電擊傷害」。
測試於執行時載入正式 locale，不複製顯示文字常數。

## 驗證

- combat core：Fireball `70h`、Lightning Bolt `87h`、inactive effect 反例。
- 正常玩家 UI：記憶法術、tile cursor、slot 消耗、visual impact 與繁中 stable
  ID 訊息。
- 長路徑：Standing Stone 經 ECL／GEO／MON6 到 37 人終戰，取真實
  Tyranthraxus fighter 建立獨立 damage boundary，兩種元素傷害均為零，原長
  路徑仍完成 `PROGRAM 8`。

## 尚未完成

- 防護與 saving throw 的原版逐次 RNG／文字／sound cue 時序。
- Fireball terrain-aware range、Lightning Bolt tie order 與完整牆角反彈。
- `4Fh` 天生 2d10 fire attack、`84h` 怪物 Lightning Bolt scheduler／AI。
- `6Ah` 對 Fireball、Lightning Bolt 與其他 Magic damage 的多目標順序。
