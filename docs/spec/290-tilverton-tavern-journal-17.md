# 290：提爾佛頓酒館與手札 17

狀態：READY

## 證據

- `GEO2.DAX` block 1 的 `(6,10)` terrain selector 是 `0x88`。
- 真實 ECL2 流程顯示 `PICTURE 4`、`HEAD4`、`BODY4`，再提供
  `PUNCH BARKEEP / HAVE A DRINK / Leave`。
- 選擇 `HAVE A DRINK → LEMONADE → YES` 後，腳本描述繫紫色腰帶的女子由側門進入；
  繼續調查會找到 `ORNATE KNIFE`，並明確要求記錄 `JOURNAL ENTRY 17`。
- 英文 Adventure Journal 的 Entry 17 是火焰形刀刃匕首插圖，沒有可逐句翻譯的英文正文。

## 實作契約

- 選單與劇情文字由 State adapter 翻成繁中；VM 保留原始英文 choices 作分支索引。
- Entry 17 只在事件真正發生後加入遊戲內手札，中文內容描述原手札插圖，不提前洩漏。
- ECL 正常 EXIT 回到原 `(6,10)` 地城格，並清空 press-button choice；不可依攻略提到的
  巷子座標自行傳送角色，除非原始 ECL 或 engine CALL 提供座標 mutation 證據。
- remake 固定使用 640×480 logical canvas。原始 PIC／HEAD／BODY 採 nearest-neighbor
  整數放大；中文在放大後畫布以約 24px 字型重新排版，兩條 pipeline 不混用。

## 驗證

- real-image regression 必須走完整 `0x88` 路徑，鎖定人物素材、四種飲料、兩次 YES/NO、
  紫色腰帶、華麗匕首、手札 17 解鎖及返回地城。
- `-tavern` 提供可重現 640×480 畫面入口。
