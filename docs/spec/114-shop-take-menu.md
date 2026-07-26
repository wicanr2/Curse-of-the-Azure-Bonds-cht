# 第 114 輪：Shop TAKE 角色／金額選單

狀態：`READY`

## 本輪成果

`STORE → 取出金幣` 現在會：

1. 列出 party roster 角色。
2. 讓玩家選擇角色。
3. 提供 1／10／100／全部 pool 金幣等可用金額。
4. 呼叫 `State.TakeGold`、更新角色 gold 與 pool，顯示繁中結果並返回 Shop Menu。

金額選項是目前 UI 的 bounded input；底層 API 仍接受任意 `uint32` amount，並保留
pool 不足與角色 `uint16` gold overflow 檢查。原版完整任意數字輸入／VIEW menu
操作仍可由後續 UI layer 擴充。
