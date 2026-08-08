# 第五百輪 PC-98 6Ah 魔法抗性資料契約與區域法術邊界

狀態：`READY`（可實作的有界規格，不代表完整法術或完整遊戲完成）

本輪把 PC-98 `MON*SPC` effect `6Ah` 的已閉合魔法抗性資料，接到可重用
Golden Box engine 的 `combat/resistance` 契約與 CoAB game-pack JSON。火球術與
閃電束現在會在自己的存豁免交易後，以同一個資料驅動的抗性規則判定；這是
「已知規則的 bounded adapter」，不是對原版每一個亂數 draw 的逐指令聲明。

## 證據輸入與定位

分析在 Docker 內使用 IDA Pro 9.4 的唯讀副本；原始 executable、overlay、symbol
與基準資料庫未被改寫。輸入雜湊如下：

| 輸入 | SHA-256 | 位址基準／用途 |
|---|---|---|
| `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | PC-98 原始 resident executable |
| `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a9` | PC-98 overlay container |
| overlay 12 code-only report | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` | overlay-local code bytes |
| overlay 23 code-only report | `a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9` | overlay-local code bytes |

重要位置均保留原始 overlay-local 位址，不把它們改名成可疑的全域位址：

- overlay 12 local `2396h..23F3h` 是共用魔法抗性 routine；local `2404h`
  是 wrapper。
- wrapper 先以 `base + (11 - casterLevel) * 5` 建立比較值，再與 `1..100`
  的擲骰比較；原始 bytes 沒有顯示先行 clamp。
- overlay 23 local `2341h..234Eh` 的 `PUTEFFECT` 將目前 effect 寫入
  `A02Dh`，再呼叫 `CHECKFX` type `9`。
- overlay 23 local `0663h..069Bh` 的 type-9 handler list 明列
  `0Ah／14h／69h／6Ah／70h` 等 effect。
- overlay 12 的 effect wrapper 對 `6Ah` 傳入 base `15`。因此「effect
  `6Ah` → 15% 基準」與等級修正式為 `exact`；這不等同於已證明所有法術
  caller 都共用相同 draw 順序。

## 推論分級

### `exact`

- `6Ah` 的 resistance base 是 `15`。
- 機制公式是 `15 + (11 - casterLevel) * 5`；CoAB engine 的
  `combat/resistance.LevelAdjustedD100Chance` 保留未 clamp 表達式。
- 只有 `Active` 或 `Innate` 的 raw effect 才能進入 effect resolver；inactive
  的相同 byte 不會觸發抗性。
- `6Ah` 的規則不是角色名稱或怪物名稱的特例，而是 game-pack 宣告的
  `effect_kind=106`、`formula=level_adjusted_d100`、`base=15`。

### `strong inference`

- 在目前 CoAB 的 Fireball／Lightning Bolt bounded core 中，抗性檢查放在每個
  目標的 Spell save 後、元素防護前；抗性成功會清除該目標的傷害，並保留
  `Resisted` 與 `Protected` 的不同結果。
- `StartCombat` 與 active-combat restore 都從 JSON 解析並重掛規則；規則切片
  不進 save JSON。
- 抵抗結果會沿 `AreaSpellImpact` → `VisualImpactTarget` → localized visual
  message 傳遞，避免 renderer 只看傷害為零就猜原因。

### `hypothesis／unknown`

- DOS／PC-98 原版在 Fireball、Lightning Bolt 多目標或反射路徑中的完整
  save、`6Ah`、元素防護與 damage draw 順序尚未以 runtime trace 關閉。
- 尚未證明 innate `4Fh` 物理命中後 `2d10` Fire effect 是否走同一個 `6Ah`
  wrapper；本輪不套用。
- effect `84h` 怪物閃電、Cloudkill、Stinking Cloud 與其他魔法／持續效果的
  `6Ah` caller 尚未由本規格擴大覆蓋。

## Engine／game-pack 契約

Engine 提供：

- `combat/resistance.Rule`：`ID`、raw `EffectKind`、`Formula`、`Base`。
- `Resolve(activeKinds, rules)`：按 game-pack 宣告順序找出 active rule。
- `LevelAdjustedD100Chance`：只實作已證實的公式，不知道作品欄位布局。
- `engine.Pack.CombatMagicResistanceRules`、JSON schema 與 validation；同一
  effect kind 不可在一個 pack 重複宣告，base 限定 `1..100`。

CoAB pack 宣告：

```json
{
  "id": "coab.monster_affect_6a.magic-resistance-15",
  "effect_kind": 106,
  "formula": "level_adjusted_d100",
  "base": 15
}
```

Battle 只保存 immutable rule slice 的 runtime projection；save／load 後由
State 再次從 pack resolve。這讓後續 SSI Gold Box 遊戲能宣告不同 effect 或
base，而不用複製 CoAB 的 `0x6A` 常數。

## 驗證

- engine：`combat/resistance` 與 `engine` package tests。
- CoAB：`gamepack`、`internal/combat`、`internal/game` focused Docker tests。
- 核心抽樣涵蓋：Sleep／Magic Missile 的既有 `6Ah` 行為、Fireball 與
  Lightning Bolt 的抵抗／未抵抗兩種 deterministic seed、inactive effect、
  save restore 的規則重掛，以及抵抗演出訊息的 stable locale IDs。
- 本輪不能用上述測試宣稱完整戰鬥、完整魔法抗性或完整通關；下一步仍需以
  DOSBox／PC-98 runtime trace 關閉 caller draw order，再逐項補齊其他 spell
  effect 與原版動態演出。
