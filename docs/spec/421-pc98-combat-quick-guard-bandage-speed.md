# 421：PC-98 戰鬥 QUICK／GUARD／BANDAGE／SPEED

狀態：`READY`（2026-08-02）

本規格只關閉戰鬥主選單的單人 QUICK、DONE 子選單的 GUARD／BANDAGE／
SPEED，以及空白鍵恢復玩家角色手動控制。`ALT+Q` 全隊快速戰鬥與
`ALT+M` 快速戰鬥施法在本輪當時仍列為後續工作；前者已由 spec 422、後者的
Magic Missile 有界切片已由 spec 424 接續。不得由本輪擴大宣稱完整自動戰鬥。

## 輸入與可重現工具

- `GAME.EXE` SHA-256：
  `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`
- `GAME.OVR` SHA-256：
  `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a`
- PC-98 Borland symbols 與 typed overlay resolver 用來解析 resident stub；
  原始 overlay bytes 另由 IDA Pro 9.4 的
  `scripts/ida/pc98_combat_commands_audit.idc` 連續輸出。
- IDA database、原始 binary 與 overlay 保持唯讀；下列語意是附加 ledger，
  不以 rename 覆蓋原始位址。

## Exact 指令契約

### 清除動作與防守

- overlay 24 local `2A5Bh..2AADh`（resident `014A:00CA`）先取得
  `Player+18Eh` 的 Action，再把 `Action+03h` delay、`+00h`、`+07h`
  guarding 與 `+06h` 清零。`+00h` 已由第 425／429 輪證明是 pending spell；
  `+06h` 仍不另取語意名稱。第 431 輪另證明 held effects 會呼叫此 consumer，
  但不呼叫 memorized-slot consumer。
- overlay 24 local `2AAEh..2B2Fh`（symbol `GUARD`、resident
  `014A:00CF`）先呼叫上述 clear，再把 `Action+07h=1`。
- overlay 13 local `0684h..078Bh` 是移動後的 consumer：建立新位置的
  鄰接目標清單，要求鄰接者 `Action+07h != 0` 且 held predicate 為假；
  在攻擊前於 `074Ah` 寫回 `Action+07h=0`，再重算 attack-received 並呼叫
  一般攻擊。raw bytes 與可讀次級重寫碼在此資料流一致。
- 因此 remake 的 Guard 是「清除本回合動作後進入一次性防守；敵方踏入其
  鄰接格時立即以一般近戰攻擊一次；攻擊前消耗旗標；被定身者不能觸發」。
  它不是離開威脅格的 free attack，也不是被動 AC bonus。
- 原選單只在武器不是純遠程，或該遠程武器也可近戰時提供 Guard。現有 typed
  投影以 `MissileWeapon && !ThrownWeapon` 表示純遠程；這項可見性映射標為
  `strong inference`，待完整 ITEMS weapon-mode consumer 再升級。

### 包紮

- overlay 24 local `35D8h..3686h`（resident `014A:010B`）依 TeamList
  順序掃描：`Action+13h==0`、`Player+198h combat_team==0`、
  `Player+196h health_status==5` 才是候選。
- apply 參數非零時，只把第一名候選的 raw health `5→4`、`Action+0Eh=0`，
  隨即把 apply 清零，所以後續候選只影響「是否可包紮」回傳值，不再被修改。
- 既有角色狀態／傷害 consumer 已把該 raw transition 投影為
  `Dying→Unconscious`，`Action+0Eh` 投影為 bleeding。remake 必須修改正式
  party roster 的第一名垂死者並止血，不能只改畫面 fighter 副本。
- BANDAGE 完成後呼叫 clear-actions 並結束目前角色回合；沒有候選時不能顯示
  可用選項，也不能消耗回合。

### 快速戰鬥與手動控制

- overlay 8 local `1375h..140Fh` 把 `Player+199h quick_fight=1`；若目前
  Action target 非空且 target `Player+198h combat_team` 與自己相同，清除
  target far pointer。現有 remake 尚未保存 per-action target pointer，因此
  「取消同隊目標」在 typed target 欄位接通前保持明確缺口。
- 主選單 QUICK 只切換目前角色後立刻由既有 combat AI 執行其本回合。
- 空白鍵只清除 `ControlMorale < 80h` 的玩家角色 quick flag；NPC／臨時怪物
  盟友仍由電腦控制。`ControlMorale` 必須由 party record 投影，不能以角色名
  或 TemporaryAlly 猜測。

### 遊戲速度

- overlay 8 local `124Dh..1374h` 讀寫 `ds:7F16h`。值域為 `0..9`；只有
  `<9` 才提供 SLOWER 並遞增，只有 `>0` 才提供 FASTER 並遞減。`0` 最快、
  `9` 最慢。
- overlay 18 local `0B07h..0B24h` 讀 `ds:7F16h`、加三，再與目前動畫 frame
  delay 相乘；後續以 elapsed time 比較該結果。故時間倍率是
  `frameDelay × (speed+3)`，不是只改選單數字。
- remake 預設採原版一般設定 `4`。為保持目前 JSON visual duration 在預設值
  下不變，renderer 將 elapsed time 乘以 `7/(speed+3)` 後交給既有 timeline。
  deterministic screenshot 的 frozen elapsed 不得被設定改寫。

## 實作與驗收門檻

- 戰鬥核心：clear、guard、quick/manual control、移動後 guard reaction；穩定
  fighter order、held suppression、一次性消耗及死亡停止均有具名測試。
- game adapter：Bandage 修改正式 roster；Guard、Done、Bandage 正確結束
  scheduler selection；Quick 交給現有 AI；所有訊息從 locale stable ID 解析。
- Ebiten：DONE 子選單 G／B／S、速度子選單 S／F／E、主選單 Q、空白鍵恢復
  玩家角色手動控制；按鍵可用性與畫面選單一致。
- Docker 跑正式非 GUI tests；Docker／Xvfb 跑 Ebiten 輸入測試與實際戰鬥
  screenshot／runtime trace。只以 direct fixture 驗證 opcode 不足以支撐完成。

## 尚未完成

- `ALT+Q` 全隊快速模式需要把 AI 回合拆成可逐幀中斷的 scheduler step，否則
  同步跑完整場戰鬥時無法用 Space 取回控制。
- `ALT+M`／特殊鍵 2 的 gate 與 Magic Missile 已由 spec 424 接通；area
  suitability、casting delay、原版 AI 全規則、per-action target pointer、
  純遠程／可近戰武器的完整原始 mode consumer 仍未完成。
- 所有動作的 DOS／PC-98 wall-clock 與音效逐幀 fidelity。
