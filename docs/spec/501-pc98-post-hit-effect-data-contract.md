# 第五百零一輪：PC-98 物理命中後效果 `4Fh` 的資料契約

狀態：`READY`
範圍：CoAB PC-98 `CHECKFX` 命中後 `4Fh` 的已閉合規則，以及將規則移入
Golden Box engine＋CoAB game pack 的資料分層。
主要原始規格：[`414-pc98-post-hit-effect-4f-runtime.md`](./414-pc98-post-hit-effect-4f-runtime.md)

## 1. 本輪結論

第 414 輪已用 PC-98 原始 bytes 與正常物理攻擊邊界閉合以下行為，本輪不重新
猜測其語意，而是把已證明的欄位交給作品中立的 `combat/posthit` 契約：

| 欄位 | 已證明值 | 推論等級 |
| --- | --- | --- |
| 原始效果 | `MON*SPC effect 4Fh`（JSON `effect_kind=79`） | `exact` |
| 觸發條件 | 物理攻擊命中，且目標在物理傷害後仍存活 | `exact` |
| 目標 | 沿用同一個攻擊目標，不重新選擇 | `exact` |
| 攻擊槽 | PC-98 傳入的第一、第二槽；一基底的 `attackSlot=1..2` | `exact` |
| 傷害骰 | `2d10` | `exact` |
| 傷害旗標 | Fire `01h`＋Magic `08h`，合計 `09h` | `exact` |
| 保護順序 | 先消耗兩顆 d10，再進入火焰／傷害保護判定 | `exact` |
| 物理擊殺、未啟用效果、第三槽以上 | 不 dispatch `4Fh` | `exact` |

這些值不是共用引擎的 CoAB 常數：`effect_kind`、槽範圍、骰池與旗標均由
`combat_post_hit_rules` 宣告；引擎只驗證形狀、解析並在命中後套用。

## 2. 原始證據

本輪使用的 PC-98 輸入與雜湊如下；位址基準保持分開，不把 overlay local offset
當成 resident address 或 file offset：

| 輸入 | SHA-256 | 用途 |
| --- | --- | --- |
| `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | PC-98 主執行檔 |
| `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a9` | overlay 容器 |
| overlay 12 code-only | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` | Fire／Magic 旗標與 `2d10` |
| overlay 13 code-only | `db1c03daaa984b94056e67d7b9e10fdc6bd3f2393f8a9e5ec720703f93835d4a` | 物理命中後 caller |
| overlay 23 code-only | `a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9` | `CHECKFX` dispatch |

關鍵連續證據：

- overlay 13 local `18A2h` 檢查物理傷害後目標仍存活；`18AAh` 讀取攻擊槽，
  `18AFh..18B0h` 傳入 `attackIndex+1`，`18B1h／18B4h` 傳入攻擊者，
  `18B7h` 呼叫 overlay 23 的 resident `CHECKFX`。
- overlay 23 local `0449h` 的 type 2 與 `0453h` 的 `mov al,4Fh`，以及
  local `04F6h` 的 type 3，共同支持第一、第二槽 dispatch `4Fh`；後續 type
  不納入本契約。
- overlay 12 local `19B3h` 同時設定 Fire `01h` 與 Magic `08h` 並執行兩顆
  d10；它使用目前攻擊目標的傷害邊界，不建立新的目標選擇。

反組譯報告採非破壞性外部註記，保留原始 local offset、bytes、輸入雜湊與 IDA
位址空間；本 spec 不以推測性 rename 取代原始定位。

## 3. Engine／game-pack 邊界

共用 engine 新增 `combat/posthit.Rule`：

- `ID` 是穩定資料識別；
- `EffectKind` 對應不透明的原始效果 kind；
- `MinAttackSlot／MaxAttackSlot` 是一基底槽範圍；
- `DamageDiceCount／DamageDiceSides／DamageMask` 保存已證明的效果 payload。

CoAB `gamepack/events/pit-of-moander.json` 目前宣告：

```json
{
  "id": "coab.monster_affect_4f.fire-magic-2d10-slots-1-2",
  "effect_kind": 79,
  "min_attack_slot": 1,
  "max_attack_slot": 2,
  "damage_dice_count": 2,
  "damage_dice_sides": 10,
  "damage_mask": 9
}
```

`Battle` 只把存活的 raw `MonsterAffect` kind 交給 engine resolver，再依解析出的
rule 擲骰與套用 `MonsterDamageAdjustment`。`PostHitRules` 是 game-pack 設定，
不進 save JSON；開戰與 active-combat restore 都重新掛入，避免存檔保存第二份
易漂移規則。

`AttackEffectResult` 保留原始 kind、旗標、原始骰值、實際傷害與 `Protected`，
不把命中後效果混入物理武器傷害，讓 UI、音效與後續動畫能逐事件接續。

## 4. 尚未閉合的邊界

以下內容刻意不由本輪升格：

- `4Fh` 的完整 DOS／PC-98 sprite、windup、impact、音效與 wall-clock timing；
- effect `4Fh` 是否還有未取樣的跨場景生命週期；
- effect `6Ah` 是否包住 `4Fh` 的 innate damage、其 caster-level 順序與
  多目標亂數 caller；目前是 `unknown`，不可直接套用魔法抗性規則；
- effect `84h` 的閃電 caller、完整目標與動畫；
- 其他 Gold Box 作品是否共用相同 kind、槽範圍或骰池。第二款作品提供獨立
  bytes／runtime 證據前，只能稱為 engine contract candidate。

因此本輪完成的是一個 exact 行為的資料分層與可重用執行邊界，不是完整戰鬥或
整作通關證明。

## 5. 驗證

- engine：`go test ./combat/posthit ./engine` 與 `go test ./...`；
- CoAB：game-pack resolver、Battle 物理命中／未命中／物理擊殺／免疫／第三槽、
  以及 active-combat save restore；
- 正式 gate 仍須在本輪收尾以 Docker／Xvfb 執行，並以 `coab-audit total=0`
  確認正式程式碼沒有新增未授權的可編輯中文或資料分層回退。
