# 第四百一十五輪：PC-98 怪物閃電排程與雙傷害池（READY）

## 範圍與結論

本輪以唯讀 PC-98 overlays、Borland symbols、raw bytes 與 Docker 內
IDA Pro 9.4 關閉提朗瑟克斯 effect `84h` 的回合階段、法術編號、兩次傷害骰
與 turn handoff，並接入 remake 正常怪物排程。這是怪物 Lightning Bolt
vertical slice；目標候選的原版 range／LOS 排序、牆角逐幀 oracle 與精確
wall-clock timing 仍未完成。

## 輸入與非破壞性

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | `ROUND`、Borland symbols／resident calls | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay corpus | `exact` |
| overlay 9 code | `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20` | action-phase `CHECKFX` caller | `exact` |
| overlay 22 code | `c54729525d576c11d731d64a1b06ee2547b2562b73e3708a1beaafc535cabbe8` | effect `84h` handler | `exact` |

原始檔與基準 database 不修改。`scripts/ida/pc98_monster_lightning_audit.idc`
只在 `/tmp` 分析副本建立指令並輸出文字報告；報告保留 overlay-local 位址與
raw bytes，語意是附加標籤，沒有 rename 原始 symbols。

## action phase 與前三回合

overlay 9 local `0DD3h` 的每戰鬥者 action routine 在其他行動前執行：

```text
0DF4  B0 0E       mov al,0Eh
0DF6  50          push ax
0DF7..0DFA        push combatant far pointer
0DFD  9A34003E01  call far 013E:0034h   ; CHECKFX type 14
```

因此 effect `84h` 位於 type-14 特殊行動階段，早於一般物理攻擊。overlay 22
handler local `62DDh` 比較 Borland `ROUND`（DS `A81Fh`）與 `04h`，值不小於
4 就直接離開。獨立 DOS 重寫來源只作交叉驗證：battle setup 將
`combat_round=0`，每輪 action 前先遞增，故 remake 從 1 起算的
`Battle.Round() < 4` 正好對應原版第 1、2、3 回合。交叉來源固定為
`simeonpilgrim/coab@9dc46f1d` 的 `ovr011.BattleSetup` 與
`ovr009.BattleRoundChecks`，不以其命名取代本機 bytes。handler 結尾呼叫
`clear_actions(caster)`，所以成功施放會消耗該怪物當回合，不再接物理攻擊。

| 斷言 | 等級 |
|---|---|
| type 14 在一般 action 前被呼叫 | `exact`（IDA＋raw bytes） |
| handler gate 是 `ROUND < 4` | `exact`（IDA＋Borland symbol） |
| remake round 1–3 對應原始值 1–3 | `strong inference`（獨立反編譯初始化／遞增交叉驗證） |
| 施放後清除該 combatant actions | `exact`（handler caller） |

## Lightning Bolt 與兩次獨立 `16d6`

effect `84h` handler local `62D7h` 的順序如下：

1. `6323h push 33h`，透過 DS `A38Ch` 的 combat spell-target delegate 選取
   Lightning Bolt `33h`。
2. `633Dh push 13h` 載入 missile icons；`6345h..6367h` 以 projectile type 4
   從施法者畫到 DS `A645h/A646h` 的目標座標。
3. `6374h push 10h; 6377h push 6; 637Ah call 013E:0052h`：擲第一份
   `16d6`，local `3BF5h` 只處理初始目標格、Spell save type 4。
4. `638Fh push 0Ah; 6392h push 10h; 6395h push 6; 6398h call
   013E:0052h`：再擲獨立第二份 `16d6`，local `3CB1h` 處理 range 10 的
   後續直線／反射、Spell save type 4。
5. `63A8h..63AEh` 清除施法者 actions。

兩次擲骰中間包含初始格 saving throw／damage consumer，不能先擲兩份再批次
套用，也不能把玩家 Lightning Bolt 的單一等級 d6 共用傷害池拿來代替。

## remake contract

- `ReflectingLineOptions` 新增中立的 initial／path dice profile；零值完全保留
  玩家法術既有「單一等級 d6 共用」行為。
- 怪物 effect `84h` 使用 initial `16d6`、path `16d6`、range budget 10、
  Electricity＋Magic flags、逐目標 Spell save 與 effect `87h` 防電。
- State 在 type-14 等價位置先處理 effect `84h`，只於 round 1–3 施放；
  成功後直接 handoff，不再嘗試 Magic Missile 或物理攻擊。
- 正式 Ebiten frontend 注入當前 DUNGCOM／WILDCOM `LineTerrain`。沒有 terrain
  callback 時失敗即關閉，不把未知牆面假裝成開放地圖。
- 繁中摘要由 stable IDs `combat_monster_lightning_bolt` 與
  `combat_monster_lightning_bolt_protected` 解析；動畫沿用已資料化的
  `lightning_bolt` COMSPR timeline。

## 驗證

- core regression 驗證 initial／path 傷害各落在 `16..96`、彼此獨立，且
  初始目標與路徑目標各自使用正確 damage pool／saving throw。
- State regression 驗證 operational／inactive `84h`、第 1 回合
  `VisualLineSpell`、繁中 stable ID，以及第 4 回合回到物理攻擊。
- Standing Stone 起始正常 GEO／ECL 長路徑載入真實 `MON6CHA／MON6SPC`，
  經最終儀式進入 37 人終戰；真實提朗瑟克斯排入 effect `84h` 動態時間軸，
  戰勝後仍由同一 ECL session 進入 `PROGRAM 8`。

## 尚未完成

- 原版 `SpellCastFunction(QuickFight,33h)` 的 reachable／visible candidate、
  range 與二十次重抽已由第 416 輪接入；原始 combatant-array 同距 tie order、
  invisibility runtime trace 與無目標 `(0,0)` 動畫仍未完成。細節以
  `416-pc98-monster-spell-target-selection.md` 為準。
- 正常長路徑測試使用有界開放 combat terrain，只證明真實 MON6 scheduler；
  正式 frontend 會使用實際地城牆，但終戰專屬牆面反射仍需 deterministic
  screenshot／DOS runtime 對照。
- 敵方施法的原版逐幀影片時間碼、距離相關 duration、caster pose、聲音先後、
  牆角與多次反射。
- Magic resistance `6Ah` 是否攔截 effect `84h`、初始格與 path save 的完整
  DOS PRNG trace，以及 effect `84h` target delegate 失敗時的原版 fallback。
