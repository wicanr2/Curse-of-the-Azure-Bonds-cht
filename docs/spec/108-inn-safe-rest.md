# 第 108 輪：城市客棧安全休息

狀態：`READY`

## 證據與行為

原版英文手冊的城市說明指出，Inns 是安全休息場所，隊伍可在客棧使用 Camp
Menu 休息。remake 目前已能從城市 map 進入 `INN`，本輪將這個已觀察的場所選項
接成第一個實際 gameplay service：

- 選擇 `INN` 後，party roster 所有角色的 current HP 恢復至 Max HP。
- 目前畫面上的 `combat.Fighter` 同步相同 HP，避免回到戰鬥時顯示舊值。
- 顯示繁中休息訊息；`Continue` 返回原本的城市場所選單。
- 沒有 party roster 時仍可恢復已存在的 fighter HP，不虛構角色資料。

商店、酒館情報、完整 Camp Menu（守夜、法術恢復、毒／疾病與中斷）仍是後續
獨立規格；客棧安全休息不等同於完整原版 CAMP。
