# 第五百六十八輪：`ROLLDICE` 與原版亂數入口

狀態：`READY`（限 `ROLLDICE`／`ROLLDAMAGEDICE` 的計算式與亂數來源）。
日期：2026-08-14

## 結論先行

`ROLLDICE` 是全遊戲被呼叫最多的共用 routine（DOS 側 118 次 far call），
兩平台逐指令同構：

```text
ROLLDICE(count, sides):          ← Pascal 呼叫慣例，count 先推入
    total = 0
    if count <= 0: return 0
    for i = 1 .. count:
        total += Random(sides) + 1
    return total                 ← byte
```

`Random` 是 **Turbo Pascal RTL 的 `Random(n)`**（回傳 `0..n-1`），因此每顆骰
是 `1..sides`，整體就是標準的 `count d sides`。

`ROLLDAMAGEDICE(count, sides)` 先把 `count` 寫進 `DS:A032h`，再原樣委派給
`ROLLDICE`。那個全域的用途尚未解出（傷害顯示或減半判定的候選）。

| 平台 | `ROLLDICE` | `ROLLDAMAGEDICE` | 亂數 far call |
|---|---|---|---|
| PC-98 | `overlay-23:1368h` | `overlay-23:13B3h` | `0A65:1155` |
| DOS | `overlay-23:137Fh` | `overlay-23:13C6h`※ | `0A54:1105` |

※ DOS 的 `ROLLDAMAGEDICE` 位址由結構對應推得，尚未逐指令核對。

## 亂數來源的位址證明

DOS 的 far call `0A54:1105` 換算成 MZ load image 位移是
`0A54h × 16 + 1105h = B645h`。`START.EXE` 的 IDA database 中
`@Random$q4Word` 位於 linear `1B645h`，而該 image 載入基址是 `10000h`，
所以 image 位移同樣是 `B645h`——**兩者是同一個函式**。

這是第 566 輪把 `@Random`／`@Randomize` 排除在「不阻塞」之外的具體理由的
延伸：原版亂數不是抽象的 PRNG，它就是 Turbo Pascal 的 `Random`，而且有
明確的消費端。

## 對 remake 的意義

- 骰法沒有隱藏規則：沒有重骰、沒有上下限修正、沒有爆擊特例。任何「原版
  骰子不只是 NdS」的假設都要拿出證據。
- **回傳值是 byte**：`count × sides` 超過 255 會截斷。目前已知用法不會超過，
  但這是資料型別事實。
- `count <= 0` 回傳 0，不是 1。
- 要對齊原版的隨機序列，必須重現 Turbo Pascal `Random` 的 LCG 與
  `RandSeed` 更新方式，並確認 `Randomize` 的種子來源。**本輪未做這件事**；
  在此之前不得宣稱 remake 的遭遇／命中序列與原版一致。

## `EFFECTS`（overlay-23）單元的函式清單

Borland 符號直接給出這個單元的 AD&D 規則骨架，全部仍 `待解讀`，但可作為
下一批解讀的優先序：

`KILLDUDE`、`CALLEFFECT`、`SPELLOFF`、`CHECKFX`、`CHECKTERRAINFX`、
`TRYTOHIT`、`ATTEMPTTOHIT`、`MAKESAVE`、`ROLLDICE`、`ROLLDAMAGEDICE`、
`ADDEFFECT`、`LOSEDUDE`、`REMOVEINVIS`、`REMOVEFX`、`ROUTINGREMOVEFX`、
`CUREEFFECT`、`CONVERTSTRTOSPEC`、`CONVERTSPECTOSTR`、`DONEWSTRENGTH`、
`RECALCULATESTATS`、`PUTDAMAGE`。

`CONVERTSTRTOSPEC`／`CONVERTSPECTOSTR` 很可能是 AD&D 的 `18/xx` 特殊力量
轉換，但名稱不是證明，未讀之前不得寫進規則。

## 這份規格明確不宣稱

- Turbo Pascal `Random` 的演算法與種子行為。
- `DS:A032h` 的用途。
- `EFFECTS` 其餘 20 個 routine 的任何語意。
