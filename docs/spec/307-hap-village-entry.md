# 第三百零七輪：哈普村入口與躲藏村民

狀態：`READY`

## ENTER CITY transition

哈普村外的 `ENTER CITY` 不是抽象 city-service menu。ECL1 world dispatcher
先設定 area selector，再 `NEWECL 0x31`；作品 adapter 必須同步切到 Area 5
與 dungeon exploration，不能繼續沿用提爾佛頓的 Area 2 renderer state。

ECL5 block `0x31` initial entry：

- `LOAD PIECES 12,0xFF,0xFF`；
- 顯示荒廢村莊、空街與緊閉窗戶的介紹；
- Continue 後進入可探索的哈普村；
- 沒有在此入口請求 PICTURE、COMBAT 或 TREASURE。

640×480 frontend 顯示「哈普村」探索提示；原始 map／人物素材仍只做
nearest-neighbour 整數倍放大，繁中敘事以 24px 重繪。

## SearchLocation terrain dispatch

SearchLocation entry `+0x04A5` 將 `C04F & 0x7F` 作為 12-entry
`ON GOTO` selector。已確認 terrain `0x84` 指向 `+0x052C` 的民宅事件。
事件另有 `4BC9 > 14` gate；本輪只保留已觀察比較，不替該 engine-owned
work byte猜測名稱。

第一次觸發：

- `4C02` 由 `0` 寫成 `1`，防止事件重播；
- HEAD selector `7EE1=0x32`；
- `PICTURE 0x32`（decimal 50）；
- 村民要求隊伍趕快離開，menu 為 `LEAVE / TRY TO TALK FURTHER`。

選 `TRY TO TALK FURTHER` 後，村民逃到街上；Continue 返回同一份
ECL `0x31` dungeon runtime。PICTURE 關閉、menu 選擇與返回探索都不可
重建新的 session，否則 `4C02` visited flag 會遺失。
