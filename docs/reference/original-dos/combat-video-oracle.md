# CoAB DOS 戰鬥影片 oracle

狀態：2026-07-29 第一筆逐格證據

## 來源

- 平台：PC／DOS
- 影片：[Curse of the Azure Bonds (PC/DOS) PART-1, 1989, AD&D, SSI, TSR](https://www.youtube.com/watch?v=wwYsij1wDC4)
- 影片 ID：`wwYsij1wDC4`
- 長度：`08:00:50`
- 本地研究檔只放在 `/tmp`，不提交完整影片。

## `00:42:25.20–00:42:25.40`：目標型法術 projectile

這段戰鬥底部顯示 `SPELL: STINKING CLOUD`。施法者確認目標後，原版先播放
一枚青色四叉 projectile，再顯示 `creates a noxious cloud` 與雲霧格。
這證明 generic spell missile 是獨立的 travel phase，而且不是用命中爆點或
雲霧本身替代。

逐 100ms frame：

| 影片時間 | 畫面 |
|---|---|
| `00:42:25.20` | [projectile frame A](./combat-spell-projectile-422520.png) |
| `00:42:25.30` | [projectile frame B](./combat-spell-projectile-422530.png) |
| `00:42:25.40` | [projectile frame C](./combat-spell-projectile-422540.png) |

`00:42:25.50` projectile 已清除，約 `00:42:25.60` 才出現文字與 Stinking
Cloud tiles。以公開影片的 29.97fps encode 只能保守判定可見 travel 約
`0.3s`；不能把壓縮影片 frame boundary 當成 DOS `SysDelay` 的精準單位。

## `00:36:15.40–00:36:17.00`：Fireball 是單次飛行、多目標逐一結算

同一部影片的 Tilverton royal-guard 戰鬥清楚保留 `SPELL: FIREBALL`：

| 影片時間 | 畫面 | 判讀 |
|---|---|---|
| `00:36:15.40` | [青色 travel projectile](./combat-fireball-travel-003615400.png) | `0x05/0x85` generic spell missile 正飛向選定中心 |
| `00:36:16.00` | [第一名目標的紅白爆點](./combat-fireball-impact-003616000.png) | travel 清除後，`0x0A/0x8A` family 在目標格播放 |
| `00:36:17.00` | [爆點後接死亡骷髏](./combat-fireball-death-003617000.png) | 同一名 Royal Guard 顯示傷害、倒地與 dying |

後續數秒可看到同一顆 Fireball 的影響目標依序重複「紅白 impact → 傷害文字
→ 必要時死亡骷髏」，而不是在中心畫一次大型圓形爆炸後同步扣除所有 HP。
因此 remake 的 Fireball choreography 必須是：

```text
caster windup
  -> generic spell travel（一次）
  -> impact[0] / damage / optional death
  -> impact[1] / damage / optional death
  -> ...
  -> handoff
```

這段影片直接證明播放順序與素材外觀，但不能單靠畫面確定 saving throw、
傷害骰、影響半徑或目標排序公式。規則仍須以 raw binary／反組譯交叉驗證。
CoAB game pack 已加入 `fireball/travel` 與 `fireball/impact` 的原始 COMSPR
frame mapping；remake 的 `VisualEvent.Impacts` 也已實作一次 travel 後逐名
impact／optional death。正常玩家路徑會檢查並消耗 `0x2F` memorized slot，
以 tile cursor 指定中心；三張 remake checkpoint 位於：

- [`combat-timeline-fireball-travel.png`](../../screenshots/combat-timeline-fireball-travel.png)
- [`combat-timeline-fireball-impact-1.png`](../../screenshots/combat-timeline-fireball-impact-1.png)
- [`combat-timeline-fireball-impact-2.png`](../../screenshots/combat-timeline-fireball-impact-2.png)

## 與反組譯的交叉驗證

Reference decompile 的共同施法路徑在 resolver 前呼叫：

```text
load_missile_icons(0x12)
draw_missile_attack(0x1E, 4, target, caster)
```

combat icon `0x12` 經初始化 mapping 對應 `COMSPR 0x05/0x85`，四格順序為：

```text
0x05 -> flip(0x05) -> flip(0x85) -> 0x85
```

影片中的青色形狀、移動順序與本機原始 COMSPR PNG 相符。因此 remake 的
Magic Missile 與其他使用共同目標型施法入口的法術，travel 必須使用這組
原版四格素材；Magic Missile 傷害回饋另使用 `0x0A/0x8A` 的 target-local
四格爆點。

Fireball resolver `sub_5F782` 另交叉證實：`Rebuild_SortedCombatantList`
以目標中心、size 1、max range 2 收集敵我所有 combatant；傷害只擲一次
caster-level d6，`SpellEntry(0x2F)` 指定 Spell save、成功減半。remake
目前精確保留這些可證明部分；有牆地形的 path reachability 與同距離時原始
combatant-array／direction tie order 尚未完成。

弓箭方向與素材目前由 raw COMSPR＋consumer mapping 證明，詳見
[`comspr-projectile-audit.md`](../../research/comspr-projectile-audit.md)；
仍需再加入一筆公開影片或 DOSBox 的逐格弓箭時間碼，才能校準速度。

## Lightning Bolt：命中後繼續延伸

同一 DOS/EGA 影片 `07:40:22.50–25.60`：

```text
22.50  COMSPR 06/86 青白電弧移向指定 dark elf mage
22.70  電弧與目標重疊
23.10  COMSPR 0A/8A 紅白 damage impact + 電擊傷害文字
23.60  saving throw 文字
24.70  COMSPR 06/86 由命中格繼續向右
25.30  電弧抵達更遠格
```

關鍵幀：

- [`combat-lightning-target-travel-074022500.png`](combat-lightning-target-travel-074022500.png)
- [`combat-lightning-target-hit-074022700.png`](combat-lightning-target-hit-074022700.png)
- [`combat-lightning-damage-impact-074023100.png`](combat-lightning-damage-impact-074023100.png)
- [`combat-lightning-line-continue-074024700.png`](combat-lightning-line-continue-074024700.png)
- [`combat-lightning-line-end-074025300.png`](combat-lightning-line-end-074025300.png)

這段是 **video-backed** 的 travel→impact/damage/save→continue ordering。
牆面反彈未出現在該場戰鬥；目前只標為 executable/reference
**code-backed strong inference**。完整規則、證據邊界與 remake mapping 見
[`355-lightning-bolt-reflecting-line.md`](../../spec/355-lightning-bolt-reflecting-line.md)。

對應 remake checkpoints：

- [`combat-timeline-lightning-target-hit.png`](../../screenshots/combat-timeline-lightning-target-hit.png)
- [`combat-timeline-lightning-line-continue.png`](../../screenshots/combat-timeline-lightning-line-continue.png)
- [`combat-timeline-lightning-reflect.png`](../../screenshots/combat-timeline-lightning-reflect.png)
