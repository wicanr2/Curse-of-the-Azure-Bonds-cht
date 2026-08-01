# 第四百一十六輪：PC-98 怪物法術選敵、射程與牆面（READY）

## 範圍與結論

本輪把 effect `84h` 怪物 Lightning Bolt 從「所有存活對手任選」推進為
Gold Box 共用的 ranged-target boundary：先依雙方 footprint、加權射程與牆面
建立候選，再以同一戰鬥 PRNG 抽樣；不可見候選會移除後重抽，最多二十次。

PC-98 raw bytes／IDA 已關閉 `PICKTARGET` 的二十次迴圈、候選建立器與隨機
移除流程；DOS 重寫來源只用來交叉確認參數語意。原始 combatant-array 的同距
tie order、PC-98 runtime 的 invisibility affect trace，以及 effect `84h` 選敵
失敗後 `(0,0)` 動畫是否仍播放，仍未達 `exact`。

## 輸入與非破壞性

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbols、resident stubs、spell table | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | typed overlay resolution | `exact` |
| overlay 13 code | `db1c03daaa984b94056e67d7b9e10fdc6bd3f2393f8a9e5ec720703f93835d4a` | `CHECKTARGET`／`PICKTARGET` | `exact` |
| overlay 22 code | `c54729525d576c11d731d64a1b06ee2547b2562b73e3708a1beaafc535cabbe8` | `FIGSPELLRANGE`／effect `84h` | `exact` |
| overlay 24 code | `0b5d1bebb367414fac93287449a230a36355c075efdc30640a4994bc11d5cd5f` | 鄰近敵方候選建立器 | `exact` |
| overlay 32 code | `c7e2a2b1166676454a54a69330ea318dae8ef9a22e5c5a7ca6e6950ef3609d93` | footprint／地形射線 consumer | `exact bytes` |

`scripts/ida/pc98_spell_targeting_audit.idc` 只讀取原 overlays 的工作副本，
輸出保留 overlay-local 位址及 bytes；未 rename 原始 symbol，也未修改基準
`.i64`。typed resolver 同時保留 resident stub 與 handler-local 位址空間。

## 原始選敵鏈

Borland symbols 提供：

```text
CHECKTARGET  00B8:11AF
PICKTARGET   00B8:3D7F
FIGSPELLRANGE 0117:0DCB
GETSPELLTARGETS 0117:112C
```

overlay 13 `PICKTARGET 3D7Fh` 的關鍵順序：

1. `3D91h..3DF2h` 讀取既有 action target；同隊、已離場或
   `CHECKTARGET 11AFh` 失敗時清空。
2. `3E39h C6 46 FD 14` 把嘗試上限設為 `14h`，即二十次。
3. `3E3Dh..3E4Bh` 以施法者 far pointer 與 range 呼叫
   `014A:00C0h`。typed resolver exact 得到 overlay 24 entry 32、handler
   `285Bh`。
4. `3E6Dh..3EA9h` 依候選數擲骰，經候選索引取回 combatant far pointer。
5. `3EBDh..3ECDh` 再呼叫 `CHECKTARGET`；失敗時 `3EECh..3F37h` 從候選表
   移除並重抽，成功則保存 action target。

overlay 24 `285Bh` 由施法者位置、shape／size、最大 range 與 opposing-team
predicate 建立 `DS:9F2E` 三 byte候選表，最後投影成 `DS:A820` 的 combatant
indices。其 `0184:003Eh` resident call 經 resolver 落到 overlay 32 entry 6、
handler `012Eh`；該 routine 逐 footprint cell 建立地圖射線並讀 combat
background tile。與固定來源 `simeonpilgrim/coab@9dc46f1d` 的
`BuildNearTargets → Rebuild_SortedCombatantList → canReachTarget` 交叉後，
可把「對方、仍在場、任一 footprint cell 可達、range 內、牆未阻擋」列為
`strong inference`；在 PC-98 tile-height 欄位與排序 comparator 完整命名前，
不升格為全部 `exact`。

## Lightning Bolt 射程

MZ header `e_cparhdr=009Ah`，DS symbol segment 為 `0C29h`。因此 spell
`33h` record 的檔案位置為：

```text
09A0h + 0C29h*10h + 61B4h + 33h*10h = 13114h
02 03 04 01 00 00 08 00 02 04 00 01 03 06 01 00
```

`FIGSPELLRANGE 0DCBh` exact 讀 record `+2` fixed range `4`、`+3`
per-level range `1`，並乘上 caster-level helper。獨立重寫顯示天生效果施法者
fallback level 是 6，故 `4 + 1×6 = 10` 為 `strong inference`；effect `84h`
後續射線 helper 又直接 push `0Ah`，獨立證明本效果的實際 line budget 是 10。

原始 `canReachTarget` 使用 `steps <= range*2+1`；交叉來源的 stepping path
與既有 Lightning Bolt bytes 共同支持 cardinal 2、diagonal 3 的加權距離。
remake 因此以 `10*2+1=21` 為候選上限，並對攻守雙方每個 footprint cell
取最短且未穿牆的射線。

## 無目標 continuation

effect `84h` 在 `6323h..632Eh` 呼叫 spell-target delegate，把 bool 寫到
`A9D8h`，但 `6332h` 起完全不檢查該值，仍移除隱形、載入 projectile、擲兩份
`16d6`，最後清 actions。這項「失敗仍消耗行動」是 `exact`。

固定 DOS 重寫的 target routine 在候選為空時把 target position 清為 `(0,0)`；
PC-98 尚缺 runtime trace，故座標只列 `strong inference`。remake 保存此
continuation：若 `(0,0)` 是有效 combat cell，仍向該格執行 line effect；若
frontend terrain 明確判定無效，則不虛構可走射線，但仍顯示資料化繁中訊息、
發出閃電聲並消耗怪物回合，不回退成物理攻擊。

## remake contract 與驗證

- `TargetSelectionOptions` 是作品中立介面，注入 max range、`LineTerrain` 與
  可選的 status visibility callback，不含提朗瑟克斯或 CoAB 劇情名稱。
- 候選依最短加權距離排序；同距暫以 stable fighter ID tie-break，原 combatant
  array order 尚未投影，所以 deterministic tie fidelity 仍未完成。
- `SelectRangedCombatTarget` 保留二十次上限與「不可見就移除再抽」的 PRNG
  consumption；候選為空是正常 `found=false`，不是 parser error。
- effect `84h` 在一般 `SelectCombatTarget` 前使用 range 10 selector，不再先
  消耗一次錯誤的全體存活者亂數。
- core tests 覆蓋超距、牆面、2×2／2×1 footprint、不可見重抽與全不可見。
- State regression 覆蓋正常閃電命中，以及唯一對手被牆阻擋時不做物理攻擊、
  不扣 HP、仍消耗 effect `84h` 回合。

本規格只將上述 bounded target selection 標為 `READY`；status visibility 與
blink／動物視覺 combat consumer 已由 spec 417／418 接續。完整怪物法術 AI、
effect 生命週期、原始 tie order、終戰牆面與逐幀演出仍是後續工作。
