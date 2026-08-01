# 第四百一十七輪：PC-98 目標可見性與 effect `18h／19h／47h`（READY）

## 範圍與結論

本輪關閉第 416 輪留下的 PC-98 status visibility callback：`CHECKTARGET`
先執行目標／觀察者的 effect dispatch，最後只依全域 target-hidden byte 判斷
是否可選。原始 handlers 證明兩種隱藏效果不能合併：

- operational `19h` 在觀察者沒有 operational `18h` 時，設定 target hidden
  並對 attack roll 套用 `-4`；有 `18h` 時兩者都不發生。
- operational `47h` 無條件設定 target hidden 並套用 `-4`，完全不檢查
  `18h`。
- 因此 `18h` 可以看穿 `19h`，不能看穿 `47h`。spec 235 與 spec 410
  將兩者一併略過的敘述已被 supersede。

這是 bounded target visibility 與物理 AC 投影，不代表 blink、動物視覺、
隱形解除時機、角色法術生命週期或所有 effect phase 已完成。

## 輸入與非破壞性分析

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | EFFPROCS resident slots、Borland symbols | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | typed TPOV resolution | `exact` |
| overlay 12 code | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` | effect `19h／47h` handlers | `exact` |
| overlay 13 code | `db1c03daaa984b94056e67d7b9e10fdc6bd3f2393f8a9e5ec720703f93835d4a` | `CHECKTARGET` | `exact` |

原始資料與基準 `.i64` 保持唯讀。`scripts/ida/pc98_visibility_target_audit.idc`
只在 Docker `/tmp` 的 code-only 副本執行，輸出原始 local address、raw bytes
與連續反組譯；沒有 rename 或把候選語意寫回 database。

## typed handler resolution

overlay 12 `INITEFFPROX` 對 effect table 的直接 writes 為：

```text
effect 18h slot A0A0/A0A2 -> resident 008B:02C3 -> overlay 12:2ECB
effect 19h slot A0A4/A0A6 -> resident 008B:00AC -> overlay 12:06F9
effect 47h slot A15C/A15E -> resident 008B:017E -> overlay 12:1713
```

後兩筆由 `pc98-ovr-audit -resolve-stub` 分別解析為 TPOV entry 28／70，
而不是把 table raw addend 直接當 handler。effect `18h` handler 本身為 no-op；
其 operational linked-list presence 由 `19h` handler 查詢。

## 原始資料流

`CHECKTARGET 00B8:11AF`（overlay 13 local `11AFh`）的 exact 順序：

1. null target 回 false，同一 fighter 回 true。
2. `11D9h` 清 `DS:A035h` target-hidden byte。
3. `11DEh..11E7h` 以 phase/type `1` dispatch target。
4. 尚未 hidden 時，`11F3h..1218h` 暫存並把對方 action target 指向觀察者，
   `121Ch..1225h` 以 phase/type `0` 再 dispatch，隨後恢復原 target pointer。
5. `123Fh..124Ch` 只在 `DS:A035h==0` 回 true。

effect `19h` handler（overlay 12 local `06F9h`）：

```text
06FF  cmp byte ptr [bp+6],0
0705  push ds:9596 / ds:9594
070D  mov al,18h
0715  call effect-query
071A  or al,al
071E  mov byte ptr ds:A035h,1   ; 僅 effect 18 不存在
0723  sub byte ptr ds:A039h,4
```

effect `47h` handler（overlay 12 local `1713h`）：

```text
1716  mov byte ptr ds:A035h,1
171B  sub byte ptr ds:A039h,4
```

`47h` 沒有 effect-query 或條件分支，所以「`18h` 也能看穿 `47h`」已被原始
bytes 直接推翻。候選語意名稱仍是附加註記；原位址與 raw effect ID 始終保留。

## remake 契約與驗證

- `Fighter.VisibleTo(observer)` 是作品中立 status visibility：inactive、
  non-innate records 不生效；`19h` 可被 observer `18h` 抵消，`47h` 不可。
- `MonsterAffectArmorClassBonusAgainst(attacker)` 使用同一差異，避免法術選敵
  與物理命中各自維護互相矛盾的規則。
- effect `84h` 的 `SelectRangedCombatTarget` callback 現呼叫正式 visibility；
  全部候選 hidden 時仍消耗怪物 action，不回退物理攻擊。
- core regression 覆蓋 `18h／19h／47h`、inactive／innate 與 AC 邊界。
- State regression 覆蓋一般 effect `84h` 怪物看不到 `19h`、帶 `18h` 可命中
  `19h`，以及帶 `18h` 仍看不到 `47h`。
- Standing Stone 起始正常長路徑載入真實 MON6 提朗瑟克斯六筆 innate
  effects，證明其中 `18h` 讓 `19h` 可見、但不能讓 `47h` 可見；原 37 人
  終戰與 `PROGRAM 8` 回歸仍必須通過。

## 尚未完成

- effect `19h／47h` 在角色法術、解除、回合到期及 save round-trip 的完整
  生命週期。
- effect `25h` blink、`45h` invisible-to-animals 等其他 visibility handlers。
- 原始 combatant-array 同距 tie order 與 effect `84h` 無目標 `(0,0)` 的
  DOS／PC-98 動態動畫。
- 終戰牆面反射、逐幀 timing、聲音順序及 effect `6Ah` 對 `84h` 的時序。
