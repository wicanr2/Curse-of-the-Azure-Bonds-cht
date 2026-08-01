# 第四百一十四輪：PC-98 命中後 `4Fh` 火焰效果 runtime（READY）

## 範圍與結論

本輪關閉提朗瑟克斯 `MON6SPC` 天生 effect `4Fh` 的主動攻擊邊界：前兩個
物理攻擊槽命中、且同一目標在物理傷害後仍存活時，原程式對該目標追加一次
Fire＋Magic 的 `2d10`，不重新選擇目標，也沒有 saving throw。

這是單一特殊能力的完整數值／排程 vertical slice；尚未證明原版火焰動畫、
專用文字、聲音及 wall-clock timing，也不代表 `84h` 怪物 Lightning Bolt
已完成。

## 輸入與非破壞性 IDA 證據

原始 PC-98 binary／overlay 與基準 database 均保持唯讀。IDA Pro 9.4 只在
Docker 內分析 code-only 副本；語意以外部報告附加，沒有 rename 或覆寫原始
位址。

| 輸入 | SHA-256 |
|---|---|
| `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` |
| `PC98-GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` |
| code-only overlay 12 | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` |
| code-only overlay 13 | `db1c03daaa984b94056e67d7b9e10fdc6bd3f2393f8a9e5ec720703f93835d4a` |
| code-only overlay 23 | `a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9` |

### 攻擊 caller

PC-98 overlay 13（Borland module `COMBAT`）的 post-hit 流程：

| local | 原始指令／資料流 | 語意 | 等級 |
|---:|---|---|---|
| `18A2h` | `cmp byte ptr es:[di+197h],0` | 物理傷害後目標仍在戰鬥才續行 | `exact` |
| `18AAh` | `mov al,[bp-15h]` | 讀取目前 attack slot index | `exact` |
| `18AFh–18B0h` | `inc ax; push ax` | 傳入 `attackIndex+1` 的 CHECKFX type | `exact` |
| `18B1h／18B4h` | push attacker far pointer | effect owner 是目前攻擊者 | `exact` |
| `18B7h` | `call far 013E:0034` | 呼叫 overlay 23 `CHECKFX` resident stub | `exact` |

`pc98-ovr-audit -resolve-code 23:03FE` 反查得到 entry 4、resident stub
`0034h`、handler local `03FEh`；segment、stub 與 overlay-local offset 分開
保存，不以相同十六進位值猜測。

### CHECKFX 與 handler

| 位址空間／local | 證據 | 結論 | 等級 |
|---|---|---|---|
| overlay 23 `0449h` | `cmp al,2` | 第一物理攻擊槽的 type 2 分支 | `exact` |
| overlay 23 `0453h` | `mov al,4Fh` | type 2 dispatch effect `4Fh` | `exact` |
| overlay 23 `04F6h` | `mov al,4Fh` | type 3 也 dispatch effect `4Fh` | `exact` |
| overlay 12 `19B3h` | Fire `01h`＋Magic `08h`、`2d10`、target `+18Eh` 座標 | 對既有 attack target 追加 2d10 Fire＋Magic | `exact` |

第二獨立 oracle 為 `simeonpilgrim/coab` commit
`9dc46f1d5911710fb2fcb7a9c0ec0ef74264d17c`：其攻擊迴圈同樣在成功命中且
目標存活後呼叫 `CheckAffectsEffect(attacker, attackIdx+1)`，type 2／3 都含
`fireAttack_2d10`，handler 使用既有 `actions.target`。它只作交叉驗證，主要
結論仍由上述原始 bytes／IDA 位址支撐。

## 作品中立 runtime 契約

- `Fighter.MonsterPostHitAffects(slot)` 只投影 operational
  `Active || Innate` 的 raw effect，不依怪物名稱、sprite 或劇情判斷。
- slot 1／2 dispatch `4Fh`；slot 3 以後不 dispatch。`Attack()` 是 slot 1，
  `AttackSequence()` 保存各 slot 編號。
- 只有物理 `Hit=true` 且物理傷害後 `TargetHP>0` 才觸發；物理擊殺、未命中、
  inactive effect 都不觸發。
- target 固定沿用該次物理攻擊目標；不重新抽選敵人。
- 每次先消耗兩顆 d10，再進入元素防護。即使 `70h` 防火使實際傷害為零，
  `RolledDamage` 仍保存且 PRNG 已前進，避免 continuation 漂移。
- `AttackEffectResult` 分開保存 raw `Kind`、`DamageFlags`、`RolledDamage`、
  實際 `Damage`、最終 HP 與 `Protected`，不把追加傷害混進武器傷害。
- 本輪未把 `6Ah` 魔法抗性套進此 damage boundary；目前沒有足夠 primary
  evidence 證明 innate 4F 的 caster-level／抗性呼叫時序，故保持未完成。

## 繁中與驗證

正式 `assets/locale/zh-TW.json` 新增三個 stable IDs，State 只依結構化 effect
結果選擇訊息，不寫死提朗瑟克斯名稱：

- `combat_hit_with_fire`
- `combat_hit_fire_protected`
- `combat_multi_attack_with_fire`

驗證涵蓋：

- operational／inactive、miss、物理擊殺與 `70h` 防火邊界；
- 三次攻擊只在前兩 slot dispatch；
- 正式繁中 JSON 的單擊、受保護及多擊訊息；
- Standing Stone 起始正常長玩家路徑取得真實 `MON6SPC` Tyranthraxus，驗證
  其 `4Fh` 投影及同一 fighter 的 2d10 post-hit boundary，原 37 人終戰仍可
  完成 `PROGRAM 8`。

## 尚未完成

- 原版 4F 火焰 sprite、逐幀動畫、專用文字、sound cue 與 wall-clock timing。
- 自由攻擊、轉移目標及第三個以上攻擊槽的完整 DOS runtime 動態 trace。
- `6Ah` 是否及如何作用於 innate 4F damage。
- `84h` 怪物 Lightning Bolt 的頻率、AI、目標、地形反彈、動畫與聲音。
