# 第 260 輪：ITEM NameNumbers 繁中顯示

狀態：`READY`（限已確認 itemNames entries、hidden flag 與 renderer-neutral display）

## 已完成

- 依 reference `Item.GenerateName` 的順序，`NameNumbers[3]`、`[2]`、`[1]` 會由高到低組合顯示。
- `HiddenNameFlags` bit `0x4/0x2/0x1` 分別控制第一／第二／第三 name number，未公開或未知的 name number 不會被假造翻譯。
- 接入已確認的常見 magical components：`+1..+5`、negative plus、偏移、屠龍者、霜寒、火焰舌、力量、治療、防護、精靈之力、寶石／珠寶等。
- `ChineseName` 保留既有 item type、cursed、plus、stack 顯示，並對已知 name number 產生繁中名稱。

## 保留邊界

完整 `itemNames` 全表翻譯、Identify 對 affect／cursed／magic library 的原版 side effect，以及未知 name number 的語意仍待逐項驗證；raw fields 不會因 UI 翻譯而改寫。

## 回歸

測試覆蓋 `+1 長劍` name-number composition，以及 `HiddenNameFlags=0x4` 隱藏第一個 name number 的 reference boundary。
