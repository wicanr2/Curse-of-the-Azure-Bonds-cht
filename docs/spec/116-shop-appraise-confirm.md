# 第 116 輪：APPRAISE 接受／拒絕報價

狀態：`READY`

## 本輪成果

APPRAISE 選擇 gems／jewelry 後不立即出售，而是顯示繁中確認選單：

- `接受`：使用 `AppraisalOffers` 報價，清除選定財寶並將 GP 加入 party pool。
- `拒絕`：保留財寶、不改變 pool，顯示繁中結果。
- `返回`：回到財寶種類選單，不產生交易。

這對齊手冊的「shopkeeper makes an offer; accept or reject」流程；報價、財寶
原始價值與更完整的 appraisal side effects 仍由資料／script 層提供。
