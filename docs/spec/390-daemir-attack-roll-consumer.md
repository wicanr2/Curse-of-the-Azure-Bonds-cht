# 規格 390：黛米爾祝福／詛咒的命中修正 consumer

狀態：`READY`

完成範圍：ECL `4CBBh` writer 到 `7F71h` 戰鬥 work cell 的投影、PC-98
命中公式 consumer、signed `+2／-2`、作品中立 game-pack binding、
Battle 暫存生命週期，以及正常玩家路徑回歸。

未完成範圍：DOS `ATTEMPTTOHIT` 的逐指令交叉版本比對、所有武器類別與
特殊攻擊 caller 列表、完整 AD&D THAC0／AC fidelity，以及黛米爾效果離開
Myth Drannor 後由哪個 ECL writer 清除。這些邊界不影響本規格已證明的
戰鬥入口投影與 attack-roll 加減值。

## 1. 輸入身分

| 輸入 | SHA-256 |
|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` |
| TPOV overlay 05 `POSTCOM` code | `4136fc40ec5647cac6ef7cbc762ecb8a2d4ada045020bff4646290d267864cdf` |
| TPOV overlay 23 `EFFECTS` code | `a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9` |
| TPOV overlay 24 `GENERIC` code | `0b5d1bebb367414fac93287449a230a36355c075efdc30640a4994bc11d5cd5f` |

原始 DOS `ECL6.DAX` 仍由 `curseoftheazurebonds.zip` 唯讀載入，不提交
解碼後商業資料。

反組譯使用指定 IDA Pro 9.4 Docker 隔離工具鏈；可重現 IDC 是
`scripts/ida/pc98_daemir_hit_consumer_audit.idc`。Borland table layout
依已保存的官方《Borland Open Architecture》交叉驗證。

## 2. 先排除的錯誤推論

### 2.1 `00E9:4CBB` 只是數值巧合

Borland symbol `CHARCONBONUS 00E9:4CBB` 的 `4CBB` 是 function
segment:offset，與 ECL work address `4CBBh` 不在同一位址空間，不能當成
consumer。

### 2.2 `VARLIST+06E0/+06E2` 不是 PARTY record offset

Borland symbols／types：

```text
VARLIST      0C29:7F09 type=705
VARLISTBASE  constant 7C00h
VARLISTEND   constant 7FFFh
type 705 VARLISTPTR size=0004h -> type 861
type 861 VARLISTTYPE id=1Ch size=0800h -> element type 10
type 10 size=0002h
```

`VARLISTTYPE` 是 `7C00h..7FFFh` 共 1024 格、每格 2 bytes 的 Pascal
array。因此：

```text
VARLIST + 06E0h = work 7F70h
VARLIST + 06E2h = work 7F71h
```

它們不是 `4CBBh` 的直接平移；必須另找 ECL 投影。

`internal/borlanddebug` 現會解析舊式 8-byte type slots，並由
`cmd/pc98-symbol-audit -type-index 705/861` 輸出：

```text
type_table=0x1BDF5 member_table=0x1EDCD member_bytes=3205 data_pool=0x1FA52
type=705 id=0x16 name="VARLISTPTR" size=0x0004 detail=005D03
type=861 id=0x1C name="VARLISTTYPE" size=0x0800 detail=000A00
```

這也驗證 641 筆 member records 正好是 `641×5=3205` bytes，table span
沒有靠模糊搜尋猜起點。

## 3. ECL 的完整投影橋接

DOS `ECL6.DAX` block `40h` exact bytes：

```text
+0249: 09 01 BB 4C 01 71 7F
```

解碼：

```text
SAVE source=[4CBBh] destination=[7F71h]
```

相同投影還出現在：

| ECL block | payload offset |
|---|---:|
| `40h` | `+0249h` |
| `42h` | `+0563h`、`+0A44h` |
| `43h` | `+0400h`、`+13BEh` |

因此第 389 輪黛米爾事件寫入的：

- 祝福：`4CBBh=02h`
- 詛咒：`4CBBh=FEh`

會在戰鬥初始化時投影到玩家側 `7F71h`。real-image regression 對
`02h` 與 `FEh` 都執行原始 `RunEntry(1)`，並逐值驗證
`SAVE [4CBBh] → [7F71h]`。

## 4. PC-98 命中公式

Borland symbols：

```text
ATTEMPTTOHIT 013E:122C
ROLLTOHIT    0C29:A039
```

IDA 在 `EFFECTS` overlay 23 證明 `ATTEMPTTOHIT` 會依 combat side 從
`VARLIST+06E0h／+06E2h` 取一個 byte，接著：

```asm
mov al, es:[di+6E0h or 6E2h]
...
mov al, ds:ROLLTOHIT
cbw
add ax, actor_attack_component
add ax, signed_side_modifier
cmp ax, required_hit_value
jl  miss
```

`cbw` 對 low byte 作 signed extension，所以 `02h` 是 `+2`，`FEh` 是
`-2`，不是 unsigned 254。

`POSTCOM` overlay 05 的 `DOPOSTCOMBAT 1775h` 又證明戰後清除：

```asm
1968  les di, ds:VARLIST
196E  mov es:[di+6E0h], ax ; AX=0
1973  les di, ds:VARLIST
1979  mov es:[di+6E2h], ax ; AX=0
```

故 `7F70／7F71` 是每場戰鬥的暫存，不應寫回角色永久 AttackBonus。

## 5. Remake contract

獨立 engine 的 JSON schema 新增 `combat_modifiers`：

- `source_namespace`
- `source_address`
- `value_encoding`
- `target`
- `side`

CoAB game pack 宣告：

```json
[
  {
    "id": "coab.ecl.enemy-attack-roll",
    "source_namespace": "ecl_work",
    "source_address": 32624,
    "value_encoding": "signed_low_byte",
    "target": "attack_roll",
    "side": "enemy"
  },
  {
    "id": "coab.ecl.party-attack-roll",
    "source_namespace": "ecl_work",
    "source_address": 32625,
    "value_encoding": "signed_low_byte",
    "target": "attack_roll",
    "side": "party"
  }
]
```

Battle 以 side-scoped modifier 計算：

```text
d20 + Fighter.AttackBonus + battle.sideAttackRollModifier
```

自然 1／20 與 held target 的既有規則不變。modifier 不改寫 Fighter，
因此 `syncPartyFromBattle` 不會把暫存效果永久累加到下一場戰鬥。

## 6. 驗收

- engine：signed low-byte `02h→+2`、`FEh→-2`、重複 side／target binding
  拒絕測試。
- ECL：真實 `ECL6.DAX` 正、負值的 `4CBB→7F71` 投影。
- combat：party／enemy 各自 modifier、命中邊界與 Fighter 不變。
- 正常玩家路徑：Standing Stone → Myth Drannor → Burial Glen →
  黛米爾 ACCEPT → 回地城 → 下一次 ECL lifecycle → `7F71=02h` →
  真實 Battle 命中邊界。
- `git diff --check` 與正式 Docker 測試是提交 gate。

