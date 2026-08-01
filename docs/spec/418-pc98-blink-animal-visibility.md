# 第四百一十八輪：PC-98 Blink、動物視覺與 action delay（READY）

## 範圍與結論

本輪延伸 spec 417 的 `CHECKTARGET` visibility，關閉 effect `25h` 與 `45h`
的 PC-98 handlers，並追出 shared Player record `+11Ah` 是怪物類型欄位：

- operational `25h` 在目標 `Action +3 == 0` 時設定 target hidden，並把
  attack roll 直接寫成 `FFh`；因此 natural 20 仍會 miss。非零 delay 時不作用。
- operational `45h` 只在觀察者 `MonsterType == 13h` 時作用。觀察者沒有
  operational `18h` 便設定 target hidden；不論有沒有 `18h`，attack roll
  都會再減 4。
- `+11Ah == 03h` 由獨立 effect `4Bh` dragon-slayer handler 消費；六章真實
  MON*CHA corpus 中 `13h` records 是 WORG、FIGHTING DOG、MONKEY、OWL BEAR。
  因而 `+11Ah` 可投影為 `MonsterType`，`13h` 可命名為 Animal。
- 原版每輪會配置非零 action delay，行動完成後歸零。remake 現把相同生命週期
  接到現行 scheduler，避免 blink 怪物因 delay 永遠維持零而永久無敵。

本規格不宣稱 remake 現行 initiative 數值公式、tie order 或 PRNG draw count
已與 PC-98／DOS 完全一致；它只將「scheduled/pending 非零、completed 零」的
visibility consumer boundary 標為 READY。

## 輸入、雜湊與非破壞性分析

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | EFFPROCS table slots | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | typed TPOV entry resolution | `exact` |
| overlay 12 code | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` | effects `25h／45h／4Bh` | `exact` |
| DOS image ZIP | `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` | MON1–6CHA corpus | `exact bytes` |

`scripts/ida/pc98_visibility_target_audit.idc` 只操作 Docker `/tmp` 內的 overlay
副本，輸出連續位址、raw bytes 與反組譯；沒有 rename 原始 symbols 或修改基準
`.i64`。原始遊戲 ZIP 只讀，MON corpus 由正式 DAX decoder 解壓。

## typed handler resolution

`INITEFFPROX` 的直接 table writes 與 resolver 結果：

```text
effect 25h slot A0D4/A0D6 -> resident 008B:00E3 -> entry 39 -> overlay 12:0BDB
effect 45h slot A154/A156 -> resident 008B:0174 -> entry 68 -> overlay 12:16C2
effect 4Bh slot A16C/A16E -> resident 008B:0192 -> entry 74 -> overlay 12:17B5
```

每筆都經五-byte resident stub 與 TPOV handler bound 驗證；不能把 raw addend
直接當 overlay local offset。

## 原始 handler 資料流

effect `25h` local `0BDBh`：

```text
0BDE  les di,[bp+0Ch]
0BE1  les di,es:[di+18Eh]       ; combatant -> Action
0BE6  cmp byte ptr es:[di+3],0
0BED  mov byte ptr ds:A035h,1   ; target hidden
0BF2  mov byte ptr ds:A039h,FFh ; replace attack roll
```

effect `45h` local `16C2h`：

```text
16C8  les di,ds:9594h
16CC  cmp byte ptr es:[di+11Ah],13h
16D4  push ds:9596 / ds:9594
16DC  mov al,18h
16E4  call effect-query
16ED  mov byte ptr ds:A035h,1   ; only when effect 18 absent
16F2  sub byte ptr ds:A039h,4   ; always for type 13h observer
```

effect `4Bh` local `17B5h` 獨立讀同一 `+11Ah`，比較 `03h` 後擲 `1d12`、
乘 3、加 4 與 strength bonus，再給 attack roll `+2`。這與既有 dragon-slayer
規則、真實 MON records 的名稱分布共同關閉欄位語意。

## 真實 MON corpus

正式 decoder 掃描 `MON1CHA.DAX` 至 `MON6CHA.DAX` 的所有 422-byte records；
`+11Ah == 13h` 只出現在：

| 資料檔／block | 原始名稱 | raw type |
|---|---|---|
| `MON1CHA／57h` | `WORG` | `13h` |
| `MON2CHA／05h` | `FIGHTING DOG` | `13h` |
| `MON2CHA／06h` | `MONKEY` | `13h` |
| `MON5CHA／37h` | `OWL BEAR` | `13h` |

名稱本身只作語意交叉驗證；正式 parser 仍保存 raw `MonsterType` byte，不把
四個名字硬編碼成動物判定。

## remake 契約與驗收

- `monster.Record` 從 raw `0x11A` 解析 `MonsterType`，並投影到
  renderer-neutral `combat.Fighter`；Animal constant 是 `13h`。
- `Fighter.VisibleTo` 加入 delay-aware `25h` 與 observer-type-aware `45h`。
- 物理命中以 `MonsterAffectForcesAttackMiss` 讓 `25h` 覆寫 natural 20；
  `45h` 對 Animal observer 保留 AC `+4`，即使 `18h` 已避免 hidden。
- `Battle.StartRound` 不新增額外 PRNG draw，沿用現行 initiative 結果建立
  正 delay marker；`CompleteAction` 在 State 所有 action path 共用的
  `advanceCombatToParty` 入口清零已完成 fighter。
- core tests 覆蓋 blink natural 20、delay 0／非零、Animal／非 Animal、
  `18h` 與 `45h` 的 hidden／AC 差異，以及 action completion lifecycle。
- Tilverton 正常玩家路徑的犬舍戰從真實 `MON2CHA` 載入 FIGHTING DOG，驗證
  `MonsterType=13h`；Standing Stone→Myth Drannor 長路徑則證明真實 phase
  spider 戰鬥仍可完成，沒有因 blink delay 永遠為零而卡死。

## 尚未完成

- PC-98／DOS exact initiative 公式、delay value、tie order 與 PRNG draw sequence。
- `25h／45h` 的施法、解除、duration、save round-trip 與所有角色 effect
  生命週期；目前只接通已載入 operational effects 的 combat consumers。
- ECL DAMAGE 若要處理 `45h`，仍須注入 attacker `MonsterType` 與 detect context；
  不得在缺 observer 時猜測。
- 其他 visibility／displacement effects、effect `84h` `(0,0)` fallback 動畫、
  終戰牆面逐幀與音效時序。
