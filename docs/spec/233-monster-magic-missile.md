# 第二百三十三輪：MON*CHA monster Magic Missile（READY）

## Reference evidence

CoAB reference `Classes/PoolRadPlayer.cs` 定義 `field_33` 為 0x38 個 spell-list slot；
`engine/ovr017.cs` 將其複製到 `Player.spellBook`。同一 routine 將 `field_B5..B7`
複製到 magic-user spell cast count。`engine/ovr023.cs` 的 `SpellNames[0x0F]` 是
`Magic Missile`，spell table 將其列為一級 combat spell；`engine/ovr010.cs` 的
monster turn 會先嘗試可用 spell，再進入 physical attack／guarding。

## 本輪 contract

- `monster.Parse` 保存 `0x33..0x6A` 的非零 spell IDs 與 `0xB5..0xB7` 的 magic-user
  level-use counts；不把未知 spell ID 映射成猜測的名稱或效果。
- `combat.Fighter` 保存上述 raw spell slots／uses。
- 僅對已核對的 spell ID `0x0F` 接入 `Battle.CastMonsterMagicMissile`：一級、單枚
  missile、每枚 2–5 damage、無 saving throw；成功施放會 atomic 消耗 level-1 use。
- State enemy turn 在 physical attack 前嘗試此 spell；無可用 spell 或施法失敗才沿用
  target selection 與 physical attack path。

## 明確 boundary

其他 monster spell、`MON*SPC` effect application、AI priority／visibility／range、
spell saving throw、wand／item casting、guarding 與完整 monster turn 仍未完成；本輪
不把 raw spell-list preservation 宣稱成完整法術引擎。

## Verification

- synthetic MON*CHA bytes regression：spell ID `0x0F` 與 `field_B5[0]` 會進入 Fighter。
- combat regression：Magic Missile 產生 1 枚、2–5 damage 並消耗一次 use。
- State regression：enemy turn 顯示繁中「魔法飛彈」、扣除 party HP 後回到下一個 party turn。
