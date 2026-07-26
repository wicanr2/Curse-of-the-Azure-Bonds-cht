# 第 115 輪：Shop APPRAISE gems／jewelry

狀態：`READY`

## 本輪成果

依原版手冊，`STORE → 估價` 會先選 party character，再選其 gems 或 jewelry；
本輪以 `AppraisalOffers` 注入店家報價，選取後將該類財寶清空並把報價加入 party
money pool，顯示繁中結果後返回 Shop Menu。

報價的 `Ready` flag 與零 GP 報價分開保存；未載入報價、角色沒有該類財寶與未知
treasure kind 都會安全失敗。這輪不推測 DOS `Gems`／`Jewelry` 的價格語意，完整
appraisal UI／拒絕報價分支仍待接入。
