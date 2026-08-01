# 第四百一十一輪：PC-98 提朗瑟克斯魔法抗性（READY）

## 範圍與結論

本規格只關閉提朗瑟克斯 `MON6SPC` effect `6Ah` 的 15% 魔法
抗性，以及現階段已有完整施放路徑的 Magic Missile。它不代表
Fireball、Lightning Bolt、持續雲霧、所有魔法傷害類型或提朗瑟克斯
全部特殊能力已完成。

- PC-98 overlay 12 的 common routine 以
  `base + (11 - casterLevel) * 5` 建立門檻。
- 它只在 current affect 存在，或 damage flags 含 Magic bit `08h` 時擲
  `1d100`；`roll <= threshold` 時呼叫 `Protected`。
- `Protected(0)` 清除 damage byte `A02Eh` 與 current affect byte `A02Dh`。
- local `23F4h` wrapper 傳入 `50`；local `2404h` wrapper 傳入 `15`。
- effect `6Ah → 15% wrapper` 的連結由同一套反編譯的二次轉寫
  交叉支持，目前推論等級為 `strong inference`；TPOV relocation 未完整
  套用前，不把 `INITEFFPROX` raw addend 冒稱為直接 handler 位址。

## 輸入與證據等級

| 輸入 | SHA-256／版本 | 用途 | 等級 |
|---|---|---|---|
| PC-98 `GAME.OVR` overlay 12 | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` | common routine、wrapper、`Protected` raw bytes | `exact` |
| PC-98 `MON6SPC` effect `6Ah` | `6A 0000 FF 00 0900191D` | 提朗瑟克斯真實 template 能力 | `exact` |
| `simeonpilgrim/coab` | commit `9dc46f1d5911710fb2fcb7a9c0ec0ef74264d17c` | `6Ah` 名稱、15% wrapper 及魔法飛彈呼叫順序交叉檢查 | 二次 oracle |
| Standing Stone 正常玩家路徑 | 真實 `MON6CHA/MON6SPC` | 確認 `47h` 怪物的 `6Ah` 進入終戰 | `exact` |

`simeonpilgrim/coab` 是依 DOS 反編譯寫成的舊 C# 重寫，只作交叉點；
公式、常數與傷害清除仍以本機 PC-98 bytes／IDA 為主證據。

## IDA Pro 9.4 與 raw bytes

原始 `GAME.OVR`、TPOV 輸入與基準 database 保持唯讀。稽核腳本只在
Docker 內的 code-only overlay 副本加建指令邊界，輸出仍並列 local
offset、raw bytes 與附加語意。

### `Protected` local `001Bh..003Bh`（`exact`）

```text
001B 55 89 E5             push bp; mov bp,sp
001E 80 7E 06 00          cmp byte ptr [bp+6],0
002C C6 06 2E A0 00       mov byte ptr ds:A02E,0
0031 C6 06 2D A0 00       mov byte ptr ds:A02D,0
0036 89 EC 5D CA 02 00    epilogue; retf 2
```

common routine local `2396h..23F3h` 的關鍵資料流（`exact`）：

```text
23AF  B8 0B 00            AX = 11
23B2  2B C2               AX -= caster-level result
23B4  B9 05 00 / F7 E9    AX *= 5
23BB  8A 46 06 / 03 C2    AX += byte argument (base)
23C5  80 3E 2D A0 00      current affect != 0 ?
23CC  A0 2F A0 / 24 08    damage flags & Magic(08h) ?
23D7  B0 01 / B0 64       request 1d100
23E2  3A 46 FE / 77 07    if roll > threshold, skip
23E7  B0 00 / ... / E8 2D DC  Protected(0)
```

wrapper（`exact`）：

```text
23F4  55 89 E5 B0 32 50 0E E8 98 FF ... CA 0A 00  ; base 50
2404  55 89 E5 B0 0F 50 0E E8 88 FF ... CA 0A 00  ; base 15
```

headless IDA 本輪曾出現「exit code 0，但新報告沒有持久化」的工具
陷阱；該次不列入新證據。上述結論來自先前已保存的 IDA 報告，
並由同一 SHA 的 raw bytes 重讀交叉驗證。後續腳本 gate 必須同時檢查
輸出是 regular file 且非空，不可只信 IDA process status。

## Remake 實作契約

- `Fighter.MonsterMagicResistanceBase()` 只投影已證明的 operational
  `6Ah`，回傳 base 15；不把作品怪物名、sprite block 或劇情寫入公式。
- `MagicResistanceChance(base, casterLevel)` 保留 exact signed expression，不擅自
  clamp 到 `0..100`。
- Magic Missile 先擲完全部 `2..5` 傷害骰，再擲抗性 `d100`；成功時
  傷害歸零並標記 `SpellResult.Resisted`。
- 施放格仍消耗，Magic Missile travel／impact 與 combat continuation 仍進行。
- 繁中訊息由 `assets/locale/zh-TW.json` 的 stable ID
  `combat_magic_resisted` 取得；測試在執行時讀取正式 JSON，不複製顯示字串。

## 驗證

- `TestMagicResistanceChanceUsesPC98LevelAdjustment`：鎖定施法者等級
  `10/11/12 → 20/15/10%`。
- `TestCastMagicMissileHonorsOperationalEffect6A`：同時覆蓋成功、失敗與
  non-innate inactive 反例，並驗證成功時 HP 不變。
- `TestCombatCastMagicMissileUsesLocalizedResistanceMessage`：由正式 locale JSON
  解析 stable ID，驗證施放格消耗、中文訊息與戰鬥輪轉。
- `TestRealPlayerPathStandingStoneToBurialGlen`：由 Standing Stone 通過
  真實 GEO／ECL／MON6 資料抵達終戰，驗證提朗瑟克斯的 `6Ah`
  投影為 15%，然後繼續完成 `PROGRAM 8`。

## 尚未完成

- TPOV relocation／fixup 的完整 typed decoder，以及 `INITEFFPROX` slot 到
  handler runtime far pointer 的一次證據橋接。
- 50% wrapper 對應的 effect，以及 `4Fh／70h／84h／87h` 的完整語意。
- Fireball、Lightning Bolt、Cloudkill 與其他 Magic damage 進入同一
  pre-damage affect boundary 的正確時序；在此之前不廣泛套用抗性。
- 原版「is unaffected」文字、影像時序、sound cue 與終戰實機畫面。
