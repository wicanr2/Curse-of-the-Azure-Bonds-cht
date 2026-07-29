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

弓箭方向與素材目前由 raw COMSPR＋consumer mapping 證明，詳見
[`comspr-projectile-audit.md`](../../research/comspr-projectile-audit.md)；
仍需再加入一筆公開影片或 DOSBox 的逐格弓箭時間碼，才能校準速度。

