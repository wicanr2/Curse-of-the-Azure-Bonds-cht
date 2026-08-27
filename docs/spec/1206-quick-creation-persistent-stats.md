# 1206 — 快捷建隊不得遺失原版起始經驗與持久戰鬥欄位

- 日期：2026-08-26
- 經驗值證據：`exact`
- 快捷模板的 HP 擲骰：`strong inference`（公式與資料表 exact；固定 seed 是 remake 的可重現取樣，不宣稱同一次原版亂數）

## 問題

前端快捷建隊由 `gamepack/pack/00-core.json` 的模板直接建立角色。舊實作只填種族、
職業、能力與豁免，留下兩個玩家可見斷鏈：

1. 原版建角依職業組合發下的 25,000／12,500／8,333 XP 沒有寫入，訓練場把新角色
   顯示成 `1 級／0 XP`。
2. 持久 roster 的 HP 是 `0/0`，只有 `Character.Fighter()` 的 fallback 臨時投影成
   可戰鬥數值；存檔、服務與重新投影讀到的不是同一份狀態。

這不是一般強度隊伍打不贏的平衡問題，而是原始資料 → typed character → combat →
UI／save 的垂直鏈中途遺失欄位。

## 原版契約

- spec 1093：單職 25,000 XP、雙職 12,500 XP、三職 8,333 XP；原版的法師／盜賊
  組合例外仍照 8,333。
- spec 1101：建角 HP 由職業槽生命骰、體質與多職除數決定，並同步寫最大 HP、
  基礎最大 HP 與目前 HP。
- spec 1140：建角同時寫入命中能力與基礎護甲欄位。
- spec 1101 ＋ spec 622：DOS NEWCHAR 對角色記錄 `+103h` 寫入 `12Ch`；該欄是
  Platinum，因此每名新角色另有 **300 PP**。這是 `exact`，不是 remake 平衡補貼。

## 修正

`starterCharacters` 現在依模板的 `RawClassID` 查同一份 `character-tables.json`：

- 寫入 `StartingExperience`；
- 呼叫與引導建角相同的 `RollCreationHitPoints`；
- 寫入目前／最大／基礎最大 HP 與一級 Hit Dice；
- 呼叫同一份 `CreationCombatBase`，寫入命中能力、基礎護甲與能力調整開關。
- 寫入原作的 300 PP 起始資金。

快捷模板仍是 remake 的省時入口；它省略互動，不再省略角色記錄本身。HP 固定 seed
只保證封包與測試可重現，不能標成原版同一次擲骰的 exact parity。

## 動態結果與限制

修正後的無強化鍵盤場（`COAB_KEY_BOOST=0`）在 12,000 幀內：364 格、4 個 ECL 段、
99 句、落回原文 0、全滅重開 0；第一名戰士透過正式訓練升到 4 級，並由正式商店
購得板甲與盾牌。機器可讀
結果在 `docs/audit/key-driven-normal-strength.json`。

這一場沒有通關，且快捷操作仍固定連選六名戰士；皇家衛兵戰仍會擊倒隊伍。因此只
證明起始資料、合法訓練／採買及早期一般強度路徑，不支持「一般強度隊伍已完整通關」。

## 驗證器勘誤

修正持久 HP 後，`cmd/cell-sweep` 的套件測試一度從 15 個 block／至少 235 格掉成
14／231。產品資料沒有少：正式命令仍量到 15／240。根因有兩層，都是測試夾具：

- 盤點專用 `BoostParty` 原本放在 `EnterSegment` 之後，入口戰鬥可能先用一般 HP
  打完；現在移到進段之前。
- Go 套件測試從 `cmd/cell-sweep` 執行，卻沒有把段內快照路徑設為相對該目錄的
  `../../workplace/campaign-frames/snapshots`；冷進不去的段因此安靜地少一整段。

這次失敗反而證實 0/0 後備值先前掩蓋了驗證器前置條件；修法保留一般角色的真實 HP，
不放寬逐格盤點門檻。
