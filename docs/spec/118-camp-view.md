# 第一百一十八輪：CAMP VIEW 角色摘要

狀態：`READY`（限目前 remake 的只讀角色查看 boundary）

## 實作 contract

`CAMP → VIEW` 會以 `partyRoster` 為資料來源建立角色選單。選取角色後顯示姓名、職業、HP、金幣、寶石、珠寶與已知裝備的繁中名稱；此畫面不修改角色資料。Enter 返回角色列表，選擇返回項才回到 CAMP Menu。

角色職業名稱集中在 game adapter，裝備名稱仍經由既有 `monster.ChineseName`；未識別 item 不在 VIEW 層猜測魔法效果或原版 ALTER side effect。

## 驗證

`TestCampMenuViewCharacterAndReturn` 覆蓋 CAMP 進入、角色列表、摘要內容、返回列表與離開 VIEW。`go test ./...` 已通過。
