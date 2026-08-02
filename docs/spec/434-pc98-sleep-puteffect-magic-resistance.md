# 第四百三十四輪：PC-98 睡眠術 PUTEFFECT 與魔法抗性

狀態：`READY`（限 ordered targets 後的 `PUTEFFECT` 結算、魔法抗性及
`EFFECTREC 35h` 寫入；`SCAN` 實機幾何、手動／Quick UI、解除與演出待續）

## 結論

第 433 輪留下的 `013E:0089h` 不是 overlay 12。依 Borland module、resident
control segment 與 typed TPOV resolver 重新驗證後，exact projection 是：

```text
EFFECTS resident 013E:0089h
  → overlay 23 entry 21
  → overlay-local 2325h PUTEFFECT
```

Sleep handler 已先完成 `4d4`／HD 容量篩選；共用 writer 隨後按保留順序逐一
呼叫 `PUTEFFECT`。後者先把 `SPELLREC.CASTON=1` 寫到 current-effect byte
`A02Dh`，再以 `CHECKFX(target,9)` dispatch 目標的效果。operational `6Ah`
沿第 411／412 輪 exact handler 擲魔法抗性；成功時 `Protected(0)` 清
`A02Dh`，`PUTEFFECT` 顯示「沒有作用」並跳過 `ADDEFFECT`。容量不會退還。

未被抗拒的 Sleep exact 寫入：

| `EFFECTREC` byte | 值 | 證據 |
|---|---:|---|
| `+0` | `35h` | `SPELLREC.EFFECTNUM` → `ADDEFFECT` |
| `+1..+2` | `5 × caster level` | spec 433 duration helper |
| `+3` | `01h` | `SPELLREC.CASTON` |
| `+4` | caster level byte | common writer `FIGCASTERLEVEL` result |
| `+5..+8` | runtime link | `ADDEFFECT` 先清，再由 list 管理器連結 |

這證明 effect record 與結算順序，不證明 `35h` 的每分鐘 decrement、受傷喚醒、
戰後保存、畫面 twinkle、聲音或完整原版 wall-clock timing。

## 非破壞性證據

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland `EFFECTS` symbols、resident stub | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | typed overlay resolution | `exact` |
| overlay 23 | `a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9` | `PUTEFFECT／ADDEFFECT／SPELLOFF` | `exact bytes` |
| overlay 12 | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` | `6Ah` 15% 魔抗 wrapper | `exact bytes` |
| overlay 22 | `c54729525d576c11d731d64a1b06ee2547b2562b73e3708a1beaafc535cabbe8` | Sleep handler／common writer | `exact bytes` |

`scripts/ida/pc98_monster_affect_loader_audit.idc` 只在 `/tmp/coab-ida-434`
的 overlay 23 副本建立新 database，輸出 729 行非空 ledger；原始 executable、
GAME.OVR、code-only overlay 與既有 database 均未 rename 或修改。

## 位址空間訂正

Borland symbols exact 提供：

```text
EFFECTS 013E:13D7 ADDEFFECT
EFFECTS 013E:2325 PUTEFFECT
EFFECTS 013E:2419 HEALDUDE
```

同一 module 的 `ADDEFFECT 13D7h` 早已由 overlay 23 raw bytes證實。typed
resolver 本輪輸出：

```text
stub_resolution overlay=23 stub=0x0089 entry=21 code=0x2325 flags=0x00
```

先前以 `12:0089h` 得到的 local `043Ah` 屬另一個 resident control segment，
不能因 stub offset 相同就視為 `PUTEFFECT`。這項舊候選已被推翻，不保留為
並存答案。

## PUTEFFECT 控制流

overlay 23 local `2325h..23F8h` 的關鍵順序：

1. `233Eh..2341h` 把 `CASTON` 參數寫入 `A02Dh`。
2. `2344h..234Eh` 以 check type `09h` 呼叫 `CHECKFX(target,9)`。
3. `2351h..2362h` 若 `A02Dh==0`，或 save outcome 非零且
   `SAVERESULT==1`，走 no-effect 分支。
4. `2364h..2384h` 以角色名接上 CP932 Pascal string
   `には影響がなかった。`，即「沒有受到影響」，然後提前離開。
5. `2386h..23B9h` 以 `SPELLON` 找同一 `CASTON` group；存在且 duration
   非零時先 `SPELLOFF`，避免重複鏈結。
6. `23BCh..23D2h` 呼叫 `ADDEFFECT(target, CASTON, duration,
   casterLevel, effectKind)`。
7. 有訊息時才於 `23D5h..23EEh` 呼叫 `TWINKLE`／清訊息視窗；這項靜態
   順序尚未等同逐幀演出完成。

Sleep record 的 `SAVERESULT=0` 已由 spec 433 證明，所以第 3 步的 save
分支不成立；可使 `A02Dh` 清零的既有 target effect 才是此路徑的阻擋來源。
第 411／412 輪已 exact 證明 `6Ah → base 15` wrapper，公式仍是：

```text
threshold = 15 + (11 - casterLevel) * 5
resisted  = d100 <= threshold
```

原程式不 clamp threshold。PRNG 順序是先四次 d4、完成全部容量選取，再按
selected target order 只對具 operational 魔抗者擲 d100。

## Remake bounded contract

`Battle.CastSleepOrdered` 明確要求 adapter 傳入上游已驗證順序：

- 不自行搜尋、擴大或排序 target；未知與重複 ID 失敗即關閉。隊伍／死亡
  predicate 屬上游 `SCAN`／handler，尚未閉合前不能在 core 自行排除。
- 呼叫 engine `combat/sleep` 擲四次 d4、保留 candidate order，已有 held
  effect 不消耗容量，昂貴 target 失敗後仍繼續。
- 五 HD 成本仍讀 shared Player raw `+74h`：零值成本 10，非零成本 20；欄位
  只以 `RawPlayer74` 保存，不命名為種族或其他規則語意。
- 容量選中後才按序擲魔抗；抗拒者保留 impact／PRNG 消耗但不寫 effect。
- 成功者寫入上述 raw `35h` record。`Raw4` 另外保存 caster level，不再把
  byte `+4` 壓成單一 `Active` bool。

focused tests 覆蓋 deterministic 四次 d4、已有 held 排除、抗性 d100 順序、
抗拒不寫入、成功 record 五欄，以及未知／重複 target fail-closed。因本輪沒有
`SCAN` runtime geometry 與 UI，Sleep 仍不得列為正常玩家可施放完成項。

## 下一步

1. 以 PC-98／DOSBox 固定戰鬥場面追 `SCAN` 三 byte record，關閉 footprint、
   terrain、第二／三欄與 tie order。
2. 將正確 target adapter 接到手動格點選取與 Quick delayed transaction，使用
   game-pack spell `15h`（十六進位）metadata，不在 State 增加中文特判。
3. 追 effect `35h` decrement、受傷／喚醒、combat end／save round-trip、
   twinkle、訊息與聲音，再做正常玩家路徑及原版影片對照。
