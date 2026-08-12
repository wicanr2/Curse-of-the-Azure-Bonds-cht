# 第五百五十六輪：手札長文分頁與睡眠術視覺資料契約

狀態：`READY`（實作與針對性回歸完成後生效）

## 問題

- Journal renderer 固定只畫七行，手札 48、56、57 等長文超出的部分會被裁掉。
- 摩安德之坑的原始事件已明示 `JOURNAL ENTRY 46`，但 game pack 只翻譯提示，
  沒有解鎖可重讀的 `journal.46`。
- 睡眠術已有 PC-98 `TWINKLE` runtime 與 1440 ms 時序，renderer 卻直接依
  `VisualTwinkle` 繪製，game pack 無法宣告或稽核這個效果。

## 契約

1. renderer 依實際字型 advance 與七行內容區切出顯示頁；所有顯示頁串回後必須
   與來源手札完全相同，不能遺失或重複文字。
2. save 與 State 仍保存來源 `journal.*` stable ID；一則來源手札可有多個顯示頁，
   不為排版另造會漂移的故事 ID。
3. `pit.zhentrim-scroll` 在原始文字規則命中時解鎖 `journal.46`。條目內容依中文
   手冊校訂摘要重述，證據等級為 `strong inference`，不冒充逐字轉錄。
4. engine `combat_visuals` 接受互斥的 `frames` 或 `generator`。CoAB 以
   `sleep／impact／pc98.twinkle` 明確綁定；renderer 無 binding 時不繪製。

## 證據邊界

- 手札 46 producer：ECL 顯示片段 `GRASPED IN THE FIGHTER'S FIST`、
  `SEAL OF ZHENTIL`、`JOURNAL ENTRY 46`；完整內容來源為
  `docs/manual/journal/entries-41-59.md` 的校訂摘要。
- 睡眠術幾何與時序沿用 READY spec 442。動態圖形的存在與時序為 `exact`；
  現行 palette 像素仍是 `layout-reconstructed`。
- 本輪只把睡眠術既有行為資料化，不代表其餘十一個法術或完整戰鬥已完成。

