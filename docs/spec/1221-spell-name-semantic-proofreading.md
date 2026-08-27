# 1221：法術名稱跨畫面語意校對

狀態：`READY`（2026-08-27）

## 範圍與方法

以原版 100 筆法術表的編號與英文名稱為身分層，對齊：

- `assets/locale/zh-TW.json` 的戰鬥、紮營與神殿 UI。
- game-pack 的英文／繁中 `spell_*` stable ID。
- `TestSpellNameKeysFollowTheOriginalTable` 量到的 81 個玩家法術名鍵。

## 修正

| 原文 | 舊譯 | 正式譯名 | 證據 |
|---|---|---|---|
| Shocking Grasp | 電擊觸手 | 電擊之握 | `grasp` 是握、不是觸手；原版表編號 20 |
| Cure Serious Wounds | 中度治療術 | 治療重傷 | 同一原文在神殿服務已用「治療重傷」 |
| Cause Serious Wounds | 中度致傷術 | 造成重傷 | 與 Cure 家族對稱 |
| Cure Critical Wounds | 重度治療術 | 治療致命傷 | 同一原文在神殿服務已用「治療致命傷」 |
| Cause Critical Wounds | 重度致傷術 | 造成致命傷 | 與 Cure 家族對稱 |

五筆均同步到 UI 與 game-pack 繁中 catalog。回歸測試另釘住
Cure Serious／Critical 在法術清單與神殿服務必須同名，避免再度漂移。

## 證據邊界

這一輪只改可由同一原文、已有專案譯名或明顯詞義錯誤證實的五筆。
其餘法術名稱雖已有鍵且跨 catalog 一致，仍不因此宣稱所有文學、桌遊
或出版慣用譯名已完成審定。
