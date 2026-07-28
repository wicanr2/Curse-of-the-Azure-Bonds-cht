# 第三百三十輪：艾森布拉露天酒館與手札 18 定位

狀態：`READY`

## 原版資料證據

ECL1 block `0x50` 的艾森布拉城市選單提供
`INN / STORE / HALL / TEMPLE / BAR / LEAVE`。選擇 `BAR` 後先進入俯瞰林地
的露天酒館，提供 `HAVE A DRINK / RELAX / EXIT`；目前固定 real-session
路徑第一次選擇 `RELAX` 時，原始 script 輸出 Tavern Tale `60`，Continue
後回到同一個酒館選單。Adventure Journal 正文為：有幾道龐大身影飛越森林，
朝南方去了。

`HAVE A DRINK` 另開
`BEER / ALE / PORT / MEAD / WHISKEY / EXIT` 六項選單；固定 regression
選擇 `BEER` 後，script 輸出 Tavern Tale `44`。手札正文指出紅袍法師偏好
火焰生物，寒冷攻擊往往是最佳防禦。Continue 返回酒館，酒館 `EXIT` 再經
原版事件頁回到艾森布拉六場所選單。

同一 DAX 區段較前方另有兜帽灰袍人與 Adventurer's Journal entry `18`，
但它不是艾森布拉 `RELAX` 的輸出。攻略的城市流程明確指出這是暗影谷酒館中
偽裝的伊爾明斯特；不能僅因 packed-text candidate 在 binary 中相鄰，就把
兩段資料視為同一城市事件。

隨附英文 Adventurer's Journal 的條目 18 說明，弦月枷印與伊爾明斯特的
徽記極為相似；隊伍應離開暗影谷，經阿沙本福德往南尋找紅袍法師的塔，要求
解除枷印。手札原文是玩家應在事件觸發後查閱的資料，不應在新遊戲開始時提前
揭露。

## 引擎與中文契約

- 城市場所選單是 ECL `ModeWilderness` 互動，`BAR` 必須送回同一個
  `BlockSession`；只有重製版自行建立的 `ModePlace` 選單才使用通用酒館。
- 艾森布拉露天酒館、五種酒名與傳聞 44／60 以 640×480 中文 HUD 顯示，
  不修改 ECL 分支、memory flag 或選項索引。
- 灰袍人事件依原始訊息組合辨識；手札 18 使用繁中完整轉述，但只在暗影谷
  script 真正輸出條目編號時加入 `JournalPages`。
- `appendJournalPages` 以「手札條目 18：」去重，重訪酒館不得重複加入頁面。

## 可沿用的 Gold Box 知識

Gold Box 城市服務不一定由引擎層的通用 `BAR`／`INN` handler 完成。同名選項
若來自 ECL menu，可能先執行一次性劇情、journal flag 或 encounter，再落入
服務介面。移植其他作品時應以「目前是否持有 ECL session」與 menu 所屬 mode
判斷 dispatch owner，不能只靠英文選項字串攔截。

Journal entry 的完整內容通常位於紙本 Adventurer's Journal，程式只輸出
條目編號。可靠重製需同時保存兩層證據：ECL 負責何時解鎖，手冊負責玩家可讀
內容；翻譯層不得以手冊文字反推或改寫 script control flow。

## 驗收

- 端到端 real-image regression 從新遊戲沿既有真實 session 通過法師塔、
  龍巫妖與艾森布拉入口，再選 `BAR → RELAX`。
- 畫面先顯示艾森布拉露天酒館，再由 `RELAX` 顯示繁中傳聞 60 並停在原版
  Continue 選單。
- 此分支不得錯誤新增手札 18。
- Continue 後仍有 `HAVE A DRINK / RELAX / EXIT`；飲酒選單六項須維持原始
  index，啤酒後顯示傳聞 44。
- 酒館 `EXIT` 經 Continue 回到艾森布拉
  `INN / STORE / HALL / TEMPLE / BAR / LEAVE`，不能停住或誤回通用酒館。
