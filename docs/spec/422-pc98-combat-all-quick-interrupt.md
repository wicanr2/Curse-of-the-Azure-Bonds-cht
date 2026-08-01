# 422：PC-98 全隊 QUICK 與可中斷 handoff

狀態：`READY`

本規格關閉 `ALT+Q` 的全 TeamList Quick transaction、目前角色
`Action.delay 14h→13h` handoff，以及視覺時間軸播放中按空白鍵恢復玩家控制。
`ALT+M` 的旗標 writer／consumer 在本輪已定位；其 Magic Missile 有界切片
後續由 spec 424 接通。本文件仍只驗收 ALT+Q 與可中斷 handoff。

## 輸入與工具

- PC-98 `GAME.EXE` SHA-256：
  `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`
- PC-98 `GAME.OVR` SHA-256：
  `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a`
- extracted overlay 08：
  `d39b6aad76af8d3ccc182b4b4d95dd9195160c9c21b36bfa5b0cd4edc1942788`
- extracted overlay 09：
  `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20`
- extracted overlay 10：
  `cc0724159c0cd7dc550e9d3937ec6a2a5e8d290716b3178b9e7a31202f14afe4`
- IDA Pro 9.4，8086／16-bit，Docker、`--network none`；原始檔唯讀掛載，
  database 與報告只輸出至 `/tmp/coab-ida-422`。
- 可重現腳本：`scripts/ida/pc98_quick_magic_audit.idc`。

本文件的函式名稱只描述附加語意；權威座標是 overlay local offset 與 raw
bytes，不回寫 pristine `GAME.OVR`。

## Primary bytes

### `ALT+Q` transaction

overlay 08 combat input：

- `0677h 3C 10`：特殊鍵 code 比較 `10h`。
- `067Bh..0687h`：沿目前 Player `+18Eh` 取得 Action，
  `26 C6 45 03 14` 寫 `Action.delay=14h`。
- `0688h..06B6h`：由 `DS:9598h` 取得 TeamList head，逐節點呼叫 local
  `1375h`，再沿 Player `+18Ah` 走 next pointer。
- local `1375h` 的 per-player setter 寫 Player `+199h=1`，並在目前 target
  同 team 時清除 Action target far pointer；完整 setter bytes 已由 spec 421
  保存。
- `06B8h..06CDh` 清 prompt、依速度等待，再把 local done 設 1，離開人工
  combat menu。

下一次 action entry 位於同 overlay：

- `0310h 26 80 7D 03 14` 比較 delay `14h`。
- `031Fh 26 C6 45 03 13` 改寫為 `13h`。
- `03BEh..03DCh` 在 delay 仍大於零時檢查 Player `+199h`；Quick 為真即呼叫
  AI routine，否則進人工 combat menu。

結論：全隊 Quick writer、TeamList 範圍、目前角色 `20→19` handoff 與 AI
dispatch 是 `exact`。特殊鍵 code `10h` 對應 `ALT+Q` 另由本機 Reference Card
交叉支持。

### 空白鍵恢復人工控制

overlay 08 `05B6h..05F2h`：普通 key `20h` 從 `DS:9598h` 掃 TeamList；只在
Player `+0F7h < 80h` 時把 Player `+199h` 清零。NPC／怪物不被收回。
writer 與條件是 `exact`；其欄位名稱由既有角色 record consumer 交叉支持。

### `ALT+M` 定位結果（實作由 spec 424 接續）

- overlay 08 `0630h 3C 11` 比較特殊鍵 `11h`，`0634h..0641h` toggle
  `DS:A86Ch`，再由兩個 Pascal string 顯示 on／off。
- overlay 10 `1C8Ah C6 06 6C A8 00` 在 combat initialization 清零該旗標。
- overlay 09 local `0627h` 建立 Quick AI spell candidate；`06A1h` 先要求
  Player `+0F7h <= 7Fh`，`06A9h 80 3E 6C A8 00` 再要求旗標非零，之後掃描
  memorized spell bytes、呼叫 spell suitability predicate，最後於
  `072Bh..0754h` 回傳選出的 spell ID 或零。

因此 `A86Ch` 是「允許可控制 PC 的 Quick AI 選擇法術」為 `proven`。候選
優先序、三次抽選、CASTCOMBATSPELL 與 Magic Missile 路徑已由 spec 424
接續；area suitability 與完整 target policy 仍未完成。

## Remake contract

1. `Battle.SetAllQuickFight(current)` 先把目前 Action delay 寫 20，再依保存的
   TeamList order 對所有 combatants 呼叫同一 Quick setter。
2. AI 處理該角色前，只有 Quick 且 delay 20 的 handoff 轉為 19；一般 Quick
   與 initiative delay 不受影響。
3. Ebiten 戰鬥模式以 `ALT+Q` 觸發；視覺時間軸播放期間仍讀取 Space，完成
   當前動畫後下一個可控制 PC 必須回到人工輸入。
4. Space 沿用 spec 421 的 `ControlMorale < 80h` 範圍，不清 NPC／怪物 Quick。
5. Space 清除後必須同步回持久 party projection；否則 ECL continuation 建立
   下一場戰鬥時會從舊 Player view 錯誤恢復 Quick。
6. 所有新文字由 locale stable ID 解析，不把繁中顯示字串寫入規則測試。

## 驗收與邊界

- core regression：全 TeamList Quick、20→19、Space 只收 PC。
- State regression：啟用真實 combat visual handoff；`ALT+Q` 的第一個 AI
  action 必須 yield 成 visual event，等待期間 Space 可收回 PC 而保留 NPC。
- 正常玩家路徑：由 Standing Stone 經 GEO／ECL 抵達紅網，以全隊 Quick
  完成四巨蛛戰；在羅剎妖 continuation 前 Space 收回後，第二戰必須停在人工
  玩家回合，不能因 stale persistent Quick 自動跳過。
- Ebiten／完整套件 gate 必須在 Docker/Xvfb 通過。

尚未完成：`ALT+M` 的 area／delay／其餘法術、原版 AI 移動／物品／Guard 策略、
特殊鍵的 PC-98 runtime scan-code trace、原版每個 Quick action 的 wall-clock
節奏，以及完整玩家戰鬥到勝利的逐鍵影片 oracle。
