# 第五百零五輪：PC-98 Quick 目標 caller 與單次抽樣邊界

狀態：`READY`（限 caller／handler 邊界、`PICKTARGET` 單次抽樣的 bounded
adapter；不宣稱完整 candidate producer、範圍、同分 tie 或原版 RNG）

## 本輪結論

第 504 輪的 priority retry 控制流仍可保留，但本輪新增的 caller audit 修正了
它的適用範圍：overlay 09 local `04CCh` 是 TPOV entry 4 的公開 handler，然而
在該 overlay 內唯一找到的直接 near caller 是 `0164h`。`0164h` 位於一個較大的
通用戰鬥動作分派函式中；它先嘗試其他 action handler，`04CCh` 回傳非零後才把
同一組 `[bp+6]／[bp+8]` 傳給另一個 `4A00:1568h` helper。這證明 `04CCh` 是
可重用的 target/action fallback 邊界，不能反推「所有 Quick 法術都直接從同一
入口進入」。

overlay 13 的 `PICKTARGET` 則提供另一個可閉合的形狀：先檢查 action 既有 target，
再由作品端建立候選，依候選數做一次抽樣，驗證 `CHECKTARGET`；失敗候選會被移除
後重抽，最多 20 次。這足以讓 remake 把「已由作品端投影的候選，依 legacy
object order 做一次抽樣」變成 engine 契約，先修正 Quick Magic Missile 目前
依字典序抽選的偏差；但不能把該候選表順序冒稱為完整 PC-98 pointer chain。

## 非破壞性輸入與重現資訊

| 輸入／產物 | SHA-256 | 用途 | 推論等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | resident symbols／TPOV control | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay chain／fixups | `exact` |
| overlay 09 code | `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20` | Quick handler／caller | `exact bytes` |
| overlay 13 code | `db1c03daaa984b94056e67d7b9e10fdc6bd3f2393f8a9e5ec720703f93835d4a` | `PICKTARGET`／`CASTCOMBATSPELL` | `exact bytes` |
| `pc98_quick_target_xref_audit.idc` | `f4334442ad894c0d91485f35bfad4fb72b649ac43e698560b069c03ec03b3bcc` | IDA caller／連續指令 audit | `exact procedure` |
| overlay 09 report | `f1bf551dd26a7e50510c32c8260e1266b91f0f8da99dd8885fa4dc3dd46df452` | 本輪 caller／helper report | `exact bytes／control flow` |
| `pc98_spell_targeting_audit.idc` | `5ffed28aae7000199ec8eefffcccd9ec2ea07aa36e861d1586e40358343b895a` | IDA `PICKTARGET` report | `exact procedure` |
| overlay 13 report | `b9f6a576dbef9fabde44a1ae12822dea92c1fae301ed25d3590295d690f548fb` | `PICKTARGET`／cast handoff | `exact bytes／control flow` |

IDA Pro 9.4 使用 image `ida-pro-9.4-ver2:uidfix-v1`，只對 Docker 內的 code-only
副本建立暫存 database；原始 executable、overlay、Borland symbol 與既有 `.i64`
均唯讀。所有數值都是 overlay-local，`exe=0x...` 僅是 `pc98-ovr-audit` 的
resident file/effective resolver 輸出，不能與 overlay-local 數值混讀。

## 已閉合的 caller／handler 證據

### overlay 09 Quick helper

- `04CCh..0624h` 連續 bytes：初始化 priority `7`，呼叫 `3E01:142Dh` 取得
  `1..7` 範圍，沿 action `+14Eh` 指向的候選與 `+52h` next pointer 走訪，
  在 `05D0h` 呼叫 `03D3h` suitability，最後由 `0615h` 呼叫 `00FA:0048h`
  handoff。這些是 `exact control flow`。
- `04CCh` 直接 xref 只有 `0164h`；raw near bytes 為 `E8 65 03`。
- `0627h` 直接 xref 只有 `01EFh`；raw near bytes 為 `E8 35 04`。
- `03D3h` 直接 xref 只有 `05D0h` 與 `0711h`；raw near bytes 分別為
  `E8 00 FE`、`E8 BF FC`。
- `pc98-ovr-audit -resolve-code 9:04CC` 對應 entry `4`、stub `0034h`、
  `exe=0x1134`；`9:03D3` 對應 entry `11`、stub `0057h`、`exe=0x1157`；
  `9:0627` 對應 entry `5`、stub `0039h`、`exe=0x1139`。`9:0164` 找不到
  handler，支持 `0164h` 是 overlay 內 caller 而非 TPOV entry。
- `0164h` 的連續 caller 先把 `[bp+8]`、`[bp+6]` 傳入其他 predicate；在
  `0164h` 呼叫 `04CCh` 後，只有非零結果才於 `016Bh` 呼叫 `4A00:1568h`。
  這是通用 action fallback 的 `exact` 形狀；參數正式名稱仍為 `unknown`。

### overlay 13 `PICKTARGET`

- `3D91h..3DF2h` 讀取 action `+0Ah/+0Ch` 的既有 far target；同隊、離場或
  `CHECKTARGET` 失敗會清除該 pointer。這是 `exact`，不是完整玩家／怪物
  target policy 的名稱證明。
- `3E39h` 將嘗試上限設為 `14h`。`3E3Dh..3E47h` 把 caster/action 與
  `+0Ah` 參數交給 `4A00:1560h` 建立候選資訊；`3E74h` 再依返回候選數做
  一次抽樣。
- `3E7Ch..3EAAh` 依抽樣索引取得候選 far pointer；`3EC9h` 呼叫
  `CHECKTARGET`。失敗時 `3EF3h` 移除該候選，`3F05h..3F25h` 依剩餘候選
  重抽；成功時 `3EE2h..3EE6h` 寫回 action target。
- 候選表的 producer、排序欄位、距離／visibility／範圍的完整 caller contract
  與同分 tie 尚未由這段 bytes 單獨閉合，均維持 `strong inference／unknown`。

## Remake 契約

- engine `combat/quicktarget.SelectOne` 只做一件事：複製並驗證候選、依
  game-pack 宣告的 `legacy_object_id` 排序，再執行一次 one-based random draw。
  它不建立候選、不解讀 spell、不處理牆面，也不宣稱重建 pointer chain。
- CoAB `State.selectQuickTargetOne` 使用與 Quick spell selector 共用的 Battle
  PRNG；Quick Magic Missile 改由 JSON `coab.pc98.quick-target-candidate-chain`
  投影的 `LegacyObjectID` 候選抽取，不再由 `Fighters()` 的 lexicographic ID
  順序抽取。手動 Magic Missile 游標與原有 damage／visual transaction 不改。
- priority Quick area／line／targeted cleric 仍使用 `Select` 的 bounded retry；
  本輪沒有把 `SelectOne` 套到那些需要 priority／geometry 的法術。

## 驗證

- engine `combat/quicktarget`：單次抽樣只呼叫一次 roll、使用 legacy order、
  非法結果 fail-closed。
- CoAB `internal/combat`／`internal/game`：Quick target integration、Magic
  Missile visual／slot path 與既有 priority selector 通過。
- 本輪仍不能宣稱完整 Quick AI、完整戰鬥或整作 remake 完成；下一個必要證據
  是候選 producer／target range／tie 的 runtime 或完整資料流，以及所有非
  Magic Missile 法術與敵方 target consumer 的逐項回歸。
