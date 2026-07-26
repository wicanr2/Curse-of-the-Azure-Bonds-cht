# 第 113 輪：Shop VIEW 角色／裝備摘要

狀態：`READY`

## 本輪成果

`STORE → 查看` 會列出 party roster 的角色、current/max HP 與 gold；選取角色後
顯示目前已保存的繁中裝備名稱，按 Enter 可返回 Shop Menu。資料直接來自
`party.Roster`，包含 DOS sidecar 匯入與 remake save/load 的欄位。

這是 renderer-neutral 的角色摘要，不宣稱完整原版 VIEW／ALTER／ID menu；角色
選擇、裝備詳細屬性、未識別物品與 magic effect 仍由後續規格處理。
