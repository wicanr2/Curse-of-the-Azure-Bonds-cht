# 第一百七十二輪：DOS thief skill preservation

狀態：`READY`（限 `.SAV/.GUY` thief skill bytes → party／save adapter）

## 反組譯證據

reference `Player.cs` 定義 `thief_skills[8]` 位於 record `0xEA–0xF1`，順序為：pick pockets、open locks、find/remove traps、move silently、hide in shadows、hear noise、climb walls、read languages。`ovr015.pick_lock` 使用 index `1`，以 `roll_dice(100,1) <= thief_skills[1]` 判定。

## 實作

- `ParseDOSPlayerRecord` 保存 `0xEA:0xF2` 的八 bytes 到 `DOSPlayerRecord.ThiefSkills`。
- `Character.ThiefSkills` 以原始順序保存並跟隨既有 versioned party/game JSON；`OpenLocksSkill()` 暴露 index 1 的 verified percentage。
- 缺少或非 DOS 匯入資料時回傳 0，代表「沒有已驗證 skill data」，不等同自動失敗或成功。

## 明確 boundary

本輪沒有實作 thief skill 重算（race／DEX／level／items）、pick-lock dice transaction、door menu、bash／knock、技能消耗或完整 DOS save container；只建立資料保存與 rules adapter input。

## 驗證

synthetic DOS player regression 覆蓋八 bytes 保存與 `open_locks` index 1 projection；既有 party JSON／sidecar parser 與 Docker gate 覆蓋相容性。
