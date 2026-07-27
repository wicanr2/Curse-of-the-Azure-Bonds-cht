# 第三百二十一輪：巫師塔後龍巫妖遭遇

狀態：`READY`

## 反組譯證據

Area 5 離場後，ECL1 block `0x50` 世界 dispatcher 回到哈普。選擇
`JOURNEY ON → ESSEMBRA → TRAIL` 會顯示一具骷髏形體從樹叢走出，宣稱隊伍
奪走了他的導師並要復仇。Continue 後執行 `+0x149A`：

1. `CLEAR MONSTERS`
2. `LOAD MONSTER 0x3C,1,0x3C`
3. `COMBAT`

`0x3C` 不在目前 ECL1 的 MON1 table，而是 `MON5CHA` 的 DRACOLICH：66 HP、
raw AC `66`（combat AC `60-66=-6`）、attack bonus `180`、`3d8` damage、
initiative `24`、ControlMorale `0xB2`。因此 spawn monster ID range，而非目前
ECL block namespace，才是 encounter record 的首要 chapter key。

## 引擎與中文契約

- `StartEncounterWithAffects` 逐一以 monster ID resolve `MON*CHA` 與
  `MON*SPC`；目前 chapter records 只作 fallback。
- packed monster AC 的已觀察範圍擴充至 50..70，仍以 `60-raw` 轉換；小於
  50 的 synthetic／intermediate 值不轉換。
- 顯示名為「龍巫妖」，復仇台詞使用繁中；script opcode、monster ID 與原始
  record 不因翻譯而改動。
- 戰鬥使用 CPIC5 block `0x3C` 原版像素小人，在 640×480 logical canvas 以
  nearest-neighbor 整數放大；中文由 24px 高解析字型另外 rasterize。
- 勝利 continuation 回到 `LocationEssembra`，保留
  `ENTER CITY / JOURNEY ON / CAMP` 世界選單。

## 驗收

- real-image regression：直接由 ECL1 block `0x50 +0x149A` 得到唯一
  `MonsterID=0x3C, Count=1, IconBlock=0x3C` 的 COMBAT request。
- parser regression：MON5 raw AC `66` 轉換為 combat AC `-6`。
- 端到端 State regression：法師塔離場、哈普到艾森布拉 TRAIL、繁中復仇對話、
  一隻 66 HP 龍巫妖戰鬥與勝利抵達艾森布拉全部沿同一真實 session 完成。
